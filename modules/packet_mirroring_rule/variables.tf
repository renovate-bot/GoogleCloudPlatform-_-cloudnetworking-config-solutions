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

variable "priority" {
  description = "The priority of the rule (0-2147483647). Lower numbers have higher priority."
  type        = number
}

variable "rule_name" {
  description = "A user-friendly rule name for the packet mirroing rule."
  type        = string
  default     = "cncs-mirroring-rule"
}

variable "firewall_policy" {
  description = "The name of the parent firewall policy."
  type        = string
}

variable "project_id" {
  description = "The project ID where the firewall policy resides."
  type        = string
}

variable "direction" {
  description = "The direction of traffic to which the rule applies. Can be 'INGRESS' or 'EGRESS'."
  type        = string
}

variable "action" {
  description = "The action to perform. Can be 'mirror', 'do_not_mirror', or 'goto_next'."
  type        = string
  default     = "mirror"
}

variable "security_profile_group" {
  description = "The full URL of the Security Profile Group to which traffic will be mirrored. Required if action is 'mirror'."
  type        = string
  default     = null
}

variable "match_src_ip_ranges" {
  description = "A list of source CIDR IP address ranges to match."
  type        = list(string)
  default     = []
}

variable "match_dest_ip_ranges" {
  description = "A list of destination CIDR IP address ranges to match."
  type        = list(string)
  default     = []
}

variable "match_layer4_configs" {
  description = "A list of L4 configs to match (protocol and optional ports)."
  type = list(object({
    ip_protocol = string
    ports       = optional(list(string))
  }))
}

variable "target_secure_tags" {
  description = "A list of secure tag 'tagValues/name' strings to apply the rule to."
  type        = list(string)
  default     = []
}

variable "description" {
  description = "An optional description for the rule."
  type        = string
  default     = "CNCS Packet Mirroring Rule"
}

variable "disabled" {
  description = "Whether the rule is disabled."
  type        = bool
  default     = false
}

variable "tls_inspect" {
  description = "Boolean flag indicating if the traffic should be TLS decrypted."
  type        = bool
  default     = false
}