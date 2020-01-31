package main

import (
	"errors"
	"github.com/Masterminds/semver/v3"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func installCmd(q *qliksense.Qliksense) *cobra.Command {
	c := &cobra.Command{
		Use:     "install",
		Short:   "install a qliksense release",
		Long:    `install a qliksesne release`,
		Example: `qliksense install <version>`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires a version (i.e. v1.0.0)")
			}
			if _, err := semver.NewVersion(args[0]); err != nil {
				return errors.New("is it not a valid version. should be something like this v1.0.0")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			q.InstallQK8s(args[0])
		},
	}
	return c
}
