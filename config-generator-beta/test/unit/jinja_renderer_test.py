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
from pyfakefs import fake_filesystem_unittest
from utilities import jinja_renderer


class JinjaRendererTest(fake_filesystem_unittest.TestCase):
    """Tests for the jinja_renderer.render_template function."""

    def setUp(self):
        super().setUp()
        self.setUpPyfakefs()
        self.template_dir = "/fake/templates"
        self.fs.create_dir(self.template_dir)
        self.module_logger = logging.getLogger(jinja_renderer.__name__)

    def test_render_template_success(self):
        """Verifies successful rendering of a valid template."""
        template_name = "test.j2"
        template_content = "Hello, {{ name }}! Your score is {{ score }}."
        self.fs.create_file(
            f"{self.template_dir}/{template_name}", contents=template_content
        )

        context = {"name": "World", "score": 100}
        result = jinja_renderer.render_template(
            self.template_dir, template_name, context
        )

        self.assertEqual(result, "Hello, World! Your score is 100.")

    def test_render_template_directory_not_found_is_caught(self):
        """Tests that a TemplateNotFound error is handled."""
        with self.assertLogs(self.module_logger, level="ERROR") as cm:
            result = jinja_renderer.render_template("/non/existent/dir", "test.j2", {})
        self.assertIsNone(result)
        self.assertIn("Failed to render template", cm.output[0])

    def test_render_template_file_not_found_is_caught(self):
        """Tests that a TemplateNotFound error is handled."""
        with self.assertLogs(self.module_logger, level="ERROR") as cm:
            result = jinja_renderer.render_template(
                self.template_dir, "non_existent.j2", {}
            )
        self.assertIsNone(result)
        self.assertIn("non_existent.j2", str(cm.output))

    def test_render_template_syntax_error_is_caught(self):
        """Tests that a TemplateSyntaxError is handled."""
        template_name = "bad_syntax.j2"
        template_content = "Hello, {{ name"  # Unclosed variable tag
        self.fs.create_file(
            f"{self.template_dir}/{template_name}", contents=template_content
        )

        with self.assertLogs(self.module_logger, level="ERROR") as cm:
            result = jinja_renderer.render_template(
                self.template_dir, template_name, {"name": "World"}
            )
        self.assertIsNone(result)
        self.assertIn("Failed to render template", cm.output[0])

    def test_render_template_handles_render_time_error(self):
        """Tests graceful failure when an exception occurs during rendering."""

        class BadObject:
            @property
            def name(self):
                raise ValueError("This is a deliberate error")

        template_name = "render_error.j2"
        template_content = "Hello, {{ user.name }}."
        self.fs.create_file(
            f"{self.template_dir}/{template_name}", contents=template_content
        )

        # Use an instance of our helper class to force an exception
        context = {"user": BadObject()}
        with self.assertLogs(self.module_logger, level="ERROR") as cm:
            result = jinja_renderer.render_template(
                self.template_dir, template_name, context
            )

        self.assertIsNone(result)
        self.assertIn("Failed to render template", cm.output[0])
        self.assertIn("deliberate error", cm.output[0])

    def test_render_template_with_custom_regex_filter(self):
        """Verifies that the custom 'regex_replace' filter works correctly."""
        template_name = "custom_filter_test.j2"
        template_content = "The region is: {{ zone | regex_replace('-[a-z]$', '') }}"
        self.fs.create_file(
            f"{self.template_dir}/{template_name}", contents=template_content
        )
        context = {"zone": "us-central1-a"}
        result = jinja_renderer.render_template(
            self.template_dir, template_name, context
        )
        self.assertEqual(result, "The region is: us-central1")
