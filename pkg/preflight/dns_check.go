package preflight

import (
	"fmt"
	"strings"
)

const (
	nginx  = "nginx"
	netcat = "netcat"
)

func (qp *QliksensePreflight) CheckDns(namespace string, kubeConfigContents []byte) error {
	clientset, _, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("error: unable to create a kubernetes client: %v\n", err)
		fmt.Println(err)
		return err
	}

	// creating deployment
	depName := "dep-dns-preflight-check"
	nginxImageName, err := qp.GetPreflightConfigObj().GetImageName(nginx, true)
	if err != nil {
		return err
	}
	dnsDeployment, err := createPreflightTestDeployment(clientset, namespace, depName, nginxImageName)
	if err != nil {
		err = fmt.Errorf("error: unable to create deployment: %v\n", err)
		fmt.Println(err)
		return err
	}
	defer deleteDeployment(clientset, namespace, depName)

	if err := waitForDeployment(clientset, namespace, dnsDeployment); err != nil {
		return err
	}

	// creating service
	serviceName := "svc-dns-pf-check"
	dnsService, err := createPreflightTestService(clientset, namespace, serviceName)
	if err != nil {
		err = fmt.Errorf("error: unable to create service : %s\n", serviceName)
		return err
	}
	defer deleteService(clientset, namespace, serviceName)

	// create a pod
	podName := "pf-pod-1"
	commandToRun := []string{"sh", "-c", "sleep 10; nc -z -v -w 1 " + dnsService.Name + " 80"}
	netcatImageName, err := qp.GetPreflightConfigObj().GetImageName(netcat, true)
	if err != nil {
		return err
	}
	dnsPod, err := createPreflightTestPod(clientset, namespace, podName, netcatImageName, commandToRun)
	if err != nil {
		err = fmt.Errorf("error: unable to create pod : %s\n", podName)
		return err
	}

	defer deletePod(clientset, namespace, podName)

	if err := waitForPod(clientset, namespace, dnsPod); err != nil {
		return err
	}
	if len(dnsPod.Spec.Containers) == 0 {
		err := fmt.Errorf("error: there are no containers in the pod")
		fmt.Println(err)
		return err
	}

	waitForPodToDie(clientset, namespace, dnsPod)

	logStr, err := getPodLogs(clientset, dnsPod)
	if err != nil {
		err = fmt.Errorf("error: unable to execute dns check in the cluster: %v", err)
		fmt.Println(err)
		return err
	}

	if strings.HasSuffix(strings.TrimSpace(logStr), "succeeded!") {
		fmt.Println("Preflight DNS check: PASSED")
	} else {
		err = fmt.Errorf("Expected response not found\n")
		return err
	}

	fmt.Println("Completed preflight DNS check")
	fmt.Println("Cleaning up resources...")

	return nil
}
