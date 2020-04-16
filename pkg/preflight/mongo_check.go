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
	fmt.Printf("Preflight mongodb check: \n")

	if preflightOpts.MongoOptions.MongodbUrl == "" {
		// infer mongoDbUrl from currentCR
		fmt.Println("MongoDbUri is empty, infer from CR")
		qConfig := qapi.NewQConfig(qp.Q.QliksenseHome)
		var currentCR *qapi.QliksenseCR

		var err error
		qConfig.SetNamespace(namespace)
		currentCR, err = qConfig.GetCurrentCR()
		if err != nil {
			fmt.Printf("Unable to retrieve current CR: %v\n", err)
			return err
		}
		decryptedCR, err := qConfig.GetDecryptedCr(currentCR)
		if err != nil {
			fmt.Printf("An error occurred while retrieving mongodbUrl from current CR: %v\n", err)
			return err
		}
		preflightOpts.MongoOptions.MongodbUrl = decryptedCR.Spec.GetFromSecrets("qliksense", "mongoDbUri")
	}

	fmt.Printf("mongodbUrl: %s\n", preflightOpts.MongoOptions.MongodbUrl)
	if err := qp.mongoConnCheck(kubeConfigContents, namespace, preflightOpts); err != nil {
		return err
	}
	fmt.Println("Completed preflight mongodb check")
	return nil
}

func (qp *QliksensePreflight) mongoConnCheck(kubeConfigContents []byte, namespace string, preflightOpts *PreflightOptions) error {
	var caCertSecretName, clientCertSecretName string
	clientset, _, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("error: unable to create a kubernetes client: %v\n", err)
		fmt.Println(err)
		return err
	}
	var secrets []string
	if preflightOpts.MongoOptions.CaCertFile != "" {
		caCertSecretName = "preflight-mongo-test-cacert"
		caCertSecret, err := createSecret(clientset, namespace, preflightOpts.MongoOptions.CaCertFile, caCertSecretName)

		if err != nil {
			err = fmt.Errorf("error: unable to create a create ca cert kubernetes secret: %v\n", err)
			fmt.Println(err)
			return err
		}

		defer deleteK8sSecret(clientset, namespace, caCertSecret)
		secrets = append(secrets, caCertSecretName)
	}
	if preflightOpts.MongoOptions.ClientCertFile != "" {
		clientCertSecretName = "preflight-mongo-test-clientcert"
		clientCertSecret, err := createSecret(clientset, namespace, preflightOpts.MongoOptions.ClientCertFile, clientCertSecretName)
		if err != nil {
			err = fmt.Errorf("error: unable to create a create client cert kubernetes secret: %v\n", err)
			fmt.Println(err)
			return err
		}

		defer deleteK8sSecret(clientset, namespace, clientCertSecret)
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
	mongoCommand.WriteString(fmt.Sprintf(" --tlsCAFile=/etc/ssl/%s/%[1]s", caCertSecretName))
	if preflightOpts.MongoOptions.CaCertFile != "" {
		api.LogDebugMessage("Adding caCertFile:  Mongo command: %s\n", mongoCommand.String())
	}
	if preflightOpts.MongoOptions.ClientCertFile != "" {
		mongoCommand.WriteString(fmt.Sprintf(" --tlsCertificateKeyFile=/etc/ssl/%s/%[1]s", clientCertSecretName))
		api.LogDebugMessage("Adding clientCertFile:  Mongo command: %s\n", mongoCommand.String())
	}
	mongoCommand.WriteString(` --eval "print(\"connected to mongo\")"`)

	commandToRun := []string{"sh", "-c", mongoCommand.String()}
	api.LogDebugMessage("Mongo commandToRun: %s\n", strings.Join(commandToRun, " "))

	// create a pod
	podName := "pf-mongo-pod"
	imageName, err := qp.GetPreflightConfigObj().GetImageName(mongo, true)
	if err != nil {
		return err
	}
	mongoPod, err := createPreflightTestPod(clientset, namespace, podName, imageName, secrets, commandToRun)
	if err != nil {
		err = fmt.Errorf("error: unable to create pod : %v\n", err)
		return err
	}
	defer deletePod(clientset, namespace, podName)

	if err := waitForPod(clientset, namespace, mongoPod); err != nil {
		return err
	}
	if len(mongoPod.Spec.Containers) == 0 {
		err := fmt.Errorf("error: there are no containers in the pod- %v\n", err)
		fmt.Println(err)
		return err
	}
	waitForPodToDie(clientset, namespace, mongoPod)
	logStr, err := getPodLogs(clientset, mongoPod)
	if err != nil {
		err = fmt.Errorf("error: unable to execute mongo check in the cluster: %v\n", err)
		fmt.Println(err)
		return err
	}

	stringToCheck := "Implicit session:"
	if strings.Contains(logStr, stringToCheck) {
		fmt.Println("Preflight mongo check: PASSED")
	} else {
		err = fmt.Errorf("Expected response not found\n")
		return err
	}
	return nil
}

func createSecret(clientset *kubernetes.Clientset, namespace, certFile, certSecretName string) (*apiv1.Secret, error) {
	certBytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	certSecret, err := createPreflightTestSecret(clientset, namespace, certSecretName, certBytes)
	if err != nil {
		err = fmt.Errorf("error: unable to create secret with ca cert : %v\n", err)
		return nil, err
	}
	return certSecret, nil
}
