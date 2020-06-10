package api

import (
	"testing"
)

func TestProcessConfigArgs(t *testing.T) {
	args := []string{
		"qliksense.mongodb=mongouri://something?ffall",
		"test_under.test=value_under",
		"test-dash.dash-key=value-dash",
		"test-dot.dot-key=127.0.0.1",
		"test123.key123=value123",
		"test-equal.keyequal=newvalue=@hj",
	}
	expectedKeys := []string{"mongodb", "test", "dash-key", "dot-key", "key123", "keyequal"}
	expectedValue := []string{"mongouri://something?ffall", "value_under", "value-dash", "127.0.0.1", "value123", "newvalue=@hj"}
	exppectedSvc := []string{"qliksense", "test_under", "test-dash", "test-dot", "test123", "test-equal"}
	sv, err := ProcessConfigArgs(args, false)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	for _, v := range sv {
		if !contains(expectedKeys, v.Key) {
			t.Fail()
			t.Log("expectd key " + v.Key + " not found")
		}
		if !contains(expectedValue, v.Value) {
			t.Fail()
			t.Log("expectd Value " + v.Value + " not found")
		}
		if !contains(exppectedSvc, v.SvcName) {
			t.Fail()
			t.Log("expectd service " + v.SvcName + " not found")
		}
	}
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func TestGetSvcAndKey(t *testing.T) {
	s1 := "qliksense[tls.cert]"
	sa := getSvcAndKey(s1)
	if sa[0] != "qliksense" || sa[1] != "tls.cert" {
		t.Fail()
		t.Logf("expected service: qliksense but got %s", sa[0])
		t.Logf("expected key: tls.cert but got %s", sa[1])
	}
	s1 = "qliksense-idps.tls"
	sa = getSvcAndKey(s1)
	for _, s := range sa {
		t.Logf("|%s|", s)
	}
	if sa[0] != "qliksense-idps" || sa[1] != "tls" {
		t.Fail()
		t.Logf("expected service: qliksense-idps but got %s", sa[0])
		t.Logf("expected key: tls but got %s", sa[1])
	}

}
