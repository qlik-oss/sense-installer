package qliksense

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/qlik-oss/sense-installer/pkg/api"
	_ "gopkg.in/yaml.v2"
)

func TestUnsetAll(t *testing.T) {
	qHome, _ := ioutil.TempDir("", "")
	testPepareDir(qHome)
	defer os.RemoveAll(qHome)
	//fmt.Print(qHome)
	args := []string{"rotateKeys", "qliksense", "qliksense2.acceptEula3", "serviceA.acceptEula", "opsRunner.watchBranch"}
	//args := []string{"opsRunner"}
	//args := []string{"opsRunner.watchBranch"}
	if err := unsetAll(qHome, args); err != nil {
		t.Log("error during unset", err)
		t.FailNow()
	}
	qc := api.NewQConfig(qHome)
	qcr, err := qc.GetCurrentCR()
	if err != nil {
		t.Log("error while getting current cr", err)
		t.FailNow()
	}
	if qcr.Spec.RotateKeys != "" {
		t.Log("Expected empty rotateKeys but got: " + qcr.Spec.RotateKeys)
		t.Fail()
	}

	if qcr.Spec.Configs["qliksense"] != nil {
		t.Log("qliksense in configs not deleted")
		t.Fail()
	}
	if len(qcr.Spec.Configs["qliksense2"]) != 1 {
		t.Log("qliksense2.acceptEula3 not deleted")
		t.Fail()
	}
	if qcr.Spec.Configs["serviceA"] != nil {
		t.Log("serviceA not deleted")
		t.Fail()
	}
	if qcr.Spec.OpsRunner == nil {
		t.Log("opsRunner not deleted")
		t.Fail()
	}
	if qcr.Spec.OpsRunner.WatchBranch != "" {
		t.Log("opsRunner.watchBranch not deleted")
		t.Fail()
	}
}

func testPepareDir(qHome string) {

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
	configFile := filepath.Join(qHome, "config.yaml")
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
  opsRunner:
    enabled: "yes"
    watchBranch: something
  configs:
    qliksense:
    - name: acceptEula
      value: some
    qliksense2:
    - name: acceptEula2
      value: some
    - name: acceptEula3
      value: some
    serviceA:
    - name: acceptEula
      value: some
`
	qlikDefaultContext := "qlik-default"
	// create contexts/qlik-default/ under tests/
	contexts := "contexts"
	contextsDir := filepath.Join(qHome, contexts, qlikDefaultContext)
	if err := os.MkdirAll(contextsDir, 0777); err != nil {
		err = fmt.Errorf("Not able to create directories")
	}

	contextFile := filepath.Join(contextsDir, qlikDefaultContext+".yaml")
	ioutil.WriteFile(contextFile, []byte(contextYaml), 0777)
}
