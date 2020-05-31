package api

import (
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
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
			name:   "valid case",
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
			name:   "error case",
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
			name:   "valid case",
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
			name:   "valid case",
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
