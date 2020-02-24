package qliksense

import (
	"errors"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (q *Qliksense) UninstallQK8s(contextName string) error {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	if contextName == "" {
		contextName = qConfig.Spec.CurrentContext
	} else if !qConfig.IsContextExist(contextName) {
		return errors.New("context name [ " + contextName + " ] not found")
	}
	cr, err := qConfig.GetCurrentCR()
	if err != nil {
		return err
	}
	str, err := q.getCRString(contextName)
	if err != nil {
		return err
	}
	return qapi.KubectlDelete(str, cr.Spec.NameSpace)
}
