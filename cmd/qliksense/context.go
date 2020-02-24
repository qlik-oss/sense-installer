package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func setContextConfigCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)

	cmd = &cobra.Command{
		Use:     "set-context",
		Short:   "Sets the context in which the Kubernetes cluster and resources live in",
		Example: `qliksense config set-context <context_name>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.SetContextConfig(args)
		},
	}
	return cmd
}

func listContextConfigCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)

	cmd = &cobra.Command{
		Use:     "list-contexts",
		Short:   "retrieves the contexts and lists them",
		Example: `qliksense config list-contexts`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.ListContextConfigs()
		},
	}
	return cmd
}

func setOtherConfigsCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)

	cmd = &cobra.Command{
		Use:     "set",
		Short:   "configure a key value pair into the current context",
		Example: `qliksense config set <key>=<value>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.SetOtherConfigs(args)
		},
	}
	return cmd
}

func setConfigsCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)

	cmd = &cobra.Command{
		Use:     "set-configs",
		Short:   "set configurations into the qliksense context",
		Example: `qliksense config set-configs <key>=<value>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.SetConfigs(args)
		},
	}
	return cmd
}

func setSecretsCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd    *cobra.Command
		secret bool
	)

	cmd = &cobra.Command{
		Use:     "set-secrets",
		Short:   "set secrets configurations into the qliksense context",
		Example: `qliksense config set-secrets <key>=<value> --secret=true`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.SetSecrets(args, secret)
		},
	}
	f := cmd.Flags()
	f.BoolVar(&secret, "secret", false, "Whether secrets should be encrypted as a Kubernetes Secret resource")
	return cmd
}
