package qliksense

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"
)

func Test_About_getImageList(t *testing.T) {
	var testCases = []struct {
		name           string
		k8sYaml        string
		expectedImages []string
	}{
		{
			name: "base",
			k8sYaml: `
apiVersion: apps/v1 # for versions before 1.9.0 use apps/v1beta2
kind: Deployment
metadata:
  name: nginx-deployment-1
spec:
  selector:
    matchLabels:
      app: nginx
  replicas: 2 # tells deployment to run 2 pods matching the template
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Secret
metadata:
  creationTimestamp: 2018-11-15T20:46:46Z
  name: mysecret
  namespace: default
  resourceVersion: "7579"
  uid: 91460ecb-e917-11e8-98f2-025000000001
type: Opaque
data:
  username: YWRtaW5pc3RyYXRvcg==
---
apiVersion: apps/v1 # for versions before 1.9.0 use apps/v1beta2
kind: Deployment
metadata:
  name: nginx-deployment-2
spec:
  selector:
    matchLabels:
      app: nginx
  replicas: 2 # tells deployment to run 2 pods matching the template
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
apiVersion: v1
kind: Service
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  ports:
  - port: 80
    name: web
  clusterIP: None
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  selector:
    matchLabels:
      app: nginx # has to match .spec.template.metadata.labels
  serviceName: "nginx"
  replicas: 3 # by default is 1
  template:
    metadata:
      labels:
        app: nginx # has to match .spec.selector.matchLabels
    spec:
      terminationGracePeriodSeconds: 10
      containers:
      - name: nginx
        image: k8s.gcr.io/nginx-slim:0.8
        ports:
        - containerPort: 80
          name: web
        volumeMounts:
        - name: www
          mountPath: /usr/share/nginx/html
  volumeClaimTemplates:
  - metadata:
      name: www
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "my-storage-class"
      resources:
        requests:
          storage: 1Gi
---
apiVersion: batch/v1
kind: Job
metadata:
  name: pi
spec:
  template:
    spec:
      containers:
      - name: pi
        image: perl
        command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
      restartPolicy: Never
  backoffLimit: 4
---
apiVersion: v1
kind: Pod
metadata:
  name: init-demo
spec:
  containers:
  - name: nginx
    image: nginx
    env:
    - name: FOO
      value: null 
    ports:
    - containerPort: 80
    volumeMounts:
    - name: workdir
      mountPath: /usr/share/nginx/html
  # These containers are run during pod initialization
  initContainers:
  - name: install
    image: busybox
    command:
    - wget
    - "-O"
    - "/work-dir/index.html"
    - http://kubernetes.io
    volumeMounts:
    - name: workdir
      mountPath: "/work-dir"
  dnsPolicy: Default
  volumes:
  - name: workdir
    emptyDir: {}
`,
			expectedImages: []string{"busybox", "k8s.gcr.io/nginx-slim:0.8", "nginx", "nginx:1.7.9", "perl"},
		},
		{
			name: "works for custom resources and CronJobs",
			k8sYaml: `
apiVersion: "qixmanager.qlik.com/v1"
kind: "Engine"
metadata:
  name: release-name-engine-reload
spec:
  metadata:
    labels:
      qix-engine: qix-engine
    annotations:
      prometheus.io/scrape: "true"
  workloadType: "reload"
  podSpec:
    imagePullSecrets:
      - name: artifactory-docker-secret      
    dnsConfig:
      options:
      - name: timeout
        value: "1"
      - name: single-request-reopen
    containers:
      - name: engine-reload
        image: another-engine
---
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: hello
spec:
  schedule: "*/1 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: hello
            image: busybox2
            args:
            - /bin/sh
            - -c
            - date; echo Hello from the Kubernetes cluster
          restartPolicy: OnFailure
`,
			expectedImages: []string{"another-engine", "busybox2"},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			images, err := getImageList([]byte(testCase.k8sYaml))
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}
			if !reflect.DeepEqual(images, testCase.expectedImages) {
				t.Fatalf("expected %v, but got: %v\n", testCase.expectedImages, images)
			}
		})
	}
}

func Test_About_getConfigDirectory(t *testing.T) {
	verifyAsdBranch := func(configDir string) (ok bool, reason string) {
		tmpDir := os.TempDir()

		if (path.Clean(path.Dir(path.Dir(configDir))) != path.Clean(tmpDir)) || (path.Base(configDir) != "repo") {
			return false, fmt.Sprintf("expected config directory path: %v to be under: %v and terminate with repo", configDir, tmpDir)
		}

		if info, err := os.Stat(path.Join(configDir, "asdczxc")); err != nil || !info.Mode().IsRegular() {
			return false, fmt.Sprintf(`expected to find file: "asdczxc" under directory: %v`, configDir)
		}
		return true, ""
	}

	verifyMasterBranch := func(configDir string) (ok bool, reason string) {
		tmpDir := os.TempDir()

		if (path.Clean(path.Dir(path.Dir(configDir))) != path.Clean(tmpDir)) || (path.Base(configDir) != "repo") {
			return false, fmt.Sprintf("expected config directory path: %v to be under: %v and terminate with repo", configDir, tmpDir)
		}

		if _, err := os.Stat(path.Join(configDir, "asdczxc")); err == nil || !os.IsNotExist(err) {
			return false, fmt.Sprintf(`expected to NOT find file: "asdczxc"" under directory: %v`, configDir)
		}

		if info, err := os.Stat(path.Join(configDir, "sad")); err != nil || !info.Mode().IsRegular() {
			return false, fmt.Sprintf(`expected to find file: "sad"" under directory: %v`, configDir)
		}
		return true, ""
	}

	var testCases = []struct {
		name    string
		setup   func(t *testing.T) (gitUrl, gitRef, profileEntered string)
		verify  func(configDir string, isTemporary bool, profile string) (ok bool, reason string, err error)
		cleanup func(configDir string) error
	}{
		{
			name: "config in current directory and default profile",
			setup: func(t *testing.T) (gitUrl, gitRef, profileEntered string) {
				currentDirectory, err := os.Getwd()
				if err != nil {
					t.Fatalf("error obtaining current directory: %v\n", err)
				}

				defaultProfilePath := path.Join(currentDirectory, "manifests", "docker-desktop")
				err = os.MkdirAll(defaultProfilePath, os.ModePerm)
				if err != nil {
					t.Fatalf("error making path: %v, err: %v\n", defaultProfilePath, err)
				}
				return "no-clone-for-you", "", ""
			},
			verify: func(configDir string, isTemporary bool, profile string) (ok bool, reason string, err error) {
				currentDirectory, err := os.Getwd()
				if err != nil {
					return false, "", err
				}

				if configDir != currentDirectory {
					return false, fmt.Sprintf("expected config directory: %v to equal current directory: %v", configDir, currentDirectory), nil
				}

				if isTemporary {
					return false, "expected isTemporary to be false", nil
				}

				if profile != "docker-desktop" {
					return false, fmt.Sprintf("expected profile to be: docker-desktop, but it was: %v", profile), nil
				}

				return true, "", nil
			},
			cleanup: func(configDir string) error {
				if currentDirectory, err := os.Getwd(); err != nil {
					return err
				} else if err := os.RemoveAll(path.Join(currentDirectory, "manifests")); err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "config in current directory and profile specified",
			setup: func(t *testing.T) (gitUrl, gitRef, profileEntered string) {
				currentDirectory, err := os.Getwd()
				if err != nil {
					t.Fatalf("error obtaining current directory: %v\n", err)
				}

				profileEntered = "foo"
				defaultProfilePath := path.Join(currentDirectory, "manifests", profileEntered)
				err = os.MkdirAll(defaultProfilePath, os.ModePerm)
				if err != nil {
					t.Fatalf("error making path: %v, err: %v\n", defaultProfilePath, err)
				}
				return "no-clone-for-you", "", profileEntered
			},
			verify: func(configDir string, isTemporary bool, profile string) (ok bool, reason string, err error) {
				currentDirectory, err := os.Getwd()
				if err != nil {
					return false, "", err
				}

				if configDir != currentDirectory {
					return false, fmt.Sprintf("expected config directory: %v to equal current directory: %v", configDir, currentDirectory), nil
				}

				if isTemporary {
					return false, "expected isTemporary to be false", nil
				}

				if profile != "foo" {
					return false, fmt.Sprintf("expected profile to be: foo, but it was: %v", profile), nil
				}

				return true, "", nil
			},
			cleanup: func(configDir string) error {
				if currentDirectory, err := os.Getwd(); err != nil {
					return err
				} else if err := os.RemoveAll(path.Join(currentDirectory, "manifests")); err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "config downloaded from git based on specific git ref and default profile used",
			setup: func(t *testing.T) (gitUrl, gitRef, profileEntered string) {
				return "https://github.com/test/HelloWorld", "asd", ""
			},
			verify: func(configDir string, isTemporary bool, profile string) (ok bool, reason string, err error) {
				ok, reason = verifyAsdBranch(configDir)
				if !ok {
					return ok, reason, nil
				}

				if !isTemporary {
					return false, "expected isTemporary to be true", nil
				}

				if profile != "docker-desktop" {
					return false, fmt.Sprintf("expected profile to be: docker-desktop, but it was: %v", profile), nil
				}

				return true, "", nil
			},
			cleanup: func(configDir string) error {
				tmpDir := os.TempDir()

				if path.Clean(path.Dir(path.Dir(configDir))) == path.Clean(tmpDir) && path.Base(configDir) == "repo" {
					tmpTmpDir := path.Dir(configDir)
					if err := os.RemoveAll(tmpTmpDir); err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			name: "config downloaded from git based on specific git ref and profile specified",
			setup: func(t *testing.T) (gitUrl, gitRef, profileEntered string) {
				return "https://github.com/test/HelloWorld", "asd", "foo"
			},
			verify: func(configDir string, isTemporary bool, profile string) (ok bool, reason string, err error) {
				ok, reason = verifyAsdBranch(configDir)
				if !ok {
					return ok, reason, nil
				}

				if !isTemporary {
					return false, "expected isTemporary to be true", nil
				}

				if profile != "foo" {
					return false, fmt.Sprintf("expected profile to be: foo, but it was: %v", profile), nil
				}

				return true, "", nil
			},
			cleanup: func(configDir string) error {
				tmpDir := os.TempDir()

				if path.Clean(path.Dir(path.Dir(configDir))) == path.Clean(tmpDir) && path.Base(configDir) == "repo" {
					tmpTmpDir := path.Dir(configDir)
					if err := os.RemoveAll(tmpTmpDir); err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			name: "config downloaded from git from master branch and default profile used",
			setup: func(t *testing.T) (gitUrl, gitRef, profileEntered string) {
				return "https://github.com/test/HelloWorld", "", ""
			},
			verify: func(configDir string, isTemporary bool, profile string) (ok bool, reason string, err error) {
				ok, reason = verifyMasterBranch(configDir)
				if !ok {
					return ok, reason, nil
				}

				if !isTemporary {
					return false, "expected isTemporary to be true", nil
				}

				if profile != "docker-desktop" {
					return false, fmt.Sprintf("expected profile to be: docker-desktop, but it was: %v", profile), nil
				}

				return true, "", nil
			},
			cleanup: func(configDir string) error {
				tmpDir := os.TempDir()

				if path.Clean(path.Dir(path.Dir(configDir))) == path.Clean(tmpDir) && path.Base(configDir) == "repo" {
					tmpTmpDir := path.Dir(configDir)
					if err := os.RemoveAll(tmpTmpDir); err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			name: "config downloaded from git from master branch and profile specified",
			setup: func(t *testing.T) (gitUrl, gitRef, profileEntered string) {
				return "https://github.com/test/HelloWorld", "", "foo"
			},
			verify: func(configDir string, isTemporary bool, profile string) (ok bool, reason string, err error) {
				ok, reason = verifyMasterBranch(configDir)
				if !ok {
					return ok, reason, nil
				}

				if !isTemporary {
					return false, "expected isTemporary to be true", nil
				}

				if profile != "foo" {
					return false, fmt.Sprintf("expected profile to be: foo, but it was: %v", profile), nil
				}

				return true, "", nil
			},
			cleanup: func(configDir string) error {
				tmpDir := os.TempDir()

				if path.Clean(path.Dir(path.Dir(configDir))) == path.Clean(tmpDir) && path.Base(configDir) == "repo" {
					tmpTmpDir := path.Dir(configDir)
					if err := os.RemoveAll(tmpTmpDir); err != nil {
						return err
					}
				}
				return nil
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			configDirectory, isTemporary, profile, err := getConfigDirectory(testCase.setup(t))
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			if ok, reason, err := testCase.verify(configDirectory, isTemporary, profile); err != nil {
				t.Fatalf("unexpected verification error: %v\n", err)
			} else if !ok {
				t.Fatalf("verification failed: %v\n", reason)
			} else if err := testCase.cleanup(configDirectory); err != nil {
				t.Fatalf("unexpected cleanup error: %v\n", err)
			}
		})
	}
}
