package qliksense

import (
	"fmt"

	"errors"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

type InstallCommandOptions struct {
	AcceptEULA   string
	Namespace    string
	StorageClass string
	MongoDbUri   string
	RotateKeys   string
}

func (q *Qliksense) InstallQK8s(version string, opts *InstallCommandOptions) error {

	// step1: fetch 1.0.0 # pull down qliksense-k8s@1.0.0
	// step2: operator view | kubectl apply -f # operator manifest (CRD)
	// step3: config apply | kubectl apply -f # generates patches (if required) in configuration directory, applies manifest
	// step4: config view | kubectl apply -f # generates Custom Resource manifest (CR)

	// fetch the version
	qConfig := qapi.NewQConfig(q.QliksenseHome)


	/*
		//TODO: CRD will be installed outside of operator
		//install crd into cluster
		fmt.Println("Installing operator CRD")
		if err := qapi.KubectlApply(q.GetCRDString()); err != nil {
			fmt.Println("cannot do kubectl apply on opeartor CRD", err)
			return err
		}
	*/
	// install generated manifests into cluster
	fmt.Println("Installing generated manifests into cluster")
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	} else if qcr.Spec.GetManifestsRoot() == "" {
		return errors.New("cannot get the manifest root. Use qliksense fetch <version> or qliksense set manifestsRoot")
	}

	if opts.AcceptEULA != "" {
		qcr.Spec.AddToConfigs("qliksense", "acceptEULA", opts.AcceptEULA)
	}
	if opts.MongoDbUri != "" {
		qcr.Spec.AddToSecrets("qliksense", "mongoDbUri", opts.MongoDbUri)
	}
	if opts.StorageClass != "" {
		qcr.Spec.StorageClassName = opts.StorageClass
	}
	if opts.Namespace != "" {
		qcr.Spec.NameSpace = opts.Namespace
	}
	if opts.RotateKeys != "" {
		qcr.Spec.RotateKeys = opts.RotateKeys
	}
	qConfig.WriteCurrentContextCR(qcr)

	if qcr.Spec.Git.Repository != "" {
		// fetching and applying manifest will be in the operator
		return q.applyCR()
	}
	if version != "" { // no need to fetch manifest root already set by some other way
		fetchAndUpdateCR(qConfig, version)
	}
	// install generated manifests into cluster
	fmt.Println("Installing generated manifests into cluster")
	qcr, err = qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	}

	if err := q.applyConfigToK8s(qcr); err != nil {
	if err := q.applyConfigToK8s(qcr, "install"); err != nil {
		fmt.Println("cannot do kubectl apply on manifests")
		return err
	}

	return q.applyCR()
}

func (q *Qliksense) applyCR() error {
	// install operator cr into cluster
	//get the current context cr
	fmt.Println("Install operator CR into cluster")
	r, err := q.getCurrentCRString()
	if err != nil {
		return err
	}
	if err := qapi.KubectlApply(r); err != nil {
		fmt.Println("cannot do kubectl apply on operator CR")
		return err
	}
	return nil
}
