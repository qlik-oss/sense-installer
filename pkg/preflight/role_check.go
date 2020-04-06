package preflight

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/api"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
)

func (qp *QliksensePreflight) CheckCreateRole(namespace string) error {
	// create a Role
	fmt.Printf("Preflight role check: \n")
	err := qp.checkCreateEntity(namespace, "Role")
	if err != nil {
		return err
	}
	fmt.Println("Completed preflight role check")
	return nil
}

func (qp *QliksensePreflight) CheckCreateRoleBinding(namespace string) error {
	// create a RoleBinding
	fmt.Printf("Preflight rolebinding check: \n")
	err := qp.checkCreateEntity(namespace, "RoleBinding")
	if err != nil {
		return err
	}
	fmt.Println("Completed preflight rolebinding check")
	return nil
}

func (qp *QliksensePreflight) CheckCreateServiceAccount(namespace string) error {
	// create a service account
	fmt.Printf("Preflight serviceaccount check: \n")
	err := qp.checkCreateEntity(namespace, "ServiceAccount")
	if err != nil {
		return err
	}
	fmt.Println("Completed preflight serviceaccount check")
	return nil
}
func (qp *QliksensePreflight) checkCreateEntity(namespace, entityToTest string) error {
	qConfig := qapi.NewQConfig(qp.Q.QliksenseHome)
	var currentCR *qapi.QliksenseCR
	mfroot := ""
	kusDir := ""
	var err error
	currentCR, err = qConfig.GetCurrentCR()
	if err != nil {
		fmt.Printf("Unable to retrieve current CR: %v\n", err)
		return err
	}
	if currentCR.IsRepoExist() {
		mfroot = currentCR.Spec.GetManifestsRoot()
	} else if tempDownloadedDir, err := qliksense.DownloadFromGitRepoToTmpDir(qliksense.QLIK_GIT_REPO, "master"); err != nil {
		fmt.Printf("Unable to Download from git repo to tmp dir: %v\n", err)
		return err
	} else {
		mfroot = tempDownloadedDir
	}

	if currentCR.Spec.Profile == "" {
		kusDir = filepath.Join(mfroot, "manifests", "docker-desktop")
	} else {
		kusDir = filepath.Join(mfroot, "manifests", currentCR.Spec.Profile)
	}
	resultYamlString, err := qliksense.ExecuteKustomizeBuild(kusDir)
	if err != nil {
		fmt.Printf("Unable to retrieve manifests from executing kustomize: %v\n", err)
		return err
	}

	sa := qliksense.GetYamlsFromMultiDoc(string(resultYamlString), entityToTest)
	if sa != "" {
		sa = strings.ReplaceAll(sa, "namespace: default\n", fmt.Sprintf("namespace: %s\n", namespace))
	} else {
		err := fmt.Errorf("Unable to retrieve yamls to apply on cluster")
		fmt.Println(err)
		return err
	}

	defer func() {
		fmt.Println("Cleaning up resources")
		api.KubectlDelete(sa, namespace)
		if err != nil {
			fmt.Println("Preflight cleanup failed!")
		}
	}()

	err = api.KubectlApply(sa, namespace)
	if err != nil {
		err := fmt.Errorf("Failed to create entity on the cluster: %v", err)
		fmt.Println(err)
		return err
	}

	fmt.Printf("Preflight %s check: PASSED\n", entityToTest)
	return nil
}

func (qp *QliksensePreflight) CheckCreateRB(namespace string, kubeConfigContents []byte) error {

	// create a role
	fmt.Printf("Preflight createRole check: \n")
	err := qp.checkCreateEntity(namespace, "Role")
	if err != nil {
		fmt.Println("Preflight role check: FAILED")
	}
	fmt.Printf("Completed preflight role check\n\n")

	// create a roleBinding
	fmt.Printf("Preflight rolebinding check: \n")
	err = qp.checkCreateEntity(namespace, "RoleBinding")
	if err != nil {
		fmt.Println("Preflight rolebinding check: FAILED")
	}
	fmt.Printf("Completed preflight rolebinding check\n\n")

	// create a service account
	fmt.Printf("Preflight serviceaccount check: \n")
	err = qp.checkCreateEntity(namespace, "ServiceAccount")
	if err != nil {
		fmt.Println("Preflight serviceaccount check: FAILED")
	}
	fmt.Printf("Completed preflight serviceaccount check\n\n")

	fmt.Println("Preflight RB check: PASSED")
	fmt.Println("Completed preflight CreateRB check")
	return nil
}
