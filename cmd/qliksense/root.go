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
	qlikSenseHomeVar = "QLIKSENSE_HOME"
	qlikSenseDirVar  = ".qliksense"
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

	// add uninstall command
	cmd.AddCommand(uninstallCmd(p))
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
	type closeCommand struct {
		Name          string
		levensteinVal int
	}
	var lev closeCommand
	lev.Name = ""
	lev.levensteinVal = 10
	if len(os.Args) > 1 {
		args := os.Args[1]
		for _, ctx := range cmd.Commands() {
			val := *ctx
			if args == val.Name() {
				//found command
				return false
			} else {
				currVal := comparator([]rune(args), []rune(val.Name()))
				if currVal < lev.levensteinVal {
					lev.levensteinVal = currVal
					lev.Name = val.Name()
				}
				fmt.Println(currVal, val.Name())
			}
		}
		if lev.levensteinVal <= 4 {
			arg := []string{}
			for _, cm := range os.Args {
				arg = append(arg, cm)
			}
			arg[1] = lev.Name
			out := ansi.NewColorableStdout()
			fmt.Fprintln(out, chalk.Green.Color("Did you mean: "), chalk.Bold.TextStyle(strings.Join(arg, " ")), "?")
			return true
		}
	}
	return false
}

func comparator(str1, str2 []rune) int {
	s1len := len(str1)
	s2len := len(str2)
	column := make([]int, len(str1)+1)

	for y := 1; y <= s1len; y++ {
		column[y] = y
	}
	for x := 1; x <= s2len; x++ {
		column[0] = x
		lastkey := x - 1
		for y := 1; y <= s1len; y++ {
			oldkey := column[y]
			var incr int
			if str1[y-1] != str2[x-1] {
				incr = 1
			}
			column[y] = minimum(column[y]+1, column[y-1]+1, lastkey+incr)
			lastkey = oldkey
		}
	}
	return column[s1len]
}

func minimum(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
	} else {
		if b < c {
			return b
		}
	}
	return c
}
