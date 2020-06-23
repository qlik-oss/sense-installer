package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func installCmd(q *qliksense.Qliksense) *cobra.Command {
	opts := &qliksense.InstallCommandOptions{
		CleanPatchFiles: true,
	}
	filePath := ""
	c := &cobra.Command{
		Use:   "install",
		Short: "install a qliksense release",
		Long:  `install a qliksense release`,
		Example: `qliksense install <version> #if no version provides, expect manifestsRoot is set somewhere in the file system
		# qliksense install -f file_name or cat cr_file | qliksense install -f -
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			version := ""
			if len(args) != 0 {
				version = args[0]
			}

			if filePath != "" {
				if err := apply(q, cmd, opts); err != nil {
					return err
				}
			} else {
				if err := q.InstallQK8s(version, opts); err != nil {
					return err
				}
			}
			postflightChecksCmd := AllPostflightChecks(q)
			postflightChecksCmd.DisableFlagParsing = true
			return postflightChecksCmd.Execute()
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
	f.BoolVarP(&opts.DryRun, "dry-run", "", false, "Dry run will generate the patches without rotating keys")

	return c
}
