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
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/test/integration/common_utils"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var (
	projectRoot, _         = filepath.Abs("../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "03-security/SecurityProfile")
	configFolderPath       = filepath.Join(projectRoot, "test/integration/security/SecurityProfile/config")
	resourceVpcSubnetRange = "10.0.1.0/24"
)

func TestSecurityProfileIntegration(t *testing.T) {
	t.Parallel()
	projectID := os.Getenv("TF_VAR_project_id")
	require.NotEmpty(t, projectID, "Skipping test: environment variable TF_VAR_project_id is not set.")

	orgID := common_utils.GetOrgIDFromProject(t, projectID)
	billingProjectID := os.Getenv("TF_VAR_billing_project_id")
	if billingProjectID == "" {
		billingProjectID = projectID
	}

	t.Log("Setting environment variables...")
	err := os.Setenv("USER_PROJECT_OVERRIDE", "true")
	require.NoError(t, err)
	err = os.Setenv("GOOGLE_BILLING_PROJECT", billingProjectID)
	require.NoError(t, err)

	instanceSuffix := strings.ToLower(random.UniqueId())
	vpcName := fmt.Sprintf("vpc-sp-test-%s", instanceSuffix)
	subnetName := fmt.Sprintf("%s-subnet", vpcName)
	firewallPolicyName := fmt.Sprintf("fwp-sp-test-%s", instanceSuffix)
	zone := "us-central1-a"
	t.Logf("Test Run Config: ProjectID=%s, OrgID=%s, Zone=%s, Suffix=%s", projectID, orgID, zone, instanceSuffix)
	region := common_utils.GetRegionFromZone(t, zone)
	common_utils.CreateVPCSubnets(t, projectID, vpcName, subnetName, region)
	defer common_utils.DeleteVPCSubnets(t, projectID, vpcName, subnetName, region)
	common_utils.CreateFirewallRules(t, projectID, vpcName, instanceSuffix)
	defer common_utils.DeleteFirewallRules(t, projectID, instanceSuffix)

	vmClientName := "vm-client-" + instanceSuffix
	vmServerName := "vm-server-" + instanceSuffix
	successMessage := "CONNECTIVITY_TEST_PASSED_AS_EXPECTED"
	failureMessage := "CONNECTIVITY_TEST_FAILED_UNEXPECTED_SUCCESS"
	serverStartupScript := "#!/bin/bash\nsudo apt-get update\nsudo apt-get install -y nginx\nsudo systemctl start nginx"
	clientStartupScript := fmt.Sprintf(`
        #!/bin/bash
        apt-get update -y; apt-get install -y curl
        curl --connect-timeout 15 http://%s
        if [ $? -ne 0 ]; then echo "%s"; else echo "%s"; fi
    `, vmServerName, successMessage, failureMessage)

	common_utils.CreateGCEInstance(t, projectID, vmServerName, zone, subnetName, serverStartupScript, "", false)
	defer common_utils.DeleteGCEInstance(t, projectID, vmServerName, zone)
	common_utils.CreateGCEInstance(t, projectID, vmClientName, zone, subnetName, clientStartupScript, "", false)
	defer common_utils.DeleteGCEInstance(t, projectID, vmClientName, zone)

	profileGroupName := "spg-integ-test-" + instanceSuffix
	createConfigYAML(t, orgID, "sp-integ-test-"+instanceSuffix, profileGroupName)

	common_utils.CreateOrgFirewallPolicy(t, orgID, firewallPolicyName)
	defer common_utils.DeleteOrgFirewallPolicy(t, orgID, firewallPolicyName)

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

	common_utils.AddSecurityProfileRuleAndAssociatePolicy(t, orgID, firewallPolicyName, vpcName, projectID, profileGroupName, resourceVpcSubnetRange)
	defer common_utils.DeleteSecurityProfileRuleAndPolicyAssociation(t, orgID, firewallPolicyName)

	t.Logf("Waiting up to 5 minutes for startup script on client VM '%s' to complete...", vmClientName)
	for i := 0; i < 20; i++ {
		time.Sleep(15 * time.Second)
		output, err := common_utils.GetSerialPortOutput(t, projectID, vmClientName, zone, 1)
		if err != nil {
			t.Logf("Warning: could not get serial port output on attempt %d: %v", i+1, err)
			continue
		}
		if strings.Contains(output, successMessage) {
			t.Logf("Success! Found expected message. The security profile correctly blocked traffic.")
			return
		}
		if strings.Contains(output, failureMessage) {
			t.Fatalf("Failure! The client VM connected to the server, but it should have been blocked.\nOutput:\n%s", output)
		}
	}
	t.Fatalf("Timeout: Did not find success or failure message in serial port logs for VM '%s'.", vmClientName)
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
