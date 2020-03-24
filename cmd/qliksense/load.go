package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func loadCrFile(q *qliksense.Qliksense) *cobra.Command {
	filePath := ""
	c := &cobra.Command{
		Use:   "load",
		Short: "load a CR a file and create necessary structure for future use",
		Long:  `load a CR a file and create necessary structure for future use`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "-" {
				if !isInputFromPipe() {
					return errors.New("No input pipe present")
				}
				return q.LoadCr(os.Stdin)
			}
			file, e := os.Open(filePath)
			if e != nil {
				return errors.Wrapf(e,
					"unable to read the file %s", filePath)
			}
			return q.LoadCr(file)
		},
	}
	f := c.Flags()
	f.StringVarP(&filePath, "file", "f", "", "File to laod CR from")
	c.MarkFlagRequired("file")
	return c
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}
