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

package unittest

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

var (
	projectRootSP, _         = filepath.Abs("../../../../")
	terraformDirectoryPathSP = filepath.Join(projectRootSP, "03-security/SecurityProfile")
	configFolderPathSP       = filepath.Join(projectRootSP, "test/unit/security/SecurityProfile/config")
)
var (
	tfVars = map[string]any{
		"config_folder_path": configFolderPathSP,
	}
)

// TestSecurityProfilePlanExitCode verifies that the plan exits with a code of 2, indicating changes are planned.
func TestSecurityProfilePlanExitCode(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathSP,
		Vars:         tfVars,
		Reconfigure:  true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})

	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	assert.Equal(t, 2, planExitCode, "Test Plan Exit Code: Expected changes to be applied")
}

// TestSecurityProfileResourcesCount verifies the number of resources to be added by the plan.
func TestSecurityProfileResourcesCount(t *testing.T) {
	expectedAddCount := 3
	expectedChangeCount := 0
	expectedDestroyCount := 0
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathSP,
		Vars:         tfVars,
		Reconfigure:  true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})
	t.Log("Initializing and planning Terraform module...")
	planStruct := terraform.InitAndPlan(t, terraformOptions)
	resourceCount := terraform.GetResourceCount(t, planStruct)
	t.Logf("Plan output: %d to add, %d to change, %d to destroy.", resourceCount.Add, resourceCount.Change, resourceCount.Destroy)
	assert.Equal(t, expectedAddCount, resourceCount.Add, "The number of resources to ADD does not match the expected value.")
	assert.Equal(t, expectedChangeCount, resourceCount.Change, "The number of resources to CHANGE does not match the expected value.")
	assert.Equal(t, expectedDestroyCount, resourceCount.Destroy, "The number of resources to DESTROY does not match the expected value.")
}

// TestSecurityProfileModuleAddressListMatch verifies that a module instance is planned for each YAML config file.
func TestSecurityProfileModuleAddressListMatch(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathSP,
		Vars:         tfVars,
		Reconfigure:  true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})

	expectedModuleKeys := []string{}
	files, err := os.ReadDir(configFolderPathSP)
	assert.NoError(t, err, "Error reading config directory")

	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
				key := strings.TrimSuffix(filename, ".yaml")
				key = strings.TrimSuffix(key, ".yml")
				expectedModuleKeys = append(expectedModuleKeys, key)
			}
		}
	}
	assert.NotEmpty(t, expectedModuleKeys, "No YAML files found in the test config directory")
	expectedModuleAddresses := []string{}
	for _, key := range expectedModuleKeys {
		expectedModuleAddresses = append(expectedModuleAddresses, fmt.Sprintf("module.security_profile[\"%s\"]", key))
	}
	planStruct := terraform.InitAndPlanAndShow(t, terraformOptions)
	content, err := terraform.ParsePlanJSON(planStruct)
	assert.NoError(t, err, "Error parsing plan JSON")

	actualModuleAddresses := make([]string, 0)
	for _, element := range content.ResourceChangesMap {
		if strings.HasPrefix(element.ModuleAddress, "module.security_profile") &&
			!slices.Contains(actualModuleAddresses, element.ModuleAddress) {
			actualModuleAddresses = append(actualModuleAddresses, element.ModuleAddress)
		}
	}

	assert.ElementsMatch(t, expectedModuleAddresses, actualModuleAddresses, "The planned module addresses do not match the expected addresses from YAML files.")
}
