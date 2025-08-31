// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package unittest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
)

var (
	projectRootPMR, _         = filepath.Abs("../../../../")
	terraformDirectoryPathPMR = filepath.Join(projectRootPMR, "08-network-security-integration/PacketMirroringRule")
	configFolderPathPMR       = filepath.Join(projectRootPMR, "test/unit/network-security-integration/PacketMirroringRule/config")
)

var (
	tfVarsPMR = map[string]any{
		"config_folder_path": configFolderPathPMR,
	}
)

func TestPMRulePlanSuccess(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathPMR,
		Vars:         tfVarsPMR,
		Reconfigure:  true,
		PlanFilePath: "./plan_pmr_success",
		NoColor:      true,
	})
	t.Cleanup(func() {
		planFilePath := filepath.Join(terraformOptions.TerraformDir, terraformOptions.PlanFilePath)
		err := os.Remove(planFilePath)
		if err != nil {
			t.Logf("Failed to remove plan file '%s', error: %v", planFilePath, err)
		}
	})
	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	assert.Equal(t, 2, planExitCode, "Test Plan Success: Expected changes to be applied")
}

func TestPMRulePlanFailure(t *testing.T) {
	tempConfigDir, err := os.MkdirTemp("", "test-invalid-config-pmr")
	assert.NoError(t, err)
	defer os.RemoveAll(tempConfigDir)
	invalidYAML := `
project_id: "my-project"
rule_name: "failure-rule"
firewall_policy_name: "my-policy"
# priority is missing
direction: "INGRESS"
match:
  layer4_configs:
    - ip_protocol: "all"
`
	err = os.WriteFile(filepath.Join(tempConfigDir, "invalid-rule.yml"), []byte(invalidYAML), 0644)
	assert.NoError(t, err)
	tfVarsFailure := map[string]any{
		"config_folder_path": tempConfigDir,
	}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathPMR,
		Vars:         tfVarsFailure,
		Reconfigure:  true,
		PlanFilePath: "./plan_pmr_failure",
		NoColor:      true,
	})
	t.Cleanup(func() {
		planFilePath := filepath.Join(terraformOptions.TerraformDir, terraformOptions.PlanFilePath)
		err := os.Remove(planFilePath)
		if err != nil {
			t.Logf("Failed to remove plan file '%s', error: %v", planFilePath, err)
		}
	})
	exitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	assert.Equal(t, 1, exitCode, "Expected Terraform to fail with exit code 1 due to invalid config")
}

func TestPMRuleResourcesCount(t *testing.T) {
	files, err := os.ReadDir(configFolderPathPMR)
	assert.NoError(t, err, "Error reading config directory")
	expectedResourceCount := 0
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			expectedResourceCount++
		}
	}
	assert.NotZero(t, expectedResourceCount, "No YAML files found in the test config directory")
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathPMR,
		Vars:         tfVarsPMR,
		Reconfigure:  true,
		PlanFilePath: "./plan_pmr_count",
		NoColor:      true,
	})
	t.Cleanup(func() {
		planFilePath := filepath.Join(terraformOptions.TerraformDir, terraformOptions.PlanFilePath)
		err := os.Remove(planFilePath)
		if err != nil {
			t.Logf("Failed to remove plan file '%s', error: %v", planFilePath, err)
		}
	})
	planStruct := terraform.InitAndPlan(t, terraformOptions)
	resourceCount := terraform.GetResourceCount(t, planStruct)
	assert.Equal(t, expectedResourceCount, resourceCount.Add, "Test Resource Count Add: Unexpected number of rules to be created")
}

func TestPMRuleModuleAddressListMatch(t *testing.T) {
	t.Parallel()
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathPMR,
		Vars:         tfVarsPMR,
		Reconfigure:  true,
		PlanFilePath: "./plan_pmr_match",
		NoColor:      true,
	})
	t.Cleanup(func() {
		planFilePath := filepath.Join(terraformOptions.TerraformDir, terraformOptions.PlanFilePath)
		err := os.Remove(planFilePath)
		if err != nil {
			t.Logf("Failed to remove plan file '%s', error: %v", planFilePath, err)
		}
	})
	configPath, ok := terraformOptions.Vars["config_folder_path"].(string)
	assert.True(t, ok, "config_folder_path not found or not a string in tfVars")
	expectedModuleKeys := []string{}
	files, err := os.ReadDir(configPath)
	assert.NoError(t, err, "Error reading config directory: %s", configPath)
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			yamlFilePath := filepath.Join(configPath, file.Name())
			yamlFile, err := os.ReadFile(yamlFilePath)
			assert.NoError(t, err, "Failed to read YAML file %s", yamlFilePath)

			var content map[string]any
			err = yaml.Unmarshal(yamlFile, &content)
			assert.NoError(t, err, "Failed to unmarshal YAML file %s", yamlFilePath)
			if priority, ok := content["priority"].(int); ok {
				expectedModuleKeys = append(expectedModuleKeys, fmt.Sprintf("%d", priority))
			}
		}
	}
	assert.NotEmpty(t, expectedModuleKeys, "No YAML files found or 'priority' key missing in test YAMLs")
	expectedModuleAddresses := []string{}
	for _, key := range expectedModuleKeys {
		expectedModuleAddresses = append(expectedModuleAddresses, fmt.Sprintf("module.packet_mirroring_rule[\"%s\"]", key))
	}
	planStruct := terraform.InitAndPlanAndShow(t, terraformOptions)
	content, err := terraform.ParsePlanJSON(planStruct)
	assert.NoError(t, err, "Error parsing plan JSON")
	actualModuleAddresses := make([]string, 0)
	for _, element := range content.ResourceChangesMap {
		if strings.HasPrefix(element.ModuleAddress, "module.packet_mirroring_rule") &&
			!slices.Contains(actualModuleAddresses, element.ModuleAddress) {
			actualModuleAddresses = append(actualModuleAddresses, element.ModuleAddress)
		}
	}
	assert.ElementsMatch(t, expectedModuleAddresses, actualModuleAddresses, "The planned module addresses do not match the expected addresses from YAML files.")
}
