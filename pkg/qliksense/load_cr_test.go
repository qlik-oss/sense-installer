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
	setup()
	sampleCr1 := `
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
    - name: mongodbUri
      value: mongodb://qlik-default-mongodb:27017/qliksense?ssl=false
  profile: docker-desktop
  rotateKeys: "yes"`
	sampleCr2 := `
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
    - name: mongodbUri
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
	crFile1 := filepath.Join(testDir, "testcr1.yaml")
	ioutil.WriteFile(crFile1, []byte(sampleCr1), 0644)
	crFile2 := filepath.Join(testDir, "testcr2.yaml")
	ioutil.WriteFile(crFile2, []byte(sampleCr2), 0644)

	dupCrFile := filepath.Join(testDir, "dupcr.yaml")
	ioutil.WriteFile(dupCrFile, []byte(duplicateCr), 0644)

	q := New(testDir)
	file1, e := os.Open(crFile1)
	if e != nil {
		t.Log(e)
		t.FailNow()
	}
	if err := q.LoadCr(file1, false); err != nil {
		t.Log(err)
		t.FailNow()
	}
	file2, e := os.Open(crFile2)
	if e != nil {
		t.Log(e)
		t.FailNow()
	}
	if err := q.LoadCr(file2, false); err != nil {
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

	cr, err = qConfig.GetCR("qlik-test3")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if cr.GetName() != "qlik-test3" {
		t.FailNow()
	}

	if qConfig.Spec.CurrentContext != "qlik-test3" {
		t.FailNow()
	}
	file, e := os.Open(dupCrFile)
	if e != nil {
		t.Log(e)
		t.FailNow()
	}
	if err := q.LoadCr(file, false); err == nil {
		t.FailNow()
	}
	td()
}
