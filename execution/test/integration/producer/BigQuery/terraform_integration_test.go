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

package integrationtest

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/common_utils"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

const (
	location               = "us-central1"
	terraformDirectoryPath = "../../../../04-producer/BigQuery"
	terraformConfigDir     = "../../test/integration/producer/BigQuery/config"
	localConfigDir         = "config"
	tableID                = "test_table"
	bqLocation             = "US"
)

// BigQueryStruct defines the structure for a BigQuery dataset configuration YAML file.
type BigQueryStruct struct {
	ProjectID   string        `yaml:"project_id"`
	DatasetID   string        `yaml:"dataset_id"`
	DatasetName string        `yaml:"dataset_name"`
	Location    string        `yaml:"location"`
	Tables      []TableStruct `yaml:"tables"`
}

// TableStruct defines the structure for a table within a BigQuery dataset.
type TableStruct struct {
	TableID     string `yaml:"table_id"`
	Description string `yaml:"description"`
	Schema      string `yaml:"schema"`
}

var randID int

// TestMain sets up a random ID to ensure resource names are unique for each test run.
func TestMain(m *testing.M) {
	randID = rand.Intn(10000)
	os.Exit(m.Run())
}

/*
TestCreateBigQueryDataset validates the basic creation of a BigQuery dataset and table.
*/
func TestCreateBigQueryDataset(t *testing.T) {
	projectID := os.Getenv("TF_VAR_project_id")
	if projectID == "" {
		t.Fatalf("Environment variable TF_VAR_project_id is not set. Please set it to your GCP project ID to run integration tests.")
	}

	datasetID := fmt.Sprintf("bq_base_test_%d", randID)

	common_utils.CleanupConfigDir(t, localConfigDir)
	createBigQueryConfigYAML(t, projectID, datasetID, localConfigDir)
	defer deleteBigQueryConfigYAML(t, localConfigDir, datasetID)

	var tfVars = map[string]any{
		"config_folder_path": terraformConfigDir,
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
	})

	defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	outputJSON := terraform.OutputJson(t, terraformOptions, "bigquery_dataset_details")
	if !gjson.Valid(outputJSON) {
		t.Fatalf("Error parsing output, invalid json: %s", outputJSON)
	}

	datasetIDPath := fmt.Sprintf("%s.dataset_id", datasetID)
	gotDatasetID := gjson.Get(outputJSON, datasetIDPath).String()
	wantDatasetID := fmt.Sprintf("projects/%s/datasets/%s", projectID, datasetID)
	if gotDatasetID != wantDatasetID {
		t.Errorf("bigquery dataset with incorrect ID created. got=%v, want=%v", gotDatasetID, wantDatasetID)
	} else {
		t.Logf("Successfully verified BigQuery dataset ID: %s", gotDatasetID)
	}

	tableIDPath := fmt.Sprintf("%s.table_ids.0", datasetID)
	gotTableID := gjson.Get(outputJSON, tableIDPath).String()
	if gotTableID != tableID {
		t.Errorf("bigquery table with incorrect ID created. got=%v, want=%v", gotTableID, tableID)
	} else {
		t.Logf("Successfully verified BigQuery table ID: %s", gotTableID)
	}
}

/*
TestGCEStartupScriptConnectsToBigQuery validates that a GCE instance with no external IP
can create and access a BigQuery dataset via its startup script, verifying Private Google Access.
*/
func TestGCEStartupScriptConnectsToBigQuery(t *testing.T) {
	projectID := os.Getenv("TF_VAR_project_id")
	if projectID == "" {
		t.Fatalf("Environment variable TF_VAR_project_id is not set.")
	}

	vpcName := fmt.Sprintf("vpc-startup-test-%d", randID)
	subnetName := fmt.Sprintf("subnet-startup-test-%d", randID)
	vmName := fmt.Sprintf("vm-startup-test-%d", randID)
	datasetIDByVM := fmt.Sprintf("ds_from_vm_%d", randID)
	zone := fmt.Sprintf("%s-a", location)
	successMessage := "BIGQUERY_CONNECTIVITY_TEST_SUCCESS"

	common_utils.CreateVPCSubnets(t, projectID, vpcName, subnetName, location)
	defer common_utils.DeleteVPCSubnets(t, projectID, vpcName, subnetName, location)

	time.Sleep(30 * time.Second)

	startupScript := fmt.Sprintf(`
		#!/bin/bash
		apt-get update -y
		apt-get install -y google-cloud-sdk
		bq --location=%s mk --dataset %s:%s
		bq show %s:%s
		if [ $? -eq 0 ]; then
			echo "%s"
		else
			echo "BIGQUERY_CONNECTIVITY_TEST_FAILURE"
		fi
	`, bqLocation, projectID, datasetIDByVM, projectID, datasetIDByVM, successMessage)

	common_utils.CreateGCEInstance(t, projectID, vmName, zone, subnetName, startupScript, "", false, "", "") // false for no external IP
	defer common_utils.DeleteGCEInstance(t, projectID, vmName, zone)
	defer deleteBigQueryDataset(t, projectID, datasetIDByVM)

	t.Logf("Waiting up to 4 minutes for startup script on '%s' to complete...", vmName)
	for i := 0; i < 24; i++ {
		time.Sleep(10 * time.Second)
		output, err := common_utils.GetSerialPortOutput(t, projectID, vmName, zone, 1)
		if err != nil {
			t.Logf("Warning: could not get serial port output on attempt %d: %v", i+1, err)
			continue
		}

		if strings.Contains(output, successMessage) {
			t.Logf("Success! Found success message in serial port logs for VM '%s'.", vmName)
			return
		}
		if strings.Contains(output, "BIGQUERY_CONNECTIVITY_TEST_FAILURE") {
			t.Fatalf("Failure message found in serial port logs for VM '%s'. Full output:\n%s", vmName, output)
		}
	}

	t.Fatalf("Timeout: Did not find success message in serial port logs for VM '%s' after 4 minutes.", vmName)
}

// --- HELPER FUNCTIONS (BigQuery) ---

func createBigQueryConfigYAML(t *testing.T, projectID string, datasetID string, configPath string) {
	tableSchema := `[{"name": "id", "type": "STRING", "mode": "REQUIRED"}, {"name": "data", "type": "STRING", "mode": "NULLABLE"}]`
	bigqueryConfig := BigQueryStruct{
		ProjectID:   projectID,
		DatasetID:   datasetID,
		DatasetName: datasetID,
		Location:    bqLocation,
		Tables: []TableStruct{
			{TableID: tableID, Description: "A test table for integration tests.", Schema: tableSchema},
		},
	}
	yamlData, err := yaml.Marshal(&bigqueryConfig)
	if err != nil {
		t.Fatalf("Error marshalling YAML: %v", err)
	}
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Unable to create config directory %s: %v", configPath, err)
	}
	fullConfigFilePath := filepath.Join(configPath, fmt.Sprintf("dataset_%s.yaml", datasetID))
	err = os.WriteFile(fullConfigFilePath, yamlData, 0644)
	if err != nil {
		t.Fatalf("Unable to write data to file %s: %v", fullConfigFilePath, err)
	}
}

func deleteBigQueryConfigYAML(t *testing.T, configPath string, datasetID string) {
	fullConfigFilePath := filepath.Join(configPath, fmt.Sprintf("dataset_%s.yaml", datasetID))
	if err := os.Remove(fullConfigFilePath); err != nil {
		t.Logf("Warning: Could not remove temporary config file %s: %v", fullConfigFilePath, err)
	}
}

func deleteBigQueryDataset(t *testing.T, projectID, datasetID string) {
	cmd := shell.Command{
		Command: "bq",
		Args: []string{
			"rm",
			"--project_id=" + projectID,
			"-f", // force delete
			"-d", // delete dataset
			datasetID,
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Warning: failed to clean up BigQuery dataset '%s'. This may require manual cleanup. Error: %v", datasetID, err)
	}
}
