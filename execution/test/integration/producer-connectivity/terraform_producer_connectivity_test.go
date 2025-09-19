// Copyright 2024-2025 Google LLC
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

package integrationtest

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Producer defines the configuration for a specific producer type (e.g., CloudSQL, AlloyDB).
type Producer struct {
	Name                  string
	GetCreateArgs         func(name, projectID, allowedProjects, region, networkName string) []string
	GetDeleteArgs         func(name, projectID, region string) []string
	GetDescribeAttachArgs func(name, projectID, region string) []string
	GetTerraformBlock     func(name string) map[string]interface{}
	GetReadyState         func(t *testing.T, name, projectID, region string) (string, error)
	ExpectedReadyState    string
	TerraformProducerKey  string
}

// producersToTest is the list of all producers to be tested.
// To add a new producer, add a new entry here.
var producersToTest = map[string]Producer{
	"cloudsql": {
		Name:                 "cloudsql",
		TerraformProducerKey: "producer_cloudsql",
		ExpectedReadyState:   "RUNNABLE",
		GetCreateArgs: func(name, projectID, allowedProjects, region, networkName string) []string {
			return []string{"sql", "instances", "create", name,
				"--project=" + projectID, "--database-version=MYSQL_8_0", "--region=" + region,
				"--enable-private-service-connect", "--allowed-psc-projects=" + allowedProjects,
				"--no-assign-ip", "--availability-type=REGIONAL", "--tier=db-n1-standard-1",
				"--enable-bin-log", "--async",
			}
		},
		GetDeleteArgs: func(name, projectID, region string) []string {
			return []string{"sql", "instances", "delete", name, "--project=" + projectID, "--quiet"}
		},
		GetDescribeAttachArgs: func(name, projectID, region string) []string {
			return []string{"sql", "instances", "describe", name, "--project=" + projectID, "--format=value(pscServiceAttachmentLink)"}
		},
		GetTerraformBlock: func(name string) map[string]interface{} {
			// The terraform module expects both 'instance_name' and 'cluster_id'.
			return map[string]interface{}{
				"instance_name": name,
				"cluster_id":    name,
			}
		},
		GetReadyState: func(t *testing.T, name, projectID, region string) (string, error) {
			return runGcloudCommandWithOutput(t, "sql", "instances", "describe", name, "--project="+projectID, "--format=value(state)")
		},
	},
}

// Constants for the Terraform directory path.
const (
	terraformDirectoryPath = "../../../05-producer-connectivity/"
)

// Global variables for test configuration.
var (
	ipAddressLiteral           = "10.10.10.30"
	ipAddressLiteralWithTarget = "10.10.10.31"
	region                     = "us-central1"
)

// runGcloudCommand executes a gcloud command and streams its output for logging.
func runGcloudCommand(_ *testing.T, args ...string) error {
	command := "gcloud"
	commandArgs := args
	if len(args) > 0 && args[0] == "bash" {
		command = args[0]
		commandArgs = args[1:]
	}

	log.Printf("Running command: %s %s", command, strings.Join(commandArgs, " "))

	var output []byte
	var err error
	for i := 0; i < 5; i++ {
		cmd := exec.Command(command, commandArgs...)
		output, err = cmd.CombinedOutput()
		if err == nil {
			log.Printf("Output:\n%s", string(output))
			return nil
		}
		log.Printf("Command failed, retrying in 5 seconds... Output:\n%s", string(output))
		time.Sleep(5 * time.Second)
	}
	log.Printf("Final output after retries:\n%s", string(output))
	return err
}

// runGcloudCommandWithOutput executes a gcloud command and returns its output as a string.
func runGcloudCommandWithOutput(_ *testing.T, args ...string) (string, error) {
	log.Printf("Running gcloud command for output: gcloud %s", strings.Join(args, " "))
	cmd := exec.Command("gcloud", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("gcloud command failed. Output:\n%s", string(output))
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// setupNetwork creates a custom VPC and Subnet in the endpoint project.
func setupNetwork(t *testing.T, projectID string, uniqueID int) (string, string, func()) {
	networkName := fmt.Sprintf("test-vpc-%d", uniqueID)
	subnetworkName := fmt.Sprintf("test-subnet-%d", uniqueID)

	log.Printf("Creating custom VPC network: %s", networkName)
	err := runGcloudCommand(t, "compute", "networks", "create", networkName,
		"--project="+projectID,
		"--subnet-mode=custom",
		"--mtu=1460",
		"--bgp-routing-mode=regional",
	)
	require.NoError(t, err, "Failed to create custom VPC")

	log.Printf("Creating custom subnetwork: %s", subnetworkName)
	err = runGcloudCommand(t, "compute", "networks", "subnets", "create", subnetworkName,
		"--network="+networkName,
		"--range=10.10.10.0/24",
		"--region="+region,
		"--project="+projectID,
	)
	require.NoError(t, err, "Failed to create custom subnetwork")
	cleanupFunc := func() {
		log.Printf("Cleaning up network resources: %s, %s", subnetworkName, networkName)
		runGcloudCommand(t, "compute", "networks", "subnets", "delete", subnetworkName, "--region="+region, "--project="+projectID, "--quiet")
		runGcloudCommand(t, "compute", "networks", "delete", networkName, "--project="+projectID, "--quiet")
	}
	return networkName, subnetworkName, cleanupFunc
}

// waitForProducer polls the status of a producer instance until it reaches the expected ready state.
func waitForProducer(t *testing.T, producer Producer, projectID, instanceName, region string) {
	log.Printf("Waiting for %s instance %s to be %s...", producer.Name, instanceName, producer.ExpectedReadyState)
	for i := 0; i < 30; i++ { // Wait for up to 30 minutes.
		status, err := producer.GetReadyState(t, instanceName, projectID, region)
		if err == nil && status == producer.ExpectedReadyState {
			log.Printf("%s instance %s is %s.", producer.Name, instanceName, producer.ExpectedReadyState)
			return
		}
		log.Printf("Instance %s not ready yet (current state: %s), waiting 60 seconds...", instanceName, status)
		time.Sleep(60 * time.Second)
	}
	t.Fatalf("Timed out waiting for %s instance %s to become %s.", producer.Name, instanceName, producer.ExpectedReadyState)
}

// getEndpointProjectID retrieves the mandatory endpoint project ID from an environment variable.
func getEndpointProjectID(t *testing.T) string {
	projectID := os.Getenv("TF_VAR_endpoint_project_id")
	require.NotEmpty(t, projectID, "Environment variable 'TF_VAR_endpoint_project_id' must be set")
	return projectID
}

// getProducerProjectID retrieves the optional producer project ID, defaulting to the endpoint project ID.
func getProducerProjectID(_ *testing.T, endpointProjectID string) string {
	producerProjectID := os.Getenv("TF_VAR_producer_project_id")
	if producerProjectID == "" {
		log.Printf("TF_VAR_producer_project_id not set, defaulting to endpoint_project_id: %s", endpointProjectID)
		return endpointProjectID
	}
	log.Printf("Using producer project ID from environment variable: %s", producerProjectID)
	return producerProjectID
}

// TestPlanFailsWithoutVars tests that the Terraform plan fails when required input variables are missing.
func TestPlanFailsWithoutVars(t *testing.T) {
	t.Parallel()
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath, Reconfigure: true, Lock: true, NoColor: true,
	})
	_, err := terraform.InitAndPlanE(t, terraformOptions)
	assert.Error(t, err, "Expected Terraform plan to fail due to missing variables")
}

// TestProducerConnectivity is the main test function that orchestrates all test cases.
func TestProducerConnectivity(t *testing.T) {
	t.Parallel()
	endpointProjectID := getEndpointProjectID(t)
	producerProjectID := getProducerProjectID(t, endpointProjectID)

	for producerName, producer := range producersToTest {
		// Capture range variables for parallel testing
		producer := producer
		producerName := producerName
		t.Run(producerName, func(t *testing.T) {
			t.Parallel()

			// ONE-TIME SETUP: Create producer instance and network once per producer type.
			rand.Seed(time.Now().UnixNano())
			uniqueID := rand.Intn(10000)
			dynamicInstanceName := fmt.Sprintf("test-%s-psc-%d", producer.Name, uniqueID)

			networkName, subnetworkName, cleanupNetwork := setupNetwork(t, endpointProjectID, uniqueID)
			defer cleanupNetwork()

			createArgs := producer.GetCreateArgs(dynamicInstanceName, producerProjectID, endpointProjectID, region, networkName)
			// 'err' is declared for the first time here.
			err := runGcloudCommand(t, createArgs...)
			require.NoError(t, err, "Failed to start producer instance creation")
			defer func() {
				log.Printf("Destroying %s instance: %s", producer.Name, dynamicInstanceName)
				deleteArgs := producer.GetDeleteArgs(dynamicInstanceName, producerProjectID, region)
				deleteErr := runGcloudCommand(t, deleteArgs...)
				assert.NoError(t, deleteErr, "Failed to destroy producer instance")
			}()

			waitForProducer(t, producer, producerProjectID, dynamicInstanceName, region)

			var serviceAttachment string
			// The existing 'err' variable from above will be reused.
			for i := 0; i < 4; i++ {
				describeArgs := producer.GetDescribeAttachArgs(dynamicInstanceName, producerProjectID, region)
				serviceAttachment, err = runGcloudCommandWithOutput(t, describeArgs...)
				if err == nil && serviceAttachment != "" {
					break // Success
				}
				log.Printf("Service attachment link not yet available for %s, retrying in 15 seconds...", dynamicInstanceName)
				time.Sleep(15 * time.Second)
			}
			require.NoError(t, err, "Failed to get service attachment link after retries")
			require.NotEmpty(t, serviceAttachment, "Service attachment link was empty after retries")

			// === RUN TEST VARIATIONS AGAINST THE CREATED PRODUCER ===
			// These sub-tests run SEQUENTIALLY to avoid race conditions.

			// Test Case 1: With a provided IP address
			t.Run("WithProvidedIPAddress", func(t *testing.T) {
				tfVars := map[string]interface{}{
					"psc_endpoints": []map[string]interface{}{{
						"endpoint_project_id":          endpointProjectID,
						"producer_instance_project_id": producerProjectID,
						"subnetwork_name":              subnetworkName,
						"network_name":                 networkName,
						"ip_address_literal":           ipAddressLiteral,
						"region":                       region,
						producer.TerraformProducerKey:  producer.GetTerraformBlock(dynamicInstanceName),
					}},
				}
				tfOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{TerraformDir: terraformDirectoryPath, Vars: tfVars})
				defer terraform.Destroy(t, tfOptions)
				terraform.InitAndApply(t, tfOptions)
				assertOutputs(t, tfOptions, producer.TerraformProducerKey)
			})

			// Test Case 2: With an auto-allocated IP address
			t.Run("WithAutoAllocatedIPAddress", func(t *testing.T) {
				tfVars := map[string]interface{}{
					"psc_endpoints": []map[string]interface{}{{
						"endpoint_project_id":          endpointProjectID,
						"producer_instance_project_id": producerProjectID,
						"subnetwork_name":              subnetworkName,
						"network_name":                 networkName,
						"ip_address_literal":           "", // Key change for this test
						"region":                       region,
						producer.TerraformProducerKey:  producer.GetTerraformBlock(dynamicInstanceName),
					}},
				}
				tfOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{TerraformDir: terraformDirectoryPath, Vars: tfVars})
				defer terraform.Destroy(t, tfOptions)
				terraform.InitAndApply(t, tfOptions)
				assertOutputsForAutoAllocatedIPAddress(t, tfOptions, producer.TerraformProducerKey)
			})

			// Test Case 3: With a direct service attachment target
			t.Run("WithDirectTarget", func(t *testing.T) {
				tfVars := map[string]interface{}{
					"psc_endpoints": []map[string]interface{}{{
						"endpoint_project_id":          endpointProjectID,
						"producer_instance_project_id": producerProjectID,
						"subnetwork_name":              subnetworkName,
						"network_name":                 networkName,
						"ip_address_literal":           ipAddressLiteralWithTarget,
						"region":                       region,
						"target":                       serviceAttachment, // Use the pre-fetched target.
					}},
				}
				tfOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{TerraformDir: terraformDirectoryPath, Vars: tfVars})
				defer terraform.Destroy(t, tfOptions)
				terraform.InitAndApply(t, tfOptions)
				assertOutputsWithTarget(t, tfOptions, serviceAttachment)
			})
		})
	}
}

// ============== ASSERTION HELPERS ==============

func assertOutputs(t *testing.T, tfOptions *terraform.Options, producerKey string) {
	actualFwdRuleMap := terraform.OutputMap(t, tfOptions, "forwarding_rule_self_link")
	actualIPMap := terraform.OutputMap(t, tfOptions, "ip_address_literal")
	vars := tfOptions.Vars["psc_endpoints"].([]map[string]interface{})[0]
	producerBlock := vars[producerKey].(map[string]interface{})
	var instanceName string
	// Get the instance/cluster name from the producer block
	for _, v := range producerBlock {
		instanceName = v.(string)
	}
	expectedFwdRuleName := fmt.Sprintf("psc-forwarding-rule-%s", instanceName)
	actualFwdRuleSelfLink := actualFwdRuleMap["0"]
	parts := strings.Split(actualFwdRuleSelfLink, "/")
	actualFwdRuleName := parts[len(parts)-1]
	actualIPAddress := actualIPMap["0"]
	assert.Equal(t, expectedFwdRuleName, actualFwdRuleName, "Forwarding rule name mismatch")
	assert.NotNil(t, actualIPAddress, "IP address is nil")
}

func assertOutputsForAutoAllocatedIPAddress(t *testing.T, tfOptions *terraform.Options, producerKey string) {
	actualFwdRuleMap := terraform.OutputMap(t, tfOptions, "forwarding_rule_self_link")
	actualIPMap := terraform.OutputMap(t, tfOptions, "ip_address_literal")
	vars := tfOptions.Vars["psc_endpoints"].([]map[string]interface{})[0]
	producerBlock := vars[producerKey].(map[string]interface{})
	var instanceName string
	for _, v := range producerBlock {
		instanceName = v.(string)
	}
	expectedFwdRuleName := fmt.Sprintf("psc-forwarding-rule-%s", instanceName)
	actualFwdRuleSelfLink := actualFwdRuleMap["0"]
	parts := strings.Split(actualFwdRuleSelfLink, "/")
	actualFwdRuleName := parts[len(parts)-1]
	actualIPAddress := actualIPMap["0"]
	assert.Equal(t, expectedFwdRuleName, actualFwdRuleName, "Forwarding rule name mismatch")
	assert.NotNil(t, actualIPAddress, "IP address should be auto-allocated and not nil")
}

func assertOutputsWithTarget(t *testing.T, tfOptions *terraform.Options, expectedTarget string) {
	actualFwdRuleMap := terraform.OutputMap(t, tfOptions, "forwarding_rule_self_link")
	actualIPMap := terraform.OutputMap(t, tfOptions, "ip_address_literal")
	actualTargetMap := terraform.OutputMap(t, tfOptions, "forwarding_rule_target")
	expectedFwdRuleName := "psc-forwarding-rule-custom-0" // As per module logic for custom targets
	actualFwdRuleSelfLink := actualFwdRuleMap["0"]
	parts := strings.Split(actualFwdRuleSelfLink, "/")
	actualFwdRuleName := parts[len(parts)-1]
	actualIPAddress := actualIPMap["0"]
	actualTarget := actualTargetMap["0"]
	assert.Equal(t, expectedFwdRuleName, actualFwdRuleName, "Forwarding rule name mismatch")
	assert.NotNil(t, actualIPAddress, "IP address is nil")
	assert.Equal(t, expectedTarget, actualTarget, "Target mismatch")
}
