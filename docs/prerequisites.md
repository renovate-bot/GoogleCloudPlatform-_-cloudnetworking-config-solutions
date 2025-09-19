# Common Prerequisites

This document outlines the common prerequisites required to use the solutions in this repository. Please note that individual solution guides may have additional, specific requirements. Always review the "Prerequisites" section of the guide you are following.

## Required Command-Line Tools

Before you begin, you must have the following command-line tools installed and configured on the machine where you will run the deployment scripts.

### Terraform

All infrastructure in this repository is deployed using Terraform.

* **Version:** The modules are tested with **Terraform v1.8+**. Some specific solutions, like those for GKE, may require **v1.9+**. We recommend using the latest stable version of Terraform.
* **Installation:** You can find and install the appropriate binary for your system from the [official Terraform releases page](https://releases.hashicorp.com/terraform/).

### Google Cloud SDK (gcloud)

The Google Cloud SDK is required to authenticate with your Google Cloud account.

* **Installation:** Follow the official instructions to [install the gcloud SDK](https://cloud.google.com/sdk/docs/install).
* **Authentication:** After installation, authenticate your session by running the following command. This allows Terraform to use your user credentials to provision resources.
    ```sh
    gcloud auth application-default login
    ```