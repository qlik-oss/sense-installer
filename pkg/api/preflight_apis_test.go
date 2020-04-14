package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func Test_Initalize(t *testing.T) {
	testCases := []struct {
		name     string
		validate func(t *testing.T, tempDir string)
	}{
		{
			name: "without account for imageRegistry",
			validate: func(t *testing.T, tempDir string) {
				preflightConfig := NewPreflightConfig(tempDir)
				imageName, err := preflightConfig.GetImageName("test", false)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if imageName != "testimage" {
					t.Fatalf("expected image name: testimage, got: %v", imageName)
				}
			},
		},
		{
			name: "with account for configured imageRegistry",
			validate: func(t *testing.T, tempDir string) {
				registry := "registryFoo"
				setupQliksenseTestDefaultContext(t, tempDir, fmt.Sprintf(`
apiVersion: qlik.com/v1
kind: Qliksense
metadata:
  name: qlik-default
spec:
  configs:
    qliksense:
    - name: imageRegistry
      value: %v
`, registry))
				preflightConfig := NewPreflightConfig(tempDir)
				imageName, err := preflightConfig.GetImageName("test", true)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				expectedImageName := fmt.Sprintf("%v/testimage", registry)
				if imageName != expectedImageName {
					t.Fatalf("expected image name: %v, got: %v", expectedImageName, imageName)
				}
			},
		},
		{
			name: "with account for un-configured imageRegistry",
			validate: func(t *testing.T, tempDir string) {
				setupQliksenseTestDefaultContext(t, tempDir, `
apiVersion: qlik.com/v1
kind: Qliksense
metadata:
  name: qlik-default
spec:
  configs:
    qliksense:
    - name: something
      value: other
`)
				preflightConfig := NewPreflightConfig(tempDir)
				imageName, err := preflightConfig.GetImageName("test", true)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				expectedImageName := "testimage"
				if imageName != expectedImageName {
					t.Fatalf("expected image name: %v, got: %v", expectedImageName, imageName)
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tempDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)
			setupPreflightConfig(t, tempDir)
			testCase.validate(t, tempDir)
		})
	}
}

func setupPreflightConfig(t *testing.T, tempDir string) {
	pf := NewPreflightConfig(tempDir)
	if err := pf.Initialize(); err != nil {
		t.Fatal(err)
	}
	p := &PreflightConfig{
		QliksenseHomePath: tempDir,
	}
	if err := ReadFromFile(p, pf.GetConfigFilePath()); err != nil {
		t.Fatal(err)
	}
	if p.GetMinK8sVersion() != "1.15" {
		t.Fatalf("expected k8 version: 1.15, but got " + p.GetMinK8sVersion())
	}
	p.AddImage("test", "testimage")
	if err := p.Write(); err != nil {
		t.Fatal(err)
	}
}

func setupQliksenseTestDefaultContext(t *testing.T, tmpQlikSenseHome, CR string) {
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
		t.Fatalf("unexpected error: %v", err)
	}

	defaultContextDir := path.Join(tmpQlikSenseHome, "contexts", "qlik-default")
	if err := os.MkdirAll(defaultContextDir, os.ModePerm); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := ioutil.WriteFile(path.Join(defaultContextDir, "qlik-default.yaml"), []byte(CR), os.ModePerm); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
