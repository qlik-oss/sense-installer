package qliksense

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

func Test_locateDockerRegistryBinary(t *testing.T) {
	binary, err := locateDockerRegistryBinary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmd := exec.Command(binary, "--version")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("output: %v\n", string(out))
}

func Test_Pull_Push_noAuth(t *testing.T) {
	registryURI := "localhost:5555"
	registry, err := setupRegistryV2At(registryURI, false)
	if registry != nil {
		defer registry.Close()
	}
	if err != nil {
		t.Fatalf("unexpected error setting up local registry: %v", err)
	}

	tmpQlikSenseHome, err := ioutil.TempDir("", "tmp-qlik-sense-home-")
	if err != nil {
		t.Fatalf("unexpected error creating tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpQlikSenseHome)

	if err := setupQlikSenseHome(t, tmpQlikSenseHome, registryURI); err != nil {
		t.Fatalf("unexpected error setting up qliksense home: %v", err)
	}
	q := &Qliksense{QliksenseHome: tmpQlikSenseHome}
	var versionOut VersionOutput

	if err := q.PullImagesForCurrentCR(); err != nil {
		t.Fatalf("unexpected pull error: %v", err)
	} else if versionOutBytes, err := ioutil.ReadFile(path.Join(tmpQlikSenseHome, "images", "foo")); err != nil {
		t.Fatalf("unexpected error reading version file: %v", err)
	} else if err = yaml.Unmarshal(versionOutBytes, &versionOut); err != nil {
		t.Fatalf("unexpected error unmarshalling version file: %v", err)
	} else if len(versionOut.Images) != 1 || versionOut.Images[0] != "alpine:latest" {
		t.Fatal("did not find alpine:latest in the version file")
	} else if infos, err := ioutil.ReadDir(path.Join(tmpQlikSenseHome, "images", "index", "alpine", "latest")); err != nil || len(infos) == 0 {
		t.Fatal("expected images/index/alpine/latest directory to be non-empty")
	} else if blobInfos, err := ioutil.ReadDir(path.Join(tmpQlikSenseHome, "images", "blobs", "sha256")); err != nil || len(blobInfos) == 0 {
		t.Fatal("expected images/blobs/sha256 directory to be non-empty")
	}

	if err := q.PushImagesForCurrentCR(registryURI); err != nil {
		t.Fatalf("unexpected push error: %v", err)
	} else if tmpImagesDir, err := ioutil.TempDir("", "tmp-images-"); err != nil {
		t.Fatalf("unexpected error creating tmp dir: %v", err)
	} else if err := pullImage(fmt.Sprintf("%s/alpine:latest", registryURI), tmpImagesDir, false); err != nil {
		t.Fatalf("unexpected error pulling alpine:latest from the local registry: %v", err)
	} else if infos, err := ioutil.ReadDir(path.Join(tmpImagesDir, "index", "alpine", "latest")); err != nil || len(infos) == 0 {
		t.Fatal("expected index/alpine/latest directory to be non-empty")
	} else if blobInfos, err := ioutil.ReadDir(path.Join(tmpImagesDir, "blobs", "sha256")); err != nil || len(blobInfos) == 0 {
		t.Fatal("expected blobs/sha256 directory to be non-empty")
	}
}

func setupQlikSenseHome(t *testing.T, tmpQlikSenseHome string, registryURI string) error {
	if err := ioutil.WriteFile(path.Join(tmpQlikSenseHome, "config.yaml"), []byte(fmt.Sprintf(`
apiVersion: config.qlik.com/v1
kind: QliksenseConfig
metadata:
  name: QliksenseConfigMetadata
spec:
  contexts:
  - name: qlik-default
    crFile: %s/contexts/qlik-default/qlik-default.yaml
  currentContext: qlik-default
`, tmpQlikSenseHome)), os.ModePerm); err != nil {
		return err
	}

	defaultContextDir := path.Join(tmpQlikSenseHome, "contexts", "qlik-default")
	if err := os.MkdirAll(defaultContextDir, os.ModePerm); err != nil {
		return err
	}

	version := "foo"
	manifestsRootDir := fmt.Sprintf("%s/repo/%s", defaultContextDir, version)
	if err := ioutil.WriteFile(path.Join(defaultContextDir, "qlik-default.yaml"), []byte(fmt.Sprintf(`
apiVersion: qlik.com/v1
kind: Qliksense
metadata:
  name: qlik-default
  labels:
    version: %s
spec:
  profile: docker-desktop
  configs:
    qliksense:
    - name: imageRegistry
      value: %s
  manifestsRoot: %s
  rotateKeys: "yes"
  releaseName: qlik-default
`, version, registryURI, manifestsRootDir)), os.ModePerm); err != nil {
		return err
	}

	profileDir := path.Join(manifestsRootDir, "manifests", "docker-desktop")
	if err := os.MkdirAll(profileDir, os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(profileDir, "kustomization.yaml"), []byte(`
resources:
- deployment.yaml
`), os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(profileDir, "deployment.yaml"), []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  template:
    spec:
      containers:
        - name: the-container
          image: alpine:latest
`), os.ModePerm); err != nil {
		return err
	}

	transformersDir := path.Join(manifestsRootDir, "transformers")
	if err := os.MkdirAll(transformersDir, os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(transformersDir, "qseokversion.yaml"), []byte(`
apiVersion: qlik.com/v1
kind: SelectivePatch
metadata:
  name: qseokversion
enabled: true
patches:
- target:
    kind: HelmChart
    labelSelector: name!=qliksense-init
  patch: |-
    chartName: qliksense
    chartVersion: 1.21.23
`), os.ModePerm); err != nil {
		return err
	}
	return nil
}

type testRegistryV2 struct {
	cmd      *exec.Cmd
	url      string
	dir      string
	username string
	password string
	email    string
}

func locateDockerRegistryBinary() (string, error) {
	if exePath, err := exec.LookPath("docker-registry"); err != nil {
		if cwd, err := os.Getwd(); err != nil {
			return "", err
		} else {
			return path.Join(cwd, "docker-registry"), nil
		}
	} else {
		return exePath, nil
	}
}

func setupRegistryV2At(url string, auth bool) (*testRegistryV2, error) {
	reg, err := newTestRegistryV2At(url, auth)
	if err != nil {
		return nil, err
	}

	// Wait for registry to be ready to serve requests.
	for i := 0; i != 50; i++ {
		if err = reg.Ping(); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err != nil {
		return reg, errors.New("Timeout waiting for test registry to become available")
	}
	return reg, nil
}

func newTestRegistryV2At(url string, auth bool) (*testRegistryV2, error) {
	tmp, err := ioutil.TempDir("", "registry-test-")
	if err != nil {
		return nil, err
	}
	template := `version: 0.1
loglevel: debug
storage:
    filesystem:
        rootdirectory: %s
    delete:
        enabled: true
http:
    addr: %s
%s`
	var (
		htpasswd string
		username string
		password string
		email    string
	)
	if auth {
		htpasswdPath := filepath.Join(tmp, "htpasswd")
		userpasswd := "testuser:$2y$05$sBsSqk0OpSD1uTZkHXc4FeJ0Z70wLQdAX/82UiHuQOKbNbBrzs63m"
		username = "testuser"
		password = "testpassword"
		email = "test@test.org"
		if err := ioutil.WriteFile(htpasswdPath, []byte(userpasswd), os.FileMode(0644)); err != nil {
			return nil, err
		}
		htpasswd = fmt.Sprintf(`auth:
    htpasswd:
        realm: basic-realm
        path: %s
`, htpasswdPath)
	}
	confPath := filepath.Join(tmp, "config.yaml")
	config, err := os.Create(confPath)
	if err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(config, template, tmp, url, htpasswd); err != nil {
		os.RemoveAll(tmp)
		return nil, err
	}

	dockerRegistryBinaryPath, err := locateDockerRegistryBinary()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(dockerRegistryBinaryPath, "serve", confPath)
	if err := cmd.Start(); err != nil {
		os.RemoveAll(tmp)
		return nil, err
	}
	return &testRegistryV2{
		cmd:      cmd,
		url:      url,
		dir:      tmp,
		username: username,
		password: password,
		email:    email,
	}, nil
}

func (t *testRegistryV2) Ping() error {
	// We always ping through HTTP for our test registry.
	resp, err := http.Get(fmt.Sprintf("http://%s/v2/", t.url))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("registry ping replied with an unexpected status code %d", resp.StatusCode)
	}
	return nil
}

func (t *testRegistryV2) Close() {
	t.cmd.Process.Kill()
	os.RemoveAll(t.dir)
}
