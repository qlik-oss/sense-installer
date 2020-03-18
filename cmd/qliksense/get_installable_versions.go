package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

const defaultVersionsLimit = 10

func getInstallableVersionsCmd(q *qliksense.Qliksense) *cobra.Command {
	opts := &qliksense.LsRemoteCmdOptions{
		IncludeBranches: false,
		Limit:           defaultVersionsLimit,
	}
	c := &cobra.Command{
		Use:     "get-versions",
		Short:   "list remote/installable versions",
		Long:    `list remote/installable versions`,
		Example: `qliksense get-versions`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.GetInstallableVersions(opts)
		},
	}

	f := c.Flags()
	f.BoolVarP(&opts.IncludeBranches, "include-branches", "", opts.IncludeBranches, "Include branches")
	f.IntVarP(&opts.Limit, "limit", "", opts.Limit, "Maximum versions to list (starting with the highest)")
	return c
}
