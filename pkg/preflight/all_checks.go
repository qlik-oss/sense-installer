package preflight

import (
	"fmt"
)

func (qp *QliksensePreflight) RunAllPreflightChecks(namespace string, kubeConfigContents []byte) {

	checkCount := 0
	// Preflight DNS check
	fmt.Printf("\nPreflight DNS check\n")
	fmt.Println("-------------------")
	if err := qp.CheckDns(namespace, kubeConfigContents); err != nil {
		fmt.Printf("Preflight DNS check: FAILED\n")
	} else {
		checkCount++
	}

	// Preflight minimum kuberenetes version check
	fmt.Printf("\nPreflight kubernetes minimum version check\n")
	fmt.Println("------------------------------------------")
	if err := qp.CheckK8sVersion(namespace, kubeConfigContents); err != nil {
		fmt.Printf("Preflight kubernetes minimum version check: FAILED\n")
	} else {
		checkCount++
	}

	// Preflight minimum kuberenetes version check
	fmt.Printf("\nPreflight deploy check\n")
	fmt.Println("-----------------------")
	if err := qp.CheckDeploy(namespace, kubeConfigContents); err != nil {
		fmt.Printf("Preflight deploy check: FAILED\n")
	} else {
		checkCount++
	}

	if checkCount == 3 {
		fmt.Printf("All preflight checks have PASSED\n")
	} else {
		fmt.Printf("1 or more preflight checks have FAILED\n")
	}
	fmt.Println("Completed running all preflight checks")
}
