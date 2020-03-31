package preflight

import (
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"
)

func (qp *QliksensePreflight) CheckDeploy(namespace string, kubeConfigContents []byte) error {
	checkCount := 0
	clientset, _, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("Kube config error: %v\n", err)
		fmt.Print(err)
		return err
	}

	// Deployment check
	fmt.Printf("Preflight deployment check: \n")
	err = checkPfDeployment(clientset, namespace)
	if err != nil {
		fmt.Println("Preflight Deployment check: FAILED")
		return err
	}
	checkCount++
	fmt.Println("Completed preflight deployment check")

	// Service check
	fmt.Printf("\nPreflight service check: \n")
	err = checkPfService(clientset, namespace)
	if err != nil {
		fmt.Println("Preflight Service check: FAILED")
		return err
	}
	checkCount++
	fmt.Println("Completed preflight service check")

	// Pod check
	fmt.Printf("\nPreflight pod check: \n")

	err = checkPfPod(clientset, namespace)
	if err != nil {
		fmt.Println("Preflight Pod check: FAILED")
		return err
	}
	checkCount++
	fmt.Println("Completed preflight pod check")

	if checkCount == 3 {
		fmt.Printf("All preflight deploy checks have PASSED\n")
	} else {
		fmt.Printf("1 or more preflight deploy checks have FAILED\n")
	}
	return nil
}

func checkPfPod(clientset *kubernetes.Clientset, namespace string) error {
	// create a pod
	podName := "pod-pf-check"
	pod, err := createPreflightTestPod(clientset, namespace, podName, nginxImage)
	if err != nil {
		err = fmt.Errorf("Unable to create pod : %s\n", podName)
		return err
	}
	defer deletePod(clientset, namespace, podName)

	if len(pod.Spec.Containers) > 0 {
		timeout := time.NewTicker(2 * time.Minute)
		defer timeout.Stop()
	OUT:
		for {
			pod, err = getPod(clientset, namespace, pod.Name)
			if err != nil {
				err = fmt.Errorf("Unable to retrieve %s pod by name", podName)
				fmt.Println(err)
				return err
			}
			select {
			case <-timeout.C:
				break OUT
			default:
				if len(pod.Status.ContainerStatuses) > 0 && pod.Status.ContainerStatuses[0].Ready {
					break OUT
				}
			}
			time.Sleep(5 * time.Second)
		}
		if len(pod.Status.ContainerStatuses) == 0 || !pod.Status.ContainerStatuses[0].Ready {
			err = fmt.Errorf("container is taking much longer than expected")
			fmt.Println(err)
			return err
		}
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
		err = fmt.Errorf("Unable to create service : %s\n", serviceName)
		return err
	}
	defer deleteService(clientset, namespace, serviceName)
	_, err = getService(clientset, namespace, pfService.GetName())
	if err != nil {
		err = fmt.Errorf("Unable to retrieve service: %s\n", serviceName)
		return err
	}
	fmt.Println("Preflight service creation check: PASSED")
	fmt.Println("Cleaning up resources...")
	return nil
}

func checkPfDeployment(clientset *kubernetes.Clientset, namespace string) error {
	// check if we are able to create a deployment
	depName := "deployment-preflight-check"
	pfDeployment, err := createPreflightTestDeployment(clientset, namespace, depName, nginxImage)
	if err != nil {
		err = fmt.Errorf("Unable to create deployment: %v\n", err)
		return err
	}
	defer deleteDeployment(clientset, namespace, depName)
	timeout := time.NewTicker(2 * time.Minute)
	defer timeout.Stop()
WAIT:
	for {
		d, err := getDeployment(pfDeployment.GetName(), clientset, namespace)
		if err != nil {
			err = fmt.Errorf("Unable to retrieve deployment: %s\n", depName)
			return err
		}
		select {
		case <-timeout.C:
			break WAIT
		default:
			if int(d.Status.ReadyReplicas) > 0 {
				break WAIT
			}
		}
		time.Sleep(5 * time.Second)
	}
	fmt.Println("Preflight Deployment check: PASSED")
	fmt.Println("Cleaning up resources...")
	return nil
}
