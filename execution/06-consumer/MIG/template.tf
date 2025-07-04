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

module "mig-template" {
  for_each   = local.mig_map
  source     = "github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/compute-vm?ref=v41.0.0"
  project_id = each.value.project_id
  name       = var.mig_template_name
  zone       = each.value.zone
  tags       = var.tags
  network_interfaces = [{
    network    = local.vpc_self_links[each.key]
    subnetwork = local.subnetwork_self_links[each.key]
    nat        = var.create_nat
    addresses  = null
  }]
  boot_disk = {
    initialize_params = {
      image = var.mig_image
    }
  }
  create_template = var.create_template
  metadata = {
    startup-script = var.metadata
  }
}
