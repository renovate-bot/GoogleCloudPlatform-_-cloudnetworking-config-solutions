// Copyright 2025 Google LLC
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
package unittest

import (
	compare "cmp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"golang.org/x/exp/slices"
)

const (
	terraformDirectoryPath = "../../../../04-producer/BigQuery"
	configFolderPath       = "../../test/unit/producer/BigQuery/config"
)

var (
	tfVars = map[string]any{
		"config_folder_path": configFolderPath,
	}
	// Used to validate an expected error code if a wrong configuration value is provided.
	invalidTFVars = map[string]any{
		"config_folder_path":  configFolderPath,
		"deletion_protection": "invalidValue", // deletion_protection expects a boolean
	}
)

/*
	TestInitAndPlanRunWithTfVars performs a sanity check to ensure that terraform init

&& terraform plan executes successfully and returns a valid "Succeeded" run code.
*/
func TestInitAndPlanRunWithTfVars(t *testing.T) {
	/*
	 0 = Succeeded with empty diff (no changes)
	 1 = Error
	 2 = Succeeded with non-empty diff (changes present)
	*/
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})
	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	want := 2
	got := planExitCode
	if got != want {
		t.Errorf("test plan exit code = %v, want = %v", got, want)
	}
}

/*
TestInitAndPlanRunWithInvalidTfVarsExpectFailureScenario performs a test run with an invalid
tfvars file to ensure the plan fails and returns an expected error code.
*/
func TestInitAndPlanRunWithInvalidTfVarsExpectFailureScenario(t *testing.T) {
	/*
	 0 = Succeeded with empty diff (no changes)
	 1 = Error
	 2 = Succeeded with non-empty diff (changes present)
	*/
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         invalidTFVars,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})
	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	want := 1
	got := planExitCode
	if !cmp.Equal(got, want) {
		t.Errorf("test plan exit code = %v, want = %v", got, want)
	}
}

/*
	TestResourcesCount validates the number of resources to be created, deleted, and

updated by the Terraform plan.
*/
func TestResourcesCount(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})
	planStruct := terraform.InitAndPlan(t, terraformOptions)
	resourceCount := terraform.GetResourceCount(t, planStruct)
	// Expects 3 resources to be added, corresponding to the 3 dummy YAML files.
	if got, want := resourceCount.Add, 3; got != want {
		t.Errorf("test resource count add = %v, want = %v", got, want)
	}
	if got, want := resourceCount.Change, 0; got != want {
		t.Errorf("test resource count change = %v, want = %v", got, want)
	}
	if got, want := resourceCount.Destroy, 0; got != want {
		t.Errorf("test resource count destroy = %v, want = %v", got, want)
	}
}

/*
	TestTerraformModuleResourceAddressListMatch compares and verifies the list of module

addresses created by the Terraform solution.
*/
func TestTerraformModuleResourceAddressListMatch(t *testing.T) {
	expectedModulesAddress := []string{"module.bigquery[\"bq_dummy1\"]", "module.bigquery[\"bq_dummy2\"]", "module.bigquery[\"bq_dummy3\"]"}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})
	planStruct := terraform.InitAndPlanAndShow(t, terraformOptions)
	content, err := terraform.ParsePlanJSON(planStruct)
	if err != nil {
		t.Error(err.Error())
	}
	actualModuleAddress := make([]string, 0)
	for _, element := range content.ResourceChangesMap {
		if element.ModuleAddress != "" && !slices.Contains(actualModuleAddress, element.ModuleAddress) {
			actualModuleAddress = append(actualModuleAddress, element.ModuleAddress)
		}
	}
	want := expectedModulesAddress
	got := actualModuleAddress
	if !cmp.Equal(got, want, cmpopts.SortSlices(compare.Less[string])) {
		t.Errorf("test element mismatch = %v, want = %v", got, want)
	}
}
