/**
 * Copyright 2025 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package integrationtest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var (
	projectRoot, _         = filepath.Abs("../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "08-network-security-integration/SecurityProfile")
	configFolderPath       = filepath.Join(projectRoot, "test/integration/network-security-integration/SecurityProfile/config")
	resourceVpcSubnetRange = "10.10.10.0/24"
)

func TestSecurityProfileIntegration(t *testing.T) {
	t.Parallel()
	projectID := os.Getenv("TF_VAR_project_id")
	if projectID == "" {
		t.Fatal("Skipping test: environment variable TF_VAR_project_id is not set.")
	}
	orgID := getOrgIDFromProject(t, projectID)
	billingProjectID := os.Getenv("TF_VAR_billing_project_id")
	if billingProjectID == "" {
		billingProjectID = projectID
	}

	t.Log("Setting environment variables for test execution...")
	t.Log("Setting USER PROJECT OVERRIDE environment variables for test execution...")
	err := os.Setenv("USER_PROJECT_OVERRIDE", "true")
	if err != nil {
		t.Fatalf("Failed to set USER_PROJECT_OVERRIDE environment variable: %v", err)
	}
	t.Log("Setting GOOGLE BILLING PROJECT environment variables for test execution...")
	err = os.Setenv("GOOGLE_BILLING_PROJECT", billingProjectID)
	if err != nil {
		t.Fatalf("Failed to set GOOGLE_BILLING_PROJECT environment variable: %v", err)
	}

	instanceSuffix := strings.ToLower(random.UniqueId())
	vpcName := fmt.Sprintf("vpc-sp-test-%s", instanceSuffix)
	firewallPolicyName := fmt.Sprintf("fwp-sp-test-%s", instanceSuffix)
	zone := "us-central1-a"
	t.Logf("Test Run Config: ProjectID=%s, OrgID=%s, Zone=%s, Suffix=%s", projectID, orgID, zone, instanceSuffix)

	createVPC(t, projectID, vpcName, zone)
	defer deleteVPC(t, projectID, vpcName, zone)

	vmClientName := "vm-client-" + instanceSuffix
	vmServerName := "vm-server-" + instanceSuffix
	successMessage := "CONNECTIVITY_TEST_PASSED_AS_EXPECTED"
	failureMessage := "CONNECTIVITY_TEST_FAILED_UNEXPECTED_SUCCESS"
	serverStartupScript := "#!/bin/bash\nsudo apt-get update\nsudo apt-get install -y nginx\nsudo systemctl start nginx"
	clientStartupScript := fmt.Sprintf(`
        #!/bin/bash
        apt-get update -y
        apt-get install -y curl
        # Try to connect to the server VM by its instance name.
        # We expect this to fail (timeout), which is a success for this test.
        curl --connect-timeout 15 http://%s
        if [ $? -ne 0 ]; then
            echo "%s"
        else
            echo "%s"
        fi
    `, vmServerName, successMessage, failureMessage)

	createVM(t, projectID, vmServerName, zone, vpcName, serverStartupScript)
	defer deleteVM(t, projectID, vmServerName, zone)
	createVM(t, projectID, vmClientName, zone, vpcName, clientStartupScript)
	defer deleteVM(t, projectID, vmClientName, zone)

	profileGroupName := "spg-integ-test-" + instanceSuffix
	createConfigYAML(t, orgID, "sp-integ-test-"+instanceSuffix, profileGroupName)

	createFirewallPolicy(t, orgID, firewallPolicyName)
	defer deleteFirewallPolicy(t, orgID, firewallPolicyName)

	tfVars := map[string]interface{}{
		"config_folder_path": configFolderPath,
	}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
		Reconfigure:  true,
		NoColor:      true,
	})

	defer terraform.Destroy(t, terraformOptions)
	t.Log("Running terraform init and apply...")
	terraform.InitAndApply(t, terraformOptions)
	t.Log("Terraform apply complete.")

	addRuleAndAssociateFirewallPolicy(t, orgID, firewallPolicyName, vpcName, projectID, profileGroupName)
	defer deleteRuleAndFirewallPolicyAssociation(t, orgID, firewallPolicyName)
	t.Logf("Waiting up to 5 minutes for startup script on client VM '%s' to complete and write to logs...", vmClientName)
	for i := 0; i < 20; i++ { // Poll for 5 minutes (20 * 15s)
		time.Sleep(15 * time.Second)
		output, err := getSerialPortOutput(t, projectID, vmClientName, zone)
		if err != nil {
			t.Logf("Warning: could not get serial port output on attempt %d: %v", i+1, err)
			continue
		}

		if strings.Contains(output, successMessage) {
			t.Logf("Success! Found expected message in serial port logs for VM '%s'. The security profile correctly blocked traffic.", vmClientName)
			return // Test passes
		}
		if strings.Contains(output, failureMessage) {
			t.Fatalf("Failure! The client VM was able to connect to the server, but it should have been blocked. Full output:\n%s", output)
		}
	}

	t.Fatalf("Timeout: Did not find success or failure message in serial port logs for VM '%s' after 5 minutes.", vmClientName)
}

func createConfigYAML(t *testing.T, orgID, profileName, groupName string) {
	type securityProfile struct {
		Create                  bool                   `yaml:"create"`
		Name                    string                 `yaml:"name"`
		Type                    string                 `yaml:"type"`
		Description             string                 `yaml:"description"`
		ThreatPreventionProfile map[string]interface{} `yaml:"threat_prevention_profile"`
	}
	type securityProfileGroup struct {
		Create bool   `yaml:"create"`
		Name   string `yaml:"name"`
	}
	type testConfig struct {
		OrgID   string               `yaml:"organization_id"`
		Profile securityProfile      `yaml:"security_profile"`
		Group   securityProfileGroup `yaml:"security_profile_group"`
		Link    bool                 `yaml:"link_profile_to_group"`
	}
	config := testConfig{
		OrgID: orgID,
		Profile: securityProfile{
			Create:      true,
			Name:        profileName,
			Type:        "THREAT_PREVENTION",
			Description: "Deny INFORMATIONAL traffic for testing",
			ThreatPreventionProfile: map[string]interface{}{
				"severity_overrides": []map[string]string{
					{"severity": "INFORMATIONAL", "action": "DENY"},
				},
			},
		},
		Group: securityProfileGroup{
			Create: true,
			Name:   groupName,
		},
		Link: true,
	}
	yamlData, err := yaml.Marshal(&config)
	assert.NoError(t, err)

	err = os.MkdirAll(configFolderPath, 0755)
	assert.NoError(t, err)

	filePath := filepath.Join(configFolderPath, "instance.yaml")
	err = os.WriteFile(filePath, yamlData, 0644)
	assert.NoError(t, err)
	t.Logf("Created test YAML config file: %s", filePath)
}

func createFirewallPolicy(t *testing.T, orgID, policyName string) {
	t.Logf("Creating Firewall Policy '%s' in Org '%s'", policyName, orgID)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-policies", "create", "--short-name=" + policyName, "--organization=" + orgID, "--description=integ-test-policy"}})
}

func addRuleAndAssociateFirewallPolicy(t *testing.T, orgID, policyName, vpcName, projectID, profileGroupName string) {
	profileGroupPath := fmt.Sprintf("organizations/%s/locations/global/securityProfileGroups/%s", orgID, profileGroupName)
	vpcPath := fmt.Sprintf("projects/%s/global/networks/%s", projectID, vpcName)

	t.Logf("Adding rule to policy '%s' to apply security profile group '%s'", policyName, profileGroupPath)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-policies", "rules", "create", "1000", "--firewall-policy=" + policyName, "--organization=" + orgID, "--action=apply_security_profile_group", "--security-profile-group=" + profileGroupPath, "--src-ip-ranges=" + resourceVpcSubnetRange, "--layer4-configs=all", "--enable-logging", "--description=test-rule"}})

	t.Logf("Associating policy '%s' with VPC '%s'", policyName, vpcPath)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-policies", "associations", "create", "--firewall-policy=" + policyName, "--organization=" + orgID, fmt.Sprintf("--name=%s-association", policyName), "--replace-association-on-target"}})
}

func deleteRuleAndFirewallPolicyAssociation(t *testing.T, orgID, policyName string) {
	if policyName == "" {
		return
	}
	t.Logf("--- Deleting Firewall Policy Association: %s-association ---", policyName)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-policies", "associations", "delete", fmt.Sprintf("%s-association", policyName), "--firewall-policy=" + policyName, "--organization=" + orgID}})

	t.Logf("--- Deleting Firewall Policy Rule '1000' from policy '%s' ---", policyName)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-policies", "rules", "delete", "1000", "--firewall-policy=" + policyName, "--organization=" + orgID}})
}

func deleteFirewallPolicy(t *testing.T, orgID, policyName string) {
	if policyName == "" {
		return
	}
	t.Logf("--- Deleting Firewall Policy: %s ---", policyName)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-policies", "delete", policyName, "--organization=" + orgID, "--quiet"}})
}

func getRegionFromZone(t *testing.T, zone string) string {
	lastHyphen := strings.LastIndex(zone, "-")
	if lastHyphen == -1 {
		t.Fatalf("Invalid zone format: %s. Expected format like 'us-central1-a'", zone)
	}
	return zone[:lastHyphen]
}

func createVPC(t *testing.T, projectID, networkName, zone string) {
	region := getRegionFromZone(t, zone)
	subnetName := fmt.Sprintf("%s-subnet", networkName)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "create", networkName, "--project=" + projectID, "--subnet-mode=custom"}})
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "subnets", "create", subnetName, "--project=" + projectID, "--network=" + networkName, "--range=" + resourceVpcSubnetRange, "--region=" + region}})
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-rules", "create", fmt.Sprintf("fw-allow-http-internal-%s", networkName), "--project=" + projectID, "--network=" + networkName, "--allow=tcp:80", "--source-ranges=" + resourceVpcSubnetRange}})
}

func deleteVPC(t *testing.T, projectID, networkName, zone string) {
	if networkName == "" {
		return
	}
	t.Logf("--- Deleting VPC: %s ---", networkName)
	region := getRegionFromZone(t, zone)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-rules", "delete", fmt.Sprintf("fw-allow-http-internal-%s", networkName), "--project=" + projectID, "--quiet"}})
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "subnets", "delete", fmt.Sprintf("%s-subnet", networkName), "--project=" + projectID, "--region=" + region, "--quiet"}})
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "delete", networkName, "--project=" + projectID, "--quiet"}})
}

func createVM(t *testing.T, projectID, vmName, zone, networkName, startupScript string) {
	t.Logf("Creating VM: %s in zone %s", vmName, zone)
	subnetName := fmt.Sprintf("%s-subnet", networkName)
	cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "instances", "create", vmName,
		"--project=" + projectID,
		"--zone=" + zone,
		"--machine-type=e2-micro",
		"--subnet=" + subnetName,
		"--no-address",
		"--image-family=ubuntu-2204-lts", "--image-project=ubuntu-os-cloud",
		fmt.Sprintf("--metadata-from-file=startup-script=%s", createStartupScriptFile(t, startupScript)),
	}}
	shell.RunCommand(t, cmd)
}

func createStartupScriptFile(t *testing.T, scriptContent string) string {
	if scriptContent == "" {
		scriptContent = "#!/bin/bash\n# No startup script"
	}
	file, err := os.CreateTemp("", "startup-script-*.sh")
	assert.NoError(t, err)
	_, err = file.WriteString(scriptContent)
	assert.NoError(t, err)
	file.Close()
	t.Cleanup(func() { os.Remove(file.Name()) })
	return file.Name()
}

func deleteVM(t *testing.T, projectID, vmName, zone string) {
	if vmName == "" {
		return
	}
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "instances", "delete", vmName, "--project=" + projectID, "--zone=" + zone, "--quiet"}})
}

func getOrgIDFromProject(t *testing.T, projectID string) string {
	t.Logf("Attempting to find Organization ID for project '%s'...", projectID)
	args := []string{"projects", "describe", projectID, "--format=value(parent.id)"}
	cmd := shell.Command{Command: "gcloud", Args: args}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	require.NoError(t, err, "Failed to run gcloud command to get organization ID for project %s: %v", projectID, err)
	orgID := strings.TrimSpace(output)
	require.NotEmpty(t, orgID, "Organization ID was not found for project %s. Ensure the project is directly under an organization.", projectID)
	t.Logf("Found Organization ID: %s", orgID)
	return orgID
}

func getSerialPortOutput(t *testing.T, projectID, vmName, zone string) (string, error) {
	args := []string{
		"compute", "instances", "get-serial-port-output", vmName,
		"--project=" + projectID,
		"--zone=" + zone,
		"--port=1",
	}
	cmd := exec.Command("gcloud", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
