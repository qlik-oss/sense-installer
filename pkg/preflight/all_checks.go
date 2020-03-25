package preflight

import (
	"fmt"
)

func (qp *QliksensePreflight) RunAllPreflightChecks() error {
	//run all preflight checks
	fmt.Println("Running all preflight checks")
	fmt.Printf("\nRunning DNS check...\n")
	qp.CheckDns()
	fmt.Printf("\nRunning minimum kubernetes version check...\n")
	qp.CheckK8sVersion()
	fmt.Println("Completed running all preflight checks")
	return nil
}
