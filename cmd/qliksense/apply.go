package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func applyCmd(q *qliksense.Qliksense) *cobra.Command {
	opts := &qliksense.InstallCommandOptions{
		CleanPatchFiles: true,
	}
	filePath := ""
	c := &cobra.Command{
		Use:     "apply",
		Short:   "install qliksense based on provided cr file",
		Long:    `install qliksense based on provided cr file`,
		Example: `qliksense apply -f file_name or cat cr_file | qliksense apply -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return apply(q, cmd, opts)
		},
	}

	f := c.Flags()
	f.StringVarP(&filePath, "file", "f", "", "Install from a CR file")
	f.StringVarP(&opts.StorageClass, "storageClass", "s", "", "Storage class for qliksense")
	f.StringVarP(&opts.MongodbUri, "mongodbUri", "m", "", "mongodbUri for qliksense (i.e. mongodb://qlik-default-mongodb:27017/qliksense?ssl=false)")
	f.BoolVar(&opts.CleanPatchFiles, cleanPatchFilesFlagName, opts.CleanPatchFiles, cleanPatchFilesFlagUsage)
	f.BoolVarP(&opts.Pull, pullFlagName, pullFlagShorthand, opts.Pull, pullFlagUsage)
	f.BoolVarP(&opts.Push, pushFlagName, pushFlagShorthand, opts.Push, pushFlagUsage)
	f.StringVarP(&opts.AcceptEULA, "acceptEULA", "a", opts.AcceptEULA, "Accept EULA for qliksense")

	if err := c.MarkFlagRequired("file"); err != nil {
		panic(err)
	}
	return c
}

func apply(q *qliksense.Qliksense, cmd *cobra.Command, opts *qliksense.InstallCommandOptions) error {
	if crBytes, err := getCrBytesFromFileFlag(cmd); err != nil {
		return err
	} else {
		return q.ApplyCRFromBytes(crBytes, opts, true)
	}
}
