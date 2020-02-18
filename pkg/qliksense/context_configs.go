package qliksense

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	b64 "encoding/base64"

	"github.com/qlik-oss/k-apis/pkg/config"
	"github.com/qlik-oss/sense-installer/pkg/api"
	yaml "gopkg.in/yaml.v2"
)

const (
	// Below are some constants to support qliksense context setup
	QliksenseConfigHome        = "/.qliksense"
	QliksenseConfigContextHome = "/.qliksense/contexts"
	QliksenseConfigApiVersion  = "config.qlik.com/v1"
	QliksenseConfigKind        = "QliksenseConfig"
	QliksenseMetadataName      = "QliksenseConfigMetadata"
	QliksenseContextApiVersion = "qlik.com/v1"
	QliksenseContextKind       = "Qliksense"
	QliksenseDefaultProfile    = "docker-desktop"
	QliksenseConfigFile        = "config.yaml"
	QliksenseContextsDir       = "contexts"
	DefaultQliksenseContext    = "qlik-default"
	DefaultRotateKeys          = "yes"
	MaxContextNameLength       = 17
	QliksenseSecretsDir        = "secrets"
	QliksensePublicKey         = "qliksensePub"
	QliksensePrivateKey        = "qliksensePriv"
)

// WriteToFile writes content into specified file
func WriteToFile(content interface{}, targetFile string) error {
	if content == nil || targetFile == "" {
		return nil
	}
	file, err := os.OpenFile(targetFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		err = fmt.Errorf("There was an error creating the file: %s, %v", targetFile, err)
		log.Println(err)
		return err
	}
	defer file.Close()
	x, err := yaml.Marshal(content)
	if err != nil {
		err = fmt.Errorf("An error occurred during marshalling CR: %v", err)
		log.Println(err)
		return err
	}

	// truncating the file before we write new content
	file.Truncate(0)
	file.Seek(0, 0)
	_, err = file.Write(x)
	if err != nil {
		log.Println(err)
		return err
	}
	LogDebugMessage("Wrote content into %s", targetFile)
	return nil
}

// ReadFromFile reads content from specified sourcefile
func ReadFromFile(content interface{}, sourceFile string) error {
	if content == nil || sourceFile == "" {
		return nil
	}
	contents, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		err = fmt.Errorf("There was an error reading from file: %s, %v", sourceFile, err)
		log.Println(err)
		return err
	}
	if err := yaml.Unmarshal(contents, content); err != nil {
		err = fmt.Errorf("An error occurred during unmarshalling: %v", err)
		log.Println(err)
		return err
	}
	return nil
}

// AddCommonConfig adds common configs into CRs
func AddCommonConfig(qliksenseCR api.QliksenseCR, contextName string) api.QliksenseCR {
	qliksenseCR.ApiVersion = QliksenseContextApiVersion
	qliksenseCR.Kind = QliksenseContextKind
	if qliksenseCR.Metadata == nil {
		qliksenseCR.Metadata = &api.Metadata{}
	}
	if qliksenseCR.Metadata.Name == "" {
		qliksenseCR.Metadata.Name = contextName
	}
	qliksenseCR.Spec = &config.CRSpec{}
	qliksenseCR.Spec.Profile = QliksenseDefaultProfile
	qliksenseCR.Spec.ReleaseName = contextName
	qliksenseCR.Spec.RotateKeys = DefaultRotateKeys
	return qliksenseCR
}

// AddBaseQliksenseConfigs adds configs into config.yaml
func AddBaseQliksenseConfigs(qliksenseConfig api.QliksenseConfig, defaultQliksenseContext string) api.QliksenseConfig {
	qliksenseConfig.ApiVersion = QliksenseConfigApiVersion
	qliksenseConfig.Kind = QliksenseConfigKind
	if qliksenseConfig.Metadata == nil {
		qliksenseConfig.Metadata = &api.Metadata{}
	}
	qliksenseConfig.Metadata.Name = QliksenseMetadataName
	if defaultQliksenseContext != "" {
		if qliksenseConfig.Spec == nil {
			qliksenseConfig.Spec = &api.ContextSpec{}
		}
		qliksenseConfig.Spec.CurrentContext = defaultQliksenseContext
	}
	return qliksenseConfig
}

func checkExists(filename string, isFile bool) os.FileInfo {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		if isFile {
			LogDebugMessage("File does not exist")
		} else {
			LogDebugMessage("Dir does not exist")
		}
		return nil
	}
	LogDebugMessage("File exists")
	return info
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	if fe := checkExists(filename, true); fe != nil && !fe.IsDir() {
		return true
	}
	return false
}

// DirExists checks if a directory exists
func DirExists(dirname string) bool {
	if fe := checkExists(dirname, false); fe != nil && fe.IsDir() {
		return true
	}
	return false
}

// LogDebugMessage logs a debug message
func LogDebugMessage(strMessage string, args ...interface{}) {
	if os.Getenv("QLIKSENSE_DEBUG") == "true" {
		log.Printf(strMessage, args...)
	}
}

// SetSecrets - set-secrets <key>=<value> commands
func SetSecrets(q *Qliksense, args []string, isK8sSecret bool) error {
	LogDebugMessage("Args received: %v\n", args)
	LogDebugMessage("isK8sSecret: %v\n", isK8sSecret)

	// retieve current context from config.yaml
	qliksenseCR, qliksenseContextsFile, err := retrieveCurrentContextInfo(q)
	if err != nil {
		return err
	}

	secretKeyPairLocation := filepath.Join(q.QliksenseHome, QliksenseSecretsDir, QliksenseContextsDir, qliksenseCR.Metadata.Name, QliksenseSecretsDir)
	LogDebugMessage("SecretKeyLocation to store key pair: %s", secretKeyPairLocation)

	if os.Getenv("QLIKSENSE_KEY_LOCATION") != "" {
		LogDebugMessage("Env variable: QLIKSENSE_KEY_LOCATION= %s", os.Getenv("QLIKSENSE_KEY_LOCATION"))
		secretKeyPairLocation = os.Getenv("QLIKSENSE_KEY_LOCATION")
	}
	// Env var: QLIKSENSE_KEY_LOCATION hasn't been set, so dropping key pair in the location:
	// /.qliksense/secrets/contexts/<current-context>/secrets/
	LogDebugMessage("Using default location to store keys: %s", secretKeyPairLocation)

	publicKeyFilePath := filepath.Join(secretKeyPairLocation, QliksensePublicKey)
	privateKeyFilePath := filepath.Join(secretKeyPairLocation, QliksensePrivateKey)

	// try to create the dir if it doesn't exist
	if !FileExists(publicKeyFilePath) || !FileExists(privateKeyFilePath) {
		LogDebugMessage("Qliksense secretKeyLocation dir does not exist, creating it now: %s", secretKeyPairLocation)
		if err := os.MkdirAll(secretKeyPairLocation, os.ModePerm); err != nil {
			err = fmt.Errorf("Not able to create %s dir: %v", secretKeyPairLocation, err)
			log.Println(err)
			return err
		}
		// generating and storing key-pair
		err1 := generateAndStoreSecretKeypair(secretKeyPairLocation)
		if err1 != nil {
			err1 = fmt.Errorf("Not able to generate and store key pair for encryption")
			log.Println(err1)
			return err1
		}
	}

	var rsaPublicKey *rsa.PublicKey
	var e1 error

	// Read Public Key
	publicKeybytes, err2 := readKeys(publicKeyFilePath)
	if err2 != nil {
		LogDebugMessage("Not able to read public key")
		return err2
	}
	// LogDebugMessage("PublicKey: %+v", publicKeybytes)

	// convert []byte into RSA public key object
	rsaPublicKey, e1 = DecodePublicKey(publicKeybytes)
	if e1 != nil {
		return e1
	}
	processingFunc := func(arg1, key, value string) {
		// Metadata name in qliksense CR is the name of the current context
		LogDebugMessage("Trying to retreive current context: %+v ----- %s", qliksenseCR.Metadata.Name, qliksenseContextsFile)

		// encrypt value with RSA key pair
		valueBytes := []byte(value)
		cipherText, e2 := Encrypt(valueBytes, rsaPublicKey)
		if e2 != nil {
			return e2
		}
		LogDebugMessage("Returned cipher text: %s", b64.StdEncoding.EncodeToString(cipherText))

		if isK8sSecret {
			// TODO: store the key value in k8s as secret
			LogDebugMessage("Need to create a Kubernetes secret")
		}
		// TODO: Extend AddToSecrets to support adding k8s key ref. (OR) create a separate method to support that
		// store the encrypted value in the file (OR)
		// TODO: k8s secret name in the file
		qliksenseCR.Spec.AddToSecrets(arg1, key, string(cipherText))
	}
	processConfigArgs(args, isK8sSecret, qliksenseCR.Spec, processingFunc)
	// write modified content into context.yaml
	WriteToFile(&qliksenseCR, qliksenseContextsFile)

	return nil
}

// SetConfigs - set-configs <key>=<value> commands
func SetConfigs(q *Qliksense, args []string) error {
	// retieve current context from config.yaml
	qliksenseCR, qliksenseContextsFile, err := retrieveCurrentContextInfo(q)
	if err != nil {
		return err
	}

	processConfigArgs(args, false, qliksenseCR.Spec, qliksenseCR.Spec.AddToConfigs)
	// write modified content into context.yaml
	WriteToFile(&qliksenseCR, qliksenseContextsFile)

	return nil
}

func readKeys(keyFile string) ([]byte, error) {
	keybyteArray, err := ioutil.ReadFile(keyFile)
	if err != nil {
		err = fmt.Errorf("There was an error reading from file: %s, %v", keyFile, err)
		log.Println(err)
		return nil, err
	} else {
		LogDebugMessage("Read key as byte[]: %+v", keybyteArray)
	}
	return keybyteArray, nil
}

func retrieveCurrentContextInfo(q *Qliksense) (api.QliksenseCR, string, error) {
	var qliksenseConfig api.QliksenseConfig
	qliksenseConfigFile := filepath.Join(q.QliksenseHome, QliksenseConfigFile)

	ReadFromFile(&qliksenseConfig, qliksenseConfigFile)
	currentContext := qliksenseConfig.Spec.CurrentContext
	LogDebugMessage("Current-context from config.yaml: %s", currentContext)
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
	if !FileExists(qliksenseContextsFile) {
		err := fmt.Errorf("Context file does not exist.\nPlease try re-running `qliksense config set-context <context-name>` and then `qliksense config view` again")
		log.Println(err)
		return api.QliksenseCR{}, "", err
	}
	ReadFromFile(&qliksenseCR, qliksenseContextsFile)

	LogDebugMessage("Read context file: %s, Read QliksenseCR: %v", qliksenseContextsFile, qliksenseCR)
	return qliksenseCR, qliksenseContextsFile, nil
}

func processConfigArgs(args []string, shouldEncrypt bool, cr *config.CRSpec, updateFn func(string, string, string)) error {
	// prepare received args
	// split args[0] into key and value
	if len(args) == 0 {
		err := fmt.Errorf("No args were provided. Please provide args to configure the current context")
		log.Println(err)
		return err
	}

	re1 := regexp.MustCompile(`(\w{1,})\[name=(\w{1,})\]=("*[\w\-_/:0-9]+"*)`)
	for _, arg := range args {
		log.Printf("Arg here: %s", arg)
		result := re1.FindStringSubmatch(arg)
		// check if result array's length is == 4 (index 0 - is the full match & indices 1,2,3- are the fields we need)
		if len(result) != 4 {
			err := fmt.Errorf("Please provide valid args for this command")
			log.Println(err)
			return err
		}
		updateFn(result[1], result[2], result[3])
	}
	return nil
}

// SetOtherConfigs - set profile/namespace/storageclassname/git.repository commands
func SetOtherConfigs(q *Qliksense, args []string) error {
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
			LogDebugMessage("Current profile after modification: %s ", qliksenseCR.Spec.Profile)
		case "namespace":
			qliksenseCR.Spec.NameSpace = argsString[1]
			LogDebugMessage("Current namespace after modification: %s ", qliksenseCR.Spec.NameSpace)
		case "git.repository":
			qliksenseCR.Spec.Git.Repository = argsString[1]
			LogDebugMessage("Current git repository after modification: %s ", qliksenseCR.Spec.Git.Repository)
		case "storageClassName":
			qliksenseCR.Spec.StorageClassName = argsString[1]
			LogDebugMessage("Current StorageClassName after modification: %s ", qliksenseCR.Spec.StorageClassName)
		case "rotateKeys":
			rotateKeys, err := validateInput(argsString[1])
			if err != nil {
				return err
			}
			qliksenseCR.Spec.RotateKeys = rotateKeys
			LogDebugMessage("Current rotateKeys after modification: %s ", qliksenseCR.Spec.RotateKeys)
		default:
			log.Println("As part of the `qliksense config set` command, please enter one of: profile, namespace, storageClassName,rotateKeys or git.repository arguments")
		}
	}
	// write modified content into context.yaml
	WriteToFile(&qliksenseCR, qliksenseContextsFile)

	return nil
}

// SetContextConfig - set the context for qliksense kubernetes resources to live in
func SetContextConfig(q *Qliksense, args []string) error {
	if len(args) == 1 {
		SetUpQliksenseContext(q.QliksenseHome, args[0], false)
	} else {
		err := fmt.Errorf("Please provide a name to configure the context with")
		log.Println(err)
		return err
	}
	return nil
}

// SetUpQliksenseDefaultContext - to setup dir structure for default qliksense context
func SetUpQliksenseDefaultContext(qlikSenseHome string) error {
	return SetUpQliksenseContext(qlikSenseHome, DefaultQliksenseContext, true)
}

// SetUpQliksenseContext - to setup qliksense context
func SetUpQliksenseContext(qlikSenseHome, contextName string, isDefaultContext bool) error {
	// check the length of the context name entered by the user, it should not exceed 17 chars
	if len(contextName) > MaxContextNameLength {
		err := fmt.Errorf("Please enter a context-name with utmost 17 characters")
		log.Println(err)
		return err
	}

	qliksenseConfigFile := filepath.Join(qlikSenseHome, QliksenseConfigFile)
	var qliksenseConfig api.QliksenseConfig
	configFileTrack := false

	if !FileExists(qliksenseConfigFile) {
		qliksenseConfig = AddBaseQliksenseConfigs(qliksenseConfig, contextName)
	} else {
		ReadFromFile(&qliksenseConfig, qliksenseConfigFile)
		if isDefaultContext { // if config file exits but a default context is requested, we want to prevent writing to config file
			configFileTrack = true
		}
	}
	// creating a file in the name of the context if it does not exist/ opening it to append/modify content if it already exists

	qliksenseContextsDir1 := filepath.Join(qlikSenseHome, QliksenseContextsDir)
	if !DirExists(qliksenseContextsDir1) {
		if err := os.Mkdir(qliksenseContextsDir1, os.ModePerm); err != nil {
			err = fmt.Errorf("Not able to create %s dir: %v", qliksenseContextsDir1, err)
			log.Println(err)
			return err
		}
	}
	LogDebugMessage("%s exists", qliksenseContextsDir1)

	// creating contexts/qlik-default/qlik-default.yaml file
	qliksenseContextFile := filepath.Join(qliksenseContextsDir1, contextName, contextName+".yaml")
	var qliksenseCR api.QliksenseCR

	defaultContextsDir := filepath.Join(qliksenseContextsDir1, contextName)
	if !DirExists(defaultContextsDir) {
		if err := os.Mkdir(defaultContextsDir, os.ModePerm); err != nil {
			err = fmt.Errorf("Not able to create %s: %v", defaultContextsDir, err)
			log.Println(err)
			return err
		}
	}
	LogDebugMessage("%s exists", defaultContextsDir)
	if !FileExists(qliksenseContextFile) {
		qliksenseCR = AddCommonConfig(qliksenseCR, contextName)
		LogDebugMessage("Added Context: %s", contextName)
	} else {
		ReadFromFile(&qliksenseCR, qliksenseContextFile)
	}

	WriteToFile(&qliksenseCR, qliksenseContextFile)
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
		WriteToFile(&qliksenseConfig, qliksenseConfigFile)
	}

	return nil
}

// func generateAndStoreSecretKeypair(q *Qliksense, contextName string) error {
func generateAndStoreSecretKeypair(secretsPath string) error {
	LogDebugMessage("%s exists", secretsPath)
	// creating contexts/qlik-default/secrets/qliksensePub and contexts/qlik-default/secrets/qliksensePriv files
	publicKeyFilePath := filepath.Join(secretsPath, QliksensePublicKey)
	privateKeyFilePath := filepath.Join(secretsPath, QliksensePrivateKey)
	LogDebugMessage("Generating public-private key pair.....")
	GenerateRSAEncryptionKeys(publicKeyFilePath, privateKeyFilePath)
	LogDebugMessage("Generated public-private key pairs")

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
