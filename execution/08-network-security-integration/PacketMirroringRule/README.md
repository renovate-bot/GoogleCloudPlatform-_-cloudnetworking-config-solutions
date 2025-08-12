# Packet Mirroring Firewall Rules

## Overview

Google Cloud Packet Mirroring Rules are a specific type of rule configured within a **Global Network Firewall Policy**. Instead of allowing or denying traffic, a mirroring rule intercepts matching traffic and forwards a copy to a specified destination for analysis, such as an intrusion detection system (IDS) or a network performance monitor.

This rule works by directing traffic to a **Security Profile Group**, which in turn contains a `CUSTOM_MIRRORING` Security Profile. This setup is a key component of Google Cloud's advanced network security capabilities, enabling deep inspection of network traffic without being in the direct path of that traffic.

This module allows for the declarative creation and management of Packet Mirroring Firewall Rules through simple YAML configuration files.

## Pre-Requisites

### Prior Step Completion:

Successful deployment of this stage requires the completion of the following stages and the existence of these prerequisite resources:

-   **01-organization:** This stage must be completed to enable the necessary Google Cloud APIs.
-   **02-networking:** This stage should be completed to set up the VPC networks that will be used.
-   **A Global Network Firewall Policy:** A parent policy must exist for the rule to be added to. This is typically created in the `03-security/Firewall/FirewallPolicy` stage.
-   **A Security Profile and Group:** A `CUSTOM_MIRRORING` Security Profile and a Security Profile Group must exist. The rule will point to this group. This is typically created in the `03-security/SecurityProfile` stage.

### Permissions:

The user or service account executing Terraform must have the following role (or equivalent permissions) at the **organization level**:

-   **Compute Organization Firewall Policy Admin** (`roles/compute.orgFirewallPolicyAdmin`)

## Execution Steps

1.  **Configuration:**

    -   Create one or more YAML configuration files (e.g., `mirror-rule-1000.yml`) inside the directory specified by the `config_folder_path` variable (e.g., `03-security/Firewall/PacketMirroringRule/config/`).
    -   Edit the YAML files to define the desired configuration for each packet mirroring rule. See the **Example** below.

2.  **Terraform Initialization:**

    -   Open your terminal and navigate to the directory containing this Terraform configuration (`03-security/Firewall/PacketMirroringRule`).
    -   Run the following command to initialize Terraform:
        ```bash
        terraform init
        ```

3.  **Review the Execution Plan:**

    -   Generate an execution plan to review the changes Terraform will make:
        ```bash
        terraform plan -var-file=../../../configuration/network-security-integration/PacketMirroringRule/packet-mirroring-rule.tfvars
        ```

4.  **Apply the Configuration:**

    -   Once satisfied with the plan, execute the `terraform apply` command to create the packet mirroring rules:
        ```bash
        terraform apply -var-file=../../../configuration/network-security-integration/PacketMirroringRule/packet-mirroring-rule.tfvars
        ```

5.  **Monitor and Manage:**

    * Use Terraform to manage updates and changes to your packet mirroring rules as needed.

## Example

The following is an example of a YAML configuration file.

```yaml
# 03-security/Firewall/PacketMirroringRule/config/mirror-web-traffic.yml

# The priority of the rule (lower number = higher priority).
priority: 1000

# The project where the parent firewall policy exists.
project_id: "<Replace with Project ID>" # e.g., "test-project-01"

# The short name of the parent firewall policy.
firewall_policy_name: "<Replace with Policy Name>" # e.g., "fwp-consumer-pm-test"

# The direction of traffic to inspect ('INGRESS' or 'EGRESS').
direction: "INGRESS"

# The full resource path of the Security Profile Group.
security_profile_group: "<Replace with Security Profile Group Path>" # e.g., "organizations/123456789012/locations/global/securityProfileGroups/my-mirroring-spg"

# An optional description for the rule.
description: "Mirrors all inbound web traffic on port 443 for inspection."

#An Action to be performed if the rule matches 
action: "mirror"

# The match block defines the traffic to be mirrored.
match:
  dest_ip_ranges:
    - "10.20.10.0/24" # e.g., the IP range of your web servers
  layer4_configs:
    - ip_protocol: "tcp"
      ports: ["443"]
```
<!-- BEGIN_TF_DOCS -->

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_packet_mirroring_rule"></a> [packet\_mirroring\_rule](#module\_packet\_mirroring\_rule) | ../../../modules/packet_mirroring_rule | n/a |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Path to the folder containing the YAML configuration files for packet mirroring rules. | `string` | `"../../../configuration/network-security-integration/PacketMirroringRule/config"` | no |
| <a name="input_description"></a> [description](#input\_description) | Description for the Packet mirroring rule. | `string` | `null` | no |
| <a name="input_disabled"></a> [disabled](#input\_disabled) | Boolean value for whether the rule is disabled. | `bool` | `false` | no |
| <a name="input_match_dest_ip_ranges"></a> [match\_dest\_ip\_ranges](#input\_match\_dest\_ip\_ranges) | List of destination CIDR IP address ranges to match. | `list(string)` | `[]` | no |
| <a name="input_match_src_ip_ranges"></a> [match\_src\_ip\_ranges](#input\_match\_src\_ip\_ranges) | List of source CIDR IP address ranges to match. | `list(string)` | `[]` | no |
| <a name="input_rule_name"></a> [rule\_name](#input\_rule\_name) | Name for the rule if not specified in the YAML file. | `string` | `null` | no |
| <a name="input_security_profile_group"></a> [security\_profile\_group](#input\_security\_profile\_group) | Full URL of the Security Profile Group to which traffic will be mirrored. | `string` | `null` | no |
| <a name="input_target_secure_tags"></a> [target\_secure\_tags](#input\_target\_secure\_tags) | List of secure tag 'tagValues/name' strings to apply the rule to. | `list(string)` | `[]` | no |
| <a name="input_tls_inspect"></a> [tls\_inspect](#input\_tls\_inspect) | Boolean Value for whether traffic should be TLS decrypted. | `bool` | `false` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_packet_mirroring_rules"></a> [packet\_mirroring\_rules](#output\_packet\_mirroring\_rules) | A map of all created packet mirroring firewall rules. |
<!-- END_TF_DOCS -->