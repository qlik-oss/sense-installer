package qliksense

import (
	"fmt"
	kapiconfig "github.com/qlik-oss/k-apis/pkg/config"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

type InstallCommandOptions struct {
	AcceptEULA   string
	Namespace    string
	StorageClass string
}

func (q *Qliksense) InstallQK8s(version string, opts *InstallCommandOptions) error {

	// step1: fetch 1.0.0 # pull down qliksense-k8s@1.0.0
	// step2: operator view | kubectl apply -f # operator manifest (CRD)
	// step3: config apply | kubectl apply -f # generates patches (if required) in configuration directory, applies manifest
	// step4: config view | kubectl apply -f # generates Custom Resource manifest (CR)

	// fetch the version
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	fetchAndUpdateCR(qConfig, version)

	//TODO: may need to check if CRD already installed, but doing apply does not hurt for now
	//install crd into cluster
	fmt.Println("Installing operator CRD")
	if err := qapi.KubectlApply(q.GetCRDString()); err != nil {
		fmt.Println("cannot do kubectl apply on opeartor CRD", err)
		return err
	}
	// install generated manifests into cluster
	fmt.Println("Installing generated manifests into cluster")
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	}
	if opts.AcceptEULA != "" {
		if qcr.Spec.Configs == nil {
			qcr.Spec.Configs = make(map[string]kapiconfig.NameValues)
		}
		qcr.Spec.AddToConfigs("qliksense", "acceptEULA", opts.AcceptEULA)
	}
	if opts.StorageClass != "" {
		qcr.Spec.StorageClassName = opts.StorageClass
	}
	if opts.Namespace != "" {
		qcr.Spec.NameSpace = opts.Namespace
	}
	//TODO: do we need to write
	//qConfig.WriteCurrentContextCR(qcr)
	if err := applyConfigToK8s(qcr); err != nil {
		fmt.Println("cannot do kubectl apply on manifests")
		return err
	}

	// install operator cr into cluster
	//get the current context cr
	fmt.Println("Install operator CR into cluster")
	r, err := q.getCurrentCRString()
	if err != nil {
		return err
	}
	if err := qapi.KubectlApply(r); err != nil {
		fmt.Println("cannot do kubectl apply on operator CR")
	}
	return nil
}
