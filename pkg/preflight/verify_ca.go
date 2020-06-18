package preflight

import (
	"crypto/tls"
	"flag"
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

	// TODO: parse out server and port frim mongodb url

	// retrieve certs from server
	server := flag.String("server", preflightOpts.MongoOptions.MongodbUrl, "Server to ping")
	port := flag.Uint("port", 27018, "Port that has TLS")
	flag.Parse()

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", *server, *port), &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	conn.Close()

	// Get the ConnectionState struct as that's the one which gives us x509.Certificate struct
	fmt.Printf("length: %d\n", len(conn.ConnectionState().PeerCertificates))
	fmt.Printf("%v\n", conn.ConnectionState().PeerCertificates)

	// TODO: execute verify cmd

	return nil
}
