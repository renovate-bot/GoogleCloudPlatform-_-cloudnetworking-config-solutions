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

output "security_profiles" {
  description = "A map of all created security profiles with their details."
  value = {
    for key, profile in module.security_profile : key => {
      id   = profile.security_profile_id
      name = profile.security_profile_name
    } if profile.security_profile_id != null
  }
}

output "security_profile_groups" {
  description = "A map of all created security profile groups with their details."
  value = {
    for key, group in module.security_profile : key => {
      id   = group.security_profile_group_id
      name = group.security_profile_group_name
    } if group.security_profile_group_id != null
  }
}
