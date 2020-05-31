package preflight

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
)

func (p *QliksensePreflight) CheckDeployment(namespace string, kubeConfigContents []byte, cleanup bool) error {
	clientset, _, err := p.CG.GetK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("Kube config error: %v\n", err)
		return err
	}

	// Deployment check
	if !cleanup {
		p.CG.LogVerboseMessage("Preflight deployment check: \n")
		p.CG.LogVerboseMessage("--------------------------- \n")
	}
	err = p.checkPfDeployment(clientset, namespace, cleanup)
	if err != nil {
		p.CG.LogVerboseMessage("Preflight Deployment check: FAILED\n")
		return err
	}
	if !cleanup {
		p.CG.LogVerboseMessage("Completed preflight deployment check\n")
	}

	return nil
}

func (p *QliksensePreflight) CheckService(namespace string, kubeConfigContents []byte, cleanup bool) error {
	clientset, _, err := p.CG.GetK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("unable to create a kubernetes client: %v\n", err)
		return err
	}
	// Service check
	if !cleanup {
		p.CG.LogVerboseMessage("Preflight service check: \n")
		p.CG.LogVerboseMessage("------------------------ \n")
	}
	err = p.checkPfService(clientset, namespace, cleanup)
	if err != nil {
		p.CG.LogVerboseMessage("Preflight Service check: FAILED\n")
		return err
	}

	if !cleanup {
		p.CG.LogVerboseMessage("Completed preflight service check\n")
	}
	return nil
}

func (p *QliksensePreflight) CheckPod(namespace string, kubeConfigContents []byte, cleanup bool) error {
	clientset, _, err := p.CG.GetK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("error: unable to create a kubernetes client: %v\n", err)
		return err
	}
	// Pod check
	if !cleanup {
		p.CG.LogVerboseMessage("Preflight pod check: \n")
		p.CG.LogVerboseMessage("-------------------- \n")
	}
	err = p.checkPfPod(clientset, namespace, cleanup)
	if err != nil {
		p.CG.LogVerboseMessage("Preflight Pod check: FAILED\n")
		return err
	}
	if !cleanup {
		p.CG.LogVerboseMessage("Completed preflight pod check\n")
	}
	return nil
}

func (p *QliksensePreflight) checkPfPod(clientset *kubernetes.Clientset, namespace string, cleanup bool) error {
	// delete the pod we are going to create, if it already exists in the cluster
	podName := "pod-pf-check"
	p.CG.DeletePod(clientset, namespace, podName)
	if cleanup {
		return nil
	}
	commandToRun := []string{}
	imageName, err := p.GetPreflightConfigObj().GetImageName(nginx, true)
	if err != nil {
		return err
	}
	// create a pod
	pod, err := p.CG.CreatePreflightTestPod(clientset, namespace, podName, imageName, nil, commandToRun)
	if err != nil {
		err = fmt.Errorf("unable to create pod - %v\n", err)
		return err
	}
	defer p.CG.DeletePod(clientset, namespace, podName)

	if err := p.CG.WaitForPod(clientset, namespace, pod); err != nil {
		return err
	}

	p.CG.LogVerboseMessage("Preflight pod creation check: PASSED\n")
	p.CG.LogVerboseMessage("Cleaning up resources...\n")
	return nil
}

func (p *QliksensePreflight) checkPfService(clientset *kubernetes.Clientset, namespace string, cleanup bool) error {
	// delete the service we are going to create, if it already exists in the cluster
	serviceName := "svc-pf-check"
	p.CG.DeleteService(clientset, namespace, serviceName)
	if cleanup {
		return nil
	}
	// creating service
	pfService, err := p.CG.CreatePreflightTestService(clientset, namespace, serviceName)
	if err != nil {
		err = fmt.Errorf("unable to create service - %v\n", err)
		return err
	}
	defer p.CG.DeleteService(clientset, namespace, serviceName)
	_, err = p.CG.GetService(clientset, namespace, pfService.GetName())
	if err != nil {
		err = fmt.Errorf("unable to retrieve service - %v\n", err)
		return err
	}
	p.CG.LogVerboseMessage("Preflight service creation check: PASSED\n")
	p.CG.LogVerboseMessage("Cleaning up resources...\n")
	return nil
}

func (p *QliksensePreflight) checkPfDeployment(clientset *kubernetes.Clientset, namespace string, cleanup bool) error {
	// delete the deployment we are going to create, if it already exists in the cluster
	depName := "deployment-preflight-check"
	p.CG.DeleteDeployment(clientset, namespace, depName)
	if cleanup {
		return nil
	}

	// check if we are able to create a deployment
	imageName, err := p.GetPreflightConfigObj().GetImageName(nginx, true)
	if err != nil {
		return err
	}
	pfDeployment, err := p.CG.CreatePreflightTestDeployment(clientset, namespace, depName, imageName)
	if err != nil {
		err = fmt.Errorf("unable to create deployment - %v\n", err)
		return err
	}
	defer p.CG.DeleteDeployment(clientset, namespace, depName)
	if err := p.CG.WaitForDeployment(clientset, namespace, pfDeployment); err != nil {
		return err
	}
	p.CG.LogVerboseMessage("Preflight Deployment check: PASSED\n")
	p.CG.LogVerboseMessage("Cleaning up resources...\n")
	return nil
}
