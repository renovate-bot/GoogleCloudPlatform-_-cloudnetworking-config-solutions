# Terraform Google Cloud Packet Mirroring Module

This module creates and manages a complete Google Cloud Packet Mirroring stack, including the Deployment Group, Endpoint Group, and lists of associated Deployments and Associations. It's designed to be flexible, allowing you to deploy a full mirroring topology or only the specific components you need.

## Usage

The following example creates a complete, high-availability mirroring setup with two zonal deployments and two consumer network associations.

```terraform
module "packet_mirroring_stack" {
  source = "./modules/nsi_out_of_band" # Or your module path

  # --- Deployment Group (Producer VPC container) ---
  create_deployment_group      = true
  deployment_group_project_id  = "deployment-group-project-id" # Must be same as the producer project id 
  deployment_group_name        = "prod-mirroring-dg"
  producer_network_link      = "projects/deployment-project-id/global/networks/producer-vpc"
  deployment_group_labels = {
    env       = "prod"
    component = "networking"
  }

  # --- Endpoint Group (Consumer VPC container) ---
  create_endpoint_group     = true
  endpoint_group_project_id = "endpoint-group-project-id"
  endpoint_group_name       = "prod-mirroring-eg"
  endpoint_group_labels = {
    team    = "security"
    service = "inspection"
  }

  # --- List of Deployments (Collectors) ---
  deployments = [
    {
      name                  = "prod-deployment-us-central1-a"
      deployment_project_id = "deployment-project-id" # Must be same as the producer project id 
      location              = "us-central1-a"
      forwarding_rule_link  = "projects/your-producer-project-id/regions/us-central1/forwardingRules/collector-fr-a"
      description           = "Collector in zone A"
      labels = {
        datacenter   = "us-central1"
        collector-id = "a"
      }
    },
    {
      name                  = "prod-deployment-us-central1-b"
      deployment_project_id = "deployment-project-id" # Must be same as the producer project id 
      location              = "us-central1-b"
      forwarding_rule_link  = "projects/your-producer-project-id/regions/us-central1/forwardingRules/collector-fr-b"
      description           = "Collector in zone B"
      labels = {
        datacenter   = "us-central1"
        collector-id = "b"
      }
    }
  ]

  # --- List of Endpoint Associations (Mirrored Destinations) ---
  endpoint_associations = [
    {
      name                            = "assoc-to-security-vpc"
      endpoint_association_project_id = "endpoint-association-project-id" # Must be same as the consumer project id 
      consumer_network_link           = "projects/your-consumer-project-id/global/networks/security-tools-vpc"
      labels = {
        owner = "security-team"
      }
    },
    {
      name                            = "assoc-to-analytics-vpc"
      endpoint_association_project_id = "endpoint-association-project-id" # Must be same as the consumer project id 
      consumer_network_link           = "projects/your-consumer-project-id/global/networks/analytics-vpc"
      labels = {
        owner = "analytics-team"
      }
    }
  ]
}
```
<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.8 |
| <a name="requirement_google"></a> [google](#requirement\_google) | >= 6.20.0, < 7.0.0 |
| <a name="requirement_google-beta"></a> [google-beta](#requirement\_google-beta) | >= 6.20.0, < 7.0.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_google"></a> [google](#provider\_google) | >= 6.20.0, < 7.0.0 |

## Resources

| Name | Type |
|------|------|
| [google_network_security_mirroring_deployment.deployment](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/network_security_mirroring_deployment) | resource |
| [google_network_security_mirroring_deployment_group.deployment_group](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/network_security_mirroring_deployment_group) | resource |
| [google_network_security_mirroring_endpoint_group.endpoint_group](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/network_security_mirroring_endpoint_group) | resource |
| [google_network_security_mirroring_endpoint_group_association.association](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/network_security_mirroring_endpoint_group_association) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_create_deployment_group"></a> [create\_deployment\_group](#input\_create\_deployment\_group) | Set to true to create the Mirroring Deployment Group. | `bool` | `false` | no |
| <a name="input_create_endpoint_group"></a> [create\_endpoint\_group](#input\_create\_endpoint\_group) | Set to true to create the Mirroring Endpoint Group. | `bool` | `false` | no |
| <a name="input_deployment_group_description"></a> [deployment\_group\_description](#input\_deployment\_group\_description) | A description for the Mirroring Deployment Group. | `string` | `null` | no |
| <a name="input_deployment_group_labels"></a> [deployment\_group\_labels](#input\_deployment\_group\_labels) | A map of labels to add to the Mirroring Deployment Group. | `map(string)` | `{}` | no |
| <a name="input_deployment_group_name"></a> [deployment\_group\_name](#input\_deployment\_group\_name) | The name (ID) for the Mirroring Deployment Group. | `string` | `null` | no |
| <a name="input_deployment_group_project_id"></a> [deployment\_group\_project\_id](#input\_deployment\_group\_project\_id) | Project where deployment group is to be deployed | `string` | `null` | no |
| <a name="input_deployments"></a> [deployments](#input\_deployments) | A list of mirroring deployments to create under the deployment group. | <pre>list(object({<br/>    deployment_project_id = string<br/>    name                 = string<br/>    location             = string<br/>    description          = optional(string)<br/>    labels               = optional(map(string))<br/>    forwarding_rule_link = string<br/>  }))</pre> | `[]` | no |
| <a name="input_endpoint_associations"></a> [endpoint\_associations](#input\_endpoint\_associations) | A list of mirroring endpoint group associations to create under the endpoint group. | <pre>list(object({<br/>    endpoint_association_project_id = string<br/>    name                  = string<br/>    labels                = optional(map(string))<br/>    consumer_network_link = string<br/>  }))</pre> | `[]` | no |
| <a name="input_endpoint_group_description"></a> [endpoint\_group\_description](#input\_endpoint\_group\_description) | A description for the Mirroring Endpoint Group. | `string` | `null` | no |
| <a name="input_endpoint_group_labels"></a> [endpoint\_group\_labels](#input\_endpoint\_group\_labels) | A map of labels to add to the Mirroring Endpoint Group. | `map(string)` | `{}` | no |
| <a name="input_endpoint_group_name"></a> [endpoint\_group\_name](#input\_endpoint\_group\_name) | The name (ID) for the Mirroring Endpoint Group. | `string` | `null` | no |
| <a name="input_endpoint_group_project_id"></a> [endpoint\_group\_project\_id](#input\_endpoint\_group\_project\_id) | Project where endpoint group is to be deployed | `string` | `null` | no |
| <a name="input_existing_deployment_group_id"></a> [existing\_deployment\_group\_id](#input\_existing\_deployment\_group\_id) | The full resource ID of an existing Mirroring Deployment Group to use if create\_deployment\_group is false. | `string` | `null` | no |
| <a name="input_existing_endpoint_group_id"></a> [existing\_endpoint\_group\_id](#input\_existing\_endpoint\_group\_id) | The full resource ID of an existing Mirroring Endpoint Group to use if create\_endpoint\_group is false. | `string` | `null` | no |
| <a name="input_global_location"></a> [global\_location](#input\_global\_location) | The location for global resources like deployment groups and endpoint groups. | `string` | `"global"` | no |
| <a name="input_producer_network_link"></a> [producer\_network\_link](#input\_producer\_network\_link) | The full resource link for the producer VPC network. | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_deployment_group_id"></a> [deployment\_group\_id](#output\_deployment\_group\_id) | The full resource ID of the Mirroring Deployment Group. |
| <a name="output_deployment_group_name"></a> [deployment\_group\_name](#output\_deployment\_group\_name) | The full resource name of the Mirroring Deployment Group. |
| <a name="output_deployments"></a> [deployments](#output\_deployments) | A map of the created mirroring deployments, keyed by their short name. |
| <a name="output_endpoint_associations"></a> [endpoint\_associations](#output\_endpoint\_associations) | A map of the created mirroring endpoint group associations, keyed by their short name. |
| <a name="output_endpoint_group_id"></a> [endpoint\_group\_id](#output\_endpoint\_group\_id) | The full resource ID of the Mirroring Endpoint Group. |
| <a name="output_endpoint_group_name"></a> [endpoint\_group\_name](#output\_endpoint\_group\_name) | The full resource name of the Mirroring Endpoint Group. |
<!-- END_TF_DOCS -->