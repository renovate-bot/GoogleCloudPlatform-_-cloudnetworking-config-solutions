// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/test/integration/common_utils"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

var (
	projectRootPMR, _         = filepath.Abs("../../../../")
	terraformDirectoryPathPMR = filepath.Join(projectRootPMR, "08-network-security-integration/PacketMirroringRule")
	configFolderPathPMR       = filepath.Join(projectRootPMR, "test/integration/network-security-integration/PacketMirroringRule/config")
)

/*
TestCreatePacketMirroringRule is an end-to-end integration test for the Packet
Mirroring Terraform module. It creates all prerequisite resources, runs `terraform apply`
on the module, validates the created rule, and tears down all resources.
*/
func TestCreatePacketMirroringRule(t *testing.T) {
	projectID := os.Getenv("TF_VAR_project_id")
	orgID := os.Getenv("TF_VAR_organization_id")
	if !assert.NotEmpty(t, projectID, "TF_VAR_project_id env var must be set for integration tests") {
		t.FailNow()
	}
	if !assert.NotEmpty(t, orgID, "TF_VAR_organization_id env var must be set for integration tests") {
		t.FailNow()
	}
	instanceSuffix := strings.ToLower(random.UniqueId())
	vpcName := "vpc-pmr-test-" + instanceSuffix
	dgName := "dg-pmr-test-" + instanceSuffix
	egName := "eg-pmr-test-" + instanceSuffix
	spName := "sp-pmr-test-" + instanceSuffix
	spgName := "spg-pmr-test-" + instanceSuffix
	fwPolicyName := "fwp-pmr-test-" + instanceSuffix
	ruleName := "mirror-internal-traffic-" + instanceSuffix
	priority := 1000
	direction := "INGRESS"
	action := "mirror"
	srcIPRanges := []string{"10.100.0.1/32"}
	layer4Configs := []any{map[string]any{"ip_protocol": "all"}}
	t.Logf("Test Run Config: ProjectID=%s, OrgID=%s, Suffix=%s", projectID, orgID, instanceSuffix)
	common_utils.CreateVPCSubnets(t, projectID, vpcName, "", "")
	defer common_utils.DeleteVPCSubnets(t, projectID, vpcName, "", "")

	common_utils.CreateMirroringDeploymentGroup(t, projectID, dgName, vpcName)
	defer common_utils.DeleteMirroringDeploymentGroup(t, projectID, dgName)

	endpointGroupID := common_utils.CreateMirroringEndpointGroup(t, projectID, egName, dgName)
	defer common_utils.DeleteMirroringEndpointGroup(t, projectID, egName)

	common_utils.CreateSecurityProfileAndGroup(t, orgID, projectID, spName, spgName, endpointGroupID)
	defer common_utils.DeleteSecurityProfileAndGroup(t, orgID, spName, spgName)

	common_utils.CreateFirewallPolicy(t, projectID, fwPolicyName)
	defer common_utils.DeleteFirewallPolicy(t, projectID, fwPolicyName)

	securityProfileGroupPath := fmt.Sprintf("organizations/%s/locations/global/securityProfileGroups/%s", orgID, spgName)
	createPMRuleConfigYAML(t, projectID, fwPolicyName, securityProfileGroupPath, direction, action, ruleName, priority, srcIPRanges, layer4Configs)
	tfVars := map[string]any{
		"config_folder_path": configFolderPathPMR,
	}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathPMR,
		Vars:         tfVars,
		Reconfigure:  true,
		NoColor:      true,
	})
	defer terraform.Destroy(t, terraformOptions)

	t.Log("Running terraform init and apply...")
	terraform.InitAndApply(t, terraformOptions)
	t.Log("Terraform apply complete.")
	t.Log("Validating created Packet Mirroring Rule...")
	rulePriority := "1000"
	ruleOutput := common_utils.DescribeFirewallPolicyRule(t, projectID, fwPolicyName, rulePriority)
	assert.Contains(t, ruleOutput, "direction: INGRESS")
	expectedRuleNameString := fmt.Sprintf("ruleName: %s", ruleName)
	assert.Contains(t, ruleOutput, expectedRuleNameString)
	fullSpgPath := fmt.Sprintf("securityProfileGroup: //networksecurity.googleapis.com/%s", securityProfileGroupPath)
	assert.Contains(t, ruleOutput, fullSpgPath)
	t.Log("Validation successful.")
}

/*
createPMRuleConfigYAML is a helper function that dynamically generates the `instance.yaml`
configuration file. This file contains the parameters for the packet mirroring rule
that the Terraform module will create.
*/
func createPMRuleConfigYAML(t *testing.T, projectID, fwPolicyName, spgPath, direction, action, ruleName string, priority int, srcIPRanges []string, layer4Configs []any) {
	config := map[string]any{
		"priority":               priority,
		"rule_name":              ruleName,
		"project_id":             projectID,
		"firewall_policy_name":   fwPolicyName,
		"direction":              direction,
		"action":                 action,
		"security_profile_group": spgPath,
		"match": map[string]any{
			"src_ip_ranges":  srcIPRanges,
			"layer4_configs": layer4Configs,
		},
	}
	yamlData, err := yaml.Marshal(&config)
	assert.NoError(t, err)

	err = os.MkdirAll(configFolderPathPMR, 0755)
	assert.NoError(t, err)
	filePath := filepath.Join(configFolderPathPMR, "instance.yaml")
	err = os.WriteFile(filePath, yamlData, 0644)
	assert.NoError(t, err)
	t.Logf("Created test YAML config file: %s", filePath)
}
