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
from utilities import defaults_json_loader
from utilities import json_loader_utils


class TestDefaultsJsonLoader(fake_filesystem_unittest.TestCase):
    """Tests for the defaults_json_loader.DefaultsJsonLoader class."""

    def setUp(self):
        super().setUp()
        self.setUpPyfakefs()
        self.module_logger = logging.getLogger(defaults_json_loader.__name__)
        self.original_module_logger_level = self.module_logger.level
        self.module_logger.setLevel(logging.DEBUG)
        self.test_data_root = "/test_data"
        self.defaults_dir = os.path.join(self.test_data_root, "defaults")
        self.fs.create_dir(self.defaults_dir)

    def tearDown(self):
        self.module_logger.setLevel(self.original_module_logger_level)
        super().tearDown()

    @mock.patch.object(json_loader_utils, "load_json_file")
    def test_load_all_defaults(self, mock_load_json_file):
        """Tests loading defaults: success, missing, firewall mapping."""

        base_path = pathlib.Path(self.defaults_dir)

        def simulate_defaults_load(file_path):
            if file_path == base_path / "vm_defaults.json":
                return {"vm_key": "vm_val"}
            if file_path == base_path / "firewall_rule_defaults.json":
                return {"fw_key": "fw_val"}
            return None

        mock_load_json_file.side_effect = simulate_defaults_load

        loader = defaults_json_loader.DefaultsJsonLoader(self.defaults_dir)
        resource_types = ["vm", "network", "firewall_rule"]
        with self.assertLogs(self.module_logger, level="DEBUG") as cm:
            result = loader.load_defaults_for_resource_types(resource_types)
        self.assertEqual(result["vm"], {"vm_key": "vm_val"})
        self.assertEqual(result["network"], {})
        self.assertEqual(result["firewall_rule"], {"fw_key": "fw_val"})
        self.assertTrue(
            any(
                "Defaults not found or failed to load for resource type: 'network'"
                in log_msg
                for log_msg in cm.output
            )
        )
        base_path = pathlib.Path(self.defaults_dir)
        expected_calls = [
            mock.call(base_path / "vm_defaults.json"),
            mock.call(base_path / "network_defaults.json"),
            mock.call(base_path / "firewall_rule_defaults.json"),
        ]
        mock_load_json_file.assert_has_calls(expected_calls, any_order=True)
        self.assertEqual(loader.get_defaults(), result)


if __name__ == "__main__":
    unittest.main()
