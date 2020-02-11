package qliksense

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/qlik-oss/k-apis/pkg/config"
	"github.com/qlik-oss/sense-installer/pkg/api"
)

var (
	testDir = "./tests"
)

func TestAddCommonConfig(t *testing.T) {
	type args struct {
		qliksenseCR api.QliksenseCR
		contextName string
	}
	tests := []struct {
		name string
		args args
		want api.QliksenseCR
	}{
		{
			name: "valid case",
			args: args{
				qliksenseCR: api.QliksenseCR{},
				contextName: "myqliksense",
			},
			want: api.QliksenseCR{
				CommonConfig: api.CommonConfig{
					ApiVersion: QliksenseContextApiVersion,
					Kind:       QliksenseContextKind,
					Metadata: &api.Metadata{
						Name: "myqliksense",
					},
				},
				Spec: &config.CRSpec{
					Profile:     QliksenseDefaultProfile,
					ReleaseName: "myqliksense",
					RotateKeys:  DefaultRotateKeys,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddCommonConfig(tt.args.qliksenseCR, tt.args.contextName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddCommonConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddBaseQliksenseConfigs(t *testing.T) {
	type args struct {
		qliksenseConfig         api.QliksenseConfig
		defaultQliksenseContext string
	}
	tests := []struct {
		name string
		args args
		want api.QliksenseConfig
	}{
		{
			name: "valid case",
			args: args{
				qliksenseConfig:         api.QliksenseConfig{},
				defaultQliksenseContext: "qlik-default",
			},
			want: api.QliksenseConfig{
				CommonConfig: api.CommonConfig{
					ApiVersion: QliksenseConfigApiVersion,
					Kind:       QliksenseConfigKind,
					Metadata: &api.Metadata{
						Name: QliksenseMetadataName,
					},
				},
				Spec: &api.ContextSpec{
					CurrentContext: "qlik-default",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddBaseQliksenseConfigs(tt.args.qliksenseConfig, tt.args.defaultQliksenseContext); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddBaseQliksenseConfigs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func setup() func() {
	// create tests dir
	if err := os.Mkdir(testDir, 0777); err != nil {
		fmt.Printf("\nError occurred: %v", err)
	}
	config :=
		`
apiVersion: config.qlik.com/v1
kind: QliksenseConfig
metadata:
  name: qliksenseConfig
spec:
  contexts:
  - name: qlik-default
    crLocation: /root/.qliksense/contexts/qlik-default.yaml
  currentContext: qlik-default
`
	configFile := filepath.Join(testDir, "config.yaml")
	// tests/config.yaml exists
	ioutil.WriteFile(configFile, []byte(config), 0777)

	contextYaml :=
		`
apiVersion: qlik.com/v1
kind: Qliksense
metadata:
  name: qlik-default
spec:
  profile: docker-desktop
  rotateKeys: "yes"
  releaseName: qlik-default
`
	qlikDefaultContext := "qlik-default"
	// create contexts/qlik-default/ under tests/
	contexts := "contexts"
	contextsDir := filepath.Join(testDir, contexts, qlikDefaultContext)
	if err := os.MkdirAll(contextsDir, 0777); err != nil {
		err = fmt.Errorf("Not able to create directories")
		log.Fatal(err)
	}

	contextFile := filepath.Join(contextsDir, qlikDefaultContext+".yaml")
	ioutil.WriteFile(contextFile, []byte(contextYaml), 0777)
	tearDown := func() {
		os.RemoveAll(testDir)
	}
	return tearDown
}

func Test_retrieveCurrentContextInfo(t *testing.T) {

	tearDown := setup()
	defer tearDown()

	q := &Qliksense{
		QliksenseHome: testDir,
	}
	_, _, err := retrieveCurrentContextInfo(q)
	if err != nil {
		t.FailNow()
	}
}

func TestSetUpQliksenseContext(t *testing.T) {
	type args struct {
		qlikSenseHome    string
		contextName      string
		isDefaultContext bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid contextname",
			args: args{
				qlikSenseHome:    testDir,
				contextName:      "testContext1",
				isDefaultContext: false,
			},
			wantErr: false,
		},
		{
			name: "invalid contextname",
			args: args{
				qlikSenseHome:    testDir,
				contextName:      "testContext_abcdefgh",
				isDefaultContext: false,
			},
			wantErr: true,
		},
	}
	tearDown := setup()
	defer tearDown()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetUpQliksenseContext(tt.args.qlikSenseHome, tt.args.contextName, tt.args.isDefaultContext); (err != nil) != tt.wantErr {
				t.Errorf("SetUpQliksenseContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetUpQliksenseDefaultContext(t *testing.T) {
	type args struct {
		qlikSenseHome string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid case",
			args: args{
				qlikSenseHome: testDir,
			},
			wantErr: false,
		},
	}
	tearDown := setup()
	defer tearDown()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetUpQliksenseDefaultContext(tt.args.qlikSenseHome); (err != nil) != tt.wantErr {
				t.Errorf("SetUpQliksenseDefaultContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetOtherConfigs(t *testing.T) {
	type args struct {
		q    *Qliksense
		args []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid case",
			args: args{
				q: &Qliksense{
					QliksenseHome: testDir,
				},
				args: []string{"profile=minikube", "namespace=qliksense", "storageClassName=efs"},
			},
			wantErr: false,
		},
	}
	tearDown := setup()
	defer tearDown()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetOtherConfigs(tt.args.q, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("SetOtherConfigs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetConfigs(t *testing.T) {
	type args struct {
		q    *Qliksense
		args []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid case",
			args: args{
				q: &Qliksense{
					QliksenseHome: testDir,
				},
				args: []string{"qliksense[name=acceptEULA]=\"yes\"", "qliksense[name=mongoDbUri]=\"mongo://mongo:3307\""},
			},
			wantErr: false,
		},
	}
	tearDown := setup()
	defer tearDown()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetConfigs(tt.args.q, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("SetConfigs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
