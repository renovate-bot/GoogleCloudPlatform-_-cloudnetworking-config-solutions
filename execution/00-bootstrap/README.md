## Introduction

The bootstrap stage is the first and most crucial step in setting up your Google Cloud infrastructure using Terraform. It lays the groundwork for subsequent stages (01-organization, 02-networking, 03-security, 04-producer, 05-producer-connectivity, 06-consumer) by provisioning essential resources and establishing security best practices. This stage focuses on creating the following:
  * **Impersonating Service Accounts:** This stage generates service accounts for each subsequent stage, allowing them to impersonate roles with the necessary permissions for their respective tasks. This approach enhances security by granting only the required privileges to each stage.
  * **Terraform State Bucket:** A Google Cloud Storage bucket is created to store the Terraform state files. This centralizes state management, making it easier to track changes and collaborate on infrastructure updates.

## Pre-Requisites

* IAM Permissions: The user or service account executing Terraform must have the following IAM roles or equivalent permissions:
  * `roles/iam.serviceAccountAdmin` : To create and manage service accounts for the project for which service account needs to be created.
  * `roles/resourcemanager.projectIamAdmin` : Provides permissions to administer allow policies on projects.
  * `roles/storage.admin` : To create and manage Google Cloud Storage buckets.

## Execution Steps:

1. Create `terraform.tfvars`:
    * Make a copy of the provided terraform.tfvars.example file and rename it to terraform.tfvars.
    * Fill in the values for [input variable](#inputs) and other required variables according to your specific requirements.
2. Initialize Terraform:
    `terraform init`
3. Review Execution Plan:
    `terraform plan -var-file="../../configuration/bootstrap.tfvars"`
4. Apply Configuration:
    `terraform apply -var-file="../../configuration/bootstrap.tfvars"`

### Example

To help you get started, we've provided examples of tfvars files that you can use :

* **Minimal tfvars (Mandatory Fields Only):**

This minimal example includes only the essential fields required to execute the bootstrap stage.

  
  ```
  folder_id                             = ""
  bootstrap_project_id                  = ""
  network_hostproject_id                = ""
  network_serviceproject_id             = "" // <service(producer/consumer)-project-id>

  organization_administrator          = ["user:organization-user-example@example.com"]
  networking_administrator            = ["user:networking-user-example@example.com"]
  security_administrator              = ["user:security-user-example@example.com"]

  producer_cloudsql_administrator     = ["user:cloudsql-user-example@example.com"]
  producer_gke_administrator          = ["user:gke-user-example@example.com"]
  producer_alloydb_administrator      = ["user:alloydb-user-example@example.com"]
  producer_vertex_administrator       = ["user:vertex-user-example@example.com"]
  producer_mrc_administrator          = ["user:mrc-user-example@example.com"]
  producer_bigquery_administrator     = ["user:bigquery-user-example@example.com"]
  producer_connectivity_administrator = ["user:connectivity-user-example@example.com"]

  consumer_gce_administrator           = ["user:gce-user-example@example.com"]
  consumer_cloudrun_administrator      = ["user:cloudrun-user-example@example.com"]
  consumer_mig_administrator           = ["user:mig-user-example@example.com"]
  consumer_umig_administrator          = ["user:umig-user-example@example.com"]  
  consumer_workbench_administrator     = ["user:workbench-user-example@example.com"]
  consumer_vpc_connector_administrator = ["user:user-example@example.com"]
  consumer_appengine_administrator     = ["user:user-example@example.com"]
  consumer_lb_administrator            = ["user:lb-user-example@example.com"]
  ```


## Important Considerations:

  * **Security**: Pay close attention to the permissions granted to the service accounts. Follow the principle of least privilege to minimize security risks.
  * **State Management:** The Terraform state bucket is critical for maintaining the state of your infrastructure. Ensure its security and accessibility.
  * **Dependencies:** This bootstrap stage is a prerequisite for all subsequent stages. Make sure it is executed successfully before proceeding with other stages.
  **Note:** You can skip the bootstrap stage if you choose, but you must ensure the following:
  * **Permissions:** The user or service account executing Terraform for each individual stage (01-organization, 02-networking, etc.) must have the necessary IAM permissions outlined in the respective stage's README file.
 * **State File Management:** You are responsible for setting up and maintaining a secure location for Terraform state files for each stage. This could involve using a Google Cloud Storage bucket, a local backend, or another suitable storage mechanism.
 **Distinct user IDs** : we strongly discourage using the same user ID for all stages and highly recommend users to follow the principle of least privilege for separate service accounts for the stages.

<!-- BEGIN_TF_DOCS -->
 ## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_alloydb_producer"></a> [alloydb\_producer](#module\_alloydb\_producer) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_appengine_consumer"></a> [appengine\_consumer](#module\_appengine\_consumer) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_cloudrun_consumer"></a> [cloudrun\_consumer](#module\_cloudrun\_consumer) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_cloudsql_producer"></a> [cloudsql\_producer](#module\_cloudsql\_producer) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_consumer_load_balancing"></a> [consumer\_load\_balancing](#module\_consumer\_load\_balancing) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_consumer_vpc_access_connector"></a> [consumer\_vpc\_access\_connector](#module\_consumer\_vpc\_access\_connector) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_gce_consumer"></a> [gce\_consumer](#module\_gce\_consumer) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_gke_producer"></a> [gke\_producer](#module\_gke\_producer) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_google_storage_bucket"></a> [google\_storage\_bucket](#module\_google\_storage\_bucket) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/gcs | v31.1.0 |
| <a name="module_mig_consumer"></a> [mig\_consumer](#module\_mig\_consumer) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_mrc_producer"></a> [mrc\_producer](#module\_mrc\_producer) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_networking"></a> [networking](#module\_networking) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_organization"></a> [organization](#module\_organization) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v34.1.0 |
| <a name="module_producer_connectivity"></a> [producer\_connectivity](#module\_producer\_connectivity) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_security"></a> [security](#module\_security) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_umig_consumer"></a> [umig\_consumer](#module\_umig\_consumer) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_vertex_producer"></a> [vertex\_producer](#module\_vertex\_producer) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |
| <a name="module_workbench_consumer"></a> [workbench\_consumer](#module\_workbench\_consumer) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/iam-service-account | v31.1.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_bootstrap_project_id"></a> [bootstrap\_project\_id](#input\_bootstrap\_project\_id) | Google Cloud Project ID which will be used to create the service account and Google Cloud storage buckets. | `string` | n/a | yes |
| <a name="input_consumer_appengine_administrator"></a> [consumer\_appengine\_administrator](#input\_consumer\_appengine\_administrator) | List of App Engine administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_consumer_appengine_sa_name"></a> [consumer\_appengine\_sa\_name](#input\_consumer\_appengine\_sa\_name) | Name of the service account to create for App Engine consumer stage. | `string` | `"consumer-appengine-sa"` | no |
| <a name="input_consumer_cloudrun_administrator"></a> [consumer\_cloudrun\_administrator](#input\_consumer\_cloudrun\_administrator) | List of Cloud Run administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_consumer_cloudrun_sa_name"></a> [consumer\_cloudrun\_sa\_name](#input\_consumer\_cloudrun\_sa\_name) | Name of the service account to create for Cloud Run consumer stage. | `string` | `"consumer-cloudrun-sa"` | no |
| <a name="input_consumer_gce_administrator"></a> [consumer\_gce\_administrator](#input\_consumer\_gce\_administrator) | List of GCE administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_consumer_gce_sa_name"></a> [consumer\_gce\_sa\_name](#input\_consumer\_gce\_sa\_name) | Name of the service account to create for GCE consumer stage. | `string` | `"consumer-gce-sa"` | no |
| <a name="input_consumer_lb_administrator"></a> [consumer\_lb\_administrator](#input\_consumer\_lb\_administrator) | List of LB administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_consumer_lb_sa_name"></a> [consumer\_lb\_sa\_name](#input\_consumer\_lb\_sa\_name) | Name of the service account to create for LB consumer stage. | `string` | `"consumer-lb-sa"` | no |
| <a name="input_consumer_mig_administrator"></a> [consumer\_mig\_administrator](#input\_consumer\_mig\_administrator) | List of MIG administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_consumer_mig_sa_name"></a> [consumer\_mig\_sa\_name](#input\_consumer\_mig\_sa\_name) | Name of the service account to create for MIG consumer stage. | `string` | `"consumer-mig-sa"` | no |
| <a name="input_consumer_vpc_connector_administrator"></a> [consumer\_vpc\_connector\_administrator](#input\_consumer\_vpc\_connector\_administrator) | List of VPC conector administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_consumer_vpc_connector_sa_name"></a> [consumer\_vpc\_connector\_sa\_name](#input\_consumer\_vpc\_connector\_sa\_name) | Name of the service account to create for VPC Access Connector consumer stage. | `string` | `"consumer-vpcaccess-sa"` | no |
| <a name="input_consumer_workbench_administrator"></a> [consumer\_workbench\_administrator](#input\_consumer\_workbench\_administrator) | List of Workbench administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_consumer_workbench_sa_name"></a> [consumer\_workbench\_sa\_name](#input\_consumer\_workbench\_sa\_name) | Name of the service account to create for Workbench consumer stage. | `string` | `"consumer-workbench-sa"` | no |
| <a name="input_folder_id"></a> [folder\_id](#input\_folder\_id) | Google Cloud folder ID designating the parent folder for both the networking host project and the service project. | `string` | n/a | yes |
| <a name="input_gcs_bucket_location"></a> [gcs\_bucket\_location](#input\_gcs\_bucket\_location) | Location of the Google Cloud storage bucket. | `string` | `"EU"` | no |
| <a name="input_gcs_bucket_name"></a> [gcs\_bucket\_name](#input\_gcs\_bucket\_name) | Name of the Google Cloud storage bucket. | `string` | `"terraform-state"` | no |
| <a name="input_network_hostproject_id"></a> [network\_hostproject\_id](#input\_network\_hostproject\_id) | Google Cloud Project ID for the networking host project to be used to create networking and security resources. | `string` | n/a | yes |
| <a name="input_network_serviceproject_id"></a> [network\_serviceproject\_id](#input\_network\_serviceproject\_id) | Google Cloud Project ID to be used to create Google Cloud resources like consumer and producer services. | `string` | n/a | yes |
| <a name="input_networking_administrator"></a> [networking\_administrator](#input\_networking\_administrator) | List of Members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_networking_sa_name"></a> [networking\_sa\_name](#input\_networking\_sa\_name) | Name of the service account to create for networking stage. | `string` | `"networking-sa"` | no |
| <a name="input_organization_administrator"></a> [organization\_administrator](#input\_organization\_administrator) | List of Members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_organization_sa_name"></a> [organization\_sa\_name](#input\_organization\_sa\_name) | Name of the service account to create for organization stage. | `string` | `"organization-sa"` | no |
| <a name="input_producer_alloydb_administrator"></a> [producer\_alloydb\_administrator](#input\_producer\_alloydb\_administrator) | List of AlloyDB administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_producer_alloydb_sa_name"></a> [producer\_alloydb\_sa\_name](#input\_producer\_alloydb\_sa\_name) | Name of the service account to create for AlloyDB's producer stage. | `string` | `"producer-alloydb-sa"` | no |
| <a name="input_producer_cloudsql_administrator"></a> [producer\_cloudsql\_administrator](#input\_producer\_cloudsql\_administrator) | List of Cloud SQL administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_producer_cloudsql_sa_name"></a> [producer\_cloudsql\_sa\_name](#input\_producer\_cloudsql\_sa\_name) | Name of the service account to create for CloudSQL's producer stage. | `string` | `"producer-cloudsql-sa"` | no |
| <a name="input_producer_connectivity_administrator"></a> [producer\_connectivity\_administrator](#input\_producer\_connectivity\_administrator) | List of Members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_producer_connectivity_sa_name"></a> [producer\_connectivity\_sa\_name](#input\_producer\_connectivity\_sa\_name) | Name of the service account to create for producer connectivity stage. | `string` | `"producer_connectivity_sa"` | no |
| <a name="input_producer_gke_administrator"></a> [producer\_gke\_administrator](#input\_producer\_gke\_administrator) | List of GKE administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_producer_gke_sa_name"></a> [producer\_gke\_sa\_name](#input\_producer\_gke\_sa\_name) | Name of the service account to create for GKE's producer stage. | `string` | `"producer-gke-sa"` | no |
| <a name="input_producer_mrc_administrator"></a> [producer\_mrc\_administrator](#input\_producer\_mrc\_administrator) | List of MRC administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_producer_mrc_sa_name"></a> [producer\_mrc\_sa\_name](#input\_producer\_mrc\_sa\_name) | Name of the service account to create for MRC's producer stage. | `string` | `"producer-mrc-sa"` | no |
| <a name="input_producer_vertex_administrator"></a> [producer\_vertex\_administrator](#input\_producer\_vertex\_administrator) | List of Vertex AI administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_producer_vertex_sa_name"></a> [producer\_vertex\_sa\_name](#input\_producer\_vertex\_sa\_name) | Name of the service account to create for Vertex AI's producer stage. | `string` | `"producer-vertex-sa"` | no |
| <a name="input_security_administrator"></a> [security\_administrator](#input\_security\_administrator) | List of Members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com) | `list(string)` | <pre>[<br/>  ""<br/>]</pre> | no |
| <a name="input_security_sa_name"></a> [security\_sa\_name](#input\_security\_sa\_name) | Name of the service account to create for security stage. | `string` | `"security-sa"` | no |
| <a name="input_versioning"></a> [versioning](#input\_versioning) | The Goocle Cloud storage bucket versioning. | `bool` | `true` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_consumer_appengine_email"></a> [consumer\_appengine\_email](#output\_consumer\_appengine\_email) | App engine consumer stage service account IAM email. |
| <a name="output_consumer_cloudrun_email"></a> [consumer\_cloudrun\_email](#output\_consumer\_cloudrun\_email) | Cloud Run consumer stage service account IAM email. |
| <a name="output_consumer_gce_email"></a> [consumer\_gce\_email](#output\_consumer\_gce\_email) | GCE consumer stage service account IAM email. |
| <a name="output_consumer_lb_email"></a> [consumer\_lb\_email](#output\_consumer\_lb\_email) | Consumer Load Balancing stage service account IAM email. |
| <a name="output_consumer_mig_email"></a> [consumer\_mig\_email](#output\_consumer\_mig\_email) | MIG consumer stage service account IAM email. |
| <a name="output_consumer_umig_email"></a> [consumer\_umig\_email](#output\_consumer\_umig\_email) | UMIG consumer stage service account IAM email. |
| <a name="output_consumer_vpc_access_connector_email"></a> [consumer\_vpc\_access\_connector\_email](#output\_consumer\_vpc\_access\_connector\_email) | VPC Access Connector consumer stage service account IAM email. |
| <a name="output_consumer_workbench_email"></a> [consumer\_workbench\_email](#output\_consumer\_workbench\_email) | Workbench consumer stage service account IAM email. |
| <a name="output_networking_email"></a> [networking\_email](#output\_networking\_email) | Networking stage service account IAM email. |
| <a name="output_organization_email"></a> [organization\_email](#output\_organization\_email) | Organization stage service account IAM email. |
| <a name="output_producer_alloydb_email"></a> [producer\_alloydb\_email](#output\_producer\_alloydb\_email) | AlloyDB producer stage service account IAM email. |
| <a name="output_producer_cloudsql_email"></a> [producer\_cloudsql\_email](#output\_producer\_cloudsql\_email) | CloudSQL producer stage service account IAM email. |
| <a name="output_producer_connectivity_email"></a> [producer\_connectivity\_email](#output\_producer\_connectivity\_email) | Producer Connectivity stage service account IAM email. |
| <a name="output_producer_gke_email"></a> [producer\_gke\_email](#output\_producer\_gke\_email) | GKE producer stage service account IAM email. |
| <a name="output_producer_mrc_email"></a> [producer\_mrc\_email](#output\_producer\_mrc\_email) | MRC producer stage service account IAM email. |
| <a name="output_producer_vertex_email"></a> [producer\_vertex\_email](#output\_producer\_vertex\_email) | Vertex producer stage service account IAM email. |
| <a name="output_security_email"></a> [security\_email](#output\_security\_email) | Security stage service account IAM email. |
| <a name="output_storage_bucket_name"></a> [storage\_bucket\_name](#output\_storage\_bucket\_name) | Google Cloud storage bucket name. |
<!-- END_TF_DOCS -->
