package preflight

import (
	"fmt"
	"testing"
)

func Test_initiateK8sOps(t *testing.T) {
	t.Skip()
	type args struct {
		opr       string
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid case",
			args: args{
				opr:       fmt.Sprintf("version"),
				namespace: "test-ns",
			},
			wantErr: false,
		},
		{
			name: "invalid case",
			args: args{
				opr:       fmt.Sprintf("versions"),
				namespace: "test-ns",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := initiateK8sOps(tt.args.opr, tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("initiateK8sOps() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
