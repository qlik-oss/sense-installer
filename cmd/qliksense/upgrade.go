package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func upgradeCmd(q *qliksense.Qliksense) *cobra.Command {
	keepPatchFiles := false
	c := &cobra.Command{
		Use:     "upgrade",
		Short:   "upgrade qliksense release",
		Long:    `upgrade qliksense release`,
		Example: `qliksense upgrade <version>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.UpgradeQK8s(keepPatchFiles)
		},
	}

	f := c.Flags()
	f.BoolVar(&keepPatchFiles, keepPatchFilesFlagName, keepPatchFiles, keepPatchFilesFlagUsage)
	return c
}
