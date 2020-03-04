package api

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/jinzhu/copier"
)

const (
	pushSecretFileName       = "image-registry-push-secret.yaml"
	pullSecretFileName       = "image-registry-pull-secret.yaml"
	qliksenseContextsDirName = "contexts"
	qliksenseSecretsDirName  = "secrets"
)

// NewQConfig create QliksenseConfig object from file ~/.qliksense/config.yaml
func NewQConfig(qsHome string) *QliksenseConfig {
	configFile := filepath.Join(qsHome, "config.yaml")
	qc := &QliksenseConfig{}

	err := ReadFromFile(qc, configFile)
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

func getUnencryptedCR() {

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
	cr := &QliksenseCR{}
	err := ReadFromFile(cr, crfile)
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
	return filepath.Join(qc.QliksenseHomePath, qliksenseContextsDirName, contextName, "qlik-k8s", version)
}

func (qc *QliksenseConfig) BuildCurrentManifestsRoot(version string) string {
	return qc.BuildRepoPath(version)
}

func (qc *QliksenseConfig) WriteCR(cr *QliksenseCR, contextName string) error {
	crf := qc.getCRFilePath(contextName)
	if crf == "" {
		return errors.New("context name " + contextName + " not found")
	}
	return WriteToFile(cr, crf)
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
		return filepath.Join(qc.QliksenseHomePath, qliksenseContextsDirName, qcr.GetObjectMeta().GetName()), nil
	}
}

func (qc *QliksenseConfig) GetCurrentContextSecretsDir() (string, error) {
	if currentContextDir, err := qc.GetCurrentContextDir(); err != nil {
		return "", err
	} else {
		return filepath.Join(currentContextDir, qliksenseSecretsDirName), nil
	}
}

func (qc *QliksenseConfig) setDockerConfigJsonSecret(filename string, dockerConfigJsonSecret *DockerConfigJsonSecret) error {
	if secretsDir, err := qc.GetCurrentContextSecretsDir(); err != nil {
		return err
	} else if publicKey, _, err := qc.GetCurrentContextEncryptionKeyPair(); err != nil {
		return err
	} else if dockerConfigJsonSecretYaml, err := dockerConfigJsonSecret.ToYaml(publicKey); err != nil {
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
	return qc.getDockerConfigJsonSecret(pushSecretFileName)
}

func (qc *QliksenseConfig) GetPullDockerConfigJsonSecret() (*DockerConfigJsonSecret, error) {
	return qc.getDockerConfigJsonSecret(pullSecretFileName)
}

func (qc *QliksenseConfig) DeletePushDockerConfigJsonSecret() error {
	return qc.deleteDockerConfigJsonSecret(pushSecretFileName)
}

func (qc *QliksenseConfig) DeletePullDockerConfigJsonSecret() error {
	return qc.deleteDockerConfigJsonSecret(pullSecretFileName)
}

func (qc *QliksenseConfig) deleteDockerConfigJsonSecret(name string) error {
	if secretsDir, err := qc.GetCurrentContextSecretsDir(); err != nil {
		return err
	} else {
		return os.Remove(filepath.Join(secretsDir, name))
	}
}

func (qc *QliksenseConfig) getDockerConfigJsonSecret(name string) (*DockerConfigJsonSecret, error) {
	dockerConfigJsonSecret := &DockerConfigJsonSecret{}
	if secretsDir, err := qc.GetCurrentContextSecretsDir(); err != nil {
		return nil, err
	} else if dockerConfigJsonSecretYaml, err := ioutil.ReadFile(filepath.Join(secretsDir, name)); err != nil {
		return nil, err
	} else if _, privateKey, err := qc.GetCurrentContextEncryptionKeyPair(); err != nil {
		return nil, err
	} else if err := dockerConfigJsonSecret.FromYaml(dockerConfigJsonSecretYaml, privateKey); err != nil {
		return nil, err
	}
	return dockerConfigJsonSecret, nil
}

func (qc *QliksenseConfig) getCurrentContextEncryptionKeyPairLocation() (string, error) {
	// Check env var: QLIKSENSE_KEY_LOCATION to determine location to store keypair
	var secretKeyPairLocation string
	if os.Getenv("QLIKSENSE_KEY_LOCATION") != "" {
		LogDebugMessage("Env variable: QLIKSENSE_KEY_LOCATION= %s", os.Getenv("QLIKSENSE_KEY_LOCATION"))
		secretKeyPairLocation = os.Getenv("QLIKSENSE_KEY_LOCATION")
	} else {
		// QLIKSENSE_KEY_LOCATION has not been set, hence storing key pair in default location:
		// /.qliksense/secrets/contexts/<current-context>/secrets/
		if qcr, err := qc.GetCurrentCR(); err != nil {
			return "", err
		} else {
			secretKeyPairLocation = filepath.Join(qc.QliksenseHomePath, qliksenseSecretsDirName, qliksenseContextsDirName, qcr.GetObjectMeta().GetName(), qliksenseSecretsDirName)
		}
	}
	LogDebugMessage("SecretKeyLocation to store key pair: %s", secretKeyPairLocation)
	return secretKeyPairLocation, nil
}

func (qc *QliksenseConfig) GetCurrentContextEncryptionKeyPair() (*rsa.PublicKey, *rsa.PrivateKey, error) {
	secretKeyPairLocation, err := qc.getCurrentContextEncryptionKeyPairLocation()
	if err != nil {
		return nil, nil, err
	}

	publicKeyFilePath := filepath.Join(secretKeyPairLocation, QliksensePublicKey)
	privateKeyFilePath := filepath.Join(secretKeyPairLocation, QliksensePrivateKey)
	// try to create the dir if it doesn't exist
	if !FileExists(publicKeyFilePath) || !FileExists(privateKeyFilePath) {
		LogDebugMessage("Qliksense secretKeyLocation dir does not exist, creating it now: %s", secretKeyPairLocation)
		if err := os.MkdirAll(secretKeyPairLocation, os.ModePerm); err != nil {
			err = fmt.Errorf("Not able to create %s dir: %v", secretKeyPairLocation, err)
			log.Println(err)
			return nil, nil, err
		}
		// generating and storing key-pair
		err1 := GenerateAndStoreSecretKeypair(secretKeyPairLocation)
		if err1 != nil {
			err1 = fmt.Errorf("Not able to generate and store key pair for encryption")
			log.Println(err1)
			return nil, nil, err1
		}
	}

	if publicKeyBytes, err := ReadKeys(publicKeyFilePath); err != nil {
		LogDebugMessage("Not able to read public key")
		return nil, nil, err
	} else if privateKeyBytes, err := ReadKeys(privateKeyFilePath); err != nil {
		LogDebugMessage("Not able to read private key")
		return nil, nil, err
	} else if rsaPublicKey, err := DecodeToPublicKey(publicKeyBytes); err != nil {
		return nil, nil, err
	} else if rsaPrivateKey, err := DecodeToPrivateKey(privateKeyBytes); err != nil {
		return nil, nil, err
	} else {
		return rsaPublicKey, rsaPrivateKey, nil
	}
}

func (cr *QliksenseCR) AddLabelToCr(key, value string) {
	m := cr.GetObjectMeta().GetLabels()
	if m == nil {
		m = make(map[string]string)
	}
	m[key] = value
	cr.GetObjectMeta().SetLabels(m)
}

func (cr *QliksenseCR) GetLabelFromCr(key string) string {
	return cr.GetObjectMeta().GetLabels()[key]
}

func (cr *QliksenseCR) GetString() (string, error) {
	out, err := K8sToYaml(cr)
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

func (cr *QliksenseCR) GetK8sSecretsFolder(qlikSenseHomeDir string) string {
	return filepath.Join(qlikSenseHomeDir, qliksenseContextsDirName, cr.GetName(), qliksenseSecretsDirName)
}
