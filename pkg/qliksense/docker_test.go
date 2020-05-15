package qliksense

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	imageTypes "github.com/containers/image/v5/types"
	"github.com/qlik-oss/sense-installer/pkg/api"
	"golang.org/x/net/context"
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

func Test_getSelfSignedCertAndKey(t *testing.T) {
	host := "andriy.registry.com"
	validity := time.Hour * 24 * 365
	selfSignedCert, key, err := getSelfSignedCertAndKey(host, validity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fmt.Print(string(selfSignedCert))
	fmt.Print(string(key))
}

type clientAuthType byte

const (
	clientAuthNotProvided clientAuthType = iota
	clientAuthProvided
	clientAuthProvidedButIncorrect
)

func Test_Pull_Push_ImagesForCurrentCR(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pull/push tests in short mode")
	}
	var testCases = []struct {
		name              string
		registryAuth      bool
		clientAuth        clientAuthType
		expectPushSuccess bool
	}{
		{
			name:              "registry does not require auth and we do not provide auth",
			registryAuth:      false,
			clientAuth:        clientAuthNotProvided,
			expectPushSuccess: true,
		},
		{
			name:              "registry does not require auth but we provide auth",
			registryAuth:      false,
			clientAuth:        clientAuthProvided,
			expectPushSuccess: true,
		},
		{
			name:              "registry requires auth but we do not provide auth",
			registryAuth:      true,
			clientAuth:        clientAuthNotProvided,
			expectPushSuccess: false,
		},
		{
			name:              "registry requires auth but we provide wrong auth",
			registryAuth:      true,
			clientAuth:        clientAuthProvidedButIncorrect,
			expectPushSuccess: false,
		},
		{
			name:              "registry requires auth and we provide auth",
			registryAuth:      true,
			clientAuth:        clientAuthProvided,
			expectPushSuccess: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registryURI := "127.0.0.1:5555"
			registry, err := setupRegistryV2At(registryURI, testCase.registryAuth)
			if registry != nil {
				defer func() {
					registry.Close()
					//fmt.Printf("registry stdout:\n%v\n", registry.stdOutBuffer.String())
					//fmt.Printf("registry stderr:\n%v\n", registry.stdErrBuffer.String())
				}()
			}
			if err != nil {
				t.Fatalf("unexpected error setting up local registry: %v", err)
			}

			tmpQlikSenseHome, err := ioutil.TempDir("", "tmp-qlik-sense-home-")
			if err != nil {
				t.Fatalf("unexpected error creating tmp dir: %v", err)
			}
			defer os.RemoveAll(tmpQlikSenseHome)

			if err := setupQlikSenseHome(t, tmpQlikSenseHome, registry, testCase.clientAuth); err != nil {
				t.Fatalf("unexpected error setting up qliksense home: %v", err)
			}
			q := &Qliksense{
				QliksenseHome: tmpQlikSenseHome,
			}
			var versionOut VersionOutput

			if err := q.PullImagesForCurrentCR(); err != nil {
				t.Fatalf("unexpected pull error: %v", err)
			} else if versionOutBytes, err := ioutil.ReadFile(path.Join(tmpQlikSenseHome, "images", "foo")); err != nil {
				t.Fatalf("unexpected error reading version file: %v", err)
			} else if err = yaml.Unmarshal(versionOutBytes, &versionOut); err != nil {
				t.Fatalf("unexpected error unmarshalling version file: %v", err)
			} else if len(versionOut.Images) != 1 || versionOut.Images[0] != "alpine:latest" {
				t.Fatal(`did not find "alpine:latest"" in the version file`)
			} else if infos, err := ioutil.ReadDir(path.Join(tmpQlikSenseHome, "images", "index", "alpine", "latest")); err != nil || len(infos) == 0 {
				t.Fatal("expected images/index/alpine/latest directory to be non-empty")
			} else if blobInfos, err := ioutil.ReadDir(path.Join(tmpQlikSenseHome, "images", "blobs", "sha256")); err != nil || len(blobInfos) == 0 {
				t.Fatal("expected images/blobs/sha256 directory to be non-empty")
			}

			if testCase.expectPushSuccess {
				if err := q.PushImagesForCurrentCR(); err != nil {
					t.Fatalf("unexpected push error: %v", err)
				} else if tmpImagesDir, err := ioutil.TempDir("", "tmp-images-"); err != nil {
					t.Fatalf("unexpected error creating tmp dir: %v", err)
				} else if err := testPullImage(fmt.Sprintf("%s/alpine:latest", registryURI), tmpImagesDir, registry); err != nil {
					t.Fatalf("unexpected error pulling alpine:latest from the local registry: %v", err)
				} else if infos, err := ioutil.ReadDir(path.Join(tmpImagesDir, "index", "alpine", "latest")); err != nil || len(infos) == 0 {
					t.Fatal("expected index/alpine/latest directory to be non-empty")
				} else if blobInfos, err := ioutil.ReadDir(path.Join(tmpImagesDir, "blobs", "sha256")); err != nil || len(blobInfos) == 0 {
					t.Fatal("expected blobs/sha256 directory to be non-empty")
				}
			} else {
				if err := q.PushImagesForCurrentCR(); err == nil {
					t.Fatal("unexpected push success")
				}
			}
		})
	}
}

func Test_appendAdditionalImages(t *testing.T) {
	tmpQlikSenseHome, err := ioutil.TempDir("", "tmp-qlik-sense-home-")
	if err != nil {
		t.Fatalf("unexpected error creating tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpQlikSenseHome)

	setupQliksenseTestDefaultContext(t, tmpQlikSenseHome, `
apiVersion: qlik.com/v1
kind: Qliksense
metadata:
  name: qlik-default
spec:
  opsRunner:
    image: some-gitops-image
`)

	q := &Qliksense{
		QliksenseHome: tmpQlikSenseHome,
	}

	pf := api.NewPreflightConfig(q.QliksenseHome)
	if err := pf.Initialize(); err != nil {
		t.Fatalf("unexpected error initializing preflight: %v", err)
	}

	qConfig := api.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		t.Fatalf("unexpected error getting current CR: %v", err)
	}

	images := make([]string, 0)
	if err := q.appendAdditionalImages(&images, qcr); err != nil {
		t.Fatalf("unexpected error appending additional images: %v", err)
	}

	expectedNumberAdditionalImages := 5
	if len(images) != expectedNumberAdditionalImages {
		t.Fatalf("unexpected number of additional images: %v, expected: %v", len(images), expectedNumberAdditionalImages)
	}

	haveMatchingImage := func(test func(string) bool) bool {
		for _, image := range images {
			if test(image) {
				return true
			}
		}
		return false
	}
	if !haveMatchingImage(func(image string) bool {
		return strings.Contains(image, "qlik-docker-oss.bintray.io/qliksense-operator:v")
	}) {
		t.Fatal("expected to find the operator image in the list, but it wasn't there")
	}
	if !haveMatchingImage(func(image string) bool {
		return image == "some-gitops-image"
	}) {
		t.Fatal("expected to find the GitOps image in the list, but it wasn't there")
	}
	if !haveMatchingImage(func(image string) bool {
		return image == "nginx"
	}) {
		t.Fatal("expected to find the nginx Preflight image in the list, but it wasn't there")
	}
	if !haveMatchingImage(func(image string) bool {
		return image == "subfuzion/netcat"
	}) {
		t.Fatal("expected to find the netcat Preflight image in the list, but it wasn't there")
	}
	if !haveMatchingImage(func(image string) bool {
		return image == "qlik-docker-oss.bintray.io/preflight-mongo"
	}) {
		t.Fatal("expected to find the mongo Preflight image in the list, but it wasn't there")
	}
}

func setupQlikSenseHome(t *testing.T, tmpQlikSenseHome string, registry *testRegistryV2, clientAuth clientAuthType) error {
	version := "foo"
	manifestsRootDir := filepath.ToSlash(path.Join(tmpQlikSenseHome, "contexts", "qlik-default", "repo", version))
	cr := fmt.Sprintf(`
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
`, version, registry.url, manifestsRootDir)
	setupQliksenseTestDefaultContext(t, tmpQlikSenseHome, cr)

	if clientAuth == clientAuthProvided || clientAuth == clientAuthProvidedButIncorrect {
		if registry.username == "" || clientAuth == clientAuthProvidedButIncorrect {
			registry.username = "bad"
		}
		if registry.password == "" || clientAuth == clientAuthProvidedButIncorrect {
			registry.password = "worse"
		}
		qConfig := api.NewQConfig(tmpQlikSenseHome)
		if err := qConfig.SetPushDockerConfigJsonSecret(&api.DockerConfigJsonSecret{
			Uri:      registry.url,
			Username: registry.username,
			Password: registry.password,
		}); err != nil {
			return err
		}
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
	cmd          *exec.Cmd
	url          string
	dir          string
	username     string
	password     string
	email        string
	stdOutBuffer *bytes.Buffer
	stdErrBuffer *bytes.Buffer
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
		if err := reg.Ping("http"); err == nil {
			fmt.Print("registry http ping succeeded\n")
			break
		} else {
			fmt.Printf("registry http ping error: %v\n", err)
		}
		if err := reg.Ping("https"); err == nil {
			fmt.Print("registry https ping succeeded\n")
			break
		} else {
			fmt.Printf("registry https ping error: %v\n", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err != nil {
		return reg, errors.New("timeout waiting for test registry to become available")
	}
	return reg, nil
}

func newTestRegistryV2At(url string, auth bool) (*testRegistryV2, error) {
	tmp, err := ioutil.TempDir("", "registry-test-")
	if err != nil {
		return nil, err
	}
	template := `version: 0.1
loglevel: info
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
	var env []string
	if auth {
		if certificate, key, err := getSelfSignedCertAndKey("localhost", time.Hour*24); err != nil {
			return nil, err
		} else {
			certPath := filepath.Join(tmp, "domain.crt")
			if err := ioutil.WriteFile(certPath, certificate, os.FileMode(0644)); err != nil {
				return nil, err
			}
			keyPath := filepath.Join(tmp, "domain.key")
			if err := ioutil.WriteFile(keyPath, key, os.FileMode(0644)); err != nil {
				return nil, err
			}
			env = append(env, fmt.Sprintf("REGISTRY_HTTP_TLS_CERTIFICATE=%v", certPath))
			env = append(env, fmt.Sprintf("REGISTRY_HTTP_TLS_KEY=%v", keyPath))
		}

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
	cmd.Env = env
	stdOutBuf, stdErrBuf, err := consumeAndLogOutputs(fmt.Sprintf("registry-%s", url), cmd)
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		os.RemoveAll(tmp)
		return nil, err
	}
	return &testRegistryV2{
		cmd:          cmd,
		url:          url,
		dir:          tmp,
		username:     username,
		password:     password,
		email:        email,
		stdOutBuffer: stdOutBuf,
		stdErrBuffer: stdErrBuf,
	}, nil
}

func (t *testRegistryV2) Ping(protocol string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(fmt.Sprintf("%v://%s/v2/", protocol, t.url))
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

func consumeAndLogOutputStream(id string, f io.ReadCloser) *bytes.Buffer {
	buff := &bytes.Buffer{}
	go func() {
		defer func() {
			f.Close()
			fmt.Fprintf(buff, "[%s]: Closed\n", id)
		}()
		buf := make([]byte, 1024)
		for {
			fmt.Fprintf(buff, "[%s]: waiting\n", id)
			n, err := f.Read(buf)
			fmt.Fprintf(buff, "[%s]: got %d,%#v: %s\n", id, n, err, strings.TrimSuffix(string(buf[:n]), "\n"))
			if n <= 0 {
				break
			}
		}
	}()
	return buff
}

// consumeAndLogOutputs causes all output to stdout and stderr from an *exec.Cmd to be logged to c
func consumeAndLogOutputs(id string, cmd *exec.Cmd) (*bytes.Buffer, *bytes.Buffer, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	return consumeAndLogOutputStream(id+" stdout", stdout), consumeAndLogOutputStream(id+" stderr", stderr), nil
}

func getSelfSignedCertAndKey(hostname string, validity time.Duration) (certificate, key []byte, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("ailed to generate serial number: %s", err)
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"self-signed"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(validity),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{hostname},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %s", err)
	}
	certificate = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal private key: %v", err)
	}
	key = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	return certificate, key, nil
}

func testPullImage(image, imagesDir string, registry *testRegistryV2) error {
	srcRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%v", image))
	if err != nil {
		return err
	}
	nameTag := getImageNameParts(image)
	targetDir := filepath.Join(imagesDir, imageIndexDirName, nameTag.name, nameTag.tag)
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return err
	}

	destRef, err := alltransports.ParseImageName(fmt.Sprintf("oci:%v", targetDir))
	if err != nil {
		return err
	}

	policyContext, err := signature.NewPolicyContext(&signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}})
	if err != nil {
		return err
	}
	defer policyContext.Destroy()

	fmt.Printf("==> Test is pulling image from %v\n", srcRef.StringWithinTransport())
	sourceCtx := &imageTypes.SystemContext{
		ArchitectureChoice:          "amd64",
		OSChoice:                    "linux",
		DockerInsecureSkipTLSVerify: imageTypes.OptionalBoolTrue,
	}
	if registry.username != "" {
		sourceCtx.DockerAuthConfig = &imageTypes.DockerAuthConfig{
			Username: registry.username,
			Password: registry.password,
		}
	}
	if _, err := copy.Image(context.Background(), policyContext, destRef, srcRef, &copy.Options{
		ReportWriter: os.Stdout,
		SourceCtx:    sourceCtx,
		DestinationCtx: &imageTypes.SystemContext{
			OCISharedBlobDirPath: filepath.Join(imagesDir, imageSharedBlobsDirName),
		},
	}); err != nil {
		return err
	}
	return nil
}
