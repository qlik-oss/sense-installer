package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/qlik-oss/sense-installer/pkg"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const requirementsFileFlag = "--requirements-file"

func buildAliasCommands(porterCmd *cobra.Command, q *qliksense.Qliksense) []*cobra.Command {

	return []*cobra.Command{
		buildBuildAlias(porterCmd),
		buildInstallAlias(porterCmd, q),
		buildAboutAlias(porterCmd),
		buildPreflightAlias(porterCmd, q),
	}

}

func buildBuildAlias(porterCmd *cobra.Command) *cobra.Command {
	var (
		c *cobra.Command
	)
	c = &cobra.Command{
		Use:                "build",
		Short:              "Build a bundle",
		Long:               "Builds the bundle in the current directory by generating a Dockerfile and a CNAB bundle.json, and then building the invocation image.",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return porterCmd.RunE(porterCmd, append([]string{"build"}, args...))
		},
		Annotations: map[string]string{
			"group": "alias",
		},
	}
	return c
}

type paramOptions struct {
	aboutOptions
	Params           []string
	ParamFiles       []string
	Name             string
	InsecureRegistry bool

	// CredentialIdentifiers is a list of credential names or paths to make available to the bundle.
	CredentialIdentifiers []string
	Driver                string
	Force                 bool
	Insecure              bool

	// Requirements.yaml file to contain minimum versions of cli, porter and mixins.
	RequirementsFile string
}

func buildInstallAlias(porterCmd *cobra.Command, q *qliksense.Qliksense) *cobra.Command {
	var (
		c        *cobra.Command
		opts     *paramOptions
		registry *string
	)

	opts = &paramOptions{}

	c = &cobra.Command{
		Use:   "install [INSTANCE]",
		Short: "Install qliksense",
		Long: `Install a new instance of a bundle.

The first argument is the bundle instance name to create for the installation. This defaults to the name of the bundle. 

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.`,
		Example: `  qliksense install
  qliksense install --version v1.0.0
  qliksense install --insecure
  qliksense install qliksense --file qliksense/bundle.json
  qliksense install --param-file base-values.txt --param-file dev-values.txt --param test-mode=true --param header-color=blue
  qliksense install --cred kubernetes
  qliksense install --driver debug
  qliksense install MyAppFromTag --tag qlik/qliksense-cnab-bundle:v1.0.0
  qliksense install MyApp --requirements-file requirements.yaml
`,
		//DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Checking for min versions here
			// case 1: pre-existing requirements.yaml file
			fmt.Println("Ash: Hello Installing things for real!")
			checkMinVersion(opts.RequirementsFile)
			// Push images here.
			// TODO: Need to get the private reg from params
			args = append(os.Args[1:], opts.getTagDefaults(args)...)
			if registry = opts.findKey("dockerRegistry"); registry != nil {
				if len(*registry) > 0 {
					q.TagAndPushImages(*registry)
				}
			}
			return porterCmd.RunE(porterCmd, append([]string{"install"}, args...))
		},
		Annotations: map[string]string{
			"group": "alias",
		},
	}
	f := c.Flags()
	f.StringVarP(&opts.Version, "version", "v", "latest",
		"Version of Qlik Sense to install")
	f.BoolVar(&opts.Insecure, "insecure", true,
		"Allow working with untrusted bundles")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringSliceVar(&opts.ParamFiles, "param-file", nil,
		"Path to a parameters definition file for the bundle, each line in the form of NAME=VALUE. May be specified multiple times.")
	f.StringVar(&opts.RequirementsFile, "requirements-file", "requirements.yaml",
		"Path to a requirements file.")
	f.StringSliceVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters set with the same name using --param-file. May be specified multiple times.")
	f.StringSliceVarP(&opts.CredentialIdentifiers, "cred", "c", nil,
		"Credential to use when installing the bundle. May be either a named set of credentials or a filepath, and specified multiple times.")
	f.StringVarP(&opts.Driver, "driver", "d", "docker",
		"Specify a driver to use. Allowed values: docker, debug")
	f.StringVarP(&opts.Tag, "tag", "t", "",
		"Use a bundle in an OCI registry specified by the given tag")
	f.BoolVar(&opts.InsecureRegistry, "insecure-registry", false,
		"Don't require TLS for the registry")
	f.BoolVar(&opts.Force, "force", false,
		"Force a fresh pull of the bundle and all dependencies")
	return c
}

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

func (o *aboutOptions) getTagDefaults(args []string) []string {
	var err error
	if len(o.Tag) > 1 {
		args = append(args, []string{"--tag", o.Tag}...)
	}
	if len(o.Tag) <= 0 && len(o.File) <= 0 && len(o.CNABFile) <= 0 {
		if _, err = os.Stat("porter.yaml"); err != nil {
			args = append(args, []string{"--tag", "qlik/qliksense-cnab-bundle:" + o.Version}...)
		}
	}
	return args
}

type aboutOptions struct {
	Version  string
	Tag      string
	File     string
	CNABFile string
}

func buildAboutAlias(porterCmd *cobra.Command) *cobra.Command {
	var (
		c    *cobra.Command
		opts *aboutOptions
	)

	opts = &aboutOptions{}

	c = &cobra.Command{
		Use:   "about",
		Short: "About Qlik Sense",
		Long:  "Gives the verion of QLik Sense on Kuberntetes and versions of images.",
		RunE: func(cmd *cobra.Command, args []string) error {
			args = opts.getTagDefaults(args)
			return porterCmd.RunE(porterCmd, append([]string{"invoke", "--action", "about"}, args...))
		},
		Annotations: map[string]string{
			"group": "alias",
		},
	}
	f := c.Flags()
	f.StringVarP(&opts.Version, "version", "v", "latest",
		"Version of Qlik Sense to install")
	f.StringVarP(&opts.Tag, "tag", "t", "",
		"Use a bundle in an OCI registry specified by the given tag")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	return c
}

func buildPreflightAlias(porterCmd *cobra.Command, q *qliksense.Qliksense) *cobra.Command {
	var (
		c    *cobra.Command
		opts *paramOptions
	)

	opts = &paramOptions{}

	c = &cobra.Command{
		Use:   "preflight",
		Short: "Preflight Checks",
		Long:  "Perform Preflight Checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append(os.Args[1:], opts.getTagDefaults(args)...)
			return porterCmd.RunE(porterCmd, append([]string{"invoke", "--action", "preflight"}, args...))
		},
		Annotations: map[string]string{
			"group": "alias",
		},
	}
	f := c.Flags()
	f.StringSliceVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters set with the same name using --param-file. May be specified multiple times.")
	f.StringSliceVar(&opts.ParamFiles, "param-file", nil,
		"Path to a parameters definition file for the bundle, each line in the form of NAME=VALUE. May be specified multiple times.")
	f.StringVarP(&opts.Tag, "tag", "t", "",
		"Use a bundle in an OCI registry specified by the given tag")
	f.StringVarP(&opts.Version, "version", "v", "latest",
		"Version of Qlik Sense to install")
	f.StringSliceVarP(&opts.CredentialIdentifiers, "cred", "c", nil,
		"Credential to use when installing the bundle. May be either a named set of credentials or a filepath, and specified multiple times.")
	return c
}

func (o *paramOptions) findKey(param string) *string {
	var (
		value *string
	)
	if value = o.findParams(param); value != nil {
		return value
	}

	if value = o.findParamFiles(param); value != nil {
		return value
	}

	return nil
}

// parsedParams parses the variable assignments in Params.
func (o *paramOptions) findParams(param string) *string {
	return o.findVariableKey(param, o.Params)
}

// parseParamFiles parses the variable assignments in ParamFiles.
func (o *paramOptions) findParamFiles(param string) *string {
	var (
		path   string
		retStr *string
	)

	for _, path = range o.ParamFiles {
		retStr = o.findParamFile(param, path)
	}

	return retStr
}

func (o *paramOptions) findParamFile(param string, path string) *string {
	var (
		f       *os.File
		err     error
		scanner *bufio.Scanner
		lines   []string
		retStr  *string
	)
	if f, err = os.Open(path); err == nil {
		defer f.Close()

		scanner = bufio.NewScanner(f)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		retStr = o.findVariableKey(param, lines)
	}
	return retStr
}

func (o *paramOptions) findVariableKey(param string, params []string) *string {
	var (
		variable, value string
	)
	for _, p := range params {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) >= 2 {
			variable = strings.TrimSpace(parts[0])
			if variable == param {
				value = strings.TrimSpace(parts[1])
				return &value
			}
		}
	}
	return nil
}
