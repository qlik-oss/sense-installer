package preflight

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/ttacon/chalk"
)

func (qp *QliksensePreflight) RunAllPreflightChecks(kubeConfigContents []byte, namespace string, preflightOpts *PreflightOptions) error {
	checkCount := 0
	totalCount := 0

	// Preflight minimum kuberenetes version check
	if err := qp.CheckK8sVersion(namespace, kubeConfigContents); err != nil {
		fmt.Printf("%s\n", chalk.Red.Color("Preflight kubernetes minimum version check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Printf("%s\n\n", chalk.Green.Color("Preflight kubernetes minimum version check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight deployment check
	if err := qp.CheckDeployment(namespace, kubeConfigContents); err != nil {
		fmt.Printf("%s\n", chalk.Red.Color("Preflight deployment check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Printf("%s\n\n", chalk.Green.Color("Preflight deployment check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight service check
	if err := qp.CheckService(namespace, kubeConfigContents); err != nil {
		fmt.Printf("%s\n", chalk.Red.Color("Preflight service check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Printf("%s\n\n", chalk.Green.Color("Preflight service check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight pod check
	if err := qp.CheckPod(namespace, kubeConfigContents); err != nil {
		fmt.Printf("%s\n", chalk.Red.Color("Preflight pod check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Printf("%s\n\n", chalk.Green.Color("Preflight pod check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight role check
	if err := qp.CheckCreateRole(namespace); err != nil {
		fmt.Printf("%s\n", chalk.Red.Color("Preflight role check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Printf("%s\n\n", chalk.Green.Color("Preflight role check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight rolebinding check
	if err := qp.CheckCreateRoleBinding(namespace); err != nil {
		fmt.Printf("%s\n", chalk.Red.Color(" Preflight rolebinding check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Printf("%s\n\n", chalk.Green.Color("Preflight rolebinding check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight serviceaccount check
	if err := qp.CheckCreateServiceAccount(namespace); err != nil {
		fmt.Printf("%s\n", chalk.Red.Color(" Preflight serviceaccount check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Printf("%s\n\n", chalk.Green.Color("Preflight serviceaccount check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight mongo check
	if err := qp.CheckMongo(kubeConfigContents, namespace, preflightOpts); err != nil {
		fmt.Printf("%s\n", chalk.Red.Color(" Preflight mongo check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Printf("%s\n\n", chalk.Green.Color("Preflight mongo check PASSED"))
		checkCount++
	}
	totalCount++

	// Preflight DNS check
	if err := qp.CheckDns(namespace, kubeConfigContents); err != nil {
		fmt.Printf("%s\n", chalk.Red.Color(" Preflight DNS check FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Printf("%s\n\n", chalk.Green.Color("Preflight DNS check PASSED"))
		checkCount++
	}
	totalCount++

	if checkCount == totalCount {
		// All preflight checks were successful
		return nil
	}
	return errors.New("1 or more preflight checks have FAILED")
}
