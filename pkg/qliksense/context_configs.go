package qliksense

import (
	"io/ioutil"
	"os"

	"github.com/qlik-oss/k-apis/config"
	"github.com/qlik-oss/sense-installer/pkg/api"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

const (
	QliksenseConfigHome        = "/.qliksense"
	QliksenseConfigContextHome = "/.qliksense/contexts"

	QliksenseConfigApiVersion = "config.qlik.com/v1"
	QliksenseConfigKind       = "QliksenseConfig"
	QliksenseMetadataName     = "QliksenseConfigMetadata"

	QliksenseContextApiVersion    = "qlik.com/v1"
	QliksenseContextKind          = "Qliksense"
	QliksenseContextLabel         = "v1.0.0"
	QliksenseContextManifestsRoot = "/Usr/ddd/my-k8-repo/manifests"
)

// ReadQliksenseContextConfig is exported
func ReadQliksenseContextConfig(qliksenseCR *api.QliksenseCR, fileName string) {
	log.Debugf("Reading file %s", fileName)
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Error reading from source: %s\n", err)
	}
	if err = yaml.Unmarshal([]byte(yamlFile), qliksenseCR); err != nil {
		log.Fatalf("Error when parsing from source: %s\n", err)
	}
}

// WriteToFile is exported
func WriteToFile(content interface{}, targetFile string) {
	log.Debug("Entry: WriteToFile()")
	if content == nil || targetFile == "" {
		return
	}
	log.Debug("This action is about writing to a file")
	// log.Debugf("File %s doesnt exist, creating it now...", targetFile)
	file, err := os.OpenFile(targetFile, os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		log.Debug("There was an error creating the file: %s, %v", targetFile, err)
		log.Fatal(err)
	}
	defer file.Close()
	x, err := yaml.Marshal(content)
	if err != nil {
		log.Fatalf("An error occurred during marshalling CR: %v", err)
	}
	log.Debugf("Marshalled yaml:\n%s\nWriting to file...", x)

	// truncating the file before we write new content
	file.Truncate(0)
	file.Seek(0, 0)
	numBytes, err := file.Write(x)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("wrote %d bytes\n", numBytes)
	log.Debugf("Wrote Struct into %s", targetFile)
}

// ReadFromFile is exported
func ReadFromFile(content interface{}, sourceFile string) {
	log.Debug("Entry: ReadFromFile()")
	if content == nil || sourceFile == "" {
		return
	}
	log.Debug("This action is about reading from a file")
	contents, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		log.Debug("There was an error reading from file: %s, %v", sourceFile, err)
		log.Fatal(err)
	}
	if err := yaml.Unmarshal(contents, content); err != nil {
		log.Fatalf("An error occurred during unmarshalling: %v", err)
	}
}

// AddCommonConfig is exported
func AddCommonConfig(qliksenseCR api.QliksenseCR, contextName string) api.QliksenseCR {
	log.Debug("Entry: addCommonConfig()")
	qliksenseCR.ApiVersion = QliksenseContextApiVersion
	qliksenseCR.Kind = QliksenseContextKind
	if qliksenseCR.Metadata.Name == "" {
		qliksenseCR.Metadata.Name = contextName
	}
	qliksenseCR.Metadata.Labels = map[string]string{}
	qliksenseCR.Metadata.Labels["Version"] = QliksenseContextLabel
	qliksenseCR.Spec = &config.CRSpec{}
	qliksenseCR.Spec.ManifestsRoot = QliksenseContextManifestsRoot
	log.Debug("Exit: addCommonConfig()")
	return qliksenseCR
}

// AddBaseQliksenseConfigs is exported
func AddBaseQliksenseConfigs(qliksenseConfig api.QliksenseConfig, defaultQliksenseContext string) api.QliksenseConfig {
	log.Debug("Entry: AddBaseQliksenseConfigs()")
	qliksenseConfig.ApiVersion = QliksenseConfigApiVersion
	qliksenseConfig.Kind = QliksenseConfigKind
	qliksenseConfig.Metadata.Name = QliksenseMetadataName
	if defaultQliksenseContext != "" {
		qliksenseConfig.Spec.CurrentContext = defaultQliksenseContext
	}
	log.Debug("Exit: AddBaseQliksenseConfigs()")
	return qliksenseConfig
}

func setOtherConfigs(q *Qliksense) error {
	return nil
}

func checkExits(filename string) os.FileInfo {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		log.Debug("File does not exist")
		return nil
	}
	log.Debug("Either File exists OR a different error occurred")
	return info
}

// FileExists is exported
func FileExists(filename string) bool {
	if fe := checkExits(filename); fe != nil && !fe.IsDir() {
		return true
	}
	return false
}

func DirExists(dirname string) bool {
	if fe := checkExits(dirname); fe != nil && fe.IsDir() {
		return true
	}
	return false
}
