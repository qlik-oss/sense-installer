package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/qlik-oss/sense-installer/pkg/api"

	"github.com/qlik-oss/sense-installer/pkg/qliksense"

	"github.com/spf13/cobra"
)

const (
	preflightRelease     = "v0.9.26"
	preflightLinuxFile   = "preflight_linux_amd64.tar.gz"
	preflightMacFile     = "preflight_darwin_amd64.tar.gz"
	preflightWindowsFile = "preflight_windows_amd64.zip"
)

var preflightBaseURL = fmt.Sprintf("https://github.com/replicatedhq/troubleshoot/releases/download/%s/", preflightRelease)

func preflightCmd(q *qliksense.Qliksense) *cobra.Command {
	var configCmd = &cobra.Command{
		Use:   "preflight",
		Short: "perform preflight checks on the cluster",
		Long:  `perform preflight checks on the cluster`,
		Example: `qliksense preflight <preflight_check_to_run>
Usage:
qliksense preflight --all
qliksense preflight dns
qliksense preflight mongo
`,
	}
	return configCmd
}

func preflightCheckDnsCmd(q *qliksense.Qliksense) *cobra.Command {
	var preflightDnsCmd = &cobra.Command{
		Use:     "dns",
		Short:   "perform preflight dns check",
		Long:    `perform preflight dns check to check DNS connectivity status in the cluster`,
		Example: `qliksense preflight dns`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := DownloadPreflight(q)
			if err != nil {
				api.LogDebugMessage("There has been an error downloading preflight.. %+v", err)
			}
			return qliksense.PerformDnsCheck(q).RunE(cmd, args)
		},
	}
	return preflightDnsCmd
}

func DownloadPreflight(q *qliksense.Qliksense) error {
	api.LogDebugMessage("Entry: DownloadPreflight")
	var preflightUrl, preflightFile string
	platform := runtime.GOOS

	const PREFLIGHTDIRNAME = "preflight"
	preflightInstallDir := filepath.Join(q.QliksenseHome, PREFLIGHTDIRNAME)
	//preflightInstallPath = filepath.Join(q.QliksenseHome)

	if !api.DirExists(preflightInstallDir) {
		api.LogDebugMessage("%s does not exist, creating now\n", preflightInstallDir)
		if err := os.Mkdir(preflightInstallDir, os.ModePerm); err != nil {
			err = fmt.Errorf("Not able to create %s dir: %v", preflightInstallDir, err)
			log.Println(err)
			return err
		}
	}
	api.LogDebugMessage("%s exists", preflightInstallDir)

	if runtime.GOARCH != `amd64` {
		return fmt.Errorf("%s architecture is not supported", runtime.GOARCH)
	}

	switch platform {
	case "windows":
		preflightFile = preflightWindowsFile
	case "darwin":
		preflightFile = preflightMacFile
	case "linux":
		preflightFile = preflightLinuxFile
	default:
		err := fmt.Errorf("Unable to download the preflight executable for the underlying platform\n")
		return err
	}
	preflightUrl = fmt.Sprintf("%s%s", preflightBaseURL, preflightFile)
	err := api.DownloadFile(preflightUrl, preflightInstallDir, preflightFile)
	if err != nil {
		return err
	}
	// TODO: download support-bundle as well and unzip it, give it permissions
	// TODO: unzip both installers before using it
	fileToUntar := filepath.Join(preflightInstallDir, preflightFile)
	err = api.UntarPackage(preflightInstallDir, fileToUntar)
	api.LogDebugMessage("Exit: DownloadPreflight")
	return nil
}
