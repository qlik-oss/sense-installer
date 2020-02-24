package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

var operatorCmd = &cobra.Command{
	Use:   "operator",
	Short: "Configuration for operator",
	Long:  `Configuration for operator`,
}

func operatorViewCmd(q *qliksense.Qliksense) *cobra.Command {
	c := &cobra.Command{
		Use:   "view",
		Short: "View CRD for operator",
		Long:  `View CRD for operator`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.ViewOperator()
		},
	}
	return c
}
