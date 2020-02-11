package qliksense

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/mitchellh/go-homedir"

	"github.com/qlik-oss/k-apis/pkg/cr"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"gopkg.in/yaml.v2"
)

const (
	Q_INIT_CRD_PATH = "manifests/base/manifests/qliksense-init"
)

func (q *Qliksense) ConfigApplyQK8s() error {

	//get the current context cr
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	}
	return q.applyConfigToK8s(qcr)
}

func (q *Qliksense) applyConfigToK8s(qcr *qapi.QliksenseCR) error {
	// apply qliksense-init crd first
	mroot := qcr.Spec.GetManifestsRoot()
	qInitMsPath := filepath.Join(mroot, Q_INIT_CRD_PATH)

	if err := os.Unsetenv("EJSON_KEY"); err != nil {
		fmt.Printf("error unsetting EJSON_KEY environment variable: %v\n", err)
		return err
	}

	if err := os.Setenv("EJSON_KEYDIR", q.QliksenseEjsonKeyDir); err != nil {
		fmt.Printf("error setting EJSON_KEYDIR environment variable: %v\n", err)
		return err
	}

	qInitByte, err := executeKustomizeBuild(qInitMsPath)
	if err != nil {
		fmt.Println("cannot generate crds for qliksense-init", err)
		return err
	}
	if err = qapi.KubectlApply(string(qInitByte)); err != nil {
		return err
	}

	userHomeDir, err := homedir.Dir()
	if err != nil {
		fmt.Printf(`error fetching user's home directory: %v\n`, err)
		return err
	}

	// generate patches
	cr.GeneratePatches(qcr.Spec, path.Join(userHomeDir, ".kube", "config"))
	// apply generated manifests
	profilePath := filepath.Join(qcr.Spec.ManifestsRoot, qcr.Spec.Profile)
	mByte, err := executeKustomizeBuild(profilePath)
	if err != nil {
		fmt.Println("cannot generate manifests for "+profilePath, err)
		return err
	}
	if err = qapi.KubectlApply(string(mByte)); err != nil {
		return err
	}

	return nil
}

func (q *Qliksense) ConfigViewCR() error {

	//get the current context cr
	r, err := q.getCurrentCRString()
	if err != nil {
		return err
	}
	fmt.Println(r)
	return nil
}

func (q *Qliksense) getCurrentCRString() (string, error) {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return "", err
	}
	out, err := yaml.Marshal(qcr)
	if err != nil {
		fmt.Println("cannot unmarshal cr ", err)
		return "", err
	}
	return string(out), nil
}
