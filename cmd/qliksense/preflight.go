package main

import (
	"fmt"
	"log"

	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func preflightCmd(q *qliksense.Qliksense) *cobra.Command {
	var configCmd = &cobra.Command{
		Use:   "preflight",
		Short: "perform preflight checks on the cluster",
		Long:  `perform preflight checks on the cluster`,
		Example: `qliksense preflight <preflight_check_to_run>
Usage:
qliksense preflight dns
`,
	}
	return configCmd
}

func preflightCheckDnsCmd(q *qliksense.Qliksense) *cobra.Command {
	var preflightDnsCmd = &cobra.Command{
		Use:     "dns",
		Short:   "perform preflight dns check",
		Long:    `perform preflight dns check to check DNS connectivity status in the cluster`,
		Example: `qliksense preflight dns`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := q.DownloadPreflight()
			if err != nil {
				err = fmt.Errorf("There has been an error downloading preflight: %+v", err)
				log.Println(err)
				return err
			}
			return q.CheckDns()
		},
	}
	return preflightDnsCmd
}
