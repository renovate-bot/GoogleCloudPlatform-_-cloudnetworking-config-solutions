// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

var (
	projectRoot, _           = filepath.Abs("../../../../")
	terraformDirectoryPath   = filepath.Join(projectRoot, "02-networking/FirewallEndpoint")
	configFolderPath         = filepath.Join(projectRoot, "test/integration/networking/FirewallEndpoint/config")
	inspectionVpcSubnetRange = "10.20.10.0/24"
	protectedVpcSubnetRange  = "10.30.10.0/24"
	sshFirewallRange         = "35.235.240.0/20"
	internalSrcRange         = "10.0.0.0/8"
)

func TestFirewallEndpointIntegration(t *testing.T) {
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

	currentUser := common_utils.GetCurrentGcloudUser(t)
	instanceSuffix := strings.ToLower(random.UniqueId())
	vpcInspectionName := fmt.Sprintf("vpc-inspection-fe-test-%s", instanceSuffix)
	vpcProtectedName := fmt.Sprintf("vpc-protected-fe-test-%s", instanceSuffix)
	zone := "us-central1-a"

	t.Logf("Test Run Config: ProjectID=%s, OrgID=%s, Zone=%s", projectID, orgID, zone)
	t.Logf("Running test as primary user: %s", currentUser)

	err = createPeeredVPCs(t, projectID, vpcInspectionName, vpcProtectedName)
	require.NoError(t, err)
	defer deletePeeredVPCs(t, projectID, vpcInspectionName, vpcProtectedName)

	endpointName := "fw-ep-integ-test-" + instanceSuffix
	assocName := "assoc-integ-test-" + instanceSuffix
	createConfigYAML(t, orgID, billingProjectID, projectID, vpcProtectedName, zone, endpointName, assocName)

	tfVars := map[string]interface{}{"config_folder_path": configFolderPath}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
		Reconfigure:  true,
		NoColor:      true,
		EnvVars:      map[string]string{"GOOGLE_PROJECT": projectID},
	})

	defer terraform.Destroy(t, terraformOptions)
	t.Log("Running terraform init and apply...")
	terraform.InitAndApply(t, terraformOptions)
	t.Log("Terraform apply complete.")

	t.Log("Validating Terraform outputs...")
	outputJson := terraform.OutputJson(t, terraformOptions, "firewall_endpoints")
	require.True(t, gjson.Valid(outputJson), "Output 'firewall_endpoints' is not valid JSON")
	endpointID := gjson.Get(outputJson, "instance.id").String()
	assert.NotEmpty(t, endpointID, "Could not find 'id' in 'firewall_endpoints' output")
	t.Logf("Validation successful: Found endpoint ID: %s", endpointID)

	t.Log("Validating control plane configuration...")
	err = verifyControlPlaneConfiguration(t, projectID)
	assert.NoError(t, err, "Control plane configuration validation failed.")
}

// createConfigYAML remains local as it's test-specific.
func createConfigYAML(t *testing.T, orgID, billingProjectID, assocProjectID, vpcName, location, endpointName, assocName string) {
	type firewallEndpoint struct {
		Create           bool   `yaml:"create"`
		Name             string `yaml:"name"`
		OrganizationID   string `yaml:"organization_id"`
		BillingProjectID string `yaml:"billing_project_id"`
	}
	type firewallEndpointAssociation struct {
		Create               bool   `yaml:"create"`
		Name                 string `yaml:"name"`
		AssociationProjectID string `yaml:"association_project_id"`
		NetworkSelfLink      string `yaml:"vpc_id"`
	}
	type testConfig struct {
		Location string                      `yaml:"location"`
		Endpoint firewallEndpoint            `yaml:"firewall_endpoint"`
		Assoc    firewallEndpointAssociation `yaml:"firewall_endpoint_association"`
	}
	networkSelfLink := fmt.Sprintf("projects/%s/global/networks/%s", assocProjectID, vpcName)
	config := testConfig{
		Location: location,
		Endpoint: firewallEndpoint{Create: true, Name: endpointName, OrganizationID: orgID, BillingProjectID: billingProjectID},
		Assoc:    firewallEndpointAssociation{Create: true, Name: assocName, AssociationProjectID: assocProjectID, NetworkSelfLink: networkSelfLink},
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

// verifyControlPlaneConfiguration remains local as it's test-specific.
func verifyControlPlaneConfiguration(t *testing.T, projectID string) error {
	cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "routes", "list", "--project=" + projectID, "--format=json"}}

	t.Logf("Verifying that an auto-generated peering route exists...")

	var lastErr error
	const maxRetries = 10
	const sleepBetweenRetries = 30 * time.Second

	for i := 0; i < maxRetries; i++ {
		routesJson, err := shell.RunCommandAndGetOutputE(t, cmd)
		if err != nil {
			lastErr = fmt.Errorf("failed to list routes with gcloud: %w", err)
			t.Logf("Attempt %d/%d: Failed to list routes, retrying in %v...", i+1, maxRetries, sleepBetweenRetries)
			time.Sleep(sleepBetweenRetries)
			continue
		}
		parsedRoutes := gjson.Parse(routesJson)
		routeFound := false
		for _, route := range parsedRoutes.Array() {
			description := route.Get("description").String()
			if strings.Contains(description, "Auto generated route via peering") {
				t.Logf("Validation successful on attempt %d/%d. Found peering route '%s'.", i+1, maxRetries, route.Get("name").String())
				routeFound = true
				break
			}
		}
		if routeFound {
			return nil
		}
		lastErr = fmt.Errorf("could not find a route with a description indicating it was auto-generated by peering")
		t.Logf("Attempt %d/%d: Required peering route not found, retrying in %v...", i+1, maxRetries, sleepBetweenRetries)
		time.Sleep(sleepBetweenRetries)
	}
	return lastErr
}

// createPeeredVPCs is kept local as it defines a complex, test-specific scenario that the generic helpers cannot support.
func createPeeredVPCs(t *testing.T, projectID, inspectionVPC, protectedVPC string) error {
	t.Logf("Creating Inspection VPC '%s' and Protected VPC '%s'", inspectionVPC, protectedVPC)
	region := "us-central1"
	commands := []shell.Command{
		{Command: "gcloud", Args: []string{"compute", "networks", "create", inspectionVPC, "--project=" + projectID, "--subnet-mode=custom"}},
		{Command: "gcloud", Args: []string{"compute", "networks", "create", protectedVPC, "--project=" + projectID, "--subnet-mode=custom"}},
		{Command: "gcloud", Args: []string{"compute", "networks", "subnets", "create", fmt.Sprintf("%s-subnet", inspectionVPC), "--project=" + projectID, "--network=" + inspectionVPC, "--range=" + inspectionVpcSubnetRange, "--region=" + region}},
		{Command: "gcloud", Args: []string{"compute", "networks", "subnets", "create", fmt.Sprintf("%s-subnet", protectedVPC), "--project=" + projectID, "--network=" + protectedVPC, "--range=" + protectedVpcSubnetRange, "--region=" + region}},
		{Command: "gcloud", Args: []string{"compute", "firewall-rules", "create", fmt.Sprintf("fw-%s-allow-all", inspectionVPC), "--project=" + projectID, "--network=" + inspectionVPC, "--allow=all", "--source-ranges=" + internalSrcRange}},
		{Command: "gcloud", Args: []string{"compute", "firewall-rules", "create", fmt.Sprintf("fw-%s-allow-all", protectedVPC), "--project=" + projectID, "--network=" + protectedVPC, "--allow=all", "--source-ranges=" + internalSrcRange}},
		{Command: "gcloud", Args: []string{"compute", "firewall-rules", "create", fmt.Sprintf("fw-%s-allow-ssh", inspectionVPC), "--project=" + projectID, "--network=" + inspectionVPC, "--allow=tcp:22", "--source-ranges=" + sshFirewallRange}},
		{Command: "gcloud", Args: []string{"compute", "firewall-rules", "create", fmt.Sprintf("fw-%s-allow-ssh", protectedVPC), "--project=" + projectID, "--network=" + protectedVPC, "--allow=tcp:22", "--source-ranges=" + sshFirewallRange}},
	}
	for _, cmd := range commands {
		if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
			return fmt.Errorf("failed to run gcloud command %s: %w", cmd.Args[2], err)
		}
	}
	common_utils.CreateBiDirectionalVPCPeering(t, projectID, inspectionVPC, protectedVPC)
	return nil
}

// deletePeeredVPCs is kept local to match the specific setup logic.
func deletePeeredVPCs(t *testing.T, projectID, inspectionVPC, protectedVPC string) {
	if inspectionVPC == "" || protectedVPC == "" {
		return
	}
	t.Logf("--- Deleting Peered VPCs and their dependent resources: %s, %s ---", inspectionVPC, protectedVPC)
	common_utils.DeleteBiDirectionalVPCPeering(t, projectID, inspectionVPC, protectedVPC)
	rulesToDelete := []string{fmt.Sprintf("fw-%s-allow-all", inspectionVPC), fmt.Sprintf("fw-%s-allow-ssh", inspectionVPC), fmt.Sprintf("fw-%s-allow-all", protectedVPC), fmt.Sprintf("fw-%s-allow-ssh", protectedVPC)}
	subnetsToDelete := []string{fmt.Sprintf("%s-subnet", inspectionVPC), fmt.Sprintf("%s-subnet", protectedVPC)}
	for _, ruleName := range rulesToDelete {
		cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-rules", "delete", ruleName, "--project=" + projectID, "--quiet"}}
		shell.RunCommand(t, cmd)
	}
	for _, subnetName := range subnetsToDelete {
		region := "us-central1"
		cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "subnets", "delete", subnetName, "--project=" + projectID, "--region=" + region, "--quiet"}}
		shell.RunCommand(t, cmd)
	}
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "delete", inspectionVPC, "--project=" + projectID, "--quiet"}})
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "delete", protectedVPC, "--project=" + projectID, "--quiet"}})
}
