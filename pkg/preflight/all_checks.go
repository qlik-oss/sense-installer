package preflight

import (
	"fmt"

	"github.com/kyokomi/emoji"
	ansi "github.com/mattn/go-colorable"
	"github.com/ttacon/chalk"
)

func (qp *QliksensePreflight) RunAllPreflightChecks(kubeConfigContents []byte, namespace string, preflightOpts *PreflightOptions) {
	out := ansi.NewColorableStdout()

	checkCount := 0
	totalCount := 0
	// Preflight minimum kuberenetes version check
	// fmt.Printf("\nPreflight kubernetes minimum version check\n")
	// fmt.Println("------------------------------------------")
	if err := qp.CheckK8sVersion(namespace, kubeConfigContents); err != nil {

		emoji.Fprintf(out, "%s\n", chalk.Red.Color(":heavy_multiplication_x: Preflight kubernetes minimum version check"))
		fmt.Printf("Error: %v\n\n", err)
		// fmt.Printf("Preflight kubernetes minimum version check: FAILED\n")
	} else {
		emoji.Fprintf(out, "%s\n\n", chalk.Green.Color(":heavy_check_mark: Preflight kubernetes minimum version check"))
		checkCount++
	}
	totalCount++

	// Preflight deployment check
	fmt.Printf("\nPreflight deployment check\n")
	fmt.Println("--------------------------")
	if err := qp.CheckDeployment(namespace, kubeConfigContents); err != nil {
		fmt.Printf("Preflight deployment check: FAILED\n")
	} else {
		checkCount++
	}
	totalCount++

	// Preflight service check
	// fmt.Printf("\nPreflight service check\n")
	// fmt.Println("-----------------------")
	// if err := qp.CheckService(namespace, kubeConfigContents); err != nil {
	// 	fmt.Printf("Preflight service check: FAILED\n")
	// } else {
	// 	checkCount++
	// }
	// totalCount++

	// Preflight pod check
	// fmt.Printf("\nPreflight pod check\n")
	// fmt.Println("-----------------------")
	// if err := qp.CheckPod(namespace, kubeConfigContents); err != nil {
	// 	fmt.Printf("Preflight pod check: FAILED\n")
	// } else {
	// 	checkCount++
	// }
	// totalCount++

	// Preflight role check
	// fmt.Printf("\nPreflight role check\n")
	// fmt.Println("--------------------------")
	// if err := qp.CheckCreateRole(namespace); err != nil {
	// 	fmt.Printf("Preflight role check: FAILED\n")
	// } else {
	// 	checkCount++
	// }
	// totalCount++

	// Preflight rolebinding check
	// fmt.Printf("\nPreflight rolebinding check\n")
	// fmt.Println("---------------------------------")
	// if err := qp.CheckCreateRoleBinding(namespace); err != nil {
	// 	fmt.Printf("Preflight rolebinding check: FAILED\n")
	// } else {
	// 	checkCount++
	// }
	// totalCount++

	// Preflight serviceaccount check
	// fmt.Printf("\nPreflight serviceaccount check\n")
	// fmt.Println("------------------------------------")
	// if err := qp.CheckCreateServiceAccount(namespace); err != nil {
	// 	fmt.Printf("Preflight serviceaccount check: FAILED\n")
	// } else {
	// 	checkCount++
	// }
	// totalCount++

	// Preflight mongo check
	// fmt.Printf("\nPreflight mongo check\n")
	// fmt.Println("---------------------")
	// if err := qp.CheckMongo(kubeConfigContents, namespace, preflightOpts); err != nil {
	// 	fmt.Printf("Preflight mongo check: FAILED\n")
	// } else {
	// 	checkCount++
	// }
	// totalCount++

	// Preflight DNS check
	// fmt.Printf("\nPreflight DNS check\n")
	// fmt.Println("-------------------")
	// if err := qp.CheckDns(namespace, kubeConfigContents); err != nil {
	// 	fmt.Printf("Preflight DNS check: FAILED\n")
	// } else {
	// 	checkCount++
	// }
	// totalCount++

	// if checkCount == totalCount {
	// 	fmt.Printf("\nAll preflight checks have PASSED\n")
	// } else {
	// 	fmt.Printf("\n1 or more preflight checks have FAILED\n")
	// }
	// fmt.Println("Completed running all preflight checks")
}
