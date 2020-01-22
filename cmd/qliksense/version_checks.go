package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/qlik-oss/sense-installer/pkg"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"gopkg.in/yaml.v2"
)

var (
	dependenciesFile             = "dependencies.yaml"
	updateMixin, updateComponent bool
	currentPorterVersion         string
	mixinURLs                    = map[string]string{
		"qliksense": "--url https://github.com/qlik-oss/porter-qliksense/releases/download",
		"kustomize": "--url https://github.com/donmstewart/porter-kustomize/releases/",
	}
	mixinsVar = map[string]string{
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
)

func checkMinVersion(tag string, q *qliksense.Qliksense) {

	logDebugMessage("Starting version checks\n")

	dependencies := map[string]string{}
	// check if tag is empty or not
	var err error
	if len(strings.TrimSpace(tag)) > 0 {
		// --tag exists
		logDebugMessage("Input tag: %s, %s\n", tag, strings.Replace(tag, "bundle", "invocation", 1))
		// Pull the image and store labels in a map
		dependencies, err = q.PullImage(strings.Replace(tag, "bundle", "invocation", 1))
		logDebugMessage("\nDependencies map from the inspected image: %v\n", dependencies)
		if err != nil {
			log.Fatalf("unable to pull the requested image: %v", err)
		}

	} else {
		// Tag is empty, hence looking for dependenciesFile.yaml, exit if this file is not present
		if fileExists(dependenciesFile) {
			// read the dependencies.yaml and store into a map
			yamlFile, err := ioutil.ReadFile(dependenciesFile)
			if err != nil {
				logDebugMessage("Exit: CheckMinVersion()\n")
				log.Fatalf("Error reading from source: %s\n", err)
			}
			err = yaml.Unmarshal(yamlFile, &dependencies)
			if err != nil {
				logDebugMessage("Exit: CheckMinVersion()\n")
				log.Fatalf("Error when parsing from source: %s\n", err)
			}
			logDebugMessage("Dependencies map from the given yaml: %v\n", dependencies)
		}
	}
	if len(dependencies) > 0 {
		for k, v := range dependencies {
			if strings.Contains(k, ".mixin.") {
				dependencies[k] = fmt.Sprintf("-v %s", v)
			}
		}
		// CLI check
		checkCLIVersion(dependencies)
		// Porter check
		checkPorterVersion(dependencies, q)
		// Mixins check
		checkMixinVersion(dependencies, q)
	} else {
		log.Fatalf("Not able to infer dependencies, hence exiting")
	}
	logDebugMessage("Completed version checks\n")
}

func checkMixinVersion(dependencies map[string]string, q *qliksense.Qliksense) {
	var tmp string
	logDebugMessage("------------ Mixins version check -----------")
	currentMixinVersions, err := retrieveCurrentInstalledMixinVersions(q)
	if err != nil {
		log.Fatal(err)
	}
	for k := range mixinsVar {
		tmp, _ = dependencies[fmt.Sprintf("org.qlik.operator.mixin.%s.version.min", k)]
		if tmp == "" {
			continue
		}
		shouldUpdateMixin := false
		mixinVersion, ok := currentMixinVersions[k]
		if !ok {
			shouldUpdateMixin = true
		} else {
			shouldUpdateMixin = versionCheck(fmt.Sprintf("Mixin %s", k), mixinVersion, tmp)
		}
		// if tmp is not empty and mixin requires download and install
		if shouldUpdateMixin {
			fmt.Println("Downloading a newer version of mixin:", k)
			// download and install the new mixin
			mURL, ok := mixinURLs[k]
			if ok {
				tmp = fmt.Sprintf("%s %s", tmp, mURL)
			}
			if _, err = installMixin(q.PorterExe, k, tmp); err != nil {
				logDebugMessage("Completed version checks\n")
				log.Fatalf("Error installing mixin %s: %s\n", k, err)
			}
		}
	}
}

func checkPorterVersion(dependencies map[string]string, q *qliksense.Qliksense) {
	// Infer info about the min porter version
	var porterVersionFromDependencies, tmp string
	var err error
	logDebugMessage("------------ Porter version check -----------")
	tmp, _ = dependencies["org.qlik.operator.cli.porter.version.min"]
	if len(tmp) != 0 {
		porterVersionFromDependencies = tmp
	}
	logDebugMessage("Porter version from dependencies map: %v\n", porterVersionFromDependencies)

	// check porter version
	currentPorterVersion, err = determineCurrentPorterVersion(q)
	if err != nil {
		log.Println("warning:", err)
	}
	logDebugMessage("Current Porter version: %v\n", currentPorterVersion)
	updateComponent = true
	if currentPorterVersion != "" {
		updateComponent = versionCheck("Porter", currentPorterVersion, porterVersionFromDependencies)
	}
	if updateComponent {
		fmt.Println("Downloading a newer version of Porter")
		// Download and install newer version of porter and mixins
		q.PorterExe, err = installPorter(q.QliksenseHome, q.PorterExe)
		if err != nil {
			logDebugMessage("Completed version checks")
			log.Fatalf("error installing porter: %v", err)
		}

		if _, err = installMixins(q.PorterExe, q.QliksenseHome); err != nil {
			logDebugMessage("Completed version checks")
			log.Fatalf("error installing mixin: %v", err)
		}
	}
}

func checkCLIVersion(dependencies map[string]string) {
	// Infer info about the minimum cli version
	var cliVersionFromDependencies, tmp string
	logDebugMessage("\n------------ CLI version check -----------\n")
	tmp, _ = dependencies["org.qlik.operator.cli.sense-installer.version.min"]
	if len(tmp) != 0 {
		cliVersionFromDependencies = tmp
	}
	logDebugMessage("\nCLI version from dependencies map: %v\n", cliVersionFromDependencies)

	// Checking version below
	updateComponent = versionCheck("CLI", pkg.Version, cliVersionFromDependencies)
	if updateComponent {
		log.Fatalf("Please download a newer version of CLI and retry the operation, exiting now.")
	}
}

func retrieveCurrentInstalledMixinVersions(q *qliksense.Qliksense) (map[string]string, error) {
	if _, err := os.Stat(filepath.Join(q.QliksenseHome, mixinDirVar)); err != nil {
		if os.IsNotExist(err) {
			// if path doesnt exist, return empty map, and let porter take care of the rest
			return map[string]string{}, nil
		}
		return nil, err
	}

	result := map[string]string{}
	currentInstalledMixinVersions, err := q.CallPorter([]string{"mixins", "list"}, func(x string) (out *string) {
		out = new(string)
		*out = strings.ReplaceAll(x, "porter", "qliksense porter")
		logDebugMessage("%s\n", *out)
		return
	})
	if err != nil {
		fmt.Printf("Error occurred when retrieving mixins list: %v", err)
		return nil, err
	}
	currentInstalledMixinVersionsArr := strings.Split(currentInstalledMixinVersions, "\n")
	for _, mix := range currentInstalledMixinVersionsArr {
		mixRow := strings.Fields(mix)
		mixRowLen := len(mixRow)
		if mixRowLen > 0 && mixRow[0] == "Name" {
			continue
		}
		// we handle the case of mixins like `kustomize` where version and author could be empty
		if mixRowLen >= 2 {
			_, err := semver.NewVersion(mixRow[1])
			if err == nil {
				result[mixRow[0]] = mixRow[1]
			}
		}
	}
	return result, nil
}

func determineVersion(versionString string) (string, error) {

	versionSlice := strings.Fields(versionString)

	var currentComponentVersionNumber *semver.Version
	var err error
	for _, value := range versionSlice {
		currentComponentVersionNumber, err = semver.NewVersion(value)
		if err == nil {
			break
		}
	}
	logDebugMessage("Version string: %v\n", currentComponentVersionNumber)
	if currentComponentVersionNumber != nil {
		return currentComponentVersionNumber.String(), nil
	}
	return "", fmt.Errorf("unable to extract version information")
}

func determineCurrentPorterVersion(q *qliksense.Qliksense) (string, error) {
	// determine current porter version
	currentPorterVersion, err := q.CallPorter([]string{"version"}, func(x string) (out *string) {
		out = new(string)
		*out = strings.ReplaceAll(x, "porter", "qliksense porter")
		logDebugMessage(*out)
		return
	})
	if err != nil {
		fmt.Printf("Error occurred during porter call: %v", err)
		return "", err
	}
	return determineVersion(currentPorterVersion)
}

func versionCheck(component string, currentVersion string, versionFromSourceOfTruth string) bool {

	if strings.HasPrefix(versionFromSourceOfTruth, "-v ") {
		versionFromSourceOfTruth = strings.Replace(versionFromSourceOfTruth, "-v ", "", 1)
	}
	componentVersionFromDependenciesYaml, err := semver.NewVersion(versionFromSourceOfTruth)
	if err != nil {
		fmt.Printf("There has been an error parsing version from source of truth: %s\n", err)
		return true
	}
	logDebugMessage("%s version from source of truth: %s\n", component, componentVersionFromDependenciesYaml)

	currentComponentVersion, err := semver.NewVersion(currentVersion)
	if err != nil {
		fmt.Printf("There has been an error parsing version from the derived current version: %s\n", err)
		return true
	}
	logDebugMessage("\nCurrently installed %s version: %v\n", component, currentComponentVersion)

	// check component version
	if currentComponentVersion.LessThan(componentVersionFromDependenciesYaml) {
		fmt.Printf("\n\nCurrent %s Component version: %s is less than minimum required version:%s\n", component, currentComponentVersion, componentVersionFromDependenciesYaml)
		return true
	}
	logDebugMessage("Current %s version is greater than version from dependencies, upgrade not necessary.\n\n", component)
	return false
}

func logDebugMessage(strMessage string, args ...interface{}) {
	if os.Getenv("QLIKSENSE_DEBUG") == "true" {
		fmt.Printf(strMessage, args...)
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
