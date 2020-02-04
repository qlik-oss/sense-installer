package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/qlik-oss/sense-installer/pkg"
	"github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"

	// "github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// To run this project in ddebug mode, run:
// export QLIKSENSE_DEBUG=true
// qliksense <command>

const (
	qlikSenseHomeVar = "QLIKSENSE_HOME"
	qlikSenseDirVar  = ".qliksense"

	qliksenseConfigFile  = "config.yaml"
	qliksenseContextsDir = "contexts"

	defaultQliksenseContext = "qliksense-default"
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
	if len(os.Args) == 1 {
		log.Debugf("QliksenseHomeDir: %s", qlikSenseHome)
		setUpQliksenseDefaultContext(qlikSenseHome)
		return nil
	}

	if err = rootCmd(qliksense.New(qlikSenseHome)).Execute(); err != nil {
		return err
	}

	return nil
}

func setUpQliksenseDefaultContext(qlikSenseHome string) {
	setUpQliksenseContext(qlikSenseHome, defaultQliksenseContext)
}

func setUpQliksenseContext(qlikSenseHome, contextName string) {
	qliksenseConfigFile := filepath.Join(qlikSenseHome, qliksenseConfigFile)
	var qliksenseConfig api.QliksenseConfig
	if !qliksense.FileExists(qliksenseConfigFile) {
		qliksenseConfig = qliksense.AddBaseQliksenseConfigs(qliksenseConfig, contextName)
	} else {
		qliksense.ReadFromFile(&qliksenseConfig, qliksenseConfigFile)
	}
	// creating a file in the name of the context if it does not exist/ opening it to append/modify content if it already exists

	qliksenseContextsDir1 := filepath.Join(qlikSenseHome, qliksenseContextsDir)
	if !qliksense.DirExists(qliksenseContextsDir1) {
		if err := os.Mkdir(qliksenseContextsDir1, 0700); err != nil {
			log.Fatalf("Not able to create the contexts/ dir: %v", err)
		}
	}
	log.Debug("Created contexts/")
	// creating contexts/qliksense-default.yaml file

	qliksenseContextFile := filepath.Join(qliksenseContextsDir1, contextName, contextName+".yaml")
	var qliksenseCR api.QliksenseCR

	if err := os.Mkdir(filepath.Join(qliksenseContextsDir1, contextName), 0700); err != nil {
		log.Fatalf("Not able to create the contexts/qliksense-default/ dir: %v", err)
	}
	log.Debug("Created contexts/qliksense-default/ directory")
	if !qliksense.FileExists(qliksenseContextFile) {
		qliksenseCR = qliksense.AddCommonConfig(qliksenseCR, contextName)
		log.Debugf("Added Context: %s", contextName)
	} else {
		qliksense.ReadFromFile(&qliksenseCR, qliksenseContextFile)
	}

	qliksense.WriteToFile(&qliksenseCR, qliksenseContextFile)
	ctxTrack := false
	if len(qliksenseConfig.Spec.Contexts) > 0 {
		for _, ctx := range qliksenseConfig.Spec.Contexts {
			if ctx.Name == contextName {
				ctx.CRLocation = qliksenseContextFile
				ctxTrack = true
				break
			}
		}
	}
	if !ctxTrack {
		qliksenseConfig.Spec.Contexts = append(qliksenseConfig.Spec.Contexts, api.Context{
			Name:       contextName,
			CRLocation: qliksenseContextFile,
		})
	}
	qliksenseConfig.Spec.CurrentContext = contextName
	qliksense.WriteToFile(&qliksenseConfig, qliksenseConfigFile)
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
	os.Mkdir(qlikSenseHome, os.ModePerm)
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
		Long: `qliksense cli tool provides a wrapper around the porter api as well as
		provides addition functionality`,
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
	}

	cmd.Flags().SetInterspersed(false)

	cobra.OnInitialize(initConfig)

	// For qliksense overrides/commands

	cmd.AddCommand(pullQliksenseImages(p))
	cmd.AddCommand(about(p))
	// add version command
	cmd.AddCommand(versionCmd)

	// add operator command
	cmd.AddCommand(operatorCmd)
	operatorCmd.AddCommand(operatorViewCmd(p))
	//add fetch command
	cmd.AddCommand(fetchCmd(p))

	// add install command
	cmd.AddCommand(installCmd(p))

	// add config command
	cmd.AddCommand(configCmd)
	configCmd.AddCommand(configApplyCmd(p))
	configCmd.AddCommand(configViewCmd(p))

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// create 'config' commands
	qliksenseConfigCmd := qliksenseConfigCmds(p)
	// add app config commands to root command
	cmd.AddCommand(qliksenseConfigCmd)

	// create set-context config sub command
	setContextCmd := setContextConfigCmd(p)
	// add the set-context config command as a sub-command to the app config command
	qliksenseConfigCmd.AddCommand(setContextCmd)

	// create set profile/namespace/storageClassName/git-repository config sub-command
	setOtherConfigsCmd := setOtherConfigsCmd(p)
	// add the set profile/namespace/storageClassName/git-repository config command as a sub-command to the app config command
	qliksenseConfigCmd.AddCommand(setOtherConfigsCmd)

	// create setConfigs sub-command
	setConfigsCmd := setConfigsCmd(p)
	// add the set ### config command as a sub-command to the app config sub-command
	qliksenseConfigCmd.AddCommand(setConfigsCmd)

	// create setConfigs sub-command
	setSecretsCmd := setSecretsCmd(p)
	// add the set ### config command as a sub-command to the app config sub-command
	qliksenseConfigCmd.AddCommand(setSecretsCmd)

	// create view config sub-command
	viewCmd := viewConfigCmd(p)
	// add the view config command as a sub-command to the app config sub-command
	qliksenseConfigCmd.AddCommand(viewCmd)

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
