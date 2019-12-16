package main

import (
	"os"
	"github.com/spf13/cobra"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"strings"
	"bufio"
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
		Use:   "build",
		Short: "Build a bundle",
		Long:  "Builds the bundle in the current directory by generating a Dockerfile and a CNAB bundle.json, and then building the invocation image.",
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
	Params []string
	ParamFiles []string
	parsedParams map[string]string
	parsedParamFiles []map[string]string
	combinedParameters map[string]string
}
func buildInstallAlias(porterCmd *cobra.Command, q *qliksense.Qliksense) *cobra.Command {
	var (
		c *cobra.Command
		opts *paramOptions
		registry *string
	)

	opts = &paramOptions{

	}

	c = &cobra.Command{
		Use:   "install [INSTANCE]",
		Short: "Install qliksense",
		Long: `Install a new instance of a bundle.

The first argument is the bundle instance name to create for the installation. This defaults to the name of the bundle. 

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.`,
		Example: `  qliksense install
  qliksense install --insecure
  qliksense install qliksense --file qliksense/bundle.json
  qliksense install --param-file base-values.txt --param-file dev-values.txt --param test-mode=true --param header-color=blue
  qliksense install --cred kubernetes
  qliksense install --driver debug
  qliksense install MyAppFromTag --tag qlik/qliksense-cnab-bundle:v1.0.0
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Push images here.
			// TODO: Need to get the private reg from params
			if registry = opts.findKey("dockerRegistry"); registry != nil{
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
	f.StringSliceVar(&opts.ParamFiles, "param-file", nil,
		"Path to a parameters definition file for the bundle, each line in the form of NAME=VALUE. May be specified multiple times.")
	f.StringSliceVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters set with the same name using --param-file. May be specified multiple times.")
	return c
}

func buildAboutAlias(porterCmd *cobra.Command) *cobra.Command {
	var (
		c *cobra.Command
	)
	c = &cobra.Command{
		Use:   "about",
		Short: "About Qlik Sense",
		Long:  "Gives the verion of QLik Sense on Kuberntetes and versions of images.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return porterCmd.RunE(porterCmd, append([]string{"invoke","--action","about"}, args...))
		},
		Annotations: map[string]string{
			"group": "alias",
		},
	}
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
		path string
		retStr *string
	)

	for _, path = range o.ParamFiles {
		retStr = o.findParamFile(param, path)
	}

	return retStr
}

func (o *paramOptions) findParamFile(param string, path string) *string {
	var (
		f *os.File
		err error
		scanner *bufio.Scanner
		lines []string
		retStr *string
	)
	if f, err = os.Open(path); err == nil {
		defer f.Close()

		scanner = bufio.NewScanner(f)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		retStr = o.findVariableKey(param,lines)
	}
	return retStr
}

func (o *paramOptions) findVariableKey(param string, params []string) (*string) {
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