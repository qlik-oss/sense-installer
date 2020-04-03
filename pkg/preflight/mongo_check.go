package preflight

import (
	"fmt"
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/api"
)

const (
	mongoImage = "mongo"
)

func (qp *QliksensePreflight) CheckMongo(kubeConfigContents []byte, namespace, mongodbUrl string) error {
	fmt.Printf("Preflight mongodb check: \n")

	if mongodbUrl == "" {
		// TODO: infer this from currentCR
		fmt.Println("MongoDbUri is empty, infer from CR")
	}
	api.LogDebugMessage("mongodbUrl: %s\n", mongodbUrl)
	mongoConnCheck(kubeConfigContents, namespace, mongodbUrl)
	fmt.Println("Completed preflight mongodb check")
	return nil

}

func mongoConnCheck(kubeConfigContents []byte, namespace, mongodbUrl string) error {
	clientset, clientConfig, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("error: unable to create a kubernetes client: %v\n", err)
		fmt.Println(err)
		return err
	}
	// create a pod
	podName := "pf-mongo-pod"
	mongoPod, err := createPreflightTestPod(clientset, namespace, podName, mongoImage)
	if err != nil {
		err = fmt.Errorf("error: unable to create pod : %s\n", podName)
		return err
	}
	defer deletePod(clientset, namespace, podName)

	if err := waitForPod(clientset, namespace, mongoPod); err != nil {
		return err
	}
	if len(mongoPod.Spec.Containers) == 0 {
		err := fmt.Errorf("error: there are no containers in the pod")
		fmt.Println(err)
		return err
	}
	api.LogDebugMessage("Exec-ing into the container...")
	stdout, stderr, err := executeRemoteCommand(clientset, clientConfig, mongoPod.Name, mongoPod.Spec.Containers[0].Name, namespace, []string{"mongo", mongodbUrl})
	if err != nil {
		err = fmt.Errorf("error: unable to execute mongo check in the cluster: %v", err)
		fmt.Println(err)
		return err
	}

	fmt.Println("stdout:", stdout)
	fmt.Println("stderr:", stderr)
	stringToCheck := "Implicit session"
	if strings.Contains(stdout, stringToCheck) || strings.Contains(stderr, stringToCheck) {
		fmt.Println("Preflight mongo check: PASSED")
	} else {
		fmt.Println("Preflight mongo check: FAILED")
	}
	return nil
}
