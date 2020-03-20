# Overview

The Qlik Sense installer CLI (`qliksense`) provides an imperative interface to many of the configurations that need to be applied against the declarative structure described in [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s). This cli facilitates:

- Installation of QSEoK
- Installation of qliksense operator to manage QSEoK
- Air gapped installation of QSEoK

!!! info ""
    This is a technology preview that uses Qlik modified [kustomize](https://github.com/qlik-oss/kustomize) for kubernetes manifests on [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s) repository

!!! info ""
    See QlikSense [edge releases on qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s/releases) repository

## Future Direction

Operations:

- Expand preflight checks
- Backup/restore operations
- Fully support airgap installation of QSEoK
- Restore unwanted deletion of kubernetes resources
