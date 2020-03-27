package main

import (
	"bytes"
	"io"
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
			return runLoadOrApplyCommandE(cmd, func(reader io.Reader) error {
				return q.LoadCr(reader, overwriteExistingContext)
			})
		},
	}
	f := c.Flags()
	f.StringVarP(&filePath, "file", "f", "", "File to load CR from")
	c.MarkFlagRequired("file")
	f.BoolVarP(&overwriteExistingContext, "overwrite", "o", overwriteExistingContext, "Overwrite any existing contexts with the same name")

	eulaPreRunHooks.addValidator(c.Name(), loadOrApplyCommandEulaPreRunHook)
	return c
}

func getCrFileFromFlag(cmd *cobra.Command, flagName string) (*os.File, error) {
	filePath := cmd.Flag(flagName).Value.String()
	if filePath == "-" {
		if !isInputFromPipe() {
			return nil, errors.New("No input pipe present")
		}
		return os.Stdin, nil
	}
	file, e := os.Open(filePath)
	if e != nil {
		return nil, errors.Wrapf(e,
			"unable to read the file %s", filePath)
	}
	return file, nil
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

func loadOrApplyCommandEulaPreRunHook(cmd *cobra.Command, q *qliksense.Qliksense) (bool, error) {
	file, err := getCrFileFromFlag(cmd, "file")
	if err != nil {
		return false, err
	}
	defer file.Close()

	if crBytes, err := ioutil.ReadAll(file); err != nil {
		return false, err
	} else {
		eulaPreRunHooks.addPostValidationArtifact("CR", crBytes)
		return q.IsEulaAcceptedInCrFile(bytes.NewBuffer(crBytes))
	}
}

func runLoadOrApplyCommandE(cmd *cobra.Command, callBack func(io.Reader) error) error {
	if crBytes := eulaPreRunHooks.getPostValidationArtifact("CR"); crBytes != nil {
		return callBack(bytes.NewBuffer(crBytes.([]byte)))
	} else {
		file, err := getCrFileFromFlag(cmd, "file")
		if err != nil {
			return err
		}
		defer file.Close()
		return callBack(file)
	}
}
