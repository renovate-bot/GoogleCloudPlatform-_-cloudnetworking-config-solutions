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

variable "security_profile_group" {
  description = "Full URL of the Security Profile Group to which traffic will be mirrored."
  type        = string
  default     = null
}

variable "match_src_ip_ranges" {
  description = "List of source CIDR IP address ranges to match."
  type        = list(string)
  default     = []
}

variable "match_dest_ip_ranges" {
  description = "List of destination CIDR IP address ranges to match."
  type        = list(string)
  default     = []
}

variable "target_secure_tags" {
  description = "List of secure tag 'tagValues/name' strings to apply the rule to."
  type        = list(string)
  default     = []
}

variable "description" {
  description = "Description for the Packet mirroring rule."
  type        = string
  default     = "CNCS Packet Mirroring Rule"
}

variable "disabled" {
  description = "Boolean value for whether the rule is disabled."
  type        = bool
  default     = false
}

variable "tls_inspect" {
  description = "Boolean Value for whether traffic should be TLS decrypted."
  type        = bool
  default     = false
}

variable "rule_name" {
  description = "Name for the rule if not specified in the YAML file."
  type        = string
  default     = null
}

variable "config_folder_path" {
  description = "Path to the folder containing the YAML configuration files for packet mirroring rules."
  type        = string
  default     = "../../../configuration/network-security-integration/PacketMirroringRule/config"
}
