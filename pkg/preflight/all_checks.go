package preflight

import (
	"fmt"
)

func (qp *QliksensePreflight) RunAllPreflightChecks(namespace string, kubeConfigContents []byte) error {

	// Preflight DNS check
	fmt.Printf("\nRunning preflight DNS check...\n")
	qp.CheckDns(namespace, kubeConfigContents)

	// Preflight minimum kuberenetes version check
	fmt.Printf("\nRunning preflight kubernetes minimum version check...\n")
	qp.CheckK8sVersion(namespace, kubeConfigContents)

	fmt.Println("Completed running all preflight checks")
	return nil
}
