package preflight

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (qp *QliksensePreflight) VerifyCAChain(kubeConfigContents []byte, namespace string, preflightOpts *PreflightOptions, cleanup bool) error {

	var currentCR *qapi.QliksenseCR
	var err error
	qConfig := qapi.NewQConfig(qp.Q.QliksenseHome)
	qConfig.SetNamespace(namespace)

	fmt.Print("Preflight verify-ca-chain check... ")
	qp.CG.LogVerboseMessage("\n----------------------------------- \n")

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

	// infer ca certs form CR
	caCertificates := strings.TrimSpace(decryptedCR.Spec.GetFromSecrets("qliksense", "caCertificates"))

	fmt.Println("Openssl verify mongodbUrl:")
	// infer mongodb url from CR
	mongodbUrl := strings.TrimSpace(decryptedCR.Spec.GetFromSecrets("qliksense", "mongodbUri"))
	qp.CG.LogVerboseMessage("Mongodb url inferred form CR: %s\n", mongodbUrl)

	// parse out server and port from mongodb url and execute openssl verify
	if err := qp.extractCertAndVerify(mongodbUrl, caCertificates); err != nil {
		return err
	}

	fmt.Printf("\nOpenssl verify discoveryUrl:\n")
	// infer idpConfigs form CR
	idpConfigs := strings.TrimSpace(decryptedCR.Spec.GetFromSecrets("identity-providers", "idpConfigs"))

	data := []map[string]interface{}{}
	if err := json.Unmarshal([]byte(idpConfigs), &data); err != nil {
		panic(err)
	}

	var discoveryUrl string
	for _, idpData := range data {
		discoveryUrl = idpData["discoveryUrl"].(string)
		qp.CG.LogVerboseMessage("Discovery url: %s\n", discoveryUrl)
	}
	if err := qp.extractCertAndVerify(discoveryUrl, caCertificates); err != nil {
		return err
	}

	qp.CG.LogVerboseMessage("Completed preflight verify-ca-chain check\n")
	return nil
}

func (qp *QliksensePreflight) extractCertAndVerify(server string, caCertificates string) error {
	u, err := url.Parse(server)
	if err != nil {
		return fmt.Errorf("unable to parse url: %v", err)
	}

	switch strings.ToLower(u.Scheme) {
	case "http":
		return fmt.Errorf("http url is not supported for this operation")
	case "https":
		if u.Port() == "" {
			u.Host += ":443"
		}
	}

	qp.CG.LogVerboseMessage("Host: %s, port: %s\n", u.Host, u.Port())
	conn, err := tls.Dial("tcp", u.Host, &tls.Config{})
	qp.CG.LogVerboseMessage("Host: %s\n", u.Host)
	if err != nil {
		return fmt.Errorf("failed to connect: " + err.Error())
	}
	defer conn.Close()

	// Get the ConnectionState struct as that's the one which gives us x509.Certificate struct
	x509Certificates := conn.ConnectionState().PeerCertificates

	var serverCert *x509.Certificate
	if len(x509Certificates) == 0 {
		return fmt.Errorf("no server certificates retrieved from the server")
	}
	// we retrieve and verify the server certificate, we ignore intermediate certificates at this point.
	for _, x509Cert := range x509Certificates {
		if !x509Cert.IsCA {
			serverCert = x509Cert
			break
		}
	}
	if serverCert == nil {
		return fmt.Errorf("no valid server certificates retrieved from the server")
	}
	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM([]byte(caCertificates)); !ok {
		return fmt.Errorf("failed to parse root certificate.")
	}

	opts := x509.VerifyOptions{
		Roots:   roots,
		DNSName: u.Hostname(),
	}
	if _, err := serverCert.Verify(opts); err != nil {
		return fmt.Errorf("failed to verify certificate: " + err.Error())
	}
	return nil
}
