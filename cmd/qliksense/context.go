package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func qliksenseConfigCmds(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)

	cmd = &cobra.Command{
		Use:     "config",
		Short:   "Set qliksense application configuration",
		Example: `qliksense config <sub-commnad>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// if len(args) == 0 {
			// 	log.Debug("We do not have any args...")
			cmd.Help()
			// } else {
			// 	log.Debug("We have some args...")
			// 	return qliksenseConfigs(q)
			// }
			return nil
		},
	}

	return cmd
}

func setContextConfigCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)

	cmd = &cobra.Command{
		Use:     "set-context",
		Short:   "Sets the context in which the Kubernetes cluster and resources live in",
		Example: `qliksense config set-context <context_name>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Debug("In set Context Config Command")
			return setContextConfig(q, args)
		},
	}
	return cmd
}

func setOtherConfigsCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
		// opts *aboutOptions
	)
	// opts = &aboutOptions{}

	cmd = &cobra.Command{
		Use:     "set",
		Short:   "configure a key value pair into the current context",
		Example: `qliksense config set <key>=<value>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setOtherConfigs(q)
		},
	}
	return cmd
}

func setConfigsCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
		// opts *aboutOptions
	)
	// opts = &aboutOptions{}

	cmd = &cobra.Command{
		Use:     "set-config",
		Short:   "",
		Example: `qliksense config set-config`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setConfigs(q)
		},
	}
	return cmd
}

func setSecretsCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
		// opts *aboutOptions
	)
	// opts = &aboutOptions{}

	cmd = &cobra.Command{
		Use:     "set-secrets",
		Short:   "",
		Example: `qliksense config set-secrets`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setSecrets(q)
		},
	}
	return cmd
}

func viewConfigCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)

	cmd = &cobra.Command{
		Use:     "view",
		Short:   "view qliksense current context configuration",
		Example: `qliksense config view`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return viewQliksenseConfig(q)
		},
	}
	return cmd
}

func viewQliksenseConfig(q *qliksense.Qliksense) error {
	// retieve current context from config.yaml
	var qliksenseConfig api.QliksenseConfig
	qliksenseConfigFile := filepath.Join(q.QliksenseHome, qliksenseConfigFile)
	log.Debugf("qliksenseConfigFile: %s", qliksenseConfigFile)

	qliksense.ReadFromFile(&qliksenseConfig, qliksenseConfigFile)
	currentContext := qliksenseConfig.Spec.CurrentContext
	log.Debugf("Current-context from config.yaml: %s", currentContext)

	// look for file with that name+.yaml in contexts/, if not exists=> output error and show qliksense-default.yaml
	// cat it out
	if currentContext != "" {
		qliksenseContextsFile := filepath.Join(q.QliksenseHome, qliksenseContextsDir, currentContext+".yaml")
		log.Debugf("Context file path: %s", qliksenseContextsFile)
		if qliksense.FileExists(qliksenseContextsFile) {
			content, err := ioutil.ReadFile(qliksenseContextsFile)
			if err != nil {
				log.Fatalf("Unable to read the file: %v", err)
			}
			fmt.Printf("%s", content)
		}
	} else {
		// current-context is empty
		fmt.Println(`Please run the "qliksense config set-context <context-name>" first before viewing the current context info`)
	}

	return nil
}

func qliksenseConfigs(q *qliksense.Qliksense) error {

	// var yamlConfig YamlConfig

	// yamlConfig.readYamlConfig(testYamlFile)
	// log.Debugf("Config read from the given yaml:\n\nApiVersion: %s\nKind: %s\n\n", string(yamlConfig.ApiVersion), string(yamlConfig.Kind))
	// yamlConfig.ApiVersion = "blah"
	// yamlConfig.Kind = "CustomResource"
	// yamlConfig.writeYamlConfigToFile("myqliksense.yaml") // TO-DO: derive the filename from the context-name

	// create a file: /.qliksense/config/qliksense_config.yaml
	// write baseConfig yaml into it.

	return nil
}

func setSecrets(q *qliksense.Qliksense) error {
	return nil
}

func setConfigs(q *qliksense.Qliksense) error {
	return nil
}

func setOtherConfigs(q *qliksense.Qliksense) error {
	return nil
}

// set the context for qliksense kubernetes resources to live in
func setContextConfig(q *qliksense.Qliksense, args []string) error {

	//check if file exists in the name of the given context name
	// if it exists: pull up it's config info and load in memory, look for more updates by way of subsequent commands, and then finally save the updated configs to the file.
	// if it doesnt exist, create a file in the name of the context specified, gather all the configs that are requested by way of subsequent commands, then save the entire
	// thing in a file with the same name as the context.
	if len(args) == 1 {
		log.Debugf("The command received: %s", args)
		setUpQliksenseContext(q.QliksenseHome, args[0])
	} else {
		log.Fatalf("Please provide a name to configure the context with.")
	}
	return nil
}
