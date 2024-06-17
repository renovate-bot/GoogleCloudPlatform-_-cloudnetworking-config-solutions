# Copyright 2024 Google LLC
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

# This variable defines a list of objects, each representing a single configuration
# for setting up a Private Service Connect (PSC) forwarding rule for a Cloud SQL instance.

variable "psc_endpoints" {
  description = "List of PSC Endpoint configurations"
  type = list(object({
    # The Google Cloud project ID where the Cloud SQL instance is located.
    producer_instance_project_id = string

    # The Google Cloud project ID where the forwarding rule and address will be created.
    endpoint_project_id = string

    # The name of the Cloud SQL instance to connect to.
    producer_instance_name = string

    # The name of the subnet where the internal IP address will be allocated.
    subnetwork_name = string

    # The name of the network where the forwarding rule will be created.
    network_name = string

    # Optional: The static internal IP address to use. If not provided,
    # Google Cloud will automatically allocate an IP address.
    ip_address_literal = optional(string, "")
    # Allow access to the PSC endpoint from any region.
    allow_psc_global_access = optional(bool, false)
    # Resource labels to apply to the forwarding rule.
    labels = optional(map(string), {})
  }))
}