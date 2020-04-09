package api

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	kapis_git "github.com/qlik-oss/k-apis/pkg/git"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
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

func TestCopyDirectory_withGit_withKuz(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short test mode")
	}

	tmpDir1, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.RemoveAll(tmpDir2)

	repoPath1 := path.Join(tmpDir1, "repo")
	repo1, err := kapis_git.CloneRepository(repoPath1, "https://github.com/qlik-oss/qliksense-k8s", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := CopyDirectory(repoPath1, tmpDir2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repoPath2 := tmpDir2
	repo2, err := kapis_git.OpenRepository(repoPath2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := kapis_git.Checkout(repo2, "v0.0.2", "", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repo2Manifest, err := kuz(path.Join(repoPath2, "manifests", "docker-desktop"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := kapis_git.Checkout(repo1, "v0.0.2", "", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repo1Manifest, err := kuz(path.Join(repoPath1, "manifests", "docker-desktop"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(repo2Manifest) != string(repo1Manifest) {
		t.Logf("manifest generated on the original config:\n%v", string(repo1Manifest))
		t.Logf("manifest generated on the copied config:\n%v", string(repo2Manifest))
		t.Fatal("expected manifests to be equal, but they were not")
	}
}

func kuz(directory string) ([]byte, error) {
	options := &krusty.Options{
		DoLegacyResourceSort: false,
		LoadRestrictions:     types.LoadRestrictionsNone,
		DoPrune:              false,
		PluginConfig:         konfig.DisabledPluginConfig(),
	}
	k := krusty.MakeKustomizer(filesys.MakeFsOnDisk(), options)
	resMap, err := k.Run(directory)
	if err != nil {
		return nil, err
	}
	return resMap.AsYaml()
}
