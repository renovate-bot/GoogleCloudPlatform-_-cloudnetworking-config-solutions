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

package common_utils

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/shell"
	stdlib_strconv "strconv"
	"strings"
	"testing"
	"time"
)

/*
CreateVPCSubnets is a helper function which creates the VPC and subnets before
execution of the test expecting to use existing VPC and subnets.
*/

func CreateVPCSubnets(t *testing.T, projectID string, networkName string, subnetworkName string, region string) {
	subnetworkIPCIDR := "10.0.1.0/24"
	text := "compute"
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{text, "networks", "create", networkName, "--project=" + projectID, "--format=json", "--bgp-routing-mode=global", "--subnet-mode=custom", "--verbosity=none"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("===Error %s Encountered while executing %s", err, text)
	}
	time.Sleep(60 * time.Second)
	if subnetworkName != "" {
		if region == "" {
			region = "us-central1"
		}
		cmd = shell.Command{
			Command: "gcloud",
			Args:    []string{text, "networks", "subnets", "create", subnetworkName, "--network=" + networkName, "--project=" + projectID, "--range=" + subnetworkIPCIDR, "--region=" + region, "--format=json", "--enable-private-ip-google-access", "--enable-flow-logs", "--verbosity=none"},
		}
		_, err = shell.RunCommandAndGetOutputE(t, cmd)
		if err != nil {
			t.Errorf("===Error %s Encountered while executing %s", err, text)
		}
	} else {
		t.Log("VPC will be created & Subnet will not be created.")
	}

}

/*
DeleteVPCSubnets is a helper function which deletes the VPC and subnets after
completion of the test expecting to use existing VPC and subnets.
*/
func DeleteVPCSubnets(t *testing.T, projectID string, networkName string, subnetworkName string, region string) {
	text := "compute"
	if subnetworkName != "" {
		cmd := shell.Command{
			Command: "gcloud",
			Args:    []string{text, "networks", "subnets", "delete", subnetworkName, "--region=" + region, "--project=" + projectID, "--quiet"},
		}
		_, err := shell.RunCommandAndGetOutputE(t, cmd)
		if err != nil {
			t.Errorf("===Error %s Encountered while executing %s", err, text)
		}
	}

	// Sleep for 60 seconds to ensure the deleted subnets is reliably reflected.
	time.Sleep(60 * time.Second)

	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{text, "networks", "delete", networkName, "--project=" + projectID, "--quiet"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("===Error %s Encountered while executing %s", err, text)
	}

}

/*
CreateServiceConnectionPolicy is a helped function that creates the service
connection policy.
*/
func CreateServiceConnectionPolicy(t *testing.T, projectID string, region string, networkName string, policyName string, subnetworkName string, serviceClass string, connectionLimit int) {
	// Get subnet self link from subnet ID using gcloud command
	subnetSelfLink := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/subnetworks/%s", projectID, region, subnetworkName)

	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"network-connectivity", "service-connection-policies", "create",
			policyName, // Add the policyName here as the first argument after "create"
			"--project", projectID,
			"--region", region,
			"--network", networkName,
			"--service-class", serviceClass,
			"--subnets", subnetSelfLink,
			"--psc-connection-limit", fmt.Sprintf("%d", connectionLimit),
			"--quiet",
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("Error creating Service Connection Policy: %s", err)
	}
}

/*
DeletePSA is a helper function which deletes the PSA range after the
execution of the test.
*/
func DeletePSA(t *testing.T, projectID string, networkName string, rangeName string) {
	// Delete PSA IP range
	time.Sleep(60 * time.Second)
	text := "compute"
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{text, "addresses", "delete", rangeName, "--project=" + projectID, "--global", "--verbosity=info", "--format=json", "--quiet"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("===Error %s Encountered while executing %s", err, text)
	}
	time.Sleep(60 * time.Second)
	// Delete PSA range
	text = "services"
	cmd = shell.Command{
		Command: "gcloud",
		Args:    []string{text, "vpc-peerings", "delete", "--service=servicenetworking.googleapis.com", "--project=" + projectID, "--network=" + networkName, "--verbosity=info", "--format=json", "--quiet"},
	}
	_, err = shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("===Error %s Encountered while executing %s", err, text)
	}
}

/*
CreatePSA is a helper function which creates the PSA range before the
execution of the test.
*/
func CreatePSA(t *testing.T, projectID string, networkName string, rangeName string) {
	// Create an IP range
	text := "compute"
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{text, "addresses", "create", rangeName, "--purpose=VPC_PEERING", "--addresses=10.0.64.0", "--prefix-length=20", "--project=" + projectID, "--network=" + networkName, "--global", "--verbosity=info", "--format=json"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("===Error %s Encountered while executing %s", err, text)
	}
	// Create PSA range
	text = "services"
	cmd = shell.Command{
		Command: "gcloud",
		Args:    []string{text, "vpc-peerings", "connect", "--service=servicenetworking.googleapis.com", "--ranges=" + rangeName, "--project=" + projectID, "--network=" + networkName, "--verbosity=info", "--format=json"},
	}
	_, err = shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("===Error %s Encountered while executing %s", err, text)
	}
	time.Sleep(60 * time.Second)
}

// getProjectNumber retrieves the project number for a given project ID.
func GetProjectNumber(t *testing.T, projectID string) (string, error) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"projects", "describe", projectID, "--format=value(projectNumber)"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		return "", fmt.Errorf("Error getting project number: %v", err)
	}

	// The gcloud command might output warnings before the actual project number,
	// especially with impersonation. The project number is expected to be the last
	// non-empty line of the output.
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("No output from gcloud projects describe for project ID %s", projectID)
	}

	// Get the last line and trim any surrounding single quotes
	projectNumber := strings.Trim(lines[len(lines)-1], "'")

	// Basic validation that it looks like a number
	if _, err := stdlib_strconv.ParseInt(projectNumber, 10, 64); err != nil {
		return "", fmt.Errorf("Extracted project number '%s' is not a valid number. Full output: %s. Error: %v", projectNumber, output, err)
	}

	return projectNumber, nil
}

// getAttachmentProjectNumber retrieves the project number for the attachment project.
// If TF_VAR_ATTACHMENT_PROJECT_ID is not set, it defaults to the primary project ID.
func GetAttachmentProjectNumber(t *testing.T, projectID string, attachmentProjectID string) (string, error) {
	// If attachmentProjectID is not set, use the primary projectID as fallback.
	if attachmentProjectID == "" {
		attachmentProjectID = projectID
		t.Logf("TF_VAR_ATTACHMENT_PROJECT_ID not set. Defaulting to primary project ID: %s", projectID)
		return GetProjectNumber(t, projectID) // Use the global projectID as the fallback
	}

	return GetProjectNumber(t, attachmentProjectID)
}
