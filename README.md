# CloudNet Config Solutions: Simplified Google Cloud Networking with Terraform 🌐

## Introduction

This repository leverages pre-built terraform templates to streamline the setup and management of Google Cloud's networking infrastructure. This project accelerates your access to managed services like AlloyDB, GKE, Vertex AI services, Cloud SQL and Memorystore for Redis Clusters while maintaining robust security boundaries between your on-premises resources and the cloud environment. By defining role-based stages, the solution ensures that only authorized users can modify specific network components, adhering to the principle of least privilege and enhancing overall security.

### Project Goals

* Simplified setup
* Enhanced security
* Scalability
* Role-based access

### Project Structure
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
              └── mrc
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
  * `04-producer`: Implements producer services like AlloyDB, Memorystore for Redis clusters, and Cloud SQL.
  * `05-producer-connectivity`: Implements networking services like Private Service Connectivity.
  * `06-consumer`: Implements consumer services like Google Compute Engine instances, Cloud Run, Workbench, AppEngine, Managed and Unmanaged Instance Groups.
  * `07-consumer-load-balancing`: Implements load balancing services. As a part of Load Balancing, the following Load Balancers are presently supported : External Application Load Balancer, External and Internal Network Passthrough Load Balancer.

* `modules`: contains reusable Terraform modules.


### Prerequisites

* **Terraform:** Ensure you have Terraform installed. Download from the official [website](https://www.terraform.io/downloads.html)

* **Google Cloud SDK (gcloud CLI):** Install and authenticate with your Google Cloud project. Follow the instructions [official documentation](https://cloud.google.com/sdk/docs/install) to install.

* **Google Cloud Project:** Have an active Google Cloud project where you'll deploy the infrastructure. You can create a new project in the Google Cloud console.

* **IAM Permissions:** Each stage's README will detail the required IAM permissions for that specific stage. Administrators must assign these permissions to users/service accounts responsible for each stage.

## Getting Started 🚀

1. **Clone the Repository:**

    ```
    git clone https://github.com/googlecloudplatform/cloudnetworking-config-solutions.git
    ```

2. **Customize Configuration:**

    Update the `*.tfvars` files in the configuration directory with your project-specific values.

3. **Navigate to a Stage:**

    Start with 00-bootstrap, then proceed sequentially through the stages.

4. **Follow Stage-Specific Instructions:**

    Each stage directory contains a README with detailed instructions. Typically, you will run:

    ```
    terraform init
    terraform plan
    terraform apply
    ```

#### Important Notes:

* Customization: Configure the provided Terraform templates to your specific networking needs.
* Dependencies: Some stages depend on resources created in earlier stages.
* State Management: Consider using a remote backend like Google Cloud Storage for robust state management.
