# How qliksense cli works

At the initialization, `qliksense` cli creates few files in the director `~/.qliksene` and it contains following files:

```console
.qliksense
├── config.yaml
├── contexts
│   └── qlik-default
│       └── qlik-default.yaml
└── ejson
    └── keys
```

`qlik-default.yaml` is a default CR created with some default values like:

```yaml
apiVersion: qlik.com/v1
kind: Qliksense
metadata:
  name: qlik-default
spec:
  profile: docker-desktop
  secrets:
    qliksense:
    - name: mongoDbUri
      value: mongodb://qlik-default-mongodb:27017/qliksense?ssl=false
  rotateKeys: "yes"
  releaseName: qlik-default
```

The `qliksense` cli creates a default qliksense context (different from kubectl context) named `qlik-default` which will be the prefix for all kubernetes resources created by the cli under this context later on. 

New context and configuration can be created by the cli, get available commands using:

```console
qliksense config -h
```

---

`qliksense` cli works in two modes

- With a git repo fork/clone of [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s)
- Without git repo

## Without git repo

In this mode `qliksense` CLI downloads the specified version from [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s) and places it in `~/.qliksense/contexts/<context-name>/qlik-k8s` folder.

The qliksense cli creates a CR for the QlikSense operator and all config operations are peformed to edit the CR.

`qliksense install` or `qliksense config apply` will generate patches in local file system (i.e `~/.qliksense/contexts/<context-name>/qlik-k8s`) and

- Install those manifests into the cluster 
- Create a custom resoruce (CR) for the `qliksene operator`.

The operator makes the association to the installed resoruces so that when `qliksense uninstall` is performed the operator can delete all kubernetes resources related to QSEoK for the current context.

## With a git repo

Create a fork or clone of [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s) and push it to your git repo/server

To add your repo into CR, perform the following:

```bash
qliksense config set git.repository="https://github.com/my-org/qliksense-k8s"
qliksense config set git.accessToken="<mySecretToken>"
```

When you perform `qliksense install` or `qliksene config apply`, qliksense operator performs these tasks:

- Download corresponding version of manifests from the your git repo
- Generate kustomize patches
- Install kubernetes resources
- Push generated patches into a new branch in the provided git repo. _Gives you ability to merge patches into your master branch_
- Create a CronJob to monitor master branch. Any changes pushed to master branch will be applied into the cluster. _This is a light weight `git-ops` model_

## GitOps

To enable gitops, the following section should be in the CR

```yaml
....
spec:
  git:
    repository: https://github.com/<OWNER>/<REPO>
    accessToken: "<git-token>"
    userName: "<git-username>"
  gitOps:
    enabled: "yes"
    schedule: "*/5 * * * *"
    watchBranch: <myBranch>
    image: qlik-docker-oss.bintray.io/qliksense-repo-watcher
....
```

##Preflight checks
Preflight checks provide pre-installation cluster conformance testing and validation before we install qliksense on the cluster. We gather a suite of conformance tests that can be easily written and run on the target cluster to verify that cluster-specific requirements are met.
The suite consists of a set of `collectors` which run the specifications of every test and `analyzers` which analyze the results of every test run by the collector.
We support the following tests at the moment as part of preflight checks, and the range of the suite will be expanded in future.

### DNS check
Run the following command to view help about the commands supported by preflight at any moment:
```console
qliksense preflight
perform preflight checks on the cluster

Usage:
  qliksense preflight [command]

Examples:
qliksense preflight <preflight_check_to_run>

Usage:
qliksense preflight dns

Available Commands:
  dns         perform preflight dns check
```

Run the following command to perform preflight DNS check. The expected output is also shown below.
```console
qliksense preflight dns
  Running Preflight checks ⠧
--- PASS DNS check
      --- DNS check passed
--- PASS   cluster-preflight-checks
PASS
```
