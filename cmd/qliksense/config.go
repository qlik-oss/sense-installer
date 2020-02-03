package main

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "do operations on/around CR",
	Long:  `do operations on/around CR`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use like: config view or config apply")
	},
}

func configApplyCmd(q *qliksense.Qliksense) *cobra.Command {
	c := &cobra.Command{
		Use:     "apply",
		Short:   "generate the patchs and apply manifests to k8s",
		Long:    `generate patches based on CR and apply manifests to k8s`,
		Example: `qliksense config apply`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.ConfigApplyQK8s()
		},
	}
	return c
}
