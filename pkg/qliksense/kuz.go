package qliksense

import (
	"fmt"
	"log"
	"os"
	"time"

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
	kuzManifestDone := make(chan bool)
	go func() {
		kuzManifest, err = executeKustomizeBuild(path)
		kuzManifestDone <- true
	}()
	progressOnTicker := time.NewTicker(500 * time.Millisecond)
	progressOffTicker := time.NewTicker(1000 * time.Millisecond)
	printProgress := func(on bool) {
		if on {
			fmt.Print("...\r")
		} else {
			fmt.Print("   \r")
		}
	}
	for {
		select {
		case <-kuzManifestDone:
			progressOnTicker.Stop()
			progressOffTicker.Stop()
			printProgress(false)
			return kuzManifest, err
		case <-progressOnTicker.C:
			printProgress(true)
		case <-progressOffTicker.C:
			printProgress(false)
		}
	}
}
