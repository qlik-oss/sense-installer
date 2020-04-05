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
	}
	expectedKeys := []string{"mongodb", "test", "dash-key", "dot-key", "key123"}
	expectedValue := []string{"mongouri://something?ffall", "value_under", "value-dash", "127.0.0.1", "value123"}
	exppectedSvc := []string{"qliksense", "test_under", "test-dash", "test-dot", "test123"}
	sv, err := ProcessConfigArgs(args)
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
