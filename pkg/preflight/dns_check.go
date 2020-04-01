package preflight

import (
	"fmt"
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/api"
)

const (
	nginxImage  = "nginx"
	netcatImage = "subfuzion/netcat:latest"
)

func (qp *QliksensePreflight) CheckDns(namespace string, kubeConfigContents []byte) error {
	clientset, clientConfig, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("error: unable to create a kubernetes client: %v\n", err)
		fmt.Println(err)
		return err
	}

	// creating deployment
	depName := "dep-dns-preflight-check"
	dnsDeployment, err := createPreflightTestDeployment(clientset, namespace, depName, nginxImage)
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
	dnsPod, err := createPreflightTestPod(clientset, namespace, podName, netcatImage)
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
	api.LogDebugMessage("Exec-ing into the container...")
	stdout, stderr, err := executeRemoteCommand(clientset, clientConfig, dnsPod.Name, dnsPod.Spec.Containers[0].Name, namespace, []string{"nc", "-z", "-v", "-w 1", dnsService.Name, "80"})
	if err != nil {
		err = fmt.Errorf("error: unable to execute dns check in the cluster: %v", err)
		fmt.Println(err)
		return err
	}

	if strings.HasSuffix(stdout, "succeeded!") || strings.HasSuffix(stderr, "succeeded!") {
		fmt.Println("Preflight DNS check: PASSED")
	} else {
		fmt.Println("Preflight DNS check: FAILED")
	}

	fmt.Println("Completed preflight DNS check")
	fmt.Println("Cleaning up resources...")

	return nil
}
