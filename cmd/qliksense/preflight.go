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
	// preflight releases have the same version
	preflightRelease       = "v0.9.26"
	preflightLinuxFile     = "preflight_linux_amd64.tar.gz"
	preflightMacFile       = "preflight_darwin_amd64.tar.gz"
	preflightWindowsFile   = "preflight_windows_amd64.zip"
	PreflightChecksDirName = "preflight_checks"
)

var preflightBaseURL = fmt.Sprintf("https://github.com/replicatedhq/troubleshoot/releases/download/%s/", preflightRelease)

func preflightCmd(q *qliksense.Qliksense) *cobra.Command {
	var configCmd = &cobra.Command{
		Use:   "preflight",
		Short: "perform preflight checks on the cluster",
		Long:  `perform preflight checks on the cluster`,
		Example: `qliksense preflight <preflight_check_to_run>
Usage:
qliksense preflight dns
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
				err = fmt.Errorf("There has been an error downloading preflight: %+v", err)
				log.Println(err)
				return err
			}
			return qliksense.PerformDnsCheck(q).RunE(cmd, args)
		},
	}
	return preflightDnsCmd
}

func DownloadPreflight(q *qliksense.Qliksense) error {
	const preflightExecutable = "preflight"

	preflightInstallDir := filepath.Join(q.QliksenseHome, PreflightChecksDirName)
	platform := runtime.GOOS

	exists, err := CheckInstalled(preflightInstallDir, preflightExecutable)
	if err != nil {
		err = fmt.Errorf("There has been an error when trying to determine if preflight installer exists")
		log.Println(err)
		return err
	}
	if exists {
		// preflight exist, no need to download again.
		api.LogDebugMessage("Preflight already exist, proceeding to perform checks")
		return nil
	}

	// Create the Preflight-check directory, download and install preflight
	if !api.DirExists(preflightInstallDir) {
		api.LogDebugMessage("%s does not exist, creating now\n", preflightInstallDir)
		if err := os.Mkdir(preflightInstallDir, os.ModePerm); err != nil {
			err = fmt.Errorf("Not able to create %s dir: %v", preflightInstallDir, err)
			log.Println(err)
			return nil
		}
	}
	api.LogDebugMessage("Preflight-checks install Dir: %s exists", preflightInstallDir)

	preflightUrl, preflightFile, err := DeterminePlatformSpecificUrls(platform)
	if err != nil {
		err = fmt.Errorf("There was an error when trying to determine platform specific paths")
		return err
	}

	// Download Preflight
	err = DownloadAndExplode(preflightUrl, preflightInstallDir, preflightFile)
	if err != nil {
		return err
	}
	fmt.Println("Downloaded Preflight")

	return nil
}

func CheckInstalled(preflightInstallDir, preflightExecutable string) (bool, error) {
	installerExists := true
	preflightInstaller := filepath.Join(preflightInstallDir, preflightExecutable)
	if api.DirExists(preflightInstallDir) {
		if !api.FileExists(preflightInstaller) {
			installerExists = false
			api.LogDebugMessage("Preflight install directory exists, but preflight installer does not exist")
		}
	} else {
		installerExists = false
	}
	return installerExists, nil
}

func DownloadAndExplode(url, installDir, file string) error {
	err := api.DownloadFile(url, installDir, file)
	if err != nil {
		return err
	}
	api.LogDebugMessage("Downloaded File: %s", file)

	fileToUntar := filepath.Join(installDir, file)
	api.LogDebugMessage("File to explode: %s", file)

	err = api.ExplodePackage(installDir, fileToUntar)
	if err != nil {
		return err
	}

	return nil
}

func DeterminePlatformSpecificUrls(platform string) (string, string, error) {

	var preflightUrl, preflightFile string

	if runtime.GOARCH != `amd64` {
		err := fmt.Errorf("%s architecture is not supported", runtime.GOARCH)
		return "", "", err
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
		return "", "", err
	}
	preflightUrl = fmt.Sprintf("%s%s", preflightBaseURL, preflightFile)

	return preflightUrl, preflightFile, nil
}
