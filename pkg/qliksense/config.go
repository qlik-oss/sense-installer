package qliksense

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"

	"github.com/qlik-oss/k-apis/pkg/cr"
	"github.com/qlik-oss/sense-installer/pkg/api"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

const (
	Q_INIT_CRD_PATH   = "manifests/base/manifests/qliksense-init"
	agreementTempalte = `
	Please read the agreement at https://www.qlik.com/us/legal/license-terms
	Accept the end user license agreement by providing acceptEULA=yes
	`
)

func (q *Qliksense) ConfigApplyQK8s() error {

	//get the current context cr
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	}
	// check if acceptEULA is yes or not
	if !qcr.IsEULA() {
		return errors.New(agreementTempalte + "\nPlease do $ qliksense config set-configs qliksense.acceptEULA=yes\n")
	}

	// create patch dependent resoruces
	fmt.Println("Installing resoruces used kuztomize patch")
	if err := q.createK8sResoruceBeforePatch(qcr); err != nil {
		return err
	}

	if qcr.Spec.Git.Repository != "" {
		// fetching and applying manifest will be in the operator controller
		if dcr, err := qConfig.GetDecryptedCr(qcr); err != nil {
			return err
		} else {
			return q.applyCR(dcr)
		}
	}
	if dcr, err := qConfig.GetDecryptedCr(qcr); err != nil {
		return err
	} else {
		return q.applyConfigToK8s(dcr)
	}
}

func (q *Qliksense) applyConfigToK8s(qcr *qapi.QliksenseCR) error {
	if qcr.Spec.RotateKeys != "None" {
		if err := os.Unsetenv("EJSON_KEY"); err != nil {
			fmt.Printf("error unsetting EJSON_KEY environment variable: %v\n", err)
			return err
		}
		if err := os.Setenv("EJSON_KEYDIR", q.QliksenseEjsonKeyDir); err != nil {
			fmt.Printf("error setting EJSON_KEYDIR environment variable: %v\n", err)
			return err
		}
	}
	userHomeDir, err := homedir.Dir()
	if err != nil {
		fmt.Printf(`error fetching user's home directory: %v\n`, err)
		return err
	}
	fmt.Println("Manifests root: " + qcr.Spec.GetManifestsRoot())
	qcr.SetNamespace(qapi.GetKubectlNamespace())
	// generate patches
	cr.GeneratePatches(&qcr.KApiCr, path.Join(userHomeDir, ".kube", "config"))
	// apply generated manifests
	profilePath := filepath.Join(qcr.Spec.GetManifestsRoot(), qcr.Spec.GetProfileDir())
	mByte, err := executeKustomizeBuild(profilePath)
	if err != nil {
		fmt.Println("cannot generate manifests for "+profilePath, err)
		return err
	}
	if err = qapi.KubectlApply(string(mByte), qcr.GetNamespace()); err != nil {
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
	return q.getCRString(qConfig.Spec.CurrentContext)
}

func (q *Qliksense) getCRString(contextName string) (string, error) {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCR(contextName)
	if err != nil {
		fmt.Println("cannot get the context cr", err)
		return "", err
	}
	out, err := qapi.K8sToYaml(qcr)
	if err != nil {
		fmt.Println("cannot unmarshal cr ", err)
		return "", err
	}
	var crString strings.Builder
	crString.Write(out)

	for svcName, v := range qcr.Spec.Secrets {
		for _, item := range v {
			if item.ValueFrom != nil && item.ValueFrom.SecretKeyRef != nil {
				secretFilePath := filepath.Join(q.QliksenseHome, QliksenseContextsDir, qcr.GetName(), QliksenseSecretsDir, svcName+".yaml")

				if api.FileExists(secretFilePath) {
					secretFile, err := ioutil.ReadFile(secretFilePath)
					if err != nil {
						return "", err
					}
					crString.WriteString("\n---\n")
					crString.Write(secretFile)
				}
			}
		}
	}
	return crString.String(), nil
}
