package qliksense

import (
	"fmt"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"gopkg.in/yaml.v2"
)

func ConfigApplyQK8s(q *Qliksense) error {

	//get the current context cr
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	}
	crByte, err := yaml.Marshal(qcr)
	if err != nil {
		fmt.Println("cannnot marshal CR", err)
		return err
	}

	return nil
}
