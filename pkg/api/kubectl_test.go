package api

import (
	"testing"
)

func TestGetKubectlNamespace(t *testing.T) {
	t.Skip()
	ns := GetKubectlNamespace()
	SetKubectlNamespace("tada")
	got := GetKubectlNamespace()
	if got != "tada" {
		t.Log(got)
		t.Fail()
	}
	SetKubectlNamespace(ns)
}
