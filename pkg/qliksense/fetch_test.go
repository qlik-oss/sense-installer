package qliksense

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func TestFetchAndUpdateCR(t *testing.T) {
	ver := "v1.58.17"
	tempHome, _ := ioutil.TempDir("", "")

	q := &Qliksense{
		QliksenseHome: tempHome,
	}
	q.SetUpQliksenseContext("test1")
	qConfig := qapi.NewQConfig(tempHome)
	if err := fetchAndUpdateCR(qConfig, ver); err != nil {
		t.Log(err)
		t.FailNow()
	}

	actualCrFile := filepath.Join(tempHome, "contexts", "test1", "test1.yaml")
	cr := &qapi.QliksenseCR{}
	if err := qapi.ReadFromFile(cr, actualCrFile); err != nil {
		t.Log(err)
		t.FailNow()
	}

	if cr.Spec.ManifestsRoot != fmt.Sprintf("contexts/test1/qlik-k8s/%s", ver) {
		t.Logf("actual path: %s, expected path: contexts/test1/qlik-k8s/%s", cr.Spec.ManifestsRoot, ver)
		t.FailNow()
	}
	//testing latest tag is fetched
	cr.AddLabelToCr("version", "")
	qConfig.WriteCR(cr)
	err := fetchAndUpdateCR(qConfig, "")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	cr = &qapi.QliksenseCR{}
	qapi.ReadFromFile(cr, actualCrFile)
	v := cr.GetLabelFromCr("version")
	if v == "" || v == ver {
		t.Log("should get latest but got version: " + v)
		t.Fail()
	}
}
