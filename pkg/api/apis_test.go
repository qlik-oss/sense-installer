package api

import (
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
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

	crFile := filepath.Join("contexts", "contx1", "contx1.yaml")
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

func TestGetDecryptedCr(t *testing.T) {
	td, dir := setup()
	qc := NewQConfig(dir)
	if qc.Spec.CurrentContext != "contx1" {
		t.Fail()
	}
	// create CR
	createCRFile(dir)

	crFile := filepath.Join("contexts", "contx1", "contx1.yaml")
	qct, e := qc.SetCrLocation("contx1", crFile)
	if e != nil {
		t.Fail()
		t.Log(e)
	}

	qcr, err := qct.GetCurrentCR()

	key, _ := setupGenerateKey(dir)
	ecn, _ := EncryptData([]byte("mongodb://qlik-default-mongodb:27017/qliksense?ssl=false"), key)
	b := b64.StdEncoding.EncodeToString(ecn)
	qcr.Spec.AddToSecrets("qliksense", "mongoDbUri", b, "")

	if err != nil {
		t.Fail()
		t.Log(err)
	}

	newCr, err := qct.GetDecryptedCr(qcr)
	if err != nil {
		t.Fail()
		t.Log(err)
	}

	decryptedValue := newCr.Spec.GetFromSecrets("qliksense", "mongoDbUri")
	orignalValue := qcr.Spec.GetFromSecrets("qliksense", "mongoDbUri")
	if decryptedValue != "mongodb://qlik-default-mongodb:27017/qliksense?ssl=false" {
		t.Fail()
		b, _ := K8sToYaml(newCr)
		t.Log(b)
	}
	if decryptedValue == orignalValue {
		t.Fail()
	}
	td()
}
func setupGenerateKey(homeDir string) (string, error) {
	secretKeyPairDir := filepath.Join(homeDir, "secrets", "contexts", "contx1", "secrets")
	if err := os.MkdirAll(secretKeyPairDir, 0777); err != nil {
		err = fmt.Errorf("Not able to create directories")
		log.Fatal(err)
	}
	os.Setenv("QLIKSENSE_KEY_LOCATION", secretKeyPairDir)

	key, _ := LoadSecretKey(secretKeyPairDir)

	if key == "" {
		return GenerateAndStoreSecretKey(secretKeyPairDir)
	}
	return key, nil
}
