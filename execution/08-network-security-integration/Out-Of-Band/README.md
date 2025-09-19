## Overview

This Terraform configuration simplifies the process of creating and managing Google Cloud Packet Mirroring policies. Using a modular, YAML-driven approach, it allows you to declaratively define your packet mirroring topology for network traffic analysis, security monitoring, and intrusion detection at scale.

## Key Features

-   **YAML-Driven Automation:** Effortlessly define and deploy packet mirroring resources using simple, readable YAML files. Manage your entire mirroring configuration as data.
-   **Flexible Resource Creation:** Create a full high-availability (HA) mirroring stack, or individual components, all from a single set of configurations.
-   **Multi-Deployment & Association Support:** A single configuration can create multiple zonal deployments for HA and multiple associations to mirror traffic to different consumer VPCs.
-   **Centralized Management:** Manage packet mirroring configurations for multiple environments or regions from a single source-controlled repository.
-   **Integration with Custom Module:** Leverages the custom Terraform Packet Mirroring module we built for reliable and consistent resource deployment.

## Prerequisites

Before using this configuration, ensure the following prerequisites are met:

1.  **Google Cloud Project:** You must have a Google Cloud project with the necessary APIs enabled (Compute Engine, Network Security).
2.  **Collector Infrastructure:** The packet mirroring collector infrastructure (VMs, Instance Groups, Health Check, Backend Service, and **Forwarding Rules**) must be deployed beforehand. This configuration only manages the mirroring policies themselves.
3.  **Terraform Installed:** Install Terraform (v1.3.0 or later) on your local machine or CI/CD environment.
4.  **Google Cloud SDK:** Install and authenticate the Google Cloud SDK (`gcloud`).
5.  **IAM Permissions:** Ensure the principal (user, service account) running Terraform has the **Packet Mirroring Admin** (`roles/networksecurity.packetMirroringAdmin`) role on the project.
6.  **Terraform Packet Mirroring Module:** The custom module must be available at the path specified in `nsioutofband.tf` (e.g., `../../../modules/nsi_out_of_band/`).

## Description

-   **`Mirroring Deployment Group`**: A top-level container for mirroring resources within a single **producer VPC**.
-   **`Mirroring Deployment`**: A resource that links a Deployment Group to a specific collector (represented by a unique forwarding rule) in a specific zone.
-   **`Mirroring Endpoint Group`**: A container that groups one or more consumer VPCs that will receive the mirrored traffic.
-   **`Mirroring Endpoint Group Association`**: A resource that links a specific **consumer VPC** to an Endpoint Group.
-   **YAML Configuration:** This setup works by reading all `.yaml` or `.yml` files from a specified directory (`config/`). Each YAML file declaratively defines a set of mirroring resources to be created.
-   **Key YAML Blocks:**
    -   `deployment_group`: Contains the data to create the `google_network_security_mirroring_deployment_group`.
    -   `endpoint_group`: Contains the data for the `google_network_security_mirroring_endpoint_group`.
    -   `deployments`: A **list** of `google_network_security_mirroring_deployment` resources to create.
    -   `associations`: A **list** of `google_network_security_mirroring_endpoint_group_association` resources to create.

## Example YAML Configuration (`ha_multi_consumer.yml`)

The following example shows how to create a complete stack: one deployment group in a producer VPC, two deployments for HA, one endpoint group, and two associations to mirror traffic to two different consumer VPCs.

Place this file inside your configuration directory.

```yaml
# config/nsi_out_of_band_readme_sample.yaml

deployment_group:
  create: true #e.g. true or false
  deployment_group_project_id: <your-gcp-project-id> #e.g. "dgp-test-project"
  name: <deployment-group-name> #e.g. "prod-web-app-dg"
  description: <deployment-group-description> #e.g. "Deployment group for the production web application VPC"
  producer_network_link: <deployment-group-producer-network-link> #e.g. "projects/acme-corp-net-prod-123456/global/networks/prod-app-vpc"

endpoint_group:
  create: true #e.g. true or false
  endpoint_group_project_id: <your-gcp-project-id> #e.g. "egp-test-project"
  name: <endpoint-group-name> #e.g. "prod-inspection-tools-eg"
  description: <endpoint-group-description> #e.g. "Endpoint group for all security inspection tools"

# List of deployments to create under the single deployment_group
deployments:
  - deployment_project_id: <your-gcp-project-id> #e.g. "dp-test-project1"
    name: <deployment-name> #e.g. "prod-web-app-dep-a"
    location: <deployment-zone> #e.g. "us-central1-a"
    forwarding_rule_link: <IPNLB-forwarding-link> #Internal Passthrough Network Load Balancer (IPNLB) forwarding rule link e.g. "projects/acme-corp-net-prod-123456/regions/us-central1/forwardingRules/pkt-mirror-collector-a"
    description: <deployment-description> #e.g. "Collector for zone A"

  - deployment_project_id: <your-gcp-project-id> #e.g. "dp-test-project2"
    name: <deployment-name> #e.g. "prod-web-app-dep-b"
    location: <deployment-zone> #e.g. "us-central1-b"
    forwarding_rule_link: <IPNLB-forwarding-link> #e.g. "projects/acme-corp-net-prod-123456/regions/us-central1/forwardingRules/pkt-mirror-collector-b"
    description: <deployment-description> #e.g. "Collector for zone B"

# List of associations to create under the single endpoint_group
associations:
  - endpoint_association_project_id: <your-gcp-project-id> #e.g. "ea-test-project1"
    name: <associations-name> #e.g. "assoc-to-security-vpc"
    consumer_network_link: <consumer-network-link> #e.g. "projects/acme-corp-net-prod-123456/global/networks/security-tools-vpc"

  - endpoint_association_project_id: <your-gcp-project-id> #e.g. "ea-test-project2"
    name: <associations-name> #e.g. "assoc-to-analytics-vpc"
    consumer_network_link: <consumer-network-link> #e.g. "projects/acme-corp-net-prod-123456/global/networks/analytics-vpc"
```
<!-- BEGIN_TF_DOCS -->

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_packet_mirroring"></a> [packet\_mirroring](#module\_packet\_mirroring) | ../../../modules/nsi_out_of_band/ | n/a |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_association_labels"></a> [association\_labels](#input\_association\_labels) | Default labels for the Mirroring Endpoint Group Association. | `map(string)` | `{}` | no |
| <a name="input_association_name"></a> [association\_name](#input\_association\_name) | Default name for the Mirroring Endpoint Group Association. | `string` | `null` | no |
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Path to the folder containing the YAML configuration files. | `string` | `"../../../configuration/network-security-integration/OutOfBand/config/"` | no |
| <a name="input_consumer_network_link"></a> [consumer\_network\_link](#input\_consumer\_network\_link) | Default full resource link for the consumer VPC network. | `string` | `null` | no |
| <a name="input_create_association"></a> [create\_association](#input\_create\_association) | Controls creation of the Mirroring Endpoint Group Association. | `bool` | `false` | no |
| <a name="input_create_deployment"></a> [create\_deployment](#input\_create\_deployment) | Controls creation of the Mirroring Deployment. | `bool` | `false` | no |
| <a name="input_create_deployment_group"></a> [create\_deployment\_group](#input\_create\_deployment\_group) | Controls creation of the Mirroring Deployment Group. | `bool` | `false` | no |
| <a name="input_create_endpoint_group"></a> [create\_endpoint\_group](#input\_create\_endpoint\_group) | Controls creation of the Mirroring Endpoint Group. | `bool` | `false` | no |
| <a name="input_deployment_description"></a> [deployment\_description](#input\_deployment\_description) | Default description for the Mirroring Deployment. | `string` | `null` | no |
| <a name="input_deployment_group"></a> [deployment\_group](#input\_deployment\_group) | Provides a default empty map {} for the deployment\_group object. | `map` | `{}` | no |
| <a name="input_deployment_group_description"></a> [deployment\_group\_description](#input\_deployment\_group\_description) | Default description for the Mirroring Deployment Group. | `string` | `null` | no |
| <a name="input_deployment_group_labels"></a> [deployment\_group\_labels](#input\_deployment\_group\_labels) | Default labels for the Mirroring Deployment Group. | `map(string)` | `{}` | no |
| <a name="input_deployment_group_name"></a> [deployment\_group\_name](#input\_deployment\_group\_name) | Default name for the Mirroring Deployment Group. | `string` | `null` | no |
| <a name="input_deployment_group_project_id"></a> [deployment\_group\_project\_id](#input\_deployment\_group\_project\_id) | Project where deployment group is to be deployed | `string` | `null` | no |
| <a name="input_deployment_labels"></a> [deployment\_labels](#input\_deployment\_labels) | Default labels for the Mirroring Deployment. | `map(string)` | `{}` | no |
| <a name="input_deployment_name"></a> [deployment\_name](#input\_deployment\_name) | Default name for the Mirroring Deployment. | `string` | `null` | no |
| <a name="input_deployments"></a> [deployments](#input\_deployments) | Provides a default empty list [] for the deployments list. | `list` | `[]` | no |
| <a name="input_endpoint_associations"></a> [endpoint\_associations](#input\_endpoint\_associations) | Provides a default empty list [] for the associations list. | `list` | `[]` | no |
| <a name="input_endpoint_group"></a> [endpoint\_group](#input\_endpoint\_group) | Provides a default empty map {} for the endpoint\_group object. | `map` | `{}` | no |
| <a name="input_endpoint_group_description"></a> [endpoint\_group\_description](#input\_endpoint\_group\_description) | Default description for the Mirroring Endpoint Group. | `string` | `null` | no |
| <a name="input_endpoint_group_labels"></a> [endpoint\_group\_labels](#input\_endpoint\_group\_labels) | Default labels for the Mirroring Endpoint Group. | `map(string)` | `{}` | no |
| <a name="input_endpoint_group_name"></a> [endpoint\_group\_name](#input\_endpoint\_group\_name) | Default name for the Mirroring Endpoint Group. | `string` | `null` | no |
| <a name="input_endpoint_group_project_id"></a> [endpoint\_group\_project\_id](#input\_endpoint\_group\_project\_id) | Project where endpoint group is to be deployed | `string` | `null` | no |
| <a name="input_existing_deployment_group_id"></a> [existing\_deployment\_group\_id](#input\_existing\_deployment\_group\_id) | Default existing Mirroring Deployment Group ID to use. | `string` | `null` | no |
| <a name="input_existing_endpoint_group_id"></a> [existing\_endpoint\_group\_id](#input\_existing\_endpoint\_group\_id) | Default existing Mirroring Endpoint Group ID to use. | `string` | `null` | no |
| <a name="input_forwarding_rule_link"></a> [forwarding\_rule\_link](#input\_forwarding\_rule\_link) | Default full resource link for the forwarding rule. | `string` | `null` | no |
| <a name="input_global_location"></a> [global\_location](#input\_global\_location) | Default location for global resources. | `string` | `"global"` | no |
| <a name="input_location"></a> [location](#input\_location) | Default location for the mirroring deployment (e.g., 'us-central1-a'). | `string` | `null` | no |
| <a name="input_producer_network_link"></a> [producer\_network\_link](#input\_producer\_network\_link) | Default full resource link for the producer VPC network. | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_all_deployment_groups"></a> [all\_deployment\_groups](#output\_all\_deployment\_groups) | A map of all created Mirroring Deployment Groups, keyed by the config file name. |
| <a name="output_all_deployments"></a> [all\_deployments](#output\_all\_deployments) | A combined map of all created deployments from all config files, keyed by their short name. |
| <a name="output_all_endpoint_associations"></a> [all\_endpoint\_associations](#output\_all\_endpoint\_associations) | A combined map of all created endpoint associations from all config files, keyed by their short name. |
| <a name="output_all_endpoint_groups"></a> [all\_endpoint\_groups](#output\_all\_endpoint\_groups) | A map of all created Mirroring Endpoint Groups, keyed by the config file name. |
<!-- END_TF_DOCS -->