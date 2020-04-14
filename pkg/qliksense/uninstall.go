package qliksense

import (
	"errors"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (q *Qliksense) UninstallQK8s(contextName string, flag string) error {
	ans := flag

	if ans == "false" {
		if i := AskForConfirmation("Are You Sure? "); i == true {
			ans = "true"
		}
	}
	if ans == "true" {
		qConfig := qapi.NewQConfig(q.QliksenseHome)
		if contextName == "" {
			contextName = qConfig.Spec.CurrentContext
		} else if !qConfig.IsContextExist(contextName) {
			return errors.New("context name [ " + contextName + " ] not found")
		}
		str, err := q.getCRString(contextName)
		if err != nil {
			return err
		}
		return qapi.KubectlDelete(str, "")
	}
	return nil
}
