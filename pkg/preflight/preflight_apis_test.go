package preflight

import (
	"io/ioutil"
	"testing"

	api "github.com/qlik-oss/sense-installer/pkg/api"
)

func Test_Initalize(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	pf := NewPreflightConfig(tempDir)
	if err := pf.Initialize(); err != nil {
		t.Log()
		t.FailNow()
	}
	p := &PreflightConfig{
		QliksenseHomePath: tempDir,
	}
	if err := api.ReadFromFile(p, pf.GetConfigFilePath()); err != nil {
		t.Log(err)
		t.FailNow()
	}
	if p.GetMinK8sVersion() != "1.15" {
		t.Log("expected k8 version: 1.15, but got " + p.GetMinK8sVersion())
		t.Fail()
	}
	p.AddImage("test", "testimage")
	if err := p.Write(); err != nil {
		t.Log(err)
		t.Fail()
	}
	p2 := NewPreflightConfig(tempDir)
	if p2.GetImageName("test") != "testimage" {
		t.Log("expected image name: testimage, got: " + p2.GetImageName("test"))
	}
}
