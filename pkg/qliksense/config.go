package qliksense

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/qlik-oss/k-apis/pkg/config"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"

	"github.com/qlik-oss/k-apis/pkg/cr"
	"github.com/qlik-oss/sense-installer/pkg/api"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

const (
	Q_INIT_CRD_PATH   = "manifests/base/crds"
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

	// create patch dependent resources
	fmt.Println("Installing resources used by the kuztomize patch")
	if err := q.createK8sResourceBeforePatch(qcr); err != nil {
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

func (q *Qliksense) configEjson() error {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	if ejsonKeyDir, err := qConfig.GetCurrentContextEjsonKeyDir(); err != nil {
		return err
	} else if err := os.Unsetenv("EJSON_KEY"); err != nil {
		return err
	} else if err := os.Setenv("EJSON_KEYDIR", ejsonKeyDir); err != nil {
		return err
	}
	return nil
}

func (q *Qliksense) applyConfigToK8s(qcr *qapi.QliksenseCR) error {
	if err := q.configEjson(); err != nil {
		return err
	}

	userHomeDir, err := homedir.Dir()
	if err != nil {
		fmt.Printf(`error fetching user's home directory: %v\n`, err)
		return err
	}
	qcr.SetNamespace(qapi.GetKubectlNamespace())
	b, _ := yaml.Marshal(qcr.KApiCr)
	fmt.Printf("%v", string(b))
	// os.Exit(0)
	// generate patches
	cr.GeneratePatches(&qcr.KApiCr, config.KeysActionRestoreOrRotate, path.Join(userHomeDir, ".kube", "config"))
	// apply generated manifests
	profilePath := filepath.Join(qcr.Spec.GetManifestsRoot(), qcr.Spec.GetProfileDir())
	fmt.Printf("Generating manifests for profile: %v\n", profilePath)
	mByte, err := ExecuteKustomizeBuild(profilePath)
	if err != nil {
		fmt.Printf("error generating manifests: %v\n", err)
		return err
	}
	fmt.Println("Applying manifests to the cluster")
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
	oth, err := q.getCurrentCrDependentResourceAsString()
	if err != nil {
		return err
	}
	fmt.Println(r + "\n" + oth)
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
	return string(out), nil

}

func (q *Qliksense) getCurrentCrDependentResourceAsString() (string, error) {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCR(qConfig.Spec.CurrentContext)
	if err != nil {
		fmt.Println("cannot get the context cr", err)
		return "", err
	}
	var crString strings.Builder

	for svcName, v := range qcr.Spec.Secrets {
		hasFile := false
		for _, item := range v {
			if item.ValueFrom != nil && item.ValueFrom.SecretKeyRef != nil {
				hasFile = true
				break
			}
		}
		if hasFile {
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
	crString.WriteString("\n---\n")
	return crString.String(), nil
}

func (q *Qliksense) EditCR(contextName string) error {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	if contextName == "" {
		cr, err := qConfig.GetCurrentCR()
		if err != nil {
			return err
		}
		contextName = cr.GetName()
	}
	crFilePath := qConfig.GetCRFilePath(contextName)
	tempFile, err := ioutil.TempFile("", "*.yaml")
	if err != nil {
		return err
	}
	crContent, err := ioutil.ReadFile(crFilePath)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(tempFile.Name(), crContent, os.ModePerm); err != nil {
		return nil
	}
	cmd := exec.Command(getKubeEditorTool(), tempFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return err
	}
	newCr, err := qapi.GetCRObject(tempFile.Name())
	if err != nil {
		return errors.New("cannot save the cr. Someting wrong in the file format. It is not saved\n" + err.Error())
	}
	oldCr, err := qapi.GetCRObject(crFilePath)

	if oldCr.GetName() != newCr.GetName() {
		return errors.New("cr name cannot be chagned")
	}
	if newCr.Validate() {
		return qConfig.WriteCR(newCr)
	}
	return nil
}

func getKubeEditorTool() string {
	editor := os.Getenv("KUBE_EDITOR")
	if editor == "" {
		editor = "vim"
	}
	return editor
}
