# Security Stage

## Overview

This Terraform stage focuses on establishing essential security configurations for various Google Cloud Platform (GCP) resources, including AlloyDB, Memorystore for Redis Clusters (MRC), CloudSQL, and GCE (Google Compute Engine). The core component of this stage is setting up firewall rules to control inbound and outbound traffic to these resources. This Stage also helps you deploy advanced security features such as firewall policies, firewall endpoints, firewall endpoint associations, security profiles and security profile groups.

## Prerequisites

### Enabled APIs:

Based on the producer (such as CloudSQL, MRC or AlloyDB) or consumer service (such as GCE, CloudRun, MIGs or Workbench) that you use, you can enable their APIs in order to manage the setup :

- Compute Engine API
- Cloud IAM API
- Cloud Logging API
- Cloud Monitoring API
- Notebooks API
- Network Security API


### Permissions:

The user or service account running Terraform should have sufficient IAM permissions to create and manage firewall rules, potentially including:

- Compute Security Admin

### Previous stage completion

**Completed Prior Stages:** Successful deployment of security resources depends on the completion of the following stages:
  * **01-organization:** This stage handles the activation of required Google Cloud APIs.
  * **02-networking:** This stage handles the creation of required networking resources.

## Components

1. AlloyDB Firewall (03-security/AlloyDB): Defines firewall rules to secure AlloyDB instances.
2. MRC Firewall (03-security/MRC): Defines firewall rules to secure Memorystore Redis Cloud instances.
3. CloudSQL Firewall (03-security/CloudSQL): Defines firewall rules to secure CloudSQL instances.
4. GCE Firewall (03-security/GCE): Defines firewall rules for GCE instances, specifically focusing on SSH access.
5. MIG Firewall (03-security/MIG) : Defines firewall rules for MIGs to allow health checks for the instance groups.
6. Workbench Firewall (03-security/Workbench): Configures firewall rules for Workbench instances, ensuring secure SSH access and enabling access to necessary resources.
7. Google Managed Compute SSL Certificates (03-security/Certificates/Compute-SSL-Certs/Google-Managed) : Configures Google Managed Compute SSL Certificates which can be used with Load Balancing for secure communication.
8. Security Profile (03-security/SecurityProfile): Defines firewall rules to simplify the process of creating and managing Google Cloud Security Profiles and Security Profile Groups

## Configuration

Each component's configuration is handled within its respective .tfvars file. The common configuration parameters include:

- project_id: Your GCP project ID.
- network: The VPC network to apply firewall rules to.
- default_rules_config: Configuration for default firewall rules (refer to variables.tf for details).
- egress_rules: Specific outbound firewall rules (refer to variables.tf for details).
- ingress_rules: Specific inbound firewall rules (for gce-firewall.tf).

Ensure that you modify these values within each file to match your environment's specific configuration requirements. You can find the confirguration files for the following security components under the `configuration/security` folder.

- GCE : gce.tfvars
- MRC : mrc.tfvars
- CloudSQL : cloudsql.tfvars
- AlloyDB : alloydb.tfvars
- MIG : mig.tfvars
- Workbench : workbench.tfvars
- SecurityProfile : securityprofile.tfvars

## Usage

1. **Adjust Variables**

Open and modify the tfvars files to set values for project_id, network, default_rules_config, egress_rules, and ingress_rules as needed.

2. **Terraform Steps**:

- Initialize: Run `terraform init`.
- Plan: Run `terraform plan -var-file=../../../configuration/security/your-component.tfvars` to review the planned changes.
- Apply:  If the plan looks good, run `terraform apply -var-file=../../../configuration/security/your-component.tfvars` to create or update the resources.

## Important Notes

- **Firewall Rules**: Carefully review and customize the firewall rules (default_rules_config, egress_rules, ingress_rules) to match your organization's security policies.

- **Additional Security**: You can consider additional security measures such as VPC Service Controls, Identity-Aware Proxy (IAP), and Security Command Center alongside these firewall rules.
