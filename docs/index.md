# Overview

The Qlik Sense on Kubernetes CLI (`qliksense`) provides an imperative interface to many of the configurations that need to be applied against the declarative structure described in [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s).

The CLI facilitates:

- Installation of QSEoK
- Installation of qliksense operator to manage QSEoK
- Air gapped installation of QSEoK

!!! info ""
    This is a technology preview that uses Qlik modified [kustomize](https://github.com/qlik-oss/kustomize) for Kubernetes manifests on [qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s) repository

!!! info ""
    See QlikSense [edge releases on qliksense-k8s](https://github.com/qlik-oss/qliksense-k8s/releases) repository
