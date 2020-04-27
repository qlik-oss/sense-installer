package qliksense

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"

	"gopkg.in/yaml.v2"

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

	kustomizationYamlFilePath := filepath.Join(tmpDir, "kustomization.yaml")
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

func Test_executeKustomizeBuild_onQlikConfig_withConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping parallel kustomize test in short mode")
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config")
	if repo, err := kapis_git.CloneRepository(configPath, defaultConfigRepoGitUrl, nil); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	} else if err := kapis_git.Checkout(repo, "v0.0.8", "v0.0.8", nil); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	configsDir := filepath.Join(configPath, "manifests", "base", "manifests")

	shouldSkipDir := func(dirName string) bool {
		skipMap := map[string]bool{
			"sense-client": true,
		}
		_, ok := skipMap[dirName]
		return ok
	}

	insertConcatenationKuzYaml := func(dir string) error {
		concatenatedMap := map[string]interface{}{}
		concatenatedMap["apiVersion"] = "kustomize.config.k8s.io/v1beta1"
		concatenatedMap["kind"] = "Kustomization"
		resources := make([]string, 0)
		infos, err := ioutil.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, info := range infos {
			if !info.IsDir() {
				resources = append(resources, info.Name())
			}
		}
		concatenatedMap["resources"] = resources

		if concatenatedMapBytes, err := yaml.Marshal(&concatenatedMap); err != nil {
			return err
		} else if err := ioutil.WriteFile(filepath.Join(dir, "kustomization.yaml"), concatenatedMapBytes, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	//sequential test:
	sequentialManifestsDir := filepath.Join(tmpDir, "sequential")
	if err := os.RemoveAll(sequentialManifestsDir); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	} else if err := os.MkdirAll(sequentialManifestsDir, os.ModePerm); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	} else if err := os.RemoveAll(filepath.Join(os.TempDir(), "dotHelm")); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	} else if err := os.RemoveAll(filepath.Join(os.TempDir(), ".chartcache")); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	infos, err := ioutil.ReadDir(configsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	fmt.Print("running a sequential execution test...\n")
	t1 := time.Now()
	for _, info := range infos {
		if shouldSkipDir(info.Name()) {
			continue
		}
		if info.IsDir() {
			if yamlResources, err := ExecuteKustomizeBuild(filepath.Join(configsDir, info.Name())); err != nil {
				t.Fatalf("unexpected error kustomizing: %v, error: %v\n", info.Name(), err)
			} else if err := ioutil.WriteFile(filepath.Join(sequentialManifestsDir, fmt.Sprintf("%v.yaml", info.Name())), yamlResources, os.ModePerm); err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}
		}
	}
	t2 := time.Now()
	fmt.Printf("sequential execution test took: %vs\n", t2.Sub(t1).Seconds())
	if err := insertConcatenationKuzYaml(sequentialManifestsDir); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	concatenatedSequentialManifests, err := ExecuteKustomizeBuild(sequentialManifestsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	//concurrent tests:
	numConcurrentTests := 5
	concatenatedConcurrentManifestsList := make([][]byte, 0)
	for i := 0; i < numConcurrentTests; i++ {
		concurrentManifestsDir := filepath.Join(tmpDir, fmt.Sprintf("concurrent-%v", i))
		if err := os.RemoveAll(concurrentManifestsDir); err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		} else if err := os.MkdirAll(concurrentManifestsDir, os.ModePerm); err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		} else if err := os.RemoveAll(filepath.Join(os.TempDir(), "dotHelm")); err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		} else if err := os.RemoveAll(filepath.Join(os.TempDir(), ".chartcache")); err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		}

		var wg sync.WaitGroup
		var concurrentErrorCounter int32
		osStderrBackup := os.Stderr
		tmpStdErrFile, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		}
		os.Stderr = tmpStdErrFile
		fmt.Print("running a concurrent execution test...\n")
		t1 = time.Now()
		for _, fi := range infos {
			wg.Add(1)
			go func(info os.FileInfo) {
				defer wg.Done()
				if shouldSkipDir(info.Name()) {
					return
				}
				if info.IsDir() {
					if yamlResources, err := ExecuteKustomizeBuild(filepath.Join(configsDir, info.Name())); err != nil {
						fmt.Printf("unexpected error: %v\n", err)
						atomic.AddInt32(&concurrentErrorCounter, 1)
					} else if err := ioutil.WriteFile(filepath.Join(concurrentManifestsDir, fmt.Sprintf("%v.yaml", info.Name())), yamlResources, os.ModePerm); err != nil {
						fmt.Printf("unexpected error: %v\n", err)
						atomic.AddInt32(&concurrentErrorCounter, 1)
					}
				}
			}(fi)
		}
		wg.Wait()
		t2 = time.Now()
		os.Stderr = osStderrBackup
		os.Remove(tmpStdErrFile.Name())
		if concurrentErrorCounter > 0 {
			t.Fatalf("there were %v errors during the concurrent execution", concurrentErrorCounter)
		}
		fmt.Printf("concurrent execution test took: %vs\n", t2.Sub(t1).Seconds())
		if err := insertConcatenationKuzYaml(concurrentManifestsDir); err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		}
		concatenatedConcurrentManifests, err := ExecuteKustomizeBuild(concurrentManifestsDir)
		if err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		}
		concatenatedConcurrentManifestsList = append(concatenatedConcurrentManifestsList, concatenatedConcurrentManifests)
	}

	getResMapBytesAdjustedForCaCertsJobName := func(resBytes []byte) ([]byte, error) {
		resMapFactory := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), nil)
		if resMap, err := resMapFactory.NewResMapFromBytes(resBytes); err != nil {
			return nil, err
		} else if resources, err := resMap.Select(types.Selector{
			Gvk: resid.Gvk{
				Group:   "batch",
				Version: "v1",
				Kind:    "Job",
			},
		}); err != nil {
			return nil, err
		} else if re, err := regexp.Compile(`^.+-ca-certificates-[a-z]{5}$`); err != nil {
			return nil, err
		} else {
			for _, res := range resources {
				res.SetName(re.ReplaceAllString(res.GetName(), "qliksense-ca-certificates"))
			}
			return resMap.AsYaml()
		}
	}

	sequentialFinalManifest, err := getResMapBytesAdjustedForCaCertsJobName(concatenatedSequentialManifests)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	concurrentFinalManifestsList := make([][]byte, 0)
	for _, concatenatedConcurrentManifest := range concatenatedConcurrentManifestsList {
		if concurrentFinalManifest, err := getResMapBytesAdjustedForCaCertsJobName(concatenatedConcurrentManifest); err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		} else {
			concurrentFinalManifestsList = append(concurrentFinalManifestsList, concurrentFinalManifest)
		}
	}

	for _, concurrentFinalManifest := range concurrentFinalManifestsList {
		if !bytes.Equal(concurrentFinalManifest, sequentialFinalManifest) {
			t.Fatalf("expected the concatenated concurrent manifest to equal the concatenated sequential manifest, but they didn't..."+
				"\nconcurrent:\n%v\nsequential:\n%v", string(concurrentFinalManifest), string(sequentialFinalManifest))
		}
	}
}

func Test_executeKustomizeBuild_onQlikConfig_regenerateKeys(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config")
	if repo, err := kapis_git.CloneRepository(configPath, defaultConfigRepoGitUrl, nil); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	} else if err := kapis_git.Checkout(repo, "e38df644e759abf0b5941c1511d1a2cd5e3c42fa", "", nil); err != nil {
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

	yamlResources, err := ExecuteKustomizeBuild(filepath.Join(configPath, "manifests", "base", "resources", "users"))
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
		if resource["kind"].(string) == "Secret" && strings.Contains(resource["metadata"].(map[interface {}]interface {})["name"].(string), "users-secrets-") {
			keyIdBase64 = resource["data"].(map[interface {}]interface {})["tokenAuthPrivateKeyId"].(string)
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
	} else if err := ioutil.WriteFile(filepath.Join(keyDir, ejsonPublicKey), []byte(ejsonPrivateKey), os.ModePerm); err != nil {
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

	kustomizationYamlFilePath := filepath.Join(tmpDir, "kustomization.yaml")
	testResFileYamlFilePath := filepath.Join(tmpDir, "test-file.yaml")
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
