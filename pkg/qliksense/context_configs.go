package qliksense

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	DefaultQliksenseContext    = "qliksense-default"
)

// WriteToFile writes content into specified file
func WriteToFile(content interface{}, targetFile string) {
	if content == nil || targetFile == "" {
		return
	}
	file, err := os.OpenFile(targetFile, os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		LogDebugMessage("There was an error creating the file: %s, %v", targetFile, err)
		log.Fatal(err)
	}
	defer file.Close()
	x, err := yaml.Marshal(content)
	if err != nil {
		log.Fatalf("An error occurred during marshalling CR: %v", err)
	}
	LogDebugMessage("Marshalled yaml:\n%s\nWriting to file...", x)

	// truncating the file before we write new content
	file.Truncate(0)
	file.Seek(0, 0)
	_, err = file.Write(x)
	if err != nil {
		log.Fatal(err)
	}
	LogDebugMessage("Wrote content into %s", targetFile)
}

// ReadFromFile reads content from specified sourcefile
func ReadFromFile(content interface{}, sourceFile string) {
	if content == nil || sourceFile == "" {
		return
	}
	contents, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		LogDebugMessage("There was an error reading from file: %s, %v", sourceFile, err)
		log.Fatal(err)
	}
	if err := yaml.Unmarshal(contents, content); err != nil {
		log.Fatalf("An error occurred during unmarshalling: %v", err)
	}
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
func SetSecrets(q *Qliksense, args []string) error {
	// retieve current context from config.yaml
	qliksenseCR, qliksenseContextsFile := retrieveCurrentContextInfo(q)

	processConfigArgs(args, qliksenseCR.Spec, qliksenseCR.Spec.AddToSecrets)
	LogDebugMessage("CR now: %v", qliksenseCR.Spec)

	// write modified content into context.yaml
	WriteToFile(&qliksenseCR, qliksenseContextsFile)

	return nil
}

// SetConfigs - set-configs <key>=<value> commands
func SetConfigs(q *Qliksense, args []string) error {
	// retieve current context from config.yaml
	qliksenseCR, qliksenseContextsFile := retrieveCurrentContextInfo(q)

	processConfigArgs(args, qliksenseCR.Spec, qliksenseCR.Spec.AddToConfigs)
	LogDebugMessage("CR now: %v", qliksenseCR.Spec)
	// write modified content into context.yaml
	WriteToFile(&qliksenseCR, qliksenseContextsFile)

	return nil
}

func retrieveCurrentContextInfo(q *Qliksense) (api.QliksenseCR, string) {
	var qliksenseConfig api.QliksenseConfig
	qliksenseConfigFile := filepath.Join(q.QliksenseHome, QliksenseConfigFile)
	LogDebugMessage("qliksenseConfigFile: %s", qliksenseConfigFile)

	ReadFromFile(&qliksenseConfig, qliksenseConfigFile)
	currentContext := qliksenseConfig.Spec.CurrentContext
	LogDebugMessage("Current-context from config.yaml: %s", currentContext)
	if currentContext == "" {
		// current-context is empty
		log.Fatal(`Please run the "qliksense config set-context <context-name>" first before viewing the current context info`)
	}
	// read the context.yaml file
	var qliksenseCR api.QliksenseCR
	if currentContext == "" {
		// current-context is empty
		log.Fatal(`Please run the "qliksense config set-context <context-name>" first before viewing the current context info`)
	}
	qliksenseContextsFile := filepath.Join(q.QliksenseHome, QliksenseContextsDir, currentContext, currentContext+".yaml")
	if !FileExists(qliksenseContextsFile) {
		log.Fatalf("Context file does not exist.\nPlease try re-running `qliksense config set-context <context-name>` and then `qliksense config view` again")
	}
	ReadFromFile(&qliksenseCR, qliksenseContextsFile)

	LogDebugMessage("Read QliksenseCR: %v", qliksenseCR)
	LogDebugMessage("Read context file: %s", qliksenseContextsFile)
	return qliksenseCR, qliksenseContextsFile
}

func processConfigArgs(args []string, cr *config.CRSpec, updateFn func(string, string, string)) {
	// prepare received args
	// split args[0] into key and value
	if len(args) == 0 {
		log.Fatalf("No args were provided. Please provide args to configure the current context")
	}

	re1 := regexp.MustCompile(`(\w{1,})\[name=(\w{1,})\]=("*\w+"*)`)
	for _, arg := range args {
		result := re1.FindStringSubmatch(arg)
		LogDebugMessage("finding matches...\n")
		LogDebugMessage("Results: %s, %s, %s", result[1], result[2], result[3])
		// check if result array's length is == 4 (index 0 - is the full match & indices 1,2,3- are the fields we need)
		if len(result) != 4 {
			log.Fatal("Please provide valid args for this command")
		}
		updateFn(result[1], result[2], result[3])
	}
}

// SetOtherConfigs - set profile/namespace/storageclassname/git.repository commands
func SetOtherConfigs(q *Qliksense, args []string) error {
	// retieve current context from config.yaml
	qliksenseCR, qliksenseContextsFile := retrieveCurrentContextInfo(q)

	// modify appropriate fields
	LogDebugMessage("Command: %s", args[0])
	// split args[0] into key and value
	if len(args) > 0 {
		argsString := strings.Split(args[0], "=")
		LogDebugMessage("Split string: %v", argsString)
		switch argsString[0] {
		case "profile":
			LogDebugMessage("Current profile: %s, Incoming profile: %s", qliksenseCR.Spec.Profile, argsString[1])
			qliksenseCR.Spec.Profile = argsString[1]
			LogDebugMessage("Current profile after modification: %s ", qliksenseCR.Spec.Profile)
		case "namespace":
			LogDebugMessage("Current namespace: %s, Incoming namespace: %s", qliksenseCR.Spec.NameSpace, argsString[1])
			qliksenseCR.Spec.NameSpace = argsString[1]
			LogDebugMessage("Current namespace after modification: %s ", qliksenseCR.Spec.NameSpace)
		case "git.repository":
			LogDebugMessage("Current git.repository: %s, Incoming git.repository: %s", qliksenseCR.Spec.Git.Repository, argsString[1])
			qliksenseCR.Spec.Git.Repository = argsString[1]
			LogDebugMessage("Current git repository after modification: %s ", qliksenseCR.Spec.Git.Repository)
		case "storageClassName":
			LogDebugMessage("Current StorageClassName: %s, Incoming StorageClassName: %s", qliksenseCR.Spec.StorageClassName, argsString[1])
			qliksenseCR.Spec.StorageClassName = argsString[1]
			LogDebugMessage("Current StorageClassName after modification: %s ", qliksenseCR.Spec.StorageClassName)
		default:
			log.Println("As part of the `qliksense config set` command, please enter one of: profile, namespace, storageClassName or git.repository arguments")
		}
	} else {
		log.Fatalf("No args were provided. Please provide args to configure the current context")
	}
	// write modified content into context.yaml
	WriteToFile(&qliksenseCR, qliksenseContextsFile)

	return nil
}

// SetContextConfig - set the context for qliksense kubernetes resources to live in
func SetContextConfig(q *Qliksense, args []string) error {
	if len(args) == 1 {
		LogDebugMessage("The command received: %s", args)
		SetUpQliksenseContext(q.QliksenseHome, args[0])
	} else {
		log.Fatalf("Please provide a name to configure the context with.")
	}
	return nil
}

// SetUpQliksenseDefaultContext - to setup dir structure for default qliksense context
func SetUpQliksenseDefaultContext(qlikSenseHome string) {
	SetUpQliksenseContext(qlikSenseHome, DefaultQliksenseContext)
}

// SetUpQliksenseContext - to setup qliksense context
func SetUpQliksenseContext(qlikSenseHome, contextName string) {
	qliksenseConfigFile := filepath.Join(qlikSenseHome, QliksenseConfigFile)
	var qliksenseConfig api.QliksenseConfig
	if !FileExists(qliksenseConfigFile) {
		qliksenseConfig = AddBaseQliksenseConfigs(qliksenseConfig, contextName)
	} else {
		ReadFromFile(&qliksenseConfig, qliksenseConfigFile)
	}
	// creating a file in the name of the context if it does not exist/ opening it to append/modify content if it already exists

	qliksenseContextsDir1 := filepath.Join(qlikSenseHome, QliksenseContextsDir)
	if !DirExists(qliksenseContextsDir1) {
		if err := os.Mkdir(qliksenseContextsDir1, 0700); err != nil {
			log.Fatalf("Not able to create the contexts/ dir: %v", err)
		}
	}
	LogDebugMessage("Created contexts/")
	// creating contexts/qliksense-default.yaml file

	qliksenseContextFile := filepath.Join(qliksenseContextsDir1, contextName, contextName+".yaml")
	var qliksenseCR api.QliksenseCR

	if err := os.Mkdir(filepath.Join(qliksenseContextsDir1, contextName), 0700); err != nil {
		log.Fatalf("Not able to create the contexts/qliksense-default/ dir: %v", err)
	}
	LogDebugMessage("Created contexts/qliksense-default/ directory")
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
	WriteToFile(&qliksenseConfig, qliksenseConfigFile)
}
