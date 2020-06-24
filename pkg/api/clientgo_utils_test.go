package api

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestClientGoUtils_getDeployment(t *testing.T) {
	t.Parallel()
	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset kubernetes.Interface
		namespace string
		depName   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *appsv1.Deployment
		wantErr bool
	}{
		{
			name:   "retrieve valid deployment",
			fields: fields{Verbose: true},
			args: args{
				clientset: fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dep",
						Namespace: "test-ns",
					},
				}),
				namespace: "test-ns",
				depName:   "test-dep",
			},
			want: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dep",
					Namespace: "test-ns",
				},
			},
			wantErr: false,
		},
		{
			name:   "retrieve non-existent deployment",
			fields: fields{Verbose: true},
			args: args{
				clientset: fake.NewSimpleClientset(),
				namespace: "test-ns",
				depName:   "test-dep",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			got, err := p.getDeployment(tt.args.clientset, tt.args.namespace, tt.args.depName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.getDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientGoUtils.getDeployment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientGoUtils_DeleteDeployment(t *testing.T) {
	t.Parallel()
	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset kubernetes.Interface
		namespace string
		name      string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "delete valid deployment",
			fields: fields{Verbose: true},
			args: args{
				clientset: fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dep",
						Namespace: "test-ns",
					},
				}),
				namespace: "test-ns",
				name:      "test-dep",
			},
			wantErr: false,
		},
		{
			name:   "delete non-existent deployment",
			fields: fields{Verbose: true},
			args: args{
				clientset: fake.NewSimpleClientset(),
				namespace: "test-ns",
				name:      "test-dep",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			if err := p.DeleteDeployment(tt.args.clientset, tt.args.namespace, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.DeleteDeployment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientGoUtils_GetService(t *testing.T) {
	t.Parallel()
	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset kubernetes.Interface
		namespace string
		svcName   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *apiv1.Service
		wantErr bool
	}{
		{
			name:   "retrieve valid service",
			fields: fields{Verbose: true},
			args: args{
				clientset: fake.NewSimpleClientset(&apiv1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-svc",
						Namespace: "test-ns",
					},
				}),
				namespace: "test-ns",
				svcName:   "test-svc",
			},
			want: &apiv1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: "test-ns",
				},
			},
			wantErr: false,
		},
		{
			name:   "retrieve non-existent service",
			fields: fields{Verbose: true},
			args: args{
				clientset: fake.NewSimpleClientset(),
				namespace: "test-ns",
				svcName:   "test-svc",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			got, err := p.GetService(tt.args.clientset, tt.args.namespace, tt.args.svcName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.GetService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientGoUtils.GetService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientGoUtils_CreatePreflightTestDeployment(t *testing.T) {
	t.Parallel()
	fk := fake.NewSimpleClientset()
	fk.Fake.PrependReactor("create", "deployments", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &appsv1.Deployment{}, errors.New("Error creating deployment")
	})

	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset kubernetes.Interface
		namespace string
		depName   string
		imageName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *appsv1.Deployment
		wantErr bool
	}{
		{
			name:   "create valid deployment",
			fields: fields{Verbose: true},
			args: args{
				clientset: fake.NewSimpleClientset(),
				namespace: "test-ns",
				depName:   "test-dep",
				imageName: "nginx",
			},
			want: &appsv1.Deployment{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-dep",
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
									Image: "nginx",
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
			},
			wantErr: false,
		},
		{
			name:   "invalid case - create deployment",
			fields: fields{Verbose: true},
			args: args{
				clientset: fk,
				namespace: "test-ns",
				depName:   "test-dep",
				imageName: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			got, err := p.CreatePreflightTestDeployment(tt.args.clientset, tt.args.namespace, tt.args.depName, tt.args.imageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.CreatePreflightTestDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientGoUtils.CreatePreflightTestDeployment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientGoUtils_DeleteService(t *testing.T) {
	t.Parallel()
	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset kubernetes.Interface
		namespace string
		name      string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid case",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(&apiv1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-svc",
						Namespace: "test-ns",
					},
				}),
				namespace: "test-ns",
				name:      "test-svc",
			},
			wantErr: false,
		},
		{
			name: "service not exists",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(),
				namespace: "test-ns",
				name:      "test-svc",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			if err := p.DeleteService(tt.args.clientset, tt.args.namespace, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.DeleteService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientGoUtils_DeletePod(t *testing.T) {
	t.Parallel()
	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset kubernetes.Interface
		namespace string
		name      string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid case",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(&apiv1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
					},
				}),
				namespace: "test-ns",
				name:      "test-pod",
			},
			wantErr: false,
		},
		{
			name: "pod not found",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(),
				namespace: "test-ns",
				name:      "test-pod",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			if err := p.DeletePod(tt.args.clientset, tt.args.namespace, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.DeletePod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientGoUtils_CreatePreflightTestPod(t *testing.T) {
	t.Parallel()
	fk := fake.NewSimpleClientset()
	fk.Fake.PrependReactor("create", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &apiv1.Pod{}, errors.New("Error creating pod")
	})

	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset    kubernetes.Interface
		namespace    string
		podName      string
		imageName    string
		secretNames  map[string]string
		commandToRun []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *apiv1.Pod
		wantErr bool
	}{
		{
			name:   "create valid pod without secret",
			fields: fields{Verbose: true},
			args: args{
				clientset:    fake.NewSimpleClientset(),
				namespace:    "test-ns",
				podName:      "test-pod",
				imageName:    "nginx",
				commandToRun: []string{"echo"},
			},
			want: &apiv1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-ns",
					Labels: map[string]string{
						"app": "preflight",
					},
				},
				Spec: apiv1.PodSpec{
					RestartPolicy: apiv1.RestartPolicyNever,
					Containers: []apiv1.Container{
						{
							Name:            "cnt",
							Image:           "nginx",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Command:         []string{"echo"},
						},
					},
				},
			},
		},
		{
			name:   "create valid pod with secret",
			fields: fields{Verbose: true},
			args: args{
				clientset:    fake.NewSimpleClientset(),
				namespace:    "test-ns",
				podName:      "test-pod",
				imageName:    "nginx",
				commandToRun: []string{"echo"},
				secretNames: map[string]string{
					"secret1": "/etc/secret1",
				},
			},
			want: &apiv1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-ns",
					Labels: map[string]string{
						"app": "preflight",
					},
				},
				Spec: apiv1.PodSpec{
					RestartPolicy: apiv1.RestartPolicyNever,
					Containers: []apiv1.Container{
						{
							Name:            "cnt",
							Image:           "nginx",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Command:         []string{"echo"},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "secret1",
									MountPath: filepath.Dir("/etc/secret1"),
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "secret1",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: "secret1",
									Items: []apiv1.KeyToPath{
										{
											Key:  "secret1",
											Path: filepath.Base("/etc/secret1"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:   "k8s error",
			fields: fields{Verbose: true},
			args: args{
				clientset:    fk,
				namespace:    "test-ns",
				podName:      "test-pod",
				imageName:    "nginx",
				commandToRun: []string{"echo"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			got, err := p.CreatePreflightTestPod(tt.args.clientset, tt.args.namespace, tt.args.podName, tt.args.imageName, tt.args.secretNames, tt.args.commandToRun)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.CreatePreflightTestPod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientGoUtils.CreatePreflightTestPod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientGoUtils_getPod(t *testing.T) {
	t.Parallel()
	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset kubernetes.Interface
		namespace string
		podName   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *apiv1.Pod
		wantErr bool
	}{
		{
			name: "valid case",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(&apiv1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
					},
				}),
				namespace: "test-ns",
				podName:   "test-pod",
			},
			want: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-ns",
				},
			},
			wantErr: false,
		},
		{
			name: "pod not found",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(),
				namespace: "test-ns",
				podName:   "test-pod",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			got, err := p.getPod(tt.args.clientset, tt.args.namespace, tt.args.podName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.getPod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientGoUtils.getPod() = %v, want %v", got, tt.want)
			}
		})
	}
}

// There is an issue with mocking logs: https://github.com/kubernetes/kubernetes/issues/84203
// We are waiting for this PR: https://github.com/kubernetes/kubernetes/pull/91485/files to be merged to be able to test this feature
func TestClientGoUtils_GetPodContainerLogs(t *testing.T) {
	t.SkipNow()
	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset kubernetes.Interface
		pod       *apiv1.Pod
		container string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		mockLog string
		wantErr bool
	}{
		{
			name:   "valid case",
			fields: fields{Verbose: true},
			args: args{
				clientset: fake.NewSimpleClientset(&apiv1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
					},
				}),
				pod: &apiv1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
					},
				},
				container: "",
			},
			want:    "blah",
			mockLog: "blah",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			got, err := p.GetPodContainerLogs(tt.args.clientset, tt.args.pod, tt.args.container)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.GetPodContainerLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ClientGoUtils.GetPodContainerLogs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientGoUtils_WaitForDeployment(t *testing.T) {
	t.Parallel()
	waitTimeout = 10 * time.Second
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dep",
			Namespace: "test-ns",
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: 0,
		},
	}

	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset    kubernetes.Interface
		namespace    string
		pfDeployment *appsv1.Deployment
	}
	tests := []struct {
		name                           string
		fields                         fields
		args                           args
		wantErr                        bool
		timeoutForChangingReplicaCount time.Duration
		mockErr                        bool
	}{
		{
			name: "valid case",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset:    fake.NewSimpleClientset(dep),
				namespace:    "test-ns",
				pfDeployment: dep,
			},
			wantErr:                        false,
			timeoutForChangingReplicaCount: 6 * time.Second,
			mockErr:                        false,
		},
		{
			name: "valid case instantly ready",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dep",
						Namespace: "test-ns",
					},
					Status: appsv1.DeploymentStatus{
						ReadyReplicas: 1,
					},
				}),
				namespace:    "test-ns",
				pfDeployment: dep,
			},
			wantErr: false,
			mockErr: false,
		},
		{
			name: "k8s returning error case",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset:    fake.NewSimpleClientset(dep),
				namespace:    "test-ns",
				pfDeployment: dep,
			},
			wantErr: true,
			mockErr: true,
		},
		{
			name: "timeout",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset:    fake.NewSimpleClientset(dep),
				namespace:    "test-ns",
				pfDeployment: dep,
			},
			wantErr: true,
			mockErr: false,
		},
		{
			name: "deployment goes missing",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset:    fake.NewSimpleClientset(),
				namespace:    "test-ns",
				pfDeployment: dep,
			},
			wantErr: true,
			mockErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeoutForChangingReplicaCount.Seconds() > 0 {
				go func() {
					time.Sleep(tt.timeoutForChangingReplicaCount)
					tt.args.clientset.(*fake.Clientset).Fake.PrependReactor("get", "deployments", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, &appsv1.Deployment{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-dep",
								Namespace: "test-ns",
							},
							Status: appsv1.DeploymentStatus{
								ReadyReplicas: 1,
							},
						}, nil
					})
				}()
			}
			if tt.mockErr {
				tt.args.clientset.(*fake.Clientset).Fake.PrependReactor("get", "deployments", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &appsv1.Deployment{}, fmt.Errorf("error")
				})
			}
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			if err := p.WaitForDeployment(tt.args.clientset, tt.args.namespace, tt.args.pfDeployment); (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.WaitForDeployment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientGoUtils_WaitForPod(t *testing.T) {
	t.Parallel()
	waitTimeout = 10 * time.Second
	po := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dep",
			Namespace: "test-ns",
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{},
			},
		},
		Status: apiv1.PodStatus{
			Phase: apiv1.PodPending,
		},
	}

	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset kubernetes.Interface
		namespace string
		pod       *apiv1.Pod
	}
	tests := []struct {
		name                           string
		fields                         fields
		args                           args
		wantErr                        bool
		timeoutForChangingReplicaCount time.Duration
		mockErr                        bool
	}{
		{
			name: "valid case",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(po),
				namespace: "test-ns",
				pod:       po,
			},
			wantErr:                        false,
			timeoutForChangingReplicaCount: 6 * time.Second,
			mockErr:                        false,
		},
		{
			name: "valid case instantly ready - pod running",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(&apiv1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dep",
						Namespace: "test-ns",
					},
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
							{},
						},
					},
					Status: apiv1.PodStatus{
						Phase: apiv1.PodRunning,
					},
				}),
				namespace: "test-ns",
				pod:       po,
			},
			wantErr: false,
			mockErr: false,
		},
		{
			name: "valid case instantly ready - pod succeeded",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(&apiv1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dep",
						Namespace: "test-ns",
					},
					Status: apiv1.PodStatus{
						Phase: apiv1.PodSucceeded,
					},
				}),
				namespace: "test-ns",
				pod:       po,
			},
			wantErr: false,
			mockErr: false,
		},
		{
			name: "valid case instantly ready - pod failed",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(&apiv1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dep",
						Namespace: "test-ns",
					},
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
							{},
						},
					},
					Status: apiv1.PodStatus{
						Phase: apiv1.PodFailed,
					},
				}),
				namespace: "test-ns",
				pod:       po,
			},
			wantErr: false,
			mockErr: false,
		},
		{
			name: "k8s returning error case",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(po),
				namespace: "test-ns",
				pod:       po,
			},
			wantErr: true,
			mockErr: true,
		},
		{
			name: "timeout",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(po),
				namespace: "test-ns",
				pod:       po,
			},
			wantErr: true,
			mockErr: false,
		},
		{
			name: "pod goes missing",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(),
				namespace: "test-ns",
				pod:       po,
			},
			wantErr: true,
			mockErr: false,
		},
		{
			name: "pod has no containers",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(&apiv1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dep",
						Namespace: "test-ns",
					},
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{},
					},
				}),
				namespace: "test-ns",
				pod:       po,
			},
			wantErr: true,
			mockErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeoutForChangingReplicaCount.Seconds() > 0 {
				go func() {
					time.Sleep(tt.timeoutForChangingReplicaCount)
					tt.args.clientset.(*fake.Clientset).Fake.PrependReactor("get", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, &apiv1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-dep",
								Namespace: "test-ns",
							},
							Status: apiv1.PodStatus{
								Phase: apiv1.PodRunning,
							},
						}, nil
					})
				}()
			}
			if tt.mockErr {
				tt.args.clientset.(*fake.Clientset).Fake.PrependReactor("get", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &apiv1.Pod{}, fmt.Errorf("error")
				})
			}
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			if err := p.WaitForPod(tt.args.clientset, tt.args.namespace, tt.args.pod); (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.WaitForPod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientGoUtils_WaitForPodToDie(t *testing.T) {
	t.Parallel()
	waitTimeout = 10 * time.Second
	po := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dep",
			Namespace: "test-ns",
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{},
			},
		},
		Status: apiv1.PodStatus{
			Phase: apiv1.PodPending,
		},
	}

	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset kubernetes.Interface
		namespace string
		pod       *apiv1.Pod
	}
	tests := []struct {
		name                           string
		fields                         fields
		args                           args
		wantErr                        bool
		timeoutForChangingReplicaCount time.Duration
		mockErr                        bool
	}{
		{
			name: "valid case",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(po),
				namespace: "test-ns",
				pod:       po,
			},
			wantErr:                        false,
			timeoutForChangingReplicaCount: 6 * time.Second,
			mockErr:                        false,
		},
		{
			name: "valid case instantly ready - pod succeeded",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(&apiv1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dep",
						Namespace: "test-ns",
					},
					Status: apiv1.PodStatus{
						Phase: apiv1.PodSucceeded,
					},
				}),
				namespace: "test-ns",
				pod:       po,
			},
			wantErr: false,
			mockErr: false,
		},
		{
			name: "valid case instantly ready - pod failed",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(&apiv1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dep",
						Namespace: "test-ns",
					},
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
							{},
						},
					},
					Status: apiv1.PodStatus{
						Phase: apiv1.PodFailed,
					},
				}),
				namespace: "test-ns",
				pod:       po,
			},
			wantErr: false,
			mockErr: false,
		},
		{
			name: "k8s returning error case",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(po),
				namespace: "test-ns",
				pod:       po,
			},
			wantErr: true,
			mockErr: true,
		},
		{
			name: "timeout",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(po),
				namespace: "test-ns",
				pod:       po,
			},
			wantErr: false,
			mockErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeoutForChangingReplicaCount.Seconds() > 0 {
				go func() {
					time.Sleep(tt.timeoutForChangingReplicaCount)
					tt.args.clientset.(*fake.Clientset).Fake.PrependReactor("get", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, &apiv1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-dep",
								Namespace: "test-ns",
							},
							Status: apiv1.PodStatus{
								Phase: apiv1.PodRunning,
							},
						}, nil
					})
				}()
			}
			if tt.mockErr {
				tt.args.clientset.(*fake.Clientset).Fake.PrependReactor("get", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &apiv1.Pod{}, fmt.Errorf("error")
				})
			}
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			if err := p.WaitForPodToDie(tt.args.clientset, tt.args.namespace, tt.args.pod); (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.WaitForPod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientGoUtils_waitForPodToDelete(t *testing.T) {
	t.Parallel()
	waitTimeout = 10 * time.Second
	po := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-ns",
		},
	}

	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset kubernetes.Interface
		namespace string
		pod       string
	}
	tests := []struct {
		name                           string
		fields                         fields
		args                           args
		wantErr                        bool
		timeoutForChangingReplicaCount time.Duration
	}{
		{
			name: "valid case",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(po),
				namespace: "test-ns",
				pod:       "test-pod",
			},
			wantErr:                        false,
			timeoutForChangingReplicaCount: 6 * time.Second,
		},
		{
			name: "valid case instant",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(),
				namespace: "test-ns",
				pod:       "test-pod",
			},
			wantErr: false,
		},
		{
			name: "timeout",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(po),
				namespace: "test-ns",
				pod:       "test-pod",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeoutForChangingReplicaCount.Seconds() > 0 {
				go func() {
					time.Sleep(tt.timeoutForChangingReplicaCount)
					tt.args.clientset.(*fake.Clientset).Fake.PrependReactor("get", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, &apiv1.Pod{}, fmt.Errorf("error")
					})
				}()
			}
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			if err := p.waitForPodToDelete(tt.args.clientset, tt.args.namespace, tt.args.pod); (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.WaitForPod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientGoUtils_WaitForDeploymentToDelete(t *testing.T) {
	t.Parallel()
	waitTimeout = 10 * time.Second
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dep",
			Namespace: "test-ns",
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: 0,
		},
	}
	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset      kubernetes.Interface
		namespace      string
		deploymentName string
	}
	tests := []struct {
		name                           string
		fields                         fields
		args                           args
		wantErr                        bool
		timeoutForChangingReplicaCount time.Duration
	}{
		{
			name: "valid case",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset:      fake.NewSimpleClientset(dep),
				namespace:      "test-ns",
				deploymentName: dep.Name,
			},
			wantErr:                        false,
			timeoutForChangingReplicaCount: 6 * time.Second,
		},
		{
			name: "valid case instant",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset:      fake.NewSimpleClientset(),
				namespace:      "test-ns",
				deploymentName: dep.Name,
			},
			wantErr: false,
		},
		{
			name: "timeout",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset:      fake.NewSimpleClientset(dep),
				namespace:      "test-ns",
				deploymentName: dep.Name,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeoutForChangingReplicaCount.Seconds() > 0 {
				go func() {
					time.Sleep(tt.timeoutForChangingReplicaCount)
					tt.args.clientset.(*fake.Clientset).Fake.PrependReactor("get", "deployments", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, &appsv1.Deployment{}, fmt.Errorf("error")
					})
				}()
			}
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			if err := p.WaitForDeploymentToDelete(tt.args.clientset, tt.args.namespace, tt.args.deploymentName); (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.WaitForDeploymentToDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientGoUtils_CreatePreflightTestSecret(t *testing.T) {
	t.Parallel()
	fk := fake.NewSimpleClientset()
	fk.Fake.PrependReactor("create", "secrets", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &apiv1.Secret{}, errors.New("Error creating deployment")
	})
	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset  kubernetes.Interface
		namespace  string
		secretName string
		secretData []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *apiv1.Secret
		wantErr bool
	}{
		{
			name:   "create valid deployment",
			fields: fields{Verbose: true},
			args: args{
				clientset:  fake.NewSimpleClientset(),
				namespace:  "test-ns",
				secretName: "test-name",
				secretData: []byte("hello"),
			},
			want: &apiv1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-name",
					Namespace: "test-ns",
					Labels: map[string]string{
						"app": "preflight",
					},
				},
				Data: map[string][]byte{
					"test-name": []byte("hello"),
				},
			},
			wantErr: false,
		},
		{
			name:   "invalid case - create secret",
			fields: fields{Verbose: true},
			args: args{
				clientset:  fk,
				namespace:  "test-ns",
				secretName: "test-name",
				secretData: []byte("hello"),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			got, err := p.CreatePreflightTestSecret(tt.args.clientset, tt.args.namespace, tt.args.secretName, tt.args.secretData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.CreatePreflightTestSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientGoUtils.CreatePreflightTestSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientGoUtils_DeleteK8sSecret(t *testing.T) {
	t.Parallel()
	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset  kubernetes.Interface
		namespace  string
		secretName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "valid case",
			fields: fields{Verbose: true},
			args: args{
				clientset: fake.NewSimpleClientset(&apiv1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-name",
						Namespace: "test-ns",
					},
				}),
				namespace:  "test-ns",
				secretName: "test-name",
			},
			wantErr: false,
		},
		{
			name:   "secret not found",
			fields: fields{Verbose: true},
			args: args{
				clientset:  fake.NewSimpleClientset(),
				namespace:  "test-ns",
				secretName: "test-name",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			if err := p.DeleteK8sSecret(tt.args.clientset, tt.args.namespace, tt.args.secretName); (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.DeleteK8sSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientGoUtils_CreateStatefulSet(t *testing.T) {
	t.Parallel()
	fk := fake.NewSimpleClientset()
	fk.Fake.PrependReactor("create", "statefulsets", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &appsv1.StatefulSet{}, errors.New("Error")
	})
	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset       kubernetes.Interface
		namespace       string
		statefulSetName string
		imageName       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *appsv1.StatefulSet
		wantErr bool
	}{
		{
			name: "valid case",
			args: args{
				clientset:       fake.NewSimpleClientset(),
				namespace:       "test-ns",
				statefulSetName: "test-sf",
				imageName:       "nginx",
			},
			want: &appsv1.StatefulSet{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-sf",
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
									Command:         []string{"bash", "-c", "for i in {1..10}; do echo \"from init container...\"; sleep 1; exit 1; done"},
								},
							},
							Containers: []apiv1.Container{
								{
									Name:  "statefulset",
									Image: "nginx",
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
			},
			wantErr: false,
		},
		{
			name: "k8s error",
			args: args{
				clientset:       fk,
				namespace:       "test-ns",
				statefulSetName: "test-sf",
				imageName:       "nginx",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			got, err := p.CreateStatefulSet(tt.args.clientset, tt.args.namespace, tt.args.statefulSetName, tt.args.imageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.CreateStatefulSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientGoUtils.CreateStatefulSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientGoUtils_waitForStatefulSet(t *testing.T) {
	t.Parallel()
	waitTimeout = 10 * time.Second
	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ss",
			Namespace: "test-ns",
		},
		Status: appsv1.StatefulSetStatus{
			ReadyReplicas: 0,
		},
	}

	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset     kubernetes.Interface
		namespace     string
		pfStatefulset *appsv1.StatefulSet
	}
	tests := []struct {
		name                           string
		fields                         fields
		args                           args
		wantErr                        bool
		timeoutForChangingReplicaCount time.Duration
		mockErr                        bool
	}{
		{
			name: "valid case",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset:     fake.NewSimpleClientset(ss),
				namespace:     "test-ns",
				pfStatefulset: ss,
			},
			wantErr:                        false,
			timeoutForChangingReplicaCount: 6 * time.Second,
			mockErr:                        false,
		},
		{
			name: "valid case instantly ready",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset: fake.NewSimpleClientset(&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ss",
						Namespace: "test-ns",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
					},
				}),
				namespace:     "test-ns",
				pfStatefulset: ss,
			},
			wantErr: false,
			mockErr: false,
		},
		{
			name: "k8s returning error case",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset:     fake.NewSimpleClientset(ss),
				namespace:     "test-ns",
				pfStatefulset: ss,
			},
			wantErr: true,
			mockErr: true,
		},
		{
			name: "timeout",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset:     fake.NewSimpleClientset(ss),
				namespace:     "test-ns",
				pfStatefulset: ss,
			},
			wantErr: true,
			mockErr: false,
		},
		{
			name: "statefulset goes missing",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset:     fake.NewSimpleClientset(),
				namespace:     "test-ns",
				pfStatefulset: ss,
			},
			wantErr: true,
			mockErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeoutForChangingReplicaCount.Seconds() > 0 {
				go func() {
					time.Sleep(tt.timeoutForChangingReplicaCount)
					tt.args.clientset.(*fake.Clientset).Fake.PrependReactor("get", "statefulsets", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, &appsv1.StatefulSet{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-ss",
								Namespace: "test-ns",
							},
							Status: appsv1.StatefulSetStatus{
								ReadyReplicas: 1,
							},
						}, nil
					})
				}()
			}
			if tt.mockErr {
				tt.args.clientset.(*fake.Clientset).Fake.PrependReactor("get", "statefulsets", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &appsv1.StatefulSet{}, fmt.Errorf("error")
				})
			}
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			if err := p.waitForStatefulSet(tt.args.clientset, tt.args.namespace, tt.args.pfStatefulset); (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.waitForStatefulSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientGoUtils_getStatefulset(t *testing.T) {
	t.Parallel()
	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset       kubernetes.Interface
		namespace       string
		statefulsetName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *appsv1.StatefulSet
		wantErr bool
	}{
		{
			name:   "valid case",
			fields: fields{Verbose: true},
			args: args{
				clientset: fake.NewSimpleClientset(&appsv1.StatefulSet{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-ss",
						Namespace: "test-ns",
					},
				}),
				namespace:       "test-ns",
				statefulsetName: "test-ss",
			},
			want: &appsv1.StatefulSet{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-ss",
					Namespace: "test-ns",
				},
			},
			wantErr: false,
		},
		{
			name:   "retrieve non-existent ss",
			fields: fields{Verbose: true},
			args: args{
				clientset:       fake.NewSimpleClientset(),
				namespace:       "test-ns",
				statefulsetName: "test-ss",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			got, err := p.getStatefulset(tt.args.clientset, tt.args.namespace, tt.args.statefulsetName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.getStatefulset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientGoUtils.getStatefulset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientGoUtils_deleteStatefulSet(t *testing.T) {
	t.Parallel()
	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset kubernetes.Interface
		namespace string
		name      string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "valid case",
			fields: fields{Verbose: true},
			args: args{
				clientset: fake.NewSimpleClientset(&appsv1.StatefulSet{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-ss",
						Namespace: "test-ns",
					},
				}),
				name:      "test-ss",
				namespace: "test-ns",
			},
			wantErr: false,
		},
		{
			name:   "delete non-existent ss case",
			fields: fields{Verbose: true},
			args: args{
				clientset: fake.NewSimpleClientset(),
				name:      "test-ss",
				namespace: "test-ns",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			if err := p.deleteStatefulSet(tt.args.clientset, tt.args.namespace, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.deleteStatefulSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientGoUtils_waitForStatefulsetToDelete(t *testing.T) {
	t.Parallel()
	waitTimeout = 10 * time.Second
	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ss",
			Namespace: "test-ns",
		},
		Status: appsv1.StatefulSetStatus{
			ReadyReplicas: 0,
		},
	}
	type fields struct {
		Verbose bool
	}
	type args struct {
		clientset       kubernetes.Interface
		namespace       string
		statefulsetName string
	}
	tests := []struct {
		name                           string
		fields                         fields
		args                           args
		wantErr                        bool
		timeoutForChangingReplicaCount time.Duration
	}{
		{
			name: "valid case",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset:       fake.NewSimpleClientset(ss),
				namespace:       "test-ns",
				statefulsetName: ss.Name,
			},
			wantErr:                        false,
			timeoutForChangingReplicaCount: 6 * time.Second,
		},
		{
			name: "valid case instant",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset:       fake.NewSimpleClientset(),
				namespace:       "test-ns",
				statefulsetName: ss.Name,
			},
			wantErr: false,
		},
		{
			name: "timeout",
			fields: fields{
				Verbose: true,
			},
			args: args{
				clientset:       fake.NewSimpleClientset(ss),
				namespace:       "test-ns",
				statefulsetName: ss.Name,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeoutForChangingReplicaCount.Seconds() > 0 {
				go func() {
					time.Sleep(tt.timeoutForChangingReplicaCount)
					tt.args.clientset.(*fake.Clientset).Fake.PrependReactor("get", "statefulsets", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, &appsv1.StatefulSet{}, fmt.Errorf("error")
					})
				}()
			}
			p := &ClientGoUtils{
				Verbose: tt.fields.Verbose,
			}
			if err := p.waitForStatefulsetToDelete(tt.args.clientset, tt.args.namespace, tt.args.statefulsetName); (err != nil) != tt.wantErr {
				t.Errorf("ClientGoUtils.waitForStatefulsetToDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
