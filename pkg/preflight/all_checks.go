package preflight

import (
	"fmt"
)

func (qp *QliksensePreflight) RunAllPreflightChecks(namespace string, kubeConfigContents []byte) error {

	// Preflight DNS check
	fmt.Printf("\nPreflight DNS check\n")
	fmt.Println("-------------------")
	qp.CheckDns(namespace, kubeConfigContents)

	// Preflight minimum kuberenetes version check
	fmt.Printf("\nPreflight kubernetes minimum version check\n")
	fmt.Println("------------------------------------------")
	qp.CheckK8sVersion(namespace, kubeConfigContents)

	fmt.Println("Completed running all preflight checks")
	return nil
}
