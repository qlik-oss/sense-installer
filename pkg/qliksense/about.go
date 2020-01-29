package qliksense

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"gopkg.in/yaml.v2"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
)

type patch struct {
	Target struct {
		Kind          string `yaml:"kind"`
		LabelSelector string `yaml:"labelSelector"`
	} `yaml:"target"`
	Patch string `yaml:"patch"`
}

type selectivePatch struct {
	APIVersion string `yaml:"apiVersion"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Enabled bool    `yaml:"enabled"`
	Patches []patch `yaml:"patches"`
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

type versionOutput struct {
	QliksenseVersion string   `yaml:"qlikSenseVersion"`
	Images           []string `yaml:"images"`
}

type nullWriter struct {
}

func (nw *nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (p *Qliksense) About(tag, directory, profile string) ([]byte, error) {
	chartVersion, err := getChartVersion(filepath.Join(directory, "transformers", "qseokversion.yaml"), "qliksense")
	if err != nil {
		return nil, err
	}

	kuzManifest, err := executeKustomizeBuild(filepath.Join(directory, "manifests", profile))
	if err != nil {
		return nil, err
	}

	images, err := getImageList(kuzManifest)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(versionOutput{
		QliksenseVersion: chartVersion,
		Images:           images,
	})
}

func executeKustomizeBuild(directory string) ([]byte, error) {
	log.SetOutput(&nullWriter{})
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	fSys := filesys.MakeFsOnDisk()
	options := &krusty.Options{
		DoLegacyResourceSort: false,
		LoadRestrictions:     types.LoadRestrictionsNone,
		DoPrune:              false,
	}
	pluginConfig, err := konfig.EnabledPluginConfig()
	if err != nil {
		return nil, err
	}
	options.PluginConfig = pluginConfig
	k := krusty.MakeKustomizer(fSys, options)
	resMap, err := k.Run(directory)
	if err != nil {
		return nil, err
	}
	return resMap.AsYaml()
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
	for image, _ := range imageMap {
		sortedImageList = append(sortedImageList, image)
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

func getChartVersion(versionFile, chartName string) (string, error) {
	var patchInst patch
	var selPatch selectivePatch
	var chart helmChart

	if bytes, err := ioutil.ReadFile(versionFile); err != nil {
		return "", err
	} else if err = yaml.Unmarshal(bytes, &selPatch); err != nil {
		return "", err
	}
	for _, patchInst = range selPatch.Patches {
		if err := yaml.Unmarshal([]byte(patchInst.Patch), &chart); err == nil {
			if chart.ChartName == chartName {
				return chart.ChartVersion, nil
			}
		}
	}
	return "", nil
}
