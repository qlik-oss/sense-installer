package qliksense

import (
	"github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/spf13/cobra"
)

const dnsCheckYAML = `
PASTE YOUR OLD DNS CHECK YAML
`

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
			api.LogDebugMessage("Entry: PerformDnsCheck")
			//return q.SetContextConfig(args)
			//cli.createServiceAccount(preflight troubleshootv1beta1.Preflight, namespace string, clientset *kubernetes.Clientset)
			//v := viper.GetViper()
			//res, err := cli.runPreflights(v, args[0])
			//return cli.VersionCmd().RunE(cmd, args)
			api.LogDebugMessage("Exit: PerformDnsCheck")
			return nil
		},
	}
	return cmd
}
