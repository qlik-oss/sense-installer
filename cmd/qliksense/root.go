package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/mitchellh/go-homedir"
	"github.com/qlik-oss/sense-installer/pkg"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"

	// "github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// To run this project in ddebug mode, run:
// export QLIKSENSE_DEBUG=true
// qliksense <command>

const (
	winOS            = "porter-windows-amd64.exe"
	linuxOS          = "porter-linux-amd64"
	macOS            = "porter-darwin-amd64"
	porterHomeVar    = "PORTER_HOME"
	qlikSenseHomeVar = "QLIKSENSE_HOME"
	qlikSenseDirVar  = ".qliksense"
	mixinDirVar      = "mixins"
	porterRuntime    = "porter-runtime"

	qliksenseConfigFile  = "config.yaml"
	qliksenseContextsDir = "contexts"
)

func initAndExecute() error {
	var (
		porterExe, qlikSenseHome string
		err                      error
	)

	// setting logrus configs
	logrusConfigs()

	porterExe, qlikSenseHome, err = setUpPaths()
	if err != nil {
		log.Fatal(err)
	}

	if err = rootCmd(qliksense.New(porterExe, qlikSenseHome)).Execute(); err != nil {
		return err
	}

	// create dirs and appropriate files for setting up contexts
	log.Debugf("QliksenseHomeDir: %s", qlikSenseHome)
	setUpContext(qlikSenseHome)
	return nil
}

func logrusConfigs() {
	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(Formatter)
}

func setUpContext(qlikSenseHome string) {
	// create config.yaml and populate it

	// create contexts/ -> qliksense-default/ -> qliksense-default.yaml
	// func createContextSetup(contextname) {
	// qliksense-default/ -> qliksense-default.yaml
	// we know at this point that qliksense home dir exists, we will proceed to create other necessary dirs and files

	if !fileExists(filepath.Join(qlikSenseHome, qliksenseConfigFile)) {
		log.Debugf("config.yaml doesnt exist, creating it now...")
		f, err := os.Create(filepath.Join(qlikSenseHome, qliksenseConfigFile))
		if err != nil {
			panic(err)
		}
		log.Debugf("File created: %s", filepath.Join(qlikSenseHome, qliksenseConfigFile))
		defer f.Close()

		// TO-DO: write default configs into config.yaml
		var qliksenseConfig *qliksense.QliksenseConfig
		qliksenseConfig.AddBaseQliksenseConfigs()
		x, err := yaml.Marshal(qliksenseConfig)
		if err != nil {
			log.Fatalf("An error occurred during marshalling config: %v", err)
		}
		log.Debugf("Marshalled yaml:\n%s\nWriting to file...", x)

		// creating a file in the name of the context if it does not exist/ opening it to append/modify content if it already exists
		var err1 error
		os.MkdirAll(qliksense.QliksenseConfigContextHome, os.ModePerm)
		f, err1 = os.OpenFile(filepath.Join(qlikSenseHome, qliksense.QliksenseConfigContextHome), os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err1 != nil {
			panic(err1)
		}

		defer f.Close()

		numBytes, err2 := f.Write(x)
		if err2 != nil {
			panic(err2)
		}
		log.Debugf("wrote %d bytes\n", numBytes)

		// create contexts/
		if err := os.Mkdir(filepath.Join(qlikSenseHome, qliksenseContextsDir), os.ModeDir); err != nil {
			log.Debug("Not able to create the contexts/ dir")
		} else {
			log.Debug("Created the contexts/")
		}
	} else {
		log.Error("config.yaml already exists..")
	}
}
func setUpPaths() (string, string, error) {
	var (
		porterExe, homeDir, qlikSenseHome string
		err                               error
	)
	porterExe = "porter"
	if runtime.GOOS == "windows" {
		porterExe = porterExe + ".exe"
	}
	if qlikSenseHome = os.Getenv(qlikSenseHomeVar); qlikSenseHome == "" {
		if homeDir, err = homedir.Dir(); err != nil {
			return "", "", err
		}
		if homeDir, err = homedir.Expand(homeDir); err != nil {
			return "", "", err
		}
		qlikSenseHome = filepath.Join(homeDir, qlikSenseDirVar)
	}
	os.Setenv(porterHomeVar, qlikSenseHome)

	porterExe = filepath.Join(qlikSenseHome, porterExe)
	return porterExe, qlikSenseHome, nil
}

func installPorter(qlikSenseHome, porterExe string) (string, error) {
	var (
		destination string
		// downloadPorter    = true
		porterDownloadURL string
		err               error
	)

	// if downloadPorter {
	os.Mkdir(qlikSenseHome, os.ModePerm)
	destination = filepath.Join(qlikSenseHome, porterRuntime)
	// construct url to download porter from
	porterDownloadURL = constructPorterURL(runtime.GOOS)

	if (runtime.GOOS == "linux" && runtime.GOARCH == "amd64") || runtime.GOOS == "darwin" {
		if err = downloadFile(porterDownloadURL, destination); err != nil {
			return "", err
		}
		os.Chmod(destination, 0755)
		if _, err = copy(destination, porterExe); err != nil {
			return "", err
		}
		os.Chmod(porterExe, 0755)
	} else if runtime.GOOS == "windows" {
		if err = downloadFile(porterDownloadURL, porterExe); err != nil {
			return "", err
		}
		os.Chmod(porterExe, 0755)
	}
	return porterExe, nil
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of qliksense cli",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s (%s, %s)\n", pkg.Version, pkg.Commit, pkg.CommitDate)
	},
}

func constructPorterURL(runtimeOS string) string {
	// FYI: Porter does not support other architectures other than amd64
	const (
		porterURLBase = "https://cdn.deislabs.io/porter/"
		winOS         = "porter-windows-amd64.exe"
		linuxOS       = "porter-linux-amd64"
		macOS         = "porter-darwin-amd64"
	)
	var url, version string
	version = retrievePorterVersion()
	if runtimeOS == "linux" {
		url = porterURLBase + version + "/" + linuxOS
	} else if runtimeOS == "windows" {
		url = porterURLBase + version + "/" + winOS
	} else if runtimeOS == "darwin" {
		url = porterURLBase + version + "/" + macOS
	}
	return url
}

func retrievePorterVersion() string {
	type apiInfo struct {
		TagName string `json:"tag_name,omitempty"`
		Name    string `json:"name,omitempty"`
	}
	const porterRepoURL = "https://api.github.com/repos/deislabs/porter/releases/latest"

	resp, err := http.Get(porterRepoURL)
	if err != nil {
		fmt.Printf("Error occurred while retrieving porter version info: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("response status was not OK while retrieving porter version info\n")
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error occurred while reading porter version info: %v\n", err)
		return ""
	}
	result := &apiInfo{}
	err = json.Unmarshal(body, result)
	if err != nil {
		fmt.Printf("Error occurred while unmarshalling porter version info: %v\n", err)
		return ""
	}
	fmt.Printf("Porter Version: %s\n", result.Name)
	return result.Name
}

func installMixins(porterExe, qlikSenseHome string) (string, error) {
	var (
		mixin, mixinOpts string
		mixinsVar        = map[string]string{
			"kustomize":  "-v 0.2-beta-3-0e19ca4 --url https://github.com/donmstewart/porter-kustomize/releases/download",
			"qliksense":  "-v v0.16.0 --url https://github.com/qlik-oss/porter-qliksense/releases/download",
			"exec":       "-v latest",
			"kubernetes": "-v latest",
			"helm":       "-v latest",
			"azure":      "-v latest",
			"terraform":  "-v latest",
			"az":         "-v latest",
			"aws":        "-v latest",
			"gcloud":     "-v latest",
		}
		downloadMixins map[string]string
		err            error
	)
	if _, err = os.Stat(filepath.Join(qlikSenseHome, mixinDirVar)); err != nil {
		if os.IsNotExist(err) {
			downloadMixins = mixinsVar
		} else {
			return "", err
		}
	} else {
		downloadMixins = make(map[string]string)
		for mixin, mixinOpts = range mixinsVar {
			if _, err = os.Stat(filepath.Join(qlikSenseHome, mixinDirVar, mixin)); err != nil {
				if os.IsNotExist(err) {
					downloadMixins[mixin] = mixinOpts
				} else {
					return "", err
				}
			}
		}
	}
	for mixin, mixinOpts = range downloadMixins {
		if _, err = installMixin(porterExe, mixin, mixinOpts); err != nil {
			return "", err
		}
	}
	return "", err
}

func installMixin(porterExe, mixin, mixinOpts string) (string, error) {
	var cmd *exec.Cmd

	args := []string{"mixin", "install", mixin}
	if mixinOpts != "" {
		args = append(args, strings.Fields(mixinOpts)...)
	}
	cmd = exec.Command(porterExe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return "", nil
}

func rootCmd(p *qliksense.Qliksense) *cobra.Command {
	var (
		cmd, porterCmd, alias *cobra.Command
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
	porterCmd = porter(p)
	cmd.AddCommand(porterCmd)
	for _, alias = range buildAliasCommands(porterCmd, p) {
		cmd.AddCommand(alias)
	}
	// add version command
	cmd.AddCommand(versionCmd)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// create 'config' commands
	qliksenseConfigCmd := qliksenseConfigCmds(p)
	// add app config commands to root command
	cmd.AddCommand(qliksenseConfigCmd)

	// create set-context config sub command
	log.Debug("About to start set-context config cmds")
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
	logDebugMessage("Porter download link: %s\n", url)
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
