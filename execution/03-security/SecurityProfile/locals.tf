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
  security_configs_with_keys = [
    for file_path in fileset(var.config_folder_path, "*.y*ml") : {
      key     = trimsuffix(trimsuffix(basename(file_path), ".yaml"), ".yml")
      content = yamldecode(file("${var.config_folder_path}/${file_path}"))
    }
  ]
  security_profiles_list = [
    for item in local.security_configs_with_keys : {
      key             = item.key
      organization_id = item.content.organization_id
      location        = try(item.content.location, var.location)
      security_profile = {
        create                    = try(item.content.security_profile.create, var.create_security_profile)
        name                      = try(item.content.security_profile.name, var.security_profile_name)
        type                      = try(item.content.security_profile.type, var.security_profile_type)
        description               = try(item.content.security_profile.description, var.security_profile_description)
        labels                    = try(item.content.security_profile.labels, var.security_profile_labels)
        threat_prevention_profile = try(item.content.security_profile.threat_prevention_profile, var.threat_prevention_profile)
        custom_mirroring_profile  = try(item.content.security_profile.custom_mirroring_profile, var.custom_mirroring_profile)
        custom_intercept_profile  = try(item.content.security_profile.custom_intercept_profile, var.custom_intercept_profile)
      }
      security_profile_group = {
        create                                = try(item.content.security_profile_group.create, var.create_security_profile_group)
        name                                  = try(item.content.security_profile_group.name, var.security_profile_group_name)
        description                           = try(item.content.security_profile_group.description, var.security_profile_group_description)
        labels                                = try(item.content.security_profile_group.labels, var.security_profile_group_labels)
        existing_threat_prevention_profile_id = try(item.content.security_profile_group.existing_threat_prevention_profile_id, var.existing_threat_prevention_profile_id)
        existing_custom_mirroring_profile_id  = try(item.content.security_profile_group.existing_custom_mirroring_profile_id, var.existing_custom_mirroring_profile_id)
        existing_custom_intercept_profile_id  = try(item.content.security_profile_group.existing_custom_intercept_profile_id, var.existing_custom_intercept_profile_id)
      }
      link_profile_to_group = try(item.content.link_profile_to_group, var.link_profile_to_group)
    }
  ]
  security_profiles_map = { for sp in local.security_profiles_list : sp.key => sp }
}
