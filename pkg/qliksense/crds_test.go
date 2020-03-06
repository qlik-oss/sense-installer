package qliksense

import (
	"testing"

	kapi_config "github.com/qlik-oss/k-apis/pkg/config"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func TestGetQliksenseInitCrd(t *testing.T) {
	someTmpRepoPath, err := downloadFromGitRepoToTmpDir(defaultConfigRepoGitUrl, "master")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	crdFromContextConfig, err := getQliksenseInitCrd(&qapi.QliksenseCR{
		KApiCr: kapi_config.KApiCr{
			Spec: &kapi_config.CRSpec{
				ManifestsRoot: someTmpRepoPath,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	crdFromDownloadedConfig, err := getQliksenseInitCrd(&qapi.QliksenseCR{
		KApiCr: kapi_config.KApiCr{
			Spec: &kapi_config.CRSpec{
				ManifestsRoot: "",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if crdFromContextConfig != crdFromDownloadedConfig {
		t.Fatalf("expected %v to equal %v, but they didn't", crdFromContextConfig, crdFromDownloadedConfig)
	}
}
