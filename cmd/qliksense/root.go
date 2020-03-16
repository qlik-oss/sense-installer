package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	ansi "github.com/mattn/go-colorable"
	"github.com/mitchellh/go-homedir"
	"github.com/qlik-oss/sense-installer/pkg"
	"github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

// To run this project in ddebug mode, run:
// export QLIKSENSE_DEBUG=true
// qliksense <command>

const (
	qlikSenseHomeVar        = "QLIKSENSE_HOME"
	qlikSenseDirVar         = ".qliksense"
	keepPatchFilesFlagName  = "keep-config-repo-patches"
	keepPatchFilesFlagUsage = "Keep config repo patch files (for debugging)"
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
	api.LogDebugMessage("QliksenseHomeDir: %s", qlikSenseHome)

	qliksenseClient, err := qliksense.New(qlikSenseHome)
	if err != nil {
		return err
	}
	qliksenseClient.SetUpQliksenseDefaultContext()
	cmd := rootCmd(qliksenseClient)
	//levenstein checks
	if levenstein(cmd) == false {
		if err := cmd.Execute(); err != nil {
			return err
		}
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
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s (%s, %s)\n", pkg.Version, pkg.Commit, pkg.CommitDate)
	},
}

func rootCmd(p *qliksense.Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)

	cmd = &cobra.Command{
		Use:   "qliksense",
		Short: "Qliksense cli tool",
		Long:  `qliksense cli tool provides functionality to perform operations on qliksense-k8s, qliksense operator, and kubernetes cluster`,
		Args:  cobra.ArbitraryArgs,
	}

	cmd.Flags().SetInterspersed(false)

	cobra.OnInitialize(initConfig)

	// For qliksense overrides/commands

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
	configCmd.AddCommand(configApplyCmd(p))
	configCmd.AddCommand(configViewCmd(p))

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	//add upgrade command
	cmd.AddCommand(upgradeCmd(p))

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

	// add uninstall command
	cmd.AddCommand(uninstallCmd(p))

	// add crds
	cmd.AddCommand(crdsCmd)
	crdsCmd.AddCommand(crdsViewCmd(p))
	crdsCmd.AddCommand(crdsInstallCmd(p))

	// add preflight command
	preflightCmd := preflightCmd(p)
	cmd.AddCommand(preflightCmd)
	preflightCmd.AddCommand(preflightCheckDnsCmd(p))
	//preflightCmd.AddCommand(preflightCheckMongoCmd(p))
	//preflightCmd.AddCommand(preflightCheckAllCmd(p))

	return cmd
}

func initConfig() {
	viper.SetEnvPrefix("QLIKSENSE")
	viper.AutomaticEnv()
}

func downloadFile(url string, filepath string) error {
	var (
		out  *os.File
		err  error
		resp *http.Response
	)
	// Create the file
	if out, err = os.Create(filepath); err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	if resp, err = http.Get(url); err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}

	return nil
}

func copy(src, dst string) (int64, error) {
	var (
		source, destination *os.File
		sourceFileStat      os.FileInfo
		err                 error
		nBytes              int64
	)
	if sourceFileStat, err = os.Stat(src); err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	if source, err = os.Open(src); err != nil {
		return 0, err
	}
	defer source.Close()

	if destination, err = os.Create(dst); err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err = io.Copy(destination, source)
	return nBytes, err
}

func levenstein(cmd *cobra.Command) bool {
	cmd.SuggestionsMinimumDistance = 4
	if len(os.Args) > 1 {
		args := os.Args[1]
		for _, ctx := range cmd.Commands() {
			val := *ctx
			if args == val.Name() {
				//found command
				return false
			}
		}
		suggest := cmd.SuggestionsFor(os.Args[1])
		if len(suggest) > 0 {
			arg := []string{}
			for _, cm := range os.Args {
				arg = append(arg, cm)
			}
			arg[1] = suggest[0]
			out := ansi.NewColorableStdout()
			fmt.Fprintln(out, chalk.Green.Color("Did you mean: "), chalk.Bold.TextStyle(strings.Join(arg, " ")), "?")
			return true
		}
	}
	return false
}
