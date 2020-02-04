package api

import (
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

// NewQConfig create QliksenseConfig object from file ~/.qliksense/config.yaml
func NewQConfig(qsHome string) *QliksenseConfig {
	configFile := filepath.Join(qsHome, "config.yaml")
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("Cannot read config file from: "+configFile, err)
		os.Exit(1)
	}
	qc := &QliksenseConfig{}
	err = yaml.Unmarshal(data, qc)
	if err != nil {
		fmt.Println("yaml unmarshalling error ", err)
		os.Exit(1)
	}
	return qc
}

// GetCR create a QliksenseCR object for a particular context
// from file ~/.qliksense/contexts/<contx-name>/<contx-name>.yaml
func (qc *QliksenseConfig) GetCR(contextName string) (*QliksenseCR, error) {
	crFilePath := ""
	for _, ctx := range qc.Spec.Contexts {
		if ctx.Name == contextName {
			crFilePath = ctx.CrFile
			break
		}
	}
	if crFilePath == "" {
		return nil, errors.New("context name " + contextName + " not found")
	}
	return getCRObject(crFilePath)
}

// GetCurrentCR create a QliksenseCR object for current context
func (qc *QliksenseConfig) GetCurrentCR() (*QliksenseCR, error) {
	return qc.GetCR(qc.Spec.CurrentContext)
}

// SetCrLocation sets the CR location for a context. Helpful during test
func (qc *QliksenseConfig) SetCrLocation(contextName, filepath string) (*QliksenseConfig, error) {
	tempQc := &QliksenseConfig{}
	copier.Copy(tempQc, qc)
	found := false
	tempQc.Spec.Contexts = []Context{}
	for _, c := range qc.Spec.Contexts {
		if c.Name == contextName {
			c.CrFile = filepath
			found = true
		}
		tempQc.Spec.Contexts = append(tempQc.Spec.Contexts, []Context{c}...)
	}
	if found {
		return tempQc, nil
	}
	return nil, errors.New("cannot find the context")
}

func getCRObject(crfile string) (*QliksenseCR, error) {
	data, err := ioutil.ReadFile(crfile)
	if err != nil {
		fmt.Println("Cannot read config file from: "+crfile, err)
		return nil, err
	}
	cr := &QliksenseCR{}
	err = yaml.Unmarshal(data, cr)
	if err != nil {
		fmt.Println("cannot unmarshal cr ", err)
		return nil, err
	}
	return cr, nil
}
