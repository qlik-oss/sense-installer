package api

import (
	"reflect"
	"strings"
	"testing"

	"github.com/qlik-oss/k-apis/pkg/config"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	testDir = "./tests"
)

func TestAddCommonConfig(t *testing.T) {
	gvk := schema.GroupVersionKind{
		Group:   QliksenseGroup,
		Kind:    QliksenseKind,
		Version: QliksenseApiVersion,
	}
	q := &QliksenseCR{}
	q.SetName("myqliksense")
	q.SetGroupVersionKind(gvk)
	q.Spec = &config.CRSpec{
		Profile:    QliksenseDefaultProfile,
		RotateKeys: DefaultRotateKeys,
		Secrets: map[string]config.NameValues{
			"qliksense": []config.NameValue{{
				Name:  DefaultMongodbUriKey,
				Value: strings.Replace(DefaultMongodbUri, "qlik-default", "myqliksense", 1),
			},
			},
		},
	}

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
			want: q,
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
	gvk := schema.GroupVersionKind{
		Group:   QliksenseConfigApiGroup,
		Kind:    QliksenseConfigKind,
		Version: QliksenseConfigApiVersion,
	}
	qc := &QliksenseConfig{}
	qc.SetGroupVersionKind(gvk)
	qc.SetName(QliksenseMetadataName)
	qc.Spec = &ContextSpec{
		CurrentContext: "qlik-default",
	}

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
			want: qc,
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
