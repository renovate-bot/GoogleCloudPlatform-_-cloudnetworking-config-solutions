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

locals {
  packet_mirroring_configs_data = [
    for file_path in fileset(var.config_folder_path, "*.y*ml") : {
      key     = trimsuffix(trimsuffix(basename(file_path), ".yaml"), ".yml")
      content = yamldecode(file("${var.config_folder_path}/${file_path}"))
    }
  ]
  packet_mirroring_list = [
    for item in local.packet_mirroring_configs_data : {
      key = item.key
      deployment_group = {
        create                      = try(item.content.deployment_group.create, var.create_deployment_group)
        deployment_group_project_id = try(item.content.deployment_group.deployment_group_project_id, var.deployment_group_project_id)
        name                        = try(item.content.deployment_group.name, var.deployment_group_name)
        producer_network_link       = try(item.content.deployment_group.producer_network_link, var.producer_network_link)
        description                 = try(item.content.deployment_group.description, var.deployment_group_description)
        labels                      = try(item.content.deployment_group.labels, var.deployment_group_labels)
      }
      endpoint_group = {
        create                    = try(item.content.endpoint_group.create, var.create_endpoint_group)
        endpoint_group_project_id = try(item.content.endpoint_group.endpoint_group_project_id, var.endpoint_group_project_id)
        name                      = try(item.content.endpoint_group.name, var.endpoint_group_name)
        description               = try(item.content.endpoint_group.description, var.endpoint_group_description)
        labels                    = try(item.content.endpoint_group.labels, var.endpoint_group_labels)
      }
      deployments                  = try(item.content.deployments, var.deployments)
      endpoint_associations        = try(item.content.endpoint_associations, var.endpoint_associations)
      existing_deployment_group_id = try(item.content.existing_deployment_group_id, var.existing_deployment_group_id)
      existing_endpoint_group_id   = try(item.content.existing_endpoint_group_id, var.existing_endpoint_group_id)
    }
  ]
  packet_mirroring_map = { for pm in local.packet_mirroring_list : pm.key => pm if pm.key != null }
  invalid_configs = {
    for key, config in local.packet_mirroring_map : key => config
    if config.deployment_group.producer_network_link != null && contains([for assoc in config.endpoint_associations : assoc.consumer_network_link], config.deployment_group.producer_network_link)
  }
}