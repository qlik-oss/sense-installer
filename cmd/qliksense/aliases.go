package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func buildAliasCommands(porterCmd *cobra.Command, q *qliksense.Qliksense) []*cobra.Command {

	return []*cobra.Command{
		buildBuildAlias(porterCmd),
		buildInstallAlias(porterCmd, q),
		buildAboutAlias(porterCmd),
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
`,
		//DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Push images here.
			// TODO: Need to get the private reg from params
			args = opts.getTagDefaults(args)
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
func (o *aboutOptions) getTagDefaults(args []string) []string {
	var err error
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
