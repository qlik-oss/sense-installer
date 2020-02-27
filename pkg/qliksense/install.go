package qliksense

import (
	"errors"
	"fmt"
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

	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	}

	if opts.AcceptEULA != "" {
		qcr.Spec.AddToConfigs("qliksense", "acceptEULA", opts.AcceptEULA)
	}
	if opts.MongoDbUri != "" {
		qcr.Spec.AddToSecrets("qliksense", "mongoDbUri", opts.MongoDbUri, "", false)
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

	//CRD will be installed outside of operator
	//install operator controller into the namespace
	fmt.Println("Installing operator controller")
	if err := qapi.KubectlApply(q.GetOperatorControllerString(), qcr.Spec.NameSpace); err != nil {
		fmt.Println("cannot do kubectl apply on opeartor controller", err)
		return err
	}

	if qcr.Spec.Git.Repository != "" {
		// fetching and applying manifest will be in the operator controller
		return q.applyCR(qcr.Spec.NameSpace)
	}
	if version != "" { // no need to fetch manifest root already set by some other way
		if err := fetchAndUpdateCR(qConfig, version); err != nil {
			return err
		}
	}

	qcr, err = qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	} else if qcr.Spec.GetManifestsRoot() == "" {
		return errors.New("cannot get the manifest root. Use qliksense fetch <version> or qliksense set manifestsRoot")
	}

	// install generated manifests into cluster
	fmt.Println("Installing generated manifests into cluster")
	if err := q.applyConfigToK8s(qcr); err != nil {
		fmt.Println("cannot do kubectl apply on manifests")
		return err
	}

	return q.applyCR(qcr.Spec.NameSpace)
}

func (q *Qliksense) applyCR(ns string) error {
	// install operator cr into cluster
	//get the current context cr
	fmt.Println("Install operator CR into cluster")
	r, err := q.getCurrentCRString()
	if err != nil {
		return err
	}
	if err := qapi.KubectlApply(r, ns); err != nil {
		fmt.Println("cannot do kubectl apply on operator CR")
		return err
	}
	return nil
}
