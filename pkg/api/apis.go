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

func (qc *QliksenseConfig) GetCR(contextName string) (*QliksenseCR, error) {
	crFilePath := ""
	for _, ctx := range qc.Spec.Contexts {
		if ctx.Name == contextName {
			crFilePath = ctx.CRLocation
			break
		}
	}
	if crFilePath == "" {
		return nil, errors.New("context name " + contextName + " not found")
	}
	return getCRObject(crFilePath)
}

func (qc *QliksenseConfig) GetCurrentCR() (*QliksenseCR, error) {
	return qc.GetCR(qc.Spec.CurrentContext)
}

func (qc *QliksenseConfig) SetCrLocation(contextName, location string) (*QliksenseConfig, error) {
	tempQc := &QliksenseConfig{}
	copier.Copy(tempQc, qc)
	found := false
	tempQc.Spec.Contexts = []Context{}
	for _, c := range qc.Spec.Contexts {
		if c.Name == contextName {
			c.CRLocation = location
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
	yaml.Unmarshal(data, cr)
	return cr, nil
}
