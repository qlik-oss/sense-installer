package preflight

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"github.com/qlik-oss/sense-installer/pkg/api"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	preflight_mongo = "preflight-mongo"
	caCertMountPath = "/etc/ssl/certs/ca-certificates.crt"
)

func (qp *QliksensePreflight) CheckMongo(kubeConfigContents []byte, namespace string, preflightOpts *PreflightOptions, cleanup bool) error {
	if !cleanup {
		qp.P.LogVerboseMessage("Preflight mongodb check: \n")
		qp.P.LogVerboseMessage("------------------------ \n")
	}
	var currentCR *qapi.QliksenseCR
	var err error
	qConfig := qapi.NewQConfig(qp.Q.QliksenseHome)
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
	if preflightOpts.MongoOptions.MongodbUrl == "" && !cleanup {
		// infer mongoDbUrl from currentCR
		qp.P.LogVerboseMessage("mongodbUri is empty, infer from CR\n")
		preflightOpts.MongoOptions.MongodbUrl = strings.TrimSpace(decryptedCR.Spec.GetFromSecrets("qliksense", "mongodbUri"))
	}

	if preflightOpts.MongoOptions.CaCertFile == "" && !cleanup {
		caCertStr := decryptedCR.Spec.GetFromSecrets("qliksense", "caCertificates")
		tmpDir := os.TempDir()
		caCrtFile := filepath.Join(tmpDir, "rootCA.crt")
		api.LogDebugMessage("received ca crt: %s\n", caCertStr)
		if err := ioutil.WriteFile(caCrtFile, []byte(caCertStr), 0644); err != nil {
			return fmt.Errorf("unable to write CA crt to file: %v", err)
		}
		preflightOpts.MongoOptions.CaCertFile = caCrtFile
	}

	if !cleanup {
		qp.P.LogVerboseMessage("MongodbUrl: %s\n", preflightOpts.MongoOptions.MongodbUrl)

		// if mongoDbUrl is empty, abort check
		if preflightOpts.MongoOptions.MongodbUrl == "" {
			qp.P.LogVerboseMessage("Mongodb Url is empty, hence aborting preflight check\n")
			return errors.New("MongoDbUrl is empty")
		}
	}

	if err := qp.mongoConnCheck(kubeConfigContents, namespace, preflightOpts, cleanup); err != nil {
		return err
	}

	if !cleanup {
		qp.P.LogVerboseMessage("Completed preflight mongodb check\n")
	}
	return nil
}

func (qp *QliksensePreflight) mongoConnCheck(kubeConfigContents []byte, namespace string, preflightOpts *PreflightOptions, cleanup bool) error {
	caCertSecretName := "ca-certificates-crt"
	mongoPodName := "pf-mongo-pod"
	clientset, _, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("unable to create a kubernetes client: %v\n", err)
		return err
	}

	// cleanup before starting check
	qp.runMongoCleanup(clientset, namespace, mongoPodName, caCertSecretName)
	if cleanup {
		return nil
	}
	secrets := map[string]string{}
	if preflightOpts.MongoOptions.CaCertFile != "" {
		caCertSecret, err := qp.createSecret(clientset, namespace, preflightOpts.MongoOptions.CaCertFile, caCertSecretName)
		if err != nil {
			err = fmt.Errorf("unable to create a ca cert kubernetes secret: %v\n", err)
			return err
		}

		defer qp.deleteK8sSecret(clientset, namespace, caCertSecret.Name)
		secrets[caCertSecretName] = caCertMountPath
	}

	commandToRun := []string{"./preflight-mongo", fmt.Sprintf(`-url="%s"`, preflightOpts.MongoOptions.MongodbUrl)}
	api.LogDebugMessage("Mongo command: %s\n", strings.Join(commandToRun, " "))

	// create a pod
	imageName, err := qp.GetPreflightConfigObj().GetImageName(preflight_mongo, true)
	if err != nil {
		err = fmt.Errorf("unable to retrieve image : %v\n", err)
		return err
	}
	api.LogDebugMessage("image name to be used: %s\n", imageName)
	mongoPod, err := qp.createPreflightTestPod(clientset, namespace, mongoPodName, imageName, secrets, commandToRun)
	if err != nil {
		err = fmt.Errorf("unable to create pod : %v\n", err)
		return err
	}
	defer qp.deletePod(clientset, namespace, mongoPodName)

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

	// check mongo server version
	ok, err := qp.checkMongoVersion(logStr)
	if !ok || err != nil {
		return err
	}

	// check if connection succeeded
	stringToCheck := "qlik - connection succeeded!!"
	if strings.Contains(logStr, stringToCheck) {
		qp.P.LogVerboseMessage("Preflight mongo check: PASSED\n")
	} else {
		err = fmt.Errorf("Connection failed: %s\n", logStr)
		return err
	}
	return nil
}

func (qp *QliksensePreflight) checkMongoVersion(logStr string) (bool, error) {
	// check mongo server version
	api.LogDebugMessage("Minimum required mongo version: %s\n", qp.GetPreflightConfigObj().GetMinMongoVersion())
	mongoVersionStrToCheck := "qlik mongo server version:"
	if strings.Contains(logStr, mongoVersionStrToCheck) {
		logLines := strings.Split(logStr, "\n")
		for _, eachline := range logLines {
			if strings.Contains(eachline, mongoVersionStrToCheck) {
				mongoVersionLog := strings.Split(eachline, ":")
				if len(mongoVersionLog) < 2 {
					continue
				}
				mongoVersionStr := strings.ReplaceAll(strings.TrimSpace(mongoVersionLog[1]), `"`, "")
				api.LogDebugMessage("Extracted mongo version from pod log: %s\n", mongoVersionStr)
				currentMongoVersionSemver, err := semver.NewVersion(mongoVersionStr)
				if err != nil {
					err = fmt.Errorf("Unable to convert minimum mongo version into semver version:%v\n", err)
					return false, err
				}
				minMongoVersionSemver, err := semver.NewVersion(qp.GetPreflightConfigObj().GetMinMongoVersion())
				if err != nil {
					err = fmt.Errorf("Unable to convert required minimum mongo version into semver version:%v\n", err)
					return false, err
				}
				if currentMongoVersionSemver.GreaterThan(minMongoVersionSemver) || currentMongoVersionSemver.Equal(minMongoVersionSemver) {
					qp.P.LogVerboseMessage("Current mongodb server version %s is greater than or equal to minimum required mongodb version: %s\n", currentMongoVersionSemver, minMongoVersionSemver)
					return true, nil
				}
				err = fmt.Errorf("Current mongodb server version %s is less than minimum required mongodb version: %s", currentMongoVersionSemver, minMongoVersionSemver)
				return false, err
			}
		}
	}
	err := errors.New("Unable to infer mongodb server version")
	return false, err
}

func (qp *QliksensePreflight) createSecret(clientset *kubernetes.Clientset, namespace, certFile, certSecretName string) (*apiv1.Secret, error) {
	certBytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	certSecret, err := qp.createPreflightTestSecret(clientset, namespace, certSecretName, certBytes)
	if err != nil {
		err = fmt.Errorf("unable to create secret with cert : %v\n", err)
		return nil, err
	}
	return certSecret, nil
}

func (qp *QliksensePreflight) runMongoCleanup(clientset *kubernetes.Clientset, namespace, mongoPodName, caCertSecretName string) {
	qp.deletePod(clientset, namespace, mongoPodName)
	qp.deleteK8sSecret(clientset, namespace, caCertSecretName)
}
