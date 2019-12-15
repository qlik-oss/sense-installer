package main

import (
	"github.com/spf13/cobra"
)

func buildAliasCommands(porterCmd *cobra.Command) []*cobra.Command {

	return []*cobra.Command{
		buildBuildAlias(porterCmd),
		buildInstallAlias(porterCmd),
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

func buildInstallAlias(porterCmd *cobra.Command) *cobra.Command {
	var (
		c *cobra.Command
	)
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
			return porterCmd.RunE(porterCmd, append([]string{"install"}, args...))
		},
		Annotations: map[string]string{
			"group": "alias",
		},
	}
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