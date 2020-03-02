package qliksense

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/qlik-oss/sense-installer/pkg/api"
)

const (
	testDir            = "./tests"
	qlikDefaultContext = "qlik-default"
	secrets            = "secrets"
	contexts           = "contexts"
)

var targetFileStringTemplate = `
apiVersion: v1
data:
  testKey1: %s
kind: Secret
metadata:
  name: ctxname-testSvcName-sense_installer
type: Opaque
`
var encText = "eiGjS2UkxXY/cLoA90hnddXCeJSZj8TF7gIHRq0c+4mIpBVFRLRo79pAMuEORFobRLQnPiwLUkLQ/BNLpA1tu8hsFaiQwIv6/iP1mNepdF2ha9Tf6XVlTYCbQJ2mmHOY0TQk/d1QwXa73+PXz0paLMB+y9/39w7SThL8NHbIKxGAs4rXurVIlmoOaXJmshUCYmEUFV26B9Y4yVQJKfOlheslYrbqVWXhA9lFa/r74yUJnYSrluj9D32eY1xJvI1tUR2oQRUuscAZ8W0v3SjoyUwiXV4L4mJb8qiNx+15PfVMK9V7LVdoRbka+rR2lOtFxQOk6mw41s244AGeU81scgt0PdFiAuNtc2Z42jF92kY4rxRrBqwx6b/oMXUuYAldU4hkNyfNGVINeqQbyMyMcBL1FSe/QuutbInCPDTODQTeZKQB58cgNl/cCORSBYuDPj8dXkPyyUU1+XGju+UEUxS6SG9WLfdozW5Q4FeBVIblS3oraQ0y4S4DYc2r44I7yEmklO3O1leSnCFCpTWMiRC++j5KDkKLYDoL/SXUTh3jMu19z4TXcZ6LCbfbA6hgEh3fcju4XCwXM8NDSzJomLsPmdsvscoqiNgJKjHJNDVHm/1VH0vzSHwMdhoWYXLbM2iX3ucV7q/OEPkLjQF0p5/9XYo0zyIa8uxcZjbimlY="
var decText = "Value1234"

func setupTargetFileAndPrivateKey() (string, error) {
	targetFileString := fmt.Sprintf(targetFileStringTemplate, encText)
	privKeyBytes := []byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`)

	publicKeyBytes := []byte(`-----BEGIN RSA PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAwUCimKCidbF3UxEHPy8K
+hvhklRB9JYhj5sJy0if4lTVibkK1MrYCykOnmC40pPU9GLY1b8HxAg9tvyRn0YH
UxOra6vVQaVcOVJhTM8D18d+lSr3Lp1yiX+UGT4nzWI9+R1CCbwXrqeQVoZs6QZK
ynEXMkFI9/wNMOwPOvQFOSTuoEoCO+zyTyUWEkNbUq825ELUQdIsjgmlWUOONudx
sAr7ESRXW9QTHVh6uWmr3VRKZHby1JdU3I/wjdlGg5M2dDuXy5nQO9w/nYLjJXiw
+zzOetZ/+t7/VOkOpNTeJQhwTM1WF7Y2VLetbi9FHgyzHatrduh07+XEiTbgDf3G
Ix2bp2p6oh0G3N2zpiLcK/aZj8rouWWydfFfsU3MZ4FfJDP8I6b9awxjmKYqIr6h
iPQCJaLBED8mwK+I5evIbnKv6E6uK+BApWA/R7ElragoFYbqQ1VpvntVMtJt9Dy5
ZrI+IQARdXD3bb34oh0IPBhClnvvMUc1cWxDoXEX6oJ4I+LzxE87Zkwnan9qOwen
golMVKFwPx1o37qrbmrXID21kKt7FL6xN4HxHLkItr1fKzdyWDFRHgASTAWfx5BI
wvPuUW0vZHkvO80VyV2L63whVhPnPASmFkbviomrBttYfpr2aGQqF/qR1Nlxe834
MFxk1pS9LMa/WnzvFr0gWakCAwEAAQ==
-----END RSA PUBLIC KEY-----
`)

	targetFile := filepath.Join(testDir, "targetfile.yaml")
	// tests/config.yaml exists
	err := ioutil.WriteFile(targetFile, []byte(targetFileString), 0777)
	if err != nil {
		log.Printf("Error while creating file: %v", err)
		return "",err
	}

	secretKeyPairDir := filepath.Join(testDir, secrets, contexts, qlikDefaultContext, secrets)
	if err := os.MkdirAll(secretKeyPairDir, 0777); err != nil {
		err = fmt.Errorf("Not able to create directories")
		log.Fatal(err)
	}
	os.Setenv("QLIKSENSE_KEY_LOCATION", secretKeyPairDir)

	privKeyFile := filepath.Join(secretKeyPairDir, "qliksensePriv")
	// construct and write priv key file into secretsDir location
	err = ioutil.WriteFile(privKeyFile, privKeyBytes, 0777)
	if err != nil {
		log.Printf("Error while creating file: %v", err)
		return "",err
	}
	pubKeyFile := filepath.Join(secretKeyPairDir, "qliksensePub")
	// construct and write pub key file into secretsDir location
	err = ioutil.WriteFile(pubKeyFile, publicKeyBytes, 0777)
	if err != nil {
		log.Printf("Error while creating file: %v", err)
		return "",err
	}
	return targetFile, nil
}

func removePrivateKey() {
	err:=os.Remove(filepath.Join(testDir, secrets, contexts, qlikDefaultContext, secrets, "qliksensePriv"))
	if err!=nil{
		log.Fatalf("Could not delete private key %v",err)
	}
return
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

func TestQliksense_PrepareK8sSecret(t *testing.T) {

	type fields struct {
		QliksenseHome        string
	}
	type args struct {
		qliksenseCR api.QliksenseCR
		targetFile  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
		setup func() (string, func())
	}{
		{
			name: "valid case",
			fields: fields{
				QliksenseHome: testDir,
			},
			args: args{
				qliksenseCR: api.QliksenseCR{
					CommonConfig: api.CommonConfig{
						ApiVersion: api.QliksenseContextApiVersion,
						Kind:       api.QliksenseContextKind,
						Metadata: &api.Metadata{
							Name: qlikDefaultContext,
						},
					},
				},
			},
			want:    fmt.Sprintf(targetFileStringTemplate, base64.StdEncoding.EncodeToString([]byte(decText))),
			wantErr: false,
			setup: func() (string, func()) {
				tearDown := setup()
				targetFile, _:= setupTargetFileAndPrivateKey()
				return targetFile, tearDown
			},
		},
		{
			name: "private key not supplied should result in decryption error",
			fields: fields{
				QliksenseHome: testDir,
			},
			args: args{
				qliksenseCR: api.QliksenseCR{
					CommonConfig: api.CommonConfig{
						ApiVersion: api.QliksenseContextApiVersion,
						Kind:       api.QliksenseContextKind,
						Metadata: &api.Metadata{
							Name: qlikDefaultContext,
						},
					},
				},
			},
			want:  "",
			wantErr: true,
			setup: func() (string, func()) {
				tearDown := setup()
				targetFile, _:= setupTargetFileAndPrivateKey()
				removePrivateKey()
				return targetFile, tearDown
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetFile, tearDown := tt.setup()

			q := &Qliksense{
				QliksenseHome: tt.fields.QliksenseHome,
			}
			got, err := q.PrepareK8sSecret(tt.args.qliksenseCR, targetFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Qliksense.PrepareK8sSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(strings.TrimSpace(got), strings.TrimSpace(tt.want)) {
				t.Errorf("Qliksense.PrepareK8sSecret() = %v, want %v", got, tt.want)
			}
			tearDown()
		})
	}
}
