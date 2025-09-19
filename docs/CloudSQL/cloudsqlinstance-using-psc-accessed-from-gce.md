# Create an Cloud SQL Instance with Secure Private Connectivity Through Private Service Connect (PSC) Accessed Using Google Compute Engine

**On this page**

1. [Introduction](#introduction)

2. [Objectives](#objectives)

3. [Architecture](#architecture)

4. [Request flow](#request-flow)

5. [Architecture Components](#architecture-components)

6. [Deploy the solution](#deploy-the-solution)

7. [Prerequisites](#prerequisites)

8. [Deploy through terraform-cli](#deploy-through-terraform-cli)

9. [Optional-Delete the deployment](#optional-delete-the-deployment)

10. [Submit feedback](#submit-feedback)

### Introduction

This guide is designed to provide a clear understanding and assist database administrators, cloud architects, and cloud engineers in optimizing the deployment process. By utilizing simplified networking within the Google Cloud platform, it automates the extensive setup required for leveraging Cloud SQL. The guide assumes familiarity with cloud computing concepts, Terraform, Cloud SQL, Virtual Private Cloud (VPC), and Private Service Connect (PSC).

This guide provides instructions on using the **cloudnetworking-config-solutions** repository on GitHub to automate establishing your Cloud SQL cluster on Google Cloud. Terraform enables you to formalize your infrastructure as code, which facilitates deployment and ensures consistency, even in complex architectures.

### Objectives

This solution guide helps you do the following :

* Set up VPC, Subnets and private connectivity using PSC
* Learn about Cloud SQL instance and configurations
* Create Cloud SQL instance with secure private connectivity in the producer project with PSC enabled
* Cloud SQL instance created will have a Service attachment for PSC connection
* Endpoint in the consumer project to access the Service Attachment
* Create a GCE instance
* Perform CRUD operations from a GCE instance to an Cloud SQL instance using the private IP of the Cloud SQL instance

### Architecture

This solution deploys a Cloud SQL instance and a GCE instance.The solution also creates all the necessary components—such as VPC, subnets, and firewall rules \- required by the Cloud SQL and GCE instance.

It covers two scenarios for Cloud SQL instance using private connectivity with PSC:

* **Scenario 1: Simple Connectivity Within a VPC:** Suitable for deployments entirely within the Google Cloud environment.

    <img src="./images/cloudsql_psc_image1.png" alt="Within-a-vpc" width="400"/>

* **Scenario 2: Connectivity with Cloud HA VPN:** Ideal for connecting your Cloud SQL cluster to on-premises or other cloud networks using Google Cloud High Availability Virtual Private Network (HA VPN).

    <img src="./images/cloudsql_psc_image2.png" alt="using-ha-vpn" width="400"/>

### Request flow

The following illustrates the request processing flow for both scenarios:

1. **Scenario 1 :** Within the Google Cloud project, a user initiates a request from a GCE instance in the consumer project. The instance routes this request to a Cloud SQL instance, utilizing Private Service Connect for secure, private connectivity. The GCE instance accesses the producer's managed services via a private IP, while the producer project exposes these services through service attachments, ensuring the SQL instance remains private and not exposed to the public Internet.

2. **Scenario 2 :** A user initiates a request from a virtual machine (VM) instance located outside the Google Cloud project. The on-premises (or external) location connects to the Google Cloud network (VPC) through a Cloud High Availability (HA) VPN. The VM instance processes the request and routes it over the HA VPN to the Google Cloud VPC, forwarding it to the Cloud SQL instance via its private IP. In this setup, the VM in the customer’s  user project accesses managed services using a Private Service Connect (PSC) endpoint created inside the customer consumer project, while the HA VPN ensures private connectivity between customer user project and customer consumer project. The producer project securely exposes its services to the consumer project through service attachments, keeping the SQL instance off the public Internet. The customer’s user and consumer projects are linked via a hybrid networking solution.

## **Architecture Components**

In this example Cloud SQL uses Private Service Connect (PSC) to enable private IP connectivity between the consumer and producer projects. The consumer project accesses managed services via private IP, while the producer exposes services through service attachments without exposing SQL instances to the public Internet.

The customer’s Google Cloud organization consists of multiple projects, each created and managed by the customer. The consumer project (12345) accesses services from the producer project (56789), which hosts resources like Cloud SQL. Google also manages a separate producer project (45678) within its own organization, handling resources created by customers in their consumer projects. Connectivity between the customer’s VM in the consumer project and the Cloud SQL service is established through a Private Service Connect (PSC) Endpoint, a reserved internal IP that forwards requests to a service attachment in the producer project. This setup ensures that the Cloud SQL producer VPC is isolated, with the VM accessing the service privately, while a NAT in the customer project enables internet access for package retrieval without exposing SQL instances to the public Internet.

## **Deploy the solution**

This section guides you through the process of deploying the solution.

### **Prerequisites**

For the common prerequisites for this repository, please refer to the **[prerequisites.md](../prerequisites.md)** guide. Any additional prerequisites specific to this user journey will be listed below.

####

### Deploy through terraform-cli

1. **Clone the** cloudnetworking-config-solutions repository repository:

    ```
    git clone https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions.git
    ```

2. Navigate to **cloudnetworking-config-solutions** folder and update the files containing the configuration values
   * **00-bootstrap stage**
     * Update configuration/bootstrap.tfvars **\-** update the google cloud project IDs and the user IDs/groups in the tfvars.

        ```
        bootstrap_project_id                      = "your-project-id"
        network_hostproject_id                    = "your-project-id"
        network_serviceproject_id                 = "your-project-id"
        organization_stage_administrator          = ["user:user-example@example.com"]
        networking_stage_administrator            = ["user:user-example@example.com"]
        security_stage_administrator              = ["user:user-example@example.com"]
        producer_stage_administrator              = ["user:user-example@example.com"]
        producer_connectivity_stage_administrator = ["user:user-example@example.com"]
        consumer_stage_administrator              = ["user:user-example@example.com"]
        ```

   * **01-organisation stage**
     * Update configuration/organization.tfvars \- update the google cloud project ID and the list of the APIs to enable for the Cloud SQL instance.

        ```
        activate_api_identities = {
          "project-01" = {
            project_id = "your-project-id",
            activate_apis = [
              "servicenetworking.googleapis.com",
              "sqladmin.googleapis.com",
              "iam.googleapis.com",
              "compute.googleapis.com",
              ],
          },
        }
        ```
   * **02-networking stage**
     * Update configuration/networking.tfvars \- update the Google Cloud Project ID and the parameters for additional resources such as VPC, subnet, and NAT as outlined below.

        ```
        project_id  = "your-project-id",
        region      = "us-central1"

        ## VPC input variables
        network_name = "CNCS_VPC"
        subnets = [
          {
            ip_cidr_range = "10.0.0.0/24"
            name          = "CNCS_VPC_Subnet_1"
            region        = "us-central1-a"
          }
        ]

        ## PSC/Service Connectivity Variables
        create_scp_policy  = false

        ## Cloud Nat input variables
        create_nat = true
        ## Cloud HA VPN input variables
        create_havpn = false
        ```

   * **03-security stage**
     * Update configuration/security/gce.tfvars file \- update the Google Cloud Project ID. This will facilitate the creation of essential firewall rules, granting GCE instances the ability to transmit traffic to Cloud SQL instances.

        ```
        project_id = "your-project-id"
        network    = "CNCS_VPC"
        ingress_rules = [
          {
            name        = "allow-ssh-custom-ranges"
            description = "Allow SSH access from specific networks"
            priority    = 1000
            source_ranges = [
              "", # Source ranges such as "192.168.1.0/24" or "10.0.0.0/8"
            ]
            target_tags = ["ssh-allowed", "https-allowed"]
            allow = [{
              protocol = "tcp"
              ports    = ["22", "443"]
            }]
          }
        ]
        ```

      * Update configuration/security/cloudsql.tfvars file \- update the Google Cloud Project ID. This will facilitate the creation of essential firewall rules, granting GCE instances the ability to transmit traffic to Cloud SQL instances.

          ```
          project_id   = "your-project-id",
          network      = "CNCS_VPC"
          egress_rules = {
            allow-egress-cloudsql = {
              deny = false
              rules = [{
                protocol = "tcp"
                ports    = ["3306"]
              }]
            }
          }
          ```

   * **04-producer stage**
     * Update the execution/04-producer/CloudSQL/config/instance.yaml.example file and rename it to instance.yaml

        ```
        name: cloudsql
        project_id: "your-project-id"
        region: us-central1
        database_version: MYSQL_8_0
        network_config:
          connectivity:
            connectivity:
              psc_allowed_consumer_projects : ["your-allowed-consumer-project-id"]
        ```

    * **05-producer-connectivity stage**
      * Update the configuration/producer-connectivity.tfvars file, you can refer following snippet for reference

          ```
          psc_endpoints = [
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
          ]
          ```

   * **06-consumer stage**
     * Update the execution/06-consumer/GCE/config/instance.yaml.example file and rename it to instance.yaml

        ```
        project_id: your-project-id
        name: CNCS-GCE
        region : us-central1
        zone: us-central1-a
        image: ubuntu-os-cloud/ubuntu-2204-lts
        network: projects/<your-project-id>/global/networks/CNCS_VPC
        subnetwork: projects/<your-project-id>/regions/us-central1/subnetworks/CNCS_VPC_Subnet_1
        ```

3. **Execute the terraform script**
   You can now deploy the stages individually using **run.sh** or you can deploy all the stages automatically using the [run.sh](http://run.sh) file. Navigate to the execution/ directory and run this command to run the automatic deployment using **run.sh .**

      ```
      ./run.sh -s all -t init-apply-auto-approve
      or
      ./run.sh --stage all --tfcommand init-apply-auto-approve
      ```

4. **Verify Cluster Creation:**
   Once the deployment is complete, navigate to the Cloud SQL section in the Google Cloud Console to confirm that your cluster has been successfully created.
5. **Connect to Your Google Compute Instance & Cloud SQL Instance:**
   * **Compute Instance**
     * You can login into your compute instance , refer [link](https://cloud.google.com/compute/docs/connect/standard-ssh)

      ```
      gcloud compute ssh --project=<your-project-id> --zone=us-central1-a CNCS-GCE
      ```

   * **Cloud SQL Instance**
     * You can connect to your Cloud SQL instance from the GCE instance, refer [link](https://cloud.google.com/sql/docs/mysql/connect-compute-engine#connect-gce-private-ip)

      ```
      gcloud compute ssh --project=<your-project-id> --zone=us-central1-a CNCS-GCE

      mysql -h CLOUD_SQL_PRIVATE_IP_ADDRESS -u USERNAME  -p
      ```

* `CLOUD_SQL_PRIVATE_IP_ADDRESS` \- is the private IP address of your Cloud SQL instance.
* `USERNAME` \- you can set username and password.

## **Optional-Delete the deployment**

1. In Cloud Shell or in your terminal, make sure that the current working directory is $HOME/cloudshell\_open/\<Folder-name\>/execution. If it isn't, go to that directory.
2. Remove the resources that were provisioned by the solution guide:

    ```
    ./run.sh -s all -t destroy-auto-approve
    ```

Terraform displays a list of the resources that will be destroyed.

3. When you're prompted to perform the actions, enter yes.

## **Submit feedback**

For common troubleshooting steps and solutions, please refer to the **[troubleshooting.md](../troubleshooting.md)** guide.

To provide feedback, please follow the instructions in our **[submit-feedback.md](../submit-feedback.md)** guide.