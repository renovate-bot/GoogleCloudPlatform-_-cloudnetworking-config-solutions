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

resource "null_resource" "prevent_invalid_apply" {
  lifecycle {
    precondition {
      condition     = length(local.invalid_configs) == 0
      error_message = "ERROR: Plan failed. The producer_network_link cannot be the same as a consumer_network_link in the following files: ${join(", ", [for f in keys(local.invalid_configs) : "${f}.yml"])}"
    }
  }
}

module "packet_mirroring" {
  for_each                     = local.packet_mirroring_map
  source                       = "../../../modules/nsi_out_of_band/"
  existing_deployment_group_id = each.value.existing_deployment_group_id
  existing_endpoint_group_id   = each.value.existing_endpoint_group_id
  create_deployment_group      = each.value.deployment_group.create
  deployment_group_project_id  = each.value.deployment_group.deployment_group_project_id
  deployment_group_name        = each.value.deployment_group.name
  producer_network_link        = each.value.deployment_group.producer_network_link
  deployment_group_description = each.value.deployment_group.description
  deployment_group_labels      = each.value.deployment_group.labels
  create_endpoint_group        = each.value.endpoint_group.create
  endpoint_group_project_id    = each.value.endpoint_group.endpoint_group_project_id
  endpoint_group_name          = each.value.endpoint_group.name
  endpoint_group_description   = each.value.endpoint_group.description
  endpoint_group_labels        = each.value.endpoint_group.labels
  deployments                  = each.value.deployments
  endpoint_associations        = each.value.endpoint_associations
}