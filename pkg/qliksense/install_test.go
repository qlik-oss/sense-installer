package qliksense

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
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
	if err := q.LoadCr(file); err != nil {
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
