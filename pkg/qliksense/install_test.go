package qliksense

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
)

func TestCreateK8sResoruceBeforePatch(t *testing.T) {
	td := setup()
	sampleCr := `
apiVersion: qlik.com/v1
kind: Qliksense
metadata:
  name: qlik-test3
  labels:
    version: v0.0.2
spec:
  git:
    repository: https://github.com/ffoysal/qliksense-k8s
    accessToken: abababababababaab
    userName: "blblbl"
  gitOps:
    enabled: "no"
    schedule: "*/1 * * * *"
    watchBranch: pr-branch-db1d26d6
    image: qlik-docker-oss.bintray.io/qliksense-repo-watcher
  configs:
    qliksense:
    - name: acceptEULA
      value: "yes"
  secrets:
    qliksense:
    - name: mongoDbUri
      value: mongodb://qlik-default-mongodb:27017/qliksense?ssl=false
  profile: docker-desktop
  rotateKeys: "yes"`

	crFile := filepath.Join(testDir, "install_test.yaml")
	ioutil.WriteFile(crFile, []byte(sampleCr), 0644)
	q := New(testDir)
	file, e := os.Open(crFile)
	if e != nil {
		t.Log(e)
		t.FailNow()
	}
	if err := q.LoadCr(file, false); err != nil {
		t.Log(err)
		t.FailNow()
	}
	qConfig := qapi.NewQConfig(testDir)
	cr, err := qConfig.GetCR("qlik-test3")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if err = q.createK8sResoruceBeforePatch(cr); err != nil {
		t.Log(err)
		t.FailNow()
	}
	td()
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

func Test_getProcessedOperatorControllerString(t *testing.T) {
	tmpQlikSenseHome, err := ioutil.TempDir("", "tmp-qlik-sense-home-")
	if err != nil {
		t.Fatalf("unexpected error creating tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpQlikSenseHome)

	registry := "registryFoo"
	setupQliksenseTestDefaultContext(t, tmpQlikSenseHome, fmt.Sprintf(`
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

	q := &Qliksense{
		QliksenseHome: tmpQlikSenseHome,
	}

	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		t.Fatalf("unexpected error getting current CR: %v", err)
	}

	originalOperatorString, err := q.GetOperatorControllerString()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	processedOperatorString, err := q.getProcessedOperatorControllerString(qcr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	controllerImageChecks := map[string]func(t *testing.T, controllerImage string){
		originalOperatorString: func(t *testing.T, controllerImage string) {
			expectedControllerImagePrefix := fmt.Sprintf("%v/%v:", qliksenseOperatorImageRepo, qliksenseOperatorImageName)
			if !strings.HasPrefix(controllerImage, expectedControllerImagePrefix) {
				t.Fatalf("expected controller image: %v to have prefix: %v", controllerImage, expectedControllerImagePrefix)
			}
		},
		processedOperatorString: func(t *testing.T, controllerImage string) {
			expectedControllerImagePrefix := fmt.Sprintf("%v/%v:", registry, qliksenseOperatorImageName)
			if !strings.HasPrefix(controllerImage, expectedControllerImagePrefix) {
				t.Fatalf("expected controller image: %v to have prefix: %v", controllerImage, expectedControllerImagePrefix)
			}
		},
	}

	resourceFactory := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), nil)
	for operatorString, controllerImageCheck := range controllerImageChecks {
		resMap, err := resourceFactory.NewResMapFromBytes([]byte(operatorString))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		res, err := resMap.GetById(resid.NewResId(resid.Gvk{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		}, "qliksense-operator"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		controllerImage, err := res.GetString("spec.template.spec.containers[0].image")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		controllerImageCheck(t, controllerImage)
	}
}
