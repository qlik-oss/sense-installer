package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type aboutCommandOptions struct {
	Profile string
}

func about(q *qliksense.Qliksense) *cobra.Command {
	opts := &aboutCommandOptions{}

	c := &cobra.Command{
		Use:   "about ref",
		Short: "Displays information pertaining to qliksense on Kubernetes",
		Long:  "Gives the version of QLik Sense on Kubernetes and versions of images.",
		Example: `
qliksense about 1.0.0
  - display default profile (docker-desktop) for Git ref 1.0.0 in the qliksense-k8s repo 
qliksense about 1.0.0 --profile=docker-desktop
  - specifying profile
qliksense about
qliksense about --profile=test
  - if no Git ref is provided, then get version information from the configuration on disk:
    - if user's current directory has a subdirectory "manifests/${profile}",
      then get version information from that
    - if using other supported commands the user has built a CR in ~/.qliksense, 
      then get version information based on the path derived like so: 
      - ${spec.manifestsRoot}/${spec.profile} # if no profile flag provided
      - ${spec.manifestsRoot}/${profile} # if profile is provided using the --profile command flag
  - if no config found on disk in locations described above, 
    then get version information based on the default profile in the qliksense-k8s repo master
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if gitRef, err := getSingleArg(args); err != nil {
				return err
			} else if vout, err := q.About(gitRef, opts.Profile); err != nil {
				return err
			} else if out, err := yaml.Marshal(vout); err != nil {
				return err
			} else if _, err := fmt.Println(string(out)); err != nil {
				return err
			}
			return nil
		},
	}
	f := c.Flags()
	f.StringVar(&opts.Profile, "profile", "", "Configuration profile")
	return c
}

func getSingleArg(args []string) (string, error) {
	if len(args) > 1 {
		return "", errors.New("too many arguments, only 1 expected")
	} else if len(args) == 1 {
		return strings.TrimSpace(args[0]), nil
	}
	return "", nil
}
