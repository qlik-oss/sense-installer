package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jinzhu/copier"
	"gopkg.in/yaml.v2"
)

const (
	pushSecretFileName = "image-registry-push-secret.yaml"
	pullSecretFileName = "image-registry-pull-secret.yaml"
)

// NewQConfig create QliksenseConfig object from file ~/.qliksense/config.yaml
func NewQConfig(qsHome string) *QliksenseConfig {
	configFile := filepath.Join(qsHome, "config.yaml")
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("Cannot read config file from: "+configFile, err)
		os.Exit(1)
	}
	qc := &QliksenseConfig{}
	err = yaml.Unmarshal(data, qc)
	if err != nil {
		fmt.Println("yaml unmarshalling error ", err)
		os.Exit(1)
	}
	qc.QliksenseHomePath = qsHome
	return qc
}

// GetCR create a QliksenseCR object for a particular context
// from file ~/.qliksense/contexts/<contx-name>/<contx-name>.yaml
func (qc *QliksenseConfig) GetCR(contextName string) (*QliksenseCR, error) {
	crFilePath := qc.getCRFilePath(contextName)
	if crFilePath == "" {
		return nil, errors.New("context name " + contextName + " not found")
	}
	return getCRObject(crFilePath)
}

// GetCurrentCR create a QliksenseCR object for current context
func (qc *QliksenseConfig) GetCurrentCR() (*QliksenseCR, error) {
	return qc.GetCR(qc.Spec.CurrentContext)
}

// SetCrLocation sets the CR location for a context. Helpful during test
func (qc *QliksenseConfig) SetCrLocation(contextName, filepath string) (*QliksenseConfig, error) {
	tempQc := &QliksenseConfig{}
	copier.Copy(tempQc, qc)
	found := false
	tempQc.Spec.Contexts = []Context{}
	for _, c := range qc.Spec.Contexts {
		if c.Name == contextName {
			c.CrFile = filepath
			found = true
		}
		tempQc.Spec.Contexts = append(tempQc.Spec.Contexts, []Context{c}...)
	}
	if found {
		return tempQc, nil
	}
	return nil, errors.New("cannot find the context")
}

func getCRObject(crfile string) (*QliksenseCR, error) {
	data, err := ioutil.ReadFile(crfile)
	if err != nil {
		fmt.Println("Cannot read config file from: "+crfile, err)
		return nil, err
	}
	cr := &QliksenseCR{}
	err = yaml.Unmarshal(data, cr)
	if err != nil {
		fmt.Println("cannot unmarshal cr ", err)
		return nil, err
	}
	return cr, nil
}

func (qc *QliksenseConfig) getCRFilePath(contextName string) string {
	crFilePath := ""
	for _, ctx := range qc.Spec.Contexts {
		if ctx.Name == contextName {
			crFilePath = ctx.CrFile
			break
		}
	}
	return crFilePath
}
func (qc *QliksenseConfig) IsRepoExist(contextName, version string) bool {
	if _, err := os.Lstat(qc.BuildRepoPathForContext(contextName, version)); err != nil {
		return false
	}
	return true
}

func (qc *QliksenseConfig) IsRepoExistForCurrent(version string) bool {
	if _, err := os.Lstat(qc.BuildRepoPath(version)); err != nil {
		return false
	}
	return true
}

func (qc *QliksenseConfig) BuildRepoPath(version string) string {
	return qc.BuildRepoPathForContext(qc.Spec.CurrentContext, version)
}

func (qc *QliksenseConfig) BuildRepoPathForContext(contextName, version string) string {
	return filepath.Join(qc.QliksenseHomePath, "contexts", contextName, "qlik-k8s", version)
}

func (qc *QliksenseConfig) BuildCurrentManifestsRoot(version string) string {
	return qc.BuildRepoPath(version)
}

func (qc *QliksenseConfig) WriteCR(cr *QliksenseCR, contextName string) error {
	crf := qc.getCRFilePath(contextName)
	if crf == "" {
		return errors.New("context name " + contextName + " not found")
	}
	by, err := yaml.Marshal(cr)
	if err != nil {
		fmt.Println("cannot marshal cr ", err)
		return err
	}
	ioutil.WriteFile(crf, by, 0644)
	return nil
}

func (qc *QliksenseConfig) WriteCurrentContextCR(cr *QliksenseCR) error {
	return qc.WriteCR(cr, qc.Spec.CurrentContext)
}

func (qc *QliksenseConfig) IsContextExist(ctxName string) bool {
	for _, ct := range qc.Spec.Contexts {
		if ct.Name == ctxName {
			return true
		}
	}
	return false
}

func (qc *QliksenseConfig) GetCurrentContextDir() (string, error) {
	if qcr, err := qc.GetCurrentCR(); err != nil {
		return "", err
	} else {
		return filepath.Join(qc.QliksenseHomePath, "contexts", qcr.Metadata.Name), nil
	}
}

func (qc *QliksenseConfig) GetCurrentContextSecretsDir() (string, error) {
	if currentContextDir, err := qc.GetCurrentContextDir(); err != nil {
		return "", err
	} else {
		return filepath.Join(currentContextDir, "secrets"), nil
	}
}

func (qc *QliksenseConfig) setDockerConfigJsonSecret(filename string, dockerConfigJsonSecret *DockerConfigJsonSecret) error {
	if secretsDir, err := qc.GetCurrentContextSecretsDir(); err != nil {
		return err
	} else if dockerConfigJsonSecretYaml, err := dockerConfigJsonSecret.ToYaml(); err != nil {
		return err
	} else if err := os.MkdirAll(secretsDir, os.ModePerm); err != nil {
		return err
	} else {
		return ioutil.WriteFile(filepath.Join(secretsDir, filename), dockerConfigJsonSecretYaml, os.ModePerm)
	}
}

func (qc *QliksenseConfig) SetPushDockerConfigJsonSecret(dockerConfigJsonSecret *DockerConfigJsonSecret) error {
	return qc.setDockerConfigJsonSecret(pushSecretFileName, dockerConfigJsonSecret)
}

func (qc *QliksenseConfig) SetPullDockerConfigJsonSecret(dockerConfigJsonSecret *DockerConfigJsonSecret) error {
	return qc.setDockerConfigJsonSecret(pullSecretFileName, dockerConfigJsonSecret)
}

func (qc *QliksenseConfig) GetPushDockerConfigJsonSecret() (*DockerConfigJsonSecret, error) {
	return qc.GetDockerConfigJsonSecret(pushSecretFileName)
}

func (qc *QliksenseConfig) GetDockerConfigJsonSecret(name string) (*DockerConfigJsonSecret, error) {
	dockerConfigJsonSecret := &DockerConfigJsonSecret{}
	if secretsDir, err := qc.GetCurrentContextSecretsDir(); err != nil {
		return nil, err
	} else if dockerConfigJsonSecretYaml, err := ioutil.ReadFile(filepath.Join(secretsDir, name)); err != nil {
		return nil, err
	} else if err := dockerConfigJsonSecret.FromYaml(dockerConfigJsonSecretYaml); err != nil {
		return nil, err
	}
	return dockerConfigJsonSecret, nil
}

func (cr *QliksenseCR) AddLabelToCr(key, value string) {
	if cr.Metadata.Labels == nil {
		cr.Metadata.Labels = make(map[string]string)
	}
	cr.Metadata.Labels[key] = value
}

func (cr *QliksenseCR) GetLabelFromCr(key string) string {
	val := ""
	if cr.Metadata.Labels != nil {
		val = cr.Metadata.Labels[key]
	}
	return val
}

func (cr *QliksenseCR) GetString() (string, error) {
	out, err := yaml.Marshal(cr)
	if err != nil {
		fmt.Println("cannot unmarshal cr ", err)
		return "", err
	}
	return string(out), nil
}

func (cr *QliksenseCR) GetImageRegistry() string {
	for _, nameValues := range cr.Spec.Configs {
		for _, nameValue := range nameValues {
			if nameValue.Name == "imageRegistry" {
				return nameValue.Value
			}
		}
	}
	return ""
}
