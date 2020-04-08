package api

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDirectory(t *testing.T) {
	src, _ := ioutil.TempDir("", "")
	f1, _ := ioutil.TempFile(src, "")
	ioutil.TempFile(src, "")

	dest, _ := ioutil.TempDir("", "")
	CopyDirectory(src, dest)
	if _, err := os.Lstat(filepath.Join(dest, filepath.Base(f1.Name()))); err != nil {
		t.Log(err)
		t.Fail()
	}

}
