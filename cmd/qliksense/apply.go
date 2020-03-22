package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func applyCmd(q *qliksense.Qliksense) *cobra.Command {

	filePath := ""
	c := &cobra.Command{
		Use:     "apply",
		Short:   "install qliksense based on provided cr file",
		Long:    `install qliksense based on provided cr file`,
		Example: `qliksense apply -f file_name or cat cr_file | qliksense apply -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "-" {
				if !isInputFromPipe() {
					return errors.New("No input pipe present")
				}
				return q.ApplyCRFromReader(os.Stdin)
			}
			file, e := os.Open(filePath)
			if e != nil {
				return errors.Wrapf(e,
					"unable to read the file %s", filePath)
			}
			return q.ApplyCRFromReader(file)
		},
	}

	f := c.Flags()
	f.StringVarP(&filePath, "file", "f", "", "Install from a CR file")
	c.MarkFlagRequired("file")
	return c
}
