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

"""Unit tests for the main config_generator.py orchestrator script."""

import os
import shutil
import sys
import unittest
from unittest import mock
from pyfakefs import fake_filesystem_unittest
import config_generator


class ConfigGeneratorTest(fake_filesystem_unittest.TestCase):
    """Tests the main function and workflow of the config generator script."""

    def setUp(self):
        """Set up the fake file system and patch external dependencies."""
        super().setUp()
        self.setUpPyfakefs()

        # Create all necessary input directories and files
        self.fs.create_dir(config_generator.ARCHITECTURE_SPEC_DIR)
        self.fs.create_dir(config_generator.SCHEMA_DIR)
        self.fs.create_dir(config_generator.DEFAULTS_DIR)
        self.fs.create_dir(config_generator.TEMPLATES_BASE_DIR)

        self.spec_file_path = os.path.join(
            config_generator.ARCHITECTURE_SPEC_DIR, "test_arch.json"
        )
        self.fs.create_file(self.spec_file_path, contents='{"name": "test"}')
        self.fs.create_file(
            config_generator.SUPPORTED_RESOURCES_PATH, contents='{"vm": {}}'
        )

        run_sh_path = os.path.join(
            config_generator.PROJECT_ROOT_DIR, "execution", "run.sh"
        )
        self.fs.create_file(run_sh_path, contents="#!/bin/bash\necho 'Mock run.sh'")

        self.complete_json_path = os.path.join(
            config_generator.MAIN_OUTPUT_DIR, "test_arch-complete.json"
        )

    def tearDown(self):
        """Clean up the fake file system after each test."""
        if os.path.exists(config_generator.MAIN_OUTPUT_DIR):
            shutil.rmtree(config_generator.MAIN_OUTPUT_DIR)
        super().tearDown()

    @mock.patch("sys.argv", ["config_generator.py", "--all"])
    @mock.patch("builtins.input", side_effect=["1", "y", "y"])
    @mock.patch("config_generator.generate_config_from_path")
    @mock.patch("config_generator.generate_all_tf_files", return_value=True)
    @mock.patch("config_generator._copy_static_files")
    @mock.patch("config_generator.run_terraform_fmt")
    @mock.patch("config_generator.open_and_advise_terraform_review")
    @mock.patch("subprocess.run")
    def test_main_all_flag_full_success_workflow(self, mock_subprocess, *mocks):
        """Tests the '--all' flag with user confirming all prompts."""
        (
            mock_open_review,
            mock_tf_fmt,
            mock_copy,
            mock_gen_tf,
            mock_gen_json,
            _,
        ) = mocks

        def fake_generate_config_from_path(*args, **kwargs):
            self.fs.create_file(self.complete_json_path, contents='{"projects":[]}')
            return self.complete_json_path

        mock_gen_json.side_effect = fake_generate_config_from_path

        config_generator.main()

        mock_gen_json.assert_called_once()
        mock_gen_tf.assert_called_once_with(
            self.complete_json_path, config_generator.MAIN_OUTPUT_DIR
        )
        mock_copy.assert_called_once()
        mock_tf_fmt.assert_called_once()
        mock_open_review.assert_called_once()
        mock_subprocess.assert_called()

    @mock.patch("sys.argv", ["config_generator.py", "--all"])
    @mock.patch("builtins.input", side_effect=["1", "n"])
    @mock.patch("config_generator.generate_config_from_path")
    @mock.patch("shutil.rmtree")
    @mock.patch("sys.exit", side_effect=SystemExit)
    def test_main_all_flag_user_aborts_review(
        self, mock_exit, mock_rmtree, mock_gen_json, mock_input
    ):
        """Tests the '--all' flag where user aborts after JSON review."""
        mock_gen_json.return_value = self.complete_json_path

        with self.assertRaises(SystemExit):
            config_generator.main()

        mock_gen_json.assert_called_once()
        mock_rmtree.assert_called_once_with(
            config_generator.MAIN_OUTPUT_DIR, ignore_errors=True
        )
        mock_exit.assert_called_once_with(0)

    @mock.patch("sys.argv", ["config_generator.py", "--full-spec"])
    @mock.patch("builtins.input", return_value="1")
    @mock.patch("config_generator.generate_config_from_path")
    @mock.patch("config_generator._open_file_in_default_app")
    @mock.patch("config_generator.generate_all_tf_files")
    def test_main_full_spec_flag_stops_after_json_generation(
        self, mock_gen_tf, mock_open_file, mock_gen_json, mock_input
    ):
        """Tests that '--full-spec' only generates the complete.json."""
        mock_gen_json.return_value = self.complete_json_path

        config_generator.main()

        mock_gen_json.assert_called_once()
        mock_open_file.assert_called_once_with(self.complete_json_path)
        mock_gen_tf.assert_not_called()

    @mock.patch("sys.argv", ["config_generator.py", "--terraform"])
    @mock.patch("config_generator.generate_all_tf_files", return_value=True)
    @mock.patch("config_generator._copy_static_files")
    @mock.patch("config_generator.run_terraform_fmt")
    @mock.patch("subprocess.run")
    def test_main_terraform_flag_generates_files(
        self, mock_subprocess, mock_tf_fmt, mock_copy, mock_gen_tf
    ):
        """Tests '--terraform' successfully generates files from an existing spec."""
        self.fs.create_dir(config_generator.MAIN_OUTPUT_DIR)
        self.fs.create_file(self.complete_json_path, contents='{"projects":[]}')

        config_generator.main()

        mock_gen_tf.assert_called_once_with(
            self.complete_json_path, config_generator.MAIN_OUTPUT_DIR
        )
        mock_copy.assert_called_once()
        mock_tf_fmt.assert_called_once()
        mock_subprocess.assert_not_called()

    @mock.patch("sys.argv", ["config_generator.py", "--terraform"])
    @mock.patch("sys.exit", side_effect=SystemExit)
    def test_main_terraform_flag_exits_if_spec_missing(self, mock_exit):
        """Tests that '--terraform' exits if complete.json is missing."""
        self.fs.create_dir(config_generator.MAIN_OUTPUT_DIR)

        with self.assertRaises(SystemExit):
            config_generator.main()
        mock_exit.assert_called_once_with(1)

    @mock.patch("sys.argv", ["config_generator.py", "--terraform-apply"])
    @mock.patch("builtins.input", return_value="y")
    @mock.patch("config_generator.generate_all_tf_files", return_value=True)
    @mock.patch("subprocess.run")
    def test_main_terraform_apply_flag_generates_and_deploys(
        self, mock_subprocess, mock_gen_tf, mock_input
    ):
        """Tests '--terraform-apply' generates files and then deploys."""
        self.fs.create_dir(config_generator.MAIN_OUTPUT_DIR)
        self.fs.create_file(self.complete_json_path, contents='{"projects":[]}')

        config_generator.main()

        execution_dir = os.path.join(config_generator.SCRIPT_DIR, "../execution")
        expected_call = mock.call(
            ["bash", "./run.sh", "-s", "all", "-t", "init-apply-auto-approve"],
            check=True,
            cwd=execution_dir,
        )
        self.assertIn(expected_call, mock_subprocess.call_args_list)

    @mock.patch("sys.argv", ["config_generator.py", "--apply"])
    @mock.patch("builtins.input", return_value="y")
    @mock.patch("subprocess.run")
    def test_main_apply_flag_runs_deployment(self, mock_subprocess, mock_input):
        """Tests that '--apply' correctly triggers the deployment step."""
        self.fs.create_dir(config_generator.MAIN_OUTPUT_DIR)
        config_generator.main()
        mock_subprocess.assert_called_once()

    @mock.patch("sys.argv", ["config_generator.py", "--apply"])
    @mock.patch("builtins.input", return_value="no")
    @mock.patch("subprocess.run")
    def test_main_apply_flag_skips_deployment_on_user_abort(
        self, mock_subprocess, mock_input
    ):
        """Tests that '--apply' is skipped if the user says no."""
        self.fs.create_dir(config_generator.MAIN_OUTPUT_DIR)
        config_generator.main()
        mock_subprocess.assert_not_called()

    # --- Tests for invalid flag usage ---

    @mock.patch("sys.argv", ["config_generator.py", "--all", "--terraform"])
    @mock.patch("argparse.ArgumentParser.print_help")
    @mock.patch("sys.exit", side_effect=SystemExit)
    def test_main_exits_if_multiple_flags_are_given(self, mock_exit, mock_print_help):
        """Tests that the script exits if more than one flag is provided."""
        with self.assertRaises(SystemExit):
            config_generator.main()
        mock_print_help.assert_called_once()
        mock_exit.assert_called_once_with(1)

    @mock.patch("sys.argv", ["config_generator.py"])
    @mock.patch("argparse.ArgumentParser.print_help")
    @mock.patch("sys.exit", side_effect=SystemExit)
    def test_main_exits_if_no_flags_are_given(self, mock_exit, mock_print_help):
        """Tests that the script exits if no flags are provided."""
        with self.assertRaises(SystemExit):
            config_generator.main()
        mock_print_help.assert_called_once()
        mock_exit.assert_called_once_with(1)

    @mock.patch("sys.argv", ["config_generator.py", "--unknown-flag"])
    def test_main_argparse_exits_on_unknown_flag(self):
        """Tests that argparse exits when an unknown flag is used."""
        # Argparse exits with code 2 for unknown arguments
        with self.assertRaises(SystemExit) as cm:
            config_generator.main()
        self.assertEqual(cm.exception.code, 2)


if __name__ == "__main__":
    unittest.main()
