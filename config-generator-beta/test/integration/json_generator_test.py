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

"""Integration tests for the json_generator core library."""

import json
import logging
import os

from pyfakefs import fake_filesystem_unittest

import unittest
from utilities import json_generator


class JsonGeneratorTest(fake_filesystem_unittest.TestCase):
    """Tests the generate_config_from_path function end-to-end."""

    def setUp(self):
        super().setUp()
        self.setUpPyfakefs()
        self.maxDiff = None

        # Define paths locally for the fake file system
        self.schemas_dir = "/schemas"
        self.defaults_dir = "/defaults"
        self.supported_resources_path = "/schemas/supported_resources.json"

        # Use the local paths to create the test files
        self.fs.create_dir(self.schemas_dir)
        self.fs.create_dir(self.defaults_dir)
        self.fs.create_file(
            self.supported_resources_path,
            contents="""
        {
          "vm": { "uriTemplate": "projects/{projectId}/zones/{zone}/instances/{name}", "referenceFields": ["network", "subnetwork"] },
          "vpc": { "nestedResources": { "subnets": "subnetwork" }, "uriTemplate": "projects/{projectId}/global/networks/{name}" },
          "subnetwork": { "uriTemplate": "projects/{projectId}/regions/{region}/subnetworks/{name}", "referenceFields": ["network"] }
        }
        """,
        )
        self.fs.create_file(
            os.path.join(self.schemas_dir, "vm_schema.json"),
            contents=(
                '{"type": "object", "properties": {"name": {}, "zone": {},'
                ' "machineType": {}, "networkInterfaces": {}}}'
            ),
        )
        self.fs.create_file(
            os.path.join(self.defaults_dir, "vm_defaults.json"),
            contents='{"machineType": "e2-medium"}',
        )
        self.fs.create_file(
            os.path.join(self.schemas_dir, "vpc_schema.json"),
            contents=(
                '{"type": "object", "properties": {"name": {}, "subnets": {}, "autoCreateSubnetworks": {}}}'
            ),
        )
        self.fs.create_file(
            os.path.join(self.defaults_dir, "vpc_defaults.json"),
            contents="{}",
        )
        self.fs.create_file(
            os.path.join(self.schemas_dir, "subnetwork_schema.json"),
            contents='{"type": "object", "properties": {"name": {}, "region": {}, "ipCidrRange": {}, "network": {}}}',
        )

    def test_end_to_end_generation_is_correct(self):
        """Tests the full pipeline from basic.json to a complete, resolved config.

        This is the primary integration test and does not use mocks.
        """
        input_path = "/test_data/basic.json"
        output_dir = "/test_output"
        self.fs.create_dir(os.path.dirname(input_path))
        self.fs.create_dir(output_dir)

        basic_config_content = {
            "projects": [
                {
                    "projectId": "my-proj",
                    "vpc": [
                        {
                            "type": "vpc",
                            "name": "my-network",
                            "subnets": [
                                {
                                    "name": "my-subnet",
                                    "region": "us-central1",
                                    "ipCidrRange": "10.0.1.0/24",
                                }
                            ],
                        }
                    ],
                    "consumers": [
                        {
                            "type": "vm",
                            "name": "my-vm",
                            "zone": "us-central1-a",
                            "networkInterfaces": [
                                {"network": "my-network", "subnetwork": "my-subnet"}
                            ],
                        }
                    ],
                }
            ]
        }
        self.fs.create_file(input_path, contents=json.dumps(basic_config_content))

        result_path = json_generator.generate_config_from_path(
            basic_config_path=input_path,
            config_name="basic",
            output_dir=output_dir,
            supported_resources_path=self.supported_resources_path,
            schemas_dir=self.schemas_dir,
            defaults_dir=self.defaults_dir,
        )

        expected_output_path = os.path.join(output_dir, "basic-complete.json")
        self.assertEqual(result_path, expected_output_path)
        self.assertTrue(os.path.exists(expected_output_path))

        with open(expected_output_path, "r") as f:
            generated_data = json.load(f)

        expected_complete_config = {
            "projects": [
                {
                    "projectId": "my-proj",
                    "vpc": [
                        {
                            "type": "vpc",
                            "name": "my-network",
                            "autoCreateSubnetworks": False,
                            "selfLink": "projects/my-proj/global/networks/my-network",
                        }
                    ],
                    "consumers": [
                        {
                            "type": "vm",
                            "name": "my-vm",
                            "zone": "us-central1-a",
                            "machineType": "e2-medium",
                            "networkInterfaces": [
                                {
                                    "network": "projects/my-proj/global/networks/my-network",
                                    "subnetwork": "projects/my-proj/regions/us-central1/subnetworks/my-subnet",
                                }
                            ],
                            "selfLink": "projects/my-proj/zones/us-central1-a/instances/my-vm",
                        }
                    ],
                    "subnets": [
                        {
                            "type": "subnetwork",
                            "name": "my-subnet",
                            "region": "us-central1",
                            "ipCidrRange": "10.0.1.0/24",
                            "network": "projects/my-proj/global/networks/my-network",
                            "selfLink": "projects/my-proj/regions/us-central1/subnetworks/my-subnet",
                        }
                    ],
                }
            ]
        }

        self.assertEqual(generated_data, expected_complete_config)


if __name__ == "__main__":
    unittest.main()
