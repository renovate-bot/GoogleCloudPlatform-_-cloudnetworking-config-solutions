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
	projectRootPM, _         = filepath.Abs("../../../../../")
	terraformDirectoryPathPM = filepath.Join(projectRootPM, "execution/08-network-security-integration/Out-Of-Band")
	configFolderPathPM       = filepath.Join(projectRootPM, "execution/test/unit/network-security-integration/OutOfBand/config")
)

var (
	tfVarsSuccess = map[string]any{
		"config_folder_path": configFolderPathPM,
	}
)

func TestPacketMirroringPlanSuccess(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathPM,
		Vars:         tfVarsSuccess,
		Reconfigure:  true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})
	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	assert.Equal(t, 2, planExitCode, "Test Plan Success: Expected changes to be applied")
}

func TestPacketMirroringResourcesCount(t *testing.T) {
	expectedAddCount := 7
	expectedChangeCount := 0
	expectedDestroyCount := 0
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathPM,
		Vars:         tfVarsSuccess,
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

func TestPacketMirroringModuleAddressListMatch(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathPM,
		Vars:         tfVarsSuccess,
		Reconfigure:  true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})
	expectedModuleKeys := []string{}
	files, err := os.ReadDir(configFolderPathPM)
	assert.NoError(t, err, "Error reading test config directory")
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
		expectedModuleAddresses = append(expectedModuleAddresses, fmt.Sprintf("module.packet_mirroring[\"%s\"]", key))
	}
	planStruct := terraform.InitAndPlanAndShow(t, terraformOptions)
	content, err := terraform.ParsePlanJSON(planStruct)
	assert.NoError(t, err, "Error parsing plan JSON")
	actualModuleAddresses := make([]string, 0)
	for _, element := range content.ResourceChangesMap {
		if strings.HasPrefix(element.ModuleAddress, "module.packet_mirroring") &&
			!slices.Contains(actualModuleAddresses, element.ModuleAddress) {
			actualModuleAddresses = append(actualModuleAddresses, element.ModuleAddress)
		}
	}
	assert.ElementsMatch(t, expectedModuleAddresses, actualModuleAddresses, "The planned module addresses do not match the expected addresses from YAML files.")
}

func TestPacketMirroringPlanFailure(t *testing.T) {
	tempConfigDir, err := os.MkdirTemp("", "test-invalid-config")
	assert.NoError(t, err)
	defer os.RemoveAll(tempConfigDir)
	invalidFile, err := os.Create(filepath.Join(tempConfigDir, "invalid.yml"))
	assert.NoError(t, err)
	invalidFile.Close()
	tfVarsFailure := map[string]any{
		"config_folder_path": tempConfigDir,
		"project_id":         "test-project",
	}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathPM,
		Vars:         tfVarsFailure,
		Reconfigure:  true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})
	exitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	assert.Equal(t, 1, exitCode, "Expected Terraform to fail with exit code 1 due to invalid YAML file")
}
