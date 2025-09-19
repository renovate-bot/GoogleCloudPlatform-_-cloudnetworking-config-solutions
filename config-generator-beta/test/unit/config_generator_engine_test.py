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

from unittest import mock

import unittest
from utilities import config_generator_engine

MOCK_SUPPORTED_RESOURCES = {
    "vm": {
        "discoveryUrl": (
            "https://compute.googleapis.com/discovery/v1/apis/compute/v1/rest"
        )
    },
    "cloudsql": {
        "discoveryUrl": (
            "https://sqladmin.googleapis.com/discovery/v1/apis/sqladmin/v1/rest"
        )
    },
}

MOCK_SHARED_VPC_HOST_CONFIG = {
    "projects": [
        {
            "projectId": "host-project-1",
            "hostProject": True,
            "vpc": [{"name": "main-shared-vpc"}],
            "producers": [{"type": "cloudsql", "name": "test-db"}],
            "consumers": [{"type": "vm", "name": "test-vm-in-host"}],
        },
        {"projectId": "service-project-A"},
        {"projectId": "service-project-B"},
    ]
}

MOCK_PSC_CONFIG = {
    "projects": [
        {
            "projectId": "producer-proj",
            "producers": [{"type": "cloudsql", "name": "my-psc-db"}],
        },
        {
            "projectId": "network-proj",
            "forwardingRules": [{"targetProducerName": "my-psc-db"}],
        },
    ]
}

MOCK_NON_HOST_VPC_CONFIG = {
    "projects": [
        {
            "projectId": "standalone-project",
            "vpc": [{"name": "standalone-vpc"}],
        }
    ]
}

MOCK_SUPPORTED_RESOURCES_FOR_YAML = {
    "cloudsql": {
        "generationConfig": {
            "category": "producer",
            "folderName": "CloudSQL",
            "templateFilename": "cloudsql.yaml.j2",
        },
        "securityConfigTemplate": "cloudsql.tfvars.j2",
    },
    "vm": {
        "generationConfig": {
            "category": "consumer",
            "folderName": "GCE",
            "templateFilename": "gce.yaml.j2",
        }
    },
}

MOCK_YAML_GEN_CONFIG = {
    "projects": [
        {
            "projectId": "host-project",
            "vpc": [{"name": "test-vpc"}],
            "producers": [{"type": "cloudsql", "name": "my-db"}],
            "consumers": [{"type": "vm", "name": "my-vm"}],
        }
    ]
}


class TerraformArtifactGeneratorEngineTest(unittest.TestCase):

    def test_prepare_organisation_context_success(self):
        engine = config_generator_engine.TerraformArtifactGenerator(
            complete_config=MOCK_SHARED_VPC_HOST_CONFIG,
            supported_resources=MOCK_SUPPORTED_RESOURCES,
            templates_base_dir="test_templates_dir",
        )
        context = engine._prepare_organisation_context()

        self.assertIn("projectApis", context)
        project_apis = context["projectApis"]
        self.assertIn("host-project-1", project_apis)

        self.assertEqual(
            project_apis["host-project-1"],
            ["compute.googleapis.com", "sqladmin.googleapis.com"],
        )

    def test_prepare_networking_context_for_host(self):
        engine = config_generator_engine.TerraformArtifactGenerator(
            complete_config=MOCK_SHARED_VPC_HOST_CONFIG,
            supported_resources=MOCK_SUPPORTED_RESOURCES,
            templates_base_dir="test_templates_dir",
        )
        context = engine._prepare_networking_context()

        self.assertTrue(context)
        self.assertEqual(context["sharedVpcHost"], "true")
        self.assertCountEqual(
            context["sharedVpcServiceProjects"],
            ["service-project-A", "service-project-B"],
        )

    def test_prepare_networking_context_for_non_host(self):
        engine = config_generator_engine.TerraformArtifactGenerator(
            complete_config=MOCK_NON_HOST_VPC_CONFIG,
            supported_resources=MOCK_SUPPORTED_RESOURCES,
            templates_base_dir="test_templates_dir",
        )
        context = engine._prepare_networking_context()

        self.assertTrue(context)
        self.assertEqual(context["sharedVpcHost"], "false")
        self.assertCountEqual(context["sharedVpcServiceProjects"], [])

    def test_prepare_networking_context_returns_none_if_no_vpc(self):
        config_no_vpc = {"projects": [{"projectId": "p1"}]}
        engine = config_generator_engine.TerraformArtifactGenerator(
            complete_config=config_no_vpc,
            supported_resources={},
            templates_base_dir="test_templates_dir",
        )
        with self.assertLogs(config_generator_engine.logger, level="WARNING") as cm:
            context = engine._prepare_networking_context()

        self.assertIsNone(context)
        self.assertIn("No VPC definition found", cm.output[0])

    def test_prepare_psc_context_success(self):
        engine = config_generator_engine.TerraformArtifactGenerator(
            complete_config=MOCK_PSC_CONFIG,
            supported_resources=MOCK_SUPPORTED_RESOURCES,
            templates_base_dir="test_templates_dir",
        )
        context = engine._prepare_psc_context()

        self.assertIn("pscEndpointsData", context)
        psc_endpoints = context["pscEndpointsData"]
        self.assertEqual(len(psc_endpoints), 1)
        self.assertEqual(psc_endpoints[0]["producerName"], "my-psc-db")
        self.assertEqual(psc_endpoints[0]["producerType"], "cloudsql")
        self.assertEqual(psc_endpoints[0]["producerInstanceProjectId"], "producer-proj")

    def test_prepare_psc_context_handles_orphan_forwarding_rule(self):
        config_with_orphan_fr = {
            "projects": [
                {"forwardingRules": [{"target_producer_name": "non-existent"}]}
            ]
        }
        engine = config_generator_engine.TerraformArtifactGenerator(
            complete_config=config_with_orphan_fr,
            supported_resources={},
            templates_base_dir="test_templates_dir",
        )
        context = engine._prepare_psc_context()

        self.assertIn("pscEndpointsData", context)
        self.assertEqual(context["pscEndpointsData"], [])

    def test_prepare_resource_files_contexts(self):
        engine = config_generator_engine.TerraformArtifactGenerator(
            complete_config=MOCK_YAML_GEN_CONFIG,
            supported_resources=MOCK_SUPPORTED_RESOURCES_FOR_YAML,
            templates_base_dir="test_templates_dir",
        )

        files_to_generate = engine._prepare_resource_files_contexts()

        self.assertEqual(len(files_to_generate), 3)

        producer_yaml = next(
            f for f in files_to_generate if f["output_path"].endswith("my-db.yaml")
        )
        self.assertIsNotNone(producer_yaml)
        self.assertEqual(producer_yaml["context"]["instance"]["name"], "my-db")

        consumer_yaml = next(
            f for f in files_to_generate if f["output_path"].endswith("my-vm.yaml")
        )
        self.assertIsNotNone(consumer_yaml)
        self.assertEqual(consumer_yaml["context"]["instance"]["name"], "my-vm")

        security_tfvars = next(
            f for f in files_to_generate if f["output_path"].endswith("cloudsql.tfvars")
        )
        self.assertIsNotNone(security_tfvars)
        self.assertEqual(security_tfvars["context"]["projectId"], "host-project")

    @mock.patch("utilities.jinja_renderer.render_template")
    @mock.patch.object(
        config_generator_engine.TerraformArtifactGenerator,
        "_prepare_resource_files_contexts",
    )
    def test_generate_all_resource_files_success(
        self, mock_prepare_contexts, mock_render_template
    ):
        """Tests that the orchestrator method correctly renders all files."""
        mock_prepare_contexts.return_value = [
            {
                "output_path": "producers/config/db.yaml",
                "template_dir": "td1",
                "template_name": "tn1",
                "context": {"k": "v1"},
            },
            {
                "output_path": "consumers/config/vm.yaml",
                "template_dir": "td2",
                "template_name": "tn2",
                "context": {"k": "v2"},
            },
        ]
        mock_render_template.return_value = "rendered content"

        engine = config_generator_engine.TerraformArtifactGenerator(
            complete_config={},
            supported_resources={},
            templates_base_dir="test_templates_dir",
        )

        result = engine.generate_all_resource_files()

        self.assertIsNotNone(result)
        self.assertEqual(mock_render_template.call_count, 2)
        self.assertIn("producers/config/db.yaml", result)
        self.assertEqual(result["consumers/config/vm.yaml"], "rendered content")

    @mock.patch.object(
        config_generator_engine.TerraformArtifactGenerator,
        "_prepare_resource_files_contexts",
    )
    def test_generate_all_resource_files_handles_render_failure(
        self, mock_prepare_contexts
    ):
        """Tests that the orchestrator returns None if any template fails."""
        mock_prepare_contexts.return_value = [
            {
                "output_path": "producers/config/db.yaml",
                "template_dir": "td1",
                "template_name": "tn1",
                "context": {},
            }
        ]

        engine = config_generator_engine.TerraformArtifactGenerator(
            complete_config={},
            supported_resources={},
            templates_base_dir="test_templates_dir",
        )
        # Patch the renderer to return None and verify our
        # orchestrator also returns None.
        with mock.patch(
            "utilities.jinja_renderer.render_template",
            return_value=None,
        ):
            result = engine.generate_all_resource_files()

        self.assertIsNone(result)

    def test_prepare_resource_files_contexts_creates_correct_paths(self):
        """Verifies the generated output_path has correct singular and cased names."""
        mock_config = {
            "projects": [
                {
                    "projectId": "test-project",
                    "vpc": [{"name": "test-vpc"}],
                    "producers": [{"type": "cloudsql", "name": "my-db"}],
                    "consumers": [{"type": "vm", "name": "my-vm"}],
                }
            ]
        }
        mock_supported = {
            "cloudsql": {
                "generationConfig": {
                    "category": "producer",
                    "folderName": "CloudSQL",
                    "templateFilename": "cloudsql.yaml.j2",
                }
            },
            "vm": {
                "generationConfig": {
                    "category": "consumer",
                    "folderName": "GCE",
                    "templateFilename": "gce.yaml.j2",
                }
            },
        }
        engine = config_generator_engine.TerraformArtifactGenerator(
            complete_config=mock_config,
            supported_resources=mock_supported,
            templates_base_dir="test_templates_dir",
        )

        files_to_generate = engine._prepare_resource_files_contexts()

        self.assertEqual(len(files_to_generate), 2)
        producer_spec = next(
            f for f in files_to_generate if "my-db.yaml" in f["output_path"]
        )
        consumer_spec = next(
            f for f in files_to_generate if "my-vm.yaml" in f["output_path"]
        )

        self.assertEqual(
            producer_spec["output_path"], "producer/CloudSQL/config/my-db.yaml"
        )
        self.assertEqual(consumer_spec["output_path"], "consumer/GCE/config/my-vm.yaml")


if __name__ == "__main__":
    unittest.main()
