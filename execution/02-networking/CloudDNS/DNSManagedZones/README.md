# Cloud DNS Managed Zones

## Overview

This Terraform configuration provides a modular and YAML-driven approach for deploying and managing Google Cloud DNS Managed Zones. It enables you to create and manage public, private, peering, forwarding, and reverse DNS zones using simple YAML files.

**Key features:**
- **YAML-driven configuration:** Define DNS zones and records in YAML files for easy management and reproducibility.
- **Supports all zone types:** Public, private, peering, forwarding, and reverse zones.
- **Flexible recordsets:** Easily manage DNS records for each zone.
- **Modularity:** Add or remove zones by editing YAML files.
- **Automation:** Supports automated deployment of complex DNS architectures.

## Prerequisites

Before creating DNS managed zones, ensure you have completed the following prerequisites:

1. **Completed Prior Stages:**
   - **01-organization:** This stage handles the activation of required Google Cloud APIs.

2. **Enable the following APIs:**
    - [Cloud DNS API](https://cloud.google.com/dns/docs/reference/rest): Enables DNS resources.
    - [Compute Engine API](https://cloud.google.com/compute/docs/reference/rest/v1): Used for network resources referenced in zones.

3. **Permissions required for this stage:**
    - [DNS Admin](https://cloud.google.com/iam/docs/understanding-roles#dns.admin): `roles/dns.admin` – Full control over DNS resources.

## Components

- `locals.tf`: Loads and processes YAML configuration files for DNS managed zones.
- `dns.tf`: Instantiates the managed zone module for each zone defined in the configuration.
- `variables.tf`: Input variables for customizing the deployment.
- `output.tf`: Exposes module outputs.

## Configuration

To configure DNS managed zones for your environment, create YAML files in the `../../../configuration/networking/CloudDNS/DNSManagedZones/config/` directory.

---

## Example YAMLs

### 1. **DNS Managed Public-Zone Creation**

```yaml
name: my-public-zone
project_id: <gcp-project-id>
description: "A description for your DNS managed zone"
force_destroy: false
zone_config:
  domain: example.com.
  visibility: public
  reverse_lookup: false
recordsets:
  "A www.example.com.":
    ttl: 300
    records:
      - "8.8.8.8"
```

---

### 2. **DNS Managed Private-Zone Creation**

```yaml
name: corp-internal-zone
project_id: <gcp-project-id>
description: "Private zone for corporate internal services"
force_destroy: false
zone_config:
  domain: corp.internal.
  visibility: private
  reverse_lookup: false
  private_visibility_config:
    networks:
      - network_url: "projects/<gcp-project-id>/global/networks/<your-network-name>"
recordsets:
  "A db.corp.internal.":
    ttl: 300
    records:
      - "10.10.1.5"
```

---

### 3. **DNS Managed Private-Reverse-Zone Creation**

```yaml
name: corp-internal-10-10-reverse-zone
project_id: <gcp-project-id>
description: "Private reverse lookup for the 10.10.0.0/16 network"
force_destroy: false
zone_config:
  domain: 10.10.in-addr.arpa.
  visibility: private
  reverse_lookup: true
  private_visibility_config:
    networks:
      - network_url: "projects/<gcp-project-id>/global/networks/<your-network-name>"
recordsets:
  "PTR 5.1.10.10.in-addr.arpa.":
    ttl: 300
    records:
      - "db.corp.internal."
```

---

### 4. **DNS Managed Peering-Zone Creation**

```yaml
name: services-peering-zone
project_id: <consumer-project-id>
description: "Peers with the producer network for service resolution"
force_destroy: false
zone_config:
  domain: shared-services.internal.
  visibility: private
  reverse_lookup: false
  private_visibility_config:
    networks:
      - network_url: "projects/<consumer-project-id>/global/networks/<consumer-network-name>"
  peering_config:
    target_network:
      network_url: "projects/<producer-project-id>/global/networks/<producer-network-name>"
```

---

### 5. **DNS Managed Forwarding-Zone Creation**

```yaml
name: onprem-forwarding-zone
project_id: <gcp-project-id>
description: "Forwards requests for onprem.corp to on-premises DNS"
force_destroy: false
zone_config:
  domain: onprem.corp.
  visibility: private
  reverse_lookup: false
  private_visibility_config:
    networks:
      - network_url: "projects/<gcp-project-id>/global/networks/<your-network-name>"
  forwarding_config:
    target_name_servers:
      - ipv4_address: "192.168.1.10"
      - ipv4_address: "192.168.1.11"
```

---

## Usage

```
 
**NOTE:** Run Terraform commands with the `-var-file` referencing your managed zones tfvars file if you override defaults.

```sh

terraform init
terraform plan
terraform apply
```

The module will read all YAML files in the config folder and create the corresponding DNS managed zones.

## Outputs

- `id`: Fully qualified zone id.
- `name`: Zone name.
- `zone`: Zone resource.

## Notes

- Ensure all required APIs are enabled and permissions are granted.
- Adjust YAML fields as per your environment and naming conventions.
- For advanced configurations, refer to the [Google Cloud DNS documentation](https://cloud.google.com/dns/docs/zones)


<!-- BEGIN_TF_DOCS -->
## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_dns"></a> [dns](#module\_dns) | git::https://github.com/GoogleCloudPlatform/cloud-foundation-fabric.git//modules/dns | v31.0.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Location of YAML files holding Cloud DNS configuration values. | `string` | `"../../../../configuration/networking/CloudDNS/DNSManagedZones/config"` | no |
| <a name="input_default_description"></a> [default\_description](#input\_default\_description) | Default description for DNS zones. | `string` | `"Terraform managed DNS Zones"` | no |
| <a name="input_default_force_destroy"></a> [default\_force\_destroy](#input\_default\_force\_destroy) | Default force\_destroy value for DNS zones. | `bool` | `false` | no |
| <a name="input_default_iam_bindings"></a> [default\_iam\_bindings](#input\_default\_iam\_bindings) | Default IAM bindings to apply to DNS zones if not specified in YAML. | `map(list(string))` | `{}` | no |
| <a name="input_default_recordsets"></a> [default\_recordsets](#input\_default\_recordsets) | Default recordsets for DNS zones. | `map(any)` | `{}` | no |
| <a name="input_default_zone_config"></a> [default\_zone\_config](#input\_default\_zone\_config) | Default zone\_config for DNS zones. | `any` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_name_servers"></a> [name\_servers](#output\_name\_servers) | A map of the name servers for each created DNS zone. |
| <a name="output_zone_names"></a> [zone\_names](#output\_zone\_names) | A map of the names of the created DNS zones. |
| <a name="output_zones"></a> [zones](#output\_zones) | A map of the created DNS zone resources. |
<!-- END_TF_DOCS -->