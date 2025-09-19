## Overview

This Terraform configuration simplifies the process of creating and managing Google Cloud Security Profiles and Security Profile Groups. Using a modular, data-driven approach, it reads declarative YAML files to deploy and manage your network security posture at scale. Properly configured security profiles are a critical component of GCP's intrusion prevention service.

## Key Features

- **YAML-Driven Automation:** Effortlessly define and deploy security profiles and groups using simple, readable YAML files. Manage all your configurations as data.
- **Flexible Resource Creation:** Create a security profile, a security profile group, or both simultaneously and link them together, all from a single YAML file.
- **Multi-Profile Support:** Supports all security profile types, including `THREAT_PREVENTION`, `CUSTOM_MIRRORING`, and `CUSTOM_INTERCEPT`.
- **Centralized Management:** Manage security configurations for your entire GCP organization from a single source-controlled repository.
- **Integration with Custom Module:** Leverages the custom Terraform module we built for reliable and consistent resource deployment.

## Prerequisites

Before using this configuration, ensure the following prerequisites are met:

1.  **Google Cloud Organization:** You must have a Google Cloud Organization, as Security Profiles are organization-level resources.
2.  **Terraform Installed:** Install Terraform (v1.3.0 or later) on your local machine or CI/CD environment.
3.  **Google Cloud SDK:** Install and authenticate the Google Cloud SDK (`gcloud`) to manage your project and resources.
4.  **IAM Permissions:** Ensure the principal (user, service account) running Terraform has the **Network Security Admin** (`roles/networksecurity.admin`) role at the **organization level**.
5.  **Terraform Security Profile Module:** The custom module for security profiles must be available at the path specified in `securityprofiles.tf` (e.g., `../../../modules/security_profile/`).
6.  **Terraform Google Provider:** Configure the Terraform Google provider with appropriate credentials and project settings (typically in a `providers.tf` file) to handle authentication and billing.

## Description

- **Security Profiles:** Define a set of threat detection and prevention behaviors. For example, a `THREAT_PREVENTION` profile specifies actions to take for different threat severities. These are created at the organization level.
- **Security Profile Groups:** A container that holds a reference to a security profile. These groups are then attached to network firewall policies to apply the profile's rules to network traffic.
- **YAML Configuration:** This setup works by reading all `.yaml` files from a specified directory (e.g., `config/`). Each YAML file declaratively defines the resources to be created. You can have one file for your production threat profile, another for a mirroring profile, etc.
- **Key YAML Blocks:**
    - `security_profile`: This block contains all the data needed to create the `google_network_security_security_profile` resource, such as its `name`, `type`, and the specific configuration block (e.g., `threat_prevention_profile`).
    - `security_profile_group`: This block contains the data for the `google_network_security_security_profile_group` resource.
    - `link_profile_to_group`: A boolean (`true` or `false`) that, if set to true, tells the module to link the profile created in the same YAML file to the group.

## Example YAML Configuration (`threat_profile_and_group.yaml`)

The following example illustrates how to define a `THREAT_PREVENTION` profile and a group, and link them together. Place this file inside your configuration directory.

```yaml
# config/threat_profile_and_group.yaml

# The GCP Organization ID where the resources will be created.
organization_id: "YOUR_ORGANIZATION_ID" # <-- Replace

# --- Defines the Security Profile resource ---
security_profile:
  create: true
  name: "prod-app-threat-profile"
  type: "THREAT_PREVENTION"
  description: "Denies critical threats for the production application"
  labels:
    app: "prod-app"
    env: "production"
  
  # Configuration block for the THREAT_PREVENTION type
  threat_prevention_profile:
    severity_overrides:
      - severity: "CRITICAL"
        action: "DENY"
      - severity: "HIGH"
        action: "ALERT"

# --- Defines the Security Profile Group resource ---
security_profile_group:
  create: true
  name: "prod-app-profile-group"
  description: "Security group for the production application"
  labels:
    app: "prod-app"

# --- Tells the module to link the two resources above ---
link_profile_to_group: true
```
<!-- BEGIN_TF_DOCS -->

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_security_profiles"></a> [security\_profiles](#module\_security\_profiles) | ../../../modules/security_profile/ | n/a |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Path to the folder containing the YAML configuration files for security profiles and groups. | `string` | `"../../../configuration/security/SecurityProfile/config"` | no |
| <a name="input_create_security_profile"></a> [create\_security\_profile](#input\_create\_security\_profile) | Default value for creating a security profile. | `bool` | `false` | no |
| <a name="input_create_security_profile_group"></a> [create\_security\_profile\_group](#input\_create\_security\_profile\_group) | Default value for creating a security profile group. | `bool` | `false` | no |
| <a name="input_custom_intercept_profile"></a> [custom\_intercept\_profile](#input\_custom\_intercept\_profile) | Default configuration for a custom intercept profile. | <pre>object({<br/>    intercept_endpoint_group = string<br/>  })</pre> | `null` | no |
| <a name="input_custom_mirroring_profile"></a> [custom\_mirroring\_profile](#input\_custom\_mirroring\_profile) | Default configuration for a custom mirroring profile. | <pre>object({<br/>    mirroring_endpoint_group = string<br/>  })</pre> | `null` | no |
| <a name="input_existing_custom_intercept_profile_id"></a> [existing\_custom\_intercept\_profile\_id](#input\_existing\_custom\_intercept\_profile\_id) | Default value for an existing custom intercept profile ID. | `string` | `null` | no |
| <a name="input_existing_custom_mirroring_profile_id"></a> [existing\_custom\_mirroring\_profile\_id](#input\_existing\_custom\_mirroring\_profile\_id) | Default value for an existing custom mirroring profile ID. | `string` | `null` | no |
| <a name="input_existing_threat_prevention_profile_id"></a> [existing\_threat\_prevention\_profile\_id](#input\_existing\_threat\_prevention\_profile\_id) | Default value for an existing threat prevention profile ID. | `string` | `null` | no |
| <a name="input_link_profile_to_group"></a> [link\_profile\_to\_group](#input\_link\_profile\_to\_group) | Default value for linking a profile to a group. | `bool` | `false` | no |
| <a name="input_location"></a> [location](#input\_location) | Default location for resources if not specified in YAML files. | `string` | `"global"` | no |
| <a name="input_security_profile_description"></a> [security\_profile\_description](#input\_security\_profile\_description) | Default description for a security profile. | `string` | `"CNCS terraform security profile"` | no |
| <a name="input_security_profile_group_description"></a> [security\_profile\_group\_description](#input\_security\_profile\_group\_description) | Default description for a security profile group. | `string` | `null` | no |
| <a name="input_security_profile_group_labels"></a> [security\_profile\_group\_labels](#input\_security\_profile\_group\_labels) | Default labels for a security profile group. | `map(string)` | `{}` | no |
| <a name="input_security_profile_group_name"></a> [security\_profile\_group\_name](#input\_security\_profile\_group\_name) | Default name for a security profile group. | `string` | `null` | no |
| <a name="input_security_profile_labels"></a> [security\_profile\_labels](#input\_security\_profile\_labels) | Default labels for a security profile. | `map(string)` | `{}` | no |
| <a name="input_security_profile_name"></a> [security\_profile\_name](#input\_security\_profile\_name) | Default name for a security profile. | `string` | `null` | no |
| <a name="input_security_profile_type"></a> [security\_profile\_type](#input\_security\_profile\_type) | Default type for a security profile. | `string` | `"THREAT_PREVENTION"` | no |
| <a name="input_threat_prevention_profile"></a> [threat\_prevention\_profile](#input\_threat\_prevention\_profile) | Default configuration for a threat prevention profile. | <pre>object({<br/>    severity_overrides = optional(list(object({<br/>      severity = string<br/>      action   = string<br/>    })), [])<br/>    threat_overrides = optional(list(object({<br/>      threat_id = string<br/>      action    = string<br/>    })), [])<br/>    antivirus_overrides = optional(list(object({<br/>      protocol = string<br/>      action   = string<br/>    })), [])<br/>  })</pre> | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_security_profile_groups"></a> [security\_profile\_groups](#output\_security\_profile\_groups) | A map of all created security profile groups with their details. |
| <a name="output_security_profiles"></a> [security\_profiles](#output\_security\_profiles) | A map of all created security profiles with their details. |
<!-- END_TF_DOCS -->