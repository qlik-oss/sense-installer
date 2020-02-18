package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/qlik-oss/k-apis/pkg/config"
	"gopkg.in/yaml.v2"
)

const (
	QliksenseConfigApiVersion  = "config.qlik.com/v1"
	QliksenseConfigKind        = "QliksenseConfig"
	QliksenseContextApiVersion = "qlik.com/v1"
	QliksenseContextKind       = "Qliksense"
	QliksenseDefaultProfile    = "docker-desktop"
	DefaultRotateKeys          = "yes"
	QliksenseMetadataName      = "QliksenseConfigMetadata"
)

// AddCommonConfig adds common configs into CRs
func (qliksenseCR *QliksenseCR) AddCommonConfig(contextName string) {
	qliksenseCR.ApiVersion = QliksenseContextApiVersion
	qliksenseCR.Kind = QliksenseContextKind
	if qliksenseCR.Metadata == nil {
		qliksenseCR.Metadata = &Metadata{}
	}
	if qliksenseCR.Metadata.Name == "" {
		qliksenseCR.Metadata.Name = contextName
	}
	qliksenseCR.Spec = &config.CRSpec{}
	qliksenseCR.Spec.Profile = QliksenseDefaultProfile
	qliksenseCR.Spec.ReleaseName = contextName
	qliksenseCR.Spec.RotateKeys = DefaultRotateKeys
}

// AddBaseQliksenseConfigs adds configs into config.yaml
func (qliksenseConfig *QliksenseConfig) AddBaseQliksenseConfigs(defaultQliksenseContext string) {
	qliksenseConfig.ApiVersion = QliksenseConfigApiVersion
	qliksenseConfig.Kind = QliksenseConfigKind
	if qliksenseConfig.Metadata == nil {
		qliksenseConfig.Metadata = &Metadata{}
	}
	qliksenseConfig.Metadata.Name = QliksenseMetadataName
	if defaultQliksenseContext != "" {
		if qliksenseConfig.Spec == nil {
			qliksenseConfig.Spec = &ContextSpec{}
		}
		qliksenseConfig.Spec.CurrentContext = defaultQliksenseContext
	}
}

// WriteToFile (QliksenseCR, QliksenseConfig) writes content into specified file
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

// ReadFromFile (QliksenseCR, QliksenseConfig) reads content from specified sourcefile
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
