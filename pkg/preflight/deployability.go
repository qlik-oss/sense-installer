package preflight

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
)

func (qp *QliksensePreflight) CheckDeployment(namespace string, kubeConfigContents []byte) error {
	clientset, _, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("Kube config error: %v\n", err)
		fmt.Print(err)
		return err
	}

	// Deployment check
	fmt.Printf("Preflight deployment check: \n")
	err = qp.checkPfDeployment(clientset, namespace, "deployment-preflight-check")
	if err != nil {
		fmt.Println("Preflight Deployment check: FAILED")
		return err
	}
	fmt.Println("Completed preflight deployment check")

	return nil
}

func (qp *QliksensePreflight) CheckService(namespace string, kubeConfigContents []byte) error {
	clientset, _, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("error: unable to create a kubernetes client: %v\n", err)
		fmt.Println(err)
		return err
	}
	// Service check
	fmt.Printf("\nPreflight service check: \n")
	err = checkPfService(clientset, namespace)
	if err != nil {
		fmt.Println("Preflight Service check: FAILED")
		return err
	}
	fmt.Println("Completed preflight service check")
	return nil
}

func (qp *QliksensePreflight) CheckPod(namespace string, kubeConfigContents []byte) error {
	clientset, _, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("error: unable to create a kubernetes client: %v\n", err)
		fmt.Print(err)
		return err
	}
	// Pod check
	fmt.Printf("\nPreflight pod check: \n")

	err = qp.checkPfPod(clientset, namespace)
	if err != nil {
		fmt.Println("Preflight Pod check: FAILED")
		return err
	}
	fmt.Println("Completed preflight pod check")
	return nil
}

func (qp *QliksensePreflight) checkPfPod(clientset *kubernetes.Clientset, namespace string) error {
	// create a pod
	podName := "pod-pf-check"
	pod, err := createPreflightTestPod(clientset, namespace, podName, qp.GetPreflightConfigObj().GetImageName(nginx))
	if err != nil {
		err = fmt.Errorf("error: unable to create pod %s - %v\n", podName, err)
		return err
	}
	defer deletePod(clientset, namespace, podName)

	if err := waitForPod(clientset, namespace, pod); err != nil {
		return err
	}

	fmt.Println("Preflight pod creation check: PASSED")
	fmt.Println("Cleaning up resources...")
	return nil
}

func checkPfService(clientset *kubernetes.Clientset, namespace string) error {
	// creating service
	serviceName := "svc-pf-check"
	pfService, err := createPreflightTestService(clientset, namespace, serviceName)
	if err != nil {
		err = fmt.Errorf("error: unable to create service : %s\n", serviceName)
		return err
	}
	defer deleteService(clientset, namespace, serviceName)
	_, err = getService(clientset, namespace, pfService.GetName())
	if err != nil {
		err = fmt.Errorf("error: unable to retrieve service: %s\n", serviceName)
		return err
	}
	fmt.Println("Preflight service creation check: PASSED")
	fmt.Println("Cleaning up resources...")
	return nil
}

func (qp *QliksensePreflight) checkPfDeployment(clientset *kubernetes.Clientset, namespace, depName string) error {
	// check if we are able to create a deployment
	pfDeployment, err := createPreflightTestDeployment(clientset, namespace, depName, qp.GetPreflightConfigObj().GetImageName(nginx))
	if err != nil {
		err = fmt.Errorf("error: unable to create deployment: %v\n", err)
		return err
	}
	defer deleteDeployment(clientset, namespace, depName)
	if err := waitForDeployment(clientset, namespace, pfDeployment); err != nil {
		return err
	}
	fmt.Println("Preflight Deployment check: PASSED")
	fmt.Println("Cleaning up resources...")
	return nil
}
