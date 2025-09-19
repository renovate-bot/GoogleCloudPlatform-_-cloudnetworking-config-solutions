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


variable "location" {
  description = "Default location for resources if not specified in YAML files."
  type        = string
  default     = "global"
}

variable "create_security_profile" {
  description = "Default value for creating a security profile."
  type        = bool
  default     = false
}

variable "security_profile_name" {
  description = "Default name for a security profile."
  type        = string
  default     = null
}

variable "security_profile_type" {
  description = "Default type for a security profile."
  type        = string
  default     = "THREAT_PREVENTION"
}

variable "security_profile_description" {
  description = "Default description for a security profile."
  type        = string
  default     = "CNCS terraform security profile"
}

variable "security_profile_labels" {
  description = "Default labels for a security profile."
  type        = map(string)
  default     = {}
}

variable "threat_prevention_profile" {
  description = "Default configuration for a threat prevention profile."
  type = object({
    severity_overrides = optional(list(object({
      severity = string
      action   = string
    })), [])
    threat_overrides = optional(list(object({
      threat_id = string
      action    = string
    })), [])
    antivirus_overrides = optional(list(object({
      protocol = string
      action   = string
    })), [])
  })
  default = null
}

variable "custom_mirroring_profile" {
  description = "Default configuration for a custom mirroring profile."
  type = object({
    mirroring_endpoint_group = string
  })
  default = null
}

variable "custom_intercept_profile" {
  description = "Default configuration for a custom intercept profile."
  type = object({
    intercept_endpoint_group = string
  })
  default = null
}

variable "create_security_profile_group" {
  description = "Default value for creating a security profile group."
  type        = bool
  default     = false
}

variable "security_profile_group_name" {
  description = "Default name for a security profile group."
  type        = string
  default     = null
}

variable "security_profile_group_description" {
  description = "Default description for a security profile group."
  type        = string
  default     = null
}
variable "security_profile_group_labels" {
  description = "Default labels for a security profile group."
  type        = map(string)
  default     = {}
}

variable "link_profile_to_group" {
  description = "Default value for linking a profile to a group."
  type        = bool
  default     = false
}

variable "existing_threat_prevention_profile_id" {
  description = "Default value for an existing threat prevention profile ID."
  type        = string
  default     = null
}

variable "existing_custom_mirroring_profile_id" {
  description = "Default value for an existing custom mirroring profile ID."
  type        = string
  default     = null
}

variable "existing_custom_intercept_profile_id" {
  description = "Default value for an existing custom intercept profile ID."
  type        = string
  default     = null
}

variable "config_folder_path" {
  description = "Path to the folder containing the YAML configuration files for security profiles and groups."
  type        = string
  default     = "../../../configuration/network-security-integration/SecurityProfile/config"
}
