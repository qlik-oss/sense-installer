package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

const defaultLsRemoteLimit = 10

func lsRemoteCmd(q *qliksense.Qliksense) *cobra.Command {
	opts := &qliksense.LsRemoteCmdOptions{
		IncludeBranches: false,
		Limit:           defaultLsRemoteLimit,
	}
	c := &cobra.Command{
		Use:     "ls-remote",
		Short:   "list remote/installable versions",
		Long:    `list remote/installable versions`,
		Example: `qliksense ls-remote`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.LsRemote(opts)
		},
	}

	f := c.Flags()
	f.BoolVarP(&opts.IncludeBranches, "include-branches", "", opts.IncludeBranches, "Include branches")
	f.IntVarP(&opts.Limit, "limit", "", opts.Limit, "Maximum versions to show (starting with the highest)")
	return c
}
