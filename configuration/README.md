# Configuration files

This directory serves as a centralized repository for all Terraform configuration files (.tfvars) used across the various stages of your infrastructure deployment. By organizing these configuration files in one place, we maintain a clear and structured approach to managing environment-specific variables and settings.

## File Organization by Stage

- 00-bootstrap stage (bootstrap.tfvars)
- 01-organisation stage (organisation.tfvars)
- 02-networking stage (networking.tfvars)
- 03-security stage
    - AlloyDB (alloydb.tfvars)
    - MRC (mrc.tfvars)
    - Cloud SQL (sql.tfvars)
    - GCE (gce.tfvars)
    - Certificates
      - Compute-SSL-Certs
        - Google-Managed
          - google_managed_ssl.tfvars
    - Firewall
      - Firewall-Policy
    - Workbench (workbench.tfvars)
    - Security-Profiles
- 04-producer stage
    - AlloyDB
      - alloydb.tfvars
      - config
        - instance.yaml.example
    - MRC
      - mrc.tfvars
      - config
        - instance.yaml.example
    - Cloud SQL
      - sql.tfvars
      - config
        - instance.yaml.example
    - GKE
      - gke.tfvars
      - config
        - instance.yaml.example
    - Vector Search
      - vectorsearch.tfvars
      - config
        - instance.yaml.example
    - Vertex-AI-Online-Endpoints
      - vertex-ai-online-endpoints.tfvars
      - config
        - endpoint.yaml.example
    - Big Query
      - bigquery.tfvars
      - config
        - dataset.yaml.example
        - dataset-with-access-controls.yaml.example
- 05-producer-connectivity stage (producer-connectivity.tfvars)
- 06-consumer stage
  - GCE
        - gce.tfvars
        - config
          - instance.yaml.example
  - MIG
    - mig.tfvars
    - config
      - instance.yaml.example
  - UMIG
    - umig.tfvars
    - config
      - instance.yaml.example
  - Serverless
    - AppEngine
      - Flexible
        - appengineflexible.tfvars
        - config
          - instance1.yaml.example
      - Standard
        - appenginestandard.tfvars
        - config
          - instance1.yaml.example
    - CloudRun
      - Job
        - cloudrunjob.tfvars
        - config
          - instance.yaml.example
      - Service
        - cloudrunservice.tfvars
        - config
          - instance.yaml.example
    - VPCAccessConnector
  - Workbench
    - config
      - instance-lite.yaml.example
      - instance-expanded.yaml.example
- 07-consumer-load-balancing stage
  - Application
    - External
      - external-application.tfvars
      - config
        - instance1.yaml.example
        - instance2.yaml.example
  - Network
    - Passthrough
      - External
        - external-network-passthrough.tfvars
        - config
          - instance-expanded.yaml.example
          - instance-lite.yaml.example


# Usage

## Specifying Variable Files

When executing a Terraform stage (e.g. terraform plan, terraform apply, terraform destroy), you must explicitly instruct Terraform to use the corresponding configuration file. This is achieved using the `-var-file` flag followed by the relative path to the .tfvars file.

## Relative Paths

Relative paths are essential for maintaining flexibility and ensuring your Terraform configuration works across different environments. While running any of the stages, use the [-var-file flag](https://developer.hashicorp.com/terraform/language/values/variables#variable-definitions-tfvars-files) to give relative path of the .tfvars file. Let's assume you're within the networking directory and want to execute terraform plan using the networking.tfvars configuration file:

```none
terraform plan -var-file=../configuration/networking.tfvars
```

This would run the terraform plan based on the values for the variables declared in the networking.tfvars file in the `configuration` folder. In this example:

- `-var-file` : instructs Terraform to load variables from the specified file.
- `../` : moves up one directory level from networking.
- `configuration/networking.tfvars` : points to the configuration folder containing the networking.tfvars file.

## Benefits of Centralized Configuration

- Improved Readability: A dedicated directory makes it easy to locate and manage configuration files.
- Enhanced Maintainability: Changes to environment-specific variables can be made in one place, minimizing the risk of errors.
- Streamlined Collaboration: Team members can easily access and understand the configuration structure.
- Simplified Automation: Terraform workflows can automatically reference the appropriate configuration file based on the stage being executed.


# Stage wise details

## 00-bootstrap

  - This tfvars file provides project IDs and administrator email addresses for different stages of the infrastructure setup. These values are used by Terraform to configure resources and permissions in the respective Google Cloud projects.
  - bootstrap project ID (`bootstrap_project_id`): project used to create resources such as service accounts or grant permissions to users to run the stages.
  - networking projects (`network_hostproject_id`/`network_serviceproject_id`) : host/service project IDs
  - Administrators : in stage wise administrator variables, you can set user accounts/groups to delegate permissions.


**Example usage**

## 00-bootstrap

```
folder_id                             = ""
bootstrap_project_id                  = ""
network_hostproject_id                = ""
network_serviceproject_id             = ""

organization_administrator          = ["user:organization-user-example@example.com"]
networking_administrator            = ["user:networking-user-example@example.com"]
security_administrator              = ["user:security-user-example@example.com"]

producer_cloudsql_administrator     = ["user:cloudsql-user-example@example.com"]
producer_gke_administrator          = ["user:gke-user-example@example.com"]
producer_alloydb_administrator      = ["user:alloydb-user-example@example.com"]
producer_vertex_administrator       = ["user:vertex-user-example@example.com"]
producer_mrc_administrator          = ["user:mrc-user-example@example.com"]

producer_connectivity_administrator = ["user:connectivity-user-example@example.com"]

consumer_gce_administrator          = ["user:gce-user-example@example.com"]
consumer_cloudrun_administrator     = ["user:cloudrun-user-example@example.com"]
consumer_mig_administrator          = ["user:mig-user-example@example.com"]
consumer_umig_administrator         = ["user:umig-user-example@example.com"]
consumer_lb_administrator           = ["user:lb-user-example@example.com"]
```

## 01-organization

**Example usage**

```
  activate_api_identities = {
    "project-01" = {
      project_id = "test-project",
      activate_apis = [
        "servicenetworking.googleapis.com",
        "alloydb.googleapis.com",
        "sqladmin.googleapis.com",
        "iam.googleapis.com",
        "compute.googleapis.com",
        "redis.googleapis.com",
        "aiplatform.googleapis.com",
        "container.googleapis.com",
        "run.googleapis.com",
        "appengine.googleapis.com",
        "cloudbuild.googleapis.com",
        "cloudresourcemanager.googleapis.com",
        "artifactregistry.googleapis.com",
        "notebooks.googleapis.com",
        "vpcaccess.googleapis.com",
        "bigquery.googleapis.com",
      ],
    },
  }
```

## 02-networking

**Example usage**

```
project_id = "test-project"
region     = "us-central1"

## VPC input variables

network_name = "network-test"
subnets = [
  {
    ip_cidr_range = "10.0.0.0/24"
    name          = "subnet-test"
    region        = "us-central1"
  }
]

create_scp_policy      = true
subnets_for_scp_policy = ["subnet-test"]

create_nat = true

create_havpn = false
peer_gateways = {
  default = {
    gcp = "" # e.g. projects/<google-cloud-peer-projectid>/regions/<google-cloud-region>/vpnGateways/<peer-vpn-name>
  }
}

tunnel_1_router_bgp_session_range = ""
tunnel_1_bgp_peer_asn             = 64514
tunnel_1_bgp_peer_ip_address      = ""
tunnel_1_shared_secret            = ""

tunnel_2_router_bgp_session_range = ""
tunnel_2_bgp_peer_asn             = 64514
tunnel_2_bgp_peer_ip_address      = ""
tunnel_2_shared_secret            = ""

create_interconnect = false
```

## 03-security

  - `project_id` : this variable identifies the GCP project where the firewall rule will be created.
  - `network` : this variable specifies the name of the Virtual Private Cloud (VPC) network to which the firewall rule will be applied. Firewall rules control traffic flow in and out of your VPC network.
  - `egress_rules/ingress_rules` : this variable defines a set of egress (outbound) firewall rules. These rules determine what kind of outgoing traffic is permitted from your VPC network to destinations outside the network.

***Example Usage**

```
project_id              = "project-id"
network                 = "network-name"
egress_rules = {
  allow-egress = {
    deny = false
    rules = [{
      protocol = "tcp"
      ports    = ["6379"]
    }]
  }
}
```

## 04-producer

Producer specific configuration examples can be found under the `/config` folder of that specific producer. Such as for AlloyDB, the example would be in the folder `configuration/producer/AlloyDB/config/instance.yaml.example`.

## 05-producer-connectivity (producer-connectivity.tfvars)

The `producer-connectivity.tfvars` file defines configurations for Private Service Connect (PSC) endpoints. These endpoints enable connectivity between consumer and producer services, such as Cloud SQL, AlloyDB, or other targets.

### Key Variables

1. `endpoint_project_id`: Consumer project ID where the forwarding rule is created.
2. `producer_instance_project_id`: Project where the producer service (e.g., Cloud SQL, AlloyDB) is created.
3. `subnetwork_name`: Name of the subnetwork within the VPC network from which the internal IP address for the PSC connection will be allocated.
4. `network_name`: VPC network hosting the subnetwork mentioned above.
5. `ip_address_literal`: **(Optional)** Specific internal IP address for the PSC connection. Leave null for automatic allocation.
6. `region`: Region where the PSC endpoint is created.
7. `producer_cloudsql`: **(Optional)** Configuration for Cloud SQL instances. Includes:
   - `instance_name`: Name of the Cloud SQL instance.
8. `producer_alloydb`: **(Optional)** Configuration for AlloyDB instances. Includes:
   - `instance_name`: Name of the AlloyDB instance.
   - `cluster_id`: ID of the AlloyDB cluster.
9. `target`: **(Optional)** Service attachment URL for other targets.

### Example Usage

```hcl
psc_endpoints = [
  // Configuration for a PSC endpoint with a CloudSQL instance
  {
    endpoint_project_id          = "your-endpoint-project-id"
    producer_instance_project_id = "your-producer-instance-project-id"
    subnetwork_name              = "subnetwork-1"
    network_name                 = "network-1"
    ip_address_literal           = "10.128.0.26"
    region                       = "us-central1"
    producer_cloudsql = {
      instance_name = "psc-instance-name"
    }
  },
  // Configuration for a PSC endpoint with an AlloyDB instance
  {
    endpoint_project_id          = "your-endpoint-project-id"
    producer_instance_project_id = "your-producer-instance-project-id"
    subnetwork_name              = "subnetwork-2"
    network_name                 = "network-2"
    ip_address_literal           = "10.128.0.27"
    region                       = "us-central2"
    producer_alloydb = {
      instance_name = "your-alloydb-instance-name"
      cluster_id    = "your-cluster-id"
    }
  },
  // Configuration for a PSC endpoint with a target
  {
    endpoint_project_id          = "your-endpoint-project-id"
    producer_instance_project_id = "your-producer-instance-project-id"
    subnetwork_name              = "subnetwork-3"
    network_name                 = "network-3"
    ip_address_literal           = "10.0.0.10"
    region                       = "us-central1"
    target                       = "projects/your-project-id/regions/us-central1/serviceAttachments/your-service-attachment-id"
  }
]
```

### Notes

- **Cloud SQL Configuration**: Use the `producer_cloudsql` block to specify the Cloud SQL instance name.
- **AlloyDB Configuration**: Use the `producer_alloydb` block to specify the AlloyDB instance name and cluster ID.
- **Target Configuration**: Use the `target` field to specify the service attachment URL for other targets.
- Ensure that the `region` field is specified for all PSC endpoints to avoid deployment issues.

## 06-consumer

Consumer specific configuration examples can be found under the `/config` folder of that specific consumer. Such as for GCE, the example would be in the folder `configuration/consumer/GCE/config/`.

## 07-consumer-load-balancing

Consumer load balancing specific configuration examples can be found under the `/config` folder of that specific load balancer. Such as for Application External Load Balancer, the example would be in the folder `configuration/consumer-load-balancing/Application/External/config/`.

## Considerations

- Sensitive Data: If your configuration files contain securrely handle sensitive values (e.g., API keys) and ensure they are securely stored. We strong recommend to not store senstive information in plain text and suggest you to carefully manage sensitive information.
