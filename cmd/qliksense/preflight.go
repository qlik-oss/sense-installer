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

func pfDnsCheckCmd(q *qliksense.Qliksense) *cobra.Command {
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
			if namespace == "" {
				namespace = "default"
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

func pfK8sVersionCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	var preflightCheckK8sVersionCmd = &cobra.Command{
		Use:     "kube-version",
		Short:   "check kubernetes version",
		Long:    `check minimum valid kubernetes version on the cluster`,
		Example: `qliksense preflight kube-version`,
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

func pfAllChecksCmd(q *qliksense.Qliksense) *cobra.Command {
	var mongodbUrl, username, password, caCertFile, clientCertFile string
	var tls bool
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
			if namespace == "" {
				namespace = "default"
			}
			qp.RunAllPreflightChecks(namespace, kubeConfigContents, tls, mongodbUrl, username, password, caCertFile, clientCertFile)
			return nil

		},
	}
	f := preflightAllChecksCmd.Flags()
	f.StringVarP(&mongodbUrl, "mongodb-url", "", "", "mongodbUrl to try connecting to")
	f.StringVarP(&username, "mongodb-username", "u", "", "username to connect to mongodb")
	f.StringVarP(&password, "mongodb-password", "p", "", "password to connect to mongodb")
	f.StringVarP(&caCertFile, "mongodb-ca-cert", "", "", "certificate to use for mongodb check")
	f.StringVarP(&clientCertFile, "mongodb-client-cert", "", "", "client-certificate to use for mongodb check")
	f.BoolVar(&tls, "mongodb-tls", false, "enable tls?")
	return preflightAllChecksCmd
}

func pfDeploymentCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	var pfDeploymentCheckCmd = &cobra.Command{
		Use:     "deployment",
		Short:   "perform preflight deploymwnt check",
		Long:    `perform preflight deployment check to ensure that we can create deployments in the cluster`,
		Example: `qliksense preflight deployment`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q}

			// Preflight deployments check
			fmt.Printf("Preflight deployment check\n")
			fmt.Println("--------------------------")
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Printf("Preflight deployment check FAILED\n")
				log.Fatal(err)
			}
			if namespace == "" {
				namespace = "default"
			}
			if err = qp.CheckDeployment(namespace, kubeConfigContents); err != nil {
				fmt.Println(err)
				fmt.Print("Preflight deploy check FAILED\n")
				log.Fatal()
			}
			return nil
		},
	}
	return pfDeploymentCheckCmd
}

func pfServiceCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	var pfServiceCheckCmd = &cobra.Command{
		Use:     "service",
		Short:   "perform preflight service check",
		Long:    `perform preflight service check to ensure that we are able to create services in the cluster`,
		Example: `qliksense preflight service`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q}

			// Preflight service check
			fmt.Printf("Preflight service check\n")
			fmt.Println("-----------------------")
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Printf("Preflight service check FAILED\n")
				log.Fatal(err)
			}
			if namespace == "" {
				namespace = "default"
			}
			if err = qp.CheckService(namespace, kubeConfigContents); err != nil {
				fmt.Println(err)
				fmt.Print("Preflight service check FAILED\n")
				log.Fatal()
			}
			return nil
		},
	}
	return pfServiceCheckCmd
}

func pfPodCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	var pfPodCheckCmd = &cobra.Command{
		Use:     "pod",
		Short:   "perform preflight pod check",
		Long:    `perform preflight pod check to ensure we can create pods in the cluster`,
		Example: `qliksense preflight pod`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q}

			// Preflight pod check
			fmt.Printf("Preflight pod check\n")
			fmt.Println("--------------------")
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Printf("Preflight pod check FAILED\n")
				log.Fatal(err)
			}
			if namespace == "" {
				namespace = "default"
			}
			if err = qp.CheckPod(namespace, kubeConfigContents); err != nil {
				fmt.Println(err)
				fmt.Print("Preflight pod check FAILED\n")
				log.Fatal()
			}
			return nil
		},
	}
	return pfPodCheckCmd
}

func pfCreateRoleCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	var preflightRoleCmd = &cobra.Command{
		Use:     "role",
		Short:   "preflight create role check",
		Long:    `perform preflight role check to ensure we are able to create a role in the cluster`,
		Example: `qliksense preflight createRole`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q}

			// Preflight role check
			fmt.Printf("Preflight role check\n")
			fmt.Println("---------------------------")
			namespace, _, err := preflight.InitPreflight()
			if err != nil {
				fmt.Printf("Preflight role check FAILED\n")
				log.Fatal(err)
			}
			if err = qp.CheckCreateRole(namespace); err != nil {
				fmt.Println(err)
				fmt.Print("Preflight role FAILED\n")
				log.Fatal()
			}
			return nil
		},
	}
	return preflightRoleCmd
}

func pfCreateRoleBindingCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	var preflightRoleBindingCmd = &cobra.Command{
		Use:     "rolebinding",
		Short:   "preflight create rolebinding check",
		Long:    `perform preflight rolebinding check to ensure we are able to create a rolebinding in the cluster`,
		Example: `qliksense preflight rolebinding`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q}

			// Preflight createRoleBinding check
			fmt.Printf("Preflight rolebinding check\n")
			fmt.Println("---------------------------")
			namespace, _, err := preflight.InitPreflight()
			if err != nil {
				fmt.Printf("Preflight rolebinding check FAILED\n")
				log.Fatal(err)
			}
			if err = qp.CheckCreateRoleBinding(namespace); err != nil {
				fmt.Println(err)
				fmt.Print("Preflight rolebinding check FAILED\n")
				log.Fatal()
			}
			return nil
		},
	}
	return preflightRoleBindingCmd
}

func pfCreateServiceAccountCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	var preflightServiceAccountCmd = &cobra.Command{
		Use:     "serviceaccount",
		Short:   "preflight create ServiceAccount check",
		Long:    `perform preflight serviceaccount check to ensure we are able to create a service account in the cluster`,
		Example: `qliksense preflight serviceaccount`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q}

			// Preflight createServiceAccount check
			fmt.Printf("Preflight ServiceAccount check\n")
			fmt.Println("-------------------------------------")
			namespace, _, err := preflight.InitPreflight()
			if err != nil {
				fmt.Printf("Preflight serviceaccount check FAILED\n")
				log.Fatal(err)
			}
			if err = qp.CheckCreateServiceAccount(namespace); err != nil {
				fmt.Println(err)
				fmt.Print("Preflight serviceaccount check FAILED\n")
				log.Fatal()
			}
			return nil
		},
	}
	return preflightServiceAccountCmd
}

func pfCreateAuthCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	var preflightCreateAuthCmd = &cobra.Command{
		Use:     "authcheck",
		Short:   "preflight authcheck",
		Long:    `perform preflight authcheck that combines the role, rolebinding and serviceaccount checks`,
		Example: `qliksense preflight authcheck`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q}

			// Preflight authcheck
			fmt.Printf("Preflight authcheck\n")
			fmt.Println("------------------------")
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Printf("Preflight authcheck FAILED\n")
				log.Fatal(err)
			}
			if err = qp.CheckCreateRB(namespace, kubeConfigContents); err != nil {
				fmt.Println(err)
				fmt.Print("Preflight authcheck FAILED\n")
				log.Fatal()
			}
			return nil
		},
	}
	return preflightCreateAuthCmd
}

func pfMongoCheckCmd(q *qliksense.Qliksense) *cobra.Command {
	var mongodbUrl, username, password, caCertFile, clientCertFile string
	var tls bool
	var preflightMongoCmd = &cobra.Command{
		Use:     "mongo",
		Short:   "preflight mongo OR preflight mongo --url=<url>",
		Long:    `perform preflight mongo check to ensure we are able to connect to a mongodb instance in the cluster`,
		Example: `qliksense preflight mongo OR preflight mongo --url=<url>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			qp := &preflight.QliksensePreflight{Q: q}

			// Preflight mongo check
			fmt.Printf("Preflight mongo check\n")
			fmt.Println("-------------------------------------")
			namespace, kubeConfigContents, err := preflight.InitPreflight()
			if err != nil {
				fmt.Printf("Preflight mongo check FAILED\n")
				log.Fatal(err)
			}
			if namespace == "" {
				namespace = "default"
			}
			if err = qp.CheckMongo(kubeConfigContents, namespace, mongodbUrl, tls, username, password, caCertFile, clientCertFile); err != nil {
				fmt.Println(err)
				fmt.Print("Preflight mongo check FAILED\n")
				log.Fatal()
			}
			return nil
		},
	}
	f := preflightMongoCmd.Flags()
	f.StringVarP(&mongodbUrl, "url", "", "", "mongodbUrl to try connecting to")
	f.StringVarP(&username, "username", "u", "", "username to connect to mongodb")
	f.StringVarP(&password, "password", "p", "", "password to connect to mongodb")
	f.StringVarP(&caCertFile, "ca-cert", "", "", "ca certificate to use for mongodb check")
	f.StringVarP(&clientCertFile, "client-cert", "", "", "client-certificate to use for mongodb check")
	f.BoolVar(&tls, "tls", false, "enable tls?")
	return preflightMongoCmd
}
