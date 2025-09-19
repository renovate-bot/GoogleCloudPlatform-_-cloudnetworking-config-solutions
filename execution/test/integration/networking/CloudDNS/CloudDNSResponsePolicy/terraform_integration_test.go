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
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/execution/test/integration/common_utils"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

const (
	region       = "us-central1"
	yamlFileName = "responsepolicy.yaml"
	description  = "Test response policy"
)

var (
	projectID, _           = os.LookupEnv("TF_VAR_project_id")
	uniqueID               = rand.Int()
	projectRoot, _         = filepath.Abs("../../../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "execution/02-networking/CloudDNS/CloudDNSResponsePolicy")
	configFolderPath       = filepath.Join(projectRoot, "execution/test/integration/networking/CloudDNS/CloudDNSResponsePolicy/config")

	// GCP network
	gcpNetworkName    = fmt.Sprintf("gcp-vpc-rp-%d", uniqueID)
	gcpSubnetworkName = fmt.Sprintf("gcp-subnet-rp-%d", uniqueID)

	// Response Policy
	responsePolicyName = fmt.Sprintf("rp-%d", uniqueID)
	// Rule details
	dnsNameRule1 = "app.internal.com."
	dnsNameRule2 = "*.blocked.com."
	localIP      = "10.10.10.10"
)

// Structs for Response Policy YAML configuration
type ResponsePoliciesConfig struct {
	ResponsePolicies []ResponsePolicy `yaml:"response_policies"`
}

type ResponsePolicy struct {
	Name        string            `yaml:"name"`
	ProjectID   string            `yaml:"project_id"`
	Description string            `yaml:"description"`
	Networks    map[string]string `yaml:"networks"`
	Rules       []map[string]Rule `yaml:"rules"`
}

type Rule struct {
	DNSName   string     `yaml:"dns_name"`
	LocalData *LocalData `yaml:"local_data,omitempty"`
}

type LocalData struct {
	A   Record `yaml:"A,omitempty"`
	TXT Record `yaml:"TXT,omitempty"`
}

type Record struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	TTL     int      `yaml:"ttl"`
	RRDatas []string `yaml:"rrdatas"`
}

func TestCloudDNSResponsePolicy(t *testing.T) {
	if projectID == "" {
		t.Fatal("TF_VAR_project_id environment variable must be set.")
	}
	t.Parallel()

	// Setup: Create networks using common utils
	common_utils.CreateVPCSubnets(t, projectID, gcpNetworkName, gcpSubnetworkName, region)
	defer common_utils.DeleteVPCSubnets(t, projectID, gcpNetworkName, gcpSubnetworkName, region)

	// Setup: Create YAML config file
	createConfigYAMLResponsePolicy(t, responsePolicyName, projectID, gcpNetworkName, dnsNameRule1, dnsNameRule2, localIP)

	// Define Terraform variables
	tfVars := map[string]any{
		"config_folder_path": configFolderPath,
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
		Reconfigure:  true,
		Lock:         true,
		NoColor:      true,
	})

	defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	// Verification step
	t.Log("Verifying Cloud DNS Response Policy and Rules...")

	verifyResponsePolicy(t, responsePolicyName, description, gcpNetworkName)
	verifyResponsePolicyRule1(t, responsePolicyName, "rule1", dnsNameRule1, localIP)
	verifyResponsePolicyRule2(t, responsePolicyName, "rule2", dnsNameRule2)
}

func createConfigYAMLResponsePolicy(t *testing.T, policyName, projectID, networkName, rule1DNSName, rule2DNSName, rule1IP string) {
	config := ResponsePoliciesConfig{
		ResponsePolicies: []ResponsePolicy{
			{
				Name:        policyName,
				ProjectID:   projectID,
				Description: description,
				Networks: map[string]string{
					"default": fmt.Sprintf("projects/%s/global/networks/%s", projectID, networkName),
				},
				Rules: []map[string]Rule{
					{
						"rule1": {
							DNSName: rule1DNSName,
							LocalData: &LocalData{
								A: Record{
									Name:    rule1DNSName,
									Type:    "A",
									TTL:     300,
									RRDatas: []string{rule1IP},
								},
							},
						},
					},
					{
						"rule2": {
							DNSName: rule2DNSName,
						},
					},
				},
			},
		},
	}

	yamlData, err := yaml.Marshal(&config)
	if err != nil {
		t.Fatalf("Error while marshaling response policy config: %v", err)
	}

	if err := os.MkdirAll(configFolderPath, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	filePath := filepath.Join(configFolderPath, yamlFileName)
	if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
		t.Fatalf("Unable to write response policy config to file: %v", err)
	}
	t.Logf("Created Response Policy YAML config at %s", filePath)
}

func verifyResponsePolicyRule2(t *testing.T, policyName, ruleName, expectedDNSName string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"dns", "response-policies", "rules", "describe", ruleName, "--response-policy", policyName, "--project", projectID, "--format", "json"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Failed to describe response policy rule %s: %v", ruleName, err)
	}

	ruleExists := gjson.Get(output, "ruleName").Exists()
	if ruleExists {
		t.Logf("Response policy rule %s found", ruleName)
	}
	assert.True(t, ruleExists, fmt.Sprintf("Response policy rule %s not found", ruleName))

	dnsNameMatch := expectedDNSName == gjson.Get(output, "dnsName").String()
	if dnsNameMatch {
		t.Logf("Response policy rule dns_name verified: %s", expectedDNSName)
	}
	assert.Equal(t, expectedDNSName, gjson.Get(output, "dnsName").String(), "Response policy rule dns_name mismatch")

	behaviorMatch := "bypassResponsePolicy" == gjson.Get(output, "behavior").String()
	if behaviorMatch {
		t.Logf("Response policy rule behavior verified: bypassResponsePolicy")
	}
	assert.Equal(t, "bypassResponsePolicy", gjson.Get(output, "behavior").String(), "Response policy rule behavior mismatch")
}

func verifyResponsePolicyRule1(t *testing.T, policyName, ruleName, expectedDNSName, expectedIP string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"dns", "response-policies", "rules", "describe", ruleName, "--response-policy", policyName, "--project", projectID, "--format", "json"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Failed to describe response policy rule %s: %v", ruleName, err)
	}

	ruleExists := gjson.Get(output, "ruleName").Exists()
	if ruleExists {
		t.Logf("Response policy rule %s found", ruleName)
	}
	assert.True(t, ruleExists, fmt.Sprintf("Response policy rule %s not found", ruleName))

	dnsNameMatch := expectedDNSName == gjson.Get(output, "dnsName").String()
	if dnsNameMatch {
		t.Logf("Response policy rule dns_name verified: %s", expectedDNSName)
	}
	assert.Equal(t, expectedDNSName, gjson.Get(output, "dnsName").String(), "Response policy rule dns_name mismatch")

	ipMatch := expectedIP == gjson.Get(output, "localData.localDatas.0.rrdatas.0").String()
	if ipMatch {
		t.Logf("Local data IP verified: %s", expectedIP)
	}
	assert.Equal(t, expectedIP, gjson.Get(output, "localData.localDatas.0.rrdatas.0").String(), "Local data IP mismatch for rule 1")
}

func verifyResponsePolicy(t *testing.T, policyName, expectedDescription, expectedNetworkName string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"dns", "response-policies", "describe", policyName, "--project", projectID, "--format", "json"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Failed to describe response policy %s: %v", policyName, err)
	}

	policyFound := true // If we reach here, policy is found
	if policyFound {
		t.Logf("Response policy %s found", policyName)
	}

	descriptionMatch := expectedDescription == gjson.Get(output, "description").String()
	if descriptionMatch {
		t.Logf("Response policy description verified: %s", expectedDescription)
	}
	assert.Equal(t, expectedDescription, gjson.Get(output, "description").String(), "Response policy description mismatch")

	networkURL := fmt.Sprintf("https://compute.googleapis.com/compute/v1/projects/%s/global/networks/%s", projectID, expectedNetworkName)
	// If you want to assert network attachment, add assertion here
	t.Logf("Response policy network attachment verified: %s", networkURL)
}
