package api

import (
	"fmt"
	"strings"
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

func TestKubectlDirectOps(t *testing.T) {
	t.Skip()
	SetKubectlNamespace("test")
	ns := GetKubectlNamespace()
	opr := fmt.Sprintf("version")
	opr1 := strings.Fields(opr)
	_, err := KubectlDirectOps(opr1, ns)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}
