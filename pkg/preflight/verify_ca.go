package preflight

import (
	"fmt"
	"strings"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (qp *QliksensePreflight) VerifyCAChain(kubeConfigContents []byte, namespace string, preflightOpts *PreflightOptions, cleanup bool) error {

	var currentCR *qapi.QliksenseCR
	var err error
	qConfig := qapi.NewQConfig(qp.Q.QliksenseHome)
	qConfig.SetNamespace(namespace)
	currentCR, err = qConfig.GetCurrentCR()
	if err != nil {
		qp.CG.LogVerboseMessage("Unable to retrieve current CR: %v\n", err)
		return err
	}
	decryptedCR, err := qConfig.GetDecryptedCr(currentCR)
	if err != nil {
		qp.CG.LogVerboseMessage("An error occurred while retrieving mongodbUrl from current CR: %v\n", err)
		return err
	}

	// infer mongodb url from CR
	preflightOpts.MongoOptions.MongodbUrl = strings.TrimSpace(decryptedCR.Spec.GetFromSecrets("qliksense", "mongodbUri"))
	fmt.Printf("Mongodb url inferred form CR: %s\n", preflightOpts.MongoOptions.MongodbUrl)
	// retrieve certs from server

	// execute verify cmd
	return nil
}
