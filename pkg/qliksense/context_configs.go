package qliksense

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/qlik-oss/k-apis/pkg/config"
	"github.com/qlik-oss/sense-installer/pkg/api"
	yaml "gopkg.in/yaml.v2"
)

const (
	// Below are some constants to support qliksense context setup
	QliksenseConfigHome           = "/.qliksense"
	QliksenseConfigContextHome    = "/.qliksense/contexts"
	QliksenseConfigApiVersion     = "config.qlik.com/v1"
	QliksenseConfigKind           = "QliksenseConfig"
	QliksenseMetadataName         = "QliksenseConfigMetadata"
	QliksenseContextApiVersion    = "qlik.com/v1"
	QliksenseContextKind          = "Qliksense"
	QliksenseContextLabel         = "v1.0.0"
	QliksenseContextManifestsRoot = "/Usr/ddd/my-k8-repo/manifests"
	QliksenseDefaultProfile       = "docker-desktop"
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
	if qliksenseCR.Metadata.Name == "" {
		qliksenseCR.Metadata.Name = contextName
	}
	qliksenseCR.Metadata.Labels = map[string]string{}
	qliksenseCR.Metadata.Labels["Version"] = QliksenseContextLabel
	qliksenseCR.Spec = &config.CRSpec{}
	qliksenseCR.Spec.ManifestsRoot = QliksenseContextManifestsRoot
	qliksenseCR.Spec.Profile = QliksenseDefaultProfile
	return qliksenseCR
}

// AddBaseQliksenseConfigs adds configs into config.yaml
func AddBaseQliksenseConfigs(qliksenseConfig api.QliksenseConfig, defaultQliksenseContext string) api.QliksenseConfig {
	qliksenseConfig.ApiVersion = QliksenseConfigApiVersion
	qliksenseConfig.Kind = QliksenseConfigKind
	qliksenseConfig.Metadata.Name = QliksenseMetadataName
	if defaultQliksenseContext != "" {
		qliksenseConfig.Spec.CurrentContext = defaultQliksenseContext
	}
	return qliksenseConfig
}

func checkExits(filename string) os.FileInfo {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		LogDebugMessage("File does not exist")
		return nil
	}
	LogDebugMessage("File exists")
	return info
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	if fe := checkExits(filename); fe != nil && !fe.IsDir() {
		return true
	}
	return false
}

// DirExists checks if a directory exists
func DirExists(dirname string) bool {
	if fe := checkExits(dirname); fe != nil && fe.IsDir() {
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
