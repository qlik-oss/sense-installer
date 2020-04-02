package preflight

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"

	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/util/retry"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"

	"github.com/pkg/errors"
	"github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

var gracePeriod int64 = 0

type QliksensePreflight struct {
	Q *qliksense.Qliksense
}

func (qp *QliksensePreflight) GetPreflightConfigObj() *PreflightConfig {
	return NewPreflightConfig(qp.Q.QliksenseHome)
}

func InitPreflight() (string, []byte, error) {
	api.LogDebugMessage("Reading .kube/config file...")

	homeDir, err := homedir.Dir()
	if err != nil {
		err = fmt.Errorf("Unable to deduce home dir\n")
		return "", nil, err
	}
	api.LogDebugMessage("Kube config location: %s\n\n", filepath.Join(homeDir, ".kube", "config"))

	kubeConfig := filepath.Join(homeDir, ".kube", "config")
	kubeConfigContents, err := ioutil.ReadFile(kubeConfig)
	if err != nil {
		err = fmt.Errorf("Unable to deduce home dir\n")
		return "", nil, err
	}
	// retrieve namespace
	namespace := api.GetKubectlNamespace()
	// if namespace comes back empty, we will run checks in the default namespace
	if namespace == "" {
		namespace = "default"
	}
	api.LogDebugMessage("Namespace: %s\n", namespace)
	return namespace, kubeConfigContents, nil
}

func initiateK8sOps(opr, namespace string) error {
	opr1 := strings.Fields(opr)
	_, err := api.KubectlDirectOps(opr1, namespace)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func int32Ptr(i int32) *int32 { return &i }

func retryOnError(mf func() error) error {
	return retry.OnError(wait.Backoff{
		Duration: 1 * time.Second,
		Factor:   1,
		Jitter:   0.1,
		Steps:    5,
	}, func(err error) bool {
		return k8serrors.IsConflict(err) || k8serrors.IsGone(err) || k8serrors.IsServerTimeout(err) ||
			k8serrors.IsServiceUnavailable(err) || k8serrors.IsTimeout(err) || k8serrors.IsTooManyRequests(err)
	}, mf)
}

func getK8SClientSet(kubeconfig []byte, contextName string) (*kubernetes.Clientset, *rest.Config, error) {
	var clientConfig *rest.Config
	var err error
	if len(kubeconfig) == 0 {
		clientConfig, err = rest.InClusterConfig()
		if err != nil {
			err = errors.Wrap(err, "Unable to load in-cluster kubeconfig")
			fmt.Println(err)
			return nil, nil, err
		}
	} else {
		config, err := clientcmd.Load(kubeconfig)
		if err != nil {
			err = errors.Wrap(err, "Unable to load kubeconfig")
			fmt.Println(err)
			return nil, nil, err
		}
		if contextName != "" {
			config.CurrentContext = contextName
		}
		clientConfig, err = clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			err = errors.Wrap(err, "Unable to create client config from config")
			fmt.Println(err)
			return nil, nil, err
		}
	}
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		err = errors.Wrap(err, "Unable to create clientset")
		fmt.Println(err)
		return nil, nil, err
	}
	return clientset, clientConfig, nil
}

func createPreflightTestDeployment(clientset *kubernetes.Clientset, namespace string, depName string, imageName string) (*appsv1.Deployment, error) {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	deployment := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: depName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "preflight-check",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"app":   "preflight-check",
						"label": "preflight-check-label",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "dep",
							Image: imageName,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	var result *appsv1.Deployment
	if err := retryOnError(func() (err error) {
		result, err = deploymentsClient.Create(deployment)
		return err
	}); err != nil {
		err = errors.Wrapf(err, "error: unable to create deployments in the %s namespace", namespace)
		fmt.Println(err)
		return nil, err
	}
	fmt.Printf("Created deployment %q\n", result.GetObjectMeta().GetName())

	return deployment, nil
}

func getDeployment(clientset *kubernetes.Clientset, namespace, depName string) (*appsv1.Deployment, error) {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	var deployment *appsv1.Deployment
	if err := retryOnError(func() (err error) {
		deployment, err = deploymentsClient.Get(depName, v1.GetOptions{})
		return err
	}); err != nil {
		err = errors.Wrapf(err, "error: unable to get deployments in the %s namespace", namespace)
		api.LogDebugMessage("%v\n", err)
		return nil, err
	}
	return deployment, nil
}

func deleteDeployment(clientset *kubernetes.Clientset, namespace, name string) error {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	// Create Deployment
	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy:  &deletePolicy,
		GracePeriodSeconds: &gracePeriod,
	}

	if err := retryOnError(func() (err error) {
		return deploymentsClient.Delete(name, &deleteOptions)
	}); err != nil {
		fmt.Println(err)
		return err
	}
	if err := waitForDeploymentToDelete(clientset, namespace, name); err != nil {
		return err
	}
	fmt.Printf("Deleted deployment: %s\n", name)
	return nil
}

func createPreflightTestService(clientset *kubernetes.Clientset, namespace string, svcName string) (*apiv1.Service, error) {
	iptr := int32Ptr(80)
	servicesClient := clientset.CoreV1().Services(namespace)
	service := &apiv1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      svcName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "preflight-check",
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{Name: "port1",
					Port: *iptr,
				},
			},
			Selector: map[string]string{
				"app": "preflight-check",
			},
			ClusterIP: "",
		},
	}
	var result *apiv1.Service
	if err := retryOnError(func() (err error) {
		result, err = servicesClient.Create(service)
		return err
	}); err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Printf("Created service %q\n", result.GetObjectMeta().GetName())

	return service, nil
}

func getService(clientset *kubernetes.Clientset, namespace, svcName string) (*apiv1.Service, error) {
	servicesClient := clientset.CoreV1().Services(namespace)
	var svc *apiv1.Service
	if err := retryOnError(func() (err error) {
		svc, err = servicesClient.Get(svcName, v1.GetOptions{})
		return err
	}); err != nil {
		err = errors.Wrapf(err, "unable to get services in the %s namespace", namespace)
		fmt.Println(err)
		return nil, err
	}

	return svc, nil
}

func deleteService(clientset *kubernetes.Clientset, namespace, name string) error {
	servicesClient := clientset.CoreV1().Services(namespace)
	// Create Deployment
	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	if err := retryOnError(func() (err error) {
		return servicesClient.Delete(name, &deleteOptions)
	}); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("Deleted service: %s\n", name)
	return nil
}

func deletePod(clientset *kubernetes.Clientset, namespace, name string) error {

	podsClient := clientset.CoreV1().Pods(namespace)
	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy:  &deletePolicy,
		GracePeriodSeconds: &gracePeriod,
	}
	if err := retryOnError(func() (err error) {
		return podsClient.Delete(name, &deleteOptions)
	}); err != nil {
		fmt.Println(err)
		return err
	}
	if err := waitForPodToDelete(clientset, namespace, name); err != nil {
		return err
	}
	fmt.Printf("Deleted pod: %s\n", name)
	return nil
}

func createPreflightTestPod(clientset *kubernetes.Clientset, namespace string, podName string, imageName string) (*apiv1.Pod, error) {
	// build the pod definition we want to deploy
	pod := &apiv1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "demo",
			},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:            "cnt",
					Image:           imageName,
					ImagePullPolicy: apiv1.PullIfNotPresent,
					Command: []string{
						"sleep",
						"3600",
					},
				},
			},
		},
	}

	// now create the pod in kubernetes cluster using the clientset
	if err := retryOnError(func() (err error) {
		pod, err = clientset.CoreV1().Pods(namespace).Create(pod)
		return err
	}); err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Printf("Created pod: %s\n", pod.Name)
	return pod, nil
}

func getPod(clientset *kubernetes.Clientset, namespace, podName string) (*apiv1.Pod, error) {
	api.LogDebugMessage("Fetching pod: %s\n", podName)
	var pod *apiv1.Pod
	if err := retryOnError(func() (err error) {
		pod, err = clientset.CoreV1().Pods(namespace).Get(podName, v1.GetOptions{})
		return err
	}); err != nil {
		api.LogDebugMessage("%v\n", err)
		return nil, err
	}
	return pod, nil
}

func execute(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}
	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	})
}

func executeRemoteCommand(clientset *kubernetes.Clientset, config *rest.Config, podName, containerName, namespace string, command []string) (string, string, error) {
	tty := false
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName)
	req.VersionedParams(&apiv1.PodExecOptions{
		Container: containerName,
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       tty,
	}, scheme.ParameterCodec)

	var stdout, stderr bytes.Buffer
	err := execute("POST", req.URL(), config, nil, &stdout, &stderr, tty)
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

func waitForDeployment(clientset *kubernetes.Clientset, namespace string, pfDeployment *appsv1.Deployment) error {
	timeout := time.NewTicker(2 * time.Minute)
	defer timeout.Stop()
	var d *appsv1.Deployment
	var err error
WAIT:
	for {
		d, err = getDeployment(clientset, namespace, pfDeployment.GetName())
		if err != nil {
			err = fmt.Errorf("error: unable to retrieve deployment: %s\n", pfDeployment.GetName())
			fmt.Println(err)
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
	if int(d.Status.ReadyReplicas) == 0 {
		err = fmt.Errorf("error: deployment took longer than expected to spin up pods")
		fmt.Println(err)
		return err
	}
	return nil
}

func waitForPod(clientset *kubernetes.Clientset, namespace string, pod *apiv1.Pod) error {
	var err error
	if len(pod.Spec.Containers) > 0 {
		timeout := time.NewTicker(2 * time.Minute)
		defer timeout.Stop()
		podName := pod.Name
	OUT:
		for {
			pod, err = getPod(clientset, namespace, podName)
			if err != nil {
				err = fmt.Errorf("error: unable to retrieve %s pod by name", podName)
				fmt.Println(err)
				return err
			}
			select {
			case <-timeout.C:
				break OUT
			default:
				if len(pod.Status.ContainerStatuses) > 0 && pod.Status.ContainerStatuses[0].Ready {
					break OUT
				}
			}
			time.Sleep(5 * time.Second)
		}
		if len(pod.Status.ContainerStatuses) == 0 || !pod.Status.ContainerStatuses[0].Ready {
			err = fmt.Errorf("error: container is taking much longer than expected")
			fmt.Println(err)
			return err
		}
		return nil
	}
	err = fmt.Errorf("error: there are no containers in the pod")
	fmt.Println(err)
	return err
}

func waitForPodToDelete(clientset *kubernetes.Clientset, namespace, podName string) error {
	var err error
	timeout := time.NewTicker(2 * time.Minute)
	defer timeout.Stop()
OUT:
	for {
		_, err = getPod(clientset, namespace, podName)
		if err != nil {
			return nil
		}
		select {
		case <-timeout.C:
			break OUT
		default:

		}
		time.Sleep(5 * time.Second)
	}
	err = fmt.Errorf("error: delete pod is taking unusually long")
	fmt.Println(err)
	return err
}

func waitForDeploymentToDelete(clientset *kubernetes.Clientset, namespace, deploymentName string) error {
	var err error
	timeout := time.NewTicker(2 * time.Minute)
	defer timeout.Stop()
OUT:
	for {
		_, err = getDeployment(clientset, namespace, deploymentName)
		if err != nil {
			return nil
		}
		select {
		case <-timeout.C:
			break OUT
		default:

		}
		time.Sleep(5 * time.Second)
	}
	err = fmt.Errorf("error: delete deployment is taking unusually long")
	fmt.Println(err)
	return err
}
