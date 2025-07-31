
# Copyright 2024-2025 Google LLC
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

output "storage_bucket_name" {
  description = "Google Cloud storage bucket name."
  value       = module.google_storage_bucket.name
}

output "organization_email" {
  description = "Organization stage service account IAM email."
  value       = module.organization.iam_email
}

output "networking_email" {
  description = "Networking stage service account IAM email."
  value       = module.networking.iam_email
}

output "dns_managed_zones_email" {
  description = "DNS Managed Zones stage service account IAM email."
  value       = module.dns_managed_zones.iam_email
}

output "dns_response_policy_email" {
  description = "DNS Response Policy stage service account IAM email."
  value       = module.dns_response_policy.iam_email
}

output "security_email" {
  description = "Security stage service account IAM email."
  value       = module.security.iam_email
}

output "producer_cloudsql_email" {
  description = "CloudSQL producer stage service account IAM email."
  value       = module.cloudsql_producer.iam_email
}

output "producer_alloydb_email" {
  description = "AlloyDB producer stage service account IAM email."
  value       = module.alloydb_producer.iam_email
}

output "producer_mrc_email" {
  description = "MRC producer stage service account IAM email."
  value       = module.mrc_producer.iam_email
}

output "producer_vertex_email" {
  description = "Vertex producer stage service account IAM email."
  value       = module.vertex_producer.iam_email
}

output "producer_gke_email" {
  description = "GKE producer stage service account IAM email."
  value       = module.gke_producer.iam_email
}

output "producer_bigquery_email" {
  description = "BigQuery producer stage service account IAM email."
  value       = module.bigquery_producer.iam_email
}

output "producer_connectivity_email" {
  description = "Producer Connectivity stage service account IAM email."
  value       = module.producer_connectivity.iam_email
}

output "consumer_gce_email" {
  description = "GCE consumer stage service account IAM email."
  value       = module.gce_consumer.iam_email
}

output "consumer_cloudrun_email" {
  description = "Cloud Run consumer stage service account IAM email."
  value       = module.cloudrun_consumer.iam_email
}

output "consumer_mig_email" {
  description = "MIG consumer stage service account IAM email."
  value       = module.mig_consumer.iam_email
}

output "consumer_vpc_access_connector_email" {
  description = "VPC Access Connector consumer stage service account IAM email."
  value       = module.consumer_vpc_access_connector.iam_email
}

output "consumer_appengine_email" {
  description = "App engine consumer stage service account IAM email."
  value       = module.appengine_consumer.iam_email
}

output "consumer_workbench_email" {
  description = "Workbench consumer stage service account IAM email."
  value       = module.workbench_consumer.iam_email
}

output "consumer_lb_email" {
  description = "Consumer Load Balancing stage service account IAM email."
  value       = module.consumer_load_balancing.iam_email
}

output "consumer_umig_email" {
  description = "UMIG consumer stage service account IAM email."
  value       = module.umig_consumer.iam_email
}