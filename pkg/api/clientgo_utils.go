package api

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

var gracePeriod int64 = 0

type ClientGoUtils struct {
	Verbose bool
}

func (p *ClientGoUtils) LogVerboseMessage(strMessage string, args ...interface{}) {
	if p.Verbose || os.Getenv("QLIKSENSE_DEBUG") == "true" {
		fmt.Printf(strMessage, args...)
	}
}

func int32Ptr(i int32) *int32 { return &i }

func (p *ClientGoUtils) LoadKubeConfigAndNamespace() (string, []byte, error) {
	fmt.Println("Reading .kube/config file...")

	homeDir, err := homedir.Dir()
	if err != nil {
		err = fmt.Errorf("Unable to deduce home dir")
		return "", nil, err
	}
	fmt.Printf("Kube config location: %s\n\n", filepath.Join(homeDir, ".kube", "config"))

	kubeConfig := filepath.Join(homeDir, ".kube", "config")
	kubeConfigContents, err := ioutil.ReadFile(kubeConfig)
	if err != nil {
		err = fmt.Errorf("Unable to deduce home dir")
		return "", nil, err
	}

	// retrieve namespace
	namespace := GetKubectlNamespace()
	// if namespace comes back empty, we will run checks in the default namespace
	if namespace == "" {
		namespace = "default"
	}

	return namespace, kubeConfigContents, nil
}

func (p *ClientGoUtils) RetryOnError(mf func() error) error {
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

func (p *ClientGoUtils) GetK8SClientSet(kubeconfig []byte, contextName string) (kubernetes.Interface, *rest.Config, error) {
	var clientConfig *rest.Config
	var err error
	if len(kubeconfig) == 0 {
		clientConfig, err = rest.InClusterConfig()
		if err != nil {
			err = fmt.Errorf("Unable to load in-cluster kubeconfig: %w", err)
			return nil, nil, err
		}
	} else {
		config, err := clientcmd.Load(kubeconfig)
		if err != nil {
			err = fmt.Errorf("Unable to load kubeconfig: %w", err)
			return nil, nil, err
		}
		if contextName != "" {
			config.CurrentContext = contextName
		}
		clientConfig, err = clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			err = fmt.Errorf("Unable to create client config from config: %w", err)
			return nil, nil, err
		}
	}
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		err = fmt.Errorf("Unable to create clientset: %w", err)
		return nil, nil, err
	}
	return clientset, clientConfig, nil
}

func (p *ClientGoUtils) CreatePreflightTestDeployment(clientset kubernetes.Interface, namespace string, depName string, imageName string) (*appsv1.Deployment, error) {
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
	if err := p.RetryOnError(func() (err error) {
		result, err = deploymentsClient.Create(deployment)
		return err
	}); err != nil {
		err = fmt.Errorf("unable to create deployments in the %s namespace: %w", namespace, err)
		return nil, err
	}
	p.LogVerboseMessage("Created deployment %q\n", result.GetObjectMeta().GetName())

	return deployment, nil
}

func (p *ClientGoUtils) getDeployment(clientset kubernetes.Interface, namespace, depName string) (*appsv1.Deployment, error) {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	var deployment *appsv1.Deployment
	if err := p.RetryOnError(func() (err error) {
		deployment, err = deploymentsClient.Get(depName, v1.GetOptions{})
		return err
	}); err != nil {
		err = fmt.Errorf("unable to get deployments in the %s namespace: %w", namespace, err)
		LogDebugMessage("%v\n", err)
		return nil, err
	}
	return deployment, nil
}

func (p *ClientGoUtils) DeleteDeployment(clientset kubernetes.Interface, namespace, name string) error {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	// Create Deployment
	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy:  &deletePolicy,
		GracePeriodSeconds: &gracePeriod,
	}

	if err := p.RetryOnError(func() (err error) {
		return deploymentsClient.Delete(name, &deleteOptions)
	}); err != nil {
		return err
	}
	if err := p.WaitForDeploymentToDelete(clientset, namespace, name); err != nil {
		return err
	}
	p.LogVerboseMessage("Deleted deployment: %s\n", name)
	return nil
}

func (p *ClientGoUtils) CreatePreflightTestService(clientset kubernetes.Interface, namespace string, svcName string) (*apiv1.Service, error) {
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
	if err := p.RetryOnError(func() (err error) {
		result, err = servicesClient.Create(service)
		return err
	}); err != nil {
		return nil, err
	}
	p.LogVerboseMessage("Created service %q\n", result.GetObjectMeta().GetName())

	return service, nil
}

func (p *ClientGoUtils) GetService(clientset kubernetes.Interface, namespace, svcName string) (*apiv1.Service, error) {
	servicesClient := clientset.CoreV1().Services(namespace)
	var svc *apiv1.Service
	if err := p.RetryOnError(func() (err error) {
		svc, err = servicesClient.Get(svcName, v1.GetOptions{})
		return err
	}); err != nil {
		err = fmt.Errorf("unable to get services in the %s namespace: %w", namespace, err)
		return nil, err
	}

	return svc, nil
}

func (p *ClientGoUtils) DeleteService(clientset kubernetes.Interface, namespace, name string) error {
	servicesClient := clientset.CoreV1().Services(namespace)
	// Create Deployment
	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	if err := p.RetryOnError(func() (err error) {
		return servicesClient.Delete(name, &deleteOptions)
	}); err != nil {
		return err
	}
	p.LogVerboseMessage("Deleted service: %s\n", name)
	return nil
}

func (p *ClientGoUtils) DeletePod(clientset kubernetes.Interface, namespace, name string) error {

	podsClient := clientset.CoreV1().Pods(namespace)
	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy:  &deletePolicy,
		GracePeriodSeconds: &gracePeriod,
	}
	if err := p.RetryOnError(func() (err error) {
		return podsClient.Delete(name, &deleteOptions)
	}); err != nil {
		return err
	}
	if err := p.waitForPodToDelete(clientset, namespace, name); err != nil {
		return err
	}
	p.LogVerboseMessage("Deleted pod: %s\n", name)
	return nil
}

func (p *ClientGoUtils) CreatePreflightTestPod(clientset kubernetes.Interface, namespace, podName, imageName string, secretNames map[string]string, commandToRun []string) (*apiv1.Pod, error) {
	// build the pod definition we want to deploy
	pod := &apiv1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "preflight",
			},
		},
		Spec: apiv1.PodSpec{
			RestartPolicy: apiv1.RestartPolicyNever,
			Containers: []apiv1.Container{
				{
					Name:            "cnt",
					Image:           imageName,
					ImagePullPolicy: apiv1.PullIfNotPresent,
					Command:         commandToRun,
				},
			},
		},
	}
	if len(secretNames) > 0 {
		for secretName, mountPath := range secretNames {
			pod.Spec.Volumes = append(pod.Spec.Volumes, apiv1.Volume{
				Name: secretName,
				VolumeSource: apiv1.VolumeSource{
					Secret: &apiv1.SecretVolumeSource{
						SecretName: secretName,
						Items: []apiv1.KeyToPath{
							{
								Key:  secretName,
								Path: filepath.Base(mountPath),
							},
						},
					},
				},
			})
			if len(pod.Spec.Containers) > 0 {
				pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, apiv1.VolumeMount{
					Name:      secretName,
					MountPath: filepath.Dir(mountPath),
					ReadOnly:  true,
				})
			}
		}
	}

	// now create the pod in kubernetes cluster using the clientset
	if err := p.RetryOnError(func() (err error) {
		pod, err = clientset.CoreV1().Pods(namespace).Create(pod)
		return err
	}); err != nil {
		return nil, err
	}
	p.LogVerboseMessage("Created pod: %s\n", pod.Name)
	return pod, nil
}

func (p *ClientGoUtils) getPod(clientset kubernetes.Interface, namespace, podName string) (*apiv1.Pod, error) {
	LogDebugMessage("Fetching pod: %s\n", podName)
	var pod *apiv1.Pod
	if err := p.RetryOnError(func() (err error) {
		pod, err = clientset.CoreV1().Pods(namespace).Get(podName, v1.GetOptions{})
		return err
	}); err != nil {
		LogDebugMessage("%v\n", err)
		return nil, err
	}
	return pod, nil
}

func (p *ClientGoUtils) GetPodLogs(clientset kubernetes.Interface, pod *apiv1.Pod) (string, error) {
	return p.GetPodContainerLogs(clientset, pod, "")
}

func (p *ClientGoUtils) GetPodContainerLogs(clientset kubernetes.Interface, pod *apiv1.Pod, container string) (string, error) {
	podLogOpts := apiv1.PodLogOptions{}
	if container != "" {
		podLogOpts.Container = container
	}

	LogDebugMessage("Retrieving logs for pod: %s   namespace: %s\n", pod.GetName(), pod.Namespace)
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream()
	if err != nil {
		return "", err
	}
	defer podLogs.Close()
	time.Sleep(15 * time.Second)
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}
	LogDebugMessage("Log from pod: %s\n", buf.String())
	return buf.String(), nil
}

func (p *ClientGoUtils) waitForResource(checkFunc func() (interface{}, error), validateFunc func(interface{}) bool) error {
	timeout := time.NewTicker(2 * time.Minute)
	defer timeout.Stop()
OUT:
	for {
		r, err := checkFunc()
		if err != nil {
			return err
		}
		select {
		case <-timeout.C:
			break OUT
		default:
			if validateFunc(r) {
				break OUT
			}
		}
		time.Sleep(5 * time.Second)
	}
	return nil
}

func (p *ClientGoUtils) WaitForDeployment(clientset kubernetes.Interface, namespace string, pfDeployment *appsv1.Deployment) error {
	var err error
	depName := pfDeployment.GetName()
	checkFunc := func() (interface{}, error) {
		pfDeployment, err = p.getDeployment(clientset, namespace, depName)
		if err != nil {
			err = fmt.Errorf("unable to retrieve deployment: %s\n", depName)
			return nil, err
		}
		return pfDeployment, nil
	}
	validateFunc := func(data interface{}) bool {
		d := data.(*appsv1.Deployment)
		return int(d.Status.ReadyReplicas) > 0
	}
	if err := p.waitForResource(checkFunc, validateFunc); err != nil {
		return err
	}
	if int(pfDeployment.Status.ReadyReplicas) == 0 {
		err = fmt.Errorf("deployment took longer than expected to spin up pods")
		return err
	}
	return nil
}

func (p *ClientGoUtils) WaitForPod(clientset kubernetes.Interface, namespace string, pod *apiv1.Pod) error {
	var err error
	if len(pod.Spec.Containers) == 0 {
		err = fmt.Errorf("there are no containers in the pod")
		return err
	}
	podName := pod.Name
	checkFunc := func() (interface{}, error) {
		pod, err = p.getPod(clientset, namespace, podName)
		if err != nil {
			err = fmt.Errorf("unable to retrieve %s pod by name", podName)
			return nil, err
		}
		return pod, nil
	}
	validateFunc := func(data interface{}) bool {
		po := data.(*apiv1.Pod)
		return po.Status.Phase == apiv1.PodRunning || po.Status.Phase == apiv1.PodSucceeded || po.Status.Phase == apiv1.PodFailed
	}

	if err := p.waitForResource(checkFunc, validateFunc); err != nil {
		return err
	}
	if pod.Status.Phase != apiv1.PodRunning && pod.Status.Phase != apiv1.PodSucceeded && pod.Status.Phase != apiv1.PodFailed {
		err = fmt.Errorf("container is taking much longer than expected")
		return err
	}
	return nil
}

func (p *ClientGoUtils) WaitForPodToDie(clientset kubernetes.Interface, namespace string, pod *apiv1.Pod) error {
	podName := pod.Name
	checkFunc := func() (interface{}, error) {
		po, err := p.getPod(clientset, namespace, podName)
		if err != nil {
			err = fmt.Errorf("unable to retrieve %s pod by name", podName)
			return nil, err
		}
		return po, nil
	}
	validateFunc := func(r interface{}) bool {
		po := r.(*apiv1.Pod)
		return po.Status.Phase == apiv1.PodFailed || po.Status.Phase == apiv1.PodSucceeded
	}
	if err := p.waitForResource(checkFunc, validateFunc); err != nil {
		return err
	}
	return nil
}

func (p *ClientGoUtils) waitForPodToDelete(clientset kubernetes.Interface, namespace, podName string) error {
	checkFunc := func() (interface{}, error) {
		po, err := p.getPod(clientset, namespace, podName)
		if err != nil {
			return nil, err
		}
		return po, nil
	}
	validateFunc := func(po interface{}) bool {
		return false
	}
	if err := p.waitForResource(checkFunc, validateFunc); err != nil {
		return nil
	}
	err := fmt.Errorf("delete pod is taking unusually long")
	return err
}

func (p *ClientGoUtils) WaitForDeploymentToDelete(clientset kubernetes.Interface, namespace, deploymentName string) error {
	checkFunc := func() (interface{}, error) {
		dep, err := p.getDeployment(clientset, namespace, deploymentName)
		if err != nil {
			return nil, err
		}
		return dep, nil
	}
	validateFunc := func(po interface{}) bool {
		return false
	}
	if err := p.waitForResource(checkFunc, validateFunc); err != nil {
		return nil
	}
	err := fmt.Errorf("delete deployment is taking unusually long")
	return err
}

func (p *ClientGoUtils) CreatePfRole(clientset kubernetes.Interface, namespace, roleName string) (*v1beta1.Role, error) {
	// build the role defination we want to create
	var role *v1beta1.Role
	roleSpec := &v1beta1.Role{
		ObjectMeta: v1.ObjectMeta{
			Name:      roleName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "preflight",
			},
		},
		Rules: []v1beta1.PolicyRule{},
	}

	// now create the role in kubernetes cluster using the clientset
	if err := p.RetryOnError(func() (err error) {
		role, err = clientset.RbacV1beta1().Roles(namespace).Create(roleSpec)
		return err
	}); err != nil {
		return nil, err
	}

	p.LogVerboseMessage("Created role: %s\n", role.Name)

	return role, nil
}

func (p *ClientGoUtils) DeleteRole(clientset kubernetes.Interface, namespace string, roleName string) error {
	rolesClient := clientset.RbacV1beta1().Roles(namespace)

	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	err := rolesClient.Delete(roleName, &deleteOptions)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}
	p.LogVerboseMessage("Deleted role: %s\n\n", roleName)
	return nil
}

func (p *ClientGoUtils) CreatePfRoleBinding(clientset kubernetes.Interface, namespace, roleBindingName string) (*v1beta1.RoleBinding, error) {
	var roleBinding *v1beta1.RoleBinding
	// build the rolebinding defination we want to create
	roleBindingSpec := &v1beta1.RoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "preflight",
			},
		},
		Subjects: []v1beta1.Subject{
			{
				Kind:      "ServiceAccount",
				APIGroup:  "",
				Name:      "preflight-check-subject",
				Namespace: namespace,
			},
		},
		RoleRef: v1beta1.RoleRef{
			APIGroup: "",
			Kind:     "Role",
			Name:     "preflight-check-roleref",
		},
	}

	// now create the roleBinding in kubernetes cluster using the clientset
	if err := p.RetryOnError(func() (err error) {
		roleBinding, err = clientset.RbacV1beta1().RoleBindings(namespace).Create(roleBindingSpec)
		return err
	}); err != nil {
		return nil, err
	}
	p.LogVerboseMessage("Created RoleBinding: %s\n", roleBindingSpec.Name)
	return roleBinding, nil
}

func (p *ClientGoUtils) DeleteRoleBinding(clientset kubernetes.Interface, namespace string, roleBindingName string) error {
	roleBindingClient := clientset.RbacV1beta1().RoleBindings(namespace)

	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	err := roleBindingClient.Delete(roleBindingName, &deleteOptions)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}
	p.LogVerboseMessage("Deleted RoleBinding: %s\n\n", roleBindingName)
	return nil
}

func (p *ClientGoUtils) CreatePfServiceAccount(clientset kubernetes.Interface, namespace, serviceAccountName string) (*apiv1.ServiceAccount, error) {
	var serviceAccount *apiv1.ServiceAccount
	// build the serviceAccount defination we want to create
	serviceAccountSpec := &apiv1.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Name:      "preflight-check-test-serviceaccount",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "preflight",
			},
		},
	}

	// now create the serviceAccount in kubernetes cluster using the clientset
	if err := p.RetryOnError(func() (err error) {
		serviceAccount, err = clientset.CoreV1().ServiceAccounts(namespace).Create(serviceAccountSpec)
		return err
	}); err != nil {
		return nil, err
	}
	p.LogVerboseMessage("Created Service Account: %s\n", serviceAccountSpec.Name)
	return serviceAccount, nil
}

func (p *ClientGoUtils) DeleteServiceAccount(clientset kubernetes.Interface, namespace string, serviceAccountName string) error {
	serviceAccountClient := clientset.CoreV1().ServiceAccounts(namespace)

	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	err := serviceAccountClient.Delete(serviceAccountName, &deleteOptions)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}
	p.LogVerboseMessage("Deleted ServiceAccount: %s\n\n", serviceAccountName)
	return nil
}

func (p *ClientGoUtils) CreatePreflightTestSecret(clientset kubernetes.Interface, namespace, secretName string, secretData []byte) (*apiv1.Secret, error) {
	var secret *apiv1.Secret
	var err error
	// build the secret defination we want to create
	secretSpec := &apiv1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "preflight",
			},
		},
		Data: map[string][]byte{
			secretName: secretData,
		},
	}

	// now create the secret in kubernetes cluster using the clientset
	if err = p.RetryOnError(func() (err error) {
		secret, err = clientset.CoreV1().Secrets(namespace).Create(secretSpec)
		return err
	}); err != nil {
		return nil, err
	}
	p.LogVerboseMessage("Created Secret: %s\n", secret.Name)
	return secret, nil
}

func (p *ClientGoUtils) DeleteK8sSecret(clientset kubernetes.Interface, namespace string, secretName string) error {
	secretClient := clientset.CoreV1().Secrets(namespace)

	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	err := secretClient.Delete(secretName, &deleteOptions)
	if err != nil {
		return err
	}
	p.LogVerboseMessage("Deleted Secret: %s\n", secretName)
	return nil
}

func (p *ClientGoUtils) CreateStatefulSet(clientset kubernetes.Interface, namespace string, statefulSetName string, imageName string) (*appsv1.StatefulSet, error) {
	statefulSetsClient := clientset.AppsV1().StatefulSets(namespace)
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: v1.ObjectMeta{
			Name: statefulSetName,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32Ptr(1),
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "postflight-check",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"app":   "postflight-check",
						"label": "postflight-check-label",
					},
				},
				Spec: apiv1.PodSpec{
					InitContainers: []apiv1.Container{
						{
							Name:            "migration",
							Image:           "ubuntu",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							// Command:         []string{"bash", "-c", "for i in {1..10}; do echo \"from init container...\"; sleep 1; done"},
							Command: []string{"bash", "-c", "for i in {1..10}; do echo \"from init container...\"; sleep 1; exit 1; done"},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:  "statefulset",
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
	var result *appsv1.StatefulSet
	if err := p.RetryOnError(func() (err error) {
		result, err = statefulSetsClient.Create(statefulset)
		return err
	}); err != nil {
		err = fmt.Errorf("unable to create statefulsets in the %s namespace: %w", namespace, err)
		return nil, err
	}
	fmt.Printf("Created statefulset %q\n", result.GetObjectMeta().GetName())

	return statefulset, nil
}

func (p *ClientGoUtils) GetPodsForStatefulset(clientset kubernetes.Interface, statefulset *appsv1.StatefulSet, namespace string) (*apiv1.PodList, error) {
	set := labels.Set(statefulset.Spec.Template.Labels)
	listOptions := v1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := clientset.CoreV1().Pods(namespace).List(listOptions)
	for _, pod := range pods.Items {
		fmt.Fprintf(os.Stdout, "*********** pod name: %v\n", pod.Name)
	}
	return pods, err
}

func (p *ClientGoUtils) waitForStatefulSet(clientset kubernetes.Interface, namespace string, pfStatefulset *appsv1.StatefulSet) error {
	var err error
	statefulsetName := pfStatefulset.GetName()
	checkFunc := func() (interface{}, error) {
		pfStatefulset, err = p.getStatefulset(clientset, namespace, statefulsetName)
		if err != nil {
			err = fmt.Errorf("unable to retrieve stateful set: %s\n", statefulsetName)
			return nil, err
		}
		return pfStatefulset, nil
	}
	validateFunc := func(data interface{}) bool {
		s := data.(*appsv1.StatefulSet)
		return int(s.Status.ReadyReplicas) > 0
	}
	if err := p.waitForResource(checkFunc, validateFunc); err != nil {
		return err
	}
	if int(pfStatefulset.Status.ReadyReplicas) == 0 {
		err = fmt.Errorf("deployment took longer than expected to spin up pods")
		return err
	}
	return nil
}

func (p *ClientGoUtils) getStatefulset(clientset kubernetes.Interface, namespace, statefulsetName string) (*appsv1.StatefulSet, error) {
	statefulsetsClient := clientset.AppsV1().StatefulSets(namespace)
	var statefulset *appsv1.StatefulSet
	if err := p.RetryOnError(func() (err error) {
		statefulset, err = statefulsetsClient.Get(statefulsetName, v1.GetOptions{})
		return err
	}); err != nil {
		err = fmt.Errorf("unable to get statefulsets in the %s namespace: %w", namespace, err)
		fmt.Printf("%v\n", err)
		return nil, err
	}
	return statefulset, nil
}

func (p *ClientGoUtils) deleteStatefulSet(clientset kubernetes.Interface, namespace, name string) error {
	statefulsetClient := clientset.AppsV1().StatefulSets(namespace)

	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy:  &deletePolicy,
		GracePeriodSeconds: &gracePeriod,
	}

	if err := p.RetryOnError(func() (err error) {
		return statefulsetClient.Delete(name, &deleteOptions)
	}); err != nil {
		return err
	}
	if err := p.waitForStatefulsetToDelete(clientset, namespace, name); err != nil {
		return err
	}
	fmt.Printf("Deleted statefulset: %s\n", name)
	return nil
}

func (p *ClientGoUtils) waitForStatefulsetToDelete(clientset kubernetes.Interface, namespace, statefulsetName string) error {
	checkFunc := func() (interface{}, error) {
		statefulset, err := p.getStatefulset(clientset, namespace, statefulsetName)
		if err != nil {
			return nil, err
		}
		return statefulset, nil
	}
	validateFunc := func(po interface{}) bool {
		return false
	}
	if err := p.waitForResource(checkFunc, validateFunc); err != nil {
		return nil
	}
	err := fmt.Errorf("delete statefulset is taking unusually long")
	return err
}

func (p *ClientGoUtils) GetPodsAndPodLogsForFailedInitContainer(clientset kubernetes.Interface, lbls map[string]string, namespace, containerName string) error {
	set := labels.Set(lbls)
	listOptions := v1.ListOptions{LabelSelector: set.AsSelector().String()}
	podList, err := clientset.CoreV1().Pods(namespace).List(listOptions)
	if err != nil {
		err = fmt.Errorf("unable to get podlist: %v", err)
		fmt.Printf("%s\n", err)
	}
	fmt.Printf("%d Pods retrieved\n ", len(podList.Items))

	for _, pod := range podList.Items {
		fmt.Fprintf(os.Stdout, "*********** pod name: %v\n", pod.Name)
		fmt.Printf("%d init containers retrieved\n", len(pod.Spec.InitContainers))
		for _, cs := range pod.Status.InitContainerStatuses {
			if (cs.State.Terminated != nil && (cs.State.Terminated.Reason != "Completed" || cs.State.Terminated.ExitCode > 0)) ||
				(cs.LastTerminationState.Terminated != nil && (cs.LastTerminationState.Terminated.Reason != "Completed" || cs.LastTerminationState.Terminated.ExitCode > 0)) && (cs.Name == containerName) {
				logs, err := p.GetPodContainerLogs(clientset, &pod, cs.Name)
				if err != nil {
					err = fmt.Errorf("unable to get pod logs: %v", err)
					fmt.Printf("%s\n", err)
					return err
				}
				fmt.Printf("Retrieved logs from init container: %s\n%s", cs.Name, logs)
			}
		}
	}
	return nil
}
