package api

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/qlik-oss/k-apis/pkg/config"
	"k8s.io/apimachinery/pkg/runtime/schema"
	machine_yaml "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	QliksenseConfigApiVersion = "v1"
	QliksenseConfigApiGroup   = "config.qlik.com"
	QliksenseConfigKind       = "QliksenseConfig"

	QliksenseApiVersion     = "v1"
	QliksenseKind           = "Qliksense"
	QliksenseGroup          = "qlik.com"
	QliksenseDefaultProfile = "docker-desktop"
	DefaultRotateKeys       = "yes"
	QliksenseMetadataName   = "QliksenseConfigMetadata"
	DefaultMongoDbUri       = "mongodb://qlik-default-mongodb:27017/qliksense?ssl=false"
	DefaultMongoDbUriKey    = "mongoDbUri"
)

// AddCommonConfig adds common configs into CRs
func (qliksenseCR *QliksenseCR) AddCommonConfig(contextName string) {
	qliksenseCR.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   QliksenseGroup,
		Kind:    QliksenseKind,
		Version: QliksenseApiVersion,
	})
	qliksenseCR.SetName(contextName)
	qliksenseCR.Spec = &config.CRSpec{
		Profile:    QliksenseDefaultProfile,
		RotateKeys: DefaultRotateKeys,
	}
	qliksenseCR.Spec.AddToSecrets("qliksense", DefaultMongoDbUriKey, DefaultMongoDbUri, "")
}

// AddBaseQliksenseConfigs adds configs into config.yaml
func (qliksenseConfig *QliksenseConfig) AddBaseQliksenseConfigs(defaultQliksenseContext string) {
	qliksenseConfig.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   QliksenseConfigApiGroup,
		Kind:    QliksenseConfigKind,
		Version: QliksenseConfigApiVersion,
	})
	qliksenseConfig.SetName(QliksenseMetadataName)
	if defaultQliksenseContext != "" {
		if qliksenseConfig.Spec == nil {
			qliksenseConfig.Spec = &ContextSpec{}
		}
		qliksenseConfig.Spec.CurrentContext = defaultQliksenseContext
	}
}

func (qliksenseConfig *QliksenseConfig) SwitchCurrentCRToVersionAndProfile(version, profile string) error {
	if qcr, err := qliksenseConfig.GetCurrentCR(); err != nil {
		return err
	} else {
		versionManifestRoot := qliksenseConfig.BuildCurrentManifestsRoot(version)
		if (qcr.Spec.ManifestsRoot != versionManifestRoot) || (profile != "" && qcr.Spec.Profile != profile) || (qcr.GetLabelFromCr("version") != version) {
			qcr.Spec.ManifestsRoot = versionManifestRoot
			if profile != "" {
				qcr.Spec.Profile = profile
			}
			qcr.AddLabelToCr("version", version)
			if err := qliksenseConfig.WriteCurrentContextCR(qcr); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteToFile (content, targetFile) writes content into specified file
func WriteToFile(content interface{}, targetFile string) error {
	if content == nil || targetFile == "" {
		return nil
	}

	x, err := K8sToYaml(content)
	if err != nil {
		err = fmt.Errorf("An error occurred during marshalling CR: %v", err)
		log.Println(err)
		return err
	}
	// Writing content
	err = ioutil.WriteFile(targetFile, x, 0644)
	if err != nil {
		log.Println(err)
		return err
	}
	LogDebugMessage("Wrote content into %s", targetFile)
	return nil
}

// ReadFromFile (content, targetFile) reads content from specified sourcefile
func ReadFromFile(content interface{}, sourceFile string) error {
	if content == nil || sourceFile == "" {
		return nil
	}
	file, e := os.Open(sourceFile)
	if e != nil {
		return e
	}
	return ReadFromStream(content, file)
}

// ReadFromStream reads from input stream and creat yaml struct of type content
func ReadFromStream(content interface{}, reader io.Reader) error {
	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		err = fmt.Errorf("There was an error reading from reader: %v", err)
		return err
	}
	// reading k8s style object
	// https://stackoverflow.com/questions/44306554/how-to-deserialize-kubernetes-yaml-file
	dec := machine_yaml.NewYAMLOrJSONDecoder(bytes.NewReader(contents), 10000)
	return dec.Decode(content)
}
