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
	region                    = "us-central1"
	zone                      = "us-central1-a"
	onpremDnsServerIP         = "192.168.1.100"
	yamlFileName              = "dns.yaml"
	privateZoneDescription    = "Private zone for corp services"
	forwardingZoneDescription = "Forwarding zone to on-prem"
	peeringZoneDescription    = "Peering zone with peer network"
	visibilityPrivate         = "private"
	visibilityForwarding      = "forwarding"
	visibilityPeering         = "peering"
	recordTypeA               = "A"
	recordTypeCNAME           = "CNAME"
	recordTTL                 = 300
	recordAData               = "10.0.0.10"
	onpremDnsServerIP2        = "192.168.1.101"
	connectivityTestName      = "reachability-test"
)

var (
	projectID, _           = os.LookupEnv("TF_VAR_project_id")
	uniqueID               = rand.Int()
	projectRoot, _         = filepath.Abs("../../../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "execution/02-networking/CloudDNS/DNSManagedZones")
	configFolderPath       = filepath.Join(projectRoot, "execution/test/integration/networking/CloudDNS/DNSManagedZones/config")

	// GCP network
	gcpNetworkName    = fmt.Sprintf("gcp-vpc-dns-%d", uniqueID)
	gcpSubnetworkName = fmt.Sprintf("gcp-subnet-dns-%d", uniqueID)

	// On-prem simulation network
	onpremNetworkName    = fmt.Sprintf("onprem-vpc-%d", uniqueID)
	onpremSubnetworkName = fmt.Sprintf("onprem-subnet-%d", uniqueID)

	// Peer network
	peerNetworkName    = fmt.Sprintf("peer-vpc-dns-%d", uniqueID)
	peerSubnetworkName = fmt.Sprintf("peer-subnet-dns-%d", uniqueID)

	// DNS names
	privateZoneName    = fmt.Sprintf("private-zone-%d", uniqueID)
	privateDomain      = "corp.internal."
	forwardingZoneName = fmt.Sprintf("forwarding-zone-%d", uniqueID)
	forwardingDomain   = "onprem.internal."
	peeringZoneName    = fmt.Sprintf("peering-zone-%d", uniqueID)
	peeringDomain      = "peered.internal."
)

// Structs for DNS YAML configuration
type DNSConfig struct {
	Zones []Zone `yaml:"zones"`
}

type Zone struct {
	Name         string      `yaml:"zone"`
	ProjectID    string      `yaml:"project_id"`
	Description  string      `yaml:"description"`
	ForceDestroy bool        `yaml:"force_destroy"`
	ZoneConfig   ZoneConfig  `yaml:"zone_config"`
	Recordsets   []RecordSet `yaml:"recordsets,omitempty"`
}

type ZoneConfig struct {
	Domain                  string                   `yaml:"domain"`
	Visibility              string                   `yaml:"visibility"`
	ReverseLookup           bool                     `yaml:"reverse_lookup,omitempty"`
	PrivateVisibilityConfig *PrivateVisibilityConfig `yaml:"private_visibility_config,omitempty"`
	ForwardingConfig        *ForwardingConfig        `yaml:"forwarding_config,omitempty"`
	PeeringConfig           *PeeringConfig           `yaml:"peering_config,omitempty"`
}

type PrivateVisibilityConfig struct {
	Networks []Network `yaml:"networks"`
}

type ForwardingConfig struct {
	TargetNameServers []TargetNameServer `yaml:"target_name_servers"`
}

type TargetNameServer struct {
	Ipv4Address string `yaml:"ipv4_address"`
}

type PeeringConfig struct {
	TargetNetwork Network `yaml:"target_network"`
}

type Network struct {
	NetworkURL string `yaml:"network_url"`
}

type RecordSet struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	TTL     int      `yaml:"ttl"`
	Records []string `yaml:"records"`
}

func TestCloudDNSUserJourney(t *testing.T) {
	if projectID == "" {
		t.Fatal("TF_VAR_project_id environment variable must be set.")
	}
	t.Parallel()

	// Setup: Create networks using common utils
	common_utils.CreateVPCSubnets(t, projectID, gcpNetworkName, gcpSubnetworkName, region)
	defer common_utils.DeleteVPCSubnets(t, projectID, gcpNetworkName, gcpSubnetworkName, region)

	common_utils.CreateVPCSubnets(t, projectID, onpremNetworkName, onpremSubnetworkName, region)
	defer common_utils.DeleteVPCSubnets(t, projectID, onpremNetworkName, onpremSubnetworkName, region)

	common_utils.CreateVPCSubnets(t, projectID, peerNetworkName, peerSubnetworkName, region)
	defer common_utils.DeleteVPCSubnets(t, projectID, peerNetworkName, peerSubnetworkName, region)

	// Setup: Create YAML config file
	createConfigYAMLDNS(t)

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
	t.Log("Verifying DNS Managed Zone and Record Set configurations...")

	// Load expected config from YAML
	configPath := filepath.Join(configFolderPath, yamlFileName)
	yamlBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config YAML: %v", err)
	}
	var config DNSConfig
	if err := yaml.Unmarshal(yamlBytes, &config); err != nil {
		t.Fatalf("Failed to unmarshal config YAML: %v", err)
	}

	for _, zone := range config.Zones {
		expectedVisibility := zone.ZoneConfig.Visibility
		t.Logf("Checking zone %s (expected visibility: %s)", zone.Name, expectedVisibility)
		verifyManagedZone(t, zone.Name, zone.ZoneConfig.Domain, expectedVisibility, zone.ZoneConfig)
		for _, rs := range zone.Recordsets {
			verifyRecordSet(t, zone.Name, rs.Name, rs.Type, rs.TTL, rs.Records)
		}
	}
}

// createConfigYAMLDNS generates a YAML configuration file for the DNS zones.
func createConfigYAMLDNS(t *testing.T) {
	config := DNSConfig{
		Zones: []Zone{
			{
				Name:         privateZoneName,
				ProjectID:    projectID,
				Description:  privateZoneDescription,
				ForceDestroy: true,
				ZoneConfig: ZoneConfig{
					Domain:     privateDomain,
					Visibility: visibilityPrivate,
					PrivateVisibilityConfig: &PrivateVisibilityConfig{
						Networks: []Network{
							{NetworkURL: fmt.Sprintf("projects/%s/global/networks/%s", projectID, gcpNetworkName)},
							{NetworkURL: fmt.Sprintf("projects/%s/global/networks/%s", projectID, peerNetworkName)},
						},
					},
				},
				Recordsets: []RecordSet{
					{
						Name:    fmt.Sprintf("app.%s", privateDomain),
						Type:    recordTypeA,
						TTL:     recordTTL,
						Records: []string{recordAData},
					},
					{
						Name:    fmt.Sprintf("alias.%s", privateDomain),
						Type:    recordTypeCNAME,
						TTL:     recordTTL,
						Records: []string{fmt.Sprintf("app.%s", privateDomain)},
					},
				},
			},
			{
				Name:         forwardingZoneName,
				ProjectID:    projectID,
				Description:  forwardingZoneDescription,
				ForceDestroy: true,
				ZoneConfig: ZoneConfig{
					Domain:     forwardingDomain,
					Visibility: visibilityForwarding,
					ForwardingConfig: &ForwardingConfig{
						TargetNameServers: []TargetNameServer{
							{Ipv4Address: onpremDnsServerIP},
							{Ipv4Address: onpremDnsServerIP2},
						},
					},
				},
			},
			{
				Name:         peeringZoneName,
				ProjectID:    projectID,
				Description:  peeringZoneDescription,
				ForceDestroy: true,
				ZoneConfig: ZoneConfig{
					Domain:     peeringDomain,
					Visibility: visibilityPeering,
					PeeringConfig: &PeeringConfig{
						TargetNetwork: Network{NetworkURL: fmt.Sprintf("projects/%s/global/networks/%s", projectID, peerNetworkName)},
					},
				},
			},
		},
	}

	yamlData, err := yaml.Marshal(&config)
	if err != nil {
		t.Fatalf("Error while marshaling DNS config: %v", err)
	}

	if err := os.MkdirAll(configFolderPath, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	filePath := filepath.Join(configFolderPath, yamlFileName)
	if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
		t.Fatalf("Unable to write DNS config to file: %v", err)
	}
	t.Logf("Created DNS YAML config at %s", filePath)
}

// verifyManagedZone checks that a managed zone exists and has the correct configuration based on logs.
func verifyManagedZone(t *testing.T, zoneName, domain, expectedVisibility string, zoneConfig ZoneConfig) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"dns", "managed-zones", "describe", zoneName, "--project", projectID, "--format", "json"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Failed to describe managed zone %s: %v", zoneName, err)
	}

	// Name check
	actualName := gjson.Get(output, "name").String()
	if actualName == zoneName {
		t.Logf("[PASS] Managed zone name: expected=%s, actual=%s", zoneName, actualName)
	} else {
		t.Errorf("[FAIL] Managed zone name: expected=%s, actual=%s", zoneName, actualName)
	}

	// Domain check
	actualDomain := gjson.Get(output, "dnsName").String()
	if actualDomain == domain {
		t.Logf("[PASS] DNS domain name: expected=%s, actual=%s", domain, actualDomain)
	} else {
		t.Errorf("[FAIL] DNS domain name: expected=%s, actual=%s", domain, actualDomain)
	}
}

// verifyRecordSet checks that a record set exists and has the correct configuration.
func verifyRecordSet(t *testing.T, zoneName, recordName, expectedType string, expectedTTL int, expectedRecords []string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"dns", "record-sets", "list", "--zone", zoneName, "--project", projectID, "--format", "json"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Failed to list record sets for zone %s: %v", zoneName, err)
	}

	record := gjson.Get(output, fmt.Sprintf("#(name==\"%s\")", recordName))
	assert.True(t, record.Exists(), fmt.Sprintf("Record set %s not found in zone %s", recordName, zoneName))

	actualType := record.Get("type").String()
	actualTTL := int(record.Get("ttl").Float())
	if actualType != expectedType {
		t.Errorf("Record type mismatch for %s: expected=%s, actual=%s", recordName, expectedType, actualType)
	}
	if actualTTL != expectedTTL {
		t.Errorf("Record TTL mismatch for %s: expected=%d, actual=%d", recordName, expectedTTL, actualTTL)
	}

	// Verify records
	actualRecords := []string{}
	record.Get("rrdatas").ForEach(func(_, value gjson.Result) bool {
		actualRecords = append(actualRecords, value.String())
		return true
	})
	for _, expected := range expectedRecords {
		if !contains(actualRecords, expected) {
			t.Errorf("Record data mismatch for %s: expected to contain %s, actual=%v", recordName, expected, actualRecords)
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
