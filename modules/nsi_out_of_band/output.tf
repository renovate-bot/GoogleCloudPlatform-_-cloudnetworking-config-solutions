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

output "deployment_group_id" {
  description = "The full resource ID of the Mirroring Deployment Group."
  value       = length(google_network_security_mirroring_deployment_group.deployment_group) > 0 ? google_network_security_mirroring_deployment_group.deployment_group[0].id : null
}

output "deployment_group_name" {
  description = "The full resource name of the Mirroring Deployment Group."
  value       = length(google_network_security_mirroring_deployment_group.deployment_group) > 0 ? google_network_security_mirroring_deployment_group.deployment_group[0].name : null
}

output "endpoint_group_id" {
  description = "The full resource ID of the Mirroring Endpoint Group."
  value       = length(google_network_security_mirroring_endpoint_group.endpoint_group) > 0 ? google_network_security_mirroring_endpoint_group.endpoint_group[0].id : null
}

output "endpoint_group_name" {
  description = "The full resource name of the Mirroring Endpoint Group."
  value       = length(google_network_security_mirroring_endpoint_group.endpoint_group) > 0 ? google_network_security_mirroring_endpoint_group.endpoint_group[0].name : null
}

output "deployments" {
  description = "A map of the created mirroring deployments, keyed by their short name."
  value = {
    for key, dep in google_network_security_mirroring_deployment.deployment : key => {
      id   = dep.id
      name = dep.name
    }
  }
}

output "endpoint_associations" {
  description = "A map of the created mirroring endpoint group associations, keyed by their short name."
  value = {
    for key, assoc in google_network_security_mirroring_endpoint_group_association.association : key => {
      id   = assoc.id
      name = assoc.name
    }
  }
}
