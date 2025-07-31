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

variable "config_folder_path" {
  description = "Location of YAML files holding Cloud DNS configuration values."
  type        = string
  default     = "../../../../configuration/networking/CloudDNS/DNSManagedZones/config"
}

variable "default_iam_bindings" {
  description = "Default IAM bindings to apply to DNS zones if not specified in YAML."
  type        = map(list(string))
  default     = {}
}

variable "default_description" {
  description = "Default description for DNS zones."
  type        = string
  default     = "Terraform managed DNS Zones"
}

variable "default_force_destroy" {
  description = "Default force_destroy value for DNS zones."
  type        = bool
  default     = false
}

variable "default_zone_config" {
  description = "Default zone_config for DNS zones."
  type        = any
  default     = null
}

variable "default_recordsets" {
  description = "Default recordsets for DNS zones."
  type        = map(any)
  default     = {}
}