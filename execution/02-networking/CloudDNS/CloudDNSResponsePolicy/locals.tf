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
  raw_response_policies = {
    for f in fileset(var.config_folder_path, "*.yaml") :
    trimsuffix(f, ".yaml") => yamldecode(file("${var.config_folder_path}/${f}"))
  }

  response_policies_map = merge(flatten([
    for key, config in local.raw_response_policies :
    [
      for policy in config.response_policies : {
        "${key}-${policy.name}" = {
          name             = policy.name
          project_id       = policy.project_id
          networks         = policy.networks
          clusters         = try(policy.clusters, var.clusters)
          description      = try(policy.description, var.description)
          factories_config = try(policy.factories_config, var.factories_config)
          policy_create    = try(policy.policy_create, var.policy_create)
          rules = {
            for rule_item in try(policy.rules, []) :
            keys(rule_item)[0] => values(rule_item)[0]
          }
        }
      }
    ]
  ])...)

  response_policies_enabled_map = {
    for key, rp in local.response_policies_map :
    key => rp if rp.policy_create == true
  }
}