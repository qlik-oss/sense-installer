# (WIP) Qlik Sense installation and operations CLI

- [Qlik Sense installation and operations CLI](#qlik-sense-installation-and-operations-cli)
  - [About](#about)
    - [Future Direction](#future-direction)
  - [Getting Started](#getting-started)
    - [Requirements](#requirements)
    - [Download](#download)
    - [TL;DR](#TL;DR)
    - [How qliksense CLI works](#how-qliksense-cli-works)
      - [Witout Git Repo](#Without-git-repo)
      - [With Git Repo](#With-a-git-repo)
    - [Air Gapped](#air-gaped)
  
## About

The Qlik Sense installer CLI (qliksense) provides an imperitive interface to many of the configurations that need to be applied against the declaritive structure described in [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s). This cli faciliates to do

- installation of QSEoK
- installation of qliksense operator to manage QSEoK
- air gapped installation of QSEoK

This is a technology preview that uses qlik modified [kustomize](https://github.com/qlik-oss/kustomize) to kubernetes manifests of the versions of the [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s) repository.

For each version of a qliksense sense edge build there should be a corresponding release in [qliksense-k8s] repository under [releases](https://github.com/qlik-oss/qliksense-k8s/releases)

### Future Direction

- More operations:
  - Expanded preflight checks
  - backup/restore operations
  - fully support airgap installation of QSEoK
  - restore unwanted deletion of kubernetes resoureces

## Getting Started

### Requirements

- `kubectl` need to be installed and configured properly so that `kubectl` can connect to the kubernetes cluser. The `qliksense` CLI uses `kubectl` under the hood to perform operations on cluster
  - (Docker Desktop setup tested for these instructions)
  
### Download

- Download the appropriate executable for your platform from the [releases page](https://github.com/qlik-oss/sense-installer/releases) and rename it to `qliksense`. All the examplease down below uses `qliksense`.
  
### TL;DR

- execute `qliksense fetch v0.0.2` to download the version `v0.0.2` from qliksense-k8s [releases](https://github.com/qlik-oss/qliksense-k8s/releases).
- execute `qliksense crds install --all` to install CRDs for QSEoK and qliksense operator into the kubernetes cluster.
- execute `qliksense install v0.0.2 --acceptEULA="yes"` to install QSEoK into a namespace in the kubernetes cluster where `kubectl` is pointing to.


## How qliksense CLI works

`qliksense` CLI workes in two modes
- with a git repo fork/clone of [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s)
- without git repo

### Without git repo

In this mode `qliksense` CLI download the specified version from [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s) and put it into `~/.qliksense/contexts/<context-name>/qlik-k8s`. The `qliksense` cli creates a default qliksense context (different from kubectl context) named `qlik-default` which will be the prefix for all kubernetes resources created by the cli under this context latter on. New context and configuration can be created by the CLI

```console
$ qliksense config
do operations on/around CR

Usage:
  qliksense config [command]

Available Commands:
  apply         generate the patchs and apply manifests to k8s
  list-contexts retrieves the contexts and lists them
  set           configure a key value pair into the current context
  set-configs   set configurations into the qliksense context
  set-context   Sets the context in which the Kubernetes cluster and resources live in
  set-secrets   set secrets configurations into the qliksense context
  view          view the qliksense operator CR

Flags:
  -h, --help   help for config

Use "qliksense config [command] --help" for more information about a command.
```

The qliksense cli create a CR for the qliksense operator and all the config operations are peformed to edit the CR. So when `qliksense install` or `qliksense config apply` both generate patches in local file system (i.e `~/.qliksense/contexts/<context-name>/qlik-k8s`) and install those manifests into the cluster and ask `qliksene operator` to take the ownership of those resource so that when `qliksenes uninstall` is performed the operator can delete all those kubernetes resources for the current context.

### With a git repo

User has to create fork or clone of [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s) and push it to their own git server. When user perform `qliksense install` or `qliksene config apply` the qliksense operator does these tasks

- downloads the corresponding version of manifests from the user's git repo.
- generate kustomize patches
- install kubernetes resoruces 
- push those generated patches into a new branch in the provided git repo. so that user user can merge those patches into their master branch. 
- spinup a cornjob to monitor master branch. If user modifies anything in the master branch those changes will be applied into the cluster. This is a light weight `git-ops` model

This is how repo info is provided into the CR

```console
qliksense config set git.repository="https://github.com/my-org/qliksense-k8s"

qliksense config set git.accessToken=blablalaala
```

## Air gaped

