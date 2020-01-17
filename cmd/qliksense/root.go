package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/qlik-oss/sense-installer/pkg"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	//	porterURLBase   = "https://deislabs.blob.core.windows.net/porter"
	porterURLBase    = "https://github.com/qlik-oss/sense-installer/releases/download"
	porterHomeVar    = "PORTER_HOME"
	qlikSenseHomeVar = "QLIKSENSE_HOME"
	qlikSenseDirVar  = ".qliksense"
	mixinDirVar      = "mixins"
	porterRuntime    = "porter-runtime"
)

func initAndExecute() error {
	var (
		porterExe string
		err       error
	)
	if porterExe, err = installPorter(); err != nil {
		return err
	}
	if err := rootCmd(qliksense.New(porterExe)).Execute(); err != nil {
		return err
	}

	return nil
}

func installPorter() (string, error) {
	var (
		porterPermaLink = pkg.Version
		//porterPermaLink                                          = "v0.3.0"
		destination, mixin, mixinOpts, qlikSenseHome, porterExe, ext string
		mixinsVar                                                    = map[string]string{
			"kustomize":  "-v 0.2-beta-3-0e19ca4 --url https://github.com/donmstewart/porter-kustomize/releases/download",
			"qliksense":  "-v v0.14.0 --url https://github.com/qlik-oss/porter-qliksense/releases/download",
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
		downloadPorter bool
		err            error
		cmd            *exec.Cmd
	)
	porterExe = "porter"
	if runtime.GOOS == "windows" {
		porterExe = porterExe + ".exe"
	}
	if qlikSenseHome, err = getQliksenseHomeDir(); err != nil {
		return "", err
	}
	os.Setenv(porterHomeVar, qlikSenseHome)
	//TODO: Check if porter version is one alreadu is one for this build
	porterExe = filepath.Join(qlikSenseHome, porterExe)
	if _, err = os.Stat(qlikSenseHome); err != nil {
		if os.IsNotExist(err) {
			downloadPorter = true
		} else {
			return "", err
		}
	} else {
		if _, err = os.Stat(porterExe); err != nil {
			if os.IsNotExist(err) {
				downloadPorter = true
			} else {
				return "", err
			}
		}
	}
	if downloadPorter {
		os.Mkdir(qlikSenseHome, os.ModePerm)
		destination = filepath.Join(qlikSenseHome, porterRuntime)
		if err = downloadFile(porterURLBase+"/"+porterPermaLink+"/porter-linux-amd64", destination); err != nil {
			return "", err
		}
		os.Chmod(destination, 0755)
		if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
			if _, err = copy(filepath.Join(qlikSenseHome, porterRuntime), porterExe); err != nil {
				return "", err
			}
			os.Chmod(porterExe, 0755)
		} else {
			if runtime.GOOS == "windows" {
				ext = ".exe"
			}
			if err = downloadFile(porterURLBase+"/"+porterPermaLink+"/"+"porter-"+runtime.GOOS+"-"+runtime.GOARCH+ext, porterExe); err != nil {
				return "", err
			}
			os.Chmod(porterExe, 0755)
		}
	}

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
		cmd = exec.Command(porterExe, append([]string{"mixin", "install", mixin}, strings.Split(mixinOpts, " ")...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return "", err
		}
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

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the qliksense loacal cache",
	Long:  `Remove the everything from ~/.qliksense/cache directory`,
	Run: func(cmd *cobra.Command, args []string) {
		qsHome, err := getQliksenseHomeDir()
		if err != nil {
			fmt.Println("Cannot find qliksense home diretory")
			return
		}
		cacheDir := filepath.Join(qsHome, "cache")
		if _, err = os.Stat(cacheDir); err != nil {
			// cache directory not exist
			fmt.Println("Cache Cleaned")
			return
		}
		if err = os.RemoveAll(cacheDir); err != nil {
			fmt.Println("cannot remove cache", err)
			return
		}
		fmt.Println("Cache Cleaned")
	},
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
	// add cache command
	var cahcheCommand = &cobra.Command{Use: "cache", Short: "Perform operations on cache"}
	cmd.AddCommand(cahcheCommand)
	cahcheCommand.AddCommand(cacheClearCmd)

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

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

func getQliksenseHomeDir() (string, error) {
	var qlikSenseHome string
	if qlikSenseHome = os.Getenv(qlikSenseHomeVar); qlikSenseHome == "" {
		var homeDir string
		var err error
		if homeDir, err = homedir.Dir(); err != nil {
			return "", err
		}
		if homeDir, err = homedir.Expand(homeDir); err != nil {
			return "", err
		}
		qlikSenseHome = filepath.Join(homeDir, qlikSenseDirVar)
	}
	return qlikSenseHome, nil
}
