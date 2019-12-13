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
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	porterURLBase   = "https://deislabs.blob.core.windows.net/porter"
	porterHomeVar   = "PORTER_HOME"
	porterDirVar    = ".porter"
	mixinDirVar     = "mixins"
	porterPermaLink = "latest"
	porterRuntime   = "porter-runtime"
)

type qlikSenseCmd struct {
	porterExe string
}

func (p *qlikSenseCmd) installPorter() error {
	var (
		homeDir, mixin, mixinOpts, porterHome, porterExe string
		mixinsVar                                        = map[string]string{
			"kustomize":  "-v 0.2-beta-3-0e19ca4 --url https://github.com/donmstewart/porter-kustomize/releases/download",
			"qliksense":  "-v v0.9.0 --url https://github.com/qlik-oss/porter-qliksense/releases/download",
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
	if porterHome = os.Getenv(porterHomeVar); porterHome == "" {
		if homeDir, err = homedir.Dir(); err != nil {
			return err
		}
		if homeDir, err = homedir.Expand(homeDir); err != nil {
			return err
		}
		porterHome = filepath.Join(homeDir, porterDirVar)
	}
	p.porterExe = filepath.Join(porterHome, porterExe)
	if _, err = os.Stat(porterHome); err != nil {
		if os.IsNotExist(err) {
			downloadPorter = true
		} else {
			return err
		}
	} else {
		if _, err = os.Stat(p.porterExe); err != nil {
			if os.IsNotExist(err) {
				downloadPorter = true
			} else {
				return err
			}
		}
	}
	if downloadPorter {
		os.Mkdir(porterHome, os.ModePerm)
		if err = DownloadFile(porterURLBase+"/"+porterPermaLink+"/porter-linux-amd64", filepath.Join(porterHome, porterRuntime)); err != nil {
			return err
		}
		if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
			if _, err = copy(filepath.Join(porterHome, porterRuntime), p.porterExe); err != nil {
				return err
			}
		} else {
			if err = DownloadFile(porterURLBase+"/"+porterPermaLink+"/"+"porter-"+runtime.GOOS+"-"+runtime.GOARCH, p.porterExe); err != nil {
				return err
			}
		}
	}

	if _, err = os.Stat(filepath.Join(porterHome, mixinDirVar)); err != nil {
		if os.IsNotExist(err) {
			downloadMixins = mixinsVar
		} else {
			return nil
		}
	} else {
		downloadMixins = make(map[string]string)
		for mixin, mixinOpts = range mixinsVar {
			if _, err = os.Stat(filepath.Join(porterHome, mixinDirVar, mixin)); err != nil {
				if os.IsNotExist(err) {
					downloadMixins[mixin] = mixinOpts
				} else {
					return err
				}
			}
		}
	}
	for mixin, mixinOpts = range downloadMixins {
		cmd = exec.Command(p.porterExe, "mixin", "install", mixin, mixinOpts)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return err
		}
	}
	return nil
	//err := DownloadFile(url, filename)

}

func DownloadFile(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
func (p *qlikSenseCmd) RootCmd() *cobra.Command {
	q := qliksense.New()
	cmd := &cobra.Command{
		Use:   "qliksense",
		Short: "Qliksense cli tool",
		Long: `qliksense cli tool provides a wrapper around the porter api as well as
		provides addition functionality`,
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
	}

	cmd.Flags().SetInterspersed(false)

	cobra.OnInitialize(initConfig)

	//cmd.AddCommand(installPorterMixin(q))

	// For qliksense overrides/commands

	cmd.AddCommand(pullQliksenseImages(q))
	cmd.AddCommand(porter(p))
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	return cmd
}

func InitAndExecute() error {

	var (
		q *qlikSenseCmd
	)
	q = new(qlikSenseCmd)
	q.installPorter()
	if err := q.RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}

func initConfig() {
	viper.SetEnvPrefix("QLIKSENSE")
	viper.AutomaticEnv()
}
