package qliksense

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	b64 "encoding/base64"

	ansi "github.com/mattn/go-colorable"
	"github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/ttacon/chalk"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Below are some constants to support qliksense context setup
	QliksenseConfigHome        = "/.qliksense"
	QliksenseConfigContextHome = "/.qliksense/contexts"
	QliksenseConfigFile        = "config.yaml"
	QliksenseContextsDir       = "contexts"
	DefaultQliksenseContext    = "qlik-default"
	MaxContextNameLength       = 17
	QliksenseSecretsDir        = "secrets"

	imageRegistryConfigKey = "imageRegistry"
	pullSecretName         = "artifactory-docker-secret"
)

// SetSecrets - set-secrets <key>=<value> commands
func (q *Qliksense) SetSecrets(args []string, isSecretSet bool) error {
	qConfig := api.NewQConfig(q.QliksenseHome)
	qliksenseCR, qliksenseContextsFile, err := retrieveCurrentContextInfo(q)
	if err != nil {
		return err
	}

	// Metadata name in qliksense CR is the name of the current context
	api.LogDebugMessage("Current context: %s", qliksenseCR.Metadata.Name)
	rsaPublicKey, _, err := qConfig.GetCurrentContextEncryptionKeyPair()
	if err != nil {
		return err
	}
	resultArgs, err := api.ProcessConfigArgs(args)
	if err != nil {
		return err
	}
	for _, ra := range resultArgs {
		api.LogDebugMessage("value args to be encrypted: %s", ra.Value)
		if err := q.processSecret(ra, rsaPublicKey, &qliksenseCR, isSecretSet); err != nil {
			return err
		}
	}
	// write modified content into context-yaml
	api.WriteToFile(&qliksenseCR, qliksenseContextsFile)

	return nil
}

func (q *Qliksense) processSecret(ra *api.ServiceKeyValue, rsaPublicKey *rsa.PublicKey, qliksenseCR *api.QliksenseCR, isSecretSet bool) error {
	// encrypt value with RSA key pair
	valueBytes := []byte(ra.Value)
	cipherText, e2 := api.Encrypt(valueBytes, rsaPublicKey)
	if e2 != nil {
		return e2
	}
	base64EncodedSecret := b64.StdEncoding.EncodeToString(cipherText)
	api.LogDebugMessage("base64 encoded secret: %s\n", base64EncodedSecret)
	secretName := ""
	if isSecretSet {
		secretFolder := filepath.Join(q.QliksenseHome, QliksenseContextsDir, qliksenseCR.Metadata.Name, QliksenseSecretsDir)
		secretFileName := filepath.Join(secretFolder, ra.SvcName+".yaml")

		secretName = fmt.Sprintf("%s-%s-%s", qliksenseCR.Metadata.Name, ra.SvcName, "sense_installer")
		api.LogDebugMessage("Constructed secret name: %s", secretName)

		k8sSecret := v1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: secretName,
			},
			Type: v1.SecretTypeOpaque,
		}

		if !api.DirExists(secretFolder) {
			if err := os.MkdirAll(secretFolder, os.ModePerm); err != nil {
				err = fmt.Errorf("Not able to create %s dir: %v", secretFolder, err)
				log.Println(err)
				return err
			}
		}
		// if read from file errors out, we can ignore it here
		_ = api.ReadFromFile(&k8sSecret, secretFileName)
		if k8sSecret.Data == nil {
			k8sSecret.Data = map[string][]byte{}
		}
		k8sSecret.Data[ra.Key] = []byte(base64EncodedSecret)

		// Write secret to file
		k8sSecretBytes, err := api.K8sSecretToYaml(k8sSecret)
		if err != nil {
			api.LogDebugMessage("Error while converting K8s secret to yaml")
			return err
		}
		if err = ioutil.WriteFile(secretFileName, k8sSecretBytes, os.ModePerm); err != nil {
			api.LogDebugMessage("Error while writing K8s secret to file")
			return err
		}
		api.LogDebugMessage("Created a Kubernetes secret")

		// Prepare args to update CR in the next step
		base64EncodedSecret = ""
	}

	// write into CR the keyref of the secret
	qliksenseCR.Spec.AddToSecrets(ra.SvcName, ra.Key, base64EncodedSecret, secretName)
	return nil
}

// SetConfigs - set-configs <key>=<value> commands
func (q *Qliksense) SetConfigs(args []string) error {
	// retieve current context from config.yaml
	qliksenseCR, qliksenseContextsFile, err := retrieveCurrentContextInfo(q)
	if err != nil {
		return err
	}

	resultArgs, err := api.ProcessConfigArgs(args)
	if err != nil {
		return err
	}
	for _, ra := range resultArgs {
		qliksenseCR.Spec.AddToConfigs(ra.SvcName, ra.Key, ra.Value)
	}
	// write modified content into context.yaml
	api.WriteToFile(&qliksenseCR, qliksenseContextsFile)

	return nil
}

func retrieveCurrentContextInfo(q *Qliksense) (api.QliksenseCR, string, error) {
	var qliksenseConfig api.QliksenseConfig
	qliksenseConfigFile := filepath.Join(q.QliksenseHome, QliksenseConfigFile)

	if err := api.ReadFromFile(&qliksenseConfig, qliksenseConfigFile); err != nil {
		log.Println(err)
		return api.QliksenseCR{}, "", err
	}
	currentContext := qliksenseConfig.Spec.CurrentContext
	api.LogDebugMessage("Current-context from config.yaml: %s", currentContext)
	if currentContext == "" {
		// current-context is empty
		err := fmt.Errorf(`Please run the "qliksense config set-context <context-name>" first before viewing the current context info`)
		log.Println(err)
		return api.QliksenseCR{}, "", err
	}
	// read the context.yaml file
	var qliksenseCR api.QliksenseCR
	if currentContext == "" {
		// current-context is empty
		err := fmt.Errorf(`Please run the "qliksense config set-context <context-name>" first before viewing the current context info`)
		log.Println(err)
		return api.QliksenseCR{}, "", err
	}
	qliksenseContextsFile := filepath.Join(q.QliksenseHome, QliksenseContextsDir, currentContext, currentContext+".yaml")
	if !api.FileExists(qliksenseContextsFile) {
		err := fmt.Errorf("Context file does not exist.\nPlease try re-running `qliksense config set-context <context-name>` and then `qliksense config view` again")
		log.Println(err)
		return api.QliksenseCR{}, "", err
	}
	if err := api.ReadFromFile(&qliksenseCR, qliksenseContextsFile); err != nil {
		log.Println(err)
		return api.QliksenseCR{}, "", err
	}

	api.LogDebugMessage("Read context file: %s, Read QliksenseCR: %v", qliksenseContextsFile, qliksenseCR)
	return qliksenseCR, qliksenseContextsFile, nil
}

// SetOtherConfigs - set profile/namespace/storageclassname/git.repository/manifestRoot commands
func (q *Qliksense) SetOtherConfigs(args []string) error {
	// retieve current context from config.yaml
	qliksenseCR, qliksenseContextsFile, err := retrieveCurrentContextInfo(q)
	if err != nil {
		return err
	}

	// modify appropriate fields
	if len(args) == 0 {
		err := fmt.Errorf("No args were provided. Please provide args to configure the current context")
		log.Println(err)
		return err
	}

	for _, arg := range args {
		argsString := strings.Split(arg, "=")
		switch argsString[0] {
		case "profile":
			qliksenseCR.Spec.Profile = argsString[1]
			api.LogDebugMessage("Current profile after modification: %s ", qliksenseCR.Spec.Profile)
		case "namespace":
			qliksenseCR.Spec.NameSpace = argsString[1]
			api.LogDebugMessage("Current namespace after modification: %s ", qliksenseCR.Spec.NameSpace)
		case "git.repository":
			qliksenseCR.Spec.Git.Repository = argsString[1]
			api.LogDebugMessage("Current git repository after modification: %s ", qliksenseCR.Spec.Git.Repository)
		case "storageClassName":
			qliksenseCR.Spec.StorageClassName = argsString[1]
			api.LogDebugMessage("Current StorageClassName after modification: %s ", qliksenseCR.Spec.StorageClassName)
		case "manifestsRoot":
			qliksenseCR.Spec.ManifestsRoot = argsString[1]
		case "rotateKeys":
			rotateKeys, err := validateInput(argsString[1])
			if err != nil {
				return err
			}
			qliksenseCR.Spec.RotateKeys = rotateKeys
			api.LogDebugMessage("Current rotateKeys after modification: %s ", qliksenseCR.Spec.RotateKeys)
		default:
			err := fmt.Errorf("Please enter one of: profile, namespace, storageClassName,rotateKeys, manifestRoot or git.repository arguments to configure the current context")
			log.Println(err)
			return err
		}
	}
	// write modified content into context.yaml
	api.WriteToFile(&qliksenseCR, qliksenseContextsFile)

	return nil
}

// SetContextConfig - set the context for qliksense kubernetes resources to live in
func (q *Qliksense) SetContextConfig(args []string) error {
	if len(args) == 1 {
		err := q.SetUpQliksenseContext(args[0], false)
		if err != nil {
			return err
		}
	} else {
		err := fmt.Errorf("Please provide a name to configure the context with")
		log.Println(err)
		return err
	}
	return nil
}

func (q *Qliksense) ListContextConfigs() error {
	qliksenseConfigFile := filepath.Join(q.QliksenseHome, QliksenseConfigFile)
	var qliksenseConfig api.QliksenseConfig
	if err := api.ReadFromFile(&qliksenseConfig, qliksenseConfigFile); err != nil {
		log.Println(err)
		return err
	}
	out := ansi.NewColorableStdout()
	w := tabwriter.NewWriter(out, 5, 8, 0, '\t', 0)
	fmt.Fprintln(w, chalk.Underline.TextStyle("Context Name"), "\t", chalk.Underline.TextStyle("CR File Location"))
	w.Flush()
	if len(qliksenseConfig.Spec.Contexts) > 0 {
		for _, cont := range qliksenseConfig.Spec.Contexts {
			fmt.Fprintln(w, cont.Name, "\t", cont.CrFile, "\t")
		}
		w.Flush()
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, chalk.Bold.TextStyle("Current Context : "), qliksenseConfig.Spec.CurrentContext)
	} else {
		fmt.Fprintln(out, "No Contexts Available")
	}
	return nil
}

// SetUpQliksenseDefaultContext - to setup dir structure for default qliksense context
func (q *Qliksense) SetUpQliksenseDefaultContext() error {
	return q.SetUpQliksenseContext(DefaultQliksenseContext, true)
}

// SetUpQliksenseContext - to setup qliksense context
func (q *Qliksense) SetUpQliksenseContext(contextName string, isDefaultContext bool) error {
	if contextName == "" {
		err := fmt.Errorf("Please enter a non-empty context-name")
		log.Println(err)
		return err
	}
	// check the length of the context name entered by the user, it should not exceed 17 chars
	if len(contextName) > MaxContextNameLength {
		err := fmt.Errorf("Please enter a context-name with utmost 17 characters")
		log.Println(err)
		return err
	}

	qliksenseConfigFile := filepath.Join(q.QliksenseHome, QliksenseConfigFile)
	var qliksenseConfig api.QliksenseConfig
	configFileTrack := false

	if !api.FileExists(qliksenseConfigFile) {
		qliksenseConfig.AddBaseQliksenseConfigs(contextName)
	} else {
		if err := api.ReadFromFile(&qliksenseConfig, qliksenseConfigFile); err != nil {
			log.Println(err)
			return err
		}
		if isDefaultContext { // if config file exits but a default context is requested, we want to prevent writing to config file
			configFileTrack = true
		}
	}
	// creating a file in the name of the context if it does not exist/ opening it to append/modify content if it already exists

	qliksenseContextsDir1 := filepath.Join(q.QliksenseHome, QliksenseContextsDir)
	if !api.DirExists(qliksenseContextsDir1) {
		if err := os.Mkdir(qliksenseContextsDir1, os.ModePerm); err != nil {
			err = fmt.Errorf("Not able to create %s dir: %v", qliksenseContextsDir1, err)
			log.Println(err)
			return err
		}
	}
	api.LogDebugMessage("%s exists", qliksenseContextsDir1)

	// creating contexts/qlik-default/qlik-default.yaml file
	qliksenseContextFile := filepath.Join(qliksenseContextsDir1, contextName, contextName+".yaml")
	var qliksenseCR api.QliksenseCR

	defaultContextsDir := filepath.Join(qliksenseContextsDir1, contextName)
	if !api.DirExists(defaultContextsDir) {
		if err := os.Mkdir(defaultContextsDir, os.ModePerm); err != nil {
			err = fmt.Errorf("Not able to create %s: %v", defaultContextsDir, err)
			log.Println(err)
			return err
		}
	}
	api.LogDebugMessage("%s exists", defaultContextsDir)
	if !api.FileExists(qliksenseContextFile) {
		qliksenseCR.AddCommonConfig(contextName)
		api.LogDebugMessage("Added Context: %s", contextName)
	} else {
		if err := api.ReadFromFile(&qliksenseCR, qliksenseContextFile); err != nil {
			log.Println(err)
			return err
		}
	}

	api.WriteToFile(&qliksenseCR, qliksenseContextFile)
	ctxTrack := false
	if len(qliksenseConfig.Spec.Contexts) > 0 {
		for _, ctx := range qliksenseConfig.Spec.Contexts {
			if ctx.Name == contextName {
				ctx.CrFile = qliksenseContextFile
				ctxTrack = true
				break
			}
		}
	}
	if !ctxTrack {
		qliksenseConfig.Spec.Contexts = append(qliksenseConfig.Spec.Contexts, api.Context{
			Name:   contextName,
			CrFile: qliksenseContextFile,
		})
	}
	qliksenseConfig.Spec.CurrentContext = contextName
	if !configFileTrack {
		api.WriteToFile(&qliksenseConfig, qliksenseConfigFile)
	}

	return nil
}

func validateInput(input string) (string, error) {
	var err error
	validInputs := []string{"yes", "no", "None"}
	isValid := false
	for _, elem := range validInputs {
		if input == elem {
			isValid = true
			break
		}
	}
	if !isValid {
		err = fmt.Errorf("Please enter one of: yes, no or None")
		log.Println(err)

	}
	return input, err
}

// PrepareK8sSecret decodes and decrypts the secret value in the secret.yaml file and returns a B64encoded string
func (q *Qliksense) PrepareK8sSecret(targetFile string) (string, error) {
	// check if targetFile exists
	if !api.FileExists(targetFile) {
		err := fmt.Errorf("Target file does not exist in the path provided")
		log.Println(err)
		return "", err
	}
	qConfig := api.NewQConfig(q.QliksenseHome)
	_, rsaPrivateKey, err := qConfig.GetCurrentContextEncryptionKeyPair()
	if err != nil {
		return "", err
	}

	// read the target file
	k8sSecret, err := readTargetfile(targetFile)
	if err != nil {
		return "", err
	}
	// retrieve value from data section
	k8sSecret1, err := api.K8sSecretFromYaml(k8sSecret)
	if err != nil {
		return "", err
	}
	dataMap := k8sSecret1.Data
	var resultMap = make(map[string][]byte)
	for k, v := range dataMap {
		decryptedString, err := api.Decrypt(v, rsaPrivateKey)
		if err != nil {
			err := fmt.Errorf("Not able to decrypt message: %v", err)
			return "", err
		}
		resultMap[k] = decryptedString
	}

	// putting the above map back into the k8sSecret struct
	k8sSecret1.Data = resultMap
	k8sSecretBytes, err := api.K8sSecretToYaml(k8sSecret1)
	if err != nil {
		return "", err
	}
	return string(k8sSecretBytes), nil
}

func readTargetfile(targetFile string) ([]byte, error) {
	k8sSecret, err := ioutil.ReadFile(targetFile)
	if err != nil {
		err := fmt.Errorf("Unable to read the targetFile")
		log.Println(err)
		return nil, err
	}
	return k8sSecret, nil
}

func (q *Qliksense) SetImageRegistry(registry, pushUsername, pushPassword, pullUsername, pullPassword string) error {
	qConfig := api.NewQConfig(q.QliksenseHome)
	qliksenseCR, qliksenseContextsFile, err := retrieveCurrentContextInfo(q)
	if err != nil {
		return err
	}
	if pushUsername != "" {
		if err := qConfig.SetPushDockerConfigJsonSecret(&api.DockerConfigJsonSecret{
			Uri:      registry,
			Username: pushUsername,
			Password: pushPassword,
		}); err != nil {
			return err
		} else if err := qConfig.SetPullDockerConfigJsonSecret(&api.DockerConfigJsonSecret{
			Name:      pullSecretName,
			Namespace: qliksenseCR.Spec.NameSpace,
			Uri:       registry,
			Username:  pullUsername,
			Password:  pullPassword,
			Email:     pullUsername,
		}); err != nil {
			return err
		}
	}
	qliksenseCR.Spec.AddToConfigs("qliksense", imageRegistryConfigKey, registry)
	return api.WriteToFile(&qliksenseCR, qliksenseContextsFile)
}
