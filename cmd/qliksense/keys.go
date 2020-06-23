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
		Short: "Rotate qliksense application keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := q.InstallQK8s("", &qliksense.InstallCommandOptions{
				CleanPatchFiles: true,
				RotateKeys:      true,
			}); err != nil {
				return err
			} else {
				postFlightChecksCmd := AllPostflightChecks(q)
				postFlightChecksCmd.DisableFlagParsing = true
				return postFlightChecksCmd.Execute()
			}
		},
	}
	return c
}
