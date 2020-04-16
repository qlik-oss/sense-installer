package preflight

import (
	"fmt"

	"github.com/kyokomi/emoji"
	ansi "github.com/mattn/go-colorable"
	"github.com/pkg/errors"
	"github.com/ttacon/chalk"
)

func (qp *QliksensePreflight) RunAllPreflightChecks(kubeConfigContents []byte, namespace string, preflightOpts *PreflightOptions) error {
	out := ansi.NewColorableStdout()
	checkCount := 0
	totalCount := 0

	// Preflight minimum kuberenetes version check
	if err := qp.CheckK8sVersion(namespace, kubeConfigContents); err != nil {
		emoji.Fprintf(out, "%s\n", chalk.Red.Color(":heavy_multiplication_x: Preflight kubernetes minimum version check"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		emoji.Fprintf(out, "%s\n\n", chalk.Green.Color(":heavy_check_mark: Preflight kubernetes minimum version check"))
		checkCount++
	}
	totalCount++

	// Preflight deployment check
	if err := qp.CheckDeployment(namespace, kubeConfigContents); err != nil {
		emoji.Fprintf(out, "%s\n", chalk.Red.Color(":heavy_multiplication_x: Preflight deployment check"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		emoji.Fprintf(out, "%s\n\n", chalk.Green.Color(":heavy_check_mark: Preflight deployment check"))
		checkCount++
	}
	totalCount++

	// Preflight service check
	if err := qp.CheckService(namespace, kubeConfigContents); err != nil {
		emoji.Fprintf(out, "%s\n", chalk.Red.Color(":heavy_multiplication_x: Preflight service check"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		emoji.Fprintf(out, "%s\n\n", chalk.Green.Color(":heavy_check_mark: Preflight service check"))
		checkCount++
	}
	totalCount++

	// Preflight pod check
	if err := qp.CheckPod(namespace, kubeConfigContents); err != nil {
		emoji.Fprintf(out, "%s\n", chalk.Red.Color(":heavy_multiplication_x: Preflight pod check"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		emoji.Fprintf(out, "%s\n\n", chalk.Green.Color(":heavy_check_mark: Preflight pod check"))
		checkCount++
	}
	totalCount++

	// Preflight role check
	if err := qp.CheckCreateRole(namespace); err != nil {
		emoji.Fprintf(out, "%s\n", chalk.Red.Color(":heavy_multiplication_x: Preflight role check"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		emoji.Fprintf(out, "%s\n\n", chalk.Green.Color(":heavy_check_mark: Preflight role check"))
		checkCount++
	}
	totalCount++

	// Preflight rolebinding check
	if err := qp.CheckCreateRoleBinding(namespace); err != nil {
		emoji.Fprintf(out, "%s\n", chalk.Red.Color(":heavy_multiplication_x: Preflight rolebinding check"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		emoji.Fprintf(out, "%s\n\n", chalk.Green.Color(":heavy_check_mark: Preflight rolebinding check"))
		checkCount++
	}
	totalCount++

	// Preflight serviceaccount check
	if err := qp.CheckCreateServiceAccount(namespace); err != nil {
		emoji.Fprintf(out, "%s\n", chalk.Red.Color(":heavy_multiplication_x: Preflight serviceaccount check"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		emoji.Fprintf(out, "%s\n\n", chalk.Green.Color(":heavy_check_mark: Preflight serviceaccount check"))
		checkCount++
	}
	totalCount++

	// Preflight mongo check
	if err := qp.CheckMongo(kubeConfigContents, namespace, preflightOpts); err != nil {
		emoji.Fprintf(out, "%s\n", chalk.Red.Color(":heavy_multiplication_x: Preflight mongo check"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		emoji.Fprintf(out, "%s\n\n", chalk.Green.Color(":heavy_check_mark: Preflight mongo check"))
		checkCount++
	}
	totalCount++

	// Preflight DNS check
	if err := qp.CheckDns(namespace, kubeConfigContents); err != nil {
		emoji.Fprintf(out, "%s\n", chalk.Red.Color(":heavy_multiplication_x: Preflight DNS check"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		emoji.Fprintf(out, "%s\n\n", chalk.Green.Color(":heavy_check_mark: Preflight DNS check"))
		checkCount++
	}
	totalCount++

	if checkCount == totalCount {
		// All preflight checks were successful
		return nil
	}
	return errors.New("1 or more preflight checks have FAILED")
}
