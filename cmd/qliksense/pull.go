package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func pullQliksenseImages(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd  *cobra.Command
		opts *aboutOptions
	)
	opts = &aboutOptions{}

	cmd = &cobra.Command{
		Use:     "pull",
		Short:   "Pull docker images for offline install",
		Example: `qliksense pull`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.PullImages(opts.getTagDefaults(args))
		},
	}
	f := cmd.Flags()
	f.StringVarP(&opts.Version, "version", "v", "latest",
		"From version of Qlik Sense The images will be pulled")
	f.StringVarP(&opts.Tag, "tag", "t", "",
		"Use a bundle in an OCI registry specified by the given tag")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	return cmd
}
