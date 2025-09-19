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
  description = "Default location for the mirroring deployment (e.g., 'us-central1-a')."
  type        = string
  default     = null
}

variable "global_location" {
  description = "Default location for global resources."
  type        = string
  default     = "global"
}

variable "producer_network_link" {
  description = "Default full resource link for the producer VPC network."
  type        = string
  default     = null
}

variable "consumer_network_link" {
  description = "Default full resource link for the consumer VPC network."
  type        = string
  default     = null
}

variable "forwarding_rule_link" {
  description = "Default full resource link for the forwarding rule."
  type        = string
  default     = null
}

variable "create_deployment_group" {
  description = "Controls creation of the Mirroring Deployment Group."
  type        = bool
  default     = false
}

variable "deployment_group_name" {
  description = "Default name for the Mirroring Deployment Group."
  type        = string
  default     = null
}

variable "deployment_group_description" {
  description = "Default description for the Mirroring Deployment Group."
  type        = string
  default     = null
}

variable "deployment_group_labels" {
  description = "Default labels for the Mirroring Deployment Group."
  type        = map(string)
  default     = {}
}

variable "existing_deployment_group_id" {
  description = "Default existing Mirroring Deployment Group ID to use."
  type        = string
  default     = null
}

variable "create_deployment" {
  description = "Controls creation of the Mirroring Deployment."
  type        = bool
  default     = false
}

variable "deployment_name" {
  description = "Default name for the Mirroring Deployment."
  type        = string
  default     = null
}

variable "deployment_description" {
  description = "Default description for the Mirroring Deployment."
  type        = string
  default     = null
}

variable "deployment_labels" {
  description = "Default labels for the Mirroring Deployment."
  type        = map(string)
  default     = {}
}

variable "create_endpoint_group" {
  description = "Controls creation of the Mirroring Endpoint Group."
  type        = bool
  default     = false
}

variable "endpoint_group_name" {
  description = "Default name for the Mirroring Endpoint Group."
  type        = string
  default     = null
}

variable "endpoint_group_description" {
  description = "Default description for the Mirroring Endpoint Group."
  type        = string
  default     = null
}

variable "endpoint_group_labels" {
  description = "Default labels for the Mirroring Endpoint Group."
  type        = map(string)
  default     = {}
}

variable "existing_endpoint_group_id" {
  description = "Default existing Mirroring Endpoint Group ID to use."
  type        = string
  default     = null
}

variable "create_association" {
  description = "Controls creation of the Mirroring Endpoint Group Association."
  type        = bool
  default     = false
}

variable "association_name" {
  description = "Default name for the Mirroring Endpoint Group Association."
  type        = string
  default     = null
}

variable "association_labels" {
  description = "Default labels for the Mirroring Endpoint Group Association."
  type        = map(string)
  default     = {}
}

variable "deployment_group" {
  description = "Provides a default empty map {} for the deployment_group object."
  type        = map(any)
  default     = {}
}

variable "endpoint_group" {
  description = "Provides a default empty map {} for the endpoint_group object."
  type        = map(any)
  default     = {}
}

variable "deployments" {
  description = "Provides a default empty list [] for the deployments list."
  type        = list(any)
  default     = []
}

variable "endpoint_associations" {
  description = "Provides a default empty list [] for the associations list."
  type        = list(any)
  default     = []
}

variable "endpoint_group_project_id" {
  description = "Project where endpoint group is to be deployed"
  type        = string
  default     = null
}

variable "deployment_group_project_id" {
  description = "Project where deployment group is to be deployed"
  type        = string
  default     = null
}

variable "config_folder_path" {
  description = "Path to the folder containing the YAML configuration files."
  type        = string
  default     = "../../../configuration/network-security-integration/OutOfBand/config/"
}