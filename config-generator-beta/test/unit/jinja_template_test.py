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

import pathlib
import unittest
import yaml

from utilities import jinja_renderer

# Path to the real configuration directory
TEMPLATES_DIR = pathlib.Path(__file__).parent.parent.parent / "configuration"

# --- Mock Data Factory ---
# This dictionary maps a template filename to its specific mock context.
# To add a test for a new template, just add a new entry here.
MOCK_DATA_FACTORY = {
    # --- YAML Templates ---
    "gke.yaml.j2": {
        "projectId": "gke-proj",
        "instance": {
            "name": "main-cluster",
            "network": "net",
            "subnetwork": "subnet",
            "ipAllocationPolicy": {
                "clusterSecondaryRangeName": "pods",
                "servicesSecondaryRangeName": "svcs",
            },
            "initialClusterVersion": "1.28",
        },
    },
    "alloydb.yaml.j2": {
        "projectId": "db-proj",
        "instance": {
            "name": "main-db",
            "region": "us-central1",
            "machineCpuCount": 4,
            "readPoolInstance": [{"nodeCount": 2, "machineCpuCount": 2}],
            "machineConfig": {"cpuCount": 2},
        },
    },
    "gce.yaml.j2": {
        "projectId": "vm-proj",
        "instance": {
            "name": "client-vm",
            "zone": "us-central1-a",
            "networkInterfaces": [{"network": "net", "subnetwork": "sub"}],
            "tags": {"items": ["frontend"]},
        },
    },
    "memorystore_redis_cluster.yaml.j2": {
        "projectId": "redis-proj",
        "instance": {
            "name": "redis-cluster",
            "shardCount": 3,
            "network": "net",
            "region": "us-central1",
            "replicaCount": 1,
        },
    },
    "vertex_ai_vector_search.yaml.j2": {
        "projectId": "vector-proj",
        "instance": {
            "name": "vector-index",
            "region": "us-central1",
            "displayName": "vector-index",
            "indexUpdateMethod": "STREAMING_UPDATE",
            "dimension": 128,
            "approximateNeighborsCount": 10,
            "shardSize": "SHARD_SIZE_SMALL",
            "distanceMeasureType": "DOT_PRODUCT_DISTANCE",
            "network": "net",
            "treeAhConfig": {
                "leafNodeEmbeddingCount": 500,
                "leafNodesToSearchPercent": 7,
            },
            "bruteForceConfig": "{}",
        },
    },
    "vertex_ai_endpoint.yaml.j2": {
        "projectId": "vertex-proj",
        "instance": {
            "name": "pred-endpoint",
            "displayName": "pred-endpoint",
            "description": "desc",
            "region": "us-central1",
            "network": "net",
        },
    },
    "cloudsql.yaml.j2": {
        "projectId": "sql-proj",
        "instance": {
            "name": "sql-instance",
            "region": "us-central1",
            "databaseVersion": "POSTGRES_15",
            "settings": {},
        },
    },
    "serverless_vpc_connector.yaml.j2": {
        "projectId": "serverless-proj",
        "instance": {"name": "vpc-connector", "region": "us-central1"},
    },
    "app_engine_flexible.yaml.j2": {
        "projectId": "app-engine-proj",
        "instance": {
            "name": "app-flex",
            "version_id": "v1",
            "runtime": "python311",
            "instance_class": "F1",
            "network": {"name": "net", "subnetwork": "sub"},
            "entrypoint": {"shell": "start"},
            "deployment": {"zip": {"source_url": "url"}},
            "flexible_runtime_settings": {},
            "automatic_scaling": {},
            "liveness_check": {},
            "readiness_check": {},
        },
    },
    "app_engine_standard.yaml.j2": {
        "projectId": "app-engine-proj",
        "instance": {
            "name": "app-standard",
            "version_id": "v1",
            "runtime": "python311",
            "vpc_access_connector": {"name": "vpc-conn"},
            "handlers": [],
            "delete_service_on_destroy": False,
        },
    },
    "cloudrun_job.yaml.j2": {
        "projectId": "cloudrun-proj",
        "instance": {"name": "cr-job", "region": "us-central1"},
    },
    "cloudrun_service.yaml.j2": {
        "projectId": "cloudrun-proj",
        "instance": {"name": "cr-svc", "region": "us-central1"},
    },
    "workbench.yaml.j2": {
        "projectId": "workbench-proj",
        "instance": {
            "name": "wb-instance",
            "zone": "us-central1-a",
            "networkInterfaces": [{"network": "net", "subnet": "sub"}],
        },
    },
    "mig.yaml.j2": {
        "projectId": "mig-proj",
        "instance": {
            "name": "mig-instance",
            "region": "us-central1",
            "zone": "us-central1-a",
            "networkInterfaces": [
                {
                    "network": "projects/p/global/networks/net",
                    "subnetwork": "projects/p/regions/r/subnetworks/sub",
                }
            ],
        },
    },
    # --- TFVars Templates ---
    "networking.tfvars.j2": {
        "projectId": "net-proj",
        "region": "us-central1",
        "networkName": "main-net",
        "subnetsData": [
            {"name": "s1", "ipCidrRange": "10.0.0.0/24", "region": "us-central1"}
        ],
        "sharedVpcHost": "false",
        "primaryVpc": {},
    },
    "organisation.tfvars.j2": {
        "projectApis": {
            "proj-a": ["compute.googleapis.com"],
            "proj-b": ["sqladmin.googleapis.com"],
        }
    },
    "producer-connectivity.tfvars.j2": {
        "pscEndpointsData": [
            {
                "endpointProjectId": "ep-proj",
                "producerInstanceProjectId": "p-proj",
                "subnetworkName": "sub",
                "networkName": "net",
                "ipAddressName": "addr",
                "region": "us-central1",
                "producerType": "cloudsql",
                "producerName": "my-db",
            }
        ]
    },
    "vm.tfvars.j2": {
        "projectId": "proj",
        "networkPath": "path",
        "sourceRanges": ["0.0.0.0/0"],
    },
    # Generic context for simple security templates
    "simple_security_context": {
        "projectId": "sec-proj",
        "networkPath": "projects/sec-proj/global/networks/sec-net",
    },
}
# --- Add simple security templates to the factory ---
simple_sec_templates = [
    "alloydb.tfvars.j2",
    "cloudsql.tfvars.j2",
    "mig.tfvars.j2",
    "online_endpoint.tfvars.j2",
    "vector_search.tfvars.j2",
    "workbench.tfvars.j2",
    "memorystore_redis_cluster.tfvars.j2",
]
for tmpl in simple_sec_templates:
    MOCK_DATA_FACTORY[tmpl] = MOCK_DATA_FACTORY["simple_security_context"]


class TestRealJinjaTemplates(unittest.TestCase):
    """
    A scalable test that discovers and validates all REAL .j2 files
    using a mock data factory.
    """

    def test_all_real_jinja_templates_are_well_formed(self):
        """
        Discovers all .j2 files and validates them using specific mock data
        from the MOCK_DATA_FACTORY.
        """
        self.assertTrue(
            TEMPLATES_DIR.is_dir(),
            f"Base templates directory not found: {TEMPLATES_DIR}",
        )

        template_files = list(TEMPLATES_DIR.rglob("*.j2"))
        self.assertGreater(len(template_files), 0, "No .j2 templates found.")

        for template_path in template_files:
            template_name = template_path.name
            with self.subTest(template=str(template_path.relative_to(TEMPLATES_DIR))):
                # 1. Check if mock data exists for this template in the factory
                self.assertIn(
                    template_name,
                    MOCK_DATA_FACTORY,
                    f"--> Please add a mock data entry for '{template_name}' to the MOCK_DATA_FACTORY in this test file.",
                )

                # 2. Render the template with its specific mock data
                context = MOCK_DATA_FACTORY[template_name]
                rendered_string = jinja_renderer.render_template(
                    str(template_path.parent), template_name, context
                )
                self.assertIsNotNone(
                    rendered_string,
                    "Template failed to render. Check for Jinja2 syntax errors.",
                )

                # 3. If it's a YAML template, validate the output
                if template_name.endswith(".yaml.j2"):
                    try:
                        yaml.safe_load(rendered_string)
                    except yaml.YAMLError as e:
                        self.fail(f"Rendered output is not valid YAML. Error: {e}")
