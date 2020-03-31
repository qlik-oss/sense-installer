package preflight

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

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

type QliksensePreflight struct {
	Q *qliksense.Qliksense
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

func createPfDeployment(clientset *kubernetes.Clientset, namespace string, depName string, imageName string) (*appsv1.Deployment, error) {
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
		err = errors.Wrapf(err, "unable to create deployments in the %s namespace", namespace)
		fmt.Println(err)
		return nil, err
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	return deployment, nil
}

func getDeployment(depName string, clientset *kubernetes.Clientset, namespace string) (*appsv1.Deployment, error) {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	var deployment *appsv1.Deployment
	if err := retryOnError(func() (err error) {
		deployment, err = deploymentsClient.Get(depName, v1.GetOptions{})
		return err
	}); err != nil {
		err = errors.Wrapf(err, "unable to get deployments in the %s namespace", namespace)
		fmt.Println(err)
		return nil, err
	}
	return deployment, nil
}

func deleteDeployment(clientset *kubernetes.Clientset, namespace, name string) {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	// Create Deployment
	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	if err := retryOnError(func() (err error) {
		return deploymentsClient.Delete(name, &deleteOptions)
	}); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Deleted deployment: %s\n", name)
}

func createPfService(clientset *kubernetes.Clientset, namespace string, svcName string) (*apiv1.Service, error) {
	//fmt.Println("Creating Service...")
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
	fmt.Printf("Created service %q.\n", result.GetObjectMeta().GetName())

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
		PropagationPolicy: &deletePolicy,
	}
	if err := retryOnError(func() (err error) {
		return podsClient.Delete(name, &deleteOptions)
	}); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("Deleted pod: %s\n", name)
	return nil
}

func createPfPod(clientset *kubernetes.Clientset, namespace string, podName string, imageName string) (*apiv1.Pod, error) {
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
	fmt.Printf("Fetching pod: %s\n", podName)
	var pod *apiv1.Pod
	if err := retryOnError(func() (err error) {
		pod, err = clientset.CoreV1().Pods(namespace).Get(podName, v1.GetOptions{})
		return err
	}); err != nil {
		fmt.Println(err)
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
