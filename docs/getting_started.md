# Getting started

## Requirements

- `kubectl` need to be installed and configured properly so that `kubectl` can connect to the kubernetes cluser. The `qliksense` CLI uses `kubectl` under the hood to perform operations on cluster
  - (Docker Desktop setup tested for these instructions)

## Download

- Download the appropriate executable for your platform from the [releases page](https://github.com/qlik-oss/sense-installer/releases) and rename it to `qliksense`. All the examples below uses `qliksense`.

### Linux

```console
curl -fsSL https://raw.githubusercontent.com/qlik-oss/sense-installer/ibiqlik/installscript/scripts/install.sh -o install-sense-cli.sh
sh install-sense-cli.sh
```

## Quick start

- To download the version `v0.0.2` from qliksense-k8s [releases](https://github.com/qlik-oss/qliksense-k8s/releases).

```shell
$qliksense fetch v0.0.2
```

- To install CRDs for QSEoK and qliksense operator into the kubernetes cluster.

```shell
$qliksense crds install --all
```

- To install QSEoK into a namespace in the kubernetes cluster where `kubectl` is pointing to.

```shell
$qliksense install --acceptEULA="yes"
```