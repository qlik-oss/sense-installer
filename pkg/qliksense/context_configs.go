package qliksense

import (
	"crypto/rsa"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	b64 "encoding/base64"
	ansi "github.com/mattn/go-colorable"
	"github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/ttacon/chalk"
)

const (
	// Below are some constants to support qliksense context setup
	QliksenseConfigHome        = "/.qliksense"
	QliksenseConfigContextHome = "/.qliksense/contexts"

	QliksenseConfigFile     = "config.yaml"
	QliksenseContextsDir    = "contexts"
	DefaultQliksenseContext = "qlik-default"
	MaxContextNameLength    = 17
	QliksenseSecretsDir     = "secrets"
)

// SetSecrets - set-secrets <key>=<value> commands
func (q *Qliksense) SetSecrets(args []string, isK8sSecret bool) error {
	api.LogDebugMessage("Args received: %v\n", args)
	api.LogDebugMessage("isK8sSecret: %v\n", isK8sSecret)

	// retieve current context from config.yaml
	qliksenseCR, qliksenseContextsFile, err := retrieveCurrentContextInfo(q)
	if err != nil {
		return err
	}

	secretKeyPairLocation := filepath.Join(q.QliksenseHome, QliksenseSecretsDir, QliksenseContextsDir, qliksenseCR.Metadata.Name, QliksenseSecretsDir)
	api.LogDebugMessage("SecretKeyLocation to store key pair: %s", secretKeyPairLocation)

	if os.Getenv("QLIKSENSE_KEY_LOCATION") != "" {
		api.LogDebugMessage("Env variable: QLIKSENSE_KEY_LOCATION= %s", os.Getenv("QLIKSENSE_KEY_LOCATION"))
		secretKeyPairLocation = os.Getenv("QLIKSENSE_KEY_LOCATION")
	}
	// Env var: QLIKSENSE_KEY_LOCATION hasn't been set, so dropping key pair in the location:
	// /.qliksense/secrets/contexts/<current-context>/secrets/
	api.LogDebugMessage("Using default location to store keys: %s", secretKeyPairLocation)

	publicKeyFilePath := filepath.Join(secretKeyPairLocation, api.QliksensePublicKey)
	privateKeyFilePath := filepath.Join(secretKeyPairLocation, api.QliksensePrivateKey)

	// try to create the dir if it doesn't exist
	if !api.FileExists(publicKeyFilePath) || !api.FileExists(privateKeyFilePath) {
		api.LogDebugMessage("Qliksense secretKeyLocation dir does not exist, creating it now: %s", secretKeyPairLocation)
		if err := os.MkdirAll(secretKeyPairLocation, os.ModePerm); err != nil {
			err = fmt.Errorf("Not able to create %s dir: %v", secretKeyPairLocation, err)
			log.Println(err)
			return err
		}
		// generating and storing key-pair
		err1 := api.GenerateAndStoreSecretKeypair(secretKeyPairLocation)
		if err1 != nil {
			err1 = fmt.Errorf("Not able to generate and store key pair for encryption")
			log.Println(err1)
			return err1
		}
	}

	var rsaPublicKey *rsa.PublicKey
	var e1 error

	// Read Public Key
	publicKeybytes, err2 := api.ReadKeys(publicKeyFilePath)
	if err2 != nil {
		api.LogDebugMessage("Not able to read public key")
		return err2
	}
	// api.LogDebugMessage("PublicKey: %+v", publicKeybytes)

	// convert []byte into RSA public key object
	rsaPublicKey, e1 = api.DecodeToPublicKey(publicKeybytes)
	if e1 != nil {
		return e1
	}

	resultArgs, err := api.ProcessConfigArgs(args)
	if err != nil {
		return err
	}

	for _, ra := range resultArgs {
		// Metadata name in qliksense CR is the name of the current context
		api.LogDebugMessage("Trying to retreive current context: %+v ----- %s", qliksenseCR.Metadata.Name, qliksenseContextsFile)

		// encrypt value with RSA key pair
		valueBytes := []byte(ra.Value)
		cipherText, e2 := api.Encrypt(valueBytes, rsaPublicKey)
		if e2 != nil {
			return e2
		}
		api.LogDebugMessage("Returned cipher text: %s", b64.StdEncoding.EncodeToString(cipherText))

		if isK8sSecret {
			// TODO: store the key value in k8s as secret
			api.LogDebugMessage("Need to create a Kubernetes secret")
		}
		// TODO: Extend AddToSecrets to support adding k8s key ref. (OR) create a separate method to support that
		// store the encrypted value in the file (OR)
		// TODO: k8s secret name in the file
		qliksenseCR.Spec.AddToSecrets(ra.SvcName, ra.Key, string(cipherText))
	}

	// write modified content into context.yaml
	api.WriteToFile(&qliksenseCR, qliksenseContextsFile)

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

	api.ReadFromFile(&qliksenseConfig, qliksenseConfigFile)
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
	api.ReadFromFile(&qliksenseCR, qliksenseContextsFile)

	api.LogDebugMessage("Read context file: %s, Read QliksenseCR: %v", qliksenseContextsFile, qliksenseCR)
	return qliksenseCR, qliksenseContextsFile, nil
}

// SetOtherConfigs - set profile/namespace/storageclassname/git.repository commands
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
		case "rotateKeys":
			rotateKeys, err := validateInput(argsString[1])
			if err != nil {
				return err
			}
			qliksenseCR.Spec.RotateKeys = rotateKeys
			api.LogDebugMessage("Current rotateKeys after modification: %s ", qliksenseCR.Spec.RotateKeys)
		default:
			log.Println("As part of the `qliksense config set` command, please enter one of: profile, namespace, storageClassName,rotateKeys or git.repository arguments")
		}
	}
	// write modified content into context.yaml
	api.WriteToFile(&qliksenseCR, qliksenseContextsFile)

	return nil
}

// SetContextConfig - set the context for qliksense kubernetes resources to live in
func (q *Qliksense) SetContextConfig(args []string) error {
	if len(args) == 1 {
		q.SetUpQliksenseContext(args[0], false)
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
	api.ReadFromFile(&qliksenseConfig, qliksenseConfigFile)
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

func (q *Qliksense) DeleteContextConfig(args []string) error {
	if len(args) == 1 {
		fmt.Println(args)
		qliksenseConfigFile := filepath.Join(q.QliksenseHome, QliksenseConfigFile)
		var qliksenseConfig api.QliksenseConfig
		api.ReadFromFile(&qliksenseConfig, qliksenseConfigFile)
		switch args[0] {
		case qliksenseConfig.Spec.CurrentContext:
			log.Println("Error: Cannot delete current context -", qliksenseConfig.Spec.CurrentContext)
			fmt.Println("Please switch contexts to be able to delete this context.")
		default:
			qliksenseContextsDir1 := filepath.Join(q.QliksenseHome, QliksenseContextsDir)
			qliksenseContextFile := filepath.Join(qliksenseContextsDir1, args[0])
			fmt.Println(qliksenseContextFile)
			if err := os.RemoveAll(qliksenseContextFile); err != nil {
				err = fmt.Errorf("Not able to delete %s dir: %v", qliksenseContextsDir1, err)
				log.Println(err)
				return err
			} else {
				count := 0
				if len(qliksenseConfig.Spec.Contexts) > 0 {
					for _, ctx := range qliksenseConfig.Spec.Contexts {
						count = count + 1
						if ctx.Name == args[0] {
							qliksenseConfig.Spec.Contexts = append(qliksenseConfig.Spec.Contexts[:count], qliksenseConfig.Spec.Contexts[count+1])
							break
						}
					}
				}
				api.WriteToFile(&qliksenseConfig, qliksenseConfigFile)
			}
		}
	} else {
		err := fmt.Errorf("Please provide a context as an arguement to delete")
		log.Println(err)
		return err
	}
	return nil
}

// SetUpQliksenseDefaultContext - to setup dir structure for default qliksense context
func (q *Qliksense) SetUpQliksenseDefaultContext() error {
	return q.SetUpQliksenseContext(DefaultQliksenseContext, true)
}

// SetUpQliksenseContext - to setup qliksense context
func (q *Qliksense) SetUpQliksenseContext(contextName string, isDefaultContext bool) error {
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
		api.ReadFromFile(&qliksenseConfig, qliksenseConfigFile)
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
		api.ReadFromFile(&qliksenseCR, qliksenseContextFile)
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
