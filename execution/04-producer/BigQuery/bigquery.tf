# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may not obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

module "bigquery" {
  source                          = "terraform-google-modules/bigquery/google"
  version                         = "~> 10.1"
  for_each                        = { for dataset in local.dataset_list : dataset.dataset_id => dataset }
  project_id                      = each.value.project_id
  dataset_id                      = each.value.dataset_id
  dataset_name                    = each.value.dataset_name
  description                     = each.value.description
  location                        = each.value.location
  access                          = each.value.access
  dataset_labels                  = each.value.dataset_labels
  default_partition_expiration_ms = each.value.default_partition_expiration_ms
  default_table_expiration_ms     = each.value.default_table_expiration_ms
  delete_contents_on_destroy      = each.value.delete_contents_on_destroy
  deletion_protection             = each.value.deletion_protection
  encryption_key                  = each.value.encryption_key
  max_time_travel_hours           = each.value.max_time_travel_hours
  storage_billing_model           = each.value.storage_billing_model
  resource_tags                   = each.value.resource_tags
  tables                          = each.value.tables
  views                           = each.value.views
  routines                        = each.value.routines
  materialized_views              = each.value.materialized_views
  external_tables                 = each.value.external_tables
}