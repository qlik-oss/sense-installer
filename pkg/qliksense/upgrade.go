package qliksense

import (
	"fmt"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

type upgradeCommandOptions struct {
	AcceptEULA   string
	Namespace    string
	StorageClass string
	MongoDbUri   string
	RotateKeys   string
}

func (q *Qliksense) UpgradeQK8s(opts *InstallCommandOptions) error {

	// step1: get CR
	// step2: run kustomize
	// step3: run kubectl apply

	// fetch the version
	qConfig := qapi.NewQConfig(q.QliksenseHome)

	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	}
	qcr.Spec.RotateKeys = "No"
	qConfig.WriteCurrentContextCR(qcr)
	if err := q.applyConfigToK8s(qcr, "upgrade"); err != nil {
		fmt.Println("cannot do kubectl apply on manifests")
		return err
	}

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
