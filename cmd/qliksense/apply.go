package main

import (
	"io"

	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func applyCmd(q *qliksense.Qliksense) *cobra.Command {
	opts := &qliksense.InstallCommandOptions{}
	filePath := ""
	keepPatchFiles := false
	c := &cobra.Command{
		Use:     "apply",
		Short:   "install qliksense based on provided cr file",
		Long:    `install qliksense based on provided cr file`,
		Example: `qliksense apply -f file_name or cat cr_file | qliksense apply -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLoadOrApplyCommandE(cmd, func(reader io.Reader) error {
				return q.ApplyCRFromReader(reader, opts, keepPatchFiles, true)
			})
		},
	}

	f := c.Flags()
	f.StringVarP(&filePath, "file", "f", "", "Install from a CR file")
	c.MarkFlagRequired("file")
	f.StringVarP(&opts.StorageClass, "storageClass", "s", "", "Storage class for qliksense")
	f.StringVarP(&opts.MongoDbUri, "mongoDbUri", "m", "", "mongoDbUri for qliksense (i.e. mongodb://qlik-default-mongodb:27017/qliksense?ssl=false)")
	f.StringVarP(&opts.RotateKeys, "rotateKeys", "r", "", "Rotate JWT keys for qliksense (yes:rotate keys/ no:use exising keys from cluster/ None: use default EJSON_KEY from env")
	f.BoolVar(&keepPatchFiles, keepPatchFilesFlagName, keepPatchFiles, keepPatchFilesFlagUsage)

	eulaPreRunHooks.addValidator(c.CommandPath(), loadOrApplyCommandEulaPreRunHook)

	return c
}
