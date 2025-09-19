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

module "security_profile" {
  for_each                              = local.security_profiles_map
  source                                = "../../../modules/security_profile/"
  organization_id                       = each.value.organization_id
  location                              = each.value.location
  create_security_profile               = each.value.security_profile.create
  security_profile_name                 = each.value.security_profile.name
  security_profile_type                 = each.value.security_profile.type
  security_profile_description          = each.value.security_profile.description
  security_profile_labels               = each.value.security_profile.labels
  threat_prevention_profile             = each.value.security_profile.threat_prevention_profile
  custom_mirroring_profile              = each.value.security_profile.custom_mirroring_profile
  custom_intercept_profile              = each.value.security_profile.custom_intercept_profile
  create_security_profile_group         = each.value.security_profile_group.create
  security_profile_group_name           = each.value.security_profile_group.name
  security_profile_group_description    = each.value.security_profile_group.description
  security_profile_group_labels         = each.value.security_profile_group.labels
  link_security_profile_to_group        = each.value.link_profile_to_group
  existing_threat_prevention_profile_id = each.value.security_profile_group.existing_threat_prevention_profile_id
  existing_custom_mirroring_profile_id  = each.value.security_profile_group.existing_custom_mirroring_profile_id
  existing_custom_intercept_profile_id  = each.value.security_profile_group.existing_custom_intercept_profile_id
}
