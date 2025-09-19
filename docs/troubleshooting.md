# Troubleshooting Guide

This guide provides common troubleshooting steps for issues that may arise during the deployment of the solutions in this repository.

## Initial Step: Check Terraform Logs

The most crucial first step for any deployment error is to **carefully review the logs and output from the Terraform command**. The output will almost always contain a specific error message indicating which resource failed to create and why. Look for messages prefixed with `Error:`.

---

## Common Issues and Solutions

This section covers common categories of issues and the steps to diagnose them.

### 1. Connectivity Issues

#### Connectivity from On-Premises / External Networks (via VPN or Interconnect)
* **Symptom:** Cannot reach Google Cloud resources from an external network, or vice-versa.
* **Troubleshooting Steps:**
    * **Check Tunnels/Attachments:** Ensure that your Cloud VPN tunnels or VLAN attachments for Cloud Interconnect are in an `ESTABLISHED` state.
    * **BGP Sessions:** Verify that the BGP session on your Cloud Router is `ESTABLISHED` and that routes are being correctly advertised and received from your on-premises router.
    * **Firewall Rules:** Check firewalls in *both* your on-premises environment and your Google Cloud VPC. Ensure that ingress/egress rules allow traffic for the correct protocols, ports, and IP ranges.
    * **Use `ping` and `traceroute`:** Test basic connectivity from an on-premises machine to the internal IP of the Cloud Router or a VM in the VPC to identify where packets are being dropped.

#### Connectivity Within Google Cloud (VPC-to-VPC or Service-to-VPC)
* **Symptom:** Resources in one VPC cannot communicate with resources in another, or a service cannot reach a resource in a VPC.
* **Troubleshooting Steps:**
    * **VPC Peering / NCC:** If using VPC Peering or Network Connectivity Center, ensure the connections are `ACTIVE` and that route tables in both VPCs have been updated.
    * **Firewall Rules:** Review VPC Firewall Rules. A common issue is forgetting to create a rule to allow traffic from the source IP range. For load balancers and managed services, ensure you allow traffic from Google Cloud's health check ranges (`130.211.0.0/22` and `35.191.0.0/16`).
    * **Network Tags:** If your firewall rules use network tags, verify that the tags are correctly applied to your VM instances.
    * **Routing Tables:** Inspect the routing tables for your VPC to ensure there are no misconfigurations or conflicting routes.

#### Private Service Access (PSA) and Private Service Connect (PSC)
* **Symptom:** Unable to connect to managed services like Cloud SQL, AlloyDB, or Memorystore.
* **Troubleshooting Steps:**
    * **PSA:** Verify that the allocated IP range for Private Service Access is configured correctly and does not overlap with other subnets.
    * **PSC:**
        * Ensure the PSC endpoint (forwarding rule) is created in the correct consumer VPC and subnetwork.
        * Verify that the service attachment in the producer project is configured to accept the connection from your consumer project.
        * Check that the necessary IAM permissions and Service Connectivity Policies are in place if required.

### 2. Performance Issues
* **Symptom:** Slow network performance or high latency between resources.
* **Troubleshooting Steps:**
    * **Monitor Link Utilization:** Check the utilization of your VPN or Interconnect links for congestion.
    * **Optimize Routing:** Ensure that traffic is not taking a suboptimal path with unnecessary hops.
    * **Review Network Topology:** Identify and address potential bottlenecks in your architecture.
    * **Application Optimization:** Profile your application to ensure the issue is not with the application code itself.

### 3. Security and IAM Issues
* **Symptom:** Terraform fails with a `403 Forbidden` error, or resources cannot access each other due to permissions.
* **Troubleshooting Steps:**
    * **IAM Permissions:** Ensure the user or service account running Terraform has all the necessary roles. Review the `Prerequisites` section of the specific guide you are following. Apply the principle of least privilege.
    * **Service Account Scopes:** If running on a GCE VM, ensure the VM's service account has the required access scopes.
    * **API Enablement:** Make sure all required APIs (e.g., `compute.googleapis.com`, `sqladmin.googleapis.com`, `container.googleapis.com`) are enabled in the project.

---

## Validating Your Deployment

After a successful deployment, you can use the following commands from a GCE instance within your VPC to test connectivity to various services:

* **Test SSH into a VM:**
    ```sh
    gcloud compute ssh [VM_NAME] --zone [ZONE] --project [PROJECT_ID]
    ```
* **Test Cloud SQL (MySQL) instance:**
    ```sh
    mysql -h [CLOUD_SQL_PRIVATE_IP_ADDRESS] -u [USERNAME] -p
    ```
   
* **Test AlloyDB (PostgreSQL) instance:**
    ```sh
    psql -h [ALLOYDB_PRIVATE_IP_ADDRESS] -U postgres
    ```
   
* **Test Memorystore for Redis instance:**
    ```sh
    redis-cli -h [REDIS_ENDPOINT_IP] -p 6379 -c PING
    ```
   
* **Interact with a GKE Cluster:**
    ```sh
    gcloud container clusters get-credentials [CLUSTER_NAME] --region [REGION] --project [PROJECT_ID]
    kubectl get nodes
    ```
   

---

## Helpful Tools and Resources

* **Google Cloud Console:** Provides a visual interface to inspect network connectivity, VPC peering status, firewall rules, and resource configurations.
* **Cloud Logging:** Collects and analyzes logs from various Google Cloud services. **VPC Flow Logs** and **Firewall Rules Logging** are especially useful for diagnosing dropped packets.
* **Cloud Monitoring:** Provides metrics and dashboards for monitoring network performance, latency, and resource utilization.
* **Network Intelligence Center:** Use the [Connectivity Tests](https://cloud.google.com/network-intelligence-center/docs/connectivity-tests/docs) to diagnose connectivity issues between endpoints in your network.

## Known Issues

Currently, there are no known issues for most examples. However, if you encounter a problem, we recommend checking the [GitHub Issues page](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/issues) for community-reported problems and solutions.