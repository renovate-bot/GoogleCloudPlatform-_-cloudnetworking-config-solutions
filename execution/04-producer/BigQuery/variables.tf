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

variable "access" {
  description = "An array of objects that define dataset access for one or more entities."
  type        = any
  default     = []
}

variable "dataset_labels" {
  description = "Key value pairs in a map for dataset labels"
  type        = map(string)
  default     = {}
}

variable "dataset_name" {
  description = "Friendly name for the dataset being provisioned."
  type        = string
  default     = null
}

variable "default_partition_expiration_ms" {
  description = "The default partition expiration for all partitioned tables in the dataset, in MS."
  type        = number
  default     = null
}

variable "default_table_expiration_ms" {
  description = "TTL of tables using the dataset in MS."
  type        = number
  default     = null
}

variable "delete_contents_on_destroy" {
  description = "If set to true, delete all the tables in the dataset when destroying the resource; otherwise, destroying the resource will fail if tables are present."
  type        = bool
  default     = false
}

variable "deletion_protection" {
  description = "Whether or not to allow deletion of tables defined by this module."
  type        = bool
  default     = false
}

variable "description" {
  description = "Dataset description."
  type        = string
  default     = "Terraform managed dataset created using the CNCS repository automation."
}

variable "encryption_key" {
  description = "Default encryption key to apply to the dataset. Defaults to null (Google-managed)."
  type        = string
  default     = null
}

variable "location" {
  description = "The location of the dataset. For multi-region, US or EU can be provided."
  type        = string
  default     = "US"
}

variable "max_time_travel_hours" {
  description = "Defines the time travel window in hours."
  type        = number
  default     = null
}

variable "resource_tags" {
  description = "A map of resource tags to add to the dataset."
  type        = map(string)
  default     = {}
}

variable "storage_billing_model" {
  description = "Specifies the storage billing model for the dataset (LOGICAL or PHYSICAL)."
  type        = string
  default     = null
}

variable "tables" {
  description = "A list of table objects to create in the dataset."
  type = list(object({
    table_id                    = string
    schema                      = optional(string)
    clustering                  = optional(list(string))
    time_partitioning           = optional(any)
    range_partitioning          = optional(any)
    expiration_time             = optional(number)
    friendly_name               = optional(string)
    description                 = optional(string)
    labels                      = optional(map(string))
    require_partition_filter    = optional(bool)
    deletion_protection         = optional(bool)
    max_staleness               = optional(string)
    encryption_configuration    = optional(object({ kms_key_name = string }))
    table_constraints           = optional(any)
    view                        = optional(any)
    materialized_view           = optional(any)
    external_data_configuration = optional(any)
    table_replication_info      = optional(object({ source_project_id = string, source_dataset_id = string, source_table_id = string, replication_interval_ms = number }))
  }))
  default = []
}

variable "views" {
  description = "A list of view objects to create in the dataset."
  type = list(object({
    view_id             = string
    query               = string
    use_legacy_sql      = optional(bool, false)
    labels              = optional(map(string), {})
    friendly_name       = optional(string)
    description         = optional(string)
    deletion_protection = optional(bool)
  }))
  default = []
}

variable "routines" {
  description = "A list of routine objects to create in the dataset."
  type = list(object({
    routine_id      = string
    routine_type    = string
    language        = string
    definition_body = string
    arguments = optional(list(object({
      name          = string
      data_type     = string
      argument_kind = optional(string)
      mode          = optional(string)
    })), [])
    return_type        = optional(string)
    return_table_type  = optional(string)
    imported_libraries = optional(list(string))
    description        = optional(string)
    determinism_level  = optional(string)
  }))
  default = []
}

variable "materialized_views" {
  description = "A list of materialized view objects to create in the dataset."
  type = list(object({
    view_id                          = string
    query                            = string
    enable_refresh                   = optional(bool, true)
    refresh_interval_ms              = optional(number)
    labels                           = optional(map(string), {})
    friendly_name                    = optional(string)
    description                      = optional(string)
    deletion_protection              = optional(bool)
    allow_non_incremental_definition = optional(bool)
    max_staleness                    = optional(string)
  }))
  default = []
}

variable "external_tables" {
  description = "A list of external table objects to create in the dataset."
  type = list(object({
    table_id              = string
    source_uris           = list(string)
    source_format         = string
    autodetect            = bool
    compression           = optional(string)
    connection_id         = optional(string)
    ignore_unknown_values = optional(bool)
    max_bad_records       = optional(number)
    schema                = optional(string)
    csv_options = optional(object({
      quote                 = optional(string)
      allow_jagged_rows     = optional(bool)
      allow_quoted_newlines = optional(bool)
      encoding              = optional(string)
      field_delimiter       = optional(string)
      skip_leading_rows     = optional(number)
    }))
    google_sheets_options = optional(object({
      skip_leading_rows = optional(number)
      range             = optional(string)
    }))
    hive_partitioning_options = optional(object({
      mode              = optional(string)
      source_uri_prefix = optional(string)
    }))
    json_options = optional(object({
      encoding = optional(string)
    }))
    labels              = optional(map(string), {})
    friendly_name       = optional(string)
    description         = optional(string)
    deletion_protection = optional(bool)
  }))
  default = []
}

variable "config_folder_path" {
  description = "Location of YAML files holding BigQuery dataset configuration values."
  type        = string
  default     = "../../../configuration/producer/BigQuery/config"
}