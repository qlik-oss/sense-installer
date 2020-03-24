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
				namespace: "ash-ns",
			},
			wantErr: false,
		},
		{
			name: "invalid case",
			args: args{
				opr:       fmt.Sprintf("versions"),
				namespace: "ash-ns",
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

func Test_determinePlatformSpecificUrls(t *testing.T) {
	type args struct {
		platform string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name: "valid platform",
			args: args{
				platform: "windows",
			},
			want:    fmt.Sprintf("%s%s", preflightBaseURL, preflightWindowsFile),
			want1:   preflightWindowsFile,
			wantErr: false,
		},
		{
			name: "invalid platform",
			args: args{
				platform: "unix",
			},
			want:    "",
			want1:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := determinePlatformSpecificUrls(tt.args.platform)
			if (err != nil) != tt.wantErr {
				t.Errorf("determinePlatformSpecificUrls() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("determinePlatformSpecificUrls() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("determinePlatformSpecificUrls() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
