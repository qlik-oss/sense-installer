package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	. "github.com/logrusorgru/aurora"
	ansi "github.com/mattn/go-colorable"
	"github.com/mitchellh/go-homedir"
	"github.com/qlik-oss/sense-installer/pkg"
	"github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// To run this project in debug mode, run:
// export QLIKSENSE_DEBUG=true
// qliksense <command>

const (
	qlikSenseHomeVar         = "QLIKSENSE_HOME"
	qlikSenseDirVar          = ".qliksense"
	cleanPatchFilesFlagName  = "clean"
	cleanPatchFilesFlagUsage = "Set --clean=false to keep any prior config repo file changes on install (for debugging)"
	pullFlagName             = "pull"
	pullFlagShorthand        = "d"
	pullFlagUsage            = "If using private docker registry, pull (download) all required qliksense images before install"
	pushFlagName             = "push"
	pushFlagShorthand        = "u"
	pushFlagUsage            = "If using private docker registry, push (upload) all downloaded qliksense images to that registry before install"
	rootCommandName          = "qliksense"
)

func initAndExecute() error {
	var (
		qlikSenseHome string
		err           error
	)
	qlikSenseHome, err = setUpPaths()
	if err != nil {
		log.Fatal(err)
	}
	// create dirs and appropriate files for setting up contexts
	api.LogDebugMessage("QliksenseHomeDir: %s\n", qlikSenseHome)

	qliksenseClient := qliksense.New(qlikSenseHome)
	cmd := rootCmd(qliksenseClient)
	if err := cmd.Execute(); err != nil {
		//levenstein checks (auto-suggestions)
		levenstein(cmd)
		return err
	}

	return nil
}

func setUpPaths() (string, error) {
	var (
		homeDir, qlikSenseHome string
		err                    error
	)

	if qlikSenseHome = os.Getenv(qlikSenseHomeVar); qlikSenseHome == "" {
		if homeDir, err = homedir.Dir(); err != nil {
			return "", err
		}
		if homeDir, err = homedir.Expand(homeDir); err != nil {
			return "", err
		}
		qlikSenseHome = filepath.Join(homeDir, qlikSenseDirVar)
	}

	if err := os.MkdirAll(qlikSenseHome, os.ModePerm); err != nil {
		return "", err
	}

	return qlikSenseHome, nil
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of qliksense cli",
	Long:  "Print the version number of qliksense cli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s (%s, %s)\n", pkg.Version, pkg.Commit, pkg.CommitDate)
	},
}

func commandUsesContext(commandName string) bool {
	return commandName != "" &&
		commandName != rootCommandName &&
		commandName != fmt.Sprintf("%v help", rootCommandName) &&
		commandName != fmt.Sprintf("%v version", rootCommandName)
}

func getRootCmd(p *qliksense.Qliksense) *cobra.Command {
	cmd := &cobra.Command{
		Use:   rootCommandName,
		Short: "qliksense cli tool",
		Long:  `qliksense cli tool provides functionality to perform operations on qliksense-k8s, qliksense operator, and kubernetes cluster`,
		Args:  cobra.ArbitraryArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if commandUsesContext(cmd.CommandPath()) {
				if err := p.SetUpQliksenseDefaultContext(); err != nil {
					panic(err)
				}
				pf := api.NewPreflightConfig(p.QliksenseHome)
				if err := pf.Initialize(); err != nil {
					panic(err)
				}
			}
		},
		SilenceUsage: true,
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func initConfig() {
	viper.SetEnvPrefix("QLIKSENSE")
	viper.AutomaticEnv()
}

func rootCmd(p *qliksense.Qliksense) *cobra.Command {
	cmd := getRootCmd(p)
	cobra.OnInitialize(initConfig)

	cmd.AddCommand(getInstallableVersionsCmd(p))
	cmd.AddCommand(pullQliksenseImages(p))
	cmd.AddCommand(pushQliksenseImages(p))
	cmd.AddCommand(about(p))
	// add version command
	cmd.AddCommand(versionCmd)

	// add operator command
	cmd.AddCommand(operatorCmd)
	//operatorCmd.AddCommand(operatorViewCmd(p))
	operatorCmd.AddCommand(operatorCrdCmd(p))
	operatorCmd.AddCommand(operatorControllerCmd(p))

	//add fetch command
	cmd.AddCommand(fetchCmd(p))

	// add install command
	cmd.AddCommand(installCmd(p))

	// add config command
	configCmd := configCmd(p)
	cmd.AddCommand(configCmd)
	/** disabling for now
	configCmd.AddCommand(configApplyCmd(p))
	**/
	configCmd.AddCommand(configViewCmd(p))

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// add the set-context config command as a sub-command to the app config command
	configCmd.AddCommand(setContextConfigCmd(p))

	// add the set profile/namespace/storageClassName/git-repository config command as a sub-command to the app config command
	configCmd.AddCommand(setOtherConfigsCmd(p))

	// add the set ### config command as a sub-command to the app config sub-command
	configCmd.AddCommand(setConfigsCmd(p))

	// add the set ### config command as a sub-command to the app config sub-command
	configCmd.AddCommand(setSecretsCmd(p))

	// add the list config command as a sub-command to the app config sub-command
	configCmd.AddCommand(listContextConfigCmd(p))

	// add the delete-context config command as a sub-command to the app config command
	configCmd.AddCommand(deleteContextConfigCmd(p))

	// add set-image-registry command as a sub-command to the app config sub-command
	configCmd.AddCommand(setImageRegistryCmd(p))

	// add clean-config-repo-patches command as a sub-command to the app config sub-command
	configCmd.AddCommand(cleanConfigRepoPatchesCmd(p))

	// open editor for config
	configCmd.AddCommand(configEditCmd(p))

	// add unset for config
	configCmd.AddCommand((unsetCmd(p)))

	// add uninstall command
	cmd.AddCommand(uninstallCmd(p))

	// add crds
	cmd.AddCommand(crdsCmd)
	crdsCmd.AddCommand(crdsViewCmd(p))
	crdsCmd.AddCommand(crdsInstallCmd(p))

	// add preflight commands
	preflightCmd := preflightCmd(p)
	preflightCmd.AddCommand(pfDnsCheckCmd(p))
	preflightCmd.AddCommand(pfK8sVersionCheckCmd(p))
	preflightCmd.AddCommand(pfAllChecksCmd(p))
	preflightCmd.AddCommand(pfMongoCheckCmd(p))
	preflightCmd.AddCommand(pfDeploymentCheckCmd(p))
	preflightCmd.AddCommand(pfServiceCheckCmd(p))
	preflightCmd.AddCommand(pfPodCheckCmd(p))
	preflightCmd.AddCommand(pfCreateRoleCheckCmd(p))
	preflightCmd.AddCommand(pfCreateRoleBindingCheckCmd(p))
	preflightCmd.AddCommand(pfCreateServiceAccountCheckCmd(p))
	preflightCmd.AddCommand(pfCreateAuthCheckCmd(p))
	preflightCmd.AddCommand(pfVerifyCAChainCmd(p))
	preflightCmd.AddCommand(pfCleanupCmd(p))

	cmd.AddCommand(preflightCmd)
	cmd.AddCommand(loadCrFile(p))
	cmd.AddCommand((applyCmd(p)))

	// add postflight command
	postflightCmd := postflightCmd(p)
	postflightCmd.AddCommand(postflightMigrationCheck(p))
	postflightCmd.AddCommand(AllPostflightChecks(p))

	cmd.AddCommand(postflightCmd)

	// add keys command
	cmd.AddCommand(keysCmd)
	keysCmd.AddCommand(keysRotateCmd(p))
	return cmd
}

func levenstein(cmd *cobra.Command) {
	cmd.SuggestionsMinimumDistance = 2
	if len(os.Args) > 1 {
		args := os.Args[1]
		suggest := cmd.SuggestionsFor(args)
		if len(suggest) > 0 {
			arg := []string{}
			for _, cm := range os.Args {
				arg = append(arg, cm)
			}
			if !strings.EqualFold(arg[1], suggest[0]) {
				arg[1] = suggest[0]
				out := ansi.NewColorableStdout()
				fmt.Fprintln(out, Green("Did you mean: "), Bold(strings.Join(arg, " ")), "?")
			}
		}
	}
}
