package qliksense

import (
	"testing"
)

func TestGetLatestTag(t *testing.T) {
	s, err := getLatestTag(defaultConfigRepoGitUrl, "")
	if s == "" || err != nil {
		t.Log(err)
		t.Fail()
	}
}
