# Terraform Google Cloud Firewall Packet Mirroring Rule Module

This module creates and manages a `google_compute_network_firewall_policy_packet_mirroring_rule` resource using the beta provider. This rule is placed within a Global Network Firewall Policy to intercept traffic based on a defined match condition and direct it to a Security Profile Group for mirroring.

## Usage

The following example creates a packet mirroring rule with a priority of 1000 to mirror all inbound HTTPS traffic from a specific IP range.

```terraform
module "mirror_https_rule" {
  source = "./modules/packet_mirroring_rule" # Or your module path

  # --- Rule Identification ---
  priority        = 1000
  rule_name       = "mirror-inbound-https"
  project_id      = "my-gcp-project-12345"
  firewall_policy = "my-global-firewall-policy"

  # --- Rule Behavior ---
  direction              = "INGRESS"
  action                 = "mirror"
  security_profile_group = "organizations/123456789012/locations/global/securityProfileGroups/my-mirroring-spg"
  description            = "Mirrors all inbound HTTPS traffic for inspection."

  # --- Traffic Matching ---
  match_layer4_configs = [
    {
      ip_protocol = "tcp"
      ports       = ["443"]
    }
  ]
  match_src_ip_ranges = ["0.0.0.0/0"]
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
| <a name="provider_google-beta"></a> [google-beta](#provider\_google-beta) | >= 6.20.0, < 7.0.0 |

## Resources

| Name | Type |
|------|------|
| [google-beta_google_compute_network_firewall_policy_packet_mirroring_rule.primary](https://registry.terraform.io/providers/hashicorp/google-beta/latest/docs/resources/google_compute_network_firewall_policy_packet_mirroring_rule) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_action"></a> [action](#input\_action) | The action to perform. Can be 'mirror', 'do\_not\_mirror', or 'goto\_next'. | `string` | `"mirror"` | no |
| <a name="input_description"></a> [description](#input\_description) | An optional description for the rule. | `string` | `null` | no |
| <a name="input_direction"></a> [direction](#input\_direction) | The direction of traffic to which the rule applies. Can be 'INGRESS' or 'EGRESS'. | `string` | n/a | yes |
| <a name="input_disabled"></a> [disabled](#input\_disabled) | Whether the rule is disabled. | `bool` | `false` | no |
| <a name="input_firewall_policy"></a> [firewall\_policy](#input\_firewall\_policy) | The name of the parent firewall policy. | `string` | n/a | yes |
| <a name="input_match_dest_ip_ranges"></a> [match\_dest\_ip\_ranges](#input\_match\_dest\_ip\_ranges) | A list of destination CIDR IP address ranges to match. | `list(string)` | `[]` | no |
| <a name="input_match_layer4_configs"></a> [match\_layer4\_configs](#input\_match\_layer4\_configs) | A list of L4 configs to match (protocol and optional ports). | <pre>list(object({<br/>    ip_protocol = string<br/>    ports       = optional(list(string))<br/>  }))</pre> | n/a | yes |
| <a name="input_match_src_ip_ranges"></a> [match\_src\_ip\_ranges](#input\_match\_src\_ip\_ranges) | A list of source CIDR IP address ranges to match. | `list(string)` | `[]` | no |
| <a name="input_priority"></a> [priority](#input\_priority) | The priority of the rule (0-2147483647). Lower numbers have higher priority. | `number` | n/a | yes |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | The project ID where the firewall policy resides. | `string` | n/a | yes |
| <a name="input_rule_name"></a> [rule\_name](#input\_rule\_name) | A user-friendly name for the rule. | `string` | `null` | no |
| <a name="input_security_profile_group"></a> [security\_profile\_group](#input\_security\_profile\_group) | The full URL of the Security Profile Group to which traffic will be mirrored. Required if action is 'mirror'. | `string` | `null` | no |
| <a name="input_target_secure_tags"></a> [target\_secure\_tags](#input\_target\_secure\_tags) | A list of secure tag 'tagValues/name' strings to apply the rule to. | `list(string)` | `[]` | no |
| <a name="input_tls_inspect"></a> [tls\_inspect](#input\_tls\_inspect) | Boolean flag indicating if the traffic should be TLS decrypted. | `bool` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_priority"></a> [priority](#output\_priority) | The priority of the created firewall rule. |
| <a name="output_rule_name"></a> [rule\_name](#output\_rule\_name) | The name of the created firewall rule. |
<!-- END_TF_DOCS -->