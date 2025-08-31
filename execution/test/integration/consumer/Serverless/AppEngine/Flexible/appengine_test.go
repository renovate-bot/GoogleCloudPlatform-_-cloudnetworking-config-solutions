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
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/test/integration/common_utils"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

var (
	projectRoot, _         = filepath.Abs("../../../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "06-consumer/Serverless/AppEngine/Flexible")
	configFolderPath       = filepath.Join(projectRoot, "test/integration/consumer/Serverless/AppEngine/Flexible/config")
)

var (
	projectID     = os.Getenv("TF_VAR_project_id")
	region        = "us-central1"
	instanceName  = fmt.Sprintf("appeng-flex-test-%d", rand.Intn(10000))
	networkName   = fmt.Sprintf("vpc-%s", instanceName)
	subnetName    = fmt.Sprintf("%s-subnet", networkName)
	gcsBucketName string
	gcsSourceURL  string
)

type AppEngineConfig struct {
	Project                 string `yaml:"project_id"`
	Service                 string `yaml:"service"`
	Runtime                 string `yaml:"runtime"`
	FlexibleRuntimeSettings struct {
		OperatingSystem string `yaml:"operating_system"`
		RuntimeVersion  string `yaml:"runtime_version"`
	} `yaml:"flexible_runtime_settings"`
	InstanceClass string `yaml:"instance_class"`
	Network       struct {
		Name       string `yaml:"name"`
		Subnetwork string `yaml:"subnetwork"`
	} `yaml:"network"`
	VersionID     string `yaml:"version_id"`
	ManualScaling *struct {
		Instances int `yaml:"instances"`
	} `yaml:"manual_scaling,omitempty"`

	Entrypoint struct {
		Shell string `yaml:"shell"`
	} `yaml:"entrypoint"`

	Deployment *struct {
		Zip *struct {
			SourceURL string `yaml:"source_url"`
		} `yaml:"zip,omitempty"`
	} `yaml:"deployment,omitempty"`
	LivenessCheck          map[string]interface{} `yaml:"liveness_check,omitempty"`
	ReadinessCheck         map[string]interface{} `yaml:"readiness_check,omitempty"`
	DeleteServiceOnDestroy bool                   `yaml:"delete_service_on_destroy,omitempty"`
	EnvVariables           map[string]string      `yaml:"env_variables,omitempty"`
	ServiceAccount         string                 `yaml:"service_account,omitempty"`
	Labels                 map[string]string      `yaml:"labels,omitempty"`
}

func prepareAppSourceZip(t *testing.T) (zipFilePath string, tempDirPath string, err error) {
	tempDir, err := os.MkdirTemp("", "appengine-source-")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	t.Logf("Created temporary directory for app source: %s", tempDir)

	baseURL := "https://raw.githubusercontent.com/GoogleCloudPlatform/python-docs-samples/main/appengine/flexible/hello_world/"
	filesToIncludeInZip := map[string]string{
		"main.py":          "main.py",
		"requirements.txt": "requirements.txt",
		"app.yaml":         "app.yaml",
	}

	for localFileName := range filesToIncludeInZip {
		_, err := common_utils.DownloadFile(t, baseURL+localFileName, tempDir, localFileName)
		if err != nil {
			_ = os.RemoveAll(tempDir)
			return "", tempDir, fmt.Errorf("failed to download %s: %w", localFileName, err)
		}
	}
	zipOutputFilePath := filepath.Join(tempDir, "app_source.zip")
	err = common_utils.CreateZipArchive(t, tempDir, zipOutputFilePath, filesToIncludeInZip)
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return "", tempDir, fmt.Errorf("failed to create zip archive: %w", err)
	}
	return zipOutputFilePath, tempDir, nil
}

func getBaseAppEngineConfig(t *testing.T) AppEngineConfig {
	return AppEngineConfig{
		Project: projectID,
		Runtime: "python",
		FlexibleRuntimeSettings: struct {
			OperatingSystem string `yaml:"operating_system"`
			RuntimeVersion  string `yaml:"runtime_version"`
		}{
			OperatingSystem: "ubuntu22",
			RuntimeVersion:  "3.12",
		},
		Network: struct {
			Name       string `yaml:"name"`
			Subnetwork string `yaml:"subnetwork"`
		}{
			Name:       networkName,
			Subnetwork: fmt.Sprintf("%s-subnet", networkName),
		},
		VersionID: "v1",
		ManualScaling: &struct {
			Instances int `yaml:"instances"`
		}{
			Instances: 1,
		},
		Entrypoint: struct {
			Shell string `yaml:"shell"`
		}{
			Shell: "pip3 install gunicorn flask && gunicorn -b :8080 main:app"},
		LivenessCheck: map[string]interface{}{
			"path":          "/",
			"initial_delay": "300s",
		},
		ReadinessCheck: map[string]interface{}{
			"path":              "/",
			"app_start_timeout": "300s",
		},
		DeleteServiceOnDestroy: true,
	}
}

func createConfigYAML(t *testing.T, currentSaEmail string, currentGcsSourceURL string) []AppEngineConfig {
	t.Log("Generating YAML configuration for a single service, with dynamic SA and GCS URL...")
	baseConfig := getBaseAppEngineConfig(t)
	service1Config := baseConfig
	service1Config.Service = "test-service1"
	service1Config.Deployment = &struct {
		Zip *struct {
			SourceURL string `yaml:"source_url"`
		} `yaml:"zip,omitempty"`
	}{
		Zip: &struct {
			SourceURL string `yaml:"source_url"`
		}{
			SourceURL: currentGcsSourceURL,
		},
	}
	service1Config.ServiceAccount = currentSaEmail
	service1Config.EnvVariables = map[string]string{
		"SERVICE_ID": "1",
		"GCS_SOURCE": currentGcsSourceURL,
	}
	service1Config.Labels = map[string]string{
		"test-type":   "integration",
		"managed-by":  "terratest",
		"specific-to": "service1",
	}
	servicesToCreate := []AppEngineConfig{service1Config}
	if err := os.RemoveAll(configFolderPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to clean config directory %s: %v", configFolderPath, err)
	}
	if err := os.MkdirAll(configFolderPath, 0755); err != nil {
		t.Fatalf("Failed to create config directory %s: %v", configFolderPath, err)
	}
	serviceCfg := servicesToCreate[0]
	yamlData, err := yaml.Marshal(&serviceCfg)
	if err != nil {
		t.Fatalf("Error marshaling YAML for service %s: %v", serviceCfg.Service, err)
	}
	filePath := filepath.Join(configFolderPath, fmt.Sprintf("%s.yaml", serviceCfg.Service))
	err = os.WriteFile(filePath, yamlData, 0644)
	if err != nil {
		t.Fatalf("Error writing YAML file for service %s at %s: %v", serviceCfg.Service, filePath, err)
	}
	t.Logf("Created YAML config file: %s\n--- Content ---\n%s\n---------------", filePath, string(yamlData))
	return servicesToCreate
}

func TestCreateAppEngine(t *testing.T) {
	if projectID == "" {
		t.Fatal("TF_VAR_project_id environment variable must be set")
	}
	gcsBucketName = fmt.Sprintf("bkt-%s-%s", strings.ToLower(projectID), instanceName)
	gcsBucketName = strings.ReplaceAll(gcsBucketName, "_", "-")
	if len(gcsBucketName) > 63 {
		gcsBucketName = gcsBucketName[:63]
	}

	t.Logf("Test Run Config: ProjectID=%s, InstanceSuffix=%s, Network=%s, Bucket=%s",
		projectID, instanceName, networkName, gcsBucketName)

	appEngineDefaultSA := fmt.Sprintf("%s@appspot.gserviceaccount.com", projectID)
	common_utils.CreateGcsBucket(t, projectID, gcsBucketName, region)
	defer common_utils.DeleteGcsBucket(t, gcsBucketName)

	localZipPath, sourceTempDir, err := prepareAppSourceZip(t)
	if err != nil {
		t.Fatalf("Failed to prepare app source zip: %v", err)
	}
	defer common_utils.RemoveTempDir(t, sourceTempDir)

	fileInfo, statErr := os.Stat(localZipPath)
	if statErr != nil {
		t.Fatalf("Zip file expected at %s was not found: %v", localZipPath, statErr)
	}
	if fileInfo.Size() == 0 {
		t.Fatalf("Zip file at %s is empty. Size: %d bytes", localZipPath, fileInfo.Size())
	}
	t.Logf("Local zip file %s prepared successfully, size: %d bytes. Proceeding to upload.", localZipPath, fileInfo.Size())
	gcsSourceURL, err = common_utils.UploadGCSObjectFromFile(t, projectID, localZipPath, gcsBucketName, "app_source.zip")
	if err != nil {
		t.Fatalf("Failed to upload app source to GCS: %v", err)
	}
	defer common_utils.DeleteGcsObjects(t, gcsBucketName, "")
	t.Logf("App source uploaded to: %s", gcsSourceURL)
	generatedConfigs := createConfigYAML(t, appEngineDefaultSA, gcsSourceURL)
	if len(generatedConfigs) == 0 {
		t.Fatal("No YAML configurations were generated.")
	}

	tfVars := map[string]interface{}{"config_folder_path": configFolderPath}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		Vars: tfVars, TerraformDir: terraformDirectoryPath, Reconfigure: true, Lock: true, NoColor: true, SetVarsAfterVarFiles: true,
	})

	common_utils.CreateVPCSubnets(t, projectID, networkName, subnetName, region)
	defer common_utils.DeleteVPCSubnets(t, projectID, networkName, subnetName, region)
	t.Log("VPC/Subnet created. Waiting for propagation...")
	time.Sleep(60 * time.Second)
	if !common_utils.CreateFirewallRules(t, projectID, networkName, instanceName) {
		t.Fatal("Firewall rule creation failed.")
	}
	defer common_utils.DeleteFirewallRules(t, projectID, instanceName)
	t.Log("Firewall rules created.")

	defer terraform.Destroy(t, terraformOptions)
	t.Log("Running terraform init and apply...")
	terraform.InitAndApply(t, terraformOptions)
	t.Log("Terraform apply complete.")
	t.Log("====== Starting Verification from Live Deployed Version. =======")
	t.Log("Fetching Terraform output for variable 'instance_service_urls'...")
	instanceServiceURLsOutput := terraform.OutputJson(t, terraformOptions, "instance_service_urls")
	instanceServiceURLsMap := gjson.Parse(instanceServiceURLsOutput).Map()
	if len(instanceServiceURLsMap) == 0 {
		t.Fatal("No instances found in 'instance_service_urls' output.")
	}
	t.Logf("Found %d instance outputs for verification.", len(instanceServiceURLsMap))

	maxRetries := 7
	retryInterval := 1 * time.Minute
	verifiedServiceCount := 0

	for instanceKey, serviceURLMapResult := range instanceServiceURLsMap {
		t.Logf("--- Verifying Instance from Output Key: %s ---", instanceKey)
		serviceURLMap := serviceURLMapResult.Map()
		if len(serviceURLMap) == 0 {
			t.Errorf("No service URLs for instance key %s.", instanceKey)
			continue
		}

		var serviceName string
		for sn := range serviceURLMap {
			serviceName = sn
			break
		}
		if serviceName == "" {
			t.Errorf("Could not extract service name for instance key %s.", instanceKey)
			continue
		}
		t.Logf("Verifying service: %s", serviceName)

		var expectedConfig *AppEngineConfig
		for i := range generatedConfigs {
			if generatedConfigs[i].Service == serviceName && generatedConfigs[i].Project == projectID {
				expectedConfig = &generatedConfigs[i]
				break
			}
		}
		if expectedConfig == nil {
			t.Errorf("No matching YAML config for service %s in project %s.", serviceName, projectID)
			continue
		}

		serviceIsReady := false
		versionID := expectedConfig.VersionID
		for i := 0; i < maxRetries; i++ {
			t.Logf("Describing %s/%s (Attempt %d/%d)...", serviceName, versionID, i+1, maxRetries)
			cmd := shell.Command{Command: "gcloud", Args: []string{"app", "versions", "describe", versionID, "--service", serviceName, "--verbosity=none", "--project", projectID, "--format", "json"}}
			gcloudOutput, errCmd := shell.RunCommandAndGetOutputE(t, cmd)
			if errCmd != nil {
				t.Logf("gcloud error for %s/%s: %v. Retrying...", serviceName, versionID, errCmd)
				time.Sleep(retryInterval)
				continue
			}
			if gcloudOutput == "" {
				t.Logf("Empty gcloud output for %s/%s. Retrying...", serviceName, versionID)
				time.Sleep(retryInterval)
				continue
			}
			t.Logf("gcloud Output : %s", gcloudOutput)
			actualServiceInfo := gjson.Parse(gcloudOutput)
			t.Logf("Actual Service Info : %s", actualServiceInfo)
			status := gjson.Get(actualServiceInfo.String(), "servingStatus").String()
			t.Logf("Status for %s/%s: %s", serviceName, versionID, status)
			if status == "SERVING" {
				t.Logf("--> Verifying 'runtime' field...")
				t.Logf("Service %s/%s is SERVING. Performing assertions...", serviceName, versionID)
				actualRuntime := actualServiceInfo.Get("runtime").String()
				t.Logf("...... Expected Runtime: %s", expectedConfig.Runtime)
				t.Logf("...... Found Runtime:    %s", actualRuntime)
				assert.Equal(t, expectedConfig.Runtime, actualRuntime, "Runtime mismatch")
				t.Logf("...... ✓ Runtime verification successful.")

				t.Logf("--> Verifying 'instanceClass' field...")
				actualInstanceClass := actualServiceInfo.Get("instanceClass").String()
				t.Logf("...... Expected Instance Class: %s", expectedConfig.InstanceClass)
				t.Logf("...... Found Instance Class:    %s", actualInstanceClass)
				assert.Equal(t, expectedConfig.InstanceClass, actualInstanceClass, "InstanceClass mismatch")
				t.Logf("...... ✓ Instance class verification successful.")

				t.Logf("--> Verifying 'serviceAccount' field...")
				actualServiceAccount := actualServiceInfo.Get("serviceAccount").String()
				t.Logf("...... Expected Service Account: %s", expectedConfig.ServiceAccount)
				t.Logf("...... Found Service Account:    %s", actualServiceAccount)
				assert.Equal(t, expectedConfig.ServiceAccount, actualServiceAccount, "ServiceAccount mismatch")
				t.Logf("...... ✓ Service account verification successful.")

				t.Logf("--> Verifying 'network name' field...")
				actualNetworkName := actualServiceInfo.Get("network.name").String()
				t.Logf("...... Expected Network Name: %s", expectedConfig.Network.Name)
				t.Logf("...... Found Network Name:    %s", actualNetworkName)
				assert.Equal(t, expectedConfig.Network.Name, actualNetworkName, "Network name mismatch")
				t.Logf("...... ✓ Network name verification successful.")

				t.Logf("--> Verifying 'subnetwork name' field...")
				actualSubnetworkPath := actualServiceInfo.Get("network.subnetworkName").String()
				t.Logf("...... Expected Subnetwork Name: %s", expectedConfig.Network.Subnetwork)
				t.Logf("...... Found Subnetwork Name:    %s", actualSubnetworkPath)
				assert.Equal(t, expectedConfig.Network.Subnetwork, actualSubnetworkPath, "Subnetwork name mismatch")
				t.Logf("...... ✓ Subnetwork name verification successful.")
				if expectedConfig.Deployment != nil && expectedConfig.Deployment.Zip != nil {
					t.Logf("Verifying deployment source URL for %s/%s against expected %s (actual field may vary).",
						serviceName, versionID, expectedConfig.Deployment.Zip.SourceURL)
				}

				serviceIsReady = true
				verifiedServiceCount++
				break
			}
			t.Logf("Waiting %v for %s/%s...", retryInterval, serviceName, versionID)
			time.Sleep(retryInterval)
		}
		if !serviceIsReady {
			t.Fatalf("Service %s/%s did not reach SERVING after %d retries.", serviceName, versionID, maxRetries)
		}
	}
	assert.Equal(t, len(generatedConfigs), verifiedServiceCount, "Number of verified services did not match generated configs.")
	t.Logf("Successfully verified %d service(s).", verifiedServiceCount)
	t.Log("Test completed. Cleanup will run via deferred calls.")
}
