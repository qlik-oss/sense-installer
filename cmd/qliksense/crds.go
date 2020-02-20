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
	c := &cobra.Command{
		Use:   "view",
		Short: "View CRDs for qliksense application. use view all to see opearator crd as well ",
		Long:  `View CRDs for qliksense application. use view all to see opearator crd as well`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return q.ViewCrds(args[0])
			}
			return q.ViewCrds("")
		},
	}
	return c
}

func crdsInstallCmd(q *qliksense.Qliksense) *cobra.Command {
	c := &cobra.Command{
		Use:   "install",
		Short: "Install CRDs fro Qliksense application. Use install all to include operator crd",
		Long:  `Install CRDs fro Qliksense application. Use install all to include operator crd`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return q.InstallCrds(args[0])
			}
			return q.InstallCrds("")
		},
	}
	return c
}
