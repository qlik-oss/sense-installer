package main

import (
	"errors"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func pullQliksenseImages(q *qliksense.Qliksense) *cobra.Command {
	opts := &aboutCommandOptions{}

	cmd := &cobra.Command{
		Use:     "pull",
		Short:   "Pull docker images for offline install",
		Example: `qliksense pull`,
		RunE: func(cmd *cobra.Command, args []string) error {
			version, err := getSingleArg(args)
			if err != nil {
				return err
			}
			return q.PullImages(version, opts.Profile)
		},
	}
	f := cmd.Flags()
	f.StringVar(&opts.Profile, "profile", "", "Configuration profile")
	return cmd
}

func pushQliksenseImages(q *qliksense.Qliksense) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "push",
		Short:   "Push docker images for offline install",
		Example: `qliksense push`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureImageRegistrySetInCR(q); err != nil {
				return err
			} else {
				return q.PushImagesForCurrentCR()
			}
		},
	}
	return cmd
}

func ensureImageRegistrySetInCR(q *qliksense.Qliksense) error {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	if qcr, err := qConfig.GetCurrentCR(); err != nil {
		return err
	} else if registry := qcr.Spec.GetImageRegistry(); registry == "" {
		return errors.New("no image registry set in the CR; to set it use: qliksense config set-image-registry")
	}
	return nil
}
