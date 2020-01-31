package qliksense

import (
	"os"
	"testing"
)

func Test_downloadFromGitRepo(t *testing.T) {
	downloadPath, err := downloadFromGitRepoToTmpDir("https://github.com/hashicorp/go-getter", "v1.4.1")
	if err != nil {
		t.Fatal(err)
	}
	if downloadPath == "" {
		t.Fatal(err)
	}
	_ = os.RemoveAll(downloadPath)
}