# CloudNet Config Solutions: Simplified Google Cloud Networking with Terraform 🌐

## Introduction

This repository leverages pre-built Terraform templates to streamline the setup and management of Google Cloud's networking infrastructure. This project accelerates your access to managed services like AlloyDB, GKE, Vertex AI services, Cloud SQL and Memorystore for Redis Clusters while maintaining robust security boundaries between your on-premises resources and the cloud environment. By defining role-based stages, the solution ensures that only authorized users can modify specific network components, adhering to the principle of least privilege and enhancing overall security.

## Key Features and Enhancements

### Network Services

*   **Network Connectivity Center (NCC):** Simplified consumption using VPC as a Spoke, Hybrid Spokes, and Producer VPC as a Spoke. ([VPC as a Spoke](docs/Networking/NCC/ncc-mesh.md), [Producer VPC as a spoke](docs/Networking/NCC/ncc-producer-vpc.md), [Hybrid VPC spokes](docs/Networking/NCC/hybrid-spoke-havpn.md))
*   **Firewall Endpoints and Firewall Endpoint Association:** New features for enhanced network security. ([Firewall Endpoint Documentation](docs/Networking/FirewallEndpoints/firewallendpoints.md))
*   **Hybrid Connectivity (VPN/Interconnect):** Extend your on-premises network to Google Cloud to allow secure access to services like AlloyDB from your on-prem environment. ([Interconnect Documentation](docs/Networking/interconnect.md))
*   **Networking Componenets:** It empowers you to create a secure, highly available, and customizable network infrastructure that aligns with your organization's specific requirements. ([Networking Documentation](execution/02-networking/README.md))

### Security Services

*   **Security Profiles and Security Profile Groups:** Added for improved security management. ([Documentation](docs/SecurityProfiles/securityprofiles.md))
*   **Secure Firewall Rules for Google Cloud Vertex AI Workbench:** Added for secure access to Vertex AI Workbench instances. ([Documentation](execution/03-security/Workbench/README.md))
*   **Firewall Rules for Google Cloud Managed Instance Groups (MIGs):** Added for secure communication between MIG instances, including health checks. ([Documentation](execution/03-security/MIG/README.md))
*   **Firewall Policies:** Added for centralized and scalable management of firewall rules. ([Documentation](execution/03-security/Firewall/FirewallPolicy/README.md))
*   **Google Compute Managed SSL Certificate:** Facilitates the creation and management of Google Compute Managed SSL Certificate. ([Documentation](execution/03-security/Certificates/Compute-SSL-Certs/Google-Managed/README.md))

### Producers

*   **AlloyDB:** Deploys AlloyDB clusters with options for both Private Service Access (PSA) and Private Service Connect (PSC). ([PSA Documentation](docs/AlloyDB/alloydbinstance-using-psa-accessed-from-gce.md), [PSC Documentation](docs/AlloyDB/alloydbinstance-using-psc.md))
*   **Cloud SQL:** Deploys Cloud SQL instances with options for both PSA and PSC. ([PSA Documentation](docs/CloudSQL/cloudsqlinstance-using-psa-accessed-from-gce.md), [PSC Documentation](docs/CloudSQL/cloudsqlinstance-using-psc-accessed-from-gce.md))
*   **GKE:** Deploys Google Kubernetes Engine (GKE) clusters with various networking configurations. ([GKE Documentation](docs/GKE/gke-gce.md))
*   **Memorystore for Redis Cluster (MRC):** Deploys MRC instances for high-performance, in-memory data storage. ([MRC Documentation](docs/MRC/mrc-accessed-using-scp-using-gce.md))
*   **Vector Search:** Deploys Vector Search for building high-performance vector similarity search engines. ([Vector Search Documentation](docs/VectorSearch/vectorsearch.md))
*   **Vertex AI Online Endpoints:** Deploys Vertex AI endpoints for real-time predictions. ([Vertex AI Documentation](execution/04-producer/Vertex-AI-Online-Endpoints/README.md))

### Producer Connectivity

*   **Private Service Connect (PSC):** Securely connects services across different VPC networks using PSC. ([Producer Connectivity Documentation](execution/05-producer-connectivity/README.md))

### Consumers

*   **Vertex AI Workbench:** Enhanced networking for creating private and secure deployments. ([Documentation](docs/Workbench/workbench-instance-using-pga.md))
*   **App Engine Standard Environments:** Smoother network integration for scalable web and mobile backends. ([Standard Documentation](docs/AppEngine/appengine-standard.md))
*   **App Engine Flexible Environments:** Smoother network integration for scalable web and mobile backends. ([Flexible Documentation](docs/AppEngine/appengine-flexible.md))
*   **Backend resources:** Increased support with MIG and UMIG as backend resources for LBs. ([MIG Documentation](execution/06-consumer/MIG/README.md), [UMIG Documentation](docs/UMIG/umig.md)))
*   **App Engine (Standard & Flexible) Environments:** Smoother network integration for scalable web and mobile backends. ([Standard Documentation](docs/AppEngine/appengine-standard.md), [Flexible Documentation](docs/AppEngine/appengine-flexible.md))
*   **Cloud Run (Jobs):** Support for running jobs with direct VPC egress or through a Serverless VPC Access connector. ([Direct VPC Egress Documentation](docs/CloudRun/cloudrun-job-direct-vpc-egress.md), [Serverless VPC Connector Documentation](docs/CloudRun/cloudrun-job-serverless-vpc-connector.md))

### Deployment Features

*   **Click-to-Deploy Functionality:**
    *   Increased coverage for AlloyDB: Expanded support with PSA and PSC. ([PSA Documentation](docs/AlloyDB/alloydbinstance-using-psa-accessed-from-gce.md)
    *   Expanded support for External Load Balancers. ([ELB Documentation](docs/LoadBalancer/external-application-lb-mig.md))
    *   Expanded support for External Network Passthrough Load Balancers. ([ENLB Documentation](docs/LoadBalancer/external-network-passthrough-lb-mig.md))
    *   Expanded support for Internal Network Passthrough Load Balancers. ([INLP Documentation](docs/LoadBalancer/internal-network-passthrough-lb-mig.md))


##  Project Structure
The project is structured into the following folders:

  ```
    cloudnetworking-config-solutions
      ├──configuration
          ├── bootstrap.tfvars
          ├── organization.tfvars
          ├── networking.tfvars
          ├── producer-connectivity.tfvars
          ├── producer
              ├── alloydb
              ├── cloudsql
              ├── gke
              ├── vectorsearch
              ├── vertex-ai-online-endpoints
              ├── mrc
              └── bigquery
          ├── consumer
              ├── cloudrun
              ├── gce
              ├── mig
              ├── workbench
              ├── umig
              ├── severless
                ├── appengine
                    ├── flexible
                    ├── standard
                ├── cloudrun
                    ├── job
                    ├── service
                ├── vpcaccessconnector
          ├──security
              ├── certificates
                ├── compute-ssl-certs
                    ├── google-managed
                        ├── google_managed_ssl.tfvars
              ├── alloydb.tfvars
              ├── cloudsql.tfvars
              ├── gce.tfvars
              ├── mig.tfvars
              ├── mrc.tfvars
              └── workbench.tfvars
          └──consumer-load-balancing
              ├── application load balancers
                ├── external
              ├── network load balancers
                ├── passthrough
                    ├── internal
                    └── external
      ├──execution
          ├── 00-bootstrap
          ├── 01-organization
          ├── 02-networking
          ├── 03-security
          ├── 04-producer
          ├── 05-producer-connectivity
          ├── 06-consumer
          └── 07-consumer-load-balancing
      ├──modules
            ├── net-vpc
            ├── psc_forwarding_rule
            ├── vector-search
            ├── vertex-ai-online-endpoints
            ├── umig
            ├── lb_http
            ├── google_compute_managed_ssl_certificate
            ├── network-connectivity-center
            └── app_engine
  ```
* `configuration`: This folder contains Terraform configuration files (*.tfvars) that hold variables used for multiple stages. These **.tfvars** files would include configurable variables such as project IDs, region or other values that you want to customize for your specific environment.

* `execution`: This folder houses the main Terraform code, organized into stages:

  * `00-bootstrap`: Sets up foundational resources like service accounts and Terraform state storage.
  * `01-organization`:  Manages organization-level policies for network resources.
  * `02-networking`: Manages VPCs, subnets, Cloud HA VPN and other core networking components like PSA, SCP, Cloud NAT.
  * `03-security`:  Configures firewalls rules, firewall policies and Google Managed SSL certificates.
  * `04-producer`: Implements producer services like AlloyDB, Memorystore for Redis clusters, BigQuery and Cloud SQL.
  * `05-producer-connectivity`: Implements networking services like Private Service Connectivity.
  * `06-consumer`: Implements consumer services like Google Compute Engine instances, Cloud Run, Workbench, AppEngine, Managed and Unmanaged Instance Groups.
  * `07-consumer-load-balancing`: Implements load balancing services. As a part of Load Balancing, the following Load Balancers are presently supported : External Application Load Balancer, External and Internal Network Passthrough Load Balancer.

* `modules`: contains reusable Terraform modules.


### Prerequisites

### [configuration](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/configuration)

Houses all the `*.tfvars` files that define customizable variables like project IDs, regions, and service-specific inputs.

- [`bootstrap.tfvars`](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/blob/main/configuration/bootstrap.tfvars)
- [`organization.tfvars`](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/blob/main/configuration/organization.tfvars)
- [`networking.tfvars`](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/blob/main/configuration/networking.tfvars)
- [`producer-connectivity.tfvars`](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/blob/main/configuration/producer-connectivity.tfvars)
- [`consumer/`](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/configuration/consumer)
- [`producer/`](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/configuration/producer)
- [`security/`](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/configuration/security)
- [`consumer-load-balancing/`](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/configuration/consumer-load-balancing)

---

### [execution](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/execution)

This is where the main Terraform logic resides — split into sequential, modular stages:

| Stage | Purpose | Link |
|-------|---------|------|
| `00-bootstrap` | Service accounts, remote state | [🔗](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/execution/00-bootstrap) |
| `01-organization` | Org policies, folders | [🔗](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/execution/01-organization) |
| `02-networking` | VPCs, Subnets, VPN, NAT, PSA, SCP, NCC, FirewallEndpoints | [🔗](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/execution/02-networking) |
| `03-security` | Firewall rules, SSL certs, Security Profiles  | [🔗](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/execution/03-security) |
| `04-producer` | AlloyDB, Cloud SQL, MRC, GKE, Vector Search, Vertex AI Online Endpoints   | [🔗](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/execution/04-producer) |
| `05-producer-connectivity` | PSC setup | [🔗](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/execution/05-producer-connectivity) |
| `06-consumer` | GCE, MIG, UMIG, Workbench, App Engine (Standard/ Flexible), Cloud Run, VPC Access Connector | [🔗](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/execution/06-consumer) 
| `07-consumer-load-balancing` | Application External Load Balancers, Network Load Balancers (Internal/External) | [🔗](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/execution/07-consumer-load-balancing) |

## Prerequisites

* **Terraform:** Ensure you have Terraform installed. Download from the official [website](https://developer.hashicorp.com/terraform/install)

* **Google Cloud SDK (gcloud CLI):** Install and authenticate with your Google Cloud project. Follow the instructions [official documentation](https://cloud.google.com/sdk/docs/install) to install.

* **Google Cloud Project:** Have an active Google Cloud project where you'll deploy the infrastructure. You can create a new project in the Google Cloud console.

* **IAM Permissions:** Each stage's README will detail the required IAM permissions for that specific stage. Administrators must assign these permissions to users/service accounts responsible for each stage.

## 🚀 Getting Started

1. **Clone the Repository**

   ```bash
   git clone https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions.git
   cd cloudnetworking-config-solutions
   ```

2. **Customize Configuration**

   Edit relevant `*.tfvars` or `yaml` configurations in the [`configuration/`](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/tree/main/configuration) folder.

3. **Execute the terraform script**
   You can now deploy the stages individually using **run.sh** or you can deploy all the stages automatically using the [run.sh](http://run.sh) file. Navigate to the execution/ directory and run this command to run the automatic deployment using **run.sh .**

      ```
      ./run.sh -s all -t init-apply-auto-approve
      or
      ./run.sh --stage all --tfcommand init-apply-auto-approve
      ```

4. **Proceed Sequentially**

   Follow `00` to `07` in order to maintain dependency consistency. Each stage has a README with instructions for updating the configuration.

---

## Important Notes

- **Customization:** Adjust templates to meet specific networking/security requirements.
- **Dependencies:** Later stages depend on outputs from earlier ones.
- **State Management:** Use Google Cloud Storage backend for state file management.