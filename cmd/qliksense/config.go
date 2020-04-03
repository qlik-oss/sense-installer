package main

import (
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func configCmd(q *qliksense.Qliksense) *cobra.Command {
	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "do operations on/around CR",
		Long:  `do operations on/around CR`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.ConfigViewCR()
		},
	}
	return configCmd
}

func configApplyCmd(q *qliksense.Qliksense) *cobra.Command {
	c := &cobra.Command{
		Use:     "apply",
		Short:   "generate the patches and apply manifests to k8s",
		Long:    `generate patches based on CR and apply manifests to k8s`,
		Example: `qliksense config apply`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.ConfigApplyQK8s()
		},
	}
	return c
}

func configViewCmd(q *qliksense.Qliksense) *cobra.Command {
	c := &cobra.Command{
		Use:     "view",
		Short:   "view the qliksense operator CR",
		Long:    `display the operator CR, that has been created for the current context`,
		Example: `qliksense config view`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.ConfigViewCR()
		},
	}
	return c
}

func configEditCmd(q *qliksense.Qliksense) *cobra.Command {
	c := &cobra.Command{
		Use:   "edit [context-name]",
		Short: "Edit the context cr",
		Long: `edit the context cr. if no context name provided default context will be edited
		It will open the vim editor unless KUBE_EDITOR is defined`,
		Example: `qliksense config edit [context-name]`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return q.EditCR(args[0])
			}
			return q.EditCR("")
		},
	}
	return c
}
