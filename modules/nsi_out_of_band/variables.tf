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

variable "global_location" {
  description = "The location for global resources like deployment groups and endpoint groups."
  type        = string
  default     = "global"
}

variable "create_deployment_group" {
  description = "Set to true to create the Mirroring Deployment Group."
  type        = bool
  default     = false
}

variable "deployment_group_name" {
  description = "The name (ID) for the Mirroring Deployment Group."
  type        = string
  default     = null
}

variable "producer_network_link" {
  description = "The full resource link for the producer VPC network."
  type        = string
  default     = null
}

variable "deployment_group_description" {
  description = "A description for the Mirroring Deployment Group."
  type        = string
  default     = null
}

variable "deployment_group_labels" {
  description = "A map of labels to add to the Mirroring Deployment Group."
  type        = map(string)
  default     = {}
}

variable "existing_deployment_group_id" {
  description = "The full resource ID of an existing Mirroring Deployment Group to use if create_deployment_group is false."
  type        = string
  default     = null
}

variable "create_endpoint_group" {
  description = "Set to true to create the Mirroring Endpoint Group."
  type        = bool
  default     = false
}

variable "endpoint_group_name" {
  description = "The name (ID) for the Mirroring Endpoint Group."
  type        = string
  default     = null
}

variable "endpoint_group_description" {
  description = "A description for the Mirroring Endpoint Group."
  type        = string
  default     = null
}

variable "endpoint_group_labels" {
  description = "A map of labels to add to the Mirroring Endpoint Group."
  type        = map(string)
  default     = {}
}

variable "existing_endpoint_group_id" {
  description = "The full resource ID of an existing Mirroring Endpoint Group to use if create_endpoint_group is false."
  type        = string
  default     = null
}

variable "deployments" {
  description = "A list of mirroring deployments to create under the deployment group."
  type = list(object({
    deployment_project_id = string
    name                  = string
    location              = string
    description           = optional(string)
    labels                = optional(map(string))
    forwarding_rule_link  = string
  }))
  default = []
}

variable "endpoint_associations" {
  description = "A list of mirroring endpoint group associations to create under the endpoint group."
  type = list(object({
    endpoint_association_project_id = string
    name                            = string
    labels                          = optional(map(string))
    consumer_network_link           = string
  }))
  default = []
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
