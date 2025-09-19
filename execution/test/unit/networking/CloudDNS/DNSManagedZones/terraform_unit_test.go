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
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

var (
	projectRoot, _         = filepath.Abs("../../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "02-networking/CloudDNS/DNSManagedZones")
	configFolderPath       = filepath.Join(projectRoot, "test/unit/networking/CloudDNS/DNSManagedZones/config")
)

var (
	tfVars = map[string]any{
		"config_folder_path": configFolderPath,
	}
)

func TestInitAndPlanRunWithTfVars(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})

	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	want := 2 // Expect changes to be applied
	got := planExitCode

	if got != want {
		t.Errorf("Test Plan Exit Code = %v, want = %v", got, want)
	}
}

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

	// The number of resources will vary based on the number of zones and recordsets in the YAML files.
	// Update this value based on your test configuration.
	// For the provided dns.yaml, we expect:
	// 3 google_dns_managed_zone resources
	// 2 google_dns_record_set resources
	// Total = 5
	if got, want := resourceCount.Add, 5; got != want {
		t.Errorf("Test Resource Count Add = %v, want = %v", got, want)
	}
}

type DNSConfig struct {
	Zones []struct {
		Zone string `yaml:"zone"`
	} `yaml:"zones"`
}

func TestDNSManagedZoneModuleResourceAddressesAndCount(t *testing.T) {
	expectedModuleAddresses := make(map[string]struct{})
	yamlFiles, err := os.ReadDir(configFolderPath)
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, file := range yamlFiles {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			yamlData, err := os.ReadFile(filepath.Join(configFolderPath, file.Name()))
			if err != nil {
				t.Fatal(err.Error())
			}

			var config DNSConfig

			err = yaml.Unmarshal(yamlData, &config)
			if err != nil {
				t.Fatal(err.Error())
			}

			for _, zone := range config.Zones {
				expectedModuleAddresses[fmt.Sprintf("module.dns[\"dns-%s\"]", zone.Zone)] = struct{}{}
			}
		}
	}

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
		t.Fatal(err.Error())
	}

	actualModuleAddresses := make(map[string]struct{})
	for _, element := range content.ResourceChangesMap {
		if strings.HasPrefix(element.ModuleAddress, "module.dns[") {
			actualModuleAddresses[element.ModuleAddress] = struct{}{}
		}
	}

	expectedSlice := make([]string, 0, len(expectedModuleAddresses))
	for address := range expectedModuleAddresses {
		expectedSlice = append(expectedSlice, address)
	}

	actualSlice := make([]string, 0, len(actualModuleAddresses))
	for address := range actualModuleAddresses {
		actualSlice = append(actualSlice, address)
	}

	if len(expectedSlice) > 0 {
		if !assert.ElementsMatch(t, actualSlice, expectedSlice) {
			t.Errorf("Test Element Mismatch = %v, want = %v", actualSlice, expectedSlice)
		}
	} else {
		if len(actualSlice) > 0 {
			t.Errorf("Unexpected module addresses found: %v", actualSlice)
		} else {
			t.Log("No modules expected, and none found in plan.")
		}
	}
}
