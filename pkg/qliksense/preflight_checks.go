package qliksense

import (
	"github.com/spf13/cobra"
)

const dnsCheckYAML = ``

// PerformDnsCheck
func PerformDnsCheck(q *Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)

	cmd = &cobra.Command{
		Use:     "preflight dns",
		Short:   "Perform preflight check on dns ",
		Example: `qliksense preflight --dns`,
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}
	return cmd
}
