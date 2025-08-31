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
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	// Import the common_utils package
	"github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/test/integration/common_utils"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

const (
	defaultRegion                = "us-central1"
	testSubnetCIDR               = "10.5.4.0/28"     // Subnet CIDR for the subnet that will host the connector
	httpCheckRetries             = 10                // Retries for HTTP check
	httpCheckInterval            = 150 * time.Second // Interval for HTTP check
	iamPropagationWait           = 30 * time.Second  // Wait after adding IAM bindings
	apiEnablementPropagationWait = 30 * time.Second  // Wait after enabling APIs
	sampleAppRuntime             = "python311"
	sampleAppEntrypoint          = "gunicorn -b :$PORT main:app"
	sampleAppGcsObjectName       = "app.yaml"
	sampleAppGcsMainPyName       = "main.py"
	service1                     = "service1"
)

var (
	uniqueID               = strings.ToLower(random.UniqueId())
	projectRoot, _         = filepath.Abs("../../../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "06-consumer/Serverless/AppEngine/Standard")
	configFolderPath       = filepath.Join(projectRoot, "test/integration/consumer/Serverless/AppEngine/Standard/config")
	projectID              = os.Getenv("TF_VAR_project_id")
	versionID1             = fmt.Sprintf("v1-%s", uniqueID)
	sampleAppGcsBucket     = common_utils.GetEnv("TF_VAR_test_gcs_bucket", fmt.Sprintf("%s-bucket-%s", projectID, uniqueID))
)

type AppEngineConfig struct {
	ProjectID              string                  `yaml:"project_id"`
	Service                string                  `yaml:"service"`
	VersionID              string                  `yaml:"version_id"`
	Runtime                string                  `yaml:"runtime"`
	Deployment             *DeploymentConfig       `yaml:"deployment,omitempty"`
	Entrypoint             *EntrypointConfig       `yaml:"entrypoint,omitempty"`
	AutomaticScaling       *AutomaticScalingConfig `yaml:"automatic_scaling,omitempty"`
	Handlers               HandlerConfig           `yaml:"handlers,omitempty"`
	AppEngineApplication   *AppEngineAppConfig     `yaml:"app_engine_application,omitempty"`
	DeleteServiceOnDestroy bool                    `yaml:"delete_service_on_destroy,omitempty"`
}
type DeploymentConfig struct {
	Files *FilesConfig `yaml:"files,omitempty"`
}
type FilesConfig struct {
	Name      string `yaml:"name"`
	SourceURL string `yaml:"source_url"`
}
type EntrypointConfig struct {
	Shell string `yaml:"shell"`
}
type AutomaticScalingConfig struct {
	MaxConcurrentRequests int `yaml:"max_concurrent_requests,omitempty"`
	MinIdleInstances      int `yaml:"min_idle_instances,omitempty"`
	MaxIdleInstances      int `yaml:"max_idle_instances,omitempty"`
}
type HandlerConfig []struct {
	URLRegex string        `yaml:"url_regex"`
	Script   *ScriptConfig `yaml:"script,omitempty"`
}
type ScriptConfig struct {
	ScriptPath string `yaml:"script_path"`
}
type VPCAccessConnectorConfig struct {
	Name string `yaml:"name"`
}
type AppEngineAppConfig struct {
	LocationID string `yaml:"location_id"`
}
type DomainMappingConfig struct {
	DomainName string `yaml:"domain_name"`
}
type FirewallRuleConfig struct {
	SourceRange string `yaml:"source_range"`
	Action      string `yaml:"action"`
}

func TestAppEngineStandardIntegration(t *testing.T) {
	t.Parallel()

	if projectID == "" {
		t.Fatal("TF_VAR_project_id environment variable must be set")
	}
	testConfigFolderPath := configFolderPath
	testGcsObjectPathPrefix := fmt.Sprintf("app-test-%s", uniqueID)

	t.Logf("Ensuring App Engine application exists in project '%s' for region '%s'...", projectID, defaultRegion)
	common_utils.EnsureAppEngineApplicationExists(t, projectID, defaultRegion)

	t.Logf("Creating GCS bucket: %s", sampleAppGcsBucket)
	common_utils.CreateGcsBucket(t, projectID, sampleAppGcsBucket, defaultRegion)
	defer common_utils.DeleteGcsBucket(t, sampleAppGcsBucket)
	defer common_utils.DeleteGcsObjects(t, sampleAppGcsBucket, testGcsObjectPathPrefix+"/")

	appYamlContent, mainPyContent := getHelloWorldAppFiles()
	appYamlGcsPath := path.Join(testGcsObjectPathPrefix, sampleAppGcsObjectName)
	mainPyGcsPath := path.Join(testGcsObjectPathPrefix, sampleAppGcsMainPyName)
	common_utils.UploadGcsObjectFromString(t, projectID, sampleAppGcsBucket, appYamlGcsPath, appYamlContent)
	common_utils.UploadGcsObjectFromString(t, projectID, sampleAppGcsBucket, mainPyGcsPath, mainPyContent)

	appYamlDisplayUrl := fmt.Sprintf("https://storage.googleapis.com/%s/%s", sampleAppGcsBucket, appYamlGcsPath)

	generatedConfigs := createConfigYAML(t, testConfigFolderPath, uniqueID, appYamlDisplayUrl)

	tfVars := map[string]interface{}{"config_folder_path": testConfigFolderPath}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
		Reconfigure:  true,
		Lock:         true,
		NoColor:      true,
	})

	defer terraform.Destroy(t, terraformOptions)
	t.Logf("====== Running terraform init & apply... ======")
	_, err := terraform.InitAndApplyE(t, terraformOptions)
	if err != nil {
		t.Fatalf("Terraform init/apply failed: %v", err)
	}
	t.Logf("Terraform apply completed successfully.")

	t.Log("====== Starting Verification of Terraform Outputs. =======")

	t.Logf("Fetching Terraform output for variable 'app_engine_standard'...")
	appEngineOutputValue := terraform.OutputJson(t, terraformOptions, "app_engine_standard")
	if !gjson.Valid(appEngineOutputValue) {
		t.Fatalf("Error parsing output, invalid json: %s", appEngineOutputValue)
	}
	result := gjson.Parse(appEngineOutputValue)
	if len(generatedConfigs) == 0 {
		t.Fatal("No configurations were generated by createConfigYAML")
	}
	for i, expectedConfig := range generatedConfigs {
		instanceKey := fmt.Sprintf("instance%d", i+1)

		t.Logf(" ========= Verifying Service: %s ========= ", expectedConfig.Service)

		t.Logf("--> Verifying 'id' field...")
		idPath := fmt.Sprintf("%s.app_engine_standard.%s.id", instanceKey, expectedConfig.Service)
		expectedID := fmt.Sprintf("apps/%s/services/%s/versions/%s", projectID, expectedConfig.Service, expectedConfig.VersionID)
		actualID := gjson.Get(result.String(), idPath).String()
		t.Logf("...... Searching for path: %s", idPath)
		t.Logf("...... Expected ID: %s", expectedID)
		t.Logf("...... Found ID:    %s", actualID)
		if expectedID != actualID {
			t.Fatalf("ID mismatch for service %s.\nExpected: %s\nActual:   %s", expectedConfig.Service, expectedID, actualID)
		}
		t.Logf("...... ✓ ID verification successful.")

		t.Logf("--> Verifying 'version_id' field...")
		versionIDPath := fmt.Sprintf("%s.app_engine_standard.%s.version_id", instanceKey, expectedConfig.Service)
		actualVersionID := gjson.Get(result.String(), versionIDPath).String()
		t.Logf("...... Searching for path: %s", versionIDPath)
		t.Logf("...... Expected Version ID: %s", expectedConfig.VersionID)
		t.Logf("...... Found Version ID:    %s", actualVersionID)
		if expectedConfig.VersionID != actualVersionID {
			t.Fatalf("Version ID mismatch for service %s.\nExpected: %s\nActual:   %s", expectedConfig.Service, expectedConfig.VersionID, actualVersionID)
		}
		t.Logf("...... ✓ Version ID verification successful.")

		t.Logf("--> Verifying 'runtime' field...")
		runtimePath := fmt.Sprintf("%s.app_engine_standard.%s.runtime", instanceKey, expectedConfig.Service)
		actualRuntime := gjson.Get(result.String(), runtimePath).String()
		t.Logf("...... Searching for path: %s", runtimePath)
		t.Logf("...... Expected Runtime: %s", sampleAppRuntime)
		t.Logf("...... Found Runtime:    %s", actualRuntime)
		if sampleAppRuntime != actualRuntime {
			t.Fatalf("Runtime mismatch for service %s.\nExpected: %s\nActual:   %s", expectedConfig.Service, sampleAppRuntime, actualRuntime)
		}
		t.Logf("...... ✓ Runtime verification successful.")

		t.Logf("--> Verifying 'service_url' field...")
		serviceURLPath := fmt.Sprintf("%s.service_urls.%s", instanceKey, expectedConfig.Service)
		serviceURL := gjson.Get(result.String(), serviceURLPath).String()
		t.Logf("...... Searching for path: %s", serviceURLPath)
		t.Logf("...... Found URL: %s", serviceURL)
		if serviceURL == "" {
			t.Fatalf("URL for service %s should not be empty", expectedConfig.Service)
		}
		t.Logf("...... ✓ URL is not empty.")
	}

	t.Log("====== Test AppEngine Standard Integration Completed. ====")
}

func createConfigYAML(t *testing.T, outputDir string, uniqueID string, deploymentFileURL string) []AppEngineConfig {
	t.Helper()
	t.Logf("Checking if 'default' service exists in project '%s'...", projectID)
	listCmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"app", "services", "list", "--project=" + projectID},
	}
	output, err := shell.RunCommandAndGetOutputE(t, listCmd)
	if err != nil {
		t.Fatalf("Failed to check for default service: %v. Output: %s", err, output)
	}
	defaultServiceExists := false
	if output != "" {
		services := strings.Fields(output)
		for _, s := range services {
			if s == "default" {
				defaultServiceExists = true
				break
			}
		}
	}
	config := AppEngineConfig{
		ProjectID:              projectID,
		Service:                service1,
		VersionID:              versionID1,
		Runtime:                sampleAppRuntime,
		Deployment:             &DeploymentConfig{Files: &FilesConfig{Name: sampleAppGcsObjectName, SourceURL: deploymentFileURL}},
		Entrypoint:             &EntrypointConfig{Shell: sampleAppEntrypoint},
		AutomaticScaling:       &AutomaticScalingConfig{MaxConcurrentRequests: 50, MinIdleInstances: 1, MaxIdleInstances: 3},
		Handlers:               HandlerConfig{{URLRegex: "/.*", Script: &ScriptConfig{ScriptPath: "auto"}}},
		DeleteServiceOnDestroy: true,
		AppEngineApplication:   &AppEngineAppConfig{LocationID: defaultRegion},
	}
	if !defaultServiceExists {
		t.Logf("'default' service not found. Creating a single config for service 'default'.")
		config.Service = "default"
		config.DeleteServiceOnDestroy = false
	} else {
		t.Logf("'default' service found. Creating standard config for '%s'.", service1)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory '%s': %v", outputDir, err)
	}
	yamlData, err := yaml.Marshal(&config)
	if err != nil {
		t.Fatalf("Error marshaling config: %v", err)
	}
	filePath := filepath.Join(outputDir, "instance1.yaml")
	if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
		t.Fatalf("Unable to write data into file %s: %v", filePath, err)
	}
	t.Logf("Created YAML config at %s", filePath)
	return []AppEngineConfig{config}
}

func getHelloWorldAppFiles() (appYamlContent string, mainPyContent string) {
	appYamlContent = fmt.Sprintf("runtime: %s\nentrypoint: %s\n", sampleAppRuntime, sampleAppEntrypoint)
	mainPyContent = `from flask import Flask
import os

app = Flask(__name__)

@app.route('/')
def hello():
    """Return a friendly HTTP greeting."""
    app_mode = os.environ.get("APP_MODE", "production")
    return f'Hello World! Mode: {app_mode}\nYour AppEngine Version is: {os.environ.get("GAE_VERSION", "unknown")}\n'

if __name__ == '__main__':
    app.run(host='127.0.0.1', port=int(os.environ.get("PORT", 8080)), debug=True)
`
	return appYamlContent, mainPyContent
}
