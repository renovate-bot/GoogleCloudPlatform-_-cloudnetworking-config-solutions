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
import pathlib
import unittest

# This constructs the path to the 'architecture-spec' directory
# relative to this test file's location.
SPEC_DIR = pathlib.Path(__file__).parent.parent.parent / "architecture-spec"


class TestArchitectureSpecs(unittest.TestCase):
    """
    Validates all architecture specification JSON files with a resource-agnostic,
    level-based approach.
    """

    # --- Level 1: Outermost Structure Validation ---
    def _validate_outermost_structure(self, data, spec_name):
        """Validates the top-level keys of the spec file."""
        # --- Define all allowed top-level keys ---
        allowed_keys = {
            "description",
            "projects",
            "defaultRegion",
            "namePrefix",
            "nameSuffix",
        }

        # Check for any unexpected keys
        unexpected_keys = set(data.keys()) - allowed_keys
        self.assertFalse(
            unexpected_keys,
            f"Unexpected top-level keys found in '{spec_name}': {', '.join(unexpected_keys)}",
        )

        # --- Required Fields ---
        self.assertIn(
            "description", data, f"'{spec_name}' must have a 'description' key."
        )
        self.assertIsInstance(
            data["description"], str, "'description' must be a string."
        )
        self.assertTrue(data["description"], "'description' cannot be empty.")

        self.assertIn("projects", data, f"'{spec_name}' must have a 'projects' key.")
        self.assertIsInstance(data["projects"], list, "'projects' must be a list.")
        self.assertGreater(len(data["projects"]), 0, "'projects' list cannot be empty.")

        # --- Optional Fields ---
        optional_fields = ["defaultRegion", "namePrefix", "nameSuffix"]
        for field in optional_fields:
            if field in data:
                self.assertIsInstance(
                    data[field], str, f"Optional key '{field}' must be a string."
                )

    # --- Level 2: Project Structure Validation ---
    def _validate_project_structure(self, project, index, spec_name):
        """Validates the basic structure of a single project entry."""
        self.assertIsInstance(
            project,
            dict,
            f"Item at index {index} in 'projects' in '{spec_name}' is not a dictionary.",
        )
        self.assertIn(
            "projectId",
            project,
            f"Project at index {index} in '{spec_name}' must have a 'projectId'.",
        )
        self.assertIsInstance(
            project["projectId"],
            str,
            f"projectId for project {index} must be a string.",
        )
        self.assertTrue(
            project["projectId"], f"projectId for project {index} cannot be empty."
        )

    # --- Level 3: Consumers Structure Validation ---
    def _validate_consumers_structure(self, consumers, project_id, spec_name):
        """Validates the generic structure for a list of consumers."""
        for i, consumer in enumerate(consumers):
            context = (
                f"consumer at index {i} in project '{project_id}' in '{spec_name}'"
            )
            self.assertIsInstance(
                consumer, dict, f"Item in 'consumers' is not a dict: {context}"
            )
            self.assertIn("type", consumer, f"Missing 'type' key for {context}")
            self.assertIn("name", consumer, f"Missing 'name' key for {context}")

            consumer_type = consumer.get("type")

            # Serverless consumers have a different structure.
            if consumer_type in ["cloudrun_service", "cloudrun_job"]:
                self.assertIn(
                    "region",
                    consumer,
                    f"Consumer '{consumer['name']}' of type '{consumer_type}' must have a region.",
                )
                continue

            # For other types, they must have a valid network definition.
            has_network_interfaces = "networkInterfaces" in consumer
            has_vpc_subnet_keys = "vpc" in consumer and "subnet" in consumer

            self.assertTrue(
                has_network_interfaces or has_vpc_subnet_keys,
                f"Consumer '{consumer['name']}' ({consumer_type}) must have either a 'networkInterfaces' block or both 'vpc' and 'subnet' keys.",
            )

            if has_network_interfaces:
                self.assertIsInstance(
                    consumer["networkInterfaces"],
                    list,
                    f"'networkInterfaces' must be a list for {context}",
                )
                self.assertGreater(
                    len(consumer["networkInterfaces"]),
                    0,
                    f"'networkInterfaces' cannot be empty for {context}",
                )

    # --- Level 4: Producer Structure Validation ---
    def _validate_producers_structure(
        self, producers, project_id, spec_name, default_region=None
    ):
        """Validates the generic structure for a list of producers."""
        for i, producer in enumerate(producers):
            context = (
                f"producer at index {i} in project '{project_id}' in '{spec_name}'"
            )
            self.assertIsInstance(
                producer, dict, f"Item in 'producers' is not a dict: {context}"
            )
            self.assertIn("type", producer, f"Missing 'type' key for {context}")
            self.assertIn("name", producer, f"Missing 'name' key for {context}")

            # A producer must have a location, defined either on the resource
            # itself ('region' or 'location') OR at the top level of the spec file.
            has_local_location = "region" in producer or "location" in producer
            self.assertTrue(
                has_local_location or default_region,
                f"Missing 'region' or 'location' key for {context}, and no top-level 'defaultRegion' was found.",
            )

    # --- Level 5: VPC Structure Validation ---
    def _validate_vpc_structure(self, vpcs, project_id, spec_name):
        """Validates the generic structure for a VPC definition."""
        for i, vpc in enumerate(vpcs):
            context = f"VPC at index {i} in project '{project_id}' in '{spec_name}'"
            self.assertIsInstance(vpc, dict, f"Item in 'vpc' is not a dict: {context}")
            self.assertIn("type", vpc, f"Missing 'type' key for {context}")
            self.assertEqual(vpc["type"], "vpc", f"'type' must be 'vpc' for {context}")
            self.assertIn("name", vpc, f"Missing 'name' key for {context}")
            self.assertIn("subnets", vpc, f"Missing 'subnets' list for {context}")
            self.assertIsInstance(
                vpc["subnets"], list, f"'subnets' must be a list for {context}"
            )

    # --- Main Test Method ---
    def test_all_specs_are_valid(self):
        """
        Dynamically discovers and validates all .json files in the
        architecture-spec directory.
        """
        self.assertTrue(
            SPEC_DIR.is_dir(), f"Architecture spec directory not found at: {SPEC_DIR}"
        )
        spec_files = list(SPEC_DIR.glob("*.json"))
        self.assertGreater(
            len(spec_files), 0, "No architecture spec files (.json) found."
        )

        for spec_path in spec_files:
            with self.subTest(spec_file=spec_path.name):
                with open(spec_path, "r", encoding="utf-8") as f:
                    try:
                        data = json.load(f)
                    except json.JSONDecodeError as e:
                        self.fail(
                            f"File '{spec_path.name}' contains invalid JSON. Error: {e}"
                        )

                # Level 1 Validation
                self._validate_outermost_structure(data, spec_path.name)
                # Get the global default region to pass down to the producer check.
                default_region = data.get("defaultRegion")

                # Level 2-5 Validation
                for i, project in enumerate(data["projects"]):
                    self._validate_project_structure(project, i, spec_path.name)
                    project_id = project.get("projectId")

                    if "consumers" in project:
                        self._validate_consumers_structure(
                            project["consumers"], project_id, spec_path.name
                        )
                    if "producers" in project:
                        # Pass the default_region to the producer validation method.
                        self._validate_producers_structure(
                            project["producers"],
                            project_id,
                            spec_path.name,
                            default_region,
                        )
                    if "vpc" in project:
                        self._validate_vpc_structure(
                            project["vpc"], project_id, spec_path.name
                        )
