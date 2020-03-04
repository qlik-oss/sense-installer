package main

import (
	"errors"

	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func setContextConfigCmd(q *qliksense.Qliksense) *cobra.Command {
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
		Use:   "set",
		Short: "configure a key value pair into the current context",
		Example: `
qliksense config set <key>=<value>
    - The above configuration will be displayed in the CR
`,
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
		Use:   "set-configs",
		Short: "set configurations into the qliksense context as key-value pairs",
		Example: `
qliksense config set-configs <service_name>.<attribute>="<value>"
    - The above configuration will be displayed in the CR
`,
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
		Use:   "set-secrets",
		Short: "set secrets configurations into the qliksense context as key-value pairs",
		Example: `
qliksense config set-secrets <service_name>.<attribute>="<value>" --secret=true
        - Encrypt the secret value into a new Kubernetes secret resource
        - The secret resource is placed in the location: <qliksense_home>/<contexts>/<context_name>/secrets/<service_name>.yaml
        - Include it's key reference in the current context

qliksense config set-secrets <service_name>.<attribute>="<value>" --secret=false
		- Encrypt the secret value and display it in the current context
		- No secret resource is created
        - The above configuration will be displayed in the CR `,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.SetSecrets(args, secret)
		},
	}
	f := cmd.Flags()
	f.BoolVar(&secret, "secret", false, "Whether secrets should be encrypted as a Kubernetes Secret resource")
	return cmd
}

func setImageRegistryCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd          *cobra.Command
		pushUsername string
		pushPassword string
		pullUsername string
		pullPassword string
		username     string
		password     string
	)

	cmd = &cobra.Command{
		Use:   "set-image-registry",
		Short: "set private image registry",
		Example: `
qliksense config set-image-registry https://your.private.registry.example.com:5000 --push-username foo1 --push-password bar1 --pull-username foo2 --pull-password bar2
qliksense config set-image-registry https://your.private.registry.example.com:5000 --username foo --password bar
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("private docker image registry FQDN is required")
			}
			registry := args[0]

			if username != "" {
				pullUsername = username
				pushUsername = username
			}
			if password != "" {
				pullPassword = password
				pushPassword = password
			}
			if (pullUsername != "" && pushUsername == "") || (pullUsername == "" && pushUsername != "") {
				return errors.New("if you specify pull credentials, you must specify push credentials as well and vise versa")
			}
			if (pullUsername == "" && pullPassword != "") || (pushUsername == "" && pushPassword != "") {
				return errors.New("if you specify passwords, you must specify usernames as well")
			}
			return q.SetImageRegistry(registry, pushUsername, pushPassword, pullUsername, pullPassword)
		},
	}
	f := cmd.Flags()
	f.StringVar(&pushUsername, "push-username", "", "Username used for pushing images")
	f.StringVar(&pushPassword, "push-password", "", "Password used for pushing images")
	f.StringVar(&pullUsername, "pull-username", "", "Username used for pulling images")
	f.StringVar(&pullPassword, "pull-password", "", "Password used for pulling images")
	f.StringVar(&username, "username", "", "Username used for both pushing and pulling images")
	f.StringVar(&password, "password", "", "Password used for both pushing and pulling images")
	return cmd
}
