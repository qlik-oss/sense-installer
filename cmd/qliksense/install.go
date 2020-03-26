package main

import (
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func installCmd(q *qliksense.Qliksense) *cobra.Command {
	opts := &qliksense.InstallCommandOptions{}
	keepPatchFiles := false
	c := &cobra.Command{
		Use:     "install",
		Short:   "install a qliksense release",
		Long:    `install a qliksense release`,
		Example: `qliksense install <version> #if no version provides, expect manifestsRoot is set somewhere in the file system`,
		RunE: func(cmd *cobra.Command, args []string) error {
			version := ""
			if len(args) != 0 {
				version = args[0]
			}
			if eulaAcceptedFromPrompt {
				opts.AcceptEULA = "yes"
			}
			return q.InstallQK8s(version, opts, keepPatchFiles)
		},
	}

	f := c.Flags()
	f.StringVarP(&opts.AcceptEULA, "acceptEULA", "a", "", "AcceptEULA for qliksense")
	f.StringVarP(&opts.StorageClass, "storageClass", "s", "", "Storage class for qliksense")
	f.StringVarP(&opts.MongoDbUri, "mongoDbUri", "m", "", "mongoDbUri for qliksense (i.e. mongodb://qlik-default-mongodb:27017/qliksense?ssl=false)")
	f.StringVarP(&opts.RotateKeys, "rotateKeys", "r", "", "Rotate JWT keys for qliksense (yes:rotate keys/ no:use exising keys from cluster/ None: use default EJSON_KEY from env")
	f.BoolVar(&keepPatchFiles, keepPatchFilesFlagName, keepPatchFiles, keepPatchFilesFlagUsage)

	eulaPreRunHooks.addValidator(c.Name(), func(cmd *cobra.Command, q *qliksense.Qliksense) (bool, error) {
		return strings.ToLower(strings.TrimSpace(opts.AcceptEULA)) == "yes", nil
	})

	return c
}
