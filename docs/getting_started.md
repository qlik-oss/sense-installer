# Getting started

## Requirements

- Kubernetes cluster (Docker Desktop with enabled Kubernetes)
- `kubectl` installed, configured and able to communicate with kubernetes cluster. _`qliksense` CLI uses `kubectl` under the hood to perform operations on cluster_

## Installing Sense installer

Download the executable for your platform from [releases page](https://github.com/qlik-oss/sense-installer/releases) and rename it to `qliksense`

??? tldr "Linux"

    ``` bash
    curl -Lo qliksense https://github.com/qlik-oss/sense-installer/releases/download/v0.7.0/qliksense-linux-amd64
    chmod +x qliksense
    sudo mv qliksense /usr/local/bin
    ```

??? tldr "MacOS"

    ``` bash
    curl -Lo qliksense https://github.com/qlik-oss/sense-installer/releases/download/v0.7.0/qliksense-darwin-amd64
    chmod +x qliksense
    sudo mv qliksense /usr/local/bin
    ```

??? tldr "Windows"
    Download Windows executable and add it in your `PATH` as `qliksense.exe`

    [https://github.com/qlik-oss/sense-installer/releases/download/v0.7.0/qliksense-windows-amd64.exe](https://github.com/qlik-oss/sense-installer/releases/download/v0.7.0/qliksense-windows-amd64.exe)
    


## Quick start

- To download the version `v0.0.2` from qliksense-k8s [releases](https://github.com/qlik-oss/qliksense-k8s/releases).

```shell
qliksense fetch v0.0.2
```

- To install CRDs for QSEoK and qliksense operator into the kubernetes cluster.

```shell
qliksense crds install --all
```

- To install QSEoK into a namespace in the kubernetes cluster where `kubectl` is pointing to.

```shell
qliksense install --acceptEULA="yes"
```
