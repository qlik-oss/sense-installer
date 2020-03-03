package qliksense

import (
	"errors"
	"fmt"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

type InstallCommandOptions struct {
	AcceptEULA   string
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
		qcr.Spec.AddToSecrets("qliksense", "mongoDbUri", opts.MongoDbUri, "")
	}
	if opts.StorageClass != "" {
		qcr.Spec.StorageClassName = opts.StorageClass
	}
	if opts.RotateKeys != "" {
		qcr.Spec.RotateKeys = opts.RotateKeys
	}
	qConfig.WriteCurrentContextCR(qcr)

	//if the docker pull secret exists on disk, install it in the cluster
	//if it doesn't exist on disk, remove it in the cluster
	if dockerConfigJsonSecret, err := qConfig.GetPullDockerConfigJsonSecret(); err == nil {
		if dockerConfigJsonSecretYaml, err := dockerConfigJsonSecret.ToYaml(nil); err != nil {
			return err
		} else if err := qapi.KubectlApply(string(dockerConfigJsonSecretYaml), ""); err != nil {
			return err
		}
	} else {
		deleteDockerConfigJsonSecret := qapi.DockerConfigJsonSecret{
			Name: pullSecretName,
		}
		if deleteDockerConfigJsonSecretYaml, err := deleteDockerConfigJsonSecret.ToYaml(nil); err != nil {
			return err
		} else if err := qapi.KubectlDelete(string(deleteDockerConfigJsonSecretYaml), ""); err != nil {
			qapi.LogDebugMessage("failed deleting %v, error: %v\n", pullSecretName, err)
		}
	}

	//CRD will be installed outside of operator
	//install operator controller into the namespace
	fmt.Println("Installing operator controller")
	if err := qapi.KubectlApply(q.GetOperatorControllerString(), ""); err != nil {
		fmt.Println("cannot do kubectl apply on opeartor controller", err)
		return err
	}

	if qcr.Spec.Git != nil && qcr.Spec.Git.Repository != "" {
		// fetching and applying manifest will be in the operator controller
		return q.applyCR()
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
	if err := qapi.KubectlApply(r, ""); err != nil {
		fmt.Println("cannot do kubectl apply on operator CR")
		return err
	}
	return nil
}
