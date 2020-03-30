package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/qlik-oss/sense-installer/pkg/api"
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
			fmt.Printf("Running preflight DNS check...\n")
			namespace, kubeConfigContents, err := initPreflight()
			if err != nil {
				log.Fatal(err)
			}
			return qp.CheckDns(namespace, kubeConfigContents)
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
			fmt.Printf("Running preflight kubernetes minimum version check...\n")
			namespace, kubeConfigContents, err := initPreflight()
			if err != nil {
				log.Fatal(err)
			}
			return qp.CheckK8sVersion(namespace, kubeConfigContents)
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
			fmt.Printf("Running all preflight checks...\n")
			namespace, kubeConfigContents, err := initPreflight()
			if err != nil {
				log.Fatal(err)
			}
			return qp.RunAllPreflightChecks(namespace, kubeConfigContents)

		},
	}
	return preflightAllChecksCmd
}

func initPreflight() (string, []byte, error) {
	api.LogDebugMessage("Reading .kube/config file...")

	homeDir, err := homedir.Dir()
	if err != nil {
		err = fmt.Errorf("Unable to deduce home dir\n")
		return "", nil, err
	}
	api.LogDebugMessage("Kube config location: %s\n\n", filepath.Join(homeDir, ".kube", "config"))

	kubeConfig := filepath.Join(homeDir, ".kube", "config")
	kubeConfigContents, err := ioutil.ReadFile(kubeConfig)
	if err != nil {
		err = fmt.Errorf("Unable to deduce home dir\n")
		return "", nil, err
	}
	// retrieve namespace
	namespace := api.GetKubectlNamespace()
	api.LogDebugMessage("Namespace: %s\n", namespace)
	return namespace, kubeConfigContents, nil
}
