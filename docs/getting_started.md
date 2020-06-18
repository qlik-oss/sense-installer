# Getting started

To get familiar with the Qlik Sense on Kubernetes Operator Command Line Interface (CLI), we will install Qlik Sense on Kubernetes on docker desktop. In subsequent sections we will enhance this configuration to include an Identity Provider (keycloak) and demonstrate air gapped capabilities as well.

## Requirements

- Kubernetes cluster (Docker Desktop with enabled Kubernetes)
- `kubectl` installed, configured and able to communicate with kubernetes cluster. _`qliksense` CLI uses `kubectl` to perform some operations on cluster_

## Installing `qliksense` CLI

Download the executable for your platform from [releases page](https://github.com/qlik-oss/sense-installer/releases) and rename it to `qliksense`

??? tldr "Linux"

    ``` bash
    # bash

    curl -LOJ https://storage.googleapis.com/kubernetes-release/release/v1.16.8/bin/linux/amd64/kubectl
    curl -LOJ https://github.com/qlik-oss/sense-installer/releases/latest/download/qliksense-linux-amd64
    sudo mv qliksense-linux-amd64 kubectl /usr/local/bin
    sudo chmod ugo+x /usr/local/bin/qliksense-linux-amd64 /usr/local/bin/kubectl
    sudo ln -s /usr/local/bin/qliksense-linux-amd64 /usr/local/bin/qliksense
    sudo ln -s /usr/local/bin/qliksense-linux-amd64 /usr/local/bin/kubectl-qliksense
    ```

??? tldr "MacOS"

    ``` bash
    # bash

    curl -LOJ https://storage.googleapis.com/kubernetes-release/release/v1.16.8/bin/darwin/amd64/kubectl
    curl -LOJ https://github.com/qlik-oss/sense-installer/releases/latest/download/qliksense-darwin-amd64
    sudo mv qliksense-darwin-amd64 kubectl /usr/local/bin
    sudo chmod ugo+x /usr/local/bin/qliksense-darwin-amd64 /usr/local/bin/kubectl
    sudo ln -s /usr/local/bin/qliksense-darwin-amd64 /usr/local/bin/qliksense
    sudo ln -s /usr/local/bin/qliksense-darwin-amd64 /usr/local/bin/kubectl-qliksense
    ```

??? tldr "Windows"

    ``` powershell
    # powershell

    Invoke-WebRequest https://storage.googleapis.com/kubernetes-release/release/v1.16.8/bin/windows/amd64/kubectl.exe -O C:\bin\kubectl.exe
    Invoke-WebRequest https://github.com/qlik-oss/sense-installer/releases/latest/download/qliksense-windows-amd64.exe -O C:\bin\qliksense.exe
    Copy-Item C:\bin\qliksense.exe C:\bin\kubectl-qliksense.exe
    # Add C:\bin to current Path
    $Env:Path += ";C:\bin"
    # Save Path to User environment scope
    [Environment]::SetEnvironmentVariable("Path",[Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User) + ";C:\bin",[EnvironmentVariableTarget]::User)
    ```

## Quick start

### Setting the contexts

By default a `qlik-default` configuration context is provided and can be used, as is.  In effect, this is the name of the Qlik Sense instance in the target cluster. All resources installed into the target namespace will be prefixed with `qlik-default`.  The name of the Qlik Sense application will correspondingly be `qliksense`.

Ex.: To change this to `qliksense-dev`:

```shell
qliksense config set-context qliksense-dev
```
For the purposes of the Quick Start we will be using `qlik-default`

The target namespace is determined by the kubectl connection context. 
ex. Ensure a connection to cluster to change the configuration context's target namespace with kubectl to `qliksense`

```shell
kubectl config set-context --current --namespace=qliksense 
For the purposes of the Quick Start we will be using the default namespace. (`default`)

### Downloading a version of Qlik Sense on Kubernetes

To download the latest version of Qlik Sense on Kubernetes from qliksense-k8s

```shell
qliksense fetch
#### More Options
- To download a specific version `v1.59.20` from qliksense-k8s [releases](https://github.com/qlik-oss/qliksense-k8s/releases).
```shell
qliksense fetch v1.58.20
```
- To download from a GitHub repository fork of the `qliksense-k8s` repository (master branch). 
Ex.:
```shell
qliksense fetch --url https://github.com/bkuschel/qliksense-k8s.git master
```

### Deployment Profiles

Deployment profiles are a sets components that require sets of key/value pairs to satisfy the requirements for the generation of a Qlik Sense on Kubernetes manifest.  Along with the profile name, sets of key/value pairs are provided through the Qlik Sense custom application resources (see here). 

Profiles can be developed and added to the qliksense-k8s repo but is considered an advanced topic (see here) not covered here.

#### Default Profile: Docker Desktop

By default, the `docker-desktop` profile is associated with the configuration context when initially created. This profile is guaranteed to work on Docker Desktop but can generally be used on other types of Kubernetes clusters, provided that the required configuration tweaks are provided specific to the hosting requirements (Ex. storage class).


The docker-desktop profile does not have any scaling characteristics and is generally set up to have the ability to work on a reasonably powerful computer (16GB, 4 cores minimum, greater is better).  It also includes a self-contained mongodb instance for non-production purposes.


It generally doesn't typically require any configuration to work except an acceptance of the Qlik User License Agreement (QULA), which is prompted on install but can also be set in advance (having read the QULA). Ex:

```shell
qliksense config set-configs qliksense.acceptEULA="yes"
```

More information on the possible configuration parameters for docker-desktop here (see here).

To access an installation of the docker desktop profile in docker desktop  the host `elastic.example` needs to be added to the system host file as an alias to 127.0.0.1. Ex.

```bash
127.0.0.1 elastic.example
```

File location:

  - Linux - `/etc/hosts`
  - MacOS - `/etc/hosts`
  - Windows - `C:\Windows\System32\drivers\etc\hosts`

### Installing Qlik Sense on Kubernetes

#### Custom Resource Definitions (CRDs)

Besides the CLI, a Kubernetes operator (read here) is a core component of the Qlik Sense Operator. Additionally, there are other Kubernetes operators in Qlik Sense on Kubernetes that provide other types functionality (ex. scaling). Depending on the profile chosen (see above), additional CRDs can also be installed for third-party components. (see gke-demo).

Kubernetes operators require Custom resource definitions (CRD) (read here), which are YAML schemas for custom resources (CR). The Qlik Sense application instance, corresponding to the name of the configuration context, corresponds to a CR (ex. `qlik-default`).

CRDs require cluster scope permissions and are shared cluster-wide across namespaces. These need to be installed first (if not done previously). 

To install CRDs for Qlik Sense on Kubernetes into the Kubernetes cluster.

```shell
qliksense crds install
```

#### Preflight Checks

#### Qlik Sense

To install Qlik Sense into a namespace in the Kubernetes cluster where `kubectl` is pointing to.

```shell
qliksense install
```
