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
	preflightRelease         = "v0.9.26"
	preflightLinuxFile       = "preflight_linux_amd64.tar.gz"
	preflightMacFile         = "preflight_darwin_amd64.tar.gz"
	preflightWindowsFile     = "preflight_windows_amd64.zip"
	supportbundleWindowsFile = "support-bundle_windows_amd64.zip"
	supportbundleLinuxFile   = "support-bundle_linux_amd64.tar.gz"
	supportbundleMacFile     = "support-bundle_darwin_amd64.tar.gz"
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
			err := DownloadPreflightAndSupportBundle(q)
			if err != nil {
				err = fmt.Errorf("There has been an error downloading preflight/support-bundle: %+v", err)
				log.Println(err)
				return err
			}
			return qliksense.PerformDnsCheck(q).RunE(cmd, args)
		},
	}
	return preflightDnsCmd
}

func DownloadPreflightAndSupportBundle(q *qliksense.Qliksense) error {
	const PREFLIGHTCHECKSDIRNAME = "preflight_checks"
	const preflightExecutable = "preflight"
	const supportbundleExecutable = "support-bundle"

	preflightInstallDir := filepath.Join(q.QliksenseHome, PREFLIGHTCHECKSDIRNAME)
	platform := runtime.GOOS

	exists, err := CheckInstalled(preflightInstallDir, preflightExecutable, supportbundleExecutable)
	if err != nil {
		err = fmt.Errorf("There has been an error when trying to determine the existence of preflight and support-bundle installers")
		log.Println(err)
		return err
	}
	if exists {
		// preflight and support-bundle exist, no need to download again.
		api.LogDebugMessage("Preflight and support-bundle already exist, proceeding to perform checks")
		return nil
	}

	// Create the Preflight-check directory, download and install preflight and support-bundle
	if !api.DirExists(preflightInstallDir) {
		api.LogDebugMessage("%s does not exist, creating now\n", preflightInstallDir)
		if err := os.Mkdir(preflightInstallDir, os.ModePerm); err != nil {
			err = fmt.Errorf("Not able to create %s dir: %v", preflightInstallDir, err)
			log.Println(err)
			return nil
		}
	}
	api.LogDebugMessage("Preflight-checks install Dir: %s exists", preflightInstallDir)

	preflightUrl, preflightFile, supportbundleUrl, supportBundleFile, err := DeterminePlatformSpecificPaths(platform, preflightInstallDir)
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

	// Download support-bundle
	err = DownloadAndExplode(supportbundleUrl, preflightInstallDir, supportBundleFile)
	if err != nil {
		return err
	}
	fmt.Println("Downloaded Support bundle")

	return nil
}

func CheckInstalled(preflightInstallDir, preflightInstaller, supportbundleInstaller string) (bool, error) {
	installerExists := true
	if api.DirExists(preflightInstallDir) {
		if !api.FileExists(preflightInstaller) {
			installerExists = false
			api.LogDebugMessage("Preflight install directory exists, but preflight installer does not exist")
		}
		if !api.FileExists(supportbundleInstaller) {
			installerExists = false
			api.LogDebugMessage("Preflight install directory exists, but support-bundle installer does not exist")
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

func DeterminePlatformSpecificPaths(platform, preflightInstallDir string) (string, string, string, string, error) {

	var preflightUrl, preflightFile, supportbundleUrl, supportBundleFile string

	if runtime.GOARCH != `amd64` {
		err := fmt.Errorf("%s architecture is not supported", runtime.GOARCH)
		return "", "", "", "", err
	}

	switch platform {
	case "windows":
		preflightFile = preflightWindowsFile
		supportBundleFile = supportbundleWindowsFile
	case "darwin":
		preflightFile = preflightMacFile
		supportBundleFile = supportbundleMacFile
	case "linux":
		preflightFile = preflightLinuxFile
		supportBundleFile = supportbundleLinuxFile
	default:
		err := fmt.Errorf("Unable to download the preflight executable for the underlying platform\n")
		return "", "", "", "", err
	}
	preflightUrl = fmt.Sprintf("%s%s", preflightBaseURL, preflightFile)
	supportbundleUrl = fmt.Sprintf("%s%s", preflightBaseURL, supportBundleFile)

	return preflightUrl, preflightFile, supportbundleUrl, supportBundleFile, nil
}
