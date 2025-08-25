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

resource "google_network_security_mirroring_deployment_group" "deployment_group" {
  count                         = var.create_deployment_group ? 1 : 0
  project                       = var.deployment_group_project_id
  location                      = var.global_location
  mirroring_deployment_group_id = var.deployment_group_name
  network                       = var.producer_network_link
  description                   = var.deployment_group_description
  labels                        = var.deployment_group_labels
}

resource "google_network_security_mirroring_endpoint_group" "endpoint_group" {
  count                       = var.create_endpoint_group ? 1 : 0
  project                     = var.endpoint_group_project_id
  location                    = var.global_location
  mirroring_endpoint_group_id = var.endpoint_group_name
  mirroring_deployment_group  = local.deployment_group_id_to_use
  description                 = var.endpoint_group_description
  labels                      = var.endpoint_group_labels
}

resource "google_network_security_mirroring_deployment" "deployment" {
  for_each                   = { for dep in var.deployments : dep.name => dep }
  project                    = each.value.deployment_project_id
  location                   = each.value.location
  mirroring_deployment_id    = each.key
  mirroring_deployment_group = local.deployment_group_id_to_use
  forwarding_rule            = each.value.forwarding_rule_link
  description                = each.value.description
  labels                     = each.value.labels
}

resource "google_network_security_mirroring_endpoint_group_association" "association" {
  for_each                                = { for assoc in var.endpoint_associations : assoc.name => assoc }
  project                                 = each.value.endpoint_association_project_id
  location                                = var.global_location
  mirroring_endpoint_group_association_id = each.key
  mirroring_endpoint_group                = local.endpoint_group_id_to_use
  network                                 = each.value.consumer_network_link
  labels                                  = each.value.labels
}
