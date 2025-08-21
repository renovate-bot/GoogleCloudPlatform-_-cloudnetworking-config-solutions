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

locals {
  config_folder_path = var.config_folder_path
  datasets           = [for file in fileset(local.config_folder_path, "[^_]*.yaml") : yamldecode(file("${local.config_folder_path}/${file}"))]
  dataset_list = flatten([
    for dataset in local.datasets : {
      project_id                      = dataset.project_id
      dataset_id                      = dataset.dataset_id
      dataset_name                    = dataset.dataset_name
      description                     = try(dataset.description, var.description)
      location                        = try(dataset.location, var.location)
      access                          = try(dataset.access, var.access)
      dataset_labels                  = try(dataset.dataset_labels, var.dataset_labels)
      default_partition_expiration_ms = try(dataset.default_partition_expiration_ms, var.default_partition_expiration_ms)
      default_table_expiration_ms     = try(dataset.default_table_expiration_ms, var.default_table_expiration_ms)
      delete_contents_on_destroy      = try(dataset.delete_contents_on_destroy, var.delete_contents_on_destroy)
      deletion_protection             = try(dataset.deletion_protection, var.deletion_protection)
      encryption_key                  = try(dataset.encryption_key, var.encryption_key)
      max_time_travel_hours           = try(dataset.max_time_travel_hours, var.max_time_travel_hours)
      storage_billing_model           = try(dataset.storage_billing_model, var.storage_billing_model)
      resource_tags                   = try(dataset.resource_tags, var.resource_tags)
      tables                          = try(dataset.tables, var.tables)
      views                           = try(dataset.views, var.views)
      routines                        = try(dataset.routines, var.routines)
      materialized_views              = try(dataset.materialized_views, var.materialized_views)
      external_tables                 = try(dataset.external_tables, var.external_tables)
    }
  ])
}