package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "keys for qliksense",
}

func keysRotateCmd(q *qliksense.Qliksense) *cobra.Command {
	c := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate Qliksense application keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.RotateKeys()
		},
	}
	return c
}
