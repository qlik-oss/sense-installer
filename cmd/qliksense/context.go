package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/qlik-oss/k-apis/config"
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
		Example: `qliksense config <sub-command>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Help()
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
	)

	cmd = &cobra.Command{
		Use:     "set",
		Short:   "configure a key value pair into the current context",
		Example: `qliksense config set <key>=<value>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setOtherConfigs(q, args)
		},
	}
	return cmd
}

func setConfigsCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)

	cmd = &cobra.Command{
		Use:     "set-configs",
		Short:   "set configurations into the qliksense context",
		Example: `qliksense config set-configs <key>=<value>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setConfigs(q, args)
		},
	}
	return cmd
}

func setSecretsCmd(q *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)

	cmd = &cobra.Command{
		Use:     "set-secrets",
		Short:   "set secrets configurations into the qliksense context",
		Example: `qliksense config set-secrets <key>=<value> --secret`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setSecrets(q, args)
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

	// check for existence of a file with that name.yaml in contexts/, if it exists-> display it.
	// If it does not exist-> output error
	if currentContext != "" {
		qliksenseContextsFile := filepath.Join(q.QliksenseHome, qliksenseContextsDir, currentContext, currentContext+".yaml")
		log.Debugf("Context file path: %s", qliksenseContextsFile)
		if qliksense.FileExists(qliksenseContextsFile) {
			content, err := ioutil.ReadFile(qliksenseContextsFile)
			if err != nil {
				log.Fatalf("Unable to read the file: %v", err)
			}
			fmt.Printf("%s", content)
		} else {
			log.Fatalf("Context file does not exist.\nPlease try re-running `qliksense config set-context <context-name>` and then `qliksense config view` again")
		}
	} else {
		// current-context is empty
		log.Info(`Please run the "qliksense config set-context <context-name>" first before viewing the current context info`)
	}

	return nil
}

func setSecrets(q *qliksense.Qliksense, args []string) error {
	// retieve current context from config.yaml
	qliksenseCR, qliksenseContextsFile := retrieveCurrentContextInfo(q)
	qliksenseCR.Spec.Secrets = processConfigArgs(args, qliksenseCR.Spec.Secrets)

	// write modified content into context.yaml
	qliksense.WriteToFile(&qliksenseCR, qliksenseContextsFile)

	return nil
}

func setConfigs(q *qliksense.Qliksense, args []string) error {
	// retieve current context from config.yaml
	qliksenseCR, qliksenseContextsFile := retrieveCurrentContextInfo(q)
	qliksenseCR.Spec.Configs = processConfigArgs(args, qliksenseCR.Spec.Configs)

	// write modified content into context.yaml
	qliksense.WriteToFile(&qliksenseCR, qliksenseContextsFile)

	return nil
}

func retrieveCurrentContextInfo(q *qliksense.Qliksense) (api.QliksenseCR, string) {
	var qliksenseConfig api.QliksenseConfig
	qliksenseConfigFile := filepath.Join(q.QliksenseHome, qliksenseConfigFile)
	log.Debugf("qliksenseConfigFile: %s", qliksenseConfigFile)

	qliksense.ReadFromFile(&qliksenseConfig, qliksenseConfigFile)
	currentContext := qliksenseConfig.Spec.CurrentContext
	log.Debugf("Current-context from config.yaml: %s", currentContext)
	if currentContext == "" {
		// current-context is empty
		log.Error(`Please run the "qliksense config set-context <context-name>" first before viewing the current context info`)
	}
	// read the context.yaml file
	var qliksenseCR api.QliksenseCR
	if currentContext == "" {
		// current-context is empty
		log.Fatal(`Please run the "qliksense config set-context <context-name>" first before viewing the current context info`)
	}
	qliksenseContextsFile := filepath.Join(q.QliksenseHome, qliksenseContextsDir, currentContext, currentContext+".yaml")
	log.Debugf("Context file path: %s", qliksenseContextsFile)
	if !qliksense.FileExists(qliksenseContextsFile) {
		log.Fatalf("Context file does not exist.\nPlease try re-running `qliksense config set-context <context-name>` and then `qliksense config view` again")
	}
	qliksense.ReadFromFile(&qliksenseCR, qliksenseContextsFile)

	log.Debugf("Read QliksenseCR: %v", qliksenseCR)
	log.Debugf("Read context file: %s", qliksenseContextsFile)
	return qliksenseCR, qliksenseContextsFile
}

func processConfigArgs(args []string, configsMap map[string]config.NameValues) map[string]config.NameValues {
	// prepare received args
	log.Debugf("Here is the command: %s", args[0])
	// split args[0] into key and value
	if len(args) == 0 {
		log.Fatalf("No args were provided. Please provide args to configure the current context")
	}

	re1 := regexp.MustCompile(`(\w{1,})\[name=(\w{1,})\]=("*\w+"*)`)
	for _, arg := range args {
		result := re1.FindStringSubmatch(arg)
		fmt.Printf("finding matches...\n")

		// check if result array's length is == 4 (index 0 - is the full match & indices 1,2,3- are the fields we need)
		if len(result) != 4 {
			log.Error("Please provide valid args for this command")
			continue
		}
		if len(configsMap) == 0 {
			configsMap = map[string]config.NameValues{}
		}
		track := false
		for k1 := range configsMap {
			if k1 == result[1] {
				track = true
				break
			}
		}
		if !track {
			configsMap[result[1]] = config.NameValues{}
		}
		nameValues := configsMap[result[1]]
		nvTrack := false
		for ind, nv := range nameValues {
			if nv.Name == result[2] {
				nv.Value = result[3]
				nameValues[ind] = nv
				nvTrack = true
				break
			}
		}
		if !nvTrack {
			nameValues = append(nameValues, config.NameValue{
				Name:  result[2],
				Value: result[3],
			})
		}
		configsMap[result[1]] = nameValues
	}
	return configsMap
}

func setOtherConfigs(q *qliksense.Qliksense, args []string) error {
	// retieve current context from config.yaml
	qliksenseCR, qliksenseContextsFile := retrieveCurrentContextInfo(q)

	// modify appropriate fields
	log.Debugf("Command: %s", args[0])
	// split args[0] into key and value
	if len(args) > 0 {
		argsString := strings.Split(args[0], "=")
		log.Debugf("Split string: %v", argsString)
		switch argsString[0] {
		case "profile":
			log.Debugf("Current profile: %s, Incoming profile: %s", qliksenseCR.Spec.Profile, argsString[1])
			qliksenseCR.Spec.Profile = argsString[1]
			log.Debugf("Current profile after modification: %s ", qliksenseCR.Spec.Profile)
		case "namespace":
			log.Debugf("Current namespace: %s, Incoming namespace: %s", qliksenseCR.Spec.NameSpace, argsString[1])
			qliksenseCR.Spec.NameSpace = argsString[1]
			log.Debugf("Current namespace after modification: %s ", qliksenseCR.Spec.NameSpace)
		case "git.repository":
			log.Debugf("Current git.repository: %s, Incoming git.repository: %s", qliksenseCR.Spec.Git.Repository, argsString[1])
			qliksenseCR.Spec.Git.Repository = argsString[1]
			log.Debugf("Current git repository after modification: %s ", qliksenseCR.Spec.Git.Repository)
		case "storageClassName":
			log.Debugf("Current StorageClassName: %s, Incoming StorageClassName: %s", qliksenseCR.Spec.StorageClassName, argsString[1])
			qliksenseCR.Spec.StorageClassName = argsString[1]
			log.Debugf("Current StorageClassName after modification: %s ", qliksenseCR.Spec.StorageClassName)
		default:
			log.Debugf("As part of the `qliksense config set` command, please enter one of: profile, namespace, storageClassName or git.repository arguments")
		}
	} else {
		log.Fatalf("No args were provided. Please provide args to configure the current context")
	}
	// write modified content into context.yaml
	qliksense.WriteToFile(&qliksenseCR, qliksenseContextsFile)

	return nil
}

// set the context for qliksense kubernetes resources to live in
func setContextConfig(q *qliksense.Qliksense, args []string) error {
	if len(args) == 1 {
		log.Debugf("The command received: %s", args)
		setUpQliksenseContext(q.QliksenseHome, args[0])
	} else {
		log.Fatalf("Please provide a name to configure the context with.")
	}
	return nil
}
