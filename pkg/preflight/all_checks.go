package preflight

import (
	"fmt"

	. "github.com/logrusorgru/aurora"
	ansi "github.com/mattn/go-colorable"
	"github.com/pkg/errors"
)

func (qp *QliksensePreflight) RunAllPreflightChecks(kubeConfigContents []byte, namespace string, preflightOpts *PreflightOptions) error {
	checkCount := 0
	totalCount := 0

	out := ansi.NewColorableStdout()
	// Preflight minimum kuberenetes version check
	if err := qp.CheckK8sVersion(namespace, kubeConfigContents); err != nil {
		fmt.Fprintf(out, "%s\n", Red("FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight deployment check
	if err := qp.CheckDeployment(namespace, kubeConfigContents, false); err != nil {
		fmt.Fprintf(out, "%s\n", Red("FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight service check
	if err := qp.CheckService(namespace, kubeConfigContents, false); err != nil {
		fmt.Fprintf(out, "%s\n", Red("FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight pod check
	if err := qp.CheckPod(namespace, kubeConfigContents, false); err != nil {
		fmt.Fprintf(out, "%s\n", Red("FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight role check
	if err := qp.CheckCreateRole(namespace, false); err != nil {
		fmt.Fprintf(out, "%s\n", Red("FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight rolebinding check
	if err := qp.CheckCreateRoleBinding(namespace, false); err != nil {
		fmt.Fprintf(out, "%s\n", Red("FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight serviceaccount check
	if err := qp.CheckCreateServiceAccount(namespace, false); err != nil {
		fmt.Fprintf(out, "%s\n", Red("FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight mongo check
	if err := qp.CheckMongo(kubeConfigContents, namespace, preflightOpts, false); err != nil {
		fmt.Fprintf(out, "%s\n", Red("FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight DNS check
	if err := qp.CheckDns(namespace, kubeConfigContents, false); err != nil {
		fmt.Fprintf(out, "%s\n", Red("FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("PASSED"))
		checkCount++
	}
	totalCount++

	if checkCount == totalCount {
		// All preflight checks were successful
		return nil
	}
	return errors.New("1 or more preflight checks have FAILED")
}
