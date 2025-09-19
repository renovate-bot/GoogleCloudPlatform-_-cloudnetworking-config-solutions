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

module "packet_mirroring_rule" {
  source                 = "../../../modules/packet_mirroring_rule"
  for_each               = local.rules_map
  rule_name              = each.value.rule_name
  priority               = each.value.priority
  project_id             = each.value.project_id
  firewall_policy        = each.value.firewall_policy
  direction              = each.value.direction
  action                 = each.value.action
  security_profile_group = each.value.security_profile_group
  match_src_ip_ranges    = each.value.match_src_ip_ranges
  match_dest_ip_ranges   = each.value.match_dest_ip_ranges
  match_layer4_configs   = each.value.match_layer4_configs
  target_secure_tags     = each.value.target_secure_tags
  description            = each.value.description
  disabled               = each.value.disabled
  tls_inspect            = each.value.tls_inspect
}