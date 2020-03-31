package preflight

import (
	"fmt"
	"strings"
	"time"
)

func (qp *QliksensePreflight) CheckDns(namespace string, kubeConfigContents []byte) error {
	clientset, clientConfig, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("Kube config error: %v\n", err)
		fmt.Print(err)
		return err
	}

	// creating deployment
	depName := "dep-dns-preflight-check"
	dnsDeployment, err := createPreflightTestDeployment(clientset, namespace, depName, "nginx")
	if err != nil {
		err = fmt.Errorf("Unable to create deployment: %v\n", err)
		return err
	}
	timeout := time.NewTicker(2 * time.Minute)
	defer timeout.Stop()
WAIT:
	for {
		d, err := getDeployment(dnsDeployment.GetName(), clientset, namespace)
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
	defer deleteDeployment(clientset, namespace, depName)

	// creating service
	serviceName := "svc-dns-pf-check"
	dnsService, err := createPreflightTestService(clientset, namespace, serviceName)
	if err != nil {
		err = fmt.Errorf("Unable to create service : %s\n", serviceName)
		return err
	}
	defer deleteService(clientset, namespace, serviceName)

	// create a pod
	podName := "pf-pod-1"
	dnsPod, err := createPreflightTestPod(clientset, namespace, podName, "subfuzion/netcat:latest")
	if err != nil {
		err = fmt.Errorf("Unable to create pod : %s\n", podName)
		return err
	}
	defer deletePod(clientset, namespace, podName)

	if len(dnsPod.Spec.Containers) > 0 {
		timeout := time.NewTicker(2 * time.Minute)
		defer timeout.Stop()
	OUT:
		for {
			dnsPod, err = getPod(clientset, namespace, dnsPod.Name)
			if err != nil {
				err = fmt.Errorf("Unable to retrieve service by name: %s\n", podName)
				fmt.Println(err)
				return err
			}
			select {
			case <-timeout.C:
				break OUT
			default:
				if len(dnsPod.Status.ContainerStatuses) > 0 && dnsPod.Status.ContainerStatuses[0].Ready {
					break OUT
				}
			}
			time.Sleep(5 * time.Second)
		}
		if len(dnsPod.Status.ContainerStatuses) == 0 || !dnsPod.Status.ContainerStatuses[0].Ready {
			err = fmt.Errorf("container is taking much longer than expected")
			fmt.Println(err)
			return err
		}
		fmt.Println("Exec-ing into the container...")
		stdout, stderr, err := executeRemoteCommand(clientset, clientConfig, dnsPod.Name, dnsPod.Spec.Containers[0].Name, namespace, []string{"nc", "-z", "-v", "-w 1", dnsService.Name, "80"})
		if err != nil {
			err = fmt.Errorf("An error occurred while executing remote command in container: %v", err)
			fmt.Println(err)
			return err
		}
		//fmt.Printf("stdout: %s\n", stdout)
		//fmt.Printf("stderr: %s\n", stderr)

		if strings.HasSuffix(stdout, "succeeded!") || strings.HasSuffix(stderr, "succeeded!") {
			fmt.Println("Preflight DNS check: PASSED")
		} else {
			fmt.Println("Preflight DNS check: FAILED")
		}
	}

	fmt.Println("Completed preflight DNS check")
	fmt.Println("Cleaning up resources...")

	return nil
}
