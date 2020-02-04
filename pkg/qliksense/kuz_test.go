package qliksense

import (
	kapis_git "github.com/qlik-oss/k-apis/git"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func Test_executeKustomizeBuild(t *testing.T) {
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

	result, err := executeKustomizeBuild(tmpDir)
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

func Test_executeKustomizeBuild_onQlikConfig_DISABLED(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := path.Join(tmpDir, "config")
	if repo, err := kapis_git.CloneRepository(configPath, defaultGitUrl, nil); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	} else if err := kapis_git.Checkout(repo, "v1.21.23-edge", "", nil); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	if _, err := executeKustomizeBuild(path.Join(configPath, "manifests", "base")); err != nil {
		t.Fatalf("unexpected kustomize error: %v\n", err)
	}
}