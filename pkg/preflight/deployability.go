package preflight

import (
	"fmt"

	"github.com/qlik-oss/sense-installer/pkg/api"
	"k8s.io/client-go/kubernetes"
)

func (qp *QliksensePreflight) CheckDeployment(namespace string, kubeConfigContents []byte, cleanup bool) error {
	clientset, _, err := api.GetK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("Kube config error: %v\n", err)
		return err
	}

	// Deployment check
	if !cleanup {
		qp.CG.LogVerboseMessage("Preflight deployment check: \n")
		qp.CG.LogVerboseMessage("--------------------------- \n")
	}
	err = qp.checkPfDeployment(clientset, namespace, cleanup)
	if err != nil {
		qp.CG.LogVerboseMessage("Preflight Deployment check: FAILED\n")
		return err
	}
	if !cleanup {
		qp.CG.LogVerboseMessage("Completed preflight deployment check\n")
	}

	return nil
}

func (qp *QliksensePreflight) CheckService(namespace string, kubeConfigContents []byte, cleanup bool) error {
	clientset, _, err := api.GetK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("unable to create a kubernetes client: %v\n", err)
		return err
	}
	// Service check
	if !cleanup {
		qp.CG.LogVerboseMessage("Preflight service check: \n")
		qp.CG.LogVerboseMessage("------------------------ \n")
	}
	err = qp.checkPfService(clientset, namespace, cleanup)
	if err != nil {
		qp.CG.LogVerboseMessage("Preflight Service check: FAILED\n")
		return err
	}

	if !cleanup {
		qp.CG.LogVerboseMessage("Completed preflight service check\n")
	}
	return nil
}

func (qp *QliksensePreflight) CheckPod(namespace string, kubeConfigContents []byte, cleanup bool) error {
	clientset, _, err := api.GetK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("error: unable to create a kubernetes client: %v\n", err)
		return err
	}
	// Pod check
	if !cleanup {
		qp.CG.LogVerboseMessage("Preflight pod check: \n")
		qp.CG.LogVerboseMessage("-------------------- \n")
	}
	err = qp.checkPfPod(clientset, namespace, cleanup)
	if err != nil {
		qp.CG.LogVerboseMessage("Preflight Pod check: FAILED\n")
		return err
	}
	if !cleanup {
		qp.CG.LogVerboseMessage("Completed preflight pod check\n")
	}
	return nil
}

func (qp *QliksensePreflight) checkPfPod(clientset *kubernetes.Clientset, namespace string, cleanup bool) error {
	// delete the pod we are going to create, if it already exists in the cluster
	podName := "pod-pf-check"
	qp.CG.DeletePod(clientset, namespace, podName)
	if cleanup {
		return nil
	}
	commandToRun := []string{}
	imageName, err := qp.GetPreflightConfigObj().GetImageName(nginx, true)
	if err != nil {
		return err
	}
	// create a pod
	pod, err := qp.CG.CreatePreflightTestPod(clientset, namespace, podName, imageName, nil, commandToRun)
	if err != nil {
		err = fmt.Errorf("unable to create pod - %v\n", err)
		return err
	}
	defer qp.CG.DeletePod(clientset, namespace, podName)

	if err := api.WaitForPod(clientset, namespace, pod); err != nil {
		return err
	}

	qp.CG.LogVerboseMessage("Preflight pod creation check: PASSED\n")
	qp.CG.LogVerboseMessage("Cleaning up resources...\n")
	return nil
}

func (qp *QliksensePreflight) checkPfService(clientset *kubernetes.Clientset, namespace string, cleanup bool) error {
	// delete the service we are going to create, if it already exists in the cluster
	serviceName := "svc-pf-check"
	qp.CG.DeleteService(clientset, namespace, serviceName)
	if cleanup {
		return nil
	}
	// creating service
	pfService, err := qp.CG.CreatePreflightTestService(clientset, namespace, serviceName)
	if err != nil {
		err = fmt.Errorf("unable to create service - %v\n", err)
		return err
	}
	defer qp.CG.DeleteService(clientset, namespace, serviceName)
	_, err = api.GetService(clientset, namespace, pfService.GetName())
	if err != nil {
		err = fmt.Errorf("unable to retrieve service - %v\n", err)
		return err
	}
	qp.CG.LogVerboseMessage("Preflight service creation check: PASSED\n")
	qp.CG.LogVerboseMessage("Cleaning up resources...\n")
	return nil
}

func (qp *QliksensePreflight) checkPfDeployment(clientset *kubernetes.Clientset, namespace string, cleanup bool) error {
	// delete the deployment we are going to create, if it already exists in the cluster
	depName := "deployment-preflight-check"
	qp.CG.DeleteDeployment(clientset, namespace, depName)
	if cleanup {
		return nil
	}

	// check if we are able to create a deployment
	imageName, err := qp.GetPreflightConfigObj().GetImageName(nginx, true)
	if err != nil {
		return err
	}
	pfDeployment, err := qp.CG.CreatePreflightTestDeployment(clientset, namespace, depName, imageName)
	if err != nil {
		err = fmt.Errorf("unable to create deployment - %v\n", err)
		return err
	}
	defer qp.CG.DeleteDeployment(clientset, namespace, depName)
	if err := api.WaitForDeployment(clientset, namespace, pfDeployment); err != nil {
		return err
	}
	qp.CG.LogVerboseMessage("Preflight Deployment check: PASSED\n")
	qp.CG.LogVerboseMessage("Cleaning up resources...\n")
	return nil
}
