/**
 * Copyright 2024 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

data "google_sql_database_instance" "instance" {
  for_each = { for idx, config in var.psc_endpoints : idx => config if config.producer_instance_name != null }
  project  = each.value.endpoint_project_id
  name     = each.value.producer_instance_name
}

resource "google_compute_address" "psc_address" {
  for_each     = { for idx, config in var.psc_endpoints : idx => config if config.ip_address_literal != null }
  project      = each.value.endpoint_project_id
  name         = "psc-compute-address-${each.value.producer_instance_name != null ? each.value.producer_instance_name : "custom-${each.key}"}"
  region       = each.value.region != null ? each.value.region : (each.value.producer_instance_name != null ? data.google_sql_database_instance.instance[each.key].region : split("/", each.value.target)[3])
  address_type = "INTERNAL"
  subnetwork   = each.value.subnetwork_name
  address      = each.value.ip_address_literal
}

resource "google_compute_forwarding_rule" "psc_forwarding_rule" {
  for_each              = { for idx, config in var.psc_endpoints : idx => config }
  project               = each.value.endpoint_project_id
  name                  = "psc-forwarding-rule-${each.value.producer_instance_name != null ? each.value.producer_instance_name : "custom-${each.key}"}"
  region                = each.value.region != null ? each.value.region : (each.value.producer_instance_name != null ? data.google_sql_database_instance.instance[each.key].region : split("/", each.value.target)[3])
  network               = each.value.network_name
  ip_address            = contains(keys(google_compute_address.psc_address), each.key) ? google_compute_address.psc_address[each.key].self_link : null
  load_balancing_scheme = ""
  target                = local.forwarding_rule_targets[each.key]
}