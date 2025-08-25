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

output "deployment_groups" {
  description = "A map of all created Mirroring Deployment Groups, keyed by the config file name."
  value = {
    for key, instance in module.packet_mirroring : key => {
      id   = instance.deployment_group_id
      name = instance.deployment_group_name
    } if instance.deployment_group_id != null
  }
}

output "endpoint_groups" {
  description = "A map of all created Mirroring Endpoint Groups, keyed by the config file name."
  value = {
    for key, instance in module.packet_mirroring : key => {
      id   = instance.endpoint_group_id
      name = instance.endpoint_group_name
    } if instance.endpoint_group_id != null
  }
}

output "deployments" {
  description = "A combined map of all created deployments from all config files, keyed by their short name."
  value = {
    for key, instance in module.packet_mirroring : key => instance.deployments if instance.deployments != null
  }
}

output "endpoint_associations" {
  description = "A combined map of all created endpoint associations from all config files, keyed by their short name."
  value = {
    for key, instance in module.packet_mirroring : key => instance.endpoint_associations if instance.endpoint_associations != null
  }
}