package qliksense

import (
	"encoding/base64"
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/qlik-oss/k-apis/pkg/config"

	"github.com/qlik-oss/sense-installer/pkg/api"
	"gopkg.in/yaml.v2"
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
  mongoDbUri: %s
kind: Secret
metadata:
  name: testctx-qliksense-senseinstaller
type: Opaque
`
var encText = "SFpVZ2t5SGsrN2lLQjlTYm9rbFUxSDFRcmVYdUxhTW9MUHlQOGtGditxMEcwZTlIZDl1dVRrV0tEYm5qUURSWFp3dStuNklueGk3anI2c1djSVdsbWlKTHdWQUJwdUg0a1NXd3llMUlMa2oxK3FRSFlMM2dQUExvN1pBYkVDeDROMUVvam12M0t0NmQwbkdhSXlWWEpmWWJUVVFDM1Y4L0ZTVXBVN0JUb0l4OVZWdmlPam5HTHk4RlF2a3RUaHJxWTUvZEh2N3pVUmhiOTc2Q2YwbEovZ3I2L2NwRk9RMUFXVXdodVhrTG9lYjVzNFdtTEZzNldqT3k0bWlKM1J6VllLaWVUSFJ2SE85eDB6dUthanRwSGEzWEZkaE5QNnpySVJJNTRFalUyblVYYUNlYXVnWnZEOUxjdWluOFhFcjExbkFINURCUDAycXhoZk5BejVoMlV2eFNWVmR0aW1QTDBhMVBJTUxGQTgyWUkrQkFOQkhkSUNnZGU5SkxIRFBoTzR6c0llaE1LRmhVQkNoOUhQa3kyRnhTeDJ3YWp3M1UycEsvcFJVZUxDazRUbkhmL25LN3h5ekdpV3dSUFFFZHdsWE5JbUhjVlVPV3gvNWh4WlJCUTZtb3pGYk1HbXR1Mkh5Z3RVV2gzNFYzd1BhS01TNFRsa0hyODFjRjVCWVpxenBFK1pKWnVyLy8zbzJsU0tFMjMxTG1pcGk1K0FqbXZvUVcyWHBocjFNVWJQY1pXUkJFRkkyQXBCM0FhQXFPa0k1MkRqNG43Mko5bCtaMzdydTk1aHk5K1lzY0FxMjZVbExYRlc0S3RUUkRLSjlMNnVmdlIrUUNudER3em5UTFRHUnEwZU5COWt6S0Q4MFlUdXozeHNXK3cxdjlHbDJaMnBZMTZWTCtEV1k9"
var decText = "mongodb://qlik-default-mongodb:27017/qliksense?ssl=false"

func setupTargetFileAndPrivateKey() ([]byte, []byte, string, error) {
	targetFileString := fmt.Sprintf(targetFileStringTemplate, encText)
	privKeyBytes := []byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`)

	publicKeyBytes := []byte(`-----BEGIN RSA PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEApFf3qCQhAr2QLRRZdhLy
B8exLjrQiXLr8hwDe0xHSLJX3w7v5z4ujJWrHUulQ2/hvS8uffxMVrp2YqeA4sjy
/ku1KqZVQv/WNTdL2v9Z1ewbRnBjDQvmkWDKZ+cP8VPdHGzQ4iM2z6BZ4RQTkdQM
KqsVwUsLO9amI2TOny/M696eFRW0pk4+W3QZZRawT0HqJPvKKXKqoO2+62W54rOV
8glJi29Do06e4S6CZUl1hBUy0VlLtrLlRSOHTois0dF2a9f7+GGgU11MHO6w1k1N
esSlfZ0vnrrkW8WFLqewk+Jj+w1xeQnYHOeHjx6zi9f9DC96eSylSB3iJ71NmXcM
c0IEZ2LiQIqTL83BLOgMBCsK3FSlGMakQUR2GJ+M0I/selYkRMhid6eOmhlsTNMP
bpcTHxZ+ReIzS+5B1X0FZ7RIL+jSL9mFcxmD3dxurrrt/DkLpXcuWdi1s1bpVn3j
IQhU0+bgA6hT0k8Kj2f3Q9QnvkHSEff+XyLvLwQeSsSAcnM+1I7fNSPEo2cq0au6
ZtjHcirXmMminAQ6cKW1XrEvJBefHHibtjJjIM4bHH7MKRA5H3km/J4CCwI1VogS
TcE05Z6ypAFU2TCrnec9c9VXkRDP94h0GuoG8sdhmQudqvghr/8T7CV+sGRQbdeq
rXwanfnGPrjcMVIO/dSOxOUCAwEAAQ==
-----END RSA PUBLIC KEY-----
`)

	targetFile := filepath.Join(testDir, "targetfile.yaml")
	// tests/config.yaml exists
	err := ioutil.WriteFile(targetFile, []byte(targetFileString), 0777)
	if err != nil {
		log.Printf("Error while creating file: %v", err)
		return nil, nil, "", err
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
		return nil, nil, "", err
	}
	pubKeyFile := filepath.Join(secretKeyPairDir, "qliksensePub")
	api.LogDebugMessage("Test setup - \npub key path: %s\n, priv key path: %s\n", pubKeyFile, privKeyFile)
	// construct and write pub key file into secretsDir location
	err = ioutil.WriteFile(pubKeyFile, publicKeyBytes, 0777)
	if err != nil {
		log.Printf("Error while creating file: %v", err)
		return nil, nil, "", err
	}
	return publicKeyBytes, privKeyBytes, targetFile, nil
}

func removePrivateKey() {
	err := os.Remove(filepath.Join(testDir, secrets, contexts, qlikDefaultContext, secrets, "qliksensePriv"))
	if err != nil {
		log.Fatalf("Could not delete private key %v", err)
	}
	return
}

func setup() func() {
	// create tests dir
	os.RemoveAll(testDir)
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
    crFile: contexts/qlik-default/qlik-default.yaml
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

func readCRFile() (*api.QliksenseCR, error) {
	qlikDefaultContext := "qlik-default"
	qliksenseCR := &api.QliksenseCR{}
	contextFileContents, err := ioutil.ReadFile(filepath.Join(testDir, contexts, qlikDefaultContext, qlikDefaultContext+".yaml"))
	if err != nil {
		log.Println(err)
		err = fmt.Errorf("Not able to read current context info")
		return nil, err
	}
	if err := yaml.Unmarshal(contextFileContents, qliksenseCR); err != nil {
		err = fmt.Errorf("An error occurred during unmarshalling: %v", err)
		return nil, err
	}
	return qliksenseCR, nil
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
		{
			name: "empty contextname",
			args: args{
				qlikSenseHome:    testDir,
				contextName:      "",
				isDefaultContext: false,
			},
			wantErr: true,
		},
	}
	tearDown := setup()
	defer tearDown()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := New(tt.args.qlikSenseHome)
			if err := q.SetUpQliksenseContext(tt.args.contextName); (err != nil) != tt.wantErr {
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
			q := New(tt.args.qlikSenseHome)
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
				args: []string{"profile=minikube", "rotateKeys=yes", "storageClassName=efs", "gitops.enabled=yes", "gitops.schedule=30 * * * *"},
			},
			wantErr: false,
		},
		{
			name: "invalid configs",
			args: args{
				q: &Qliksense{
					QliksenseHome: testDir,
				},
				args: []string{"someconfig=somevalue, gitops.schedule=bar", "gitops.enabled=bar"},
			},
			wantErr: true,
		},
		{
			name: "empty configs",
			args: args{
				q: &Qliksense{
					QliksenseHome: testDir,
				},
				args: []string{},
			},
			wantErr: true,
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
				args: []string{"qliksense.acceptEULA=\"yes\"", "qliksense.mongoDbUri=\"mongo://mongo:3307\""},
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
		if err := ioutil.WriteFile(path.Join(tmpQlikSenseHome, "config.yaml"), []byte(`
apiVersion: config.qlik.com/v1
kind: QliksenseConfig
metadata:
  name: QliksenseConfigMetadata
spec:
  contexts:
  - name: qlik-default
    crFile: contexts/qlik-default/qlik-default.yaml
  currentContext: qlik-default
`), os.ModePerm); err != nil {
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
				if pullSecret, err := qConfig.GetPullDockerConfigJsonSecret(); err != nil {
					t.Fatalf("unexpected error: %v", err)
				} else if pullSecret.Uri != testCase.registry ||
					pullSecret.Name != "artifactory-docker-secret" ||
					pullSecret.Username != testCase.pullUsername || pullSecret.Password != testCase.pullPassword {
					t.Fatalf("unexpected pull secret content: %v", pullSecret)
				}
			} else {
				if _, err := qConfig.GetPushDockerConfigJsonSecret(); err == nil {
					t.Fatal("unexpected image-registry-push-secret.yaml")
				} else if _, err := qConfig.GetPullDockerConfigJsonSecret(); err == nil {
					t.Fatal("unexpected image-registry-pull-secret.yaml")
				}
			}
		})
	}
}

func Test_PrepareK8sSecret(t *testing.T) {

	type fields struct {
		QliksenseHome string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
		setup   func() (string, func())
	}{
		{
			name: "valid case",
			fields: fields{
				QliksenseHome: testDir,
			},
			want:    fmt.Sprintf(targetFileStringTemplate, base64.StdEncoding.EncodeToString([]byte(decText))),
			wantErr: false,
			setup: func() (string, func()) {
				tearDown := setup()
				_, _, targetFile, _ := setupTargetFileAndPrivateKey()
				return targetFile, tearDown
			},
		},
		{
			name: "private key not supplied should result in decryption error",
			fields: fields{
				QliksenseHome: testDir,
			},
			want:    "",
			wantErr: true,
			setup: func() (string, func()) {
				tearDown := setup()
				_, _, targetFile, _ := setupTargetFileAndPrivateKey()
				removePrivateKey()
				return targetFile, tearDown
			},
		},
		{
			name: "target file not supplied",
			fields: fields{
				QliksenseHome: testDir,
			},
			want:    "",
			wantErr: true,
			setup: func() (string, func()) {
				tearDown := setup()
				_, _, _, _ = setupTargetFileAndPrivateKey()
				removePrivateKey()
				return "", tearDown
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetFile, tearDown := tt.setup()

			q := &Qliksense{
				QliksenseHome: tt.fields.QliksenseHome,
			}
			got, err := q.PrepareK8sSecret(targetFile)
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

func Test_ListContextConfigs(t *testing.T) {
	type fields struct {
		QliksenseHome string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		setup   func() (string, func())
	}{
		{
			name: "valid case",
			fields: fields{
				QliksenseHome: testDir,
			},
			wantErr: false,
			setup: func() (string, func()) {
				tearDown := setup()
				return "", tearDown
			},
		},
		{
			name: "config yaml does not exist",
			fields: fields{
				QliksenseHome: testDir,
			},
			wantErr: true,
			setup: func() (string, func()) {
				return "", func() {}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, tearDown := tt.setup()

			q := &Qliksense{
				QliksenseHome: tt.fields.QliksenseHome,
			}
			if err := q.ListContextConfigs(); (err != nil) != tt.wantErr {
				t.Errorf("ListContextConfigs() error = %v, wantErr %v", err, tt.wantErr)
			}
			tearDown()
		})
	}
}

func Test_SetSecrets(t *testing.T) {
	type fields struct {
		QliksenseHome string
	}
	type args struct {
		args        []string
		isSecretSet bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid secret secrets=false",
			fields: fields{
				QliksenseHome: testDir,
			},
			args: args{
				args:        []string{"qliksense.mongoDbUri=\"mongodb://qlik-default-mongodb:27017/qliksense?ssl=false\""},
				isSecretSet: false,
			},
			wantErr: false,
		},
		{
			name: "test1 valid secret secrets=true",
			fields: fields{
				QliksenseHome: testDir,
			},
			args: args{
				args:        []string{"qliksense.mongoDbUri=\"mongo://mongo:3307\""},
				isSecretSet: true,
			},
			wantErr: false,
		},
		{
			name: "test2 valid secret secrets=true",
			fields: fields{
				QliksenseHome: testDir,
			},
			args: args{
				args:        []string{"qliksense.mongoDbUri=\"mongodb://qlik-default-mongodb:27017/qliksense?ssl=false\""},
				isSecretSet: true,
			},
			wantErr: false,
		},
		{
			name: "invalid secret secrets=false",
			fields: fields{
				QliksenseHome: testDir,
			},
			args: args{
				args:        []string{},
				isSecretSet: false,
			},
			wantErr: true,
		},
	}
	tearDown := setup()
	_, privateKeyBytes, _, err := setupTargetFileAndPrivateKey()
	if err != nil {
		t.FailNow()
	}
	defer tearDown()

	privKey, err := api.DecodeToPrivateKey(privateKeyBytes)
	if err != nil {
		t.FailNow()
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Qliksense{
				QliksenseHome: tt.fields.QliksenseHome,
			}
			if err := q.SetSecrets(tt.args.args, tt.args.isSecretSet); (err != nil) != tt.wantErr {
				t.Errorf("SetSecrets() error = %v, wantErr %v", err, tt.wantErr)
				t.FailNow()
			}
			if tt.wantErr || len(tt.args.args) == 0 {
				return
			}
			// VERIFICATION PART BELOW
			// extract the value for testing
			testValueArr := strings.SplitN(tt.args.args[0], "=", 2)
			testValue := strings.ReplaceAll(testValueArr[1], "\"", "")

			qliksenseCR, err := readCRFile()
			if err != nil {
				err = fmt.Errorf("Not able to read from context file: %v", err)
				log.Println(err)
				t.FailNow()
			}

			for svcName := range qliksenseCR.Spec.Secrets { // we are sure we only have one service
				for _, v := range qliksenseCR.Spec.Secrets {
					for _, item := range v { // we are sure we only have one entry
						valToBeEncrypted, err := getValueToBeDecodedForSetSecrets(item, qliksenseCR, svcName)
						if err != nil {
							err := fmt.Errorf("Error occurred while decoding: %v", err)
							log.Printf("decode error: %v", err)
							t.FailNow()
						}
						decodedValue, err := b64.StdEncoding.DecodeString(valToBeEncrypted)
						if err != nil {
							err := fmt.Errorf("Error occurred while decoding: %v", err)
							log.Printf("decode error: %v", err)
							t.FailNow()
						}
						decryptedVal, err := api.Decrypt(decodedValue, privKey)
						if err != nil {
							err := fmt.Errorf("Error occurred while testing decryption: %v", err)
							log.Printf("No Data in Secret: %v", err)
							t.FailNow()
						}
						if string(decryptedVal) != testValue {
							t.FailNow()
						}
					}

				}
			}
		})
	}
}

func getValueToBeDecodedForSetSecrets(item config.NameValue, qliksenseCR *api.QliksenseCR, svcName string) (string, error) {
	if item.ValueFrom != nil && item.ValueFrom.SecretKeyRef != nil {
		// secret=true
		secretFilePath := filepath.Join(testDir, contexts, qliksenseCR.GetName(), QliksenseSecretsDir, svcName+".yaml")
		if api.FileExists(secretFilePath) {
			secretFileContents, err := ioutil.ReadFile(secretFilePath)
			if err != nil {
				err = fmt.Errorf("An error occurred during unmarshalling: %v", err)
				return "", err
			}
			k8sSecret, err := api.K8sSecretFromYaml(secretFileContents)
			if err != nil {
				err = fmt.Errorf("An error occurred during unmarshalling: %v", err)
				return "", err
			}
			if k8sSecret.Data == nil {
				err = fmt.Errorf("No Data in Secret: %v", err)
				return "", err
			}
			return string(k8sSecret.Data[item.ValueFrom.SecretKeyRef.Key]), nil
		}
	}
	// secret=false
	if item.Value != "" {
		return item.Value, nil
	}
	err := fmt.Errorf("Both Value and ValueFrom are empty")
	return "", err
}

func setupDeleteContext() func() {
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
    crFile: contexts/qlik-default.yaml
  - name: qlik1
    crFile: contexts/qlik1.yaml
  - name: qlik2
    crFile: contexts/qlik2.yaml
  currentContext: qlik1
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
	contextsDir1 := filepath.Join(testDir, contexts, qlikDefaultContext)
	if err := os.MkdirAll(contextsDir1, 0777); err != nil {
		err = fmt.Errorf("Not able to create directories")
		log.Fatal(err)
	}
	contextYaml1 :=
		`
apiVersion: qlik.com/v1
kind: Qliksense
metadata:
  name: qlik1
spec:
  profile: docker-desktop
  rotateKeys: "yes"
  releaseName: qlik1`

	contextYaml2 :=
		`
apiVersion: qlik.com/v1
kind: Qliksense
metadata:
  name: qlik2
spec:
  profile: docker-desktop
  rotateKeys: "yes"
  releaseName: qlik2`

	contextsDir := filepath.Join(testDir, contexts, "qlik1")
	if err := os.MkdirAll(contextsDir, 0777); err != nil {
		err = fmt.Errorf("Not able to create directories")
		log.Fatal(err)
	}

	contextsDir2 := filepath.Join(testDir, contexts, "qlik2")
	if err := os.MkdirAll(contextsDir2, 0777); err != nil {
		err = fmt.Errorf("Not able to create directories")
		log.Fatal(err)
	}

	contextFile := filepath.Join(contextsDir, "qlik1.yaml")
	ioutil.WriteFile(contextFile, []byte(contextYaml1), 0777)

	contextFile2 := filepath.Join(contextsDir2, "qlik2.yaml")
	ioutil.WriteFile(contextFile2, []byte(contextYaml2), 0777)

	contextFile1 := filepath.Join(contextsDir1, "qlik-default.yaml")
	ioutil.WriteFile(contextFile1, []byte(contextYaml), 0777)

	tearDown := func() {
		os.RemoveAll(testDir)
	}
	return tearDown
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
	tearDown := setupDeleteContext()
	defer tearDown()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := New(tt.args.qlikSenseHome)
			var arg []string
			arg = append(arg, tt.args.contextName)
			if err := q.DeleteContextConfig(arg, true); (err != nil) != tt.wantErr {
				t.Errorf("DeleteContext() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}
