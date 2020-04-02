package qliksense

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/Shopify/ejson"
	"github.com/qlik-oss/k-apis/pkg/config"
	"github.com/qlik-oss/k-apis/pkg/qust"

	kapis_git "github.com/qlik-oss/k-apis/pkg/git"
)

func Test_ExecuteKustomizeBuild(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	kustomizationYamlFilePath := path.Join(tmpDir, "kustomization.yaml")
	kustomizationYaml := `
generatorOptions:
  disableNameSuffixHash: true
configMapGenerator:
- name: foo-config
  literals:    
  - foo=bar
`
	err = ioutil.WriteFile(kustomizationYamlFilePath, []byte(kustomizationYaml), os.ModePerm)
	if err != nil {
		t.Fatalf("error writing kustomization file to path: %v error: %v\n", kustomizationYamlFilePath, err)
	}

	result, err := ExecuteKustomizeBuild(tmpDir)
	if err != nil {
		t.Fatalf("unexpected kustomize error: %v\n", err)
	}

	expectedK8sYaml := `apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  name: foo-config
`
	if string(result) != expectedK8sYaml {
		t.Fatalf("expected k8s yaml: [%v] but got: [%v]\n", expectedK8sYaml, string(result))
	}
}

func Test_executeKustomizeBuild_onQlikConfig_regenerateKeys(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := path.Join(tmpDir, "config")
	if repo, err := kapis_git.CloneRepository(configPath, defaultConfigRepoGitUrl, nil); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	} else if err := kapis_git.Checkout(repo, "v1.21.23-edge", "", nil); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	cr := &config.CRSpec{
		ManifestsRoot: configPath,
	}

	if err := os.Setenv("EJSON_KEYDIR", tmpDir); err != nil {
		t.Fatalf("unexpected error setting EJSON_KEYDIR environment variable: %v\n", err)
	}

	if err := os.Unsetenv("EJSON_KEY"); err != nil {
		t.Fatalf("unexpected error unsetting EJSON_KEY: %v\n", err)
	}

	generateKeys(cr, "won't-use")

	yamlResources, err := ExecuteKustomizeBuild(path.Join(configPath, "manifests", "base", "resources", "users"))
	if err != nil {
		t.Fatalf("unexpected kustomize error: %v\n", err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(yamlResources))
	var resource map[string]interface{}
	keyIdBase64 := ""
	for {
		err := decoder.Decode(&resource)
		if err != nil {
			if err != io.EOF {
				t.Fatalf("unexpected yaml decode error: %v\n", err)
			}
			break
		}
		if resource["kind"].(string) == "Secret" && strings.Contains(resource["metadata"].(map[string]interface{})["name"].(string), "users-secrets-") {
			keyIdBase64 = resource["data"].(map[string]interface{})["tokenAuthPrivateKeyId"].(string)
			break
		}
	}

	untransformedKeyId := `(( (ds "data").kid ))`
	if keyIdBase64 == "" {
		t.Fatalf("expected keyIdBase64 for users secret to be non empty:\n")
	} else if keyId, err := base64.StdEncoding.DecodeString(keyIdBase64); err != nil {
		t.Fatalf("unexpected base64 decode error: %v\n", err)
	} else if string(keyId) == untransformedKeyId {
		t.Fatalf("unexpected users keyId: %v\n", untransformedKeyId)
	}
}

func generateKeys(cr *config.CRSpec, defaultKeyDir string) {
	log.Println("rotating all keys")
	keyDir := getEjsonKeyDir(defaultKeyDir)
	if ejsonPublicKey, ejsonPrivateKey, err := ejson.GenerateKeypair(); err != nil {
		log.Printf("error generating an ejson key pair: %v\n", err)
	} else if err := qust.GenerateKeys(cr, ejsonPublicKey); err != nil {
		log.Printf("error generating application keys: %v\n", err)
	} else if err := os.MkdirAll(keyDir, os.ModePerm); err != nil {
		log.Printf("error makeing sure private key storage directory: %v exists, error: %v\n", keyDir, err)
	} else if err := ioutil.WriteFile(path.Join(keyDir, ejsonPublicKey), []byte(ejsonPrivateKey), os.ModePerm); err != nil {
		log.Printf("error storing ejson private key: %v\n", err)
	}
}

func getEjsonKeyDir(defaultKeyDir string) string {
	ejsonKeyDir := os.Getenv("EJSON_KEYDIR")
	if ejsonKeyDir == "" {
		ejsonKeyDir = defaultKeyDir
	}
	return ejsonKeyDir
}

func Test_GetYamlDocKindFromMultiDoc(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	kustomizationYamlFilePath := path.Join(tmpDir, "kustomization.yaml")
	testResFileYamlFilePath := path.Join(tmpDir, "test-file.yaml")
	kustomizationYaml := `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- test-file.yaml
`
	testYaml := `
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  name: foo-config
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: qix-sessions
    chart: qix-sessions-4.0.10
    heritage: Helm
    release: qliksense
  name: qliksense-qix-sessions
  namespace: default
   ---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  labels:
    app: chronos
    chart: chronos-1.5.7
    heritage: Helm
    release: qliksense
  name: qliksense-chronos
  namespace: default
  rules:
  - apiGroups:
    - ""
    resources:
    - endpoints
    verbs:
    - get
    - update
`
	err = ioutil.WriteFile(kustomizationYamlFilePath, []byte(kustomizationYaml), os.ModePerm)
	if err != nil {
		t.Fatalf("error writing kustomization file to path: %v error: %v\n", kustomizationYamlFilePath, err)
	}
	err = ioutil.WriteFile(testResFileYamlFilePath, []byte(testYaml), os.ModePerm)
	if err != nil {
		t.Fatalf("error writing test-file to path: %v error: %v\n", testResFileYamlFilePath, err)
	}
	result, err := ExecuteKustomizeBuild(tmpDir)
	if err != nil {
		t.Fatalf("unexpected kustomize error: %v\n", err)
	}
	resultYaml := GetYamlsFromMultiDoc(string(result), "Role")

	expectedK8sYaml := `
---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  labels:
    app: chronos
    chart: chronos-1.5.7
    heritage: Helm
    release: qliksense
  name: qliksense-chronos
  namespace: default
  rules:
  - apiGroups:
    - ""
    resources:
    - endpoints
    verbs:
    - get
    - update
`
	if resultYaml != expectedK8sYaml {
		t.Fatalf("expected k8s yaml: [%v] but got: [%v]\n", expectedK8sYaml, resultYaml)
	}
}
