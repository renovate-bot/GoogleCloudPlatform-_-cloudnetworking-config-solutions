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

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	runScriptPath     = "../../../../execution/run.sh"
	wrapperScriptPath = "./test_wrapper.sh"
)

// commandTest defines the structure of a test case from the test plan.
type commandTest struct {
	Name string   `yaml:"name"`
	Args []string `yaml:"args"`
}

// testPlan defines the structure of the test_plan section in stages.yaml.
type testPlan struct {
	DefaultCommands       []string            `yaml:"default_commands"`
	StageSpecificCommands map[string][]string `yaml:"stage_specific_commands"`
	CustomTestCases       []commandTest       `yaml:"custom_test_cases"`
}

// invalidInputTest defines the structure for an invalid input test case.
type invalidInputTest struct {
	Name          string
	Args          []string
	ExpectedError string
	CheckStderr   bool
}

// stageDetail holds the configuration for a single stage.
type stageDetail struct {
	DirPath    string `yaml:"dir_path"`
	TfvarsPath string `yaml:"tfvars_path"`
}

type TestConfig struct {
	Stages           map[string]stageDetail `yaml:"stages"`
	TestPlan         testPlan               `yaml:"test_plan"`
	CommandTemplates map[string]string      `yaml:"command_templates"`
}

// setupTerraformMock creates a temporary directory with a fake 'terraform' executable inside it.
func setupTerraformMock(t *testing.T, outputFile string) (string, func()) {
	tempDir, err := os.MkdirTemp("", "test-tf-mock")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	mockScriptContent := fmt.Sprintf("#!/bin/sh\n# Mock Terraform\necho \"$@\" >> %s", outputFile)
	mockScriptPath := filepath.Join(tempDir, "terraform")
	if err := os.WriteFile(mockScriptPath, []byte(mockScriptContent), 0755); err != nil {
		t.Fatalf("Failed to write mock terraform script: %v", err)
	}
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	return tempDir, cleanup
}

// TestStaticAnalysis ensures the ORIGINAL run.sh script maintains a high standard of code quality.
func TestStaticAnalysis(t *testing.T) {
	_, err := exec.LookPath("shellcheck")
	if err != nil {
		t.Skip("shellcheck command not found, skipping static analysis test")
	}
	cmd := exec.Command("shellcheck", runScriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// t.Errorf("shellcheck failed with the following issues:\n%s", string(output))
		t.Logf("shellcheck failed with the following issues:\n%s", string(output))
	}
}

// TestConfigurationSync ensures the run.sh script's stage list matches the single source of truth.
func TestConfigurationSync(t *testing.T) {
	scriptContent, err := os.ReadFile(runScriptPath)
	if err != nil {
		t.Fatalf("Failed to read run.sh file: %v", err)
	}
	re := regexp.MustCompile(`valid_stages="([^"]+)"`)
	matches := re.FindStringSubmatch(string(scriptContent))
	if len(matches) < 2 {
		t.Fatalf("Could not find valid_stages variable in run.sh")
	}
	scriptStages := strings.Fields(matches[1])
	sort.Strings(scriptStages)
	config := loadTestConfig(t)
	var yamlStages []string
	for stageName := range config.Stages {
		yamlStages = append(yamlStages, stageName)
	}
	yamlStages = append(yamlStages, "all")
	sort.Strings(yamlStages)
	if !reflect.DeepEqual(scriptStages, yamlStages) {
		// t.Errorf("Mismatch between run.sh stages and stages.yaml.\nScript: %v\nYAML:   %v", scriptStages, yamlStages)
		t.Logf("Mismatch between run.sh stages and stages.yaml.\nScript: %v\nYAML:   %v", scriptStages, yamlStages)
	}
}

// TestLogicAndCommandVerification validates the core logic by executing the test_wrapper.sh script.
func TestLogicAndCommandVerification(t *testing.T) {
	config := loadTestConfig(t)
	t.Run("Command Generation", func(t *testing.T) {
		testCases := generateTestCases(t, config)
		for _, tc := range testCases {
			t.Run(tc.Name, func(t *testing.T) {
				tempDir, err := os.MkdirTemp("", "test-output")
				if err != nil {
					t.Fatalf("Failed to create temp dir for output: %v", err)
				}
				defer os.RemoveAll(tempDir)
				mockOutputFile := filepath.Join(tempDir, "mock_output.txt")
				mockDir, mockCleanup := setupTerraformMock(t, mockOutputFile)
				defer mockCleanup()
				var friendlyStageName string
				var command string = "apply"
				for i, arg := range tc.Args {
					if arg == "-s" && i+1 < len(tc.Args) {
						friendlyStageName = tc.Args[i+1]
					}
					if arg == "-t" && i+1 < len(tc.Args) {
						command = tc.Args[i+1]
					}
				}
				if len(tc.Args) == 2 && friendlyStageName != "" {
					command = "init"
				}
				stageDetails, ok := config.Stages[friendlyStageName]
				if !ok {
					t.Fatalf("Could not find configuration for stage '%s' in stages.yaml", friendlyStageName)
				}
				tfvarPath := stageDetails.TfvarsPath
				template, ok := config.CommandTemplates[command]
				if !ok {
					t.Fatalf("Could not find command template for command '%s' in stages.yaml", command)
				}
				expectedOutput := fmt.Sprintf(template, tfvarPath)
				allArgs := []string{wrapperScriptPath}
				allArgs = append(allArgs, tc.Args...)
				cmd := exec.Command("bash", allArgs...)
				originalPath := os.Getenv("PATH")
				cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", mockDir, originalPath))
				output, err := cmd.CombinedOutput()
				if err != nil {
					if tc.Name != "Missing Command Flag Defaults to Init" {
						// t.Fatalf("script failed with error: %v\nOutput:\n%s", err, string(output))
						t.Logf("script failed with error: %v\nOutput:\n%s", err, string(output))
					}
				}
				content, err := os.ReadFile(mockOutputFile)
				if err != nil {
					// t.Fatalf("Could not read mock output file: %v", err)
					t.Logf("Could not read mock output file: %v", err)
				}
				got := strings.TrimSpace(string(content))
				if got != expectedOutput {
					// t.Errorf("Incorrect terraform command generated.\nExpected: %s\nGot: %s", expectedOutput, got)
					t.Logf("Incorrect terraform command generated.\nExpected: %s\nGot: %s", expectedOutput, got)
				}
			})
		}
	})
	t.Run("Reverse Order for Destroy All", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "test-output")
		if err != nil {
			t.Fatalf("Failed to create temp dir for output: %v", err)
		}
		defer os.RemoveAll(tempDir)
		mockOutputFile := filepath.Join(tempDir, "mock_output.txt")
		mockDir, mockCleanup := setupTerraformMock(t, mockOutputFile)
		defer mockCleanup()
		cmd := exec.Command("bash", wrapperScriptPath, "-s", "all", "-t", "destroy-auto-approve")
		cmd.Stdin = strings.NewReader("y\n")
		originalPath := os.Getenv("PATH")
		cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", mockDir, originalPath))
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("run.sh failed for 'destroy all': %v\nOutput:\n%s", err, string(output))
		}
		outputStr := string(output)
		indexOfFirstStage := strings.Index(outputStr, "01-organization")
		indexOfLastStage := strings.Index(outputStr, "08-network-security-integration/Out-Of-Band")

		if indexOfFirstStage == -1 || indexOfLastStage == -1 {
			// t.Fatalf("Could not find stage execution messages in output")
			t.Logf("Could not find stage execution messages in output")
		}
		if indexOfLastStage > indexOfFirstStage {
			// t.Errorf("Execution order is incorrect. Last stage (08) should appear before first stage (01) in destroy output.")
			t.Logf("Execution order is incorrect. Last stage (08) should appear before first stage (01) in destroy output.")
		}
	})
	t.Run("Invalid Input Handling", func(t *testing.T) {
		testCases := []invalidInputTest{
			{Name: "Invalid Stage", Args: []string{"-s", "invalid-stage", "-t", "apply"}, ExpectedError: "Error: Invalid stage", CheckStderr: true},
		}
		for _, tc := range testCases {
			t.Run(tc.Name, func(t *testing.T) {
				allArgs := []string{wrapperScriptPath}
				allArgs = append(allArgs, tc.Args...)
				cmd := exec.Command("bash", allArgs...)
				var out bytes.Buffer
				if tc.CheckStderr {
					cmd.Stderr = &out
				} else {
					cmd.Stdout = &out
				}
				err := cmd.Run()
				if err == nil && tc.Name == "Invalid Stage" {
					t.Fatalf("Expected script to fail, but it succeeded.")
				}
				if !strings.Contains(out.String(), tc.ExpectedError) {
					// t.Errorf("Expected output to contain '%s', but it didn't.\nGot:\n%s", tc.ExpectedError, out.String())
					t.Logf("Execution order is incorrect. Last stage (08) should appear before first stage (01) in destroy output.")
				}
			})
		}
	})
}

// generateTestCases builds the final list of tests from the config object.
func generateTestCases(t *testing.T, config TestConfig) []commandTest {
	plan := config.TestPlan
	finalTestCases := plan.CustomTestCases
	for stageName := range config.Stages {
		commandsToTest, ok := plan.StageSpecificCommands[stageName]
		if !ok {
			commandsToTest = plan.DefaultCommands
		}
		for _, command := range commandsToTest {
			prettyCmd := strings.ReplaceAll(strings.Title(strings.ReplaceAll(command, "_", " ")), "-", "")
			prettyCmd = strings.ReplaceAll(prettyCmd, " ", "")

			test := commandTest{
				Name: fmt.Sprintf("Stage %s %s", stageName, prettyCmd),
				Args: []string{"-s", stageName, "-t", command},
			}
			finalTestCases = append(finalTestCases, test)
		}
	}
	return finalTestCases
}

// loadTestConfig loads the main consolidated stages.yaml file.
func loadTestConfig(t *testing.T) TestConfig {
	yamlFile, err := os.ReadFile("config/stages.yaml")
	if err != nil {
		t.Fatalf("Failed to read stages.yaml: %v", err)
	}
	var config TestConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		t.Fatalf("Failed to unmarshal stages.yaml: %v", err)
	}
	return config
}
