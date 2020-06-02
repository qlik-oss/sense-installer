package qliksense

import (
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestGetLatestTag(t *testing.T) {
	s, err := getLatestTag(defaultConfigRepoGitUrl, "")
	if s == "" || err != nil {
		t.Log(err)
		t.Fail()
	}
	sv, err := semver.NewVersion(s)
	if err != nil {
		t.Log(err)
		t.Log(sv)
	}
	baseV, _ := semver.NewVersion("v0.0.8")
	if !sv.GreaterThan(baseV) {
		t.Log("Expected greater than v0.0.8, but got:  " + s)
		t.Fail()
	}
}
