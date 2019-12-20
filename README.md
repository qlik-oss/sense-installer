# Qlik Sense installation and operations CLI

The Qlik Sense installations and operations CLI provides capabilities for installing the Qlik Sense on Kubernetes packaging and performing operations on qliksense.

## Getting started

Download the appropriate executable for your platform from the [releases page](https://github.com/qlik-oss/sense-installer/releases). When used, the CLI will check to see if porter is installed, if not, will download and install it. Once done, you can find porter through `echo $HOME/.porter` on Linux and MacOS and in `$Env:USERPROFILE\.porter` on Windows. You can also install it in advance, release > 0.22.1-beta.1 is required.

To make sure everything is order, you can fetch the Qlik Sense bundle version and corresponding image list from:
 - `qliksense about --tag qlik/qliksense-cnab-bundle:latest `

## Qliksense Packaging
Packaging of Qlik Sense on Kubernetes is done through a [Porter](https://porter.sh/) definition in the [Qlik Sense on Kubernetes configuration repository](https://github.com/qlik-oss/qliksense-k8s/blob/master/porter.yaml), the resulting bundle published on DockerHub as a [Cloud Natvie Application Bundle](https://cnab.io/) called [qliksense-cnab-bundle](https://hub.docker.com/r/qlik/qliksense-cnab-bundle).
### Versioning
A version of [qliksense-cnab-bundle](https://hub.docker.com/r/qlik/qliksense-cnab-bundle) is published corresponding to an edge release. To get the latest edge release simply specify `qliksense-cnab-bundle:latest`
