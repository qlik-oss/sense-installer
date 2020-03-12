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
