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
	dependenciesFile     = "dependencies.yaml"
	porterPermaLink      = pkg.Version
	currentPorterVersion string
	mixinsVar            = map[string]string{
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
	dependencies = map[string]string{}
)

func checkMinVersion(tag string, q *qliksense.Qliksense) {
	fmt.Println("Ash: Hello from checkMinVersion..")

	// check if tag is empty or not
	if len(strings.TrimSpace(tag)) == 0 {
		// Tag is empty, hence doing DIR way. Looking for dependenciesFile.yaml, exit if this file is not present
		if fileExists(dependenciesFile) {
			// read the dependencies.yaml and store into a map
			yamlFile, err := ioutil.ReadFile(dependenciesFile)
			if err != nil {
				log.Fatalf("Ash: Error reading YAML file: %s\n", err)
			}
			err = yaml.Unmarshal(yamlFile, &dependencies)
			if err != nil {
				log.Fatalf("Ash: Error parsing YAML file: %s\n", err)
			}
			// fmt.Printf("Ash: read file: %+v\n", dependencies)

			// Infer info about the minimum cli version
			var cliVersionFromDependencies, porterVersionFromDependencies, tmp string
			tmp = getVersionFromDependencyYaml("org.qlik.operator.cli.sense-installer.version.min")
			if len(tmp) != 0 {
				cliVersionFromDependencies = tmp
			}
			fmt.Printf("\nAsh: CLI version from dependencies.yaml: %v\n", cliVersionFromDependencies)

			// Infer info about the min porter version
			tmp = getVersionFromDependencyYaml("org.qlik.operator.cli.porter.version.min")
			if len(tmp) != 0 {
				// if tmp > currentPorterVersion {
				// 	currentPorterVersion = tmp // TO-DO: We need to download and install newer porter version now.
				// }
				porterVersionFromDependencies = tmp
			}
			fmt.Printf("Ash: Porter version from dependencies.yaml: %v\n", porterPermaLink)

			// Infer info about the minimum mixin version
			fmt.Println("\nAsh: MixinsVar BEFORE modification:")
			for key, value := range mixinsVar {
				fmt.Printf("%s: %s\n", key, value)
			}

			for k, _ := range mixinsVar {
				if k == "qliksense" {
					tmp = getVersionFromDependencyYaml("org.qlik.operator.mixin.qliksense.version.min")
					if tmp != "" {
						mixinsVar[k] = tmp
					}
				}
				if k == "kustomize" {
					tmp = getVersionFromDependencyYaml("org.qlik.operator.mixin.kustomize.version.min")
					if tmp != "" {
						mixinsVar[k] = tmp
					}
				}
				if k == "exec" {
					tmp = getVersionFromDependencyYaml("org.qlik.operator.mixin.exec.version.min")
					if tmp != "" {
						mixinsVar[k] = tmp
					}
				}
				if k == "kubernetes" {
					tmp = getVersionFromDependencyYaml("org.qlik.operator.mixin.kubernetes.version.min")
					if tmp != "" {
						mixinsVar[k] = tmp
					}
				}
			}
			fmt.Println("\nAsh: MixinsVar AFTER modification:")
			for key, value := range mixinsVar {
				fmt.Printf("%s: %s\n", key, value)
			}

			// Checking version info below.

			fmt.Printf("\n--------CLI version Check--------\n")
			// First, strip 'v' from the prefix of the version info
			// cliVersionFromDependenciesStripped := stripVFromVersionInfo(cliVersionFromDependencies)
			// fmt.Printf("Ash: After stripping 'v' from cliversion from requiremtns: %v\n", cliVersionFromDependenciesStripped)
			versionCheck("CLI", pkg.Version, cliVersionFromDependencies)

			// check porter version
			fmt.Printf("\n--------Porter version Check--------\n")
			currentPorterVersion = determineCurrentPorterVersion(q)
			fmt.Printf("Current Porter version: %v\n", currentPorterVersion)

			versionCheck("porter", currentPorterVersion, porterVersionFromDependencies)

			// check mixin version
			fmt.Printf("\n--------Mixins version Check--------\n")

			os.Exit(1)

		} else {
			fmt.Println("dependencies.yaml doesnt exist, hence exiting")
			os.Exit(1)
		}

	} else {
		// tag exists

	}

	// installPorter(porterPermaLink, mixinsVar)
}

func determineCurrentPorterVersion(q *qliksense.Qliksense) string {
	// determine current porter version
	fmt.Println("Determining current Porter Version")
	currentPorterVersion, err := q.CallPorter([]string{"version"}, func(x string) (out *string) {
		out = new(string)
		*out = strings.ReplaceAll(x, "porter", "qliksense porter")
		// fmt.Println(*out)
		return
	})
	if err != nil {
		log.Fatalf("ERROR occurred during porter call: %v", err)
	}
	fmt.Printf("Output from porter version: %v\n", currentPorterVersion)

	porterVersionSlice := strings.Fields(currentPorterVersion)
	// fmt.Printf("Ash: String slice: %v, length of slice: %d\n", porterVersionSlice, len(porterVersionSlice))

	var currentPorterVersionNumber, tmpVer *semver.Version
	for _, value := range porterVersionSlice {
		tmpVer, err = semver.NewVersion(value)
		if err != nil {
			fmt.Errorf("Error parsing version: %s", err)
		} else {
			currentPorterVersionNumber = tmpVer
			fmt.Printf("Ash: HERE IS THE RIGHT PORTER VERSION: %s\n\n", currentPorterVersionNumber)
		}
	}
	return currentPorterVersionNumber.String()
}

func getVersionFromDependencyYaml(key string) string {
	if v, found := dependencies[key]; found {
		// fmt.Printf("Ash: Key: %s, Found value: %s", key, v)
		return v
	}
	return ""
}

func versionCheck(component string, currentVersion string, versionFromSourceOfTruth string) {

	// Commented code start
	// cliVersionFromDependenciesYaml, err := semver.NewVersion(cliVersionFromDependenciesStripped)
	// fmt.Printf("Ash 1: %s", cliVersionFromDependenciesYaml)
	// if err != nil {
	// 	fmt.Printf("There has been an error! %s", err)
	// }

	// // current CLI version
	// currentCLIVersion, _ := semver.NewVersion(pkg.Version)
	// fmt.Printf("\nAsh: Current CLI version: %v\n", currentCLIVersion)

	// check CLI version
	// if currentCLIVersion.LessThan(cliVersionFromDependenciesYaml) {
	// 	fmt.Printf("\n\nCurrent CLI version:%s is less than minimum required version:%s, please download minimum version or greater. Exiting for now.\n", currentCLIVersion, cliVersionFromDependenciesStripped)
	// 	os.Exit(1)
	// } else {
	// 	fmt.Println("Current CLI version is greater than version from dependencies, nothing to do.")
	// }

	// Comented code end

	fmt.Printf("%s version Check\n", component)
	fmt.Printf("current component version: %s\n", currentVersion)
	fmt.Printf("component version from source of truth: %s\n", versionFromSourceOfTruth)

	// First, strip 'v' from the prefix of the version info
	// componentVersionFromDependenciesStripped := stripVFromVersionInfo(versionFromSourceOfTruth)
	// fmt.Printf("Ash: After stripping 'v' from component version from requirements: %v\n", componentVersionFromDependenciesStripped)

	// componentVersionFromDependenciesYaml, err := semver.NewVersion(componentVersionFromDependenciesStripped)
	componentVersionFromDependenciesYaml, err := semver.NewVersion(versionFromSourceOfTruth)

	fmt.Printf("Ash from source of truth: %s", componentVersionFromDependenciesYaml)
	if err != nil {
		fmt.Printf("There has been an error! %s", err)
	}

	// current Component version
	currentComponentVersion, _ := semver.NewVersion(currentVersion)
	fmt.Printf("\nAsh: Current Component version: %v\n", currentComponentVersion)

	// check Component version
	if currentComponentVersion.LessThan(componentVersionFromDependenciesYaml) {
		fmt.Printf("\n\nCurrent Component version:%s is less than minimum required version:%s\n", currentComponentVersion, componentVersionFromDependenciesYaml)
		if component == "porter" {
			fmt.Println("TO-DO: Download and install newer version of Porter")
		} else if component == "CLI" {
			fmt.Println("Please download and install newer CLI component. Exiting now.")
			os.Exit(1)
		}

	} else {
		fmt.Println("Current Component version is greater than version from dependencies, nothing to do.")
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func stripVFromVersionInfo(versionString string) string {
	if versionString[0] == 'v' {
		versionString = versionString[1:]
		fmt.Printf("Trimmed String: %s\n", versionString)
	}
	return versionString
}
