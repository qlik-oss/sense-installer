package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func pullQliksenseImages(q *qliksense.Qliksense) *cobra.Command {
	opts := &aboutCommandOptions{}

	cmd := &cobra.Command{
		Use:     "pull",
		Short:   "Pull docke images for offline install",
		Example: `qliksense pull`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if gitRef, err := getAboutCommandGitRef(args); err != nil {
				return err
			} else if err = q.PullImages(gitRef, opts.Profile, false); err != nil {
				return err
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&opts.Profile, "profile", "", "Configuration profile")
	return cmd
}
