package api

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
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
	ver := "master"
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
	_, err = kapis_git.CloneRepository(repoPath1, "https://github.com/qlik-oss/qliksense-k8s", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	repo1, err := kapis_git.OpenRepository(repoPath1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := kapis_git.Checkout(repo1, ver, "", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := CopyDirectory(repoPath1, tmpDir2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d, _ := directoryContentsEqual(repoPath1, tmpDir2); !d {
		t.Log("Directory was not copied properly")
		t.Fail()
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

func directoryContentsEqual(dir1 string, dir2 string) (bool, error) {
	if map1, err := getDirMap(dir1); err != nil {
		return false, err
	} else if map2, err := getDirMap(dir2); err != nil {
		return false, err
	} else if !reflect.DeepEqual(map1, map2) {
		return false, nil
	}
	return true, nil
}

func getDirMap(dir string) (map[string][]byte, error) {
	dirMap := make(map[string][]byte)
	if err := filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fpath != dir && !info.IsDir() {
			if fileContent, err := ioutil.ReadFile(fpath); err != nil {
				return err
			} else {
				dirMap[path.Base(fpath)] = fileContent
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return dirMap, nil
}
