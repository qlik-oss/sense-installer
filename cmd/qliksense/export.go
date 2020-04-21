package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func exportCmd(q *qliksense.Qliksense) *cobra.Command {

	c := &cobra.Command{
		Use:     "export",
		Short:   "export files for corresponding context",
		Long:    `exports all context files in zip format`,
		Example: `qliksense export <context_name>`,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 0 {
				context := args[0]
				return q.ExportContext(context)
			}
			return nil
		},
	}

	return c
}