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

import json
import logging
import pathlib
from unittest import mock

from pyfakefs import fake_filesystem_unittest

import unittest
from utilities import json_loader_utils


class LoadJsonFileTest(fake_filesystem_unittest.TestCase):
    """Tests for the json_loader_utils.load_json_file utility function."""

    def setUp(self):
        super().setUp()
        self.setUpPyfakefs()
        self.module_logger = logging.getLogger(json_loader_utils.__name__)
        self.original_module_logger_level = self.module_logger.level
        self.module_logger.setLevel(logging.DEBUG)

    def tearDown(self):
        self.module_logger.setLevel(self.original_module_logger_level)
        super().tearDown()

    def test_load_valid_json_file(self):
        """Tests loading a correctly formatted JSON file."""
        complex_valid_json_path = "complex_valid.json"
        complex_valid_json_content = {
            "key": "value",
            "number": 123,
            "nested": {"a": True},
        }
        self.fs.create_file(
            complex_valid_json_path, contents=json.dumps(complex_valid_json_content)
        )

        with self.assertNoLogs(logger=self.module_logger, level=logging.WARNING):
            data = json_loader_utils.load_json_file(complex_valid_json_path)
        self.assertEqual(data, complex_valid_json_content)

    def test_load_file_not_found(self):
        """Tests behavior when the JSON file does not exist."""
        non_existent_path = "non_existent.json"

        with self.assertLogs(self.module_logger, level="ERROR") as cm:
            data = json_loader_utils.load_json_file(non_existent_path)
        self.assertIsNone(data)
        log_content = "\n".join(cm.output)
        self.assertIn(f"File not found: {non_existent_path}", log_content)

    def test_load_invalid_json_decode_error(self):
        """Tests behavior when the file contains malformed JSON."""
        invalid_json_path = "invalid.json"
        self.fs.create_file(invalid_json_path, contents='{"key": "malformed",}')

        with self.assertLogs(self.module_logger, level="ERROR") as cm:
            data = json_loader_utils.load_json_file(invalid_json_path)
        self.assertIsNone(data)
        log_content = "\n".join(cm.output)
        self.assertIn("Error decoding JSON", log_content)

    def test_load_empty_file_results_in_decode_error(self):
        """Tests behavior when the JSON file is empty."""
        empty_file_path = "empty.json"
        self.fs.create_file(empty_file_path, contents="")

        with self.assertLogs(self.module_logger, level="ERROR") as cm:
            data = json_loader_utils.load_json_file(empty_file_path)
        self.assertIsNone(data)
        log_content = "\n".join(cm.output)
        self.assertIn("Error decoding JSON", log_content)

    @mock.patch("utilities.json_loader_utils.open")
    def test_load_json_os_error_during_open(self, mock_open_func):
        """Tests behavior when an OS-level error (e.g., permission) occurs."""
        mock_open_func.side_effect = OSError("Simulated permission denied")
        dummy_path_for_os_error = "file_with_os_error.json"

        with self.assertLogs(self.module_logger, level="ERROR") as cm:
            data = json_loader_utils.load_json_file(dummy_path_for_os_error)
        self.assertIsNone(data)
        log_content = "\n".join(cm.output)
        self.assertIn("An unexpected error occurred loading", log_content)
        mock_open_func.assert_called_once_with(
            pathlib.PosixPath("file_with_os_error.json"), "r", encoding="utf-8"
        )


if __name__ == "__main__":
    unittest.main()
