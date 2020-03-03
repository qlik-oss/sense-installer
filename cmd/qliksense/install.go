package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func installCmd(q *qliksense.Qliksense) *cobra.Command {
	opts := &qliksense.InstallCommandOptions{}
	c := &cobra.Command{
		Use:     "install",
		Short:   "install a qliksense release",
		Long:    `install a qliksense release`,
		Example: `qliksense install <version>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return q.InstallQK8s("", opts)
			}
			return q.InstallQK8s(args[0], opts)
		},
	}

	f := c.Flags()
	f.StringVarP(&opts.AcceptEULA, "acceptEULA", "a", "", "AcceptEULA for qliksense")
	f.StringVarP(&opts.Namespace, "namespace", "n", "", "Namespace where to install the qliksense")
	f.StringVarP(&opts.StorageClass, "storageClass", "s", "", "Storage class for qliksense")
	f.StringVarP(&opts.MongoDbUri, "mongoDbUri", "m", "", "mongoDbUri for qliksense (i.e. mongodb://qlik-default-mongodb:27017/qliksense?ssl=false)")
	f.StringVarP(&opts.RotateKeys, "rotateKeys", "r", "", "Rotate JWT keys for qliksense (yes:rotate keys/ no:use exising keys from cluster/ None: use default EJSON_KEY from env")
	return c
}
