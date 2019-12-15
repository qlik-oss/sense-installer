package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func pullQliksenseImages(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)
	cmd = &cobra.Command{
		Use:     "pull",
		Short:   "Pull docke images for offline install",
		Example: `  qliksense pull`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.PullImages()
		},
	}

	return cmd
}
