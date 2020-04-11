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
var decText = "mongodb://qlik-default-mongodb:27017/qliksense?ssl=false"

func setupTargetFileAndPrivateKey() (string, string, error) {

	secretKeyLocation := filepath.Join(testDir, secrets, contexts, qlikDefaultContext, secrets)
	if err := os.MkdirAll(secretKeyLocation, 0777); err != nil {
		err = fmt.Errorf("Not able to create directories")
		log.Fatal(err)
	}
	os.Setenv("QLIKSENSE_KEY_LOCATION", secretKeyLocation)

	//privKeyFile := filepath.Join(secretKeyLocation, "user_secret_key")
	key, err := api.LoadSecretKey(secretKeyLocation)
	if key == "" {
		key, err = api.GenerateAndStoreSecretKey(secretKeyLocation)
	}
	encData, _ := api.EncryptData([]byte(decText), key)
	encText := b64.StdEncoding.EncodeToString(encData)

	targetFileString := fmt.Sprintf(targetFileStringTemplate, encText)
	targetFile := filepath.Join(testDir, "targetfile.yaml")
	// tests/config.yaml exists
	err = ioutil.WriteFile(targetFile, []byte(targetFileString), 0777)
	if err != nil {
		log.Printf("Error while creating file: %v", err)
		return "", "", err
	}

	return targetFile, key, err
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
	qConfig := api.NewQConfig(q.QliksenseHome)
	_, err := qConfig.GetCurrentCR()
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
				args: []string{"profile=minikube", "rotateKeys=yes", "storageClassName=efs", "gitops.enabled=yes", "gitops.schedule=30 * * * *", "git.repository=master", "git.username=foo", "git.accesstoken=1234"},
			},
			wantErr: false,
		},
		{
			name: "invalid configs",
			args: args{
				q: &Qliksense{
					QliksenseHome: testDir,
				},
				args: []string{"someconfig=somevalue, gitops.schedule=bar", "gitops.enabled=bar", "git.foo=bar", "rotatekeys=bar"},
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
			if err := tt.args.q.SetConfigs(tt.args.args, false); (err != nil) != tt.wantErr {
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
func removePrivateKey() {
	err := os.Remove(filepath.Join(testDir, secrets, contexts, qlikDefaultContext, secrets, "user_secret_key"))
	if err != nil {
		log.Fatalf("Could not delete private key %v", err)
	}
	return
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
				targetFile, _, _ := setupTargetFileAndPrivateKey()
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
				targetFile, _, _ := setupTargetFileAndPrivateKey()
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
				setupTargetFileAndPrivateKey()
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
		base64      bool
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
			name: "valid secret secrets=false base64 encoded",
			fields: fields{
				QliksenseHome: testDir,
			},
			args: args{
				args:        []string{"qliksense.mongoDbUri=bW9uZ29kYjovL3FsaWstZGVmYXVsdC1tb25nb2RiOjI3MDE3L3FsaWtzZW5zZT9zc2w9ZmFsc2U="},
				isSecretSet: false,
				base64:      true,
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
	_, encryptionKey, err := setupTargetFileAndPrivateKey()
	if err != nil {
		t.FailNow()
	}
	defer tearDown()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Qliksense{
				QliksenseHome: tt.fields.QliksenseHome,
			}
			if err := q.SetSecrets(tt.args.args, tt.args.isSecretSet, tt.args.base64); (err != nil) != tt.wantErr {
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
			if tt.args.base64 {
				d, _ := b64.StdEncoding.DecodeString(testValue)
				testValue = strings.Trim(string(d), "\n ")
			}
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
						decryptedVal, err := api.DecryptData([]byte(valToBeEncrypted), encryptionKey)
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
		b, err := b64.RawStdEncoding.DecodeString(item.Value)
		return string(b), err
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
			if err := q.DeleteContextConfig(arg); (err != nil) != tt.wantErr {
				t.Errorf("DeleteContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}
