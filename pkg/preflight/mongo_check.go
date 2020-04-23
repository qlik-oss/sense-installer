package preflight

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/api"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	mongo = "mongo"
)

func (qp *QliksensePreflight) CheckMongo(kubeConfigContents []byte, namespace string, preflightOpts *PreflightOptions) error {
	qp.P.LogVerboseMessage("Preflight mongodb check: \n")
	qp.P.LogVerboseMessage("------------------------ \n")

	if preflightOpts.MongoOptions.MongodbUrl == "" {
		// infer mongoDbUrl from currentCR
		qp.P.LogVerboseMessage("MongoDbUri is empty, infer from CR\n")
		qConfig := qapi.NewQConfig(qp.Q.QliksenseHome)
		var currentCR *qapi.QliksenseCR

		var err error
		qConfig.SetNamespace(namespace)
		currentCR, err = qConfig.GetCurrentCR()
		if err != nil {
			qp.P.LogVerboseMessage("Unable to retrieve current CR: %v\n", err)
			return err
		}
		decryptedCR, err := qConfig.GetDecryptedCr(currentCR)
		if err != nil {
			qp.P.LogVerboseMessage("An error occurred while retrieving mongodbUrl from current CR: %v\n", err)
			return err
		}
		preflightOpts.MongoOptions.MongodbUrl = decryptedCR.Spec.GetFromSecrets("qliksense", "mongoDbUri")
	}

	qp.P.LogVerboseMessage("MongodbUrl: %s\n", preflightOpts.MongoOptions.MongodbUrl)
	if err := qp.mongoConnCheck(kubeConfigContents, namespace, preflightOpts); err != nil {
		return err
	}
	qp.P.LogVerboseMessage("Completed preflight mongodb check\n")
	return nil
}

func (qp *QliksensePreflight) mongoConnCheck(kubeConfigContents []byte, namespace string, preflightOpts *PreflightOptions) error {
	var caCertSecretName, clientCertSecretName string
	clientset, _, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("unable to create a kubernetes client: %v\n", err)
		return err
	}
	caCertSecretName = "preflight-mongo-test-cacert"
	clientCertSecretName = "preflight-mongo-test-clientcert"
	podName := "pf-mongo-pod"

	// cleanup before starting check
	qp.runMongoCleanup(clientset, namespace, podName, caCertSecretName, clientCertSecretName)

	var secrets []string
	if preflightOpts.MongoOptions.CaCertFile != "" {
		caCertSecret, err := qp.createSecret(clientset, namespace, preflightOpts.MongoOptions.CaCertFile, caCertSecretName)
		if err != nil {
			err = fmt.Errorf("unable to create a ca cert kubernetes secret: %v\n", err)
			return err
		}

		defer qp.deleteK8sSecret(clientset, namespace, caCertSecret.Name)
		secrets = append(secrets, caCertSecretName)
	}
	if preflightOpts.MongoOptions.ClientCertFile != "" {
		clientCertSecret, err := qp.createSecret(clientset, namespace, preflightOpts.MongoOptions.ClientCertFile, clientCertSecretName)
		if err != nil {
			err = fmt.Errorf("unable to create a client cert kubernetes secret: %v\n", err)
			return err
		}

		defer qp.deleteK8sSecret(clientset, namespace, clientCertSecret.Name)
		secrets = append(secrets, clientCertSecretName)
	}

	mongoCommand := strings.Builder{}
	mongoCommand.WriteString(fmt.Sprintf("sleep 10;mongo %s", preflightOpts.MongoOptions.MongodbUrl))
	if preflightOpts.MongoOptions.Username != "" {
		mongoCommand.WriteString(fmt.Sprintf(" --username %s", preflightOpts.MongoOptions.Username))
		api.LogDebugMessage("Adding username: Mongo command: %s\n", mongoCommand.String())
	}
	if preflightOpts.MongoOptions.Password != "" {
		mongoCommand.WriteString(fmt.Sprintf(" --password %s", preflightOpts.MongoOptions.Password))
		api.LogDebugMessage("Adding username and password\n")
	}
	if preflightOpts.MongoOptions.Tls || preflightOpts.MongoOptions.CaCertFile != "" || preflightOpts.MongoOptions.ClientCertFile != "" {
		mongoCommand.WriteString(" --tls")
		api.LogDebugMessage("Adding --tls: Mongo command: %s\n", mongoCommand.String())
	}
	if preflightOpts.MongoOptions.CaCertFile != "" {
		mongoCommand.WriteString(fmt.Sprintf(" --tlsCAFile=/etc/ssl/%s/%[1]s", caCertSecretName))
		api.LogDebugMessage("Adding caCertFile:  Mongo command: %s\n", mongoCommand.String())
	}
	if preflightOpts.MongoOptions.ClientCertFile != "" {
		mongoCommand.WriteString(fmt.Sprintf(" --tlsCertificateKeyFile=/etc/ssl/%s/%[1]s", clientCertSecretName))
		api.LogDebugMessage("Adding clientCertFile:  Mongo command: %s\n", mongoCommand.String())
	}
	mongoCommand.WriteString(` --eval "print(\"connected to mongo\")"`)

	commandToRun := []string{"sh", "-c", mongoCommand.String()}
	api.LogDebugMessage("Mongo command: %s\n", strings.Join(commandToRun, " "))

	// create a pod
	imageName, err := qp.GetPreflightConfigObj().GetImageName(mongo, true)
	if err != nil {
		err = fmt.Errorf("unable to retrieve image : %v\n", err)
		return err
	}
	mongoPod, err := qp.createPreflightTestPod(clientset, namespace, podName, imageName, secrets, commandToRun)
	if err != nil {
		err = fmt.Errorf("unable to create pod : %v\n", err)
		return err
	}
	defer qp.deletePod(clientset, namespace, podName)

	if err := waitForPod(clientset, namespace, mongoPod); err != nil {
		return err
	}
	if len(mongoPod.Spec.Containers) == 0 {
		err := fmt.Errorf("there are no containers in the pod- %v\n", err)
		return err
	}
	waitForPodToDie(clientset, namespace, mongoPod)
	logStr, err := getPodLogs(clientset, mongoPod)
	if err != nil {
		err = fmt.Errorf("unable to execute mongo check in the cluster: %v\n", err)
		return err
	}

	stringToCheck := "Implicit session:"
	if strings.Contains(logStr, stringToCheck) {
		qp.P.LogVerboseMessage("Preflight mongo check: PASSED\n")
	} else {
		err = fmt.Errorf("Connection failed: %s\n", logStr)
		return err
	}
	return nil
}

func (qp *QliksensePreflight) createSecret(clientset *kubernetes.Clientset, namespace, certFile, certSecretName string) (*apiv1.Secret, error) {
	certBytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	certSecret, err := qp.createPreflightTestSecret(clientset, namespace, certSecretName, certBytes)
	if err != nil {
		err = fmt.Errorf("unable to create secret with ca cert : %v\n", err)
		return nil, err
	}
	return certSecret, nil
}

func (qp *QliksensePreflight) runMongoCleanup(clientset *kubernetes.Clientset, namespace, podName, caCertSecretName, clientCertSecretName string) {
	qp.deleteK8sSecret(clientset, namespace, caCertSecretName)
	qp.deleteK8sSecret(clientset, namespace, clientCertSecretName)
	qp.deletePod(clientset, namespace, podName)
}
