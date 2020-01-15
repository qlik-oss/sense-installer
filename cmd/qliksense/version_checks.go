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
	dependencies = map[string]string{}
)

func checkMinVersion(tag string, q *qliksense.Qliksense) {
	fmt.Println("Hello from checkMinVersion..")

	// check if tag is empty or not
	if len(strings.TrimSpace(tag)) == 0 {
		// Tag is empty, hence doing DIR way. Looking for dependenciesFile.yaml, exit if this file is not present
		if fileExists(dependenciesFile) {
			// read the dependencies.yaml and store into a map
			yamlFile, err := ioutil.ReadFile(dependenciesFile)
			if err != nil {
				log.Fatalf("Error reading YAML file: %s\n", err)
			}
			err = yaml.Unmarshal(yamlFile, &dependencies)
			if err != nil {
				log.Fatalf("Error parsing YAML file: %s\n", err)
			}
			// fmt.Printf("read file: %+v\n", dependencies)

			// Infer info about the minimum cli version
			var cliVersionFromDependencies, porterVersionFromDependencies, tmp string
			tmp = getVersionFromDependencyYaml("org.qlik.operator.cli.sense-installer.version.min")
			if len(tmp) != 0 {
				cliVersionFromDependencies = tmp
			}
			fmt.Printf("\nCLI version from dependencies.yaml: %v\n", cliVersionFromDependencies)

			// Checking version info below

			fmt.Printf("\n--------CLI version Check--------\n")
			updateComponent = versionCheck("CLI", pkg.Version, cliVersionFromDependencies)
			if updateComponent {
				fmt.Println("Please download a newer version of CLI and retry the operation, exiting now.")
				log.Fatalf("Error reading YAML file: %s\n", err)
			}

			// Infer info about the min porter version
			tmp = getVersionFromDependencyYaml("org.qlik.operator.cli.porter.version.min")
			if len(tmp) != 0 {
				porterVersionFromDependencies = tmp
			}
			fmt.Printf("Porter version from dependencies.yaml: %v\n", porterVersionFromDependencies)

			// check porter version
			fmt.Printf("\n--------Porter version Check--------\n")
			currentPorterVersion, err = determineCurrentPorterVersion(q)
			if err != nil {
				log.Println("warning:", err)
			}
			fmt.Printf("Current Porter version: %v\n", currentPorterVersion)
			updateComponent = true //
			if currentPorterVersion != "" {
				updateComponent = versionCheck("porter", currentPorterVersion, porterVersionFromDependencies)
			}
			if updateComponent {
				fmt.Println("Downloading a newer version of Porter and retrying the operation.")
				// TO-DO: download and install newer version of porter and retry the original command that was issued.
				q.PorterExe, err = installPorter(q.QliksenseHome)
				if err != nil {
					log.Fatal(err)
				}

				if _, err = installMixins(q.PorterExe, q.QliksenseHome); err != nil {
					log.Fatal(err)
				}
			}

			// // Infer info about the minimum mixin version
			// fmt.Println("\nMixinsVar BEFORE modification:")
			// for key, value := range mixinsVar {
			// 	fmt.Printf("%s: %s\n", key, value)
			// }

			currentMixinVersions, err := retrieveCurrentInstalledMixinVersions(q)
			if err != nil {
				log.Fatal(err)
			}
			for k := range mixinsVar {
				tmp = getVersionFromDependencyYaml(fmt.Sprintf("org.qlik.operator.mixin.%s.version.min", k))
				if tmp == "" {
					continue
				}
				shouldUpdateMixin := false
				mixinVersion, ok := currentMixinVersions[k]
				if !ok {
					shouldUpdateMixin = true
				} else {
					// if k == "qliksense" {
					// check mixin version
					fmt.Printf("\n--------%s Mixin version Check--------\n", k)

					// currentMixinVersion, err := determineVersion(currentMixinVersions[k])
					// if err != nil {
					// 	log.Fatal(err)
					// }
					shouldUpdateMixin = versionCheck(fmt.Sprintf("Mixin %s", k), mixinVersion, tmp)
				}
				// if tmp != "" and mixin requires Download and install
				if shouldUpdateMixin {
					fmt.Println("Downloading a newer version of mixin and retrying the operation.")
					// download and install the new mixin
					mURL, ok := mixinURLs[k]
					if ok {
						tmp = fmt.Sprintf("%s %s", tmp, mURL)
					}
					if _, err = installMixin(q.PorterExe, k, tmp); err != nil {
						// return err
						log.Fatalf("Error reading YAML file: %s\n", err)
					}
				}
			}

			// FOR MY DEVELOPMENT ONLY, DO NOT COMMIT INTO MASTER
			os.Exit(1)

		} else {
			fmt.Println("dependencies.yaml doesnt exist, hence exiting")
			os.Exit(1)
		}

	} else {
		// tag exists

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
		// have to handle the case of mixins like `kustomize` where version could be empty & the author too
		if mixRowLen >= 2 {
			_, err := semver.NewVersion(mixRow[1])
			if err == nil {
				result[mixRow[0]] = mixRow[1]
			}
		}
	}
	fmt.Printf("Ash: Output from porter mixins version: \n%v\n", result)
	return result, nil
}

func determineVersion(versionString string) (string, error) {
	fmt.Printf("Ash: Current version string: %v\n", versionString)

	versionSlice := strings.Fields(versionString)
	fmt.Printf("Ash: String slice: %v, length of slice: %d\n", versionSlice, len(versionSlice))

	var currentComponentVersionNumber *semver.Version
	var err error
	for _, value := range versionSlice {
		currentComponentVersionNumber, err = semver.NewVersion(value)
		if err == nil {
			break
		}
	}
	fmt.Printf("Ash: Version string at this point : %v\n", currentComponentVersionNumber)
	if currentComponentVersionNumber != nil {
		return currentComponentVersionNumber.String(), nil
	}
	return "", fmt.Errorf("unable to extract version information")
}

func determineCurrentPorterVersion(q *qliksense.Qliksense) (string, error) {
	// determine current porter version
	fmt.Println("Ash: Determining current Porter Version")
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
	fmt.Printf("Ash: Output from porter version: %v\n", currentPorterVersion)
	return determineVersion(currentPorterVersion)
}

func getVersionFromDependencyYaml(key string) string {
	if v, found := dependencies[key]; found {
		// fmt.Printf("Key: %s, Found value: %s", key, v)
		return v
	}
	return ""
}

func versionCheck(component string, currentVersion string, versionFromSourceOfTruth string) bool {
	fmt.Printf("%s version Check\n", component)
	fmt.Printf("current component version: %s\n", currentVersion)
	fmt.Printf("component version from source of truth: %s\n", versionFromSourceOfTruth)

	componentVersionFromDependenciesYaml, err := semver.NewVersion(versionFromSourceOfTruth)
	if err != nil {
		fmt.Printf("There has been an error! %s", err)
		return true
	}
	fmt.Printf("Ash: from source of truth: %s", componentVersionFromDependenciesYaml)
	// current Component version
	currentComponentVersion, err := semver.NewVersion(currentVersion)
	if err != nil {
		fmt.Printf("There has been an error! %s", err)
		return true
	}
	fmt.Printf("\nCurrent Component version: %v\n", currentComponentVersion)

	// check Component version
	if currentComponentVersion.LessThan(componentVersionFromDependenciesYaml) {
		fmt.Printf("\n\nCurrent %s Component version: %s is less than minimum required version:%s\n", component, currentComponentVersion, componentVersionFromDependenciesYaml)
		return true
	}
	fmt.Println("Current Component version is greater than version from dependencies, nothing to do.")
	return false
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
