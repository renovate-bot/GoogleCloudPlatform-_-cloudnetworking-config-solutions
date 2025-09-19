Create a Google Kubernetes Engine Cluster using Private Service Connectivity with Cloud SQL using Private Service Access accessed using Google Compute Engine
---

**On this page**

  [Introduction](#introduction)

  [Objectives](#objectives)

  [Request flow](#request-flow)

  [Architecture](#architecture)

  [Prerequisites](#prerequisites)

  [Deploy the solution](#deploy-the-solution)

  [Optional:Delete the deployment](#optional-delete-the-deployment)

  [Known Issues](#known-issues)

  [Troubleshoot Errors](#troubleshoot-errors)

  [Submit feedback](#submit-feedback)

Introduction
---

*Simplified networking solution so you can spin up infrastructure in minutes using terraform\!*

Deploy your containerized applications with speed and simplicity using Google Kubernetes Engine (GKE)\!

This guide provides a comprehensive walkthrough for deploying a Google Kubernetes Engine cluster in your Google Cloud Platform (GCP) environment and connecting it to Cloud SQL for persistent data storage. Google Kubernetes Engine is a managed Kubernetes service that allows you to easily run containerized applications at scale. Cloud SQL, on the other hand, offers fully managed relational databases like MySQL, PostgreSQL, and SQL Server. By combining Google Kubernetes Engine with Cloud SQL, you gain the flexibility and scalability of containers with the reliability and performance of managed databases. In this solution, you shall also create a Google Compute Engine instance in order to access and manage your clusters. 

Objectives
---

This walkthrough will guide you through the essential stages of Google Kubernetes Engine deployment, including network configuration, and application deployment, with a focus on integrating Cloud SQL for your application's data needs. We'll leverage infrastructure-as-code with Terraform to automate the deployment process, enabling you to spin up your Google Kubernetes Engine cluster in minutes.

This guide would have Google Kubernetes Engine control plane communication use Private Service Connectivity and Cloud SQL would be using Private Service Access. This is to commonly used while different networking options are adopted in a single architecture. To read more about Private Service Connectivity and Private Service Access, refer to our [official documentation - PSA](https://cloud.google.com/vpc/docs/private-service-connect) and [official documentation - PSA](https://cloud.google.com/vpc/docs/private-services-access).

We'll provide clear instructions and code examples to ensure a smooth and successful deployment. While going through each stage, please ensure that you have the necessary permissions required. The stages are:

* Bootstrap stage :  Granting the required IAM permissions to your service accounts.  
* Organization stage : Enabling the necessary APIs for Google Kubernetes Engine and related services.  
* Networking stage : Creating a Virtual Private Cloud (VPC) for your Google Kubernetes Engine cluster. Defining subnets for your nodes and services.  
* Security stage : Create firewall rules for your Cloud SQL instance
* Producer stage : Provisioning your Google Kubernetes Engine cluster with desired configurations (e.g., machine type, node pools, autoscaling), Provisioning your CloudSQL instance with desired configurations
* Producer Connectivity : **skipped**
* Consumer stage : Provisioning your Google Compute Engine instance with desired configurations (e.g., machine type, attached disks)

Throughout each stage, we'll provide guidance on recommended variables and configurations to tailor the deployment to your specific needs.

Let's get started\!

Request Flow
---

The request processing flow for the deployed topology, which allows Google Kubernetes Engine Cluster using Private Service Connectivity with Cloud SQL using Private Service Access accessed using Google Compute Engine : 

1. **User initiates a request:** This could be a user accessing an application deployed on the Google Kubernetes Engine cluster, triggering an action that requires data interaction.

2. **Application interacts with Cloud SQL:** The application running on the Google Kubernetes Engine cluster sends a request to the Cloud SQL instance for data retrieval or storage. This communication happens over the VPC network using the Cloud SQL Private Service Access connection.

3. **Cloud SQL processes the request:** The Cloud SQL instance receives the request, processes it (e.g., retrieves data from the database), and sends a response back to the application.

4. **Application receives the response:** The application on the Google Kubernetes Engine cluster receives the response from Cloud SQL and continues its operation.

5. **User accesses the application via Compute Engine:** A user, possibly an administrator or developer, connects to the Google Compute Engine instance.

6. **Compute Engine interacts with GKE:** The user on the Compute Engine instance uses tools like `kubectl` to interact with the Google Kubernetes Engine cluster, potentially to manage deployments, monitor applications, or troubleshoot issues. This communication also happens securely within the VPC network, leveraging the Private Service Connect connection to the Google Kubernetes Engine control plane.

7. **Compute Engine may interact with Cloud SQL:** If needed, the user on the Compute Engine instance can also directly interact with the Cloud SQL instance for database administration or data analysis tasks. This again uses the Private Service Access connection for secure and private communication.

This flow illustrates how the different components – Google Kubernetes Engine, Cloud SQL, and Google Compute Engine – work together to provide a secure, private, and efficient environment for running and managing applications with persistent data storage.

Architecture
---

<img src="./images/gke-psc-sql-psa.png" alt="gke-psc-sql-psa" width="400"/>

(**GKE with PSC and External Endpoint with SQL using PSA**)

This solution will guide you how to establish a connection to a

The main components that are deployed in this architecture are the following : 

1. **Consumer Google Kubernetes Engine Project & Virtual Private Cloud**  
2. **Producer Google Kubernetes Engine Project & Virtual Private Cloud**  
3. **Producer CloudSQL Project & Virtual Private Cloud**  
4. **Google Kubernetes Engine Cluster : Google Kubernetes Engine Nodes & Pods**  
5. **Google Kubernetes Engine Internal Endpoint**  
6. **Cloud SQL instance**  
7. **SQL Private Service Connect Endpoint**
8. **Google Compute Engine instance**

Deploy the Solution
---
This section guides you through the process of deploying the solution.

Prerequisites
---

For the common prerequisites for this repository, please refer to the **[prerequisites.md](../prerequisites.md)** guide. Any additional prerequisites specific to this user journey will be listed below.

Deploy through terraform-cli
---

Here’s a guide to configure tfvars for each stage and then use run.sh to automatically deploy the solution.

1. **Clone the** cloudnetworking-config-solutions repository repository:

    ```
    git clone https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions.git
    ```

2. Navigate to the configuration/ directory and use the following tfvars for reference for in-place modifications. 

    Bootstrap stage (configuration/bootstrap.tfvars) : 

    * You will need to create a service account with the necessary permissions to access your Cloud SQL instance.
    * Add the following project IDs and user IDs/groups in the tfvars.

    ```c
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

    Organisation Stage (configuration/organisation.tfvars) : 

    * You will need to enable the required APIs for Cloud SQL and Google Kubernetes Engine.
    * Add your project ID here in which you wish to enable the APIs for Google Kubernetes Engine Clusters.

    ```
    activate_api_identities = {
      "project-01" = {
        project_id = "your-project-id",
        activate_apis = [
          "servicenetworking.googleapis.com",
          "iam.googleapis.com",
          "compute.googleapis.com",
          "sqladmin.googleapis.com",
          "container.googleapis.com"
        ],
      },
    }
    ```

    Networking Stage (configuration/networking.tfvars) : 

    * You will need to create the required Virtual Private Cloud (VPC), subnets & IP ranges for Google Kubernetes Engine clusters.
    * Add your project ID here in which you wish to create the VPC, Subnet and NAT for Google Kubernetes Engine Clusters.

    ```c

    project_id = "your-project-id"

    region     = "us-central1"

    ## VPC input variables

    network_name = "CNCS_VPC"
    subnets = [
      {
        ip_cidr_range = "10.0.0.0/24"
        name          = "CNCS_VPC_Subnet_1"
        region        = "us-west1-a"
        secondary_ip_ranges = {
            ip_range_pods = "192.168.0.0/16"
            ip_range_services = "192.169.0.0/24"
          }
      }
    ]

    # PSC/Service Connecitvity Variables

    create_scp_policy      = false
    subnets_for_scp_policy = [""]

    ## Cloud Nat input variables

    create_nat = true

    ## Cloud HA VPN input variables

    create_havpn = false
    ```
    Security Stage (configuration/security/cloudsql.tfvars) :

    **NOTE :** 

    1. **Before moving forward, please delete the security/alloydb.tfvars and security/mrc.tfvars files as our CUJ only involves CloudSQL.**  
    2. **GKE Firewall rules are automatically created, as a user you wouldn’t need to create new firewall rules.**

    For CloudSQL (configuration/security/cloudsql.tfvars): 

    * You will need to configure firewall rules to allow traffic between your GKE cluster and Cloud SQL instance and to your GKE cluster and Cloud SQL instance as well.
    * Use the same project ID as used above for enabling APIs and creation of networking resources. This stage should create the necessary firewall rules for CloudSQL security.

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

    For Google Compute Engine (configuration/security/gce.tfvars): 

    * You will need to configure firewall rules to allow traffic between your GKE cluster and Google Compute Engine instance and to your GKE cluster and Google Compute Engine instance as well.
    * Use the same project ID as used above for enabling APIs and creation of networking resources. This stage should create the necessary firewall rules for Google Compute Engine security.

    ```
    project_id = "your-project-id"

    name = "CNCS_VPC"
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

    Producer Stage (configuration/producer/GKE/config/instance1.yaml.example) : 

    * You will need to create configuration YAML files for creation of Google Kubernetes Engine clusters.
    * Use the same project ID as used above for the creation of the Google Kubernetes Engine cluster.

    ```c

    project_id: your-project-id

    name : gke-CNCS-cluster
    network : CNCS_VPC
    subnetwork : CNCS_VPC_Subnet_1
    ip_range_pods : gke-cluster-range-for-pods
    ip_range_services : gke-cluster-range-for-services
    kubernetes_version : 1.29

    ```

    Producer Stage (configuration/producer/CloudSQL/config/instance1.yaml.example) : 

    * You will need to create configuration YAML files for creation of Cloud SQL instances.
    * Use the same project ID as used above for the creation of the CloudSQL PSA instance.

    ```c

    name: CNCS_CloudSQL_instance
    project_id: your-project-id
    region: us-central1
    database_version: MYSQL_8_0
    network_config:
      connectivity:
        psa_config:
          private_network : projects/your-project-id/global/networks/CNCS_VPC
    ```

    Consumer Stage (configuration/consumer/GCE/instance1.yaml) : 

    * You will need to create configuration YAML files for creation of Google Compute Engine clusters.
    * Use the same project ID as used above for the creation of the Google Compute Engine cluster.

    ```c
    project_id: your-project-id

    name: CNCS_GCE_instance
    region : us-central1
    zone: us-central1-a
    image: ubuntu-os-cloud/ubuntu-2204-lts
    network: projects/your-project-id/global/networks/CNCS_VPC
    subnetwork: projects/your-project-id/regions/us-central1/subnetworks/CNCS_VPC_Subnet_1
    ```

3. Now, navigate to the execution/ directory and run this command to run the automatic deployment using run.sh : 

```c
sh run.sh -s all -t init-apply
```

Here, \-s flag with all values will run all **s**tages and \-t flag with value init-apply will ask **t**erraform to use init and apply steps.

Usage  
---

This solution shall help your applications on GKE connect to PSA based Cloud SQL instances. Once your deployment is complete, you can deploy containerized applications to your GKE cluster. This can be done through various methods, such as:

* **kubectl:** Use the `kubectl` command-line tool to deploy your applications from YAML manifests or Helm charts.

**Now, to connect to your newly created Google Kubernetes Clusters:**

You can connect to your GKE cluster using the following methods:

* **gcloud CLI:** Use the `gcloud container clusters get-credentials` command to configure your `kubectl` to interact with your cluster.  
* **kubectl:** Once your `kubectl` is configured, you can use it to interact with your cluster, deploy applications, and manage resources.  
* **Cloud Console:** Access your GKE cluster through the Google Cloud Console to view its status, manage resources, and troubleshoot issues.
* **Compute Engine:** Access your GCE instance to use `gcloud/kubectl` to administer or monitor your GKE cluster

To learn more about connecting to GKE clusters go through our [public documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl).

Optional-Delete the deployment
---

Once you’re done with using the environment, you can destroy the resources using the run.sh automated script with this command from parent folder : 

```c
sh run.sh -s all -t destroy
```

Before destroying, ensure that if you’d any critical data/applications you’ve safely moved them.

Known Issues
---

No known issues for this example at the moment, however if you run into any issues please feel free to create an issue/bug in this repository. 

Troubleshoot Errors
---
For common troubleshooting steps and solutions, please refer to the **[troubleshooting.md](../troubleshooting.md)** guide.

Submit feedback
---

To provide feedback, please follow the instructions in our **[submit-feedback.md](../submit-feedback.md)** guide.