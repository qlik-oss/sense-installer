package preflight

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
)

func (qp *QliksensePreflight) CheckCreateRole(namespace string, kubeConfigContents []byte) error {
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
	fmt.Println("Completed preflight create-role check")

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
