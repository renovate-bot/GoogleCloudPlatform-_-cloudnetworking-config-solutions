# Cloud DNS

## Centralized DNS Management with Cloud DNS

**On this page**

1. [Introduction](#introduction)
2. [Objectives](#objectives)
3. [Architecture](#architecture)
4. [Request flow](#request-flow)
5. [Deploy the solution](#deploy-the-solution)
6. [Prerequisites](#prerequisites)
7. [Deploy through “terraform-cli”](#deploy-through-terraform-cli)
8. [Optional: Delete the deployment](#optional---delete-the-deployment)
9. [Submit feedback](#submit-feedback)

## Introduction

This document guides you on implementing centralized DNS management using Google Cloud DNS. Cloud DNS provides scalable, reliable, and managed authoritative DNS services, enabling you to manage DNS zones and records for your cloud resources and hybrid environments.

The steps involved in this user journey are:

1. Define DNS zones for your organization’s domains.
2. Configure DNS records for services and resources.
3. Integrate DNS with VPC networks for private DNS resolution.
4. Validate DNS resolution from workloads in your VPCs.

Centralized DNS management simplifies service discovery, enhances security, and ensures consistent DNS resolution across your cloud and on-premises environments.

## Objectives

- Create a VPC and subnet with Private Google Access enabled.
- Configure a private DNS zone for internal service discovery.
- Deploy a Vertex AI Workbench instance in the subnet.
- Set up firewall rules for secure access and DNS resolution.
- Access the Workbench and validate DNS resolution.

## Architecture

**Centralized DNS Resolution for Cloud and Hybrid Resources:**  
The following diagram illustrates a centralized DNS architecture using Cloud DNS. Multiple VPCs and on-premises networks resolve DNS queries via Cloud DNS, ensuring consistent and secure name resolution.

<img src="./images/dns.png" alt="clouddns-architecture" width="800"/>

### Request Flow

This diagram illustrates DNS resolution and data flow between an on-premises environment and Google Cloud using Cloud DNS peering and private zones.

#### DNS Query Flow (Blue Dashed Arrows: 1–6)

1. **User VM initiates DNS query:**  
   The user VM in the on-premises network requests the IP address for `service.project-1.internal` or `service.project-2.internal`.

2. **DNS query sent to on-premises DNS server:**  
   The query is forwarded from the user VM to the local DNS server.

3. **DNS server forwards query to Cloud DNS (Peering Zone) in Host Project-1:**  
   The on-premises DNS server is configured to forward queries for specific internal zones to the Cloud DNS peering zone in Host Project-1 via Cloud Interconnect or VPN.

4. **Cloud DNS (Peering Zone) forwards query to Cloud DNS (Private Zone) in Host Project-2:**  
   If the requested record belongs to a zone managed by Host Project-2, the peering zone forwards the query to the private zone in Host Project-2.

5. **Cloud DNS (Private Zone) resolves the query:**  
   The private zone in Host Project-2 finds the record (e.g., `service.project-1.internal = 10.20.30.5`) and returns the result to the peering zone.

6. **Cloud DNS (Peering Zone) returns the resolved IP to the on-premises DNS server:**  
   The peering zone sends the resolved IP address back to the on-premises DNS server.

#### Data Path Flow (Red Dashed Arrows: 7–9)

7. **On-premises DNS server returns IP to User VM:**  
   The DNS server provides the resolved IP (e.g., `10.20.30.5`) to the user VM.

8. **User VM initiates data connection to the resolved IP:**  
   The user VM uses the returned IP to connect to the Compute Engine instance in Service Project 1.

9. **Traffic flows through Cloud Interconnect/VPN to Service Project 1:**  
   The data connection traverses the Cloud Router and Cloud Interconnect/VPN, reaching the Compute Engine instance at `10.20.30.5`.

---

**Legend:**  
- **Blue dashed arrows (1–6):** DNS query flow  
- **Red dashed arrows (7–9):** Data path flow

This setup enables seamless DNS resolution and secure connectivity between on-premises resources and Google Cloud services using Cloud DNS peering and private zones.

## Deploy the solution

This section guides you through deploying centralized DNS using Cloud DNS.

### Prerequisites

For common prerequisites, refer to the **[prerequisites.md](../../prerequisites.md)** guide.

### Deploy through terraform-cli

1. **Clone the cloudnetworking-config-solutions repository:**

  ```sh
  git clone https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions.git
  ```

2. **Navigate to the `cloudnetworking-config-solutions` folder and update the configuration files:**

  - **00-bootstrap stage**

    Update `configuration/bootstrap.tfvars` with your project IDs and user IDs/groups:

    ```hcl
    folder_id                = "<your-project-id>"
    bootstrap_project_id     = "<your-project-id>"
    dns_administrator        = ["user:user-example@example.com"]
    ```

  - **01-organisation stage**

    Update `configuration/organization.tfvars` to enable required APIs:

    ```hcl
    activate_api_identities = {
     "project-01" = {
      project_id = "your-project-id",
      activate_apis = [
        "dns.googleapis.com",
        "compute.googleapis.com",
        "notebooks.googleapis.com",
        "aiplatform.googleapis.com",
      ],
     },
    }
    ```

  - **02-networking stage**

    Update the `configuration/networking.tfvars` file with the following parameters to configure the Google Cloud Project ID, VPC, subnet, NAT, and other networking resources:

    ```hcl
    project_id = "" # Replace with your Google Cloud Project ID
    region     = "us-central1" # Specify the region for your resources

    ## VPC input variables

    network_name = "workbench-vpc" # Name of the VPC
    subnets = [
     {
      name                  = "workbench-subnet" # Name of the subnet
      ip_cidr_range         = "10.20.0.0/24" # CIDR range for the subnet
      region                = "us-central1" # Region for the subnet
      enable_private_access = true # Set to true to enable Private Google Access (required for Workbench)
     }
    ]

    # Configuration for setting up a Shared VPC Host project, enabling centralized network management and resource sharing across multiple projects.
    shared_vpc_host = false # Set to true if using a Shared VPC Host

    ## PSC/Service Connectivity Variables

    create_scp_policy      = false # Set to true to create a Service Connectivity Policy
    subnets_for_scp_policy = []  # List subnets for the SCP policy in the same region

    ## Cloud NAT input variables

    create_nat = true # Set to true to create a Cloud NAT instance

    ## Cloud HA VPN input variables

    create_havpn = false # Set to true to create a High Availability VPN
    peer_gateways = {
     default = {
      gcp = "" # Specify the peer VPN gateway, e.g., projects/<peer-project-id>/regions/<region>/vpnGateways/<vpn-name>
     }
    }

    tunnel_1_router_bgp_session_range = "169.254.1.0/30" # BGP session range for Tunnel 1
    tunnel_1_bgp_peer_asn             = 64514 # ASN for Tunnel 1 BGP peer
    tunnel_1_bgp_peer_ip_address      = "" # IP address for Tunnel 1 BGP peer
    tunnel_1_shared_secret            = "" # Shared secret for Tunnel 1

    tunnel_2_router_bgp_session_range = "169.254.2.0/30" # BGP session range for Tunnel 2
    tunnel_2_bgp_peer_asn             = 64514 # ASN for Tunnel 2 BGP peer
    tunnel_2_bgp_peer_ip_address      = "" # IP address for Tunnel 2 BGP peer
    tunnel_2_shared_secret            = "" # Shared secret for Tunnel 2

    ## Cloud Interconnect input variables

    create_interconnect = false # Set to true to create a Cloud Interconnect
    ```

    Create a YAML file for your private DNS zone in `configuration/networking/CloudDNS/DNSManagedZones/config/private-zone.yaml`:

    ```yaml
    zones:
    - zone: "workbench-internal-zone"
      project_id: "<your-gcp-project-id>"
      description: "Private zone for Workbench internal services"
      force_destroy: false
      zone_config:
        domain: "workbench.internal."
        visibility: "private"
        reverse_lookup: false
        private_visibility_config:
         networks:
          - network_url: "projects/<your-gcp-project-id>/global/networks/workbench-vpc"
      recordsets:
       - name: "db.workbench.internal."
         type: "A"
         ttl: 300
         records:
          - "10.20.0.5"
       - name: "notebook.workbench.internal."
         type: "A"
         ttl: 300
         records:
          - "<workbench-private-ip>" # Replace after deployment
    ```

  - **03-security stage**

    Update `configuration/security/workbench.tfvars` file - update the Google Cloud Project ID. This will facilitate the creation of essential firewall rules, including rules to allow SSH access to port 22 for Workbench instances with private IP configurations.

    ```hcl
    project_id = "<your-gcp-project-id>"
    network = "projects/<your-gcp-project-id>/global/networks/workbench-vpc"

    ingress_rules = {
     "allow-ssh-workbench" = {
      deny        = false
      description = "Allow SSH access to Workbench"
      priority    = 1000
      source_ranges = ["10.20.0.0/24"] # Only allow from subnet
      targets     = ["workbench-instance"]
      rules = [
        {
         protocol = "tcp"
         ports    = ["22"]
        }
      ]
     }
     "allow-dns-egress" = {
      deny        = false
      description = "Allow DNS egress from Workbench"
      priority    = 1001
      source_ranges = ["10.20.0.0/24"] # Only allow from subnet
      targets     = ["workbench-instance"]
      rules = [
        {
         protocol = "udp"
         ports    = ["53"]
        },
        {
         protocol = "tcp"
         ports    = ["53"]
        }
      ]
     }
    }
    ```
  - **03-security stage(Google Managed SSL Certificates)**
      
    Update `configuration/security/Certificates/Compute-SSL-Certs/Google-Managed/google_managed_ssl.tfvars` - update the google cloud project ID in the google_managed_ssl.tfvars.

    ```
    project_id           = "<producer-project-id>"
    ssl_certificate_name = "my-managed-ssl-cert"
    ssl_managed_domains = [
      {
        domains = ["example.com", "www.example.com"]
      }
    ]
    ```


  - **05-producer-connectivity stage**
         
    For this user journey we do not need to create any psc connections hence go to `configuration/producer-connectivity.tfvars` and update the contents of producer-connectivity.tfvars file to look like this.

    ```
    psc_endpoints = []
    ```

  - **06-consumer stage**

    Update the `configuration/consumer/workbench/config/instance-lite.yaml.example` file with the following content and rename it to `instance-lite.yaml`:

    ```yaml
    name: default-workbench-instance # Default Workbench instance name
    project_id: project-id          # Replace with your GCP project ID
    location: us-central1-a         # Default zone
    gce_setup:
     network_interfaces:
      - network: projects/project-id/global/networks/workbench-vpc   # Default network path
        subnet: projects/project-id/regions/us-central1/subnetworks/workbench-subnet # Default subnet path
    ```

3. **Execute the terraform script:**

  You can now deploy the stages individually using **run.sh** or you can deploy all the stages automatically using the run.sh file. Navigate to the `execution/` directory and run this command to run the automatic deployment using **run.sh**:

  ```sh
  ./run.sh -s all -t init-apply-auto-approve
  # or
  ./run.sh --stage all --tfcommand init-apply-auto-approve
  ```

4. **Verify your creation:**

  After deployment, verify your setup as follows:
  1. **Check Cloud DNS zones and records:**  
    In the Google Cloud Console, navigate to the Cloud DNS section and confirm that your DNS zones and records have been created.

  2. **List Workbench instances:**  
    Ensure your Workbench instance appears in the list:
  ```sh
    gcloud compute instances list --filter="name=default-workbench-instance"
  ```

  3. **SSH into the Workbench instance and set up an SSH tunnel for JupyterLab:**
     ```sh
     gcloud compute ssh default-workbench-instance \
       --project <your-gcp-project-id> \
       --zone us-central1-a \
       -- -NL 8080:localhost:8080
     ```
     Open [https://localhost:8080](https://localhost:8080) in your browser, log in, and verify you can create and run a notebook.

  4. **Validate DNS resolution:** Run DNS lookup commands on the Workbench instance:
     ```sh
     nslookup db.workbench.internal.
     nslookup notebook.workbench.internal.
     ```
     You should see the IP addresses configured in your DNS zone YAML.

     Optionally, test with `dig` (install DNS utilities such as `dnsutils` or `bind-utils` if needed):
     ```sh
     dig db.workbench.internal.
     dig notebook.workbench.internal.
     ```

  **References:**  
  - [Vertex AI Workbench Quickstart](https://cloud.google.com/vertex-ai/docs/workbench/managed/quickstart)  
  - [Cloud DNS Documentation](https://cloud.google.com/dns/docs/quickstart)

## Optional - Delete the deployment

1. In Cloud Shell or in your terminal, make sure that the current working directory is `$HOME/cloudshell_open/<Folder-name>/execution`. If it isn't, go to that directory.
2. Remove the resources that were provisioned by the solution guide:

  ```sh
  ./run.sh -s all -t destroy-auto-approve
  ```

3. When you're prompted to perform the actions, enter `yes`.

## Troubleshoot Errors

Currently, there are no known issues specific to the Vertex AI Workbench setup using Cloud DNS when following this guide.

## Submit feedback

For common troubleshooting steps and solutions, please refer to the **[troubleshooting.md](../../troubleshooting.md)** guide.

To provide feedback, please follow the instructions in our **[submit-feedback.md](../../submit-feedback.md)** guide.