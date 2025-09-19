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

import logging
import os
import pathlib
from unittest import mock

from pyfakefs import fake_filesystem_unittest

import unittest
from utilities import schemas_loader


class TestSchemasLoader(fake_filesystem_unittest.TestCase):
    """Tests for the schemas_loader.SchemasLoader class."""

    def setUp(self):
        super().setUp()
        self.setUpPyfakefs()
        self.module_logger = logging.getLogger(schemas_loader.__name__)
        self.original_module_logger_level = self.module_logger.level
        self.module_logger.setLevel(logging.DEBUG)
        self.test_data_root = "/test_data"
        self.schemas_dir = os.path.join(self.test_data_root, "schemas")
        self.fs.create_dir(self.schemas_dir)

    def tearDown(self):
        self.module_logger.setLevel(self.original_module_logger_level)
        super().tearDown()

    @mock.patch("utilities.json_loader_utils.load_json_file")
    def test_load_all_schemas(self, mock_load_json_file):
        """Tests loading schemas: supported, unsupported, missing."""
        supported_resources_config = {
            "vm": {"details": "..."},
            "db_instance": {"details": "..."},
        }

        base_path = pathlib.Path(self.schemas_dir)

        def simulate_schema_load(file_path):
            if file_path == base_path / "vm_schema.json":
                return {"vm_schema": "details"}
            return None

        mock_load_json_file.side_effect = simulate_schema_load

        loader = schemas_loader.SchemasLoader(self.schemas_dir)
        resource_types = ["vm", "db_instance", "network"]
        with self.assertLogs(self.module_logger, level="WARNING") as cm:
            result = loader.load_schemas_for_resource_types(
                resource_types, supported_resources_config
            )

        self.assertEqual(result["vm"], {"vm_schema": "details"})
        self.assertEqual(result["db_instance"], {})
        self.assertEqual(result["network"], {})

        self.assertTrue(
            any(
                "Resource type 'network' is not in supported_resources." in log_msg
                for log_msg in cm.output
            )
        )
        self.assertTrue(
            any(
                "Schema not found or failed to load for supported resource type: "
                "'db_instance'" in log_msg
                for log_msg in cm.output
            )
        )
        base_path = pathlib.Path(self.schemas_dir)
        expected_calls = [
            mock.call(base_path / "vm_schema.json"),
            mock.call(base_path / "db_instance_schema.json"),
        ]
        mock_load_json_file.assert_has_calls(expected_calls, any_order=True)
        self.assertEqual(loader.get_schemas(), result)


if __name__ == "__main__":
    unittest.main()
