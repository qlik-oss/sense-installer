package preflight

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/qliksense"

	"github.com/qlik-oss/sense-installer/pkg/api"
)

type QliksensePreflight struct {
	Q *qliksense.Qliksense
}

const (
	// preflight releases have the same version
	preflightRelease       = "v0.9.28"
	preflightLinuxFile     = "preflight_linux_amd64.tar.gz"
	preflightMacFile       = "preflight_darwin_amd64.tar.gz"
	preflightWindowsFile   = "preflight_windows_amd64.zip"
	PreflightChecksDirName = "preflight_checks"
	preflightFileName      = "preflight"
)

var preflightBaseURL = fmt.Sprintf("https://github.com/replicatedhq/troubleshoot/releases/download/%s/", preflightRelease)

func (qp *QliksensePreflight) DownloadPreflight() error {
	preflightExecutable := "preflight"
	if runtime.GOOS == "windows" {
		preflightExecutable += ".exe"
	}

	preflightInstallDir := filepath.Join(qp.Q.QliksenseHome, PreflightChecksDirName)

	exists, err := checkInstalled(preflightInstallDir, preflightExecutable)
	if err != nil {
		err = fmt.Errorf("There has been an error when trying to determine if preflight installer exists")
		log.Println(err)
		return err
	}
	if exists {
		// preflight exist, no need to download again.
		api.LogDebugMessage("Preflight already exists, proceeding to perform checks")
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

	preflightUrl, preflightFile, err := determinePlatformSpecificUrls(runtime.GOOS)
	if err != nil {
		err = fmt.Errorf("There was an error when trying to determine platform specific paths")
		return err
	}

	// Download Preflight
	err = downloadAndExplode(preflightUrl, preflightInstallDir, preflightFile)
	if err != nil {
		return err
	}
	fmt.Println("Downloaded Preflight")

	return nil
}

func checkInstalled(preflightInstallDir, preflightExecutable string) (bool, error) {
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

func downloadAndExplode(url, installDir, file string) error {
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

func determinePlatformSpecificUrls(platform string) (string, string, error) {

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

func initiateK8sOps(opr, namespace string) error {
	opr1 := strings.Fields(opr)
	err := api.KubectlDirectOps(opr1, namespace)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func invokePreflight(preflightCommand string, yamlFile *os.File) error {
	arguments := []string{}
	// check for 2nd character is ':' then take file location from after ':'
	tempYamlName := yamlFile.Name()
	if tempYamlName[1] == ':' {
		tempYamlName = tempYamlName[2:]
		api.LogDebugMessage("This is the Windows yaml file path after modification: %s", tempYamlName)
	}

	arguments = append(arguments, tempYamlName, "--interactive=false")
	cmd := exec.Command(preflightCommand, arguments...)

	sterrBuffer := &bytes.Buffer{}
	cmd.Stdout = sterrBuffer
	cmd.Stderr = sterrBuffer
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error when running preflight command: %v\n", err)
	}
	ind := strings.Index(sterrBuffer.String(), "---")
	output := sterrBuffer.String()
	if ind > -1 {
		output = fmt.Sprintf("%s\n%s", output[:ind], output[ind:])
	}
	fmt.Printf("%v\n", output)

	// Maybe good to retain this part in case we need to process the output in future.
	// We are going to look for the first occurance of PASS or FAIL from the end
	// there are also some space-like deceiving characters which are being hard to get by

	//outputArr := strings.Fields(strings.TrimSpace(output))
	//trackSuccess := false
	//trackPrg := false

	//for i := len(outputArr) - 1; i >= 0; i-- {
	//	if strings.TrimSpace(outputArr[i]) != "" {
	//		if outputArr[i] == "PASS" {
	//			trackSuccess = true
	//			trackPrg = true
	//		} else if outputArr[i] == "FAIL" {
	//			trackPrg = true
	//		}
	//	}
	//	if trackPrg {
	//		break
	//	}
	//}

	return nil
}
