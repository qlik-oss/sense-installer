package main

import (
	"errors"
	"fmt"
	"os"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"

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
	base64Encoded := false
	cmd = &cobra.Command{
		Use:   "set-configs",
		Short: "set configurations into the qliksense context as key-value pairs",
		Example: `
qliksense config set-configs <service_name>.<attribute>="<value>"
		- The above configuration will be displayed in the CR
qliksense config set-configs <service_name>.<attribute>="<value" --base64
		- if the value is base64 encoded
echo "something" | base64 | qliksense config set-configs <service_name>.<attribute> --base64
	- value is coming from input pipe as base64 encoded
echo "something" | qliksense config set-configs <service_name>.<attribute>
	- value is coming from input pipe

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if isInputFromPipe() && len(args) == 1 {
				return q.SetConfigFromReader(args[0], os.Stdin, base64Encoded)
			}
			return q.SetConfigs(args, base64Encoded)
		},
	}
	f := cmd.Flags()
	f.BoolVarP(&base64Encoded, "base64", "", false, "if the arguments value is base64 encoded")
	return cmd
}

func setSecretsCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd    *cobra.Command
		secret bool
	)
	base64Encoded := false
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
		  - The above configuration will be displayed in the CR 
qliksense config set-secrets <service_name>.<attribute>="<value>" --base64
		- the <value> is base64 encoded
echo "something" | base64 | qliksense config set-secrets <service_name>.<attribute> --base64
		- value coming from input pipe as base64 encoded
echo "something" | qliksense config set-secrets <service_name>.<attribute>
		- value coming from input pipe`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if isInputFromPipe() && len(args) == 1 {
				return q.SetSecretsFromReader(args[0], os.Stdin, secret, base64Encoded)
			}
			return q.SetSecrets(args, secret, base64Encoded)
		},
	}
	f := cmd.Flags()
	f.BoolVar(&secret, "secret", false, "Whether secrets should be encrypted as a Kubernetes Secret resource")
	f.BoolVarP(&base64Encoded, "base64", "", false, "if the arguments value is base64 encoded")

	return cmd
}

func deleteContextConfigCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)
	skipConfirmation := false
	cmd = &cobra.Command{
		Use:     "delete-context",
		Short:   "deletes a specific context locally (not in-cluster)",
		Example: `qliksense config delete-contexts <context_name>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.DeleteContextConfig(args, skipConfirmation)
		},
	}
	f := cmd.Flags()

	f.BoolVar(&skipConfirmation, "yes", skipConfirmation, "skips confirmation")
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

func cleanConfigRepoPatchesCmd(q *qliksense.Qliksense) *cobra.Command {
	return &cobra.Command{
		Use:     "clean-config-repo-patches",
		Short:   "Clean config repo patch files",
		Example: "qliksense config clean-config-repo-patches",
		RunE: func(cmd *cobra.Command, args []string) error {
			qConfig := qapi.NewQConfig(q.QliksenseHome)
			if err := q.DiscardAllUnstagedChangesFromGitRepo(qConfig); err != nil {
				return fmt.Errorf("error removing temporary changes to the config: %v\n", err)
			}
			fmt.Println("done")
			return nil
		},
	}
}

func unsetCmd(q *qliksense.Qliksense) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset",
		Short: "remove a key from a context or a secrets or a configs from the context",
		Example: `
# remove the key from CR
qliksense config unset <key>

# remove the key from service inside configs/secrets of CR
qliksense config unset <service>.<key> 

# remove the service from inside configs/secrets of CR
qliksense config usnet <servcie> 

all of the above supports space separated multiple arguments
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.UnsetCmd(args)
		},
		Args: cobra.MinimumNArgs(1),
	}
	return cmd
}
