package preflight

import (
	"fmt"

	ansi "github.com/mattn/go-colorable"
	"github.com/pkg/errors"
	. "github.com/logrusorgru/aurora"
)

func (qp *QliksensePreflight) RunAllPreflightChecks(kubeConfigContents []byte, namespace string, preflightOpts *PreflightOptions) error {
	checkCount := 0
	totalCount := 0

	out := ansi.NewColorableStdout()
	// Preflight minimum kuberenetes version check
	if err := qp.CheckK8sVersion(namespace, kubeConfigContents); err != nil {
		fmt.Fprintf(out, "%s\n", Red("Preflight kubernetes minimum version check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("Preflight kubernetes minimum version check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight deployment check
	if err := qp.CheckDeployment(namespace, kubeConfigContents); err != nil {
		fmt.Fprintf(out, "%s\n", Red("Preflight deployment check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("Preflight deployment check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight service check
	if err := qp.CheckService(namespace, kubeConfigContents); err != nil {
		fmt.Fprintf(out, "%s\n", Red("Preflight service check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("Preflight service check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight pod check
	if err := qp.CheckPod(namespace, kubeConfigContents); err != nil {
		fmt.Fprintf(out, "%s\n", Red("Preflight pod check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("Preflight pod check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight role check
	if err := qp.CheckCreateRole(namespace); err != nil {
		fmt.Fprintf(out, "%s\n", Red("Preflight role check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("Preflight role check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight rolebinding check
	if err := qp.CheckCreateRoleBinding(namespace); err != nil {
		fmt.Fprintf(out, "%s\n", Red(" Preflight rolebinding check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("Preflight rolebinding check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight serviceaccount check
	if err := qp.CheckCreateServiceAccount(namespace); err != nil {
		fmt.Fprintf(out, "%s\n", Red(" Preflight serviceaccount check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("Preflight serviceaccount check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight mongo check
	if err := qp.CheckMongo(kubeConfigContents, namespace, preflightOpts); err != nil {
		fmt.Fprintf(out, "%s\n", Red(" Preflight mongo check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("Preflight mongo check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight DNS check
	if err := qp.CheckDns(namespace, kubeConfigContents); err != nil {
		fmt.Fprintf(out, "%s\n", Red(" Preflight DNS check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("Preflight DNS check PASSED"))
		checkCount++
	}
	totalCount++

	if checkCount == totalCount {
		// All preflight checks were successful
		return nil
	}
	return errors.New("1 or more preflight checks have FAILED")
}
