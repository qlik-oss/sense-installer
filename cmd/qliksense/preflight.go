package main

import (
	"fmt"

	ansi "github.com/mattn/go-colorable"
	"github.com/qlik-oss/sense-installer/pkg/preflight"
	"github.com/ttacon/chalk"

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

func pfDnsCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	out := ansi.NewColorableStdout()
	preflightOpts := &preflight.PreflightOptions{
		MongoOptions: &preflight.MongoOptions{},
	}
	var preflightDnsCmd = &cobra.Command{
		Use:     "dns",
		Short:   "perform preflight dns check",
		Long:    `perform preflight dns check to check DNS connectivity status in the cluster`,
		Example: `qliksense preflight dns`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q, P: preflightOpts}

			// Preflight DNS check
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight DNS check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			if namespace == "" {
				namespace = "default"
			}
			if err = qp.CheckDns(namespace, kubeConfigContents); err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight DNS check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			fmt.Fprintf(out, "%s\n", chalk.Green.Color("Preflight DNS check PASSED"))
			return nil
		},
	}
	f := preflightDnsCmd.Flags()
	f.BoolVarP(&preflightOpts.Verbose, "verbose", "v", false, "verbose mode")
	return preflightDnsCmd
}

func pfK8sVersionCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	out := ansi.NewColorableStdout()
	preflightOpts := &preflight.PreflightOptions{
		MongoOptions: &preflight.MongoOptions{},
	}

	var preflightCheckK8sVersionCmd = &cobra.Command{
		Use:     "kube-version",
		Short:   "check kubernetes version",
		Long:    `check minimum valid kubernetes version on the cluster`,
		Example: `qliksense preflight kube-version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q, P: preflightOpts}

			// Preflight Kubernetes minimum version check
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight kubernetes minimum version check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			if err = qp.CheckK8sVersion(namespace, kubeConfigContents); err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight kubernetes minimum version check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			fmt.Fprintf(out, "%s\n", chalk.Green.Color("Preflight kubernetes minimum version check PASSED"))
			return nil
		},
	}
	f := preflightCheckK8sVersionCmd.Flags()
	f.BoolVarP(&preflightOpts.Verbose, "verbose", "v", false, "verbose mode")

	return preflightCheckK8sVersionCmd
}

func pfAllChecksCmd(q *qliksense.Qliksense) *cobra.Command {
	out := ansi.NewColorableStdout()
	preflightOpts := &preflight.PreflightOptions{
		MongoOptions: &preflight.MongoOptions{},
	}

	var preflightAllChecksCmd = &cobra.Command{
		Use:     "all",
		Short:   "perform all checks",
		Long:    `perform all preflight checks on the target cluster`,
		Example: `qliksense preflight all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q, P: preflightOpts}

			// Preflight run all checks
			fmt.Printf("Running all preflight checks...\n\n")
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Unable to run the preflight checks suite"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			if namespace == "" {
				namespace = "default"
			}
			if err = qp.RunAllPreflightChecks(kubeConfigContents, namespace, preflightOpts); err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("1 or more preflight checks have FAILED"))
				fmt.Println("Completed running all preflight checks")
				return nil
			}
			fmt.Fprintf(out, "%s\n\n", chalk.Green.Color("All preflight checks have PASSED"))
			return nil
		},
	}
	f := preflightAllChecksCmd.Flags()
	f.BoolVarP(&preflightOpts.Verbose, "verbose", "v", false, "verbose mode")
	f.StringVarP(&preflightOpts.MongoOptions.MongodbUrl, "mongodb-url", "", "", "mongodbUrl to try connecting to")
	f.StringVarP(&preflightOpts.MongoOptions.Username, "mongodb-username", "", "", "username to connect to mongodb")
	f.StringVarP(&preflightOpts.MongoOptions.Password, "mongodb-password", "", "", "password to connect to mongodb")
	f.StringVarP(&preflightOpts.MongoOptions.CaCertFile, "mongodb-ca-cert", "", "", "certificate to use for mongodb check")
	f.StringVarP(&preflightOpts.MongoOptions.ClientCertFile, "mongodb-client-cert", "", "", "client-certificate to use for mongodb check")
	f.BoolVar(&preflightOpts.MongoOptions.Tls, "mongodb-tls", false, "enable tls?")

	return preflightAllChecksCmd
}

func pfDeploymentCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	out := ansi.NewColorableStdout()
	preflightOpts := &preflight.PreflightOptions{
		MongoOptions: &preflight.MongoOptions{},
	}
	var pfDeploymentCheckCmd = &cobra.Command{
		Use:     "deployment",
		Short:   "perform preflight deploymwnt check",
		Long:    `perform preflight deployment check to ensure that we can create deployments in the cluster`,
		Example: `qliksense preflight deployment`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q, P: preflightOpts}

			// Preflight deployments check
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight deployment check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			if namespace == "" {
				namespace = "default"
			}
			if err = qp.CheckDeployment(namespace, kubeConfigContents); err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight deployment check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			fmt.Fprintf(out, "%s\n", chalk.Green.Color("Preflight deployment check PASSED"))
			return nil
		},
	}
	f := pfDeploymentCheckCmd.Flags()
	f.BoolVarP(&preflightOpts.Verbose, "verbose", "v", false, "verbose mode")
	return pfDeploymentCheckCmd
}

func pfServiceCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	out := ansi.NewColorableStdout()
	preflightOpts := &preflight.PreflightOptions{
		MongoOptions: &preflight.MongoOptions{},
	}

	var pfServiceCheckCmd = &cobra.Command{
		Use:     "service",
		Short:   "perform preflight service check",
		Long:    `perform preflight service check to ensure that we are able to create services in the cluster`,
		Example: `qliksense preflight service`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q, P: preflightOpts}

			// Preflight service check
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight service check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}

			if namespace == "" {
				namespace = "default"
			}
			if err = qp.CheckService(namespace, kubeConfigContents); err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight service check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			fmt.Fprintf(out, "%s\n", chalk.Green.Color("Preflight service check PASSED"))
			return nil
		},
	}
	f := pfServiceCheckCmd.Flags()
	f.BoolVarP(&preflightOpts.Verbose, "verbose", "v", false, "verbose mode")
	return pfServiceCheckCmd
}

func pfPodCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	out := ansi.NewColorableStdout()
	preflightOpts := &preflight.PreflightOptions{
		MongoOptions: &preflight.MongoOptions{},
	}

	var pfPodCheckCmd = &cobra.Command{
		Use:     "pod",
		Short:   "perform preflight pod check",
		Long:    `perform preflight pod check to ensure we can create pods in the cluster`,
		Example: `qliksense preflight pod`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q, P: preflightOpts}

			// Preflight pod check
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight pod check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			if namespace == "" {
				namespace = "default"
			}
			if err = qp.CheckPod(namespace, kubeConfigContents); err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight pod check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			fmt.Fprintf(out, "%s\n", chalk.Green.Color("Preflight pod check PASSED"))
			return nil
		},
	}
	f := pfPodCheckCmd.Flags()
	f.BoolVarP(&preflightOpts.Verbose, "verbose", "v", false, "verbose mode")
	return pfPodCheckCmd
}

func pfCreateRoleCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	out := ansi.NewColorableStdout()
	preflightOpts := &preflight.PreflightOptions{
		MongoOptions: &preflight.MongoOptions{},
	}

	var preflightRoleCmd = &cobra.Command{
		Use:     "role",
		Short:   "preflight create role check",
		Long:    `perform preflight role check to ensure we are able to create a role in the cluster`,
		Example: `qliksense preflight createRole`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q, P: preflightOpts}

			// Preflight role check
			namespace, _, err := preflight.InitPreflight()
			if err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight role check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			if err = qp.CheckCreateRole(namespace); err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight role check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			fmt.Fprintf(out, "%s\n", chalk.Green.Color("Preflight role check PASSED"))
			return nil
		},
	}
	f := preflightRoleCmd.Flags()
	f.BoolVarP(&preflightOpts.Verbose, "verbose", "v", false, "verbose mode")
	return preflightRoleCmd
}

func pfCreateRoleBindingCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	out := ansi.NewColorableStdout()
	preflightOpts := &preflight.PreflightOptions{
		MongoOptions: &preflight.MongoOptions{},
	}

	var preflightRoleBindingCmd = &cobra.Command{
		Use:     "rolebinding",
		Short:   "preflight create rolebinding check",
		Long:    `perform preflight rolebinding check to ensure we are able to create a rolebinding in the cluster`,
		Example: `qliksense preflight rolebinding`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q, P: preflightOpts}

			// Preflight createRoleBinding check
			namespace, _, err := preflight.InitPreflight()
			if err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight rolebinding check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			if err = qp.CheckCreateRoleBinding(namespace); err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight rolebinding check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			fmt.Fprintf(out, "%s\n", chalk.Green.Color("Preflight rolebinding check PASSED"))
			return nil
		},
	}
	f := preflightRoleBindingCmd.Flags()
	f.BoolVarP(&preflightOpts.Verbose, "verbose", "v", false, "verbose mode")
	return preflightRoleBindingCmd
}

func pfCreateServiceAccountCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	out := ansi.NewColorableStdout()
	preflightOpts := &preflight.PreflightOptions{
		MongoOptions: &preflight.MongoOptions{},
	}

	var preflightServiceAccountCmd = &cobra.Command{
		Use:     "serviceaccount",
		Short:   "preflight create ServiceAccount check",
		Long:    `perform preflight serviceaccount check to ensure we are able to create a service account in the cluster`,
		Example: `qliksense preflight serviceaccount`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q, P: preflightOpts}

			// Preflight createServiceAccount check
			namespace, _, err := preflight.InitPreflight()
			if err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight ServiceAccount check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			if err = qp.CheckCreateServiceAccount(namespace); err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight ServiceAccount check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			fmt.Fprintf(out, "%s\n", chalk.Green.Color("Preflight ServiceAccount check PASSED"))
			return nil
		},
	}
	f := preflightServiceAccountCmd.Flags()
	f.BoolVarP(&preflightOpts.Verbose, "verbose", "v", false, "verbose mode")
	return preflightServiceAccountCmd
}

func pfCreateAuthCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	out := ansi.NewColorableStdout()
	preflightOpts := &preflight.PreflightOptions{
		MongoOptions: &preflight.MongoOptions{},
	}
	var preflightCreateAuthCmd = &cobra.Command{
		Use:     "authcheck",
		Short:   "preflight authcheck",
		Long:    `perform preflight authcheck that combines the role, rolebinding and serviceaccount checks`,
		Example: `qliksense preflight authcheck`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q, P: preflightOpts}

			// Preflight authcheck
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight authcheck FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			if err = qp.CheckCreateRB(namespace, kubeConfigContents); err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight authcheck FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			fmt.Fprintf(out, "%s\n", chalk.Green.Color("Preflight authcheck PASSED"))
			return nil
		},
	}
	f := preflightCreateAuthCmd.Flags()
	f.BoolVarP(&preflightOpts.Verbose, "verbose", "v", false, "verbose mode")
	return preflightCreateAuthCmd
}

func pfMongoCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	out := ansi.NewColorableStdout()
	preflightOpts := &preflight.PreflightOptions{
		MongoOptions: &preflight.MongoOptions{},
	}

	var preflightMongoCmd = &cobra.Command{
		Use:     "mongo",
		Short:   "preflight mongo OR preflight mongo --url=<url>",
		Long:    `perform preflight mongo check to ensure we are able to connect to a mongodb instance in the cluster`,
		Example: `qliksense preflight mongo OR preflight mongo --url=<url>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q, P: preflightOpts}

			// Preflight mongo check
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight mongo check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			if namespace == "" {
				namespace = "default"
			}
			if err = qp.CheckMongo(kubeConfigContents, namespace, preflightOpts); err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight mongo check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			fmt.Fprintf(out, "%s\n", chalk.Green.Color("Preflight mongo check PASSED"))
			return nil
		},
	}
	f := preflightMongoCmd.Flags()
	f.BoolVarP(&preflightOpts.Verbose, "verbose", "v", false, "verbose mode")
	f.StringVarP(&preflightOpts.MongoOptions.MongodbUrl, "url", "", "", "mongodbUrl to try connecting to")
	f.StringVarP(&preflightOpts.MongoOptions.Username, "username", "", "", "username to connect to mongodb")
	f.StringVarP(&preflightOpts.MongoOptions.Password, "password", "", "", "password to connect to mongodb")
	f.StringVarP(&preflightOpts.MongoOptions.CaCertFile, "ca-cert", "", "", "ca certificate to use for mongodb check")
	f.StringVarP(&preflightOpts.MongoOptions.ClientCertFile, "client-cert", "", "", "client-certificate to use for mongodb check")
	f.BoolVar(&preflightOpts.MongoOptions.Tls, "tls", false, "enable tls?")
	return preflightMongoCmd
}

func pfCleanupCmd(q *qliksense.Qliksense) *cobra.Command {
	out := ansi.NewColorableStdout()
	preflightOpts := &preflight.PreflightOptions{
		MongoOptions: &preflight.MongoOptions{},
	}

	var pfCleanCmd = &cobra.Command{
		Use:     "clean",
		Short:   "perform preflight clean",
		Long:    `perform preflight clean to ensure that all resources are cleared up in the cluster`,
		Example: `qliksense preflight clean`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q, P: preflightOpts}

			// Preflight clean
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight cleanup FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}

			if namespace == "" {
				namespace = "default"
			}
			if err = qp.Cleanup(namespace, kubeConfigContents); err != nil {
				fmt.Fprintf(out, "%s\n", chalk.Red.Color("Preflight cleanup FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			fmt.Fprintf(out, "%s\n", chalk.Green.Color("Preflight cleanup complete"))
			return nil
		},
	}
	f := pfCleanCmd.Flags()
	f.BoolVarP(&preflightOpts.Verbose, "verbose", "v", false, "verbose mode")
	return pfCleanCmd
}
