# Getting started

## Requirements

- Docker Desktop with Kubernetes enabled

## Installing `qliksense` CLI

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

Pick a released version from qliksense-k8s [releases](https://github.com/qlik-oss/qliksense-k8s/releases)

- Check that your environment fullfills Qlik Sense requirements

```shell
qliksense preflight all
```

- Fetch version `v0.0.2`

```shell
qliksense fetch v0.0.2
```

- Install CRDs for QSEoK and qliksense operator into the kubernetes cluster

```shell
qliksense crds install --all
```

- To install QSEoK into a namespace in the kubernetes cluster where `kubectl` is pointing to.

```shell
qliksense install --acceptEULA="yes"
```

### Accessing newly installed Qlik Sense

- Add to your hosts file the following line:

    ```
    127.0.0.1 elastic.example
    ```

    - Linux - `/etc/hosts`
    - MacOS - `/etc/hosts`
    - Windows - `C:\Windows\System32\drivers\etc\hosts`

- Point your browser to <https://elastic.example>
