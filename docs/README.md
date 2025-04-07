
![Skale Logo Placeholder](logo_w_name-.png)

[![Awesome](https://cdn.jsdelivr.net/gh/sindresorhus/awesome@latest/media/badge.svg)](https://github.com/sindresorhus/awesome)

![Version](https://img.shields.io/badge/version-v0.1.0-blue)

[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://GitHub.com/your-username/your-repo/graphs/commit-activity)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/skale)](https://artifacthub.io/packages/search?repo=skale)

![GitHub followers](https://img.shields.io/github/followers/TheWendril)

![GitHub contributors](https://img.shields.io/github/contributors/TheWendril/Skale)


**Manage your deployments with ease!**  
Skale provides a declarative approach to installing the necessary Helm chart and creating `skale.yaml` configuration files for your scalability based on resource limits specific needs.


## ✨ Features

* **Effortless Helm Installation:** Clear and concise instructions to get the Skale Helm chart up and running.
* **Intuitive `skale.yaml` Generation:** Guidance on creating Skale files for Deployments.
* **Modern and Clean Design:** Easy-to-read documentation for a smooth user experience.
* **Scalable and Flexible:** Horizontal Pod Autoscaller based on Resource Limits.

---

## 🛠️ Installation

### Prerequisites

Before you begin, ensure you have the following installed on your system:

* **kubectl:** The Kubernetes command-line tool. You can find installation instructions [here](https://kubernetes.io/docs/tasks/tools/).
* **Helm:** The package manager for Kubernetes. Installation instructions can be found [here](https://helm.sh/docs/intro/install/).

### Installing the Skale Helm Chart

Follow these steps to install the Skale Helm chart on your Kubernetes cluster:

1.  **Add the Skale Helm repository:**

    ```bash
    helm repo add skale
    ```

2.  **Update your Helm repositories:**

    ```bash
    helm repo update
    ```

3.  **Install the Skale chart:**

    You can install the chart with its default configuration or provide a custom `values.yaml` file.

    * **Default Installation:**

        ```bash
        helm install skale -n your_namespace thewendril/skale
        ```

    * **Custom Installation (using a `values.yaml` file):**

        Create a `values.yaml` file with your desired configurations (refer to the chart's documentation for available options). Then, run:

        ```bash
        helm install skale -f your-custom-values.yaml -n your_namespace thewendril/skale 
        ```

4.  **Verify the installation:**

    Check if the Skale pods are running in your Kubernetes cluster:

    ```bash
    kubectl get pods -n your_namespace 
    ```

    You should see pods related to the Skale deployment in a `Running` state.

---

## 📄 Creating `skale.yaml` Files

The `skale.yaml` file is crucial for defining the configuration of your Skale component. Here's a guide to creating these files:

### Basic Skale Node Configuration

A basic `skale.yaml` for deploying a Skale object might look like this:

```yaml
    apiVersion: core.skale.io/v1
    kind: Skale
    metadata:
    name: your_skale_name-skaler
    namespace: your_namespace
    spec:
    scaleTargetRef:
        apiVersion: apps/v1
        kind: Deployment # Deployment only
        name: deployment_name
    minReplicas: 1
    maxReplicas: 10
    metrics:
    - type: Resource
        resource: 
        name: cpu
        targetAverageUtilization: 80
```