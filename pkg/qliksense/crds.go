package qliksense

import (
	"fmt"
	"os"
	"path/filepath"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
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
	engineCRD, err := getQliksenseInitCrd(qcr)
	if err != nil {
		return err
	}
	customCrd, err := getCustomCrd(qcr)
	if err != nil {
		return nil
	}

	fmt.Println(engineCRD)
	if customCrd != "" {
		fmt.Println("---")
		fmt.Println(customCrd)
	}

	if opts.All {
		fmt.Println("---")
		fmt.Printf("%s", q.GetOperatorCRDString())
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
	} else if err = qapi.KubectlApply(engineCRD, ""); err != nil {
		return err
	}
	if customCrd, err := getCustomCrd(qcr); err != nil {
		return err
	} else if customCrd != "" {
		if err = qapi.KubectlApply(customCrd, ""); err != nil {
			return err
		}
	}

	if opts.All { // install opeartor crd
		if err := qapi.KubectlApply(q.GetOperatorCRDString(), ""); err != nil {
			fmt.Println("cannot do kubectl apply on opeartor CRD", err)
			return err
		}
	}
	return nil
}

func getQliksenseInitCrd(qcr *qapi.QliksenseCR) (string, error) {
	var repoPath string
	var err error

	if qcr.Spec.GetManifestsRoot() != "" {
		repoPath = qcr.Spec.GetManifestsRoot()
	} else {
		if repoPath, err = DownloadFromGitRepoToTmpDir(defaultConfigRepoGitUrl, "master"); err != nil {
			return "", err
		}
	}

	qInitMsPath := filepath.Join(repoPath, Q_INIT_CRD_PATH)
	if _, err := os.Lstat(qInitMsPath); err != nil {
		// older version of qliksense-init used
		qInitMsPath = filepath.Join(repoPath, "manifests/base/manifests/qliksense-init")
	}
	qInitByte, err := ExecuteKustomizeBuild(qInitMsPath)
	if err != nil {
		fmt.Println("cannot generate crds for qliksense-init", err)
		return "", err
	}
	return string(qInitByte), nil
}

func getCustomCrd(qcr *qapi.QliksenseCR) (string, error) {
	crdPath := qcr.GetCustomCrdsPath()
	if crdPath == "" {
		return "", nil
	}
	qInitByte, err := ExecuteKustomizeBuild(crdPath)
	if err != nil {
		fmt.Println("cannot generate custom crds", err)
		return "", err
	}
	return string(qInitByte), nil
}
