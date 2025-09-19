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

resource "google_compute_network_firewall_policy_packet_mirroring_rule" "primary" {
  provider               = google-beta
  priority               = var.priority
  rule_name              = var.rule_name
  firewall_policy        = var.firewall_policy
  project                = var.project_id
  direction              = var.direction
  action                 = var.action
  security_profile_group = var.security_profile_group
  description            = var.description
  disabled               = var.disabled
  tls_inspect            = var.tls_inspect
  match {
    src_ip_ranges  = var.match_src_ip_ranges
    dest_ip_ranges = var.match_dest_ip_ranges
    dynamic "layer4_configs" {
      for_each = var.match_layer4_configs
      content {
        ip_protocol = layer4_configs.value.ip_protocol
        ports       = layer4_configs.value.ports
      }
    }
  }
  dynamic "target_secure_tags" {
    for_each = var.target_secure_tags
    content {
      name = target_secure_tags.value
    }
  }
}