package qliksense

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"

	kapis_git "github.com/qlik-oss/k-apis/pkg/git"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"gopkg.in/yaml.v2"
)

type patch struct {
	Target struct {
		Kind          string `yaml:"kind"`
		LabelSelector string `yaml:"labelSelector"`
	} `yaml:"target"`
	Patch string `yaml:"patch"`
}

type annotationTransformer struct {
	APIVersion string `yaml:"apiVersion"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type helmChart struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	ReleaseNamespace string `yaml:"releaseNamespace"`
	ChartHome        string `yaml:"chartHome"`
	ChartRepo        string `yaml:"chartRepo"`
	ChartName        string `yaml:"chartName"`
	ChartVersion     string `yaml:"chartVersion"`
}

type VersionOutput struct {
	QliksenseVersion string   `yaml:"qlikSenseVersion"`
	Images           []string `yaml:"images"`
}

type nullWriter struct {
}

func (nw *nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

const (
	defaultProfile          = "docker-desktop"
	defaultConfigRepoGitUrl = "https://github.com/qlik-oss/qliksense-k8s"
)

func (q *Qliksense) About(gitRef, profile string) (*VersionOutput, error) {
	configDirectory, isTemporary, profile, err := q.getConfigDirectory(defaultConfigRepoGitUrl, gitRef, profile)
	if err != nil {
		return nil, err
	} else if isTemporary {
		defer os.RemoveAll(configDirectory)
	}
	return q.AboutDir(configDirectory, profile)
}

func (q *Qliksense) AboutDir(configDirectory, profile string) (*VersionOutput, error) {
	if chartVersion, err := getChartVersion(filepath.Join(configDirectory, "manifests", "base", "transformers", "release", "annotations.yaml"), "app.kubernetes.io/version"); err != nil {
		return nil, err
	} else if kuzManifest, err := executeKustomizeBuildWithStdoutProgress(filepath.Join(configDirectory, "manifests", profile)); err != nil {
		return nil, err
	} else if images, err := getImageList(kuzManifest); err != nil {
		return nil, err
	} else {
		return &VersionOutput{
			QliksenseVersion: chartVersion,
			Images:           images,
		}, nil
	}
}

func (q *Qliksense) getConfigDirectory(gitUrl, gitRef, profileEntered string) (dir string, isTemporary bool, profile string, err error) {
	profile = profileEntered
	if profile == "" {
		profile = defaultProfile
	}

	if gitRef != "" {
		if dir, err = DownloadFromGitRepoToTmpDir(gitUrl, gitRef); err != nil {
			return "", false, "", err
		} else {
			return dir, true, profile, nil
		}
	}

	var exists bool
	exists, dir, err = configExistsInCurrentDirectory(profile)
	if err != nil {
		return "", false, "", err
	} else if exists {
		return dir, false, profile, nil
	}

	var profileFromCurrentContext string
	exists, dir, profileFromCurrentContext, err = q.configExistsInCurrentContext()
	if err != nil {
		return "", false, "", err
	} else if exists {
		if profileEntered == "" {
			profile = profileFromCurrentContext
		}
		return dir, false, profile, nil
	}

	if dir, err = DownloadFromGitRepoToTmpDir(gitUrl, "master"); err != nil {
		return "", false, "", err
	} else {
		return dir, true, profile, nil
	}
}

//DownloadFromGitRepoToTmpDir download git repo to a temporary directory
func DownloadFromGitRepoToTmpDir(gitUrl, gitRef string) (string, error) {
	if tmpDir, err := ioutil.TempDir("", ""); err != nil {
		return "", err
	} else {
		downloadPath := path.Join(tmpDir, "repo")
		if err := downloadFromGitRepo(gitUrl, gitRef, downloadPath); err != nil {
			_ = os.RemoveAll(tmpDir)
			return "", err
		} else {
			return downloadPath, nil
		}
	}
}

func downloadFromGitRepo(gitUrl, gitRef, destDir string) error {
	if repo, err := kapis_git.CloneRepository(destDir, gitUrl, nil); err != nil {
		return err
	} else {
		return kapis_git.Checkout(repo, gitRef, "", nil)
	}
}

func configExistsInCurrentDirectory(profile string) (exists bool, currentDirectory string, err error) {
	currentDirectory, err = os.Getwd()
	if err == nil {
		info, err := os.Stat(path.Join(currentDirectory, "manifests", profile))
		if err == nil && info.IsDir() {
			exists = true
		}
	}
	return exists, currentDirectory, err
}

func (q *Qliksense) configExistsInCurrentContext() (exists bool, directory string, profile string, err error) {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	if currentCr, err := qConfig.GetCurrentCR(); err != nil {
		return false, "", "", err
	} else if currentCr.Spec.ManifestsRoot == "" {
		return false, "", "", nil
	} else {
		return true, currentCr.Spec.GetManifestsRoot(), currentCr.Spec.Profile, nil
	}
}

func getImageList(yamlContent []byte) ([]string, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(yamlContent))
	var resource map[string]interface{}
	imageMap := make(map[string]bool)
	for {
		err := decoder.Decode(&resource)
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		traverseYamlDecodedMapRecursively(reflect.ValueOf(resource), []string{}, func(path []string, val interface{}) {
			if len(path) >= 2 && path[len(path)-1] == "image" &&
				(path[len(path)-2] == "containers" || path[len(path)-2] == "initContainers") {
				if image, ok := val.(string); ok {
					imageMap[image] = true
				}
			}
		})
	}
	var sortedImageList []string
	for image, v := range imageMap {
		sortedImageList = append(sortedImageList, image)
		// a warning "simplify range expression" if written like this 'for image _ :=range imageMap'
		_ = v
	}
	sort.Strings(sortedImageList)
	return sortedImageList, nil
}

func traverseYamlDecodedMapRecursively(val reflect.Value, path []string, visitorFunc func(path []string, val interface{})) {
	kind := val.Kind()
	switch kind {
	case reflect.Interface:
		traverseYamlDecodedMapRecursively(val.Elem(), path, visitorFunc)
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			traverseYamlDecodedMapRecursively(val.Index(i), path, visitorFunc)
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			traverseYamlDecodedMapRecursively(val.MapIndex(key), append(path, key.Interface().(string)), visitorFunc)
		}
	default:
		if kind != reflect.Invalid {
			visitorFunc(path, val.Interface())
		}
	}
}

func getChartVersion(versionFile, versionAnnotation string) (string, error) {
	var annTransformer annotationTransformer

	if bytes, err := ioutil.ReadFile(versionFile); err != nil {
		return "", err
	} else if err = yaml.Unmarshal(bytes, &annTransformer); err != nil {
		return "", err
	}
	if version, ok := annTransformer.Annotations[versionAnnotation]; ok {
		return version, nil
	}
	return "", nil
}
