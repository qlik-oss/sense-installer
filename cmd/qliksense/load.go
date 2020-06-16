package main

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func loadCrFile(q *qliksense.Qliksense) *cobra.Command {
	filePath := ""
	overwriteExistingContext := false
	c := &cobra.Command{
		Use:     "load",
		Short:   "load a CR a file and create necessary structure for future use",
		Long:    `load a CR a file and create necessary structure for future use`,
		Example: `qliksense load -f file_name or cat cr_file | qliksense load -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if crBytes, err := getCrBytesFromFileFlag(cmd); err != nil {
				return err
			} else {
				return q.LoadCr(crBytes, overwriteExistingContext)
			}
		},
	}
	f := c.Flags()
	f.StringVarP(&filePath, "file", "f", "", "File to load CR from")
	f.BoolVarP(&overwriteExistingContext, "overwrite", "o", overwriteExistingContext, "Overwrite any existing contexts with the same name")

	if err := c.MarkFlagRequired("file"); err != nil {
		panic(err)
	}
	return c
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

func getCrFileFromFlag(cmd *cobra.Command, flagName string) (*os.File, error) {
	filePath := cmd.Flag(flagName).Value.String()
	if filePath == "-" {
		if !isInputFromPipe() {
			return nil, errors.New("No input pipe present")
		} else {
			return os.Stdin, nil
		}
	} else if file, err := os.Open(filePath); err != nil {
		return nil, errors.Wrapf(err, "unable to read the file %s", filePath)
	} else {
		return file, nil
	}
}

func getCrBytesFromFileFlag(cmd *cobra.Command) ([]byte, error) {
	if file, err := getCrFileFromFlag(cmd, "file"); err != nil {
		return nil, err
	} else {
		defer file.Close()
		return ioutil.ReadAll(file)
	}
}
