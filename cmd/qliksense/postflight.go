package main

import (
	"fmt"

	. "github.com/logrusorgru/aurora"
	ansi "github.com/mattn/go-colorable"
	postflight "github.com/qlik-oss/sense-installer/pkg/postflight"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func postflightCmd(q *qliksense.Qliksense) *cobra.Command {
	postflightOpts := &postflight.PostflightOptions{}
	var postflightCmd = &cobra.Command{
		Use:     "postflight",
		Short:   "perform postflight checks on the cluster",
		Long:    `perform postflight checks on the cluster`,
		Example: `qliksense postflight <postflight_check_to_run>`,
	}
	f := postflightCmd.Flags()
	f.BoolVarP(&postflightOpts.Verbose, "verbose", "v", false, "verbose mode")
	return postflightCmd
}

func pfMigrationCheck(q *qliksense.Qliksense) *cobra.Command {
	out := ansi.NewColorableStdout()
	postflightOpts := &postflight.PostflightOptions{}
	var postflightMigrationCmd = &cobra.Command{
		Use:     "db-migration-check",
		Short:   "check mongodb migration status on the cluster",
		Long:    `check mongodb migration status on the cluster`,
		Example: `qliksense postflight db-migration-check`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pf := &postflight.QliksensePostflight{Q: q, P: postflightOpts}

			// Postflight db_migration_check
			namespace, kubeConfigContents, err := pf.CG.LoadKubeConfigAndNamespace()
			if err != nil {
				fmt.Fprintf(out, "%s\n", Red("Postflight db_migration_check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			if namespace == "" {
				namespace = "default"
			}
			if err = pf.DbMigrationCheck(namespace, kubeConfigContents); err != nil {
				fmt.Fprintf(out, "%s\n", Red("Postflight db_migration_check FAILED"))
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			fmt.Fprintf(out, "%s\n", Green("Postflight db_migration_check completed"))
			return nil
		},
	}
	f := postflightMigrationCmd.Flags()
	f.BoolVarP(&postflightOpts.Verbose, "verbose", "v", false, "verbose mode")
	return postflightMigrationCmd
}
