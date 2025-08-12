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
  rules_configs_data = [
    for file_path in fileset(var.config_folder_path, "*.y*ml") : {
      key     = trimsuffix(trimsuffix(basename(file_path), ".yaml"), ".yml")
      content = yamldecode(file("${var.config_folder_path}/${file_path}"))
    }
  ]
  rules_list = [
    for rules_config in local.rules_configs_data : {
      rule_name              = try(rules_config.content.rule_name, var.rule_name)
      priority               = rules_config.content.priority
      project_id             = rules_config.content.project_id
      firewall_policy        = rules_config.content.firewall_policy_name
      direction              = rules_config.content.direction
      action                 = rules_config.content.action
      security_profile_group = try(rules_config.content.security_profile_group, var.security_profile_group)
      match_src_ip_ranges    = try(rules_config.content.match.src_ip_ranges, var.match_src_ip_ranges)
      match_dest_ip_ranges   = try(rules_config.content.match.dest_ip_ranges, var.match_dest_ip_ranges)
      match_layer4_configs   = rules_config.content.match.layer4_configs
      target_secure_tags     = try(rules_config.content.target_secure_tags, var.target_secure_tags)
      description            = try(rules_config.content.description, var.description)
      disabled               = try(rules_config.content.disabled, var.disabled)
      tls_inspect            = try(rules_config.content.tls_inspect, var.tls_inspect)
    }
  ]
  rules_map = { for rule in local.rules_list : rule.priority => rule }
}