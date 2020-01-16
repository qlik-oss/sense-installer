package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
	fmt.Println("Entry: CheckMinVersion()", "tag:", tag)
	dependencies := map[string]string{}
	// check if tag is empty or not
	var err error
	if len(strings.TrimSpace(tag)) > 0 {
		// --tag exists
		// fmt.Printf("Here is the tag: %s", tag)
		dependencies, err = q.PullImage(tag)
		fmt.Printf("\nDependencies map from the inspected image: %v\n", dependencies)
		if err != nil {
			log.Fatalf("unable to pull the requested image: %v", err)
		}

	} else {
		// Tag is empty, hence looking for dependenciesFile.yaml, exit if this file is not present
		if fileExists(dependenciesFile) {
			// read the dependencies.yaml and store into a map
			yamlFile, err := ioutil.ReadFile(dependenciesFile)
			if err != nil {
				fmt.Println("Exit: CheckMinVersion()")
				log.Fatalf("Error reading from source: %s\n", err)
			}
			err = yaml.Unmarshal(yamlFile, &dependencies)
			if err != nil {
				fmt.Println("Exit: CheckMinVersion()")
				log.Fatalf("Error when parsing from source: %s\n", err)
			}
		}
	}
	if len(dependencies) > 0 {
		// CLI check
		checkCLIVersion(dependencies)
		// Porter check
		checkPorterVersion(dependencies, q)
		// Mixins check
		checkMixinVersion(dependencies, q)
	} else {
		log.Fatalf("Not able to infer dependencies, hence exiting")
	}
	fmt.Println("Exit: CheckMinVersion()")
	// os.Exit(1)
}

func checkMixinVersion(dependencies map[string]string, q *qliksense.Qliksense) {
	var tmp string
	fmt.Println("------------ Mixins version check -----------")
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
			fmt.Println("Downloading a newer version of mixin")
			// download and install the new mixin
			mURL, ok := mixinURLs[k]
			if ok {
				tmp = fmt.Sprintf("%s %s", tmp, mURL)
			}
			if _, err = installMixin(q.PorterExe, k, tmp); err != nil {
				// return err
				fmt.Println("Exit: CheckMinVersion()")
				log.Fatalf("Error reading YAML file: %s\n", err)
			}
		}
	}
}

func checkPorterVersion(dependencies map[string]string, q *qliksense.Qliksense) {
	// Infer info about the min porter version
	var porterVersionFromDependencies, tmp string
	var err error
	fmt.Println("------------ Porter version check -----------")
	tmp, _ = dependencies["org.qlik.operator.cli.porter.version.min"]
	if len(tmp) != 0 {
		porterVersionFromDependencies = tmp
	}
	fmt.Printf("Porter version from dependencies map: %v\n", porterVersionFromDependencies)

	// check porter version
	currentPorterVersion, err = determineCurrentPorterVersion(q)
	if err != nil {
		log.Println("warning:", err)
	}
	fmt.Printf("Current Porter version: %v\n", currentPorterVersion)
	updateComponent = true //
	if currentPorterVersion != "" {
		updateComponent = versionCheck("Porter", currentPorterVersion, porterVersionFromDependencies)
	}
	if updateComponent {
		fmt.Println("Downloading a newer version of Porter")
		// Download and install newer version of porter and mixins
		q.PorterExe, err = installPorter(q.QliksenseHome)
		if err != nil {
			fmt.Println("Exit: CheckMinVersion()")
			log.Fatal(err)
		}

		if _, err = installMixins(q.PorterExe, q.QliksenseHome); err != nil {
			fmt.Println("Exit: CheckMinVersion()")
			log.Fatal(err)
		}
	}
}

func checkCLIVersion(dependencies map[string]string) {
	// Infer info about the minimum cli version
	var cliVersionFromDependencies, tmp string
	var err error
	fmt.Printf("\n------------ CLI version check -----------\n")
	tmp, _ = dependencies["org.qlik.operator.cli.sense-installer.version.min"]
	if len(tmp) != 0 {
		cliVersionFromDependencies = tmp
	}
	fmt.Printf("\nCLI version from dependencies map: %v\n", cliVersionFromDependencies)

	// Checking version below
	updateComponent = versionCheck("CLI", pkg.Version, cliVersionFromDependencies)
	if updateComponent {
		fmt.Println("Please download a newer version of CLI and retry the operation, exiting now.")
		fmt.Println("Exit: CheckMinVersion()")
		log.Fatalf("Error reading YAML file: %s\n", err)
	}
}

func retrieveCurrentInstalledMixinVersions(q *qliksense.Qliksense) (map[string]string, error) {
	result := map[string]string{}
	currentInstalledMixinVersions, err := q.CallPorter([]string{"mixins", "list"}, func(x string) (out *string) {
		out = new(string)
		*out = strings.ReplaceAll(x, "porter", "qliksense porter")
		fmt.Println(*out)
		return
	})
	if err != nil {
		log.Printf("ERROR occurred during porter call: %v", err)
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
	fmt.Printf("Output from porter mixins version: \n%v\n", result)
	return result, nil
}

func determineVersion(versionString string) (string, error) {
	// fmt.Printf("Current version string: %v\n", versionString)

	versionSlice := strings.Fields(versionString)
	// fmt.Printf("String slice: %v, length of slice: %d\n", versionSlice, len(versionSlice))

	var currentComponentVersionNumber *semver.Version
	var err error
	for _, value := range versionSlice {
		currentComponentVersionNumber, err = semver.NewVersion(value)
		if err == nil {
			break
		}
	}
	fmt.Printf("Version string: %v\n", currentComponentVersionNumber)
	if currentComponentVersionNumber != nil {
		return currentComponentVersionNumber.String(), nil
	}
	return "", fmt.Errorf("unable to extract version information")
}

func determineCurrentPorterVersion(q *qliksense.Qliksense) (string, error) {
	// determine current porter version
	fmt.Printf("Determining current Porter Version\n")
	currentPorterVersion, err := q.CallPorter([]string{"version"}, func(x string) (out *string) {
		out = new(string)
		*out = strings.ReplaceAll(x, "porter", "qliksense porter")
		fmt.Println(*out)
		return
	})
	if err != nil {
		log.Printf("ERROR occurred during porter call: %v", err)
		return "", err
	}
	fmt.Printf("Output from porter version: %v", currentPorterVersion)
	return determineVersion(currentPorterVersion)
}

func versionCheck(component string, currentVersion string, versionFromSourceOfTruth string) bool {
	// fmt.Printf("----------%s Version check----------\n", component)
	// fmt.Printf("Current component version: %s\n", currentVersion)
	// fmt.Printf("Component version from source of truth: %s\n", versionFromSourceOfTruth)

	componentVersionFromDependenciesYaml, err := semver.NewVersion(versionFromSourceOfTruth)
	if err != nil {
		fmt.Printf("There has been an error! %s", err)
		return true
	}
	fmt.Printf("%s version from source of truth: %s", component, componentVersionFromDependenciesYaml)

	currentComponentVersion, err := semver.NewVersion(currentVersion)
	if err != nil {
		fmt.Printf("There has been an error! %s", err)
		return true
	}
	fmt.Printf("\nCurrently installed %s version: %v\n", component, currentComponentVersion)

	// check component version
	if currentComponentVersion.LessThan(componentVersionFromDependenciesYaml) {
		fmt.Printf("\n\nCurrent %s Component version: %s is less than minimum required version:%s\n", component, currentComponentVersion, componentVersionFromDependenciesYaml)
		return true
	}
	fmt.Printf("Current %s version is greater than version from dependencies, nothing to do.\n\n", component)

	return false
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
