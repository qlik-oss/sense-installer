package preflight

import (
	"fmt"
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/api"
	"k8s.io/client-go/kubernetes"
)

const (
	nginx  = "nginx"
	netcat = "netcat"
)

func (qp *QliksensePreflight) CheckDns(namespace string, kubeConfigContents []byte, cleanup bool) error {
	depName := "dep-dns-preflight-check"
	serviceName := "svc-dns-pf-check"
	podName := "pf-pod-1"

	if !cleanup {
		qp.CG.LogVerboseMessage("Preflight DNS check: \n")
		qp.CG.LogVerboseMessage("------------------- \n")
	}
	clientset, _, err := api.GetK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("unable to create a kubernetes client: %v\n", err)
		return err
	}

	// delete the deployment we are going to create, if it already exists in the cluster
	qp.runDNSCleanup(clientset, namespace, podName, serviceName, depName)
	if cleanup {
		return nil
	}
	// creating deployment
	nginxImageName, err := qp.GetPreflightConfigObj().GetImageName(nginx, true)
	if err != nil {
		return err
	}

	dnsDeployment, err := qp.CG.CreatePreflightTestDeployment(clientset, namespace, depName, nginxImageName)
	if err != nil {
		err = fmt.Errorf("unable to create deployment: %v\n", err)
		return err
	}
	defer qp.CG.DeleteDeployment(clientset, namespace, depName)

	if err := api.WaitForDeployment(clientset, namespace, dnsDeployment); err != nil {
		return err
	}

	// creating service
	dnsService, err := qp.CG.CreatePreflightTestService(clientset, namespace, serviceName)
	if err != nil {
		err = fmt.Errorf("unable to create service : %s, %s\n", serviceName, err)
		return err
	}
	defer qp.CG.DeleteService(clientset, namespace, serviceName)

	// create a pod
	commandToRun := []string{"sh", "-c", "sleep 10; nc -z -v -w 1 " + dnsService.Name + " 80"}
	netcatImageName, err := qp.GetPreflightConfigObj().GetImageName(netcat, true)
	if err != nil {
		err = fmt.Errorf("unable to retrieve image : %v\n", err)
		return err
	}

	dnsPod, err := qp.CG.CreatePreflightTestPod(clientset, namespace, podName, netcatImageName, nil, commandToRun)
	if err != nil {
		err = fmt.Errorf("unable to create pod : %s, %s\n", podName, err)
		return err
	}

	defer qp.CG.DeletePod(clientset, namespace, podName)

	if err := api.WaitForPod(clientset, namespace, dnsPod); err != nil {
		return err
	}
	if len(dnsPod.Spec.Containers) == 0 {
		err := fmt.Errorf("there are no containers in the pod")
		return err
	}

	api.WaitForPodToDie(clientset, namespace, dnsPod)

	logStr, err := api.GetPodLogs(clientset, dnsPod)
	if err != nil {
		err = fmt.Errorf("unable to execute dns check in the cluster: %v", err)
		return err
	}

	if strings.HasSuffix(strings.TrimSpace(logStr), "succeeded!") {
		qp.CG.LogVerboseMessage("Preflight DNS check: PASSED\n")
	} else {
		err = fmt.Errorf("Expected response not found\n")
		return err
	}
	if !cleanup {
		qp.CG.LogVerboseMessage("Completed preflight DNS check\n")
		qp.CG.LogVerboseMessage("Cleaning up resources...\n")
	}

	return nil
}

func (qp *QliksensePreflight) runDNSCleanup(clientset *kubernetes.Clientset, namespace, podName, serviceName, depName string) {
	qp.CG.DeleteDeployment(clientset, namespace, depName)
	qp.CG.DeletePod(clientset, namespace, podName)
	qp.CG.DeleteService(clientset, namespace, serviceName)
}
