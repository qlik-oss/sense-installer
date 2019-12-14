package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func porter(q *qliksense.Qliksense) *cobra.Command {
	return &cobra.Command{
		Use:   "porter",
		Short: "Execute a porter command",
		RunE: func(cobCmd *cobra.Command, args []string) error {
			return q.CallPorter(args)
		},
	}
}
