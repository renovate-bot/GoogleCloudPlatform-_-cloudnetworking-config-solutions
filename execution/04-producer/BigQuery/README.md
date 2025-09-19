## Introduction

**BigQuery** is Google's fully managed, serverless data warehouse that enables super-fast SQL queries using the processing power of Google's infrastructure. It allows you to collect, store, and analyze massive amounts of data in real-time without having to manage any infrastructure.

BigQuery is your ideal data solution if you're looking for:

  * **Massive Scalability**: Effortlessly scale from gigabytes to petabytes of data.
  * **High Performance**: Run complex analytical queries on large datasets in seconds.
  * **Serverless Architecture**: No infrastructure to manage, so you can focus on analyzing data.
  * **Built-in ML & BI**: Leverage integrated tools like BigQuery ML and connect seamlessly with BI tools like Looker Studio.

## Pre-Requisites

Before creating your first BigQuery dataset, ensure you have completed the following prerequisites:

1.  **Completed Prior Stages**: Successful deployment of BigQuery resources depends on the completion of the following stage:

      * **01-organization**: This stage handles the creation of your Google Cloud project and the activation of required APIs.

2.  **API Enablement**: Ensure the following Google Cloud APIs have been enabled in your project:

      * IAM API (`iam.googleapis.com`)
      * BigQuery API (`bigquery.googleapis.com`)
      * Cloud Resource Manager API (`cloudresourcemanager.googleapis.com`)

3.  **IAM Permissions**: Grant the user or service account executing Terraform the following IAM role at the project level:

      * **BigQuery Admin** (`roles/bigquery.admin`) or **BigQuery Data Owner** (`roles/bigquery.dataOwner`).

-----

## Let's Get Started\! 🚀

With the prerequisites in place and your BigQuery configuration files ready, you can now leverage Terraform to automate the creation of your datasets, tables, and views.

### Execution Steps

1.  **Create your configuration files**:

      * Create YAML files defining the properties of each BigQuery dataset you want to create. Store these files in the `configuration/producer/BigQuery/config` folder.
      * Each YAML file should map to a single dataset, providing details such as `dataset_id`, `location`, and any `tables` or `views` you want to include.
      * For reference, see the [example section](#example) below or the sample file at `configuration/producer/BigQuery/config/dataset.yaml.example`.

2.  **Initialize Terraform**:

      * Open your terminal and navigate to the directory containing the BigQuery Terraform configuration.
      * Run the following command to initialize Terraform:
        ```bash
        terraform init
        ```

3.  **Review the Execution Plan**:

      * Use the `terraform plan` command to see the changes Terraform will make to your infrastructure. You will need a `bigquery.tfvars` file in the appropriate configuration directory to specify shared variables if any.
        ```bash
        terraform plan -var-file=../../../configuration/producer/BigQuery/bigquery.tfvars
        ```
      * Carefully review the plan to ensure it aligns with your intended configuration.

4.  **Apply the Configuration**:

      * Once you're satisfied, execute the `terraform apply` command to provision your BigQuery resources:
        ```bash
        terraform apply -var-file=../../../configuration/producer/BigQuery/bigquery.tfvars
        ```
      * Terraform will read the YAML files from the `configuration/producer/BigQuery/config` folder and create the corresponding resources in your Google Cloud project.

5.  **Monitor and Manage**:

      * After the resources are created, you can interact with them through the Google Cloud Console or the `bq` command-line tool.
      * Continue to use Terraform to manage updates and changes to your BigQuery resources declaratively.

-----

### Example {#example}

To help you get started, we've provided examples of YAML configuration files that you can use as templates.

  * **Minimal YAML (Dataset only)**: This example includes only the essential fields required to create a basic BigQuery dataset.

    ```yaml
    # A simple dataset for marketing analytics.
    project_id: <your-gcp-project-id>
    dataset_id: marketing_analytics
    dataset_name: Marketing Analytics
    description: Dataset for storing marketing campaign data.
    dataset_labels:
      owner: <marketing-team>
      environment: production
    ```

  * **Comprehensive YAML (Dataset with Table and View)**: This example includes more advanced configurations, such as a table with a defined schema and a view.

    ```yaml
    # A comprehensive dataset for the finance department.
    project_id: <your-gcp-project-id>
    dataset_id: finance_reports
    dataset_name: Finance Reports
    description: Confidential dataset for quarterly finance reports.
    location: EU

    # Grant the finance-auditors group the Data Viewer role.
    access:
      - role: "roles/bigquery.dataViewer"
        group_by_email: <finance-auditors@your-company.com>

    # Define a table for quarterly earnings.
    tables:
      - table_id: quarterly_earnings
        description: Table containing raw data for quarterly earnings reports.
        schema: >
          [
            {"name": "report_id", "type": "STRING", "mode": "REQUIRED"},
            {"name": "fiscal_quarter", "type": "STRING", "mode": "NULLABLE"},
            {"name": "revenue", "type": "NUMERIC", "mode": "NULLABLE"}
          ]

    # Define a view to show revenue from the earnings table.
    views:
      - view_id: quarterly_revenue_view
        use_legacy_sql: false
        query: "SELECT fiscal_quarter, revenue FROM `${project_id}.${dataset_id}.quarterly_earnings`"
    ```

-----

## Important Notes:

  * This `README.md` is a starting point. Customize it to include specific details about your data warehousing projects.
  * Refer to the official [Google Cloud BigQuery documentation](https://cloud.google.com/bigquery/docs) for the most up-to-date information and best practices.
  * **Order of Execution**: Ensure you have completed the `01-organization` stage before attempting to create BigQuery datasets.
  * **Troubleshooting**: If you encounter errors, verify that all prerequisites are satisfied and that the service account or user has the correct IAM permissions.

<!-- BEGIN_TF_DOCS -->

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_bigquery"></a> [bigquery](#module\_bigquery) | terraform-google-modules/bigquery/google | ~> 10.1 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_access"></a> [access](#input\_access) | An array of objects that define dataset access for one or more entities. | `any` | `[]` | no |
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Location of YAML files holding BigQuery dataset configuration values. | `string` | `"../../../configuration/producer/BigQuery/config"` | no |
| <a name="input_dataset_labels"></a> [dataset\_labels](#input\_dataset\_labels) | Key value pairs in a map for dataset labels | `map(string)` | `{}` | no |
| <a name="input_dataset_name"></a> [dataset\_name](#input\_dataset\_name) | Friendly name for the dataset being provisioned. | `string` | `null` | no |
| <a name="input_default_partition_expiration_ms"></a> [default\_partition\_expiration\_ms](#input\_default\_partition\_expiration\_ms) | The default partition expiration for all partitioned tables in the dataset, in MS. | `number` | `null` | no |
| <a name="input_default_table_expiration_ms"></a> [default\_table\_expiration\_ms](#input\_default\_table\_expiration\_ms) | TTL of tables using the dataset in MS. | `number` | `null` | no |
| <a name="input_delete_contents_on_destroy"></a> [delete\_contents\_on\_destroy](#input\_delete\_contents\_on\_destroy) | If set to true, delete all the tables in the dataset when destroying the resource; otherwise, destroying the resource will fail if tables are present. | `bool` | `false` | no |
| <a name="input_deletion_protection"></a> [deletion\_protection](#input\_deletion\_protection) | Whether or not to allow deletion of tables defined by this module. | `bool` | `false` | no |
| <a name="input_description"></a> [description](#input\_description) | Dataset description. | `string` | `"Terraform managed dataset created using the CNCS repository automation."` | no |
| <a name="input_encryption_key"></a> [encryption\_key](#input\_encryption\_key) | Default encryption key to apply to the dataset. Defaults to null (Google-managed). | `string` | `null` | no |
| <a name="input_external_tables"></a> [external\_tables](#input\_external\_tables) | A list of external table objects to create in the dataset. | `any` | `[]` | no |
| <a name="input_location"></a> [location](#input\_location) | The location of the dataset. For multi-region, US or EU can be provided. | `string` | `"US"` | no |
| <a name="input_materialized_views"></a> [materialized\_views](#input\_materialized\_views) | A list of materialized view objects to create in the dataset. | `any` | `[]` | no |
| <a name="input_max_time_travel_hours"></a> [max\_time\_travel\_hours](#input\_max\_time\_travel\_hours) | Defines the time travel window in hours. | `number` | `null` | no |
| <a name="input_resource_tags"></a> [resource\_tags](#input\_resource\_tags) | A map of resource tags to add to the dataset. | `map(string)` | `{}` | no |
| <a name="input_routines"></a> [routines](#input\_routines) | A list of routine objects to create in the dataset. | `any` | `[]` | no |
| <a name="input_storage_billing_model"></a> [storage\_billing\_model](#input\_storage\_billing\_model) | Specifies the storage billing model for the dataset (LOGICAL or PHYSICAL). | `string` | `null` | no |
| <a name="input_tables"></a> [tables](#input\_tables) | A list of table objects to create in the dataset. | `any` | `[]` | no |
| <a name="input_views"></a> [views](#input\_views) | A list of view objects to create in the dataset. | `any` | `[]` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_bigquery_dataset_details"></a> [bigquery\_dataset\_details](#output\_bigquery\_dataset\_details) | A map of the created BigQuery datasets and their nested resources. |

<!-- END_TF_DOCS -->