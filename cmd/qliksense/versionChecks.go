
import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/qlik-oss/sense-installer/pkg"
	"gopkg.in/yaml.v2"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func stripFlagFromArgs(args []string, flagName string) []string {
	requirementsFileIndex := -1
	for i, element := range args {
		if element == flagName {
			fmt.Printf("Ash: %v found, index: %d\n", flagName, i)
			requirementsFileIndex = i
			break
		}
	}
	// if flag not found, return original
	if requirementsFileIndex == -1 {
		return args
	}

	argLength := len(args)
	var argsWithoutRequirementFile []string
	if requirementsFileIndex+2 < argLength {
		// removing the flagName and its value we added earlier
		argsWithoutRequirementFile = append(args[:requirementsFileIndex], args[requirementsFileIndex+2:]...)
	} else {
		argsWithoutRequirementFile = args[:requirementsFileIndex]
	}
	fmt.Printf("Ash: argsWithoutRequirementFile: %v\n", argsWithoutRequirementFile)
	return argsWithoutRequirementFile
}

func checkMinVersion(requirementsFile string) {
	// Check for existance of --requirementsFile flag and its value
	fmt.Printf("Ash: File: %v\n", requirementsFile)
	fmt.Printf("Ash: Command: %v\n", os.Args)
	var requirementsFileStatus bool
	// port over requirements into a map for ease of use
	dependencies := map[string]string{}

	for _, element := range os.Args {
		if element == requirementsFileFlag {
			requirementsFileStatus = true
			break
		}
	}
	fmt.Printf("\nAsh: requirements-file flag present? %t\n", requirementsFileStatus)
	if requirementsFileStatus {
		// flag --requirementsFile is present, now check for existence of requirements.yaml file
		if !fileExists(requirementsFile) {
			// exit if requirements.yaml is not found
			log.Fatalf("Ash: %s file doesn't exist, aborting operation.\n", requirementsFile)
			// os.Exit(1)
		} else {
			fmt.Printf("\nAsh: %s exists\n", requirementsFile)
			// read the requirements.yaml file and infer labels about the minimum cli, porter, and mixin versions
			yamlFile, err := ioutil.ReadFile(requirementsFile)
			if err != nil {
				log.Fatalf("Ash: Error reading YAML file: %s\n", err)
			}
			err = yaml.Unmarshal(yamlFile, &dependencies)
			if err != nil {
				log.Fatalf("Ash: Error parsing YAML file: %s\n", err)
			}
			fmt.Printf("Ash: read file: %+v\n", dependencies)

			os.Args = stripFlagFromArgs(os.Args, requirementsFileFlag)
		}
	}
	fmt.Printf("Ash: Args with tag: %v\n", os.Args)
	// os.Exit(1)

	// check CLI version
	currentCLIVersion, _ := semver.NewVersion(pkg.Version)
	fmt.Printf("\nAsh: Current CLI version: %v\n", currentCLIVersion)

	cliVersionFromRequirements, ok := dependencies["org.qlik.operator.cli.sense-installer.version.min"]
	if !ok {
		fmt.Errorf("There was an error when trying to retrieve the key: %s", ok)
	}
	fmt.Printf("Ash: CLI version from requirements.yaml: %v\n", cliVersionFromRequirements)

	// strip 'v' from the prefix of the version info
	cliVersionFromRequirementsStripped := stripVFromVersionInfo(cliVersionFromRequirements)
	fmt.Printf("Ash: After stripping 'v' from cliversion from requiremtns: %v\n", cliVersionFromRequirementsStripped)

	cliVersionFromRequirementsYaml, err := semver.NewVersion(cliVersionFromRequirementsStripped)
	fmt.Printf("Ash 1: %s", cliVersionFromRequirementsYaml)
	if err != nil {
		fmt.Printf("There has been an error! %s", err)
	}

	if currentCLIVersion.LessThan(cliVersionFromRequirementsYaml) {
		fmt.Printf("\n\nCurrent CLI version:%s is less than minimum required version:%s, please download minimum version or greater. Exiting for now.\n", currentCLIVersion, cliVersionFromRequirementsStripped)
		os.Exit(1)
	} else {
		fmt.Println("Current CLI version is greater than version from requirements")
	}

	// check porter version

	// check mixin version

	os.Exit(1)
}

func stripVFromVersionInfo(versionString string) string {
	if versionString[0] == 'v' {
		versionString = versionString[1:]
		fmt.Printf("Trimmed String: %s\n", versionString)
	}
	return versionString
}
