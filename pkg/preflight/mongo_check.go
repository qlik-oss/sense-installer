package preflight

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/api"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	apiv1 "k8s.io/api/core/v1"
)

const (
	mongo = "mongo"
)

func (qp *QliksensePreflight) CheckMongo(kubeConfigContents []byte, namespace, mongodbUrl string, tls bool, username, password, caCertFile, clientCertFile string) error {
	fmt.Printf("Preflight mongodb check: \n")

	if mongodbUrl == "" {
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
		mongodbUrl = decryptedCR.Spec.GetFromSecrets("qliksense", "mongoDbUri")
	}

	fmt.Printf("mongodbUrl: %s\n", mongodbUrl)
	if err := qp.mongoConnCheck(kubeConfigContents, namespace, mongodbUrl, tls, username, password, caCertFile, clientCertFile); err != nil {
		return err
	}
	fmt.Println("Completed preflight mongodb check")
	return nil
}

func (qp *QliksensePreflight) mongoConnCheck(kubeConfigContents []byte, namespace, mongodbUrl string, tls bool, username, password, caCertFile, clientCertFile string) error {
	var caCertSecret, clientSecret *apiv1.Secret
	var caCertSecretName, clientCertSecretName string
	clientset, _, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("error: unable to create a kubernetes client: %v\n", err)
		fmt.Println(err)
		return err
	}
	var secrets []string
	if caCertFile != "" {
		caCertBytes, err := ioutil.ReadFile(caCertFile)
		if err != nil {
			return err
		}
		caCertSecretName = "preflight-mongo-test-cacert"
		caCertSecret, err = createPreflightTestSecret(clientset, namespace, caCertSecretName, caCertBytes)
		if err != nil {
			err = fmt.Errorf("error: unable to create secret with ca cert : %v\n", err)
			return err
		}
		defer deleteK8sSecret(clientset, namespace, caCertSecret)
		secrets = append(secrets, caCertSecretName)
	}
	if clientCertFile != "" {
		clientCertBytes, err := ioutil.ReadFile(clientCertFile)
		if err != nil {
			return err
		}
		clientCertSecretName = "preflight-mongo-test-clientcert"
		clientSecret, err = createPreflightTestSecret(clientset, namespace, clientCertSecretName, clientCertBytes)
		if err != nil {
			err = fmt.Errorf("error: unable to create secret with client cert : %v\n", err)
			return err
		}
		defer deleteK8sSecret(clientset, namespace, clientSecret)
		secrets = append(secrets, clientCertSecretName)
	}

	mongoCommand := strings.Builder{}
	mongoCommand.WriteString(fmt.Sprintf("sleep 10;mongo %s", mongodbUrl))
	if username != "" {
		mongoCommand.WriteString(fmt.Sprintf(" --username %s", username))
		api.LogDebugMessage("Adding username: Mongo command: %s\n", mongoCommand.String())
	}
	if password != "" {
		mongoCommand.WriteString(fmt.Sprintf(" --password %s", password))
		api.LogDebugMessage("Adding username and password\n")
	}
	if tls || caCertFile != "" || clientCertFile != "" {
		mongoCommand.WriteString(" --tls")
		api.LogDebugMessage("Adding --tls: Mongo command: %s\n", mongoCommand.String())
	}
	if caCertFile != "" {
		mongoCommand.WriteString(fmt.Sprintf(" --tlsCAFile=/etc/ssl/%s/%[1]s", caCertSecretName))
		api.LogDebugMessage("Adding caCertFile:  Mongo command: %s\n", mongoCommand.String())
	}
	if clientCertFile != "" {
		mongoCommand.WriteString(fmt.Sprintf(" --tlsCertificateKeyFile=/etc/ssl/%s/%[1]s", clientCertSecretName))
		api.LogDebugMessage("Adding clientCertFile:  Mongo command: %s\n", mongoCommand.String())
	}

	commandToRun := []string{"sh", "-c", mongoCommand.String()}
	api.LogDebugMessage("Mongo commandToRun: %s\n", strings.Join(commandToRun, " "))

	// create a pod
	podName := "pf-mongo-pod"

	mongoPod, err := createPreflightTestPod(clientset, namespace, podName, qp.GetPreflightConfigObj().GetImageName(mongo), secrets, commandToRun)
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
