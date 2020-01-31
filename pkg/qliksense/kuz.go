package qliksense

import (
	"log"
	"os"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
)

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
