package main

import (
	"errors"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func installCmd(q *qliksense.Qliksense) *cobra.Command {
	opts := &qliksense.InstallCommandOptions{}
	c := &cobra.Command{
		Use:     "install",
		Short:   "install a qliksense release",
		Long:    `install a qliksesne release`,
		Example: `qliksense install <version>`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires a version (i.e. v1.0.0)")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.InstallQK8s(args[0], opts)
		},
	}

	f := c.Flags()
	f.StringVarP(&opts.AcceptEULA, "acceptEULA", "a", "", "AcceptEULA for qliksense")
	f.StringVarP(&opts.Namespace, "namespace", "n", "", "Namespace where to install the qliksense")
	f.StringVarP(&opts.StorageClass, "storageClass", "s", "", "Storage class for qliksense")
	f.StringVarP(&opts.MongoDbUri, "mongoDbUri", "m", "", "mongoDbUri for qliksense (i.e. mongodb://qliksense-mongodb:27017/qliksense?ssl=false)")

	return c
}
