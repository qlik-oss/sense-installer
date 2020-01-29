package main

import (
	"fmt"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

type aboutCommandOptions struct {
	Tag      		string
	Directory   string
	Profile 		string
}

func about(q *qliksense.Qliksense) *cobra.Command {
	opts := &aboutCommandOptions{}

	c := &cobra.Command{
		Use:   "about",
		Short: "About Qlik Sense",
		Long:  "Gives the verion of QLik Sense on Kuberntetes and versions of images.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if out, err := q.About(opts.Tag, opts.Directory, opts.Profile); err != nil {
				return err
			} else if _, err := fmt.Println(string(out)); err != nil {
				return err
			}
			return nil
		},
	}
	f := c.Flags()
	f.StringVar(&opts.Tag, "tag", "",
		"Use configuration specified by the given release tag in the qliksense-k8s repository")
	f.StringVar(&opts.Directory, "directory", "",
		"Use configuration in the local directory")
	f.StringVar(&opts.Profile, "profile", "",
		"Configuration profile")
	return c
}
