package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func fetchCmd(q *qliksense.Qliksense) *cobra.Command {
	opts := &qliksense.FetchCommandOptions{}
	c := &cobra.Command{
		Use:     "fetch",
		Short:   "fetch a release from qliksense-k8s repo, if version not supplied, will use from context",
		Long:    `fetch a release from qliksense-k8s repo, if version not supplied, will use from context`,
		Example: `qliksense fetch [version]`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				opts.Version = args[0]
			}
			return q.FetchK8sWithOpts(opts)
		},
	}

	f := c.Flags()
	f.StringVarP(&opts.GitUrl, "url", "", "", "git url from where configuration will be pulled")
	f.StringVarP(&opts.AccessToken, "accessToken", "", "", "access token for git url")
	f.StringVarP(&opts.SecretName, "secretName", "", "", "kubernetes secret name where a key name accessToken exist")

	return c
}
