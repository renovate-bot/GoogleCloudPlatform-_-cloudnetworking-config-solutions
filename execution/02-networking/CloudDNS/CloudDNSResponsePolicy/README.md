# Cloud DNS Response Policy

## Overview

This Terraform configuration provides a modular and YAML-driven approach for deploying and managing Google Cloud DNS Response Policies. It enables you to create and manage response policies, including custom DNS rules and local data overrides, using simple YAML files.

**Key features:**
- **YAML-driven configuration:** Define response policies and rules in YAML files for easy management and reproducibility.
- **Flexible rule definitions:** Support for custom DNS rules and local data records.
- **Modularity:** Easily add or remove response policies by editing YAML files.
- **Automation:** Supports automated deployment of complex DNS response policies.

## Prerequisites

Before creating DNS response policies, ensure you have completed the following prerequisites:

1. **Completed Prior Stages:**
   - **01-organization:** This stage handles the activation of required Google Cloud APIs.

2. **Enable the following APIs:**
    - [Cloud DNS API](https://cloud.google.com/dns/docs/reference/rest): Enables DNS resources.
    - [Compute Engine API](https://cloud.google.com/compute/docs/reference/rest/v1): Used for network resources referenced in policies.

3. **Permissions required for this stage:**
    - [DNS Admin](https://cloud.google.com/iam/docs/understanding-roles#dns.admin): `roles/dns.admin` – Full control over DNS resources.

## Components

- `locals.tf`: Loads and processes YAML configuration files for DNS response policies.
- `responsepolicy.tf`: Instantiates the response policy module for each policy defined in the configuration.
- `variables.tf`: Input variables for customizing the deployment.
- `output.tf`: Exposes module outputs.

## Configuration

To configure DNS response policies for your environment, create YAML files in the `../../../configuration/networking/CloudDNS/CloudDNSResponsePolicy/config/` directory.

**Example YAML:**

```yaml
project_id: <gcp-project-id>
name: my-response-policy
description: Internal overrides for pubsub
create_policy: true

networks:
  default: projects/<gcp-project-id>/global/networks/<your-network-name>

rules:
  override-pubsub:
    dns_name: "pubsub.googleapis.com."
    behavior: "bypass"
    local_data:
      A:
        name: "pubsub.googleapis.com."
        type: "A"
        ttl: 300
        rrdatas:
          - "10.0.0.1"
          - "10.0.0.2"

  bypass-googleapis:
    dns_name: "*.googleapis.com."
    behavior: bypassResponsePolicy
```

## Usage

**NOTE:** Run Terraform commands with the `-var-file` referencing your response policy tfvars file if you override defaults.

```sh
terraform init
terraform plan
terraform apply
```

The module will read all YAML files in the config folder and create the corresponding DNS response policies.

## Example Scenarios

### 1. Create a new response policy with custom rules and local data

```yaml
project_id: <gcp-project-id>
name: my-response-policy
description: Internal overrides for pubsub
create_policy: true

networks:
  default: projects/<gcp-project-id>/global/networks/<your-network-name>

rules:
  override-pubsub:
    dns_name: "pubsub.googleapis.com."
    behavior: "bypass"
    local_data:
      A:
        name: "pubsub.googleapis.com."
        type: "A"
        ttl: 300
        rrdatas:
          - "10.0.0.1"
          - "10.0.0.2"
  bypass-googleapis:
    dns_name: "*.googleapis.com."
    behavior: bypassResponsePolicy
```

### 2. Create a response policy that only bypasses certain domains

```yaml
project_id: <gcp-project-id>
name: bypass-policy
description: Bypass googleapis domains
create_policy: true

networks:
  default: projects/<gcp-project-id>/global/networks/<your-network-name>

rules:
  bypass-googleapis:
    dns_name: "*.googleapis.com."
    behavior: bypassResponsePolicy
```

### 3. Disable creation of a response policy via YAML

```yaml
project_id: <gcp-project-id>
name: disabled-policy
description: This policy will not be created
create_policy: false

networks:
  default: projects/<gcp-project-id>/global/networks/<your-network-name>

rules: {}
```

## Outputs

- `id`: Fully qualified policy id.
- `name`: Policy name.
- `policy`: Policy resource.

## Notes

- Ensure all required APIs are enabled and permissions are granted.
- Adjust YAML fields as per your environment and naming conventions.
- For advanced configurations, refer to the [Google Cloud DNS documentation](https://cloud.google.com/dns/docs/response-policies).

<!-- BEGIN_TF_DOCS -->

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_response_policy"></a> [response\_policy](#module\_response\_policy) | git::https://github.com/GoogleCloudPlatform/cloud-foundation-fabric.git//modules/dns-response-policy | v41.0.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Path to YAML config files defining response policies. | `string` | `"../../../../configuration/networking/CloudDNS/CloudDNSResponsePolicy/config"` | no |
| <a name="input_default_create_policy"></a> [default\_create\_policy](#input\_default\_create\_policy) | Whether to create the response policy if not specified in YAML | `bool` | `true` | no |
| <a name="input_default_description"></a> [default\_description](#input\_default\_description) | n/a | `string` | `"Managed by Terraform"` | no |
| <a name="input_default_networks"></a> [default\_networks](#input\_default\_networks) | Map of networks to attach the response policy to. | `map(string)` | `{}` | no |
| <a name="input_default_rules"></a> [default\_rules](#input\_default\_rules) | Map of response policy rules | <pre>map(object({<br/>    dns_name   = string<br/>    behavior   = optional(string)<br/>    local_data = optional(map(object({<br/>      name    = string<br/>      type    = string<br/>      ttl     = number<br/>      rrdatas = list(string)<br/>    })))<br/>  }))</pre> | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_id"></a> [id](#output\_id) | Fully qualified policy id. |
| <a name="output_name"></a> [name](#output\_name) | Policy name. |
| <a name="output_policy"></a> [policy](#output\_policy) | Policy resource. |
<!-- END_TF_DOCS -->