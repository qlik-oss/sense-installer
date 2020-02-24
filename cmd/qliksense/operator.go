package main

import (
	"fmt"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

var operatorCmd = &cobra.Command{
	Use:   "operator",
	Short: "Configuration for operator",
	Long:  `Configuration for operator`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("User like: operator view")
	},
}

func operatorViewCmd(q *qliksense.Qliksense) *cobra.Command {
	c := &cobra.Command{
		Use:   "view",
		Short: "View Configuration for operator",
		Long:  `View Configuration for operator`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.ViewOperator()
		},
	}
	return c
}
