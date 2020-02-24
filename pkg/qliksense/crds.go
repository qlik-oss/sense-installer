package qliksense

import (
	"errors"
	"fmt"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"path/filepath"
)

type CrdCommandOptions struct {
	All bool
}

func (q *Qliksense) ViewCrds(opts *CrdCommandOptions) error {
	//io.WriteString(os.Stdout, q.GetCRDString())
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	}
	if engineCRD, err := getQliksenseInitCrd(qcr); err != nil {
		return err
	} else if opts.All {
		fmt.Printf("%s\n%s", q.GetOperatorCRDString(), engineCRD)
	} else {
		fmt.Printf("%s", engineCRD)
	}
	return nil
}

func (q *Qliksense) InstallCrds(opts *CrdCommandOptions) error {
	// install qliksense-init crd
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	}

	if engineCRD, err := getQliksenseInitCrd(qcr); err != nil {
		return err
	} else if err = qapi.KubectlApply(engineCRD, qcr.Spec.NameSpace); err != nil {
		return err
	}
	if opts.All { // install opeartor crd
		if err := qapi.KubectlApply(q.GetOperatorCRDString(), qcr.Spec.NameSpace); err != nil {
			fmt.Println("cannot do kubectl apply on opeartor CRD", err)
			return err
		}
	}
	return nil
}

func getQliksenseInitCrd(qcr *qapi.QliksenseCR) (string, error) {

	if qcr.Spec.GetManifestsRoot() == "" {
		return "", errors.New("Cannot find manifests root. Please use `qliksense fetch <version>`")
	}

	qInitMsPath := filepath.Join(qcr.Spec.GetManifestsRoot(), Q_INIT_CRD_PATH)

	qInitByte, err := executeKustomizeBuild(qInitMsPath)
	if err != nil {
		fmt.Println("cannot generate crds for qliksense-init", err)
		return "", err
	}
	return string(qInitByte), nil
}
