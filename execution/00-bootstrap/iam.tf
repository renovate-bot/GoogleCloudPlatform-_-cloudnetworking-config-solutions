
# Copyright 2024 Google LLC
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

/********************************************
 Service Account used to run Organization Stage
*********************************************/

module "organization" {
  source     = "github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account?ref=v38.1.0"
  project_id = var.bootstrap_project_id
  name       = var.organization_sa_name
  iam = {
    "roles/iam.serviceAccountTokenCreator" = var.organization_administrator
  }
  iam_project_roles = {
    (var.network_hostproject_id) = [
      "roles/iam.serviceAccountUser",
      "roles/serviceusage.serviceUsageAdmin",
    ]
    (var.network_serviceproject_id) = [
      "roles/iam.serviceAccountUser",
      "roles/serviceusage.serviceUsageAdmin",
    ]
  }
  iam_storage_roles = {
    (module.google_storage_bucket.name) = [
      "roles/storage.objectAdmin"
    ]
  }
}

/********************************************
 Service Account used to run Networking Stage
*********************************************/

module "networking" {
  source     = "github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account?ref=v38.1.0"
  project_id = var.bootstrap_project_id
  name       = var.networking_sa_name
  iam = {
    "roles/iam.serviceAccountTokenCreator" = var.networking_administrator
  }
  iam_folder_roles = {
    (var.folder_id) = [
      "roles/compute.xpnAdmin",
    ]
  }
  iam_project_roles = {
    (var.network_hostproject_id) = [
      "roles/compute.networkAdmin",
    ]
    (var.network_serviceproject_id) = [
      "roles/cloudsql.viewer"
    ]
  }
  iam_storage_roles = {
    (module.google_storage_bucket.name) = [
      "roles/storage.objectAdmin"
    ]
  }
}

/********************************************
 Service Account used to run Security Stage
*********************************************/

module "security" {
  source     = "github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account?ref=v38.1.0"
  project_id = var.bootstrap_project_id
  name       = var.security_sa_name
  iam = {
    "roles/iam.serviceAccountTokenCreator" = var.security_administrator
  }
  iam_project_roles = {
    (var.network_hostproject_id) = [
      "roles/compute.securityAdmin"
    ]
  }
  iam_storage_roles = {
    (module.google_storage_bucket.name) = [
      "roles/storage.objectAdmin"
    ]
  }
}

/********************************************
 Service Account used to run CloudSQL Producer Stage
*********************************************/

module "cloudsql_producer" {
  source     = "github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account?ref=v38.1.0"
  project_id = var.bootstrap_project_id
  name       = var.producer_cloudsql_sa_name
  iam = {
    "roles/iam.serviceAccountTokenCreator" = var.producer_cloudsql_administrator
  }
  iam_project_roles = {
    (var.network_serviceproject_id) = [
      "roles/cloudsql.admin"
    ]
  }
  iam_storage_roles = {
    (module.google_storage_bucket.name) = [
      "roles/storage.objectAdmin"
    ]
  }
}

/********************************************
 Service Account used to run AlloyDB Producer Stage
*********************************************/

module "alloydb_producer" {
  source     = "github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account?ref=v38.1.0"
  project_id = var.bootstrap_project_id
  name       = var.producer_alloydb_sa_name
  iam = {
    "roles/iam.serviceAccountTokenCreator" = var.producer_alloydb_administrator
  }
  iam_project_roles = {
    (var.network_serviceproject_id) = [
      "roles/alloydb.admin"
    ]
  }
  iam_storage_roles = {
    (module.google_storage_bucket.name) = [
      "roles/storage.objectAdmin"
    ]
  }
}

/********************************************
 Service Account used to run MRC Producer Stage
*********************************************/

module "mrc_producer" {
  source     = "github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account?ref=v38.1.0"
  project_id = var.bootstrap_project_id
  name       = var.producer_mrc_sa_name
  iam = {
    "roles/iam.serviceAccountTokenCreator" = var.producer_mrc_administrator
  }
  iam_project_roles = {
    (var.network_serviceproject_id) = [
      "roles/redis.admin"
    ]
  }
  iam_storage_roles = {
    (module.google_storage_bucket.name) = [
      "roles/storage.objectAdmin"
    ]
  }
}

/********************************************
 Service Account used to run Vertex AI Producer Stages
*********************************************/

module "vertex_producer" {
  source     = "github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account?ref=v38.1.0"
  project_id = var.bootstrap_project_id
  name       = var.producer_vertex_sa_name
  iam = {
    "roles/iam.serviceAccountTokenCreator" = var.producer_vertex_administrator
  }
  iam_project_roles = {
    (var.network_serviceproject_id) = [
      "roles/aiplatform.admin"
    ]
  }
  iam_storage_roles = {
    (module.google_storage_bucket.name) = [
      "roles/storage.objectAdmin"
    ]
  }
}

/********************************************
 Service Account used to run GKE Producer Stage
*********************************************/

module "gke_producer" {
  source     = "github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account?ref=v38.1.0"
  project_id = var.bootstrap_project_id
  name       = var.producer_gke_sa_name
  iam = {
    "roles/iam.serviceAccountTokenCreator" = var.producer_gke_administrator
  }
  iam_project_roles = {
    (var.network_serviceproject_id) = [
      "roles/container.admin",
      "roles/compute.instanceAdmin",
      "roles/iam.serviceAccountAdmin",
      "roles/iam.serviceAccountUser",
      "roles/resourcemanager.projectIamAdmin",
    ]
  }
  iam_storage_roles = {
    (module.google_storage_bucket.name) = [
      "roles/storage.objectAdmin"
    ]
  }
}

/****************************************************
 Service Account used to run Networking Manual Stage
*****************************************************/

module "networking_manual" {
  source     = "github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account?ref=v38.1.0"
  project_id = var.bootstrap_project_id
  name       = var.networking_manual_sa_name
  iam = {
    "roles/iam.serviceAccountTokenCreator" = var.networking_manual_administrator
  }
  iam_project_roles = {
    (var.network_hostproject_id) = [
      "roles/compute.networkAdmin",
    ]
    (var.network_serviceproject_id) = [
      "roles/cloudsql.viewer",
    ]
  }
  iam_storage_roles = {
    (module.google_storage_bucket.name) = [
      "roles/storage.objectAdmin"
    ]
  }
}

/********************************************
 Service Account used to run GCE Consumer Stage
*********************************************/

module "gce_consumer" {
  source     = "github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account?ref=v38.1.0"
  project_id = var.bootstrap_project_id
  name       = var.consumer_gce_sa_name
  iam = {
    "roles/iam.serviceAccountTokenCreator" = var.consumer_gce_administrator
  }
  iam_project_roles = {
    (var.network_hostproject_id) = [
      "roles/compute.networkUser",
    ]
    (var.network_serviceproject_id) = [
      "roles/compute.instanceAdmin.v1",
      "roles/iam.serviceAccountUser",
    ]
  }
  iam_storage_roles = {
    (module.google_storage_bucket.name) = [
      "roles/storage.objectAdmin"
    ]
  }
}

/********************************************
 Service Account used to run Cloud Run Consumer Stage
*********************************************/

module "cloudrun_consumer" {
  source     = "github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account?ref=v38.1.0"
  project_id = var.bootstrap_project_id
  name       = var.consumer_cloudrun_sa_name
  iam = {
    "roles/iam.serviceAccountTokenCreator" = var.consumer_cloudrun_administrator
  }
  iam_project_roles = {
    (var.network_hostproject_id) = [
      "roles/compute.networkUser",
    ]
    (var.network_serviceproject_id) = [
      "roles/iam.serviceAccountUser",
      "roles/run.admin"
    ]
  }
  iam_storage_roles = {
    (module.google_storage_bucket.name) = [
      "roles/storage.objectAdmin"
    ]
  }
}
