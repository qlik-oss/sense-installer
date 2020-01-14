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
	// "version_checks"

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

	// install mixins
	if _, err = installMixins(); err != nil {
		return err
	}

	// porterExe = "porter"
	if err = rootCmd(qliksense.New(porterExe)).Execute(); err != nil {
		return err
	}

	return nil
}

func installPorter() (string, error) {
	var (
		porterPermaLink = pkg.Version
		destination, homeDir,
		// mixin,
		// mixinOpts,
		qlikSenseHome, porterExe, ext string
		// mixinsVar                                                             = map[string]string{
		// 	"kustomize":  "-v 0.2-beta-3-0e19ca4 --url https://github.com/donmstewart/porter-kustomize/releases/download",
		// 	"qliksense":  "-v v0.11.0 --url https://github.com/qlik-oss/porter-qliksense/releases/download",
		// 	"exec":       "-v latest",
		// 	"kubernetes": "-v latest",
		// 	"helm":       "-v latest",
		// 	"azure":      "-v latest",
		// 	"terraform":  "-v latest",
		// 	"az":         "-v latest",
		// 	"aws":        "-v latest",
		// 	"gcloud":     "-v latest",
		// }
		// downloadMixins map[string]string
		downloadPorter bool
		err            error
		// cmd            *exec.Cmd
	)
	porterExe = "porter"
	if runtime.GOOS == "windows" {
		porterExe = porterExe + ".exe"
	}
	if qlikSenseHome = os.Getenv(qlikSenseHomeVar); qlikSenseHome == "" {
		if homeDir, err = homedir.Dir(); err != nil {
			return "", err
		}
		if homeDir, err = homedir.Expand(homeDir); err != nil {
			return "", err
		}
		qlikSenseHome = filepath.Join(homeDir, qlikSenseDirVar)
	}
	os.Setenv(porterHomeVar, qlikSenseHome)

	porterExe = filepath.Join(qlikSenseHome, porterExe)
	// // get Porter version from dependency.yaml
	// var porterVersion= getVersionFromDependencyYaml("org.qlik.operator.cli.porter.version.min")
	if _, err = os.Stat(qlikSenseHome); err != nil {
		if os.IsNotExist(err) /* ||  porterVersion > porterPermaLink */ {
			// porterPermaLink = porterVersion
			downloadPorter = true
		} else {
			return "", err
		}
	} else {
		if _, err = os.Stat(porterExe); err != nil {
			if os.IsNotExist(err) /* || getVersionFromDependencyYaml("org.qlik.operator.cli.porter.version.min") > porterPermaLink */ {
				// porterPermaLink = getVersionFromDependencyYaml("org.qlik.operator.cli.porter.version.min")
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

	// if _, err = os.Stat(filepath.Join(qlikSenseHome, mixinDirVar)); err != nil {
	// 	if os.IsNotExist(err) {
	// 		downloadMixins = mixinsVar
	// 	} else {
	// 		return "", err
	// 	}
	// } else {
	// 	downloadMixins = make(map[string]string)
	// 	for mixin, mixinOpts = range mixinsVar {
	// 		if _, err = os.Stat(filepath.Join(qlikSenseHome, mixinDirVar, mixin)); err != nil {
	// 			if os.IsNotExist(err) {
	// 				downloadMixins[mixin] = mixinOpts
	// 			} else {
	// 				return "", err
	// 			}
	// 		}
	// 	}
	// }
	// for mixin, mixinOpts = range downloadMixins {
	// 	cmd = exec.Command(porterExe, append([]string{"mixin", "install", mixin}, strings.Split(mixinOpts, " ")...)...)
	// 	cmd.Stdout = os.Stdout
	// 	cmd.Stderr = os.Stderr
	// 	if err = cmd.Run(); err != nil {
	// 		return "", err
	// 	}
	// }

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

func installMixins() (string, error) {
	var (
		mixin, mixinOpts, qlikSenseHome, porterExe string
		mixinsVar                                  = map[string]string{
			"kustomize":  "-v 0.2-beta-3-0e19ca4 --url https://github.com/donmstewart/porter-kustomize/releases/download",
			"qliksense":  "-v v0.11.0 --url https://github.com/qlik-oss/porter-qliksense/releases/download",
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
		// cmd            *exec.Cmd
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

	cmd = exec.Command(porterExe, append([]string{"mixin", "install", mixin}, strings.Split(mixinOpts, " ")...)...)
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
