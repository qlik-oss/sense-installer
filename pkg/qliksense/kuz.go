package qliksense

import (
	"log"
	"os"

	"github.com/qlik-oss/sense-installer/pkg/api"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
)

func executeKustomizeBuild(directory string) ([]byte, error) {
	return executeKustomizeBuildForFileSystem(directory, filesys.MakeFsOnDisk())
}

func executeKustomizeBuildForFileSystem(directory string, fSys filesys.FileSystem) ([]byte, error) {
	log.SetOutput(&nullWriter{})
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	options := &krusty.Options{
		DoLegacyResourceSort: false,
		LoadRestrictions:     types.LoadRestrictionsNone,
		DoPrune:              false,
		PluginConfig:         konfig.DisabledPluginConfig(),
	}
	k := krusty.MakeKustomizer(fSys, options)
	resMap, err := k.Run(directory)
	if err != nil {
		return nil, err
	}
	return resMap.AsYaml()
}

func executeKustomizeBuildWithStdoutProgress(path string) (kuzManifest []byte, err error) {
	result, err := api.ExecuteTaskWithBlinkingStdoutFeedback(func() (interface{}, error) {
		return executeKustomizeBuild(path)
	}, "...")
	if err != nil {
		return nil, err
	}
	return result.([]byte), nil
}
