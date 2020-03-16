package qliksense

import (
	"github.com/spf13/cobra"
)

// PerformDnsCheck
func PerformDnsCheck(q *Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)

	cmd = &cobra.Command{
		Use:   "set-context",
		Short: "Sets the context in which the Kubernetes cluster and resources live in",
		Example: `
qliksense config set-context <context_name>
   - The above configuration will be displayed in the CR
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			//return q.SetContextConfig(args)
			cli
			return nil
		},
	}
	return cmd
}
