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
	fmt.Printf("Preflight createRole check: \n")
	err := qp.checkCreateEntity(namespace, "Role")
	if err != nil {
		// fmt.Println("Preflight createRole check: FAILED")
		return err
	}
	fmt.Println("Completed preflight createRole check")
	return nil
}

func (qp *QliksensePreflight) CheckCreateRoleBinding(namespace string) error {
	// create a RoleBinding
	fmt.Printf("Preflight createRoleBinding check: \n")
	err := qp.checkCreateEntity(namespace, "RoleBinding")
	if err != nil {
		// fmt.Println("Preflight createRoleBinding check: FAILED")
		return err
	}
	fmt.Println("Completed preflight createRoleBinding check")
	return nil
}

func (qp *QliksensePreflight) CheckCreateServiceAccount(namespace string) error {
	// create a service account
	fmt.Printf("Preflight createServiceAccount check: \n")
	err := qp.checkCreateEntity(namespace, "ServiceAccount")
	if err != nil {
		// fmt.Println("Preflight createServiceAccount check: FAILED")
		return err
	}
	fmt.Println("Completed preflight createServiceAccount check")
	return nil
}
func (qp *QliksensePreflight) checkCreateEntity(namespace, entityToTest string) error {
	qConfig := qapi.NewQConfig(qp.Q.QliksenseHome)
	var currentCR *qapi.QliksenseCR
	var tempDownloadedDir string
	var resultYamlString []byte
	var err error
	qConfig.SetNamespace(namespace)
	currentCR, err = qConfig.GetCurrentCR()
	if err != nil {
		fmt.Printf("Unable to retrieve current CR: %v\n", err)
		return err
	}

	if tempDownloadedDir, err = qliksense.DownloadFromGitRepoToTmpDir(qliksense.QLIK_GIT_REPO, "master"); err != nil {
		fmt.Printf("Unable to Download from git repo to tmp dir: %v\n", err)
		return err
	}

	if currentCR.Spec.Profile == "" {
		resultYamlString, err = qliksense.ExecuteKustomizeBuild(filepath.Join(tempDownloadedDir, "manifests", "docker-desktop"))
		if err != nil {
			fmt.Printf("Unable to retrieve manifests from executing kustomize: %v\n", err)
			return err
		}

	} else {
		resultYamlString, err = qliksense.ExecuteKustomizeBuild(filepath.Join(tempDownloadedDir, currentCR.Spec.GetManifestsRoot(), currentCR.Spec.GetProfileDir()))
		if err != nil {
			fmt.Printf("Unable to retrieve manifests from executing kustomize: %v\n", err)
			return err
		}
	}

	sa := qliksense.GetYamlsFromMultiDoc(string(resultYamlString), entityToTest)
	if sa != "" {
		sa = strings.ReplaceAll(sa, "namespace: default\n", fmt.Sprintf("namespace: %s\n", namespace))
	} else {
		err := fmt.Errorf("Unable to retrieve yamls to apply on cluster\n")
		fmt.Println(err)
		return err
	}

	err = api.KubectlApply(sa, namespace)
	if err != nil {
		fmt.Printf("Failed to create entity on the cluster: ", err)
		return err
	}

	fmt.Printf("Preflight create%s check: PASSED\n", entityToTest)
	fmt.Println("Cleaning up resources")
	err = api.KubectlDelete(sa, namespace)
	if err != nil {
		fmt.Println("Preflight cleanup failed!")
		return err
	}
	return nil
}

func (qp *QliksensePreflight) CheckCreateRB(namespace string, kubeConfigContents []byte) error {

	// create a role
	fmt.Printf("Preflight createRole check: \n")
	err := qp.checkCreateEntity(namespace, "Role")
	if err != nil {
		fmt.Println("Preflight createRole check: FAILED")
	}
	fmt.Printf("Completed preflight create-role check\n\n")

	// create a roleBinding
	fmt.Printf("Preflight createRoleBinding check: \n")
	err = qp.checkCreateEntity(namespace, "RoleBinding")
	if err != nil {
		fmt.Println("Preflight createRoleBinding check: FAILED")
	}
	fmt.Printf("Completed preflight createRoleBinding check\n\n")

	// create a service account
	fmt.Printf("Preflight createServiceAccount check: \n")
	err = qp.checkCreateEntity(namespace, "ServiceAccount")
	if err != nil {
		fmt.Println("Preflight createServiceAccount check: FAILED")
	}
	fmt.Printf("Completed preflight createServiceAccount check\n\n")

	fmt.Println("Preflight CreateRB check: PASSED")
	fmt.Println("Completed preflight CreateRB check")
	return nil
}
