package preflight

import (
	"fmt"
	"strings"

	"k8s.io/client-go/kubernetes"
)

const (
	nginx  = "nginx"
	netcat = "netcat"
)

func (p *QliksensePreflight) CheckDns(namespace string, kubeConfigContents []byte, cleanup bool) error {
	depName := "dep-dns-preflight-check"
	serviceName := "svc-dns-pf-check"
	podName := "pf-pod-1"

	if !cleanup {
		p.CG.LogVerboseMessage("Preflight DNS check: \n")
		p.CG.LogVerboseMessage("------------------- \n")
	}
	clientset, _, err := p.CG.GetK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("unable to create a kubernetes client: %v\n", err)
		return err
	}

	// delete the deployment we are going to create, if it already exists in the cluster
	p.runDNSCleanup(clientset, namespace, podName, serviceName, depName)
	if cleanup {
		return nil
	}
	// creating deployment
	nginxImageName, err := p.GetPreflightConfigObj().GetImageName(nginx, true)
	if err != nil {
		return err
	}

	dnsDeployment, err := p.CG.CreatePreflightTestDeployment(clientset, namespace, depName, nginxImageName)
	if err != nil {
		err = fmt.Errorf("unable to create deployment: %v\n", err)
		return err
	}
	defer p.CG.DeleteDeployment(clientset, namespace, depName)

	if err := p.CG.WaitForDeployment(clientset, namespace, dnsDeployment); err != nil {
		return err
	}

	// creating service
	dnsService, err := p.CG.CreatePreflightTestService(clientset, namespace, serviceName)
	if err != nil {
		err = fmt.Errorf("unable to create service : %s, %s\n", serviceName, err)
		return err
	}
	defer p.CG.DeleteService(clientset, namespace, serviceName)

	// create a pod
	commandToRun := []string{"sh", "-c", "sleep 10; nc -z -v -w 1 " + dnsService.Name + " 80"}
	netcatImageName, err := p.GetPreflightConfigObj().GetImageName(netcat, true)
	if err != nil {
		err = fmt.Errorf("unable to retrieve image : %v\n", err)
		return err
	}

	dnsPod, err := p.CG.CreatePreflightTestPod(clientset, namespace, podName, netcatImageName, nil, commandToRun)
	if err != nil {
		err = fmt.Errorf("unable to create pod : %s, %s\n", podName, err)
		return err
	}

	defer p.CG.DeletePod(clientset, namespace, podName)

	if err := p.CG.WaitForPod(clientset, namespace, dnsPod); err != nil {
		return err
	}
	if len(dnsPod.Spec.Containers) == 0 {
		err := fmt.Errorf("there are no containers in the pod")
		return err
	}

	p.CG.WaitForPodToDie(clientset, namespace, dnsPod)

	logStr, err := p.CG.GetPodLogs(clientset, dnsPod)
	if err != nil {
		err = fmt.Errorf("unable to execute dns check in the cluster: %v", err)
		return err
	}

	if strings.HasSuffix(strings.TrimSpace(logStr), "succeeded!") {
		p.CG.LogVerboseMessage("Preflight DNS check: PASSED\n")
	} else {
		err = fmt.Errorf("Expected response not found\n")
		return err
	}
	if !cleanup {
		p.CG.LogVerboseMessage("Completed preflight DNS check\n")
		p.CG.LogVerboseMessage("Cleaning up resources...\n")
	}

	return nil
}

func (p *QliksensePreflight) runDNSCleanup(clientset kubernetes.Interface, namespace, podName, serviceName, depName string) {
	p.CG.DeleteDeployment(clientset, namespace, depName)
	p.CG.DeletePod(clientset, namespace, podName)
	p.CG.DeleteService(clientset, namespace, serviceName)
}
