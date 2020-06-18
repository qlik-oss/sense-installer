package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/qlik-oss/k-apis/pkg/config"

	b64 "encoding/base64"

	"github.com/jinzhu/copier"
)

const (
	pushSecretFileName       = "image-registry-push-secret.yaml"
	pullSecretFileName       = "image-registry-pull-secret.yaml"
	qliksenseContextsDirName = "contexts"
	qliksenseSecretsDirName  = "secrets"
	qliksenseEjsonDirName    = "ejson"
	QLIK_GIT_REPO            = "https://github.com/qlik-oss/qliksense-k8s"
)

// NewQConfig create QliksenseConfig object from file ~/.qliksense/config.yaml
func NewQConfig(qsHome string) *QliksenseConfig {
	qc, err := NewQConfigE(qsHome)
	if err != nil {
		fmt.Println("yaml unmarshalling error ", err)
		os.Exit(1)
	}
	return qc
}

func NewQConfigE(qsHome string) (*QliksenseConfig, error) {
	configFile := filepath.Join(qsHome, "config.yaml")
	qc := &QliksenseConfig{}

	err := ReadFromFile(qc, configFile)
	if err != nil {
		return nil, err
	}
	qc.QliksenseHomePath = qsHome
	return qc, nil
}
func NewQConfigEmpty(qsHome string) *QliksenseConfig {
	return &QliksenseConfig{
		QliksenseHomePath: qsHome,
	}
}

// GetCR create a QliksenseCR object for a particular context
// from file ~/.qliksense/contexts/<contx-name>/<contx-name>.yaml
func (qc *QliksenseConfig) GetCR(contextName string) (*QliksenseCR, error) {
	crFilePath := qc.GetCRFilePath(contextName)
	if crFilePath == "" {
		return nil, errors.New("context name " + contextName + " not found")
	}
	return qc.GetAndTransformCrObject(crFilePath)
}

// GetCurrentCR create a QliksenseCR object for current context
func (qc *QliksenseConfig) GetCurrentCR() (*QliksenseCR, error) {
	return qc.GetCR(qc.Spec.CurrentContext)
}

// SetCrLocation sets the CR location for a context. Helpful during test
func (qc *QliksenseConfig) SetCrLocation(contextName, filePath string) (*QliksenseConfig, error) {
	tempQc := &QliksenseConfig{}
	copier.Copy(tempQc, qc)
	found := false
	tempQc.Spec.Contexts = []Context{}
	for _, c := range qc.Spec.Contexts {
		if c.Name == contextName {
			c.CrFile = filePath
			found = true
		}
		tempQc.Spec.Contexts = append(tempQc.Spec.Contexts, []Context{c}...)
	}
	if found {
		return tempQc, nil
	}
	return nil, errors.New("cannot find the context")
}

// GetCRObject create a qliksense CR object from file
func GetCRObject(crfile string) (*QliksenseCR, error) {
	cr := &QliksenseCR{}
	err := ReadFromFile(cr, crfile)
	if err != nil {
		fmt.Println("cannot unmarshal cr ", err)
		return nil, err
	}

	return cr, nil
}

func (qc *QliksenseConfig) GetAndTransformCrObject(crfile string) (*QliksenseCR, error) {
	cr, err := GetCRObject(crfile)
	if err != nil {
		return nil, err
	}
	if cr.Spec.ManifestsRoot != "" && !filepath.IsAbs(cr.Spec.ManifestsRoot) {
		cr.Spec.ManifestsRoot = filepath.Join(qc.QliksenseHomePath, cr.Spec.ManifestsRoot)
	}
	return cr, nil
}

//CreateCRObjectFromString create a QliksenseCR from string content
func CreateCRObjectFromString(crContent string) (*QliksenseCR, error) {
	if crContent == "" {
		return nil, errors.New("empty string cannot qliksensecr")
	}
	cr := &QliksenseCR{}
	err := ReadFromStream(cr, strings.NewReader(crContent))
	if err != nil {
		fmt.Println("cannot unmarshal cr ", err)
		return nil, err
	}
	return cr, nil
}

func (qc *QliksenseConfig) GetCRFilePath(contextName string) string {
	crFilePath := ""
	for _, ctx := range qc.Spec.Contexts {
		if ctx.Name == contextName {
			crFilePath = filepath.Join(qc.QliksenseHomePath, ctx.CrFile)
			break
		}
	}
	return crFilePath
}

func (cr *QliksenseCR) IsRepoExist() bool {
	if cr.Spec.ManifestsRoot == "" {
		return false
	}
	if _, err := os.Lstat(cr.Spec.ManifestsRoot); err != nil {
		return false
	}
	return true
}

func (cr *QliksenseCR) GetFetchUrl() string {
	if cr.Spec.Git == nil || cr.Spec.Git.Repository == "" {
		return QLIK_GIT_REPO
	}
	return cr.Spec.Git.Repository
}

func (cr *QliksenseCR) GetFetchAccessToken(encryptionKey string) string {
	if cr.Spec.Git == nil {
		return ""
	}
	if tok, err := cr.Spec.Git.GetAccessToken(); err != nil {
		fmt.Println(err)
		return ""
	} else if tok == "" {
		return tok
	} else {
		by, _ := b64.StdEncoding.DecodeString(tok)
		res, err := DecryptData(by, encryptionKey)
		if err != nil {
			fmt.Println(err)
			return ""
		}
		return string(res)
	}
}

func (cr *QliksenseCR) SetFetchUrl(url string) {
	if cr.Spec.Git == nil {
		cr.Spec.Git = &config.Repo{}
	}
	cr.Spec.Git.Repository = url
}

func (cr *QliksenseCR) SetFetchAccessToken(token, encryptionKey string) error {
	if cr.Spec.Git == nil {
		cr.Spec.Git = &config.Repo{}
	}
	res, err := EncryptData([]byte(token), encryptionKey)
	if err != nil {
		return err
	}
	cr.Spec.Git.AccessToken = b64.StdEncoding.EncodeToString(res)
	return nil
}

func (cr *QliksenseCR) SetFetchAccessSecretName(sec string) {
	if cr.Spec.Git == nil {
		cr.Spec.Git = &config.Repo{}
	}
	cr.Spec.Git.SecretName = sec
}

//DeleteRepo delete the manifest repo and unset manifestsRoot
func (cr *QliksenseCR) DeleteRepo() error {
	if err := os.RemoveAll(cr.Spec.ManifestsRoot); err != nil {
		return err
	}
	cr.Spec.ManifestsRoot = ""
	return nil
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

func (qc *QliksenseConfig) DeleteRepoForCurrent(version string) error {
	path := qc.BuildRepoPath(version)
	return os.RemoveAll(path)
}

func (qc *QliksenseConfig) BuildRepoPath(version string) string {
	return qc.BuildRepoPathForContext(qc.Spec.CurrentContext, version)
}

func (qc *QliksenseConfig) BuildRepoPathForContext(contextName, version string) string {
	return filepath.Join(qc.GetContextPath(contextName), "qlik-k8s", version)
}

func (qc *QliksenseConfig) BuildCurrentManifestsRoot(version string) string {
	return qc.BuildRepoPath(version)
}

func (qc *QliksenseConfig) WriteCR(cr *QliksenseCR) error {
	crf := qc.GetCRFilePath(cr.GetName())
	if crf == "" {
		return errors.New("context name " + cr.GetName() + " not found")
	}

	return qc.TransformAndWriteCr(cr, crf)
}

//CreateOrWriteCrAndContext create necessary folder structure, update config.yaml and context yaml files
func (qc *QliksenseConfig) CreateOrWriteCrAndContext(cr *QliksenseCR) error {
	if qc.QliksenseHomePath == "" {
		return errors.New("qliksense home is not set")
	}
	crf := qc.GetCRFilePath(cr.GetName())
	if crf == "" {
		// create direcotry structure for context
		cDir := filepath.Join(qc.QliksenseHomePath, "contexts", cr.GetName())
		if err := os.MkdirAll(cDir, os.ModePerm); err != nil {
			return err
		}
		crf = filepath.Join(cDir, cr.GetName()+".yaml")
		ctx := Context{
			Name:   cr.GetName(),
			CrFile: "contexts/" + cr.GetName() + "/" + cr.GetName() + ".yaml", //filepath.Join("contexts", cr.GetName(), cr.GetName()+".yaml"),
		}
		qc.AddToContexts(ctx)

		if err := qc.Write(); err != nil {
			return err
		}
	}

	return qc.TransformAndWriteCr(cr, crf)
}

func (qc *QliksenseConfig) TransformAndWriteCr(cr *QliksenseCR, file string) error {
	if strings.HasPrefix(cr.Spec.ManifestsRoot, qc.QliksenseHomePath) {
		cr.Spec.ManifestsRoot = strings.Replace(cr.Spec.ManifestsRoot, qc.QliksenseHomePath+"/", "", 1)
		cr.Spec.ManifestsRoot = strings.Replace(cr.Spec.ManifestsRoot, qc.QliksenseHomePath+"\\", "", 1)
		cr.Spec.ManifestsRoot = strings.Replace(cr.Spec.ManifestsRoot, "\\", "/", -1)
	}
	if err := WriteToFile(cr, file); err != nil {
		return err
	}
	if cr.Spec.ManifestsRoot != "" {
		cr.Spec.ManifestsRoot = filepath.Join(qc.QliksenseHomePath, cr.Spec.ManifestsRoot)
	}
	return nil
}
func (qc *QliksenseConfig) AddToContexts(ctx Context) error {
	//TODO: additional duplicate check may be added latter
	qc.Spec.Contexts = append(qc.Spec.Contexts, ctx)

	return nil
}
func (qc *QliksenseConfig) WriteCurrentContextCR(cr *QliksenseCR) error {
	return qc.WriteCR(cr)
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
	} else if encryptionKey, err := qc.GetEncryptionKeyForCurrent(); err != nil {
		return err
	} else if dockerConfigJsonSecretYaml, err := dockerConfigJsonSecret.ToYaml(encryptionKey); err != nil {
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
	} else if encryptionKey, err := qc.GetEncryptionKeyForCurrent(); err != nil {
		return nil, err
	} else if err := dockerConfigJsonSecret.FromYaml(dockerConfigJsonSecretYaml, encryptionKey); err != nil {
		return nil, err
	}
	return dockerConfigJsonSecret, nil
}

func (qc *QliksenseConfig) getCurrentContextEncryptionKeyPairLocation() (string, error) {

	if qcr, err := qc.GetCurrentCR(); err != nil {
		return "", err
	} else {
		return qc.getContextEncryptionKeyLocation(qcr.GetName())
	}
}

func (qc *QliksenseConfig) getContextEncryptionKeyLocation(contextName string) (string, error) {
	// Check env var: QLIKSENSE_KEY_LOCATION to determine location to store keypair
	var secretKeyPairLocation string
	if os.Getenv("QLIKSENSE_KEY_LOCATION") != "" {
		LogDebugMessage("Env variable: QLIKSENSE_KEY_LOCATION= %s", os.Getenv("QLIKSENSE_KEY_LOCATION"))
		secretKeyPairLocation = os.Getenv("QLIKSENSE_KEY_LOCATION")
	} else {
		// QLIKSENSE_KEY_LOCATION has not been set, hence storing key pair in default location:
		// /.qliksense/secrets/contexts/<current-context>/secrets/
		secretKeyPairLocation = filepath.Join(qc.QliksenseHomePath, qliksenseSecretsDirName, qliksenseContextsDirName, contextName, qliksenseSecretsDirName)
	}

	return secretKeyPairLocation, os.MkdirAll(secretKeyPairLocation, os.ModePerm)
}

func (qc *QliksenseConfig) GetCurrentContextEjsonKeyDir() (string, error) {
	if qcr, err := qc.GetCurrentCR(); err != nil {
		return "", err
	} else {
		ejsonKeyDir := filepath.Join(qc.QliksenseHomePath, qliksenseSecretsDirName, qliksenseContextsDirName, qcr.GetObjectMeta().GetName(), qliksenseEjsonDirName)
		if err := os.MkdirAll(ejsonKeyDir, os.ModePerm); err != nil {
			return "", err
		}
		return ejsonKeyDir, nil
	}
}

func (qc *QliksenseConfig) GetEncryptionKeyForCurrent() (string, error) {
	if qcr, err := qc.GetCurrentCR(); err != nil {
		return "", err
	} else {
		return qc.GetEncryptionKeyFor(qcr.GetName())
	}
}

func (qc *QliksenseConfig) GetEncryptionKeyFor(contextName string) (string, error) {
	secretKeyLocation, err := qc.getContextEncryptionKeyLocation(contextName)
	if err != nil {
		return "", err
	}
	key, err := LoadSecretKey(secretKeyLocation)
	if key != "" {
		return key, nil
	}
	fmt.Println("Generating new encryption key for the context: " + contextName)
	return GenerateAndStoreSecretKey(secretKeyLocation)
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

func (cr *QliksenseCR) GetK8sSecretsFolder(qlikSenseHomeDir string) string {
	return filepath.Join(qlikSenseHomeDir, qliksenseContextsDirName, cr.GetName(), qliksenseSecretsDirName)
}

func (cr *QliksenseCR) IsEULA() bool {
	for k, nvs := range cr.Spec.Configs {
		if k == "qliksense" {
			for _, nv := range nvs {
				if nv.Name == "acceptEULA" {
					return nv.Value == "yes"
				}
			}
		}
	}
	return false
}

func (cr *QliksenseCR) SetEULA(value string) {
	cr.Spec.AddToConfigs("qliksense", "acceptEULA", value)
}

// GetCustomCrdsPath get crds path if exist in the profile dir
func (cr *QliksenseCR) GetCustomCrdsPath() string {
	if cr.Spec.ManifestsRoot == "" || cr.Spec.Profile == "" {
		return ""
	}
	crdsPath := filepath.Join(cr.Spec.GetManifestsRoot(), "manifests", cr.Spec.Profile, "crds")
	if _, err := os.Lstat(crdsPath); err != nil {
		return ""
	}
	return crdsPath
}

// GetDecryptedCr it decrypts all the encrypted value and return a new CR
func (qc *QliksenseConfig) GetDecryptedCr(cr *QliksenseCR) (*QliksenseCR, error) {
	newCr := &QliksenseCR{}
	copier.Copy(newCr, cr)
	encryptionKey, err := qc.GetEncryptionKeyFor(cr.GetName())
	if err != nil {
		return nil, err
	}
	finalSecrets := map[string]config.NameValues{}
	for k, nvs := range newCr.Spec.Secrets {
		newNvs := config.NameValues{}
		for _, nv := range nvs {
			if nv.Value != "" {
				b, err := b64.StdEncoding.DecodeString(strings.TrimSpace(nv.Value))
				if err != nil {
					return nil, err
				}
				db, err := DecryptData(b, encryptionKey)
				if err != nil {
					return nil, err
				}
				newNvs = append(newNvs, config.NameValue{
					Name:  nv.Name,
					Value: string(db),
				})
			}
		}
		finalSecrets[k] = newNvs
	}
	newCr.Spec.Secrets = finalSecrets

	if newCr.Spec.Git != nil && newCr.Spec.Git.AccessToken != "" {
		decData := cr.GetFetchAccessToken(encryptionKey)
		newCr.Spec.Git.AccessToken = decData
	}
	return newCr, nil
}

//Validate validate CR
func (cr *QliksenseCR) Validate() bool {
	return true
}

//CreateContextDirs create context dir structure ~/.qliksense/contexts/contextName
func (qc *QliksenseConfig) CreateContextDirs(contextName string) error {
	return os.MkdirAll(qc.GetContextPath(contextName), os.ModePerm)
}

func (qc *QliksenseConfig) GetContextPath(contextName string) string {
	return filepath.Join(qc.QliksenseHomePath, qliksenseContextsDirName, contextName)
}

//BuildCrFileAbsolutePath build absolute path for a cr ie. ~/.qliksense/contexts/qlik-defautl/qlik-default.yaml
func (qc *QliksenseConfig) BuildCrFileAbsolutePath(contextName string) string {
	return filepath.Join(qc.GetContextPath(contextName), contextName+".yaml")
}

//BuildCrFilePath build cr file path i.e. contexts/qlik-default/qlik-default.yaml
func (qc *QliksenseConfig) BuildCrFilePath(contextName string) string {
	return filepath.Join(qc.GetContextPath(contextName), contextName+".yaml")
}

//AddToContexts add the context into qc.Spec.Contexts
func (qc *QliksenseConfig) AddToContextsRaw(crName, crFile string) {
	qc.Spec.Contexts = append(qc.Spec.Contexts, []Context{
		{CrFile: crFile,
			Name: crName},
	}...)
}

//SetCurrentContextName set the qc.Spec.CurrentContext
func (qc *QliksenseConfig) SetCurrentContextName(name string) {
	qc.Spec.CurrentContext = name
}

//Write write QliksenseConfig into config.yaml
func (qc *QliksenseConfig) Write() error {
	return WriteToFile(qc, filepath.Join(qc.QliksenseHomePath, "config.yaml"))
}
