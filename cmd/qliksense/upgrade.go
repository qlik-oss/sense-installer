package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func upgradeCmd(q *qliksense.Qliksense) *cobra.Command {
	c := &cobra.Command{
		Use:     "upgrade",
		Short:   "upgrade qliksense release",
		Long:    `upgrade qliksesne release`,
		Example: `qliksense upgrade <version>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.UpgradeQK8s()
		},
	}

	return c
}
