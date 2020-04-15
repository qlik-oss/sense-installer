package main

import (
	"io"

	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func installCmd(q *qliksense.Qliksense) *cobra.Command {
	opts := &qliksense.InstallCommandOptions{}
	keepPatchFiles := false
	filePath := ""
	c := &cobra.Command{
		Use:   "install",
		Short: "install a qliksense release",
		Long:  `install a qliksense release`,
		Example: `qliksense install <version> #if no version provides, expect manifestsRoot is set somewhere in the file system
		# qliksense install -f file_name or cat cr_file | qliksense install -f -
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath != "" {
				return runLoadOrApplyCommandE(cmd, func(reader io.Reader) error {
					return q.ApplyCRFromReader(reader, opts, keepPatchFiles, true)
				})
			}
			version := ""
			if len(args) != 0 {
				version = args[0]
			}
			return q.InstallQK8s(version, opts, keepPatchFiles)
		},
	}

	f := c.Flags()
	f.StringVarP(&opts.StorageClass, "storageClass", "s", "", "Storage class for qliksense")
	f.StringVarP(&opts.MongoDbUri, "mongoDbUri", "m", "", "mongoDbUri for qliksense (i.e. mongodb://qlik-default-mongodb:27017/qliksense?ssl=false)")
	f.StringVarP(&opts.RotateKeys, "rotateKeys", "r", "", "Rotate JWT keys for qliksense (yes:rotate keys/ no:use exising keys from cluster/ None: use default EJSON_KEY from env")
	f.BoolVar(&keepPatchFiles, keepPatchFilesFlagName, keepPatchFiles, keepPatchFilesFlagUsage)
	f.StringVarP(&filePath, "file", "f", "", "Install from a CR file")

	return c
}
