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
	opts := &qliksense.CrdCommandOptions{}
	c := &cobra.Command{
		Use:   "view",
		Short: "View CRDs for qliksense application. use view --all to see opearator crd as well ",
		Long:  `View CRDs for qliksense application. use view --all to see opearator crd as well`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.ViewCrds(opts)
		},
	}
	f := c.Flags()
	f.BoolVarP(&opts.All, "all", "", false, "Include All CRDs")
	return c
}

func crdsInstallCmd(q *qliksense.Qliksense) *cobra.Command {
	opts := &qliksense.CrdCommandOptions{}
	c := &cobra.Command{
		Use:   "install",
		Short: "Install CRDs fro Qliksense application. Use install --all to include operator crd",
		Long:  `Install CRDs fro Qliksense application. Use install --all to include operator crd`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.InstallCrds(opts)
		},
	}
	f := c.Flags()
	f.BoolVarP(&opts.All, "all", "", false, "Include All CRDs")
	return c
}
