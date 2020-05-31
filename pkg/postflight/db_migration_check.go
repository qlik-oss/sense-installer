package postflight

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const initContainerNameToCheck = "migration"

func (p *QliksensePostflight) DbMigrationCheck(namespace string, kubeConfigContents []byte) error {

	clientset, _, err := p.CG.GetK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("unable to create a kubernetes client: %v", err)
		fmt.Printf("%s\n", err)
		return err
	}

	// Retrieve all deployments
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	deployments, err := deploymentsClient.List(v1.ListOptions{})
	fmt.Printf("Number of deployments found: %d\n", deployments.Size())
	for _, deployment := range deployments.Items {
		fmt.Printf("Deployment name: %s\n", deployment.GetName())
		if err = p.CG.GetPodsAndPodLogsForFailedInitContainer(clientset, deployment.Spec.Template.Labels, namespace, initContainerNameToCheck); err != nil {
			// err = fmt.Errorf("unable to fetch d a kubernetes client: %v", err)
			fmt.Printf("%s\n", err)
			return err
		}
	}

	// retrieve all statefulsets
	statefulsetsClient := clientset.AppsV1().StatefulSets(namespace)
	statefulsets, err := statefulsetsClient.List(v1.ListOptions{})
	fmt.Printf("Number of statefulsets found: %d\n", statefulsets.Size())
	for _, statefulset := range statefulsets.Items {
		fmt.Printf("Deployment name: %s\n", statefulset.GetName())
		if err = p.CG.GetPodsAndPodLogsForFailedInitContainer(clientset, statefulset.Spec.Template.Labels, namespace, initContainerNameToCheck); err != nil {
			// err = fmt.Errorf("unable to create a kubernetes client: %v", err)
			fmt.Printf("%s\n", err)
			return err
		}
	}

	fmt.Printf("all done!\n")
	return nil
}
