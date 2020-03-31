package main

import (
	"fmt"
	"log"

	"github.com/qlik-oss/sense-installer/pkg/preflight"

	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func preflightCmd(q *qliksense.Qliksense) *cobra.Command {
	var preflightCmd = &cobra.Command{
		Use:     "preflight",
		Short:   "perform preflight checks on the cluster",
		Long:    `perform preflight checks on the cluster`,
		Example: `qliksense preflight <preflight_check_to_run>`,
	}
	return preflightCmd
}

func preflightCheckDnsCmd(q *qliksense.Qliksense) *cobra.Command {
	var preflightDnsCmd = &cobra.Command{
		Use:     "dns",
		Short:   "perform preflight dns check",
		Long:    `perform preflight dns check to check DNS connectivity status in the cluster`,
		Example: `qliksense preflight dns`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q}

			// Preflight DNS check
			fmt.Printf("Preflight DNS check\n")
			fmt.Println("---------------------")
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Printf("Preflight DNS check FAILED\n")
				log.Fatal(err)
			}
			if err = qp.CheckDns(namespace, kubeConfigContents); err != nil {
				fmt.Println(err)
				fmt.Print("Preflight DNS check FAILED\n")
				log.Fatal()
			}
			return nil
		},
	}
	return preflightDnsCmd
}

func preflightCheckK8sVersionCmd(q *qliksense.Qliksense) *cobra.Command {
	var preflightCheckK8sVersionCmd = &cobra.Command{
		Use:     "k8s-version",
		Short:   "check k8s version",
		Long:    `check minimum valid k8s version on the cluster`,
		Example: `qliksense preflight k8s-version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q}

			// Preflight Kubernetes minimum version check
			fmt.Printf("Preflight kubernetes minimum version check\n")
			fmt.Println("------------------------------------------")
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Printf("Preflight kubernetes minimum version check FAILED\n")
				log.Fatal(err)
			}
			if err = qp.CheckK8sVersion(namespace, kubeConfigContents); err != nil {
				fmt.Println(err)
				fmt.Printf("Preflight kubernetes minimum version check FAILED\n")
				log.Fatal()
			}
			return nil
		},
	}
	return preflightCheckK8sVersionCmd
}

func preflightAllChecksCmd(q *qliksense.Qliksense) *cobra.Command {
	var preflightAllChecksCmd = &cobra.Command{
		Use:     "all",
		Short:   "perform all checks",
		Long:    `perform all preflight checks on the target cluster`,
		Example: `qliksense preflight all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q}

			// Preflight run all checks
			fmt.Printf("Running all preflight checks\n")
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Println(err)
				fmt.Printf("Running preflight check suite has FAILED...\n")
				log.Fatal()
			}
			qp.RunAllPreflightChecks(namespace, kubeConfigContents)
			return nil

		},
	}
	return preflightAllChecksCmd
}
