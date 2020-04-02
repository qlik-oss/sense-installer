package qliksense

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/api"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // don't delete this line ref: https://github.com/kubernetes/client-go/issues/242
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
)

//ExecuteKustomizeBuild execute kustomize to the directory and return manifest as byte array
func ExecuteKustomizeBuild(directory string) ([]byte, error) {
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
		return ExecuteKustomizeBuild(path)
	}, "...")
	if err != nil {
		return nil, err
	}
	return result.([]byte), nil
}

//GetYamlsFromMultiDoc filter yaml docs from multiyaml based on kind
func GetYamlsFromMultiDoc(multiYaml string, kind string) string {
	yamlDocs := strings.Split(string(multiYaml), "---")
	resultDocs := ""
	for _, doc := range yamlDocs {
		scanner := bufio.NewScanner(strings.NewReader(doc))
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "kind: "+kind) {
				resultDocs = resultDocs + "\n---\n" + doc
				break
			}
		}
	}
	return resultDocs
}
