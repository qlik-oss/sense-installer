package qliksense

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	apixv1beta1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
)

type CrdCommandOptions struct {
	All bool
}

func (q *Qliksense) ViewCrds(opts *CrdCommandOptions) error {
	//io.WriteString(os.Stdout, q.GetCRDString())
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	}
	engineCRD, err := getQliksenseInitCrds(qcr)
	if err != nil {
		return err
	}
	customCrd, err := getCustomCrds(qcr)
	if err != nil {
		return nil
	}

	fmt.Println(engineCRD)
	if customCrd != "" {
		fmt.Println("---")
		fmt.Println(customCrd)
	}

	if opts.All {
		fmt.Println("---")
		fmt.Printf("%s", q.GetOperatorCRDString())
	}
	return nil
}

func (q *Qliksense) InstallCrds(opts *CrdCommandOptions) error {
	// install qliksense-init crd
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	}

	if engineCRD, err := getQliksenseInitCrds(qcr); err != nil {
		return err
	} else if err = qapi.KubectlApply(engineCRD, ""); err != nil {
		return err
	}
	if customCrd, err := getCustomCrds(qcr); err != nil {
		return err
	} else if customCrd != "" {
		if err = qapi.KubectlApply(customCrd, ""); err != nil {
			return err
		}
	}

	if opts.All { // install opeartor crd
		if err := qapi.KubectlApply(q.GetOperatorCRDString(), ""); err != nil {
			fmt.Println("cannot do kubectl apply on opeartor CRD", err)
			return err
		}
	}
	return nil
}

func getQliksenseInitCrds(qcr *qapi.QliksenseCR) (string, error) {
	var repoPath string
	var err error

	if qcr.Spec.GetManifestsRoot() != "" {
		repoPath = qcr.Spec.GetManifestsRoot()
	} else {
		if repoPath, err = DownloadFromGitRepoToTmpDir(defaultConfigRepoGitUrl, "master"); err != nil {
			return "", err
		}
	}

	qInitMsPath := filepath.Join(repoPath, Q_INIT_CRD_PATH)
	if _, err := os.Lstat(qInitMsPath); err != nil {
		// older version of qliksense-init used
		qInitMsPath = filepath.Join(repoPath, "manifests/base/manifests/qliksense-init")
	}
	qInitByte, err := ExecuteKustomizeBuild(qInitMsPath)
	if err != nil {
		fmt.Println("cannot generate crds for qliksense-init", err)
		return "", err
	}
	return string(qInitByte), nil
}

func getCustomCrds(qcr *qapi.QliksenseCR) (string, error) {
	crdPath := qcr.GetCustomCrdsPath()
	if crdPath == "" {
		return "", nil
	}
	qInitByte, err := ExecuteKustomizeBuild(crdPath)
	if err != nil {
		fmt.Println("cannot generate custom crds", err)
		return "", err
	}
	return string(qInitByte), nil
}

func (q *Qliksense) CheckAllCrdsInstalled() (bool, error) {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		return false, err
	}

	customResourceDefinitionInterface, err := getCustomResourceDefinitionInterface()
	if err != nil {
		return false, err
	}

	if engineCRDs, err := getQliksenseInitCrds(qcr); err != nil {
		return false, err
	} else if allInstalled, err := checkCrdsInstalled(engineCRDs, customResourceDefinitionInterface); err != nil {
		return false, err
	} else if !allInstalled {
		return false, nil
	}

	if customCrds, err := getCustomCrds(qcr); err != nil {
		return false, err
	} else if allInstalled, err := checkCrdsInstalled(customCrds, customResourceDefinitionInterface); err != nil {
		return false, err
	} else if !allInstalled {
		return false, nil
	}

	if allInstalled, err := checkCrdsInstalled(q.GetOperatorCRDString(), customResourceDefinitionInterface); err != nil {
		return false, err
	} else if !allInstalled {
		return false, nil
	}

	return true, nil
}

func checkCrdsInstalled(crds string, customResourceDefinitionInterface apixv1beta1client.CustomResourceDefinitionInterface) (bool, error) {
	kuzResourceFactory := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), nil)
	if kuzResMap, err := kuzResourceFactory.NewResMapFromBytes([]byte(crds)); err != nil {
		return false, err
	} else {
		for _, kuzRes := range kuzResMap.Resources() {
			if customResourceDefinition, err := customResourceDefinitionInterface.Get(kuzRes.GetName(), v1.GetOptions{}); err != nil && apierrors.IsNotFound(err) {
				return false, nil
			} else if err != nil {
				return false, err
			} else if customResourceDefinition == nil {
				return false, fmt.Errorf("failed looking up crd: %v", kuzRes.GetName())
			}
		}
		return true, nil
	}
}

func getCustomResourceDefinitionInterface() (apixv1beta1client.CustomResourceDefinitionInterface, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	kubeconfigPath := filepath.Join(homeDir, ".kube", "config")
	k8sRestConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	apixClient, err := apixv1beta1client.NewForConfig(k8sRestConfig)
	if err != nil {
		return nil, err
	}

	return apixClient.CustomResourceDefinitions(), nil
}
