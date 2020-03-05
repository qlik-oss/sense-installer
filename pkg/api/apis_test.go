package api

import (
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
  namespace: myqliksense
  storageClassName: efs
  configs:
    qliksense:
      - name: acceptEULA
        value: "yes"
  secrets:
    qliksense:
    - name: mongoDbUri
      # this is rsa encrypted value, the pub and pri keys are in setuPublicAndPrivateKey() method
      value: dnlpOC9xRnQ3NlFYWk9FZnVDWW1BbSsvR3AyaTFRZDF5bW05bmNyRFM4QzlRT3BoRVEwMmpRNlVHUmkvdzFyVG9oQmlQS1k3NDlMWDlyc3BuUDhzNHBCL0YyRDNrdlZaVDdSazVabW56Z0QrRk96N1djVExSVDQ0aHllMjB4ajFiTzF6bW1sdGdhN2dTK01hSWEveW9ZUnlDclhHZ0xxS0lycWtEMUt5WE9MS20wK0VZa3NlMllqZzVQS3ZzOWhWN1ZoditsbUU3eUoxMHlSUTdUMWlLUWhtNDhBWW5iMmR0UWwyTTNTdGs1cURHWlZLZ1U4cHJTT01pOXRwTmxmUTRSZ0tlWnVmWGEyY094MDNWcllNYW9xemM4YjNwRXVmWGttSktFMld3akFKSjhuUUJjeWc3UkZIbElmOEJXRGFwR2haVTJRaUJLaTBweWxFYklrZXlsQ05wd1I1M3ZnSncrRzFkSnd3T0RUbm5YSk1ZSXphMVFOcllWclFxemxEUGVhTERmN3VWNis3SG9BTDlSK09aR1dlRHh2MjRGV1Z0Z3pjTS9YUVJxQ3kzeSt3cjVFdmtOMkV3RXhQakNvc2V4QzZkbC9nUWVja0F1VFVXUnhEeVJ3SG90dlY5cC9TM0VNUXJJcHNwLzhxQVNVNWN3aC83dHRGRzZpRTlzZVUyVHNTdFJqSkt3a3UrK1dZN3Jjb3Bsa0tvaWc0aXNJOHFQMFUraDNFWlo2eWE5Ym8rWlpCYldjMmI0R0JRSnVsQ2ZYSTZvd3ZtR0phbnpOM2hQWlhqRUF1dDZGaEN0Lzd5WlFTaDgrTGJvZUVXTnBlWTlKTUJMbUE0a0ZsQzdObGwxOGdSd1BlNGg0Qi94aDdJWTBXUFQ2OG9QZU8rNmlLOWp5YmRTZ3M0RTQ9
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

func TestGetDecryptedCr(t *testing.T) {
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

	setuPublicAndPrivateKey(dir)
	b, _ := K8sToYaml(qcr)
	t.Log(string(b))
	_, err = qct.GetDecryptedCr(qcr)
	if err != nil {
		t.Fail()
		t.Log(err)
	}
	// b, _ = K8sToYaml(ncr)
	// t.Log(string(b))
	td()
}
func setuPublicAndPrivateKey(homeDir string) ([]byte, []byte, error) {
	privKeyBytes := []byte(`
-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`)

	publicKeyBytes := []byte(`
-----BEGIN RSA PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAwCj1Wb2q9quzvePpcq+V
nVUUhULWFWsI8lE+Pf5NZD/H+jLzvv6EG2PTidYXj8GninyRBx3eS+0tSn01THgU
wqwdpJ5szdL9HGuZZZlSgaYw1KBJXzF5TUi5KfKrQesYUgPorjt/38Ub5IEHX4h3
b2IwXX8QIOQyIDWZMCyYUm4rVARVBePrMPi2CfyrV5Picgphc9QrajNDNjPEl/AQ
Z/eH9BYImzjvdjg283irqlU1b5emTIu1G3NFwQRgIhVOATVamZaQOHKaInnM6RdO
2T6RXC54Jsv+5djDHBrG5XsL8bZS5qcdTs7PjDHizDJGzMjd66p60kdwh4ERX7v6
uaDz0o+HfIV4/PgS+PIAVCfT8b+X5XFaATCpOnvCBT/BVKWlV7HGL4d2BDG2KrDS
yZj7Bi+EFZnRql8c1lkDik9/a6FMGvSUCVmnPXsryY5AJtyzmSi4g40v46wr0Fcr
3JsY/sDbQ6OjJdh52jwX1F4i21RIj7M2hu4H+30PEMauI9M+5twQ+DSmjv13zu+Z
Vw/I+6S1yV8Ls4atOWNrfJcuTFy/zEswWpMuiVti5em/r1qeiEWYA3nmPu7d2IFP
4nBK3UojALvFU74OhCcTuETXPmw0NypaQE8V7Q78kHmbh+c9i8EoR3yx6lASJ17z
fXuR9dSdHCGH0HDT5oFsmZECAwEAAQ==
-----END RSA PUBLIC KEY-----
`)

	secretKeyPairDir := filepath.Join(homeDir, "secrets", "contexts", "contx1", "secrets")
	if err := os.MkdirAll(secretKeyPairDir, 0777); err != nil {
		err = fmt.Errorf("Not able to create directories")
		log.Fatal(err)
	}
	os.Setenv("QLIKSENSE_KEY_LOCATION", secretKeyPairDir)

	privKeyFile := filepath.Join(secretKeyPairDir, "qliksensePriv")
	// construct and write priv key file into secretsDir location
	err := ioutil.WriteFile(privKeyFile, privKeyBytes, 0777)
	if err != nil {
		log.Printf("Error while creating file: %v", err)
		return nil, nil, err
	}
	pubKeyFile := filepath.Join(secretKeyPairDir, "qliksensePub")
	// construct and write pub key file into secretsDir location
	err = ioutil.WriteFile(pubKeyFile, publicKeyBytes, 0777)
	if err != nil {
		log.Printf("Error while creating file: %v", err)
		return nil, nil, err
	}
	return publicKeyBytes, privKeyBytes, nil
}
