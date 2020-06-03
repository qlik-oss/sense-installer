package api

import (
	"errors"
	"reflect"
	"testing"

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
