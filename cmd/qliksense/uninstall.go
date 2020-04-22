package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func uninstallCmd(q *qliksense.Qliksense) *cobra.Command {
	skipConfirmation := false
	c := &cobra.Command{
		Use:     "uninstall",
		Short:   "Uninstall the deployed qliksense.",
		Long:    `Uninstall the deployed qliksense. By default uninstall the current context`,
		Example: `qliksense uninstall <context-name>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return q.UninstallQK8s(args[0], skipConfirmation)
			}
			return q.UninstallQK8s("", skipConfirmation)
		},
	}

	f := c.Flags()

	f.BoolVar(&skipConfirmation, "yes", skipConfirmation, "skips confirmation")

	return c
}
