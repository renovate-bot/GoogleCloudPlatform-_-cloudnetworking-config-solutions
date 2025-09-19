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
  dns_configs_raw = {
    for file in fileset(var.config_folder_path, "*.yaml") :
    trimsuffix(file, ".yaml") => yamldecode(file("${var.config_folder_path}/${file}"))
  }

  dns_zones_map = merge(flatten([
    for key, config in local.dns_configs_raw :
    [
      for zone in config.zones : {
        "${key}-${zone.zone}" = {
          project_id    = zone.project_id
          name          = zone.zone
          description   = try(zone.description, var.default_description)
          force_destroy = try(zone.force_destroy, var.default_force_destroy)
          iam           = try(zone.iam, var.default_iam_bindings)
          zone_config   = try(zone.zone_config, var.default_zone_config)
          recordsets    = { for rs in try(zone.recordsets, []) : "${rs.type} ${rs.name}" => { ttl = rs.ttl, records = rs.records } }
        }
      }
    ]
  ])...)
}