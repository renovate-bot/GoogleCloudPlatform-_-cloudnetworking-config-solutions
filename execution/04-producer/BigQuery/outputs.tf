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

output "bigquery_dataset_details" {
  description = "A map of the created BigQuery datasets and their nested resources."
  value = { for key, dataset in module.bigquery : key => {
    dataset_id         = dataset.bigquery_dataset.id
    dataset_self_link  = dataset.bigquery_dataset.self_link
    table_ids          = dataset.table_ids
    view_ids           = dataset.view_ids
    routine_ids        = dataset.routine_ids
    external_table_ids = dataset.external_table_ids
  } }
}