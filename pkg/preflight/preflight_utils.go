package preflight

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

type PreflightOptions struct {
	Verbose      bool
	MongoOptions *MongoOptions
}

// LogVerboseMessage logs a verbose message
func (p *PreflightOptions) LogVerboseMessage(strMessage string, args ...interface{}) {
	if p.Verbose || os.Getenv("QLIKSENSE_DEBUG") == "true" {
		fmt.Printf(strMessage, args...)
	}
}

type MongoOptions struct {
	MongodbUrl     string
	Username       string
	Password       string
	CaCertFile     string
	ClientCertFile string
	Tls            bool
}

var gracePeriod int64 = 0

type QliksensePreflight struct {
	Q *qliksense.Qliksense
	P *PreflightOptions
}

func (qp *QliksensePreflight) GetPreflightConfigObj() *api.PreflightConfig {
	return api.NewPreflightConfig(qp.Q.QliksenseHome)
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
			return nil, nil, err
		}
	} else {
		config, err := clientcmd.Load(kubeconfig)
		if err != nil {
			err = errors.Wrap(err, "Unable to load kubeconfig")
			return nil, nil, err
		}
		if contextName != "" {
			config.CurrentContext = contextName
		}
		clientConfig, err = clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			err = errors.Wrap(err, "Unable to create client config from config")
			return nil, nil, err
		}
	}
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		err = errors.Wrap(err, "Unable to create clientset")
		return nil, nil, err
	}
	return clientset, clientConfig, nil
}

func (qp *QliksensePreflight) createPreflightTestDeployment(clientset *kubernetes.Clientset, namespace string, depName string, imageName string) (*appsv1.Deployment, error) {
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
		return nil, err
	}
	qp.P.LogVerboseMessage("Created deployment %q\n", result.GetObjectMeta().GetName())

	return deployment, nil
}

func getDeployment(clientset *kubernetes.Clientset, namespace, depName string) (*appsv1.Deployment, error) {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	var deployment *appsv1.Deployment
	if err := retryOnError(func() (err error) {
		deployment, err = deploymentsClient.Get(depName, v1.GetOptions{})
		return err
	}); err != nil {
		err = errors.Wrapf(err, "unable to get deployments in the %s namespace", namespace)
		api.LogDebugMessage("%v\n", err)
		return nil, err
	}
	return deployment, nil
}

func (qp *QliksensePreflight) deleteDeployment(clientset *kubernetes.Clientset, namespace, name string) error {
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
		return err
	}
	if err := waitForDeploymentToDelete(clientset, namespace, name); err != nil {
		return err
	}
	qp.P.LogVerboseMessage("Deleted deployment: %s\n", name)
	return nil
}

func (qp *QliksensePreflight) createPreflightTestService(clientset *kubernetes.Clientset, namespace string, svcName string) (*apiv1.Service, error) {
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
		return nil, err
	}
	qp.P.LogVerboseMessage("Created service %q\n", result.GetObjectMeta().GetName())

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
		return nil, err
	}

	return svc, nil
}

func (qp *QliksensePreflight) deleteService(clientset *kubernetes.Clientset, namespace, name string) error {
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
	qp.P.LogVerboseMessage("Deleted service: %s\n", name)
	return nil
}

func (qp *QliksensePreflight) deletePod(clientset *kubernetes.Clientset, namespace, name string) error {

	podsClient := clientset.CoreV1().Pods(namespace)
	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy:  &deletePolicy,
		GracePeriodSeconds: &gracePeriod,
	}
	if err := retryOnError(func() (err error) {
		return podsClient.Delete(name, &deleteOptions)
	}); err != nil {
		return err
	}
	if err := waitForPodToDelete(clientset, namespace, name); err != nil {
		return err
	}
	qp.P.LogVerboseMessage("Deleted pod: %s\n", name)
	return nil
}

func (qp *QliksensePreflight) createPreflightTestPod(clientset *kubernetes.Clientset, namespace, podName, imageName string, secretNames []string, commandToRun []string) (*apiv1.Pod, error) {
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
		for _, secretName := range secretNames {
			pod.Spec.Volumes = append(pod.Spec.Volumes, apiv1.Volume{
				Name: secretName,
				VolumeSource: apiv1.VolumeSource{
					Secret: &apiv1.SecretVolumeSource{
						SecretName: secretName,
						Items: []apiv1.KeyToPath{
							{
								Key:  secretName,
								Path: secretName,
							},
						},
					},
				},
			})
			if len(pod.Spec.Containers) > 0 {
				pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, apiv1.VolumeMount{
					Name:      secretName,
					MountPath: "/etc/ssl/" + secretName,
					ReadOnly:  true,
				})
			}
		}
	}

	// now create the pod in kubernetes cluster using the clientset
	if err := retryOnError(func() (err error) {
		pod, err = clientset.CoreV1().Pods(namespace).Create(pod)
		return err
	}); err != nil {
		return nil, err
	}
	qp.P.LogVerboseMessage("Created pod: %s\n", pod.Name)
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

func getPodLogs(clientset *kubernetes.Clientset, pod *apiv1.Pod) (string, error) {
	podLogOpts := apiv1.PodLogOptions{}

	api.LogDebugMessage("Retrieving logs for pod: %s   namespace: %s\n", pod.GetName(), pod.Namespace)
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
	api.LogDebugMessage("Log from pod: %s\n", buf.String())
	return buf.String(), nil
}

func waitForResource(checkFunc func() (interface{}, error), validateFunc func(interface{}) bool) error {
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

func waitForDeployment(clientset *kubernetes.Clientset, namespace string, pfDeployment *appsv1.Deployment) error {
	var err error
	depName := pfDeployment.GetName()
	checkFunc := func() (interface{}, error) {
		pfDeployment, err = getDeployment(clientset, namespace, depName)
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
	if err := waitForResource(checkFunc, validateFunc); err != nil {
		return err
	}
	if int(pfDeployment.Status.ReadyReplicas) == 0 {
		err = fmt.Errorf("deployment took longer than expected to spin up pods")
		return err
	}
	return nil
}

func waitForPod(clientset *kubernetes.Clientset, namespace string, pod *apiv1.Pod) error {
	var err error
	if len(pod.Spec.Containers) == 0 {
		err = fmt.Errorf("there are no containers in the pod")
		return err
	}
	podName := pod.Name
	checkFunc := func() (interface{}, error) {
		pod, err = getPod(clientset, namespace, podName)
		if err != nil {
			err = fmt.Errorf("unable to retrieve %s pod by name", podName)
			return nil, err
		}
		return pod, nil
	}
	validateFunc := func(data interface{}) bool {
		po := data.(*apiv1.Pod)
		return len(po.Status.ContainerStatuses) > 0 && po.Status.ContainerStatuses[0].Ready
	}

	if err := waitForResource(checkFunc, validateFunc); err != nil {
		return err
	}
	if len(pod.Status.ContainerStatuses) == 0 || !pod.Status.ContainerStatuses[0].Ready {
		err = fmt.Errorf("container is taking much longer than expected")
		return err
	}
	return nil
}

func waitForPodToDie(clientset *kubernetes.Clientset, namespace string, pod *apiv1.Pod) error {
	podName := pod.Name
	checkFunc := func() (interface{}, error) {
		po, err := getPod(clientset, namespace, podName)
		if err != nil {
			err = fmt.Errorf("unable to retrieve %s pod by name", podName)
			return nil, err
		}
		api.LogDebugMessage("pod status: %v\n", po.Status.Phase)
		return po, nil
	}
	validateFunc := func(r interface{}) bool {
		po := r.(*apiv1.Pod)
		return po.Status.Phase == apiv1.PodFailed || po.Status.Phase == apiv1.PodSucceeded
	}
	if err := waitForResource(checkFunc, validateFunc); err != nil {
		return err
	}
	return nil
}

func waitForPodToDelete(clientset *kubernetes.Clientset, namespace, podName string) error {
	checkFunc := func() (interface{}, error) {
		po, err := getPod(clientset, namespace, podName)
		if err != nil {
			return nil, err
		}
		return po, nil
	}
	validateFunc := func(po interface{}) bool {
		return false
	}
	if err := waitForResource(checkFunc, validateFunc); err != nil {
		return nil
	}
	err := fmt.Errorf("delete pod is taking unusually long")
	return err
}

func waitForDeploymentToDelete(clientset *kubernetes.Clientset, namespace, deploymentName string) error {
	checkFunc := func() (interface{}, error) {
		dep, err := getDeployment(clientset, namespace, deploymentName)
		if err != nil {
			return nil, err
		}
		return dep, nil
	}
	validateFunc := func(po interface{}) bool {
		return false
	}
	if err := waitForResource(checkFunc, validateFunc); err != nil {
		return nil
	}
	err := fmt.Errorf("delete deployment is taking unusually long")
	return err
}

func (qp *QliksensePreflight) createPfRole(clientset *kubernetes.Clientset, namespace, roleName string) (*v1beta1.Role, error) {
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
	if err := retryOnError(func() (err error) {
		role, err = clientset.RbacV1beta1().Roles(namespace).Create(roleSpec)
		return err
	}); err != nil {
		return nil, err
	}

	qp.P.LogVerboseMessage("Created role: %s\n", role.Name)

	return role, nil
}

func (qp *QliksensePreflight) deleteRole(clientset *kubernetes.Clientset, namespace string, role *v1beta1.Role) {
	rolesClient := clientset.RbacV1beta1().Roles(namespace)

	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	err := rolesClient.Delete(role.GetName(), &deleteOptions)
	if err != nil {
		log.Fatal(err)
	}
	qp.P.LogVerboseMessage("Deleted role: %s\n\n", role.Name)
}

func (qp *QliksensePreflight) createPfRoleBinding(clientset *kubernetes.Clientset, namespace, roleBindingName string) (*v1beta1.RoleBinding, error) {
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
	if err := retryOnError(func() (err error) {
		roleBinding, err = clientset.RbacV1beta1().RoleBindings(namespace).Create(roleBindingSpec)
		return err
	}); err != nil {
		return nil, err
	}
	qp.P.LogVerboseMessage("Created RoleBinding: %s\n", roleBindingSpec.Name)
	return roleBinding, nil
}

func (qp *QliksensePreflight) deleteRoleBinding(clientset *kubernetes.Clientset, namespace string, roleBinding *v1beta1.RoleBinding) {
	roleBindingClient := clientset.RbacV1beta1().RoleBindings(namespace)

	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	err := roleBindingClient.Delete(roleBinding.GetName(), &deleteOptions)
	if err != nil {
		log.Fatal(err)
	}
	qp.P.LogVerboseMessage("Deleted RoleBinding: %s\n\n", roleBinding.Name)
}

func (qp *QliksensePreflight) createPfServiceAccount(clientset *kubernetes.Clientset, namespace, serviceAccountName string) (*apiv1.ServiceAccount, error) {
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
	if err := retryOnError(func() (err error) {
		serviceAccount, err = clientset.CoreV1().ServiceAccounts(namespace).Create(serviceAccountSpec)
		return err
	}); err != nil {
		return nil, err
	}
	qp.P.LogVerboseMessage("Created Service Account: %s\n", serviceAccountSpec.Name)
	return serviceAccount, nil
}

func (qp *QliksensePreflight) deleteServiceAccount(clientset *kubernetes.Clientset, namespace string, serviceAccount *apiv1.ServiceAccount) {
	serviceAccountClient := clientset.CoreV1().ServiceAccounts(namespace)

	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	err := serviceAccountClient.Delete(serviceAccount.GetName(), &deleteOptions)
	if err != nil {
		log.Fatal(err)
	}
	qp.P.LogVerboseMessage("Deleted ServiceAccount: %s\n\n", serviceAccount.Name)
}

func (qp *QliksensePreflight) createPreflightTestSecret(clientset *kubernetes.Clientset, namespace, secretName string, secretData []byte) (*apiv1.Secret, error) {
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
	if err = retryOnError(func() (err error) {
		secret, err = clientset.CoreV1().Secrets(namespace).Create(secretSpec)
		return err
	}); err != nil {
		return nil, err
	}
	qp.P.LogVerboseMessage("Created Secret: %s\n", secret.Name)
	return secret, nil
}

func (qp *QliksensePreflight) deleteK8sSecret(clientset *kubernetes.Clientset, namespace string, k8sSecret *apiv1.Secret) {
	secretClient := clientset.CoreV1().Secrets(namespace)

	deletePolicy := v1.DeletePropagationForeground
	deleteOptions := v1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	err := secretClient.Delete(k8sSecret.GetName(), &deleteOptions)
	if err != nil {
		log.Fatal(err)
	}
	qp.P.LogVerboseMessage("Deleted Secret: %s\n\n", k8sSecret.Name)
}
