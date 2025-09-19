// Copyright 2024-2025 Google LLC
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
	"github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/common_utils"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
	"math/rand"
	"os"
	"path"
	"reflect"
	"testing"
	"time"
)

var (
	projectID              = os.Getenv("TF_VAR_project_id")
	region                 = "us-central1"
	terraformDirectoryPath = "../../../../04-producer/AlloyDB"
	configFolderPath       = "../../test/integration/producer/AlloyDB/config"
	rangeName              = fmt.Sprintf("psatestrangealloydb-%s", clusterDisplayName)
	clusterDisplayName     = fmt.Sprint(rand.Int())
	networkName            = fmt.Sprintf("vpc-%s-test", clusterDisplayName)
	alloyDBClusterID       = fmt.Sprintf("cid-%s-test", clusterDisplayName)
	instanceID             = fmt.Sprintf("id-%s-test", clusterDisplayName)
	networkID              = fmt.Sprintf("projects/%s/global/networks/%s", projectID, networkName)
)

type PrimaryInstanceStruct struct {
	InstanceID    string      `yaml:"instance_id"`
	DisplayName   string      `yaml:"display_name"`
	InstanceType  string      `yaml:"instance_type"`
	MachineCPUs   int         `yaml:"machine_cpu_count"`
	DatabaseFlags interface{} `yaml:"database_flags"`
}

type AlloyDBStruct struct {
	ClusterID                  string                `yaml:"cluster_id"`
	ClusterDisplayName         string                `yaml:"cluster_display_name"`
	ProjectID                  string                `yaml:"project_id"`
	Region                     string                `yaml:"region"`
	NetworkID                  string                `yaml:"network_id"`
	PrimaryInstance            PrimaryInstanceStruct `yaml:"primary_instance"`
	AllocatedIPRange           string                `yaml:"allocated_ip_range"`
	PscAllowedConsumerProjects []string              `yaml:"psc_allowed_consumer_projects"`
	ConnectivityOptions        string                `yaml:"connectivity_options"`
	ReadPoolInstance           interface{}           `yaml:"read_pool_instance"`
	AutomatedBackupPolicy      interface{}           `yaml:"automated_backup_policy"`
	DeletionProtection         bool                  `yaml:"deletion_protection"`
}

/*
This test creates all the pre-requsite resources including the vpc network, subnetwork along with a PSA range.
It then validates if
1. AlloyDB instance is created.
2. AlloyDB instance is created in the correct network and correct PSA range.
3. AlloyDB instance is in ACTIVE state.
*/
func TestCreateAlloyDB(t *testing.T) {
	// Initialize AlloyDB config YAML files
	createConfigYAMLs(t)

	// Get the project number
	projectNumber, err := common_utils.GetProjectNumber(t, projectID)
	if err != nil {
		t.Fatal(err)
	}
	attachmentProjectID := os.Getenv("TF_VAR_ATTACHMENT_PROJECT_ID")
	// Get the attachment project number
	attachmentProjectNumber, err := common_utils.GetAttachmentProjectNumber(t, projectID, attachmentProjectID)
	if err != nil {
		t.Fatal(err)
	}

	var (
		tfVars = map[string]any{
			"config_folder_path": configFolderPath,
		}
	)

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		// Set the path to the Terraform code that will be tested.
		Vars:                 tfVars,
		TerraformDir:         terraformDirectoryPath,
		Reconfigure:          true,
		Lock:                 true,
		NoColor:              true,
		SetVarsAfterVarFiles: true,
	})
	// Create VPC outside of the terraform module.
	common_utils.CreateVPCSubnets(t, projectID, networkName, "", "")
	if err != nil {
		t.Fatal(err)
	}
	// Create PSA in the VPC.
	common_utils.CreatePSA(t, projectID, networkName, rangeName)

	// Delete VPC created outside of the terraform module.
	defer common_utils.DeleteVPCSubnets(t, projectID, networkName, "", "")

	// Remove PSA from the VPC.
	defer common_utils.DeletePSA(t, projectID, networkName, rangeName)

	// Clean up resources with "terraform destroy" at the end of the test.
	defer terraform.Destroy(t, terraformOptions)

	// Run "terraform init" and "terraform apply". Fail the test if there are any errors.
	terraform.InitAndApply(t, terraformOptions)

	// Wait for 60 seconds to let resource acheive stable state.
	time.Sleep(60 * time.Second)

	// Run `terraform output` to get the values of output variables
	alloyDBOutputValue := terraform.OutputJson(t, terraformOptions, "cluster_details")
	if !gjson.Valid(alloyDBOutputValue) {
		t.Fatalf("Error parsing output, invalid JSON: %s", alloyDBOutputValue)
	}

	result := gjson.Parse(alloyDBOutputValue)

	// Define cluster keys for easier access
	psaClusterKey := fmt.Sprint(clusterDisplayName)
	pscClusterKey := fmt.Sprintf("%s-psc", clusterDisplayName)
	clusterKeys := []string{psaClusterKey, pscClusterKey}

	for _, clusterKey := range clusterKeys {
		t.Logf(" ========= Verifying Cluster: %s ========= ", clusterKey)

		// Verify Cluster ID
		t.Log(" ========= Verify AlloyDB Cluster ID ========= ")
		clusterIDPath := fmt.Sprintf("%s.cluster_id", clusterKey)
		gotClusterID := gjson.Get(result.String(), clusterIDPath).String()
		gotClusterID = path.Base(gotClusterID)

		wantClusterID := fmt.Sprintf("cid-%s-test", clusterDisplayName)
		if clusterKey == pscClusterKey {
			wantClusterID = fmt.Sprintf("cid-%s-test-psc", clusterDisplayName)
		}
		if gotClusterID != wantClusterID {
			t.Errorf("AlloyDB Cluster with invalid Cluster ID = %v, want = %v", gotClusterID, wantClusterID)
		}

		// Verify AlloyDB Cluster Status
		t.Log(" ========= Verify AlloyDB Cluster Status ========= ")
		wantStatus := "READY"
		clusterStatusPath := fmt.Sprintf("%s.cluster_status", clusterKey)
		gotStatus := gjson.Get(result.String(), clusterStatusPath).String()
		if gotStatus != wantStatus {
			t.Errorf("AlloyDB Cluster with invalid Cluster status = %v, want = %v", gotStatus, wantStatus)
		}

		// Verify Allocated IP Range (only for PSA)
		t.Log(" ========= Verify Allocated IP Range ========= ")
		allocatedIPRangePath := fmt.Sprintf("%s.network_config.allocated_ip_range", clusterKey)
		gotAllocatedIPRange := gjson.Get(result.String(), allocatedIPRangePath).String()

		if clusterKey == psaClusterKey {
			wantAllocatedIPRange := fmt.Sprintf("psatestrangealloydb-%s", clusterDisplayName)
			if gotAllocatedIPRange != wantAllocatedIPRange {
				t.Errorf("Allocated IP range mismatch for PSA cluster. Got: %v, Want: %v", gotAllocatedIPRange, wantAllocatedIPRange)
			}
		} else {
			if gotAllocatedIPRange != "" {
				t.Errorf("Allocated IP range should be empty for non-PSA cluster. Got: %v", gotAllocatedIPRange)
			}
		}

		// Verify Connectivity Options
		t.Log(" ========= Verify Connectivity Options ========= ")
		connectivityOptionsPath := fmt.Sprintf("%s.connectivity_options", clusterKey)
		gotConnectivityOptions := gjson.Get(result.String(), connectivityOptionsPath).String()
		wantConnectivityOptions := "PSA"
		if clusterKey == pscClusterKey {
			wantConnectivityOptions = "PSC"
		}
		if gotConnectivityOptions != wantConnectivityOptions {
			t.Errorf("Connectivity Options mismatch. Got: %v, Want: %v", gotConnectivityOptions, wantConnectivityOptions)
		}

		// Verify PSC Allowed Consumer Projects
		t.Log(" ========= Verify PSC Allowed Consumer Projects ========= ")
		consumerProjectsPath := fmt.Sprintf("%s.network_config.psc_config.configured_allowed_consumer_projects", clusterKey)
		gotConsumerProjects := gjson.Get(result.String(), consumerProjectsPath).Array()
		if clusterKey == pscClusterKey {
			wantConsumerProjects := []string{projectNumber, attachmentProjectNumber}
			gotConsumerProjectsStr := []string{}
			for _, v := range gotConsumerProjects {
				gotConsumerProjectsStr = append(gotConsumerProjectsStr, v.String())
			}
			if !reflect.DeepEqual(gotConsumerProjectsStr, wantConsumerProjects) {
				t.Errorf("PSC consumer projects mismatch. Got: %v, Want: %v", gotConsumerProjectsStr, wantConsumerProjects)
			}
		} else {
			if len(gotConsumerProjects) > 0 {
				t.Errorf("PSC consumer projects expected to be empty. Got: %v", gotConsumerProjects)
			}
		}

		// Verify Database Version
		t.Log(" ========= Verify Database Version ========= ")
		databaseVersionPath := fmt.Sprintf("%s.database_version", clusterKey)
		gotDatabaseVersion := gjson.Get(result.String(), databaseVersionPath).String()
		wantDatabaseVersion := "POSTGRES_15"
		if gotDatabaseVersion != wantDatabaseVersion {
			t.Errorf("Database version mismatch. Got: %v, Want: %v", gotDatabaseVersion, wantDatabaseVersion)
		}

	}
}

/*
createConfigYAML is a helper function which creates the configigration YAML file
for an alloydb instance range before the.
*/
func createConfigYAMLs(t *testing.T) {
	// Get the project number
	projectNumber, err := common_utils.GetProjectNumber(t, projectID)
	if err != nil {
		t.Fatal(err)
	}
	attachmentProjectID := os.Getenv("TF_VAR_ATTACHMENT_PROJECT_ID")
	// Get the attachment project number (replace with your actual logic)
	attachmentProjectNumber, err := common_utils.GetAttachmentProjectNumber(t, projectID, attachmentProjectID)
	if err != nil {
		t.Fatal(err)
	}

	instance1 := AlloyDBStruct{ // PSA config
		ClusterID:          alloyDBClusterID,
		ClusterDisplayName: clusterDisplayName,
		ProjectID:          projectID,
		Region:             region,
		NetworkID:          networkID,
		AllocatedIPRange:   rangeName,
		PrimaryInstance: PrimaryInstanceStruct{
			InstanceID:    instanceID,
			DisplayName:   instanceID,
			InstanceType:  "PRIMARY",
			MachineCPUs:   2,
			DatabaseFlags: nil,
		},
		ReadPoolInstance:           nil,
		AutomatedBackupPolicy:      nil,
		DeletionProtection:         false,
		ConnectivityOptions:        "psa",
		PscAllowedConsumerProjects: []string{projectNumber, attachmentProjectNumber},
	}

	instance2 := AlloyDBStruct{ // PSC config
		ClusterID:          alloyDBClusterID + "-psc",
		ClusterDisplayName: clusterDisplayName + "-psc",
		ProjectID:          projectID,
		Region:             region,
		NetworkID:          networkID,
		PrimaryInstance: PrimaryInstanceStruct{
			InstanceID:    instanceID + "-psc",
			DisplayName:   instanceID + "-psc",
			InstanceType:  "PRIMARY",
			MachineCPUs:   2,
			DatabaseFlags: nil,
		},
		ReadPoolInstance:           nil,
		AutomatedBackupPolicy:      nil,
		DeletionProtection:         false,
		ConnectivityOptions:        "psc",
		PscAllowedConsumerProjects: []string{projectNumber, attachmentProjectNumber},
		AllocatedIPRange:           "", // No Allocated IP Range for PSC
	}

	yamlData1, err := yaml.Marshal(&instance1)
	if err != nil {
		t.Errorf("Error marshalling instance1: %v", err)
	}
	filePath1 := fmt.Sprintf("%s/%s", "config", "instance1.yaml")
	err = os.WriteFile(filePath1, []byte(yamlData1), 0666)
	if err != nil {
		t.Errorf("Unable to write instance1 data: %v", err)
	}

	yamlData2, err := yaml.Marshal(&instance2)
	if err != nil {
		t.Errorf("Error marshalling instance2: %v", err)
	}
	filePath2 := fmt.Sprintf("%s/%s", "config", "instance2.yaml")
	err = os.WriteFile(filePath2, []byte(yamlData2), 0666)
	if err != nil {
		t.Errorf("Unable to write instance2 data: %v", err)
	}
}
