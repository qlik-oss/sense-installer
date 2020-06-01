package postflight

import (
	"fmt"
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/api"
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

	var logs string
	// Retrieve all deployments
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	deployments, err := deploymentsClient.List(v1.ListOptions{})
	api.LogDebugMessage("Number of deployments found: %d\n", deployments.Size())
	for _, deployment := range deployments.Items {
		api.LogDebugMessage("Deployment name: %s\n", deployment.GetName())
		if logs, err = p.CG.GetPodsAndPodLogsFromFailedInitContainer(clientset, deployment.Spec.Template.Labels, namespace, initContainerNameToCheck); err != nil {
			fmt.Printf("%s\n", err)
			return err
		}
		p.filterLogsForErrors(logs)
	}

	// retrieve all statefulsets
	statefulsetsClient := clientset.AppsV1().StatefulSets(namespace)
	statefulsets, err := statefulsetsClient.List(v1.ListOptions{})
	api.LogDebugMessage("Number of statefulsets found: %d\n", statefulsets.Size())
	for _, statefulset := range statefulsets.Items {
		api.LogDebugMessage("Statefulset name: %s\n", statefulset.GetName())
		if logs, err = p.CG.GetPodsAndPodLogsFromFailedInitContainer(clientset, statefulset.Spec.Template.Labels, namespace, initContainerNameToCheck); err != nil {
			fmt.Printf("%s\n", err)
			return err
		}
		p.filterLogsForErrors(logs)
	}

	return nil
}

func (p *QliksensePostflight) filterLogsForErrors(logs string) {
	containerLogs := strings.Split(logs, "\n")
	if len(containerLogs) > 0 {
		for _, logLine := range containerLogs {
			api.LogDebugMessage("init container logs: \n")
			if strings.Contains(strings.ToLower(logLine), "error") {
				fmt.Printf("%s\n", logLine)
			}
		}
	} else {
		p.CG.LogVerboseMessage("no logs obtained\n")
	}
}
