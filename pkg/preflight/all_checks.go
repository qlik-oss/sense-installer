package preflight

import (
	"fmt"
)

func (qp *QliksensePreflight) RunAllPreflightChecks(namespace string, kubeConfigContents []byte) {

	checkCount := 0
	// Preflight minimum kuberenetes version check
	fmt.Printf("\nPreflight kubernetes minimum version check\n")
	fmt.Println("------------------------------------------")
	if err := qp.CheckK8sVersion(namespace, kubeConfigContents); err != nil {
		fmt.Printf("Preflight kubernetes minimum version check: FAILED\n")
	} else {
		checkCount++
	}

	// Preflight deployment check
	fmt.Printf("\nPreflight deployment check\n")
	fmt.Println("--------------------------")
	if err := qp.CheckDeployment(namespace, kubeConfigContents); err != nil {
		fmt.Printf("Preflight deployment check: FAILED\n")
	} else {
		checkCount++
	}

	// Preflight service check
	fmt.Printf("\nPreflight service check\n")
	fmt.Println("-----------------------")
	if err := qp.CheckService(namespace, kubeConfigContents); err != nil {
		fmt.Printf("Preflight service check: FAILED\n")
	} else {
		checkCount++
	}

	// Preflight pod check
	fmt.Printf("\nPreflight pod check\n")
	fmt.Println("-----------------------")
	if err := qp.CheckPod(namespace, kubeConfigContents); err != nil {
		fmt.Printf("Preflight pod check: FAILED\n")
	} else {
		checkCount++
	}

	// Preflight role check
	fmt.Printf("\nPreflight role check\n")
	fmt.Println("--------------------------")
	if err := qp.CheckCreateRole(namespace); err != nil {
		fmt.Printf("Preflight role check: FAILED\n")
	} else {
		checkCount++
	}

	// Preflight rolebinding check
	fmt.Printf("\nPreflight rolebinding check\n")
	fmt.Println("---------------------------------")
	if err := qp.CheckCreateRoleBinding(namespace); err != nil {
		fmt.Printf("Preflight rolebinding check: FAILED\n")
	} else {
		checkCount++
	}

	// Preflight serviceaccount check
	fmt.Printf("\nPreflight serviceaccount check\n")
	fmt.Println("------------------------------------")
	if err := qp.CheckCreateServiceAccount(namespace); err != nil {
		fmt.Printf("Preflight serviceaccount check: FAILED\n")
	} else {
		checkCount++
	}

	// Preflight mongo check
	mongodbUrl := "mongodb://192.168.2.24:27017"
	fmt.Printf("\nPreflight mongo check\n")
	fmt.Println("---------------------")
	if err := qp.CheckMongo(kubeConfigContents, namespace, mongodbUrl); err != nil {
		fmt.Printf("Preflight mongo check: FAILED\n")
	} else {
		checkCount++
	}

	// Preflight DNS check
	fmt.Printf("\nPreflight DNS check\n")
	fmt.Println("-------------------")
	if err := qp.CheckDns(namespace, kubeConfigContents); err != nil {
		fmt.Printf("Preflight DNS check: FAILED\n")
	} else {
		checkCount++
	}

	if checkCount == 9 {
		fmt.Printf("\nAll preflight checks have PASSED\n")
	} else {
		fmt.Printf("\n1 or more preflight checks have FAILED\n")
	}
	fmt.Println("Completed running all preflight checks")
}
