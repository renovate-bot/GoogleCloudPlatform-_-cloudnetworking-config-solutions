# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Tests for the orchestrator.JsonGenerator class."""

from collections.abc import Mapping
import logging
from unittest import mock
import json

from pyfakefs import fake_filesystem_unittest

import unittest
from utilities import orchestrator


def _patch_loader(
    mock_loader_cls,
    schemas: Mapping[str, object] | None = None,
    defaults: Mapping[str, object] | None = None,
):
    """Helper function to configure mock loaders."""
    mock_loader_instance = mock_loader_cls.return_value
    if schemas:
        mock_loader_instance.load_schemas_for_resource_types.return_value = schemas
    if defaults:
        mock_loader_instance.load_defaults_for_resource_types.return_value = defaults


class JsonGeneratorTest(fake_filesystem_unittest.TestCase):
    """Tests for the orchestrator.JsonGenerator class."""

    def setUp(self):
        super().setUp()
        self.setUpPyfakefs()

        # Create a comprehensive supported_resources.json for the tests
        supported_resources_content = {
            "vm": {"uriTemplate": "projects/{projectId}/zones/{zone}/instances/{name}"},
            "network": {},
            "widget": {},
            "subnetwork": {},
            "vpc": {"nestedResources": {"subnets": "subnetwork"}},
            "router": {},
            "firewall_rule": {},
            "memorystore_redis_cluster": {"connectivityOptions": ["scp"]},
            "cloudsql": {"connectivityOptions": ["psc"]},
            "address": {},
            "forwardingrule": {},
        }
        self.fs.create_file(
            "supported_resources.json", contents=json.dumps(supported_resources_content)
        )
        self.fs.create_dir("schemas")
        self.fs.create_dir("defaults")

    def tearDown(self):
        super().tearDown()

    def test_process_explicit_resources_expands_count(self):
        """Tests that resources with a 'count' field are expanded correctly."""
        with (
            mock.patch.object(
                orchestrator, "SchemasLoader", spec_set=True
            ) as mock_schemas,
            mock.patch.object(
                orchestrator, "DefaultsJsonLoader", spec_set=True
            ) as mock_defaults,
        ):
            _patch_loader(
                mock_schemas,
                schemas={"vm": {"type": "object", "properties": {"name": None}}},
            )
            _patch_loader(mock_defaults, defaults={"vm": {}})
            basic_config = {
                "namePrefix": "test",
                "projects": [
                    {
                        "projectId": "proj-1",
                        "vms": [{"type": "vm", "name": "instance", "count": 3}],
                    }
                ],
            }
            generator = orchestrator.JsonGenerator(
                supported_resources_path="supported_resources.json",
                schemas_dir="schemas",
                defaults_dir="defaults",
            )

            result = generator.generate(basic_config)

            processed_vms = result["projects"][0]["vms"]
            self.assertEqual(len(processed_vms), 3)
            self.assertEqual(
                [vm["name"] for vm in processed_vms],
                ["test-instance-1", "test-instance-2", "test-instance-3"],
            )
            self.assertNotIn("count", processed_vms[0])

    @mock.patch.object(orchestrator.logger, "warning")
    def test_process_resource_instance_handles_item_with_no_type(self, mock_warning):
        """Verifies that a resource with a missing 'type' key is handled."""
        generator = orchestrator.JsonGenerator("supported_resources.json", "s", "d")
        untyped_item = {"name": "untyped-widget"}

        result = generator._process_resource_instance(untyped_item)
        self.assertEqual(result, untyped_item)
        mock_warning.assert_called_once()

    def test_generate_preloads_schemas_and_defaults(self):
        """Tests that generate() correctly pre-loads schemas and defaults."""
        basic_config = {
            "projects": [
                {"projectId": "test-proj", "vms": [{"type": "vm", "name": "instance"}]}
            ]
        }
        generator = orchestrator.JsonGenerator("supported_resources.json", "s", "d")

        with (
            mock.patch.object(
                generator._schemas_loader,
                "load_schemas_for_resource_types",
                return_value={},
            ) as mock_load_schemas,
            mock.patch.object(
                generator._defaults_loader,
                "load_defaults_for_resource_types",
                return_value={},
            ) as mock_load_defaults,
        ):
            generator.generate(basic_config)
            mock_load_schemas.assert_called_once_with(["vm"], mock.ANY)
            mock_load_defaults.assert_called_once_with(["vm"])

    def test_create_instance_from_schema_skips_readonly_properties(self):
        """Verifies that readOnly properties in a schema are skipped."""
        generator = orchestrator.JsonGenerator("supported_resources.json", "s", "d")
        schema_with_readonly = {
            "type": "object",
            "properties": {
                "normal_prop": {"type": "string"},
                "secret_prop": {"type": "string", "readOnly": True},
            },
        }

        result = generator._create_instance_from_schema(schema_with_readonly)

        self.assertIn("normal_prop", result)
        self.assertNotIn("secret_prop", result)

    def test_create_instance_from_schema_initializes_to_none(self):
        """Verifies that properties from the schema are initialized to None."""
        generator = orchestrator.JsonGenerator("supported_resources.json", "s", "d")
        schema = {
            "type": "object",
            "properties": {
                "prop1": {"type": "string"},
                "prop2": {"type": "integer"},
            },
        }
        result = generator._create_instance_from_schema(schema)

        self.assertIn("prop1", result)
        self.assertIsNone(result["prop1"])
        self.assertIn("prop2", result)
        self.assertIsNone(result["prop2"])

    def test_create_instance_from_schema_initializes_properties_to_none(self):
        """Verifies that properties from the schema are correctly initialized."""
        generator = orchestrator.JsonGenerator("supported_resources.json", "s", "d")
        schema = {
            "type": "object",
            "properties": {
                "prop1": {"type": "string"},
                "prop2": {"type": "integer"},
            },
        }

        result = generator._create_instance_from_schema(schema)

        self.assertIn("prop1", result)
        self.assertIsNone(result["prop1"])
        self.assertIn("prop2", result)
        self.assertIsNone(result["prop2"])

    def test_process_resource_instance_applies_defaults(self):
        """Verifies that default values are correctly layered onto a resource."""
        generator = orchestrator.JsonGenerator(
            supported_resources_path="supported_resources.json",
            schemas_dir="schemas",
            defaults_dir="defaults",
        )
        generator._schemas["vm"] = {
            "type": "object",
            "properties": {"name": None, "zone": None},
        }
        generator._defaults["vm"] = {"zone": "us-central1-c"}
        resource_data = {"type": "vm", "name": "test-vm"}
        result = generator._process_resource_instance(resource_data)

        self.assertEqual(result["zone"], "us-central1-c")

    def test_create_instance_from_schema_handles_non_dict_input(self):
        """Verifies _create_instance_from_schema returns {} for non-dict inputs."""
        generator = orchestrator.JsonGenerator(
            supported_resources_path="supported_resources.json",
            schemas_dir="schemas",
            defaults_dir="defaults",
        )
        invalid_list = []
        invalid_string = "I am not a schema"
        result_from_list = generator._create_instance_from_schema(invalid_list)
        result_from_string = generator._create_instance_from_schema(invalid_string)
        self.assertEqual(result_from_list, {})
        self.assertEqual(result_from_string, {})

    def test_derive_implicit_resources_creates_nat_router(self):
        """Verifies that a NAT router is derived for a VPC with createNat."""
        basic_config = {
            "defaultRegion": "us-central1",
            "projects": [
                {
                    "projectId": "host-project",
                    "vpc": [{"type": "vpc", "name": "my-net", "createNat": True}],
                }
            ],
        }
        with (
            mock.patch.object(
                orchestrator, "SchemasLoader", spec_set=True
            ) as mock_schemas,
            mock.patch.object(
                orchestrator, "DefaultsJsonLoader", spec_set=True
            ) as mock_defaults,
        ):
            _patch_loader(mock_schemas, schemas={"vpc": {}, "router": {}})
            _patch_loader(mock_defaults, defaults={"vpc": {}, "router": {}})
            generator = orchestrator.JsonGenerator(
                "supported_resources.json", "schemas", "defaults"
            )
            result = generator.generate(basic_config)
            host_project_config = result["projects"][0]
            self.assertIn("routers", host_project_config)
            self.assertEqual(len(host_project_config["routers"]), 1)
            self.assertEqual(
                host_project_config["routers"][0]["name"], "router-my-net-nat"
            )

    def test_derive_firewall_rule_with_explicit_network(self):
        """Verifies a firewall rule is derived using an explicit network key."""
        basic_config = {
            "projects": [
                {
                    "projectId": "network-project",
                    "vpc": [{"type": "vpc", "name": "my-net"}],
                },
                {
                    "projectId": "service-project",
                    "producers": [
                        {
                            "type": "cloudsql",
                            "name": "my-db",
                            "createRequiredFwRules": True,
                            "networkForFirewall": "my-net",
                            "allowedConsumersTags": ["allow-db-access"],
                        }
                    ],
                },
            ]
        }
        with (
            mock.patch.object(
                orchestrator, "SchemasLoader", spec_set=True
            ) as mock_schemas,
            mock.patch.object(
                orchestrator, "DefaultsJsonLoader", spec_set=True
            ) as mock_defaults,
        ):
            _patch_loader(
                mock_schemas,
                schemas={"cloudsql": {}, "firewall_rule": {}, "vpc": {}},
            )
            _patch_loader(
                mock_defaults,
                defaults={"cloudsql": {}, "firewall_rule": {}, "vpc": {}},
            )
            generator = orchestrator.JsonGenerator(
                "supported_resources.json", "schemas", "defaults"
            )
            result = generator.generate(basic_config)
            network_project_config = result["projects"][0]
            self.assertIn("firewalls", network_project_config)
            self.assertEqual(len(network_project_config["firewalls"]), 1)
            firewall_rule = network_project_config["firewalls"][0]
            self.assertEqual(firewall_rule["name"], "fw-allow-my-db")
            self.assertEqual(firewall_rule["targetTags"], ["allow-db-access"])

    def test_derive_firewall_rule_infers_network_from_psc_settings(self):
        """Verifies a firewall rule is derived by inferring from pscSettings."""
        basic_config = {
            "projects": [
                {
                    "projectId": "network-project",
                    "pscSettings": {"networkForPsc": "my-net"},
                    "vpc": [{"type": "vpc", "name": "my-net"}],
                },
                {
                    "projectId": "service-project",
                    "producers": [
                        {
                            "type": "cloudsql",
                            "name": "my-db",
                            "createRequiredFwRules": True,
                            "allowedConsumersTags": ["allow-db-access"],
                        }
                    ],
                },
            ]
        }
        with (
            mock.patch.object(
                orchestrator, "SchemasLoader", spec_set=True
            ) as mock_schemas,
            mock.patch.object(
                orchestrator, "DefaultsJsonLoader", spec_set=True
            ) as mock_defaults,
        ):
            _patch_loader(
                mock_schemas,
                schemas={"cloudsql": {}, "firewall_rule": {}, "vpc": {}},
            )
            _patch_loader(
                mock_defaults,
                defaults={"cloudsql": {}, "firewall_rule": {}, "vpc": {}},
            )
            generator = orchestrator.JsonGenerator(
                "supported_resources.json", "schemas", "defaults"
            )
            result = generator.generate(basic_config)
            network_project_config = result["projects"][0]
            self.assertIn("firewalls", network_project_config)
            self.assertEqual(len(network_project_config["firewalls"]), 1)
            self.assertEqual(
                network_project_config["firewalls"][0]["name"], "fw-allow-my-db"
            )

    def test_derive_implicit_resources_creates_scp_policy(self):
        """Verifies that SCP flags are correctly set on a VPC."""
        basic_config = {
            "projects": [
                {
                    "projectId": "service-project",
                    "vpc": [{"type": "vpc", "name": "my-scp-net"}],
                    "subnets": [
                        {
                            "type": "subnetwork",
                            "name": "my-scp-subnet",
                            "network": "my-scp-net",
                        }
                    ],
                    "producers": [
                        {
                            "type": "memorystore_redis_cluster",
                            "name": "my-mrc",
                            "connectivityType": "scp",
                            "subnet": "my-scp-subnet",
                        }
                    ],
                }
            ]
        }
        with (
            mock.patch.object(
                orchestrator, "SchemasLoader", spec_set=True
            ) as mock_schemas,
            mock.patch.object(
                orchestrator, "DefaultsJsonLoader", spec_set=True
            ) as mock_defaults,
        ):
            _patch_loader(
                mock_schemas,
                schemas={
                    "memorystore_redis_cluster": {},
                    "vpc": {},
                    "subnetwork": {},
                },
            )
            _patch_loader(
                mock_defaults,
                defaults={
                    "memorystore_redis_cluster": {},
                    "vpc": {},
                    "subnetwork": {},
                },
            )
            generator = orchestrator.JsonGenerator(
                "supported_resources.json", "schemas", "defaults"
            )
            result = generator.generate(basic_config)
            vpc_config = result["projects"][0]["vpc"][0]
            self.assertTrue(vpc_config["createScpPolicy"])
            self.assertEqual(vpc_config["subnetsForScpPolicy"], ["my-scp-subnet"])

    def test_derive_psc_endpoint_from_per_producer_settings(self):
        """Verifies PSC endpoint is created using network/subnet on the producer."""
        basic_config = {
            "projects": [
                {
                    "projectId": "host-project",
                    "vpc": [{"type": "vpc", "name": "my-net"}],
                    "subnets": [
                        {
                            "name": "my-psc-subnet",
                            "region": "us-central1",
                            "network": "my-net",
                        }
                    ],
                },
                {
                    "projectId": "service-project",
                    "producers": [
                        {
                            "type": "cloudsql",
                            "name": "my-db",
                            "network": "my-net",
                            "subnet": "my-psc-subnet",
                        }
                    ],
                },
            ]
        }
        with (
            mock.patch.object(
                orchestrator, "SchemasLoader", spec_set=True
            ) as mock_schemas,
            mock.patch.object(
                orchestrator, "DefaultsJsonLoader", spec_set=True
            ) as mock_defaults,
        ):
            _patch_loader(
                mock_schemas,
                schemas={
                    "cloudsql": {},
                    "address": {},
                    "forwardingrule": {},
                    "vpc": {},
                    "subnetwork": {},
                },
            )
            _patch_loader(
                mock_defaults,
                defaults={
                    "cloudsql": {},
                    "address": {},
                    "forwardingrule": {},
                    "vpc": {},
                    "subnetwork": {},
                },
            )
            generator = orchestrator.JsonGenerator(
                "supported_resources.json", "schemas", "defaults"
            )
            result = generator.generate(basic_config)
            host_project_config = result["projects"][0]
            self.assertIn("addresses", host_project_config)
            self.assertIn("forwardingRules", host_project_config)
            self.assertEqual(len(host_project_config["addresses"]), 1)
            self.assertEqual(
                host_project_config["addresses"][0]["name"], "addr-my-db-psc"
            )

    def test_derive_psc_endpoint_falls_back_to_global_settings(self):
        """Verifies PSC endpoint is created using global pscSettings as a fallback."""
        basic_config = {
            "projects": [
                {
                    "projectId": "host-project",
                    "pscSettings": {
                        "networkForPsc": "my-net",
                        "subnetForPsc": "my-psc-subnet",
                    },
                    "vpc": [{"type": "vpc", "name": "my-net"}],
                    "subnets": [
                        {
                            "name": "my-psc-subnet",
                            "region": "us-central1",
                            "network": "my-net",
                        }
                    ],
                },
                {
                    "projectId": "service-project",
                    "producers": [{"type": "cloudsql", "name": "my-db-fallback"}],
                },
            ]
        }
        with (
            mock.patch.object(
                orchestrator, "SchemasLoader", spec_set=True
            ) as mock_schemas,
            mock.patch.object(
                orchestrator, "DefaultsJsonLoader", spec_set=True
            ) as mock_defaults,
        ):
            _patch_loader(
                mock_schemas,
                schemas={
                    "cloudsql": {},
                    "address": {},
                    "forwardingrule": {},
                    "vpc": {},
                    "subnetwork": {},
                },
            )
            _patch_loader(
                mock_defaults,
                defaults={
                    "cloudsql": {},
                    "address": {},
                    "forwardingrule": {},
                    "vpc": {},
                    "subnetwork": {},
                },
            )
            generator = orchestrator.JsonGenerator(
                "supported_resources.json", "schemas", "defaults"
            )
            result = generator.generate(basic_config)
            host_project_config = result["projects"][0]
            self.assertIn("addresses", host_project_config)
            self.assertIn("forwardingRules", host_project_config)
            self.assertEqual(len(host_project_config["addresses"]), 1)
            self.assertEqual(
                host_project_config["addresses"][0]["name"], "addr-my-db-fallback-psc"
            )

    def test_extract_nested_resources_moves_subnets_and_links_network(self):
        """Verifies subnetworks are extracted and linked to their parent VPC."""
        basic_config = {
            "projects": [
                {
                    "projectId": "host-project",
                    "vpc": [
                        {
                            "type": "vpc",
                            "name": "my-net",
                            "subnets": [{"name": "my-subnet"}],
                        }
                    ],
                }
            ]
        }
        with (
            mock.patch.object(
                orchestrator, "SchemasLoader", spec_set=True
            ) as mock_schemas,
            mock.patch.object(
                orchestrator, "DefaultsJsonLoader", spec_set=True
            ) as mock_defaults,
        ):
            _patch_loader(mock_schemas, schemas={"vpc": {}, "subnetwork": {}})
            _patch_loader(mock_defaults, defaults={"vpc": {}, "subnetwork": {}})
            generator = orchestrator.JsonGenerator(
                "supported_resources.json", "schemas", "defaults"
            )
            # Test the public generate() method to see the final state
            result = generator.generate(basic_config)
            final_project = result["projects"][0]
            self.assertNotIn("subnets", final_project["vpc"][0])
            self.assertIn("subnets", final_project)
            self.assertEqual(len(final_project["subnets"]), 1)

    def test_populate_psc_allowed_consumer_projects_updates_producer(self):
        """Verifies that the PSC allow-list is correctly populated based on tags."""
        generator = orchestrator.JsonGenerator(
            supported_resources_path="supported_resources.json",
            schemas_dir="schemas",
            defaults_dir="defaults",
        )
        config = {
            "projects": [
                {
                    "projectId": "service-project",
                    "producers": [
                        {
                            "type": "cloudsql",
                            "name": "my-db",
                            "connectivityType": "psc",
                            "allowedConsumersTags": ["web-app"],
                        }
                    ],
                },
                {
                    "projectId": "consumer-project",
                    "consumers": [
                        {
                            "type": "vm",
                            "name": "frontend-vm",
                            "tags": {"items": ["web-app", "another-tag"]},
                        }
                    ],
                },
            ]
        }

        # Instead of calling the private method, call the public generate() method
        result = generator.generate(config)

        # Assert against the returned result
        producer = result["projects"][0]["producers"][0]
        self.assertIn("settings", producer)
        self.assertIn("ipConfiguration", producer["settings"])
        self.assertIn("pscConfig", producer["settings"]["ipConfiguration"])

        allow_list = producer["settings"]["ipConfiguration"]["pscConfig"][
            "allowedConsumerProjects"
        ]
        self.assertEqual(len(allow_list), 1)
        self.assertIn("consumer-project", allow_list)

    def test_find_subnet_details_handles_null_vpc_in_list(self):
        """Verifies _find_subnet_details doesn't crash if a VPC in the list is null."""
        generator = orchestrator.JsonGenerator("supported_resources.json", "s", "d")
        all_projects = [
            {
                "projectId": "host-project",
                "vpc": [None, {"name": "wrong-net"}, {"name": "my-net"}],
                "subnets": [
                    {"name": "my-subnet", "region": "us-central1", "network": "my-net"}
                ],
            }
        ]

        _, _, path = generator._find_subnet_details("my-subnet", "my-net", all_projects)
        self.assertIsNotNone(path)
        self.assertIn("my-subnet", path)

    def test_find_subnet_details_finds_nested_subnet_pre_extraction(self):
        """Verifies subnets can be found when still nested inside a VPC."""
        generator = orchestrator.JsonGenerator("supported_resources.json", "s", "d")
        all_projects = [
            {
                "projectId": "host-project",
                "vpc": [
                    {
                        "name": "my-net",
                        "subnetworks": [
                            {"name": "my-nested-subnet", "region": "us-central1"}
                        ],
                    }
                ],
                "subnets": [],
            }
        ]

        _, _, path = generator._find_subnet_details(
            "my-nested-subnet", "my-net", all_projects
        )
        self.assertIsNotNone(path, "Should have found the nested subnet")
        self.assertIn("my-nested-subnet", path)


if __name__ == "__main__":
    unittest.main()
