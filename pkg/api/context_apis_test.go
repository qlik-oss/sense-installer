package api

import (
	"reflect"
	"testing"

	"github.com/qlik-oss/k-apis/pkg/config"
)

var (
	testDir = "./tests"
)

func TestAddCommonConfig(t *testing.T) {
	type args struct {
		qliksenseCR *QliksenseCR
		contextName string
	}
	tests := []struct {
		name string
		args args
		want *QliksenseCR
	}{
		{
			name: "valid case",
			args: args{
				qliksenseCR: &QliksenseCR{},
				contextName: "myqliksense",
			},
			want: &QliksenseCR{
				CommonConfig: CommonConfig{
					ApiVersion: QliksenseContextApiVersion,
					Kind:       QliksenseContextKind,
					Metadata: &Metadata{
						Name: "myqliksense",
					},
				},
				Spec: &config.CRSpec{
					Profile:     QliksenseDefaultProfile,
					ReleaseName: "myqliksense",
					RotateKeys:  DefaultRotateKeys,
					Secrets: map[string]config.NameValues{
						"qliksense": []config.NameValue{{
							Name:  DefaultMongoDbUriKey,
							Value: DefaultMongoDbUri,
						},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.qliksenseCR.AddCommonConfig(tt.args.contextName)
			if !reflect.DeepEqual(tt.args.qliksenseCR, tt.want) {
				t.Errorf("AddCommonConfig() = %+v, want %+v", tt.args.qliksenseCR, tt.want)
			}
		})
	}
}

func TestAddBaseQliksenseConfigs(t *testing.T) {
	type args struct {
		qliksenseConfig         *QliksenseConfig
		defaultQliksenseContext string
	}
	tests := []struct {
		name string
		args args
		want *QliksenseConfig
	}{
		{
			name: "valid case",
			args: args{
				qliksenseConfig:         &QliksenseConfig{},
				defaultQliksenseContext: "qlik-default",
			},
			want: &QliksenseConfig{
				CommonConfig: CommonConfig{
					ApiVersion: QliksenseConfigApiVersion,
					Kind:       QliksenseConfigKind,
					Metadata: &Metadata{
						Name: QliksenseMetadataName,
					},
				},
				Spec: &ContextSpec{
					CurrentContext: "qlik-default",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.qliksenseConfig.AddBaseQliksenseConfigs(tt.args.defaultQliksenseContext)
			if !reflect.DeepEqual(tt.args.qliksenseConfig, tt.want) {
				t.Errorf("AddBaseQliksenseConfigs() = %+v, want %+v", tt.args.qliksenseConfig, tt.want)
			}
		})
	}
}
