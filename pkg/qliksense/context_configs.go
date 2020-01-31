package qliksense

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/prometheus/common/log"
	"github.com/qlik-oss/k-apis/config"
	yaml "gopkg.in/yaml.v2"
)

const (
	QliksenseConfigHome        = "/.qliksense"
	QliksenseConfigContextHome = "/.qliksense/contexts"

	QliksenseConfigApiVersion = "config.qlik.com/v1"
	QliksenseConfigKind       = "QliksenseConfig"

	QliksenseContextApiVersion    = "qlik.com/v1"
	QliksenseContextKind          = "Qliksense"
	QliksenseContextLabel         = "v1.0.0"
	QliksenseContextManifestsRoot = "/Usr/ddd/my-k8-repo/manifests"
)

// ReadQliksenseContextConfig is exported
func (qliksenseContext *QliksenseContext) ReadQliksenseContextConfig(fileName string) {
	log.Debugf("Reading file %s", fileName)
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Error reading from source: %s\n", err)
	}
	if err = yaml.Unmarshal([]byte(yamlFile), qliksenseContext); err != nil {
		log.Fatalf("Error when parsing from source: %s\n", err)
	}
}

// WriteQliksenseConfigToFile is exported
func (qliksenseContext *QliksenseContext) WriteQliksenseConfigToFile(contextName string) {

	qliksenseContext.addCommonConfig(contextName)
	x, err := yaml.Marshal(qliksenseContext)
	if err != nil {
		log.Fatalf("An error occurred during marshalling config: %v", err)
	}
	log.Debugf("Marshalled yaml:\n%s\nWriting to file...", x)

	var f *os.File
	var err1 error

	// creating a file in the name of the context if it does not exist/ opening it to append/modify content if it already exists
	os.MkdirAll(QliksenseConfigHome, os.ModePerm)
	f, err1 = os.OpenFile(filepath.Join(QliksenseConfigHome, contextName+".yaml"), os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err1 != nil {
		panic(err1)
	}

	defer f.Close()

	numBytes, err2 := f.Write(x)
	if err2 != nil {
		panic(err2)
	}
	log.Debugf("wrote %d bytes\n", numBytes)
}

func (qliksenseContext *QliksenseContext) addCommonConfig(contextName string) {
	qliksenseContext.ApiVersion = QliksenseContextApiVersion
	qliksenseContext.Kind = QliksenseContextKind
	qliksenseContext.Metadata.Name = contextName
	qliksenseContext.Metadata.Labels["Version"] = QliksenseContextLabel
	qliksenseContext.Spec.ManifestsRoot = QliksenseContextManifestsRoot
}

func (qliksenseConfig *QliksenseConfig) AddBaseQliksenseConfigs() {
	qliksenseConfig.ApiVersion = QliksenseConfigApiVersion
	qliksenseConfig.Kind = QliksenseConfigKind
	qliksenseConfig.Metadata.Name = 
}
func setOtherConfigs(q *Qliksense) error {
	return nil
}
