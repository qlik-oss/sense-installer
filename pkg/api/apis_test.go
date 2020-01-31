package api

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const tempPermissionCode os.FileMode = 0777

func setup() (func(), string) {
	dir, _ := ioutil.TempDir("", "testing_path")
	config :=
		`
apiVersion: config.qlik.com/v1
kind: QliksenseConfig
metadata:
  name: whatever
spec:
  contexts:
  - name: contx1
    crLocation: /Users/mqb/.qliksense/contexts/contx1
  - name: cotx2
    crLocation: /root/.qliksense/contexts/cotx2.yaml
  currentContext: contx1
`
	configFile := filepath.Join(dir, "config.yaml")
	ioutil.WriteFile(configFile, []byte(config), tempPermissionCode)
	tearDown := func() {
		os.RemoveAll(dir)
	}
	return tearDown, dir
}

func createCRFile(homeDir string) {
	cr :=
		`
apiVersion: qlik.com/v1
kind: QlikSense
metadata:
  name: contx1
  labels:
    version: v1.0.0
spec:
  profile: docker-desktop
  manifestsRoot: /Users/mqb/.qliksense/contexts/contx1/qlik-k8s/v0.0.1/manifests
  namespace: myqliksense
  storageClassName: efs
  configs:
    qliksense:
      - name: acceptEULA
        value: "yes"
`
	ctx1Dir := filepath.Join(homeDir, "contexts", "contx1")
	crFile := filepath.Join(ctx1Dir, "contx1.yaml")
	os.MkdirAll(ctx1Dir, tempPermissionCode)
	ioutil.WriteFile(crFile, []byte(cr), tempPermissionCode)

}
func TestGetCR(t *testing.T) {
	td, dir := setup()
	qc := NewQConfig(dir)
	if qc.Spec.CurrentContext != "contx1" {
		t.Fail()
	}
	// create CR
	createCRFile(dir)

	crFile := filepath.Join(dir, "contexts", "contx1", "contx1.yaml")
	qct, e := qc.SetCrLocation("contx1", crFile)
	if e != nil {
		t.Fail()
		t.Log(e)
	}
	qcr, err := qct.GetCurrentCR()
	if err != nil {
		t.Fail()
		t.Log(err)
	}
	if qcr.Spec.Profile != "docker-desktop" {
		t.Fail()
	}
	td()
}
