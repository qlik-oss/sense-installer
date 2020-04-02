package preflight

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/api"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"k8s.io/client-go/kubernetes"
)

func (qp *QliksensePreflight) CheckCreateRole(namespace string, kubeConfigContents []byte) error {
	qConfig := qapi.NewQConfig(qp.Q.QliksenseHome)
	var currentCR *qapi.QliksenseCR
	var tempDownloadedDir string
	var resultYamlString []byte
	var err error
	qConfig.SetNamespace(namespace)
	currentCR, err = qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println(err)
		log.Fatal("Unable to retrieve current CR")
	}

	// currentCR.Spec.RotateKeys = "None"
	// currentCR.SetNamespace(namespace)
	// fmt.Println("Manifests root: " + currentCR.Spec.GetManifestsRoot())
	// fmt.Println("Namespace: " + currentCR.GetNamespace())
	// // generate patches
	// cr.GeneratePatches(&currentCR.KApiCr, path.Join(qp.Q.QliksenseHome, ".kube", "config"))
	if tempDownloadedDir, err = qliksense.DownloadFromGitRepoToTmpDir(qliksense.QLIK_GIT_REPO, "master"); err != nil {
		fmt.Println("Unable to Download from git repo to tmp dir", err)
		return err
	}
	fmt.Printf("qliksense.QLIK_GIT_REPO: %s\n", qliksense.QLIK_GIT_REPO)
	fmt.Printf("currentCR.Spec.Profile: %s\n", currentCR.Spec.Profile)
	if currentCR.Spec.Profile == "" {
		resultYamlString, err = qliksense.ExecuteKustomizeBuild(filepath.Join(tempDownloadedDir, "manifests", "docker-desktop"))
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			log.Fatal("Unable to retrieve manifests from executing kustomize")
		}

	} else {
		resultYamlString, err = qliksense.ExecuteKustomizeBuild(filepath.Join(tempDownloadedDir, currentCR.Spec.GetManifestsRoot(), currentCR.Spec.GetProfileDir()))
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			log.Fatal("Unable to retrieve manifests from executing kustomize")
		}
	}
	fmt.Printf("Tmp downloaded dir: %s\n", filepath.Join(tempDownloadedDir, "manifests", "docker-desktop"))
	// fmt.Printf("resultYaml String: %s\n", string(resultYamlString))

	// replace namespace with my namespace in the resulting yaml file

	sa := qliksense.GetYamlsFromMultiDoc(string(resultYamlString), "Role")
	// fmt.Println("generate yaml 1:", sa)
	sa = strings.ReplaceAll(sa, "namespace: default\n", fmt.Sprintf("namespace: %s\n", namespace))
	fmt.Println("generate yaml 2:", sa)
	// os.Exit(1)
	// fmt.Printf("SA: %s\n", sa)
	err = api.KubectlApply(sa, namespace)
	if err != nil {
		fmt.Println("Preflight create-role check: FAILED")
		return err
	}
	fmt.Println("Preflight create-role check: PASSED")
	fmt.Println("cleaning up resources")
	err = api.KubectlDelete(sa, namespace)
	if err != nil {
		fmt.Println("Preflight cleanup: FAILED")
		return err
	}
	fmt.Println("Completed preflight createRole check")

	return nil
}

func checkPfRole(clientset *kubernetes.Clientset, namespace, roleName string) error {
	// check if we are able to create a role
	pfRole, err := createPfRole(clientset, namespace, roleName)
	if err != nil {
		fmt.Println("Preflight create-role check: FAILED")
		return err
	}
	defer deleteRole(clientset, namespace, pfRole)

	fmt.Println("Preflight create-role check: PASSED")
	fmt.Println("Cleaning up resources...")

	return nil
}

func (qp *QliksensePreflight) CheckCreateRoleBinding(namespace string, kubeConfigContents []byte) error {
	clientset, _, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("Kube config error: %v\n", err)
		fmt.Print(err)
		return err
	}

	// create a roleBinding
	fmt.Printf("Preflight create RoleBinding check: \n")
	err = checkPfRoleBinding(clientset, namespace, "role-binding-preflight-check")
	if err != nil {
		fmt.Println("Preflight create RoleBinding check: FAILED")
		return err
	}
	fmt.Println("Completed preflight create RoleBinding check")
	return nil
}

func checkPfRoleBinding(clientset *kubernetes.Clientset, namespace, roleBindingName string) error {
	// check if we are able to create a role binding
	pfRoleBinding, err := createPfRoleBinding(clientset, namespace, roleBindingName)
	if err != nil {
		fmt.Println("Preflight create RoleBinding check: FAILED")
		return err
	}
	defer deleteRoleBinding(clientset, namespace, pfRoleBinding)

	fmt.Println("Preflight create RoleBinding check: PASSED")
	fmt.Println("Cleaning up resources...")

	return nil
}

func (qp *QliksensePreflight) CheckCreateServiceAccount(namespace string, kubeConfigContents []byte) error {
	clientset, _, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("Kube config error: %v\n", err)
		fmt.Print(err)
		return err
	}

	// create a service account
	fmt.Printf("Preflight createServiceAccount check: \n")
	err = checkPfServiceAccount(clientset, namespace, "service-account-preflight-check")
	if err != nil {
		fmt.Println("Preflight createServiceAccount check: FAILED")
		return err
	}
	fmt.Println("Completed preflight createServiceAccount check")
	return nil
}

func checkPfServiceAccount(clientset *kubernetes.Clientset, namespace, serviceAccountName string) error {
	// check if we are able to create a service account
	pfRole, err := createPfServiceAccount(clientset, namespace, serviceAccountName)
	if err != nil {
		fmt.Println("Preflight createServiceAccount check: FAILED")
		return err
	}
	defer deleteServiceAccount(clientset, namespace, pfRole)

	fmt.Println("Preflight createServiceAccount check: PASSED")
	fmt.Println("Cleaning up resources...")

	return nil
}

func (qp *QliksensePreflight) CheckCreateRB(namespace string, kubeConfigContents []byte) error {
	clientset, _, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("Kube config error: %v\n", err)
		fmt.Print(err)
		return err
	}

	// create a role
	fmt.Printf("Preflight create-role check: \n")
	err = checkPfRole(clientset, namespace, "role-preflight-check")
	if err != nil {
		fmt.Println("Preflight create-role check: FAILED")
		return err
	}
	fmt.Printf("Completed preflight create-role check\n\n")

	// create a roleBinding
	fmt.Printf("Preflight create RoleBinding check: \n")
	err = checkPfRoleBinding(clientset, namespace, "role-binding-preflight-check")
	if err != nil {
		fmt.Println("Preflight create RoleBinding check: FAILED")
		return err
	}
	fmt.Printf("Completed preflight create RoleBinding check\n\n")

	// create a service account
	fmt.Printf("Preflight createServiceAccount check: \n")
	err = checkPfServiceAccount(clientset, namespace, "service-account-preflight-check")
	if err != nil {
		fmt.Println("Preflight createServiceAccount check: FAILED")
		return err
	}
	fmt.Println("Completed preflight createServiceAccount check")

	fmt.Println("Completed preflight CreateRB check")
	return nil
}
