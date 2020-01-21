# Qlik Sense installation and operations CLI

- [Qlik Sense installation and operations CLI](#qlik-sense-installation-and-operations-cli)
  - [About](#about)
    - [Future Direction](#future-direction)
  - [Getting Started](#getting-started)
    - [Requirements](#requirements)
    - [Download](#download)
    - [Generate Credentials from published bundle](#generate-credentials-from-published-bundle)
    - [Qlik Sense version and image list](#qliksense-version-and-image-list)
    - [Optional: Pulling images in manifest locally, "air gap"](#optional-pulling-images-in-manifest-locally-%22air-gap%22)
    - [Running Preflight checks](#running-preflight-checks)
    - [Installation](#installation)
      - [Supported Parameters during install](#supported-parameters-during-install)
      - [How To Add Identity Provider Config](#how-to-add-identity-provider-config)
  - [Packaging a Custom bundle](#packaging-a-custom-bundle)
  
## About

The Qlik Sense installer CLI (sense-installer) provides an imperitive interface to many of the configurations that need to be applied against the declaritive structure described in [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s).

This is a technology preview that uses [porter](https://porter.sh) to execute "actions" (operations) and bundle versions of the [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s) repository.

These bundles are posted to [docker hub](https://hub.docker.com/) at the following location: [qliksense-cnab-bundle](https://hub.docker.com/r/qlik/qliksense-cnab-bundle/tags).

For each version of a qliksense sense edge build there should be a corresponding release current posted on docker hub. ex. `qlik/qliksense-cnab-bundle:v1.21.23-edge` for `v1.21.23` edge release of qliksense. The latest version posted will also be labelled as `latest`

### Future Direction

- Porter is currently used as a core technology for the CLI. In the future Porter will be moved "up the stack" to allow the CLI to perform the current and expanded operations independently and encapsulate core functionality currently provided by Porter and other dependent tooling.
- More operations:
  - Expanded preflight checks
  - backup/restore operations

## Getting Started

### Requirements

- Docker Client connected to a docker engine into which images can built, pulled and pushed.
  - (Docker Desktop setup tested for these instructions)
  
### Download

- Download the appropriate executable for your platform from the [releases page](https://github.com/qlik-oss/sense-installer/releases).
- To allow the CLI to download and initialize dependencies (including porter and it's associated mixins), simply execute `qliksense` with no arguments
  - `qliksense`
  
#### Porter CLI
- *Optional*: If wanting to use porter CLI directly, two environment variables will need to be set so as not to conflict with an existing porter installation:
  - _Bash_

    ```shell
    bash# export PORTER_HOME="$HOME\.qliksense"
    bash# export PATH="$HOME\.qliksense;$PATH"
    ```

  - _PowerShell_

    ```shell
    PS> $Env:PORTER_HOME="$Env:USERPROFILE\.qliksense"
    PS> $Env:PATH="$Env:USERPROFILE\.qliksense;$Env:PATH"
    ```



### Generate Credentials from published bundle

- Ensure connectivity to the target cluster create a kubeconfig credential for a target bundle. 
  - generating a file as follows, replace `<credential_name>` with a name of your choosing.
    - _Bash_
    ```shell
    cat <<EOF > $HOME/.qliksense/credentials/kube-cred.yaml
    name: <credential_name>
    credentials:
    - name: kubeconfig
      source:
        path: $HOME/.kube/config
    EOF
    ```
    - _PowerShell_
    ```shell
    PS> Add-Content -Value @"
    name: <credential_name>
    credentials:
    - name: kubeconfig
      source:
        path: $Env:USERPROFILE\.kube\config
    "@ -Path $Env:USERPROFILE\.qliksense\credentials\kube-cred.yaml
    ```
  - credentials can also be created using the [porter](https://porter.sh) CLI *(the correct environmental variable need to have been set up as shown in [Porter CLI](#porter-cli) above)*
    - `porter cred generate <credential_name> --tag qlik/qliksense-cnab-bundle:v1.21.23-edge`, replace `<credential_name>` with a name of your choosing.
    - Select `file path` and specify full path to a kube config file ex. _Bash_: `/home/user/.kube/config` or _PowerShell_ `C:\Users\user\.kube\config`

### Qlik Sense version and image list

It is possible verify the version of the [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s) repository bundled into the `qlik/qliksense-cnab-bundle` image and retreive the list of images included in that release. (This operation can take a minute or so)<https://github.com/qlik-oss/kustomize/issues/13> as the entire manifests needs to be rendered:

- `qliksense about --tag qlik/qliksense-cnab-bundle:<qliksense_version>`

### Optional: Pulling images in manifest locally, "air gap"

If the `dockerRegistry` parameter is specified as the private docker registry to be used by the kubernetes cluster hosting qliksense, it is possible to pull images to the local docker engine for an eventual push during a `qliksense install` or `qliksense upgrade`

### Running Preflight checks

You can run preflight checks to ensure that the cluster is in a healthy state before installing Qliksense.

- `qliksense preflight -c <credential_name> --tag qlik/qliksense-cnab-bundle:<qliksense_version>`

The above command runs the checks in the default namespace. If you want to specify the namespace to run preflight checks on:

- `qliksense preflight --param namespace=<value> -c <credential_name> --tag qlik/qliksense-cnab-bundle:<qliksense_version>`

### Installation

- Install the bundle : `qliksense install --param acceptEULA=yes -c <credential_name> --tag qlik/qliksense-cnab-bundle:<qliksense_version>`

#### Supported Parameters during install

| Name        | Descriptions           | Default  |
| ------------- |:-------------:| -----:|
| profile      | select a profile i.e docker-desktop, aws-eks, gke | docker-desktop |
| acceptEULA      | yes | has to be yes |
| namespace      | any kubernetes namespace      |   default |
| dockerRegistry      | A private docker image regitry for pods    |   not specified (public) |
| rotateKeys | regenerate application PKI keys on upgrade (yes/no)      |    no |
| mongoDbUri | the mongodb URI to use      |    URI of development mongodb |
| scName | storage class name      |    none |

#### How To Add Identity Provider Config

Since idp configs are usually multiline configs it is not conventional to pass to porter during install as a `param`. Rather put the configs in a file and refer to that file during `porter install` command. For example to add `keycloak` IDP create file named `idpconfigs.txt` and put

```shell
idpConfigs=[{"discoveryUrl":"http://keycloak-insecure:8089/keycloak/realms/master22/.well-known/openid-configuration","clientId":"edge-auth","clientSecret":"e15b5075-9399-4b20-a95e-023022aa4aed","realm":"master","hostname":"elastic.example","claimsMapping":{"sub":["sub","client_id"],"name":["name","given_name","family_name","preferred_username"]}}]

```

Then pass that file during install command like this

```shell
qliksense install --param acceptEULA=yes -c  <credential_name> --param-file idpconfigs.txt --tag qlik/qliksense-cnab-bundle:<qliksense_version>`
```

## Packaging a Custom bundle

If files need to be added to the [qliksense-k8s repository](https://github.com/qlik-oss/qliksense-k8s) in order to perform advanced configuration outside the scope of the what the operator provides, a custom bundle needs to be built.

Packaging of Qlik Sense on Kubernetes is done through a [Porter](https://porter.sh/) definition in the [Qlik Sense on Kubernetes configuration repository](https://github.com/qlik-oss/qliksense-k8s/blob/master/porter.yaml), the resulting bundle published on DockerHub as a [Cloud Natvie Application Bundle](https://cnab.io/) called [qliksense-cnab-bundle](https://hub.docker.com/r/qlik/qliksense-cnab-bundle)

To start, clone [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s) and modify the repo as desired, once finished make sure to be in the `qliksense-k8s` directory from which the porter bundle can be built:

```shell
git clone git@github.com:qlik-oss/qliksense-k8s.git

cd qliksense-k8s

qliksense build
```

Once built, all of the `porter` command that were used with `--tag` can be now be used without this flag provided that porter is executed with the `qliksense-k8s` directory. `porter` will automatically use the qliksense-k8s (and the porter.yaml) in the current directory.
