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

package integrationtest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var (
	projectRoot, _         = filepath.Abs("../../../..")
	terraformDirectoryPath = filepath.Join(projectRoot, "08-network-security-integration/Out-Of-Band")
	configFolderPath       = filepath.Join(projectRoot, "test/integration/network-security-integration/OutOfBand/config")
	producerVPCRange       = "10.10.10.0/24"
	consumerVPCRange       = "10.20.10.0/24"
	hcPort                 = "8080"
	backendPort            = "6081"
	listenerPort           = "50051"
	gcsBucketLocation      = "US-CENTRAL1"
	location               = "us-central1"
	zoneA                  = "us-central1-a"
	zoneB                  = "us-central1-b"
)

func TestPacketMirroringIntegration(t *testing.T) {
	projectID := os.Getenv("TF_VAR_project_id")
	if projectID == "" {
		t.Fatal("Skipping test: environment variable TF_VAR_project_id is not set.")
	}
	orgID := getOrgIDFromProject(t, projectID)
	instanceSuffix := strings.ToLower(random.UniqueId())
	gcsBucketName := fmt.Sprintf("%s-pm-test-logs-%s", strings.ReplaceAll(projectID, "google.com:", ""), instanceSuffix)
	gcsLogPath := fmt.Sprintf("gs://%s/pm-logs-%s/", gcsBucketName, instanceSuffix)
	producerVPC := "producer-vpc-" + instanceSuffix
	consumerVPC := "consumer-vpc-" + instanceSuffix
	consumerFwPolicyName := "fwp-consumer-pm-test-" + instanceSuffix
	securityProfileName := "sp-pm-test-" + instanceSuffix
	securityProfileGroupName := "spg-pm-test-" + instanceSuffix
	collectorImage := fmt.Sprintf("gcr.io/%s/terminus-lite:v1", projectID)
	t.Logf("Test Run Config: ProjectID=%s, OrgID=%s, Suffix=%s", projectID, orgID, instanceSuffix)
	createGCSBucket(t, projectID, gcsBucketName, gcsBucketLocation)
	defer deleteGCSBucket(t, gcsBucketName)
	createVPC(t, projectID, producerVPC, producerVPCRange, consumerVPCRange, location)
	defer deleteVPC(t, projectID, producerVPC, location)
	createVPC(t, projectID, consumerVPC, consumerVPCRange, producerVPCRange, location)
	defer deleteVPC(t, projectID, consumerVPC, location)
	createCollectorStack(t, projectID, producerVPC, zoneA, zoneB, hcPort, backendPort, gcsLogPath, collectorImage)
	defer deleteCollectorStack(t, projectID, zoneA, zoneB, location, instanceSuffix)
	createWorkloadVMs(t, projectID, producerVPC, consumerVPC, zoneA, listenerPort, instanceSuffix)
	defer deleteWorkloadVMs(t, projectID, zoneA, instanceSuffix)
	createFirewallPolicy(t, projectID, consumerFwPolicyName)
	defer deleteFirewallPolicy(t, projectID, consumerFwPolicyName)
	associateFirewallPolicy(t, projectID, consumerVPC, consumerFwPolicyName)
	defer deleteFirewallPolicyAssociation(t, projectID, consumerVPC, consumerFwPolicyName)
	t.Log("Applying Packet Mirroring Terraform configuration...")
	createTestYAML(t, projectID, instanceSuffix)
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         map[string]any{"config_folder_path": configFolderPath},
		Reconfigure:  true,
		NoColor:      true,
	})
	defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)
	t.Log("Packet Mirroring apply complete.")
	t.Log("Validation: Verifying creation of packet mirroring resources from Terraform outputs...")
	tfOutputs := terraform.OutputAll(t, terraformOptions)
	validateMapOutput := func(outputName string) map[string]interface{} {
		rawOutput, ok := tfOutputs[outputName]
		if !ok {
			t.Fatalf("'%s' output not found in terraform outputs", outputName)
		}
		mapOutput, ok := rawOutput.(map[string]interface{})
		if !ok {
			t.Fatalf("'%s' output is not of the expected type (map)", outputName)
		}
		if len(mapOutput) == 0 {
			t.Fatalf("'%s' map should not be empty", outputName)
		}
		t.Logf("Validation successful: '%s' output found and is not empty.", outputName)
		return mapOutput
	}
	validateMapOutput("deployment_groups")
	validateMapOutput("deployments")
	validateMapOutput("endpoint_associations")
	allEndpointGroups := validateMapOutput("endpoint_groups")
	instanceDataRaw, ok := allEndpointGroups["instance"]
	if !ok {
		t.Fatal("Key 'instance' not found in 'endpoint_groups' output map")
	}
	instanceData, ok := instanceDataRaw.(map[string]interface{})
	if !ok {
		t.Fatal("Instance data is not of the expected type (map)")
	}
	endpointGroupID, ok := instanceData["id"].(string)
	if !ok {
		t.Fatal("'id' key not found or not a string in instance data")
	}
	if endpointGroupID == "" {
		t.Fatal("Endpoint Group ID is empty")
	}
	t.Logf("Validation successful: Extracted Endpoint Group ID: %s", endpointGroupID)
	createSecurityProfile(t, orgID, projectID, securityProfileName, securityProfileGroupName, endpointGroupID)
	defer deleteSecurityProfile(t, orgID, securityProfileName, securityProfileGroupName)
	verifySecurityProfiles(t, orgID, securityProfileName, securityProfileGroupName)
	createFirewallMirroringRule(t, orgID, projectID, consumerFwPolicyName, securityProfileGroupName)
	defer deleteFirewallMirroringRule(t, projectID, consumerFwPolicyName)
	expectedSpgPath := fmt.Sprintf("organizations/%s/locations/global/securityProfileGroups/%s", orgID, securityProfileGroupName)
	verifyFirewallMirroringRule(t, projectID, consumerFwPolicyName, expectedSpgPath)
	t.Log("--- Overall Test Successful: All resources created and validated. ---")
}

func createTestYAML(t *testing.T, projectID, suffix string) {
	dgName := "integ-test-dg-" + suffix
	egName := "integ-test-eg-" + suffix
	depAName := "integ-test-dep-a-" + suffix
	depBName := "integ-test-dep-b-" + suffix
	assocAName := "integ-test-assoc-a-" + suffix
	producerVPCName := "producer-vpc-" + suffix
	consumerVPCName := "consumer-vpc-" + suffix
	frLinkA := fmt.Sprintf("projects/%s/regions/%s/forwardingRules/collector-fr-a-%s", projectID, location, suffix)
	frLinkB := fmt.Sprintf("projects/%s/regions/%s/forwardingRules/collector-fr-b-%s", projectID, location, suffix)
	producerNetworkLink := fmt.Sprintf("projects/%s/global/networks/%s", projectID, producerVPCName)
	consumerNetworkLink := fmt.Sprintf("projects/%s/global/networks/%s", projectID, consumerVPCName)
	config := map[string]any{
		"deployment_group": map[string]any{
			"create":                      true,
			"deployment_group_project_id": projectID,
			"name":                        dgName,
			"producer_network_link":       producerNetworkLink,
		},
		"endpoint_group": map[string]any{
			"create":                    true,
			"endpoint_group_project_id": projectID,
			"name":                      egName,
		},
		"deployments": []any{
			map[string]any{"deployment_project_id": projectID, "name": depAName, "location": zoneA, "forwarding_rule_link": frLinkA},
			map[string]any{"deployment_project_id": projectID, "name": depBName, "location": zoneB, "forwarding_rule_link": frLinkB},
		},
		"endpoint_associations": []any{
			map[string]any{"endpoint_association_project_id": projectID, "name": assocAName, "consumer_network_link": consumerNetworkLink},
		},
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

func runCmdWithRetry(t *testing.T, args ...string) {
	maxRetries := 3
	sleepBetweenRetries := 20 * time.Second
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		t.Logf("Attempt %d/%d: Running gcloud %s", i+1, maxRetries, strings.Join(args, " "))
		_, err := shell.RunCommandAndGetOutputE(t, shell.Command{Command: "gcloud", Args: args})
		if err == nil {
			t.Logf("Command succeeded on attempt %d/%d", i+1, maxRetries)
			return
		}
		lastErr = err
		t.Logf("Attempt %d/%d failed with error: %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			t.Logf("Sleeping for %v before retrying...", sleepBetweenRetries)
			time.Sleep(sleepBetweenRetries)
		}
	}
	t.Fatalf("Command failed after %d retries. Last error: %v", maxRetries, lastErr)
}

func createGCSBucket(t *testing.T, projectID, bucketName, gcsBucketLocation string) {
	t.Logf("Creating GCS bucket '%s'...", bucketName)
	runCmdWithRetry(t, "storage", "buckets", "create", "gs://"+bucketName, "--project="+projectID, "--location="+gcsBucketLocation)
}

func deleteGCSBucket(t *testing.T, bucketName string) {
	t.Logf("--- Deleting GCS Bucket: %s ---", bucketName)
	runCmdWithRetry(t, "storage", "rm", "-r", "gs://"+bucketName)
}

func createVPC(t *testing.T, projectID, networkName, ipRange1, ipRange2, region string) {
	t.Logf("Creating VPC '%s'...", networkName)
	subnetName := networkName + "-subnet"
	allTestRanges := fmt.Sprintf("%s,%s", producerVPCRange, consumerVPCRange)
	runCmdWithRetry(t, "compute", "networks", "create", networkName, "--project="+projectID, "--subnet-mode=custom", "--bgp-routing-mode=global")
	runCmdWithRetry(t, "compute", "networks", "subnets", "create", subnetName, "--project="+projectID, "--network="+networkName, "--range="+ipRange1, "--region="+region)
	runCmdWithRetry(t, "compute", "firewall-rules", "create", "allow-internal-test-"+networkName, "--project="+projectID, "--network="+networkName, "--allow=all", "--source-ranges="+allTestRanges)
}

func deleteVPC(t *testing.T, projectID, networkName, region string) {
	t.Logf("--- Deleting VPC: %s ---", networkName)
	subnetName := networkName + "-subnet"
	runCmdWithRetry(t, "compute", "firewall-rules", "delete", "allow-internal-test-"+networkName, "--project="+projectID, "--quiet")
	runCmdWithRetry(t, "compute", "networks", "subnets", "delete", subnetName, "--project="+projectID, "--region="+region, "--quiet")
	runCmdWithRetry(t, "compute", "networks", "delete", networkName, "--project="+projectID, "--quiet")
}

func createCollectorStack(t *testing.T, projectID, vpc, zoneA, zoneB, hcPort, backendPort, gcsLogPath, containerImage string) {
	t.Log("Creating collector stack (Virtual Machines, Instance Groups, Health Checks, Backend Services, Forwarding Rules)...")
	suffix := strings.Split(vpc, "-")[2]
	region := location
	subnet := vpc + "-subnet"
	for _, zone := range []string{zoneA, zoneB} {
		zoneSuffix := string(zone[len(zone)-1])
		vmName := fmt.Sprintf("collector-vm-%s-%s", zoneSuffix, suffix)
		igName := fmt.Sprintf("collector-ig-%s-%s", zoneSuffix, suffix)
		runCmdWithRetry(t, "compute", "instances", "create-with-container", vmName, "--project="+projectID, "--zone="+zone, "--machine-type=e2-medium", fmt.Sprintf("--network-interface=network=%s,subnet=%s,no-address", vpc, subnet), "--image-family=cos-stable", "--image-project=cos-cloud", "--container-image="+containerImage, fmt.Sprintf("--container-env=BUCKET_PATH=%s,HEALTH_PORT=%s", gcsLogPath, hcPort))
		runCmdWithRetry(t, "compute", "instance-groups", "unmanaged", "create", igName, "--project="+projectID, "--zone="+zone)
		runCmdWithRetry(t, "compute", "instance-groups", "unmanaged", "add-instances", igName, "--project="+projectID, "--zone="+zone, "--instances="+vmName)
	}
	hcName := "collector-hc-" + suffix
	bsName := "collector-bs-" + suffix
	frNameA := "collector-fr-a-" + suffix
	frNameB := "collector-fr-b-" + suffix
	runCmdWithRetry(t, "compute", "health-checks", "create", "tcp", hcName, "--project="+projectID, "--region="+region, "--port="+hcPort)
	runCmdWithRetry(t, "compute", "backend-services", "create", bsName, "--project="+projectID, "--load-balancing-scheme=INTERNAL", "--protocol=UDP", "--health-checks="+hcName, "--health-checks-region="+region, "--network="+vpc, "--region="+region)
	runCmdWithRetry(t, "compute", "backend-services", "add-backend", bsName, "--project="+projectID, "--region="+region, fmt.Sprintf("--instance-group=collector-ig-a-%s", suffix), "--instance-group-zone="+zoneA)
	runCmdWithRetry(t, "compute", "backend-services", "add-backend", bsName, "--project="+projectID, "--region="+region, fmt.Sprintf("--instance-group=collector-ig-b-%s", suffix), "--instance-group-zone="+zoneB)
	runCmdWithRetry(t, "compute", "forwarding-rules", "create", frNameA, "--project="+projectID, "--region="+region, "--load-balancing-scheme=INTERNAL", "--network="+vpc, "--subnet="+subnet, "--ip-protocol=UDP", "--ports="+backendPort, "--backend-service="+bsName, "--is-mirroring-collector")
	runCmdWithRetry(t, "compute", "forwarding-rules", "create", frNameB, "--project="+projectID, "--region="+region, "--load-balancing-scheme=INTERNAL", "--network="+vpc, "--subnet="+subnet, "--ip-protocol=UDP", "--ports="+backendPort, "--backend-service="+bsName, "--is-mirroring-collector")
}

func deleteCollectorStack(t *testing.T, projectID, zoneA, zoneB, region, suffix string) {
	t.Logf("--- Deleting collector stack for suffix: %s ---", suffix)
	frNameA, frNameB := "collector-fr-a-"+suffix, "collector-fr-b-"+suffix
	bsName := "collector-bs-" + suffix
	hcName := "collector-hc-" + suffix
	igNameA, igNameB := "collector-ig-a-"+suffix, "collector-ig-b-"+suffix
	vmNameA, vmNameB := "collector-vm-a-"+suffix, "collector-vm-b-"+suffix
	runCmdWithRetry(t, "compute", "forwarding-rules", "delete", frNameA, "--project="+projectID, "--region="+region, "--quiet")
	runCmdWithRetry(t, "compute", "forwarding-rules", "delete", frNameB, "--project="+projectID, "--region="+region, "--quiet")
	runCmdWithRetry(t, "compute", "backend-services", "delete", bsName, "--project="+projectID, "--region="+region, "--quiet")
	runCmdWithRetry(t, "compute", "health-checks", "delete", hcName, "--project="+projectID, "--region="+region, "--quiet")
	runCmdWithRetry(t, "compute", "instance-groups", "unmanaged", "delete", igNameA, "--project="+projectID, "--zone="+zoneA, "--quiet")
	runCmdWithRetry(t, "compute", "instance-groups", "unmanaged", "delete", igNameB, "--project="+projectID, "--zone="+zoneB, "--quiet")
	runCmdWithRetry(t, "compute", "instances", "delete", vmNameA, "--project="+projectID, "--zone="+zoneA, "--quiet")
	runCmdWithRetry(t, "compute", "instances", "delete", vmNameB, "--project="+projectID, "--zone="+zoneB, "--quiet")
}

func createWorkloadVMs(t *testing.T, projectID, producerVPC, consumerVPC, zone, listenerPort, suffix string) {
	t.Log("Creating workload VMs (sender in producer-vpc, listener in consumer-vpc)...")
	producerSubnet := producerVPC + "-subnet"
	consumerSubnet := consumerVPC + "-subnet"
	listenerName, senderName := "listener-vm-"+suffix, "sender-vm-"+suffix
	listenerImage := fmt.Sprintf("gcr.io/%s/chatter-listener:v1", projectID)
	senderImage := fmt.Sprintf("gcr.io/%s/chatter-sender:v1", projectID)
	runCmdWithRetry(t, "compute", "instances", "create-with-container", listenerName, "--project="+projectID, "--zone="+zone, "--machine-type=e2-medium", fmt.Sprintf("--network-interface=network=%s,subnet=%s,no-address", consumerVPC, consumerSubnet), "--image-family=cos-stable", "--image-project=cos-cloud", "--container-image="+listenerImage, "--container-env=LISTEN_PORT="+listenerPort)
	runCmdWithRetry(t, "compute", "instances", "create-with-container", senderName, "--project="+projectID, "--zone="+zone, "--machine-type=e2-medium", fmt.Sprintf("--network-interface=network=%s,subnet=%s,no-address", producerVPC, producerSubnet), "--image-family=cos-stable", "--image-project=cos-cloud", "--container-image="+senderImage, fmt.Sprintf("--container-env=TARGET_HOST=%s,TARGET_PORT=%s", listenerName, listenerPort))
}

func deleteWorkloadVMs(t *testing.T, projectID, zone, suffix string) {
	t.Logf("--- Deleting workload VMs for suffix: %s ---", suffix)
	listenerName, senderName := "listener-vm-"+suffix, "sender-vm-"+suffix
	runCmdWithRetry(t, "compute", "instances", "delete", listenerName, "--project="+projectID, "--zone="+zone, "--quiet")
	runCmdWithRetry(t, "compute", "instances", "delete", senderName, "--project="+projectID, "--zone="+zone, "--quiet")
}

func createFirewallPolicy(t *testing.T, projectID, policyName string) {
	t.Logf("Creating firewall policy '%s'", policyName)
	runCmdWithRetry(t, "compute", "network-firewall-policies", "create", policyName, "--project="+projectID, "--description=integ-test", "--global")
}

func associateFirewallPolicy(t *testing.T, projectID, vpc, policyName string) {
	t.Logf("Associating firewall policy '%s' with VPC '%s'", policyName, vpc)
	runCmdWithRetry(t, "compute", "network-firewall-policies", "associations", "create", "--firewall-policy="+policyName, "--network="+vpc, "--name=assoc-"+vpc, "--project="+projectID, "--global-firewall-policy")
}

func deleteFirewallPolicyAssociation(t *testing.T, projectID, vpc, policyName string) {
	associationName := "assoc-" + vpc
	t.Logf("--- Deleting firewall policy association: %s ---", associationName)
	runCmdWithRetry(t, "compute", "network-firewall-policies", "associations", "delete", "--name="+associationName, "--firewall-policy="+policyName, "--project="+projectID, "--global-firewall-policy")
}

func deleteFirewallPolicy(t *testing.T, projectID, policyName string) {
	t.Logf("--- Deleting firewall policy: %s ---", policyName)
	runCmdWithRetry(t, "compute", "network-firewall-policies", "delete", policyName, "--project="+projectID, "--global", "--quiet")
}

func createSecurityProfile(t *testing.T, orgID, projectID, spName, spgName, endpointGroupID string) {
	t.Logf("Creating security profile '%s' and group '%s'", spName, spgName)
	spPath := fmt.Sprintf("organizations/%s/locations/global/securityProfiles/%s", orgID, spName)
	runCmdWithRetry(t, "network-security", "security-profiles", "custom-mirroring", "create", spName, "--organization="+orgID, "--location=global", "--billing-project="+projectID, "--mirroring-endpoint-group="+endpointGroupID)
	runCmdWithRetry(t, "network-security", "security-profile-groups", "create", spgName, "--organization="+orgID, "--location=global", "--billing-project="+projectID, "--custom-mirroring-profile="+spPath)
}

func deleteSecurityProfile(t *testing.T, orgID, spName, spgName string) {
	t.Logf("--- Deleting security profile group '%s' and profile '%s' ---", spgName, spName)
	runCmdWithRetry(t, "network-security", "security-profile-groups", "delete", spgName, "--organization="+orgID, "--location=global", "--quiet")
	runCmdWithRetry(t, "network-security", "security-profiles", "custom-mirroring", "delete", spName, "--organization="+orgID, "--location=global", "--quiet")
}

func createFirewallMirroringRule(t *testing.T, orgID, projectID, policyName, spgName string) {
	t.Logf("Creating firewall mirroring rule using SPG '%s'", spgName)
	spgPath := fmt.Sprintf("organizations/%s/locations/global/securityProfileGroups/%s", orgID, spgName)
	runCmdWithRetry(t, "compute", "network-firewall-policies", "mirroring-rules", "create", "200", "--project="+projectID, "--firewall-policy="+policyName, "--action=MIRROR", "--direction=INGRESS", "--layer4-configs=udp:50051", "--src-ip-ranges=10.20.10.0/24", "--security-profile-group="+spgPath, "--global-firewall-policy")
}

func deleteFirewallMirroringRule(t *testing.T, projectID, policyName string) {
	t.Logf("--- Deleting firewall mirroring rule '200' from policy '%s' ---", policyName)
	runCmdWithRetry(t, "compute", "network-firewall-policies", "mirroring-rules", "delete", "200", "--project="+projectID, "--firewall-policy="+policyName, "--global-firewall-policy")
}

func verifySecurityProfiles(t *testing.T, orgID, spName, spgName string) {
	t.Logf("Verifying creation of Security Profile '%s' and Group '%s'", spName, spgName)
	spDescribeCmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"network-security", "security-profiles", "custom-mirroring", "describe", spName, "--organization=" + orgID, "--location=global"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, spDescribeCmd)
	if err != nil {
		t.Fatalf("Validation failed: Could not describe security profile %s: %v", spName, err)
	}
	spgDescribeCmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"network-security", "security-profile-groups", "describe", spgName, "--organization=" + orgID, "--location=global"},
	}
	_, err = shell.RunCommandAndGetOutputE(t, spgDescribeCmd)
	if err != nil {
		t.Fatalf("Validation failed: Could not describe security profile group %s: %v", spgName, err)
	}
	t.Logf("Validation successful: Security Profile '%s' and Group '%s' exist.", spName, spgName)
}

func verifyFirewallMirroringRule(t *testing.T, projectID, policyName, expectedSpgPath string) {
	t.Logf("Verifying creation of firewall mirroring rule in policy '%s'", policyName)
	ruleDescribeCmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "network-firewall-policies", "mirroring-rules", "describe", "200", "--project=" + projectID, "--firewall-policy=" + policyName, "--global-firewall-policy", "--format=json"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, ruleDescribeCmd)
	if err != nil {
		t.Fatalf("Validation failed: Could not describe firewall mirroring rule '200' in policy %s: %v", policyName, err)
	}
	var rules []map[string]interface{}
	err = json.Unmarshal([]byte(output), &rules)
	if err != nil {
		t.Fatalf("Failed to parse JSON output for firewall rule: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("Expected to find exactly one firewall rule with priority 200, but found %d", len(rules))
	}
	ruleData := rules[0]
	assert.Equal(t, "mirror", ruleData["action"], "Firewall rule action should be 'mirror'")
	assert.False(t, ruleData["disabled"].(bool), "Firewall rule should be enabled")
	fullSpgPath, ok := ruleData["securityProfileGroup"].(string)
	if !ok {
		t.Fatal("securityProfileGroup field is not a string")
	}
	assert.Contains(t, fullSpgPath, expectedSpgPath, "Firewall rule is not linked to the correct security profile group")
	t.Logf("Validation successful: Firewall mirroring rule '200' exists and is correctly configured.")
}

func getOrgIDFromProject(t *testing.T, projectID string) string {
	t.Logf("Attempting to find Organization ID for project '%s'...", projectID)
	args := []string{"projects", "describe", projectID, "--format=value(parent.id)"}
	cmd := shell.Command{Command: "gcloud", Args: args}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	require.NoError(t, err, "Failed to run gcloud command to get organization ID for project %s", projectID)
	orgID := strings.TrimSpace(output)
	require.NotEmpty(t, orgID, "Organization ID was not found for project %s. Ensure the project is directly under an organization.", projectID)
	t.Logf("Found Organization ID: %s", orgID)
	return orgID
}
