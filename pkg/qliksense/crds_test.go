package qliksense

import (
	"io/ioutil"
	"os"
	"testing"

	apixv1beta1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"

	"github.com/gobuffalo/packr/v2"

	kapi_config "github.com/qlik-oss/k-apis/pkg/config"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func TestGetQliksenseInitCrd(t *testing.T) {
	someTmpRepoPath, err := DownloadFromGitRepoToTmpDir(defaultConfigRepoGitUrl, "master")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	crdFromContextConfig, err := getQliksenseInitCrds(&qapi.QliksenseCR{
		KApiCr: kapi_config.KApiCr{
			Spec: &kapi_config.CRSpec{
				ManifestsRoot: someTmpRepoPath,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	crdFromDownloadedConfig, err := getQliksenseInitCrds(&qapi.QliksenseCR{
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

func TestCheckAllCrdsInstalled(t *testing.T) {
	t.Skip("Skipping this test because it makes kubernetes calls")

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
  profile: docker-desktop
`)

	q := &Qliksense{
		QliksenseHome: tmpQlikSenseHome,
		CrdBox:        packr.New("crds", "./crds"),
	}

	if err := q.FetchQK8s("v1.50.3"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if allInstalled, err := q.CheckAllCrdsInstalled(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if allInstalled {
		t.Fatal("expected crds to NOT be installed at this point")
	}

	if err := q.InstallCrds(&CrdCommandOptions{All: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if allInstalled, err := q.CheckAllCrdsInstalled(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if !allInstalled {
		t.Fatal("expected crds to BE installed at this point")
	}

	//cleanup:
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	customResourceDefinitionInterface, err := getCustomResourceDefinitionInterface()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if engineCRDs, err := getQliksenseInitCrds(qcr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if err := deleteCrds(engineCRDs, customResourceDefinitionInterface); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if customCrd, err := getCustomCrds(qcr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if err := deleteCrds(customCrd, customResourceDefinitionInterface); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := deleteCrds(q.GetOperatorCRDString(), customResourceDefinitionInterface); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func deleteCrds(crds string, customResourceDefinitionInterface apixv1beta1client.CustomResourceDefinitionInterface) error {
	kuzResourceFactory := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), nil)
	if kuzResMap, err := kuzResourceFactory.NewResMapFromBytes([]byte(crds)); err != nil {
		return err
	} else {
		for _, kuzRes := range kuzResMap.Resources() {
			if err := customResourceDefinitionInterface.Delete(kuzRes.GetName(), &v1.DeleteOptions{}); err != nil {
				return err
			}
		}
		return nil
	}
}
