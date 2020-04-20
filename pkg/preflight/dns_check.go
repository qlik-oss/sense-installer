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
	qp.P.LogVerboseMessage("Preflight DNS check: \n")
	qp.P.LogVerboseMessage("------------------- \n")
	clientset, _, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("unable to create a kubernetes client: %v\n", err)
		return err
	}

	// creating deployment
	depName := "dep-dns-preflight-check"
	nginxImageName, err := qp.GetPreflightConfigObj().GetImageName(nginx, true)
	if err != nil {
		return err
	}
	dnsDeployment, err := qp.createPreflightTestDeployment(clientset, namespace, depName, nginxImageName)
	if err != nil {
		err = fmt.Errorf("unable to create deployment: %v\n", err)
		return err
	}
	defer qp.deleteDeployment(clientset, namespace, depName)

	if err := waitForDeployment(clientset, namespace, dnsDeployment); err != nil {
		return err
	}

	// creating service
	serviceName := "svc-dns-pf-check"
	dnsService, err := qp.createPreflightTestService(clientset, namespace, serviceName)
	if err != nil {
		err = fmt.Errorf("unable to create service : %s, %s\n", serviceName, err)
		return err
	}
	defer qp.deleteService(clientset, namespace, serviceName)

	// create a pod
	podName := "pf-pod-1"
	commandToRun := []string{"sh", "-c", "sleep 10; nc -z -v -w 1 " + dnsService.Name + " 80"}
	netcatImageName, err := qp.GetPreflightConfigObj().GetImageName(netcat, true)
	if err != nil {
		err = fmt.Errorf("unable to retrieve image : %v\n", err)
		return err
	}
	dnsPod, err := qp.createPreflightTestPod(clientset, namespace, podName, netcatImageName, nil, commandToRun)
	if err != nil {
		err = fmt.Errorf("unable to create pod : %s, %s\n", podName, err)
		return err
	}

	defer qp.deletePod(clientset, namespace, podName)

	if err := waitForPod(clientset, namespace, dnsPod); err != nil {
		return err
	}
	if len(dnsPod.Spec.Containers) == 0 {
		err := fmt.Errorf("there are no containers in the pod")
		return err
	}

	waitForPodToDie(clientset, namespace, dnsPod)

	logStr, err := getPodLogs(clientset, dnsPod)
	if err != nil {
		err = fmt.Errorf("unable to execute dns check in the cluster: %v", err)
		return err
	}

	if strings.HasSuffix(strings.TrimSpace(logStr), "succeeded!") {
		qp.P.LogVerboseMessage("Preflight DNS check: PASSED\n")
	} else {
		err = fmt.Errorf("Expected response not found\n")
		return err
	}

	qp.P.LogVerboseMessage("Completed preflight DNS check\n")
	qp.P.LogVerboseMessage("Cleaning up resources...\n")

	return nil
}
