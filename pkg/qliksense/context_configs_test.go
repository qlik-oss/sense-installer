package qliksense

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/qlik-oss/sense-installer/pkg/api"
)

var (
	testDir = "./tests"
)

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
  - name: qlik1
	crLocation: /root/.qliksense/contexts/qlik1.yaml
  - name: qlik2
	crLocation: /root/.qliksense/contexts/qlik2.yaml
  currentContext: qlik1
`
	configFile := filepath.Join(testDir, "config.yaml")
	// tests/config.yaml exists
	ioutil.WriteFile(configFile, []byte(config), 0777)

	contexts := "contexts"
	contextsDir := filepath.Join(testDir, contexts)
	if err := os.MkdirAll(contextsDir, 0777); err != nil {
		err = fmt.Errorf("Not able to create directories")
		log.Fatal(err)
	}
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
	contextYaml2 :=
		`
	apiVersion: qlik.com/v1
	kind: Qliksense
	metadata:
	  name: qlik1
	spec:
	  profile: docker-desktop
	  rotateKeys: "yes"
	  releaseName: qlik1
	`

	contextYaml3 :=
		`
	apiVersion: qlik.com/v1
	kind: Qliksense
	metadata:
	  name: qlik2
	spec:
	  profile: docker-desktop
	  rotateKeys: "yes"
	  releaseName: qlik2
	`
	contextfile := "./tests/contexts"
	createYaml("qlik-default", contextYaml, contextfile)
	createYaml("qlik1", contextYaml2, contextfile)
	createYaml("qlik2", contextYaml3, contextfile)

	tearDown := func() {
		os.RemoveAll(testDir)
	}
	return tearDown
}

func createYaml(context string, contextYaml string, contextfile string) {
	contextsDir := filepath.Join(contextfile, context)
	if err := os.MkdirAll(contextsDir, 0777); err != nil {
		err = fmt.Errorf("Not able to create directories")
		log.Fatal(err)
	}

	contextFile := filepath.Join(contextsDir, context+".yaml")
	ioutil.WriteFile(contextFile, []byte(contextYaml), 0777)
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

func TestSetImageRegistry(t *testing.T) {
	getQlikSense := func(tmpQlikSenseHome string) (*Qliksense, error) {
		if err := ioutil.WriteFile(path.Join(tmpQlikSenseHome, "config.yaml"), []byte(fmt.Sprintf(`
apiVersion: config.qlik.com/v1
kind: QliksenseConfig
metadata:
  name: QliksenseConfigMetadata
spec:
  contexts:
  - name: qlik-default
    crFile: %s/contexts/qlik-default/qlik-default.yaml
  currentContext: qlik-default
`, tmpQlikSenseHome)), os.ModePerm); err != nil {
			return nil, err
		}

		defaultContextDir := path.Join(tmpQlikSenseHome, "contexts", "qlik-default")
		if err := os.MkdirAll(defaultContextDir, os.ModePerm); err != nil {
			return nil, err
		}

		version := "foo"
		manifestsRootDir := fmt.Sprintf("%s/repo/%s", defaultContextDir, version)
		if err := ioutil.WriteFile(path.Join(defaultContextDir, "qlik-default.yaml"), []byte(fmt.Sprintf(`
apiVersion: qlik.com/v1
kind: Qliksense
metadata:
  name: qlik-default
  labels:
    version: %s
spec:
  profile: docker-desktop
  manifestsRoot: %s
  namespace: some-namespace
`, version, manifestsRootDir)), os.ModePerm); err != nil {
			return nil, err
		}
		return &Qliksense{
			QliksenseHome: tmpQlikSenseHome,
		}, nil
	}
	testCases := []struct {
		name               string
		registry           string
		pushUsername       string
		pushPassword       string
		pullUsername       string
		pullPassword       string
		expectSecretsExist bool
	}{
		{
			name:               "no auth",
			registry:           "foobar",
			pushUsername:       "",
			pushPassword:       "",
			pullUsername:       "",
			pullPassword:       "",
			expectSecretsExist: false,
		},
		{
			name:               "auth",
			registry:           "foobar",
			pushUsername:       "foo-push",
			pushPassword:       "bar-push",
			pullUsername:       "foo-pull",
			pullPassword:       "bar-pull",
			expectSecretsExist: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tmpQlikSenseHome, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer os.RemoveAll(tmpQlikSenseHome)

			q, err := getQlikSense(tmpQlikSenseHome)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if err := q.SetImageRegistry(testCase.registry, testCase.pushUsername, testCase.pushPassword,
				testCase.pullUsername, testCase.pullPassword); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			qConfig := api.NewQConfig(q.QliksenseHome)
			if testCase.expectSecretsExist {
				if pushSecret, err := qConfig.GetPushDockerConfigJsonSecret(); err != nil {
					t.Fatalf("unexpected error: %v", err)
				} else if pushSecret.Uri != testCase.registry ||
					pushSecret.Username != testCase.pushUsername || pushSecret.Password != testCase.pushPassword {
					t.Fatalf("unexpected push secret content: %v", pushSecret)
				}
				if pullSecret, err := qConfig.GetDockerConfigJsonSecret("image-registry-pull-secret.yaml"); err != nil {
					t.Fatalf("unexpected error: %v", err)
				} else if pullSecret.Uri != testCase.registry ||
					pullSecret.Name != "artifactory-docker-secret" || pullSecret.Namespace != "some-namespace" ||
					pullSecret.Username != testCase.pullUsername || pullSecret.Password != testCase.pullPassword {
					t.Fatalf("unexpected pull secret content: %v", pullSecret)
				}
			} else {
				if _, err := qConfig.GetPushDockerConfigJsonSecret(); err == nil {
					t.Fatal("unexpected image-registry-push-secret.yaml")
				} else if _, err := qConfig.GetDockerConfigJsonSecret("image-registry-pull-secret.yaml"); err == nil {
					t.Fatal("unexpected image-registry-pull-secret.yaml")
				}
			}
		})
	}
}

func TestDeleteContexts(t *testing.T) {
	type args struct {
		qlikSenseHome string
		contextName   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid context",
			args: args{
				qlikSenseHome: testDir,
				contextName:   "qlik2",
			},
			wantErr: false,
		},
		{
			name: "default context",
			args: args{
				qlikSenseHome: testDir,
				contextName:   "qlik-default",
			},
			wantErr: true,
		},
		{
			name: "non-existent context",
			args: args{
				qlikSenseHome: testDir,
				contextName:   "qlik3",
			},
			wantErr: true,
		},
		{
			name: "current context",
			args: args{
				qlikSenseHome: testDir,
				contextName:   "qlik1",
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
			var arg []string
			arg[0] = tt.args.contextName
			if err := q.DeleteContextConfig(arg); (err != nil) != tt.wantErr {
				t.Errorf("DeleteContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}
