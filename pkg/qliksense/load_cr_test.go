package qliksense

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func TestLoadCrFile(t *testing.T) {
	td := setup()
	sampleCr := `
apiVersion: qlik.com/v1
kind: Qliksense
metadata:
  name: qlik-test
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

	duplicateCr := `
apiVersion: qlik.com/v1
kind: Qliksense
metadata:
  name: qlik-default
  labels:
    version: v0.0.2
spec:
  git:
    repository: https://github.com/ffoysal/qliksense-k8s
    accessToken: abababababababaab
    userName: "blblbl"`
	crFile := filepath.Join(testDir, "testcr.yaml")
	ioutil.WriteFile(crFile, []byte(sampleCr), 0644)

	dupCrFile := filepath.Join(testDir, "dupcr.yaml")
	ioutil.WriteFile(dupCrFile, []byte(duplicateCr), 0644)

	q := New(testDir)
	file, e := os.Open(crFile)
	if e != nil {
		t.Log(e)
		t.FailNow()
	}
	if err := q.LoadCr(file); err != nil {
		t.Log(err)
		t.FailNow()
	}
	qConfig := qapi.NewQConfig(testDir)
	cr, err := qConfig.GetCR("qlik-test")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if cr.GetName() != "qlik-test" {
		t.FailNow()
	}
	if qConfig.Spec.CurrentContext != "qlik-test" {
		t.FailNow()
	}
	file, e = os.Open(dupCrFile)
	if e != nil {
		t.Log(e)
		t.FailNow()
	}
	if err := q.LoadCr(file); err == nil {
		t.FailNow()
	}
	td()
}
