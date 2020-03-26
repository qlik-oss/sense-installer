package main

import (
	"io"
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func applyCmd(q *qliksense.Qliksense) *cobra.Command {
	filePath := ""
	acceptEULA := ""
	keepPatchFiles := true
	c := &cobra.Command{
		Use:     "apply",
		Short:   "install qliksense based on provided cr file",
		Long:    `install qliksense based on provided cr file`,
		Example: `qliksense apply -f file_name`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLoadOrApplyCommandE(cmd, func(reader io.Reader) error {
				opts := &qliksense.InstallCommandOptions{
					AcceptEULA: acceptEULA,
				}
				if eulaAcceptedFromPrompt {
					opts.AcceptEULA = "yes"
				}
				return q.ApplyCRFromReader(reader, opts, keepPatchFiles)
			})
		},
	}

	f := c.Flags()
	f.StringVarP(&filePath, "file", "f", "", "Install from a CR file")
	f.StringVarP(&acceptEULA, "acceptEULA", "a", "", "AcceptEULA for qliksense")
	c.MarkFlagRequired("file")

	eulaPreRunHooks.addValidator(c.Name(), func(cmd *cobra.Command, q *qliksense.Qliksense) (bool, error) {
		if strings.ToLower(strings.TrimSpace(acceptEULA)) == "yes" {
			return true, nil
		} else {
			return loadOrApplyCommandEulaPreRunHook(cmd, q)
		}
	})

	return c
}
