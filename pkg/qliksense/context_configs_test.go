package qliksense

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gobuffalo/packr/v2"
	"github.com/qlik-oss/sense-installer/pkg/api"
)

var (
	testDir = "./tests"
)

func setupTargetFile() {
	targetFileString :=
		`
	apiVersion: v1
data:
  testKey1: WTFuMWtOYURhNndvS3JCZ05PbWtHbWpkdy91bEVnbGowTGw0QWIxc0phWDBBUmtWQWpTRHNmbFc5ajlFMkQvN3MyUjJvaDN3Z25YSWUwb3lxamxYdUIxa2NNU2M4UEZEa1YrZ3ZDOUFzdzhmcHBVQ0lBS2h1M2lTdzZKRWlKWURLUmplQnJ5dmtxcUdpQ2xtSXlac29QZGdlNWI5Ynh1M0lWN1JYdjdPdExZWjhOS000MHJCTE04dTFRbDU4RHhvVEc0dkdINHpsajczNVQra0w1U3NSWkQraDhxc2tvN0pZMzBBMEFVTngyK1FrbGpCeWdWR0VteUZHUy9xQk95Y284TWFsWlZPd2dnaThLMDZIMnBCdDBEMG5DdzV6UGZyMHpEbmdiT2V3ZS9XYjNlZlJOSEdoMStsTlNYTzhNd1RJc3RMbWdYbDRBTnJ3QXEyenhGbWcvNmUrc0FIYVFGc1pNd0tuVWxvSkpNUUFkSm1XdWJRNDNPOTlDZEVjeFBOOWw0Mi9Uc0Y5NVp5dzRtM0ZTRTB1bVdxbjQzeXNsOXVuQUhOL3FodG9tZlZOS08vdm1yTERSYzJjV1hDOTlNcEVjc0J0amNPeFNFcVBCNkNDUDZIMXdCR0NNcE5zS2lncms1aXVGL1BxZllOVEZaeGh1QjFBeVdlbUsybGpsWSt1Tjc3aWRNcEJQV29OVTVMR0pTb3AzRzNaajVMbXhmTXluYy9tRU1INnFmUGg3TVV2MmdsVTZ5N1puRUtpUkkrcmRIdU9EM2pPOThoWEJVcS9JK1N5VGNzS3VFQ2dVTlp0UlZXaGRkQitPck0vamYyQnZPc20yRFZWcjIremlzK0hhano3cmg2Z29YUmdnVHpxajBNdXFZK1gyeDBidE54ckZBaHdKQUhvbTQ9
kind: Secret
metadata:
  name: ctxname-testSvcName-sense_installer
type: Opaque
`
	targetFile := filepath.Join(testDir, "secretfile.yaml")
	// tests/config.yaml exists
	ioutil.WriteFile(targetFile, []byte(targetFileString), 0777)
}
func setup() func() {
	// create tests dir
	if err := os.Mkdir(testDir, 0777); err != nil {
		log.Printf("\nError occurred: %v", err)
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
			q, err := New(tt.args.qlikSenseHome)
			if err != nil {
				t.Errorf("unable to create a qliksense instance")
				return
			}
			if err := q.SetUpQliksenseContext(tt.args.contextName, tt.args.isDefaultContext); (err != nil) != tt.wantErr {
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
			q, err := New(tt.args.qlikSenseHome)
			if err != nil {
				t.Errorf("unable to create a qliksense instance")
				return
			}
			if err := q.SetUpQliksenseDefaultContext(); (err != nil) != tt.wantErr {
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
			if err := tt.args.q.SetOtherConfigs(tt.args.args); (err != nil) != tt.wantErr {
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
			if err := tt.args.q.SetConfigs(tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("SetConfigs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQliksense_PrepareK8sSecret(t *testing.T) {
	type fields struct {
		QliksenseHome        string
		QliksenseEjsonKeyDir string
		CrdBox               *packr.Box
	}
	type args struct {
		qliksenseCR *api.QliksenseCR
		targetFile  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "valid case",
			args: args{
				qliksenseCR: &QliksenseCR{
					CommonConfig: &CommonConfig{
						ApiVersion: QliksenseContextApiVersion,
						Kind:       QliksenseContextKind,
						Metadata: &Metadata{
							Name: "myqliksense",
						},
					},
				},
				targetFile: "secretfile.yaml",
			},
			want:    map[string]string{},
			wantErr: false,
		},
	}
	tearDown := setup()
	setupTargetFile()
	defer tearDown()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Qliksense{
				QliksenseHome:        tt.fields.QliksenseHome,
				QliksenseEjsonKeyDir: tt.fields.QliksenseEjsonKeyDir,
				CrdBox:               tt.fields.CrdBox,
			}
			got, err := q.PrepareK8sSecret(tt.args.qliksenseCR, tt.args.targetFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Qliksense.PrepareK8sSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Qliksense.PrepareK8sSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}
