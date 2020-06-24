package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

var crdsCmd = &cobra.Command{
	Use:   "crds",
	Short: "crds for qliksense and operators",
	Long:  `crds for qliksense and operators`,
}

func crdsViewCmd(q *qliksense.Qliksense) *cobra.Command {
	opts := &qliksense.CrdCommandOptions{
		All: true,
	}
	c := &cobra.Command{
		Use:   "view",
		Short: "View CRDs for qliksense application. Use view --all=false to exclude the operator CRD",
		Long:  "View CRDs for qliksense application. Use view --all=false to exclude the operator CRD",
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.ViewCrds(opts)
		},
	}
	f := c.Flags()
	f.BoolVarP(&opts.All, "all", "", opts.All, "If set to false, then the operator CRD is excluded")
	return c
}

func crdsInstallCmd(q *qliksense.Qliksense) *cobra.Command {
	opts := &qliksense.CrdCommandOptions{
		All: true,
	}
	c := &cobra.Command{
		Use:   "install",
		Short: "Install CRDs for qliksense application. Use install --all=false to exclude the operator CRD",
		Long:  "Install CRDs for qliksense application. Use install --all=false to exclude the operator CRD",
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.InstallCrds(opts)
		},
	}
	f := c.Flags()
	f.BoolVarP(&opts.All, "all", "", opts.All, "If set to false, then the operator CRD is excluded")
	return c
}
