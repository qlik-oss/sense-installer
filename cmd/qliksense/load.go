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
	c := &cobra.Command{
		Use:     "load",
		Short:   "load a CR a file and create necessary structure for future use",
		Long:    `load a CR a file and create necessary structure for future use`,
		Example: `qliksense load -f file_name`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLoadOrApplyCommandE(cmd, func(reader io.Reader) error {
				return q.LoadCr(reader)
			})
		},
	}
	f := c.Flags()
	f.StringVarP(&filePath, "file", "f", "", "File to load CR from")
	c.MarkFlagRequired("file")

	eulaPreRunHooks.addValidator(c.Name(), loadOrApplyCommandEulaPreRunHook)
	return c
}

func getCrFileFromFlag(cmd *cobra.Command, flagName string) (*os.File, error) {
	filePath := cmd.Flag(flagName).Value.String()
	file, e := os.Open(filePath)
	if e != nil {
		return nil, errors.Wrapf(e, "unable to read the file %s", filePath)
	}
	return file, nil
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
		eulaPreRunHooks.addPostValidationArtifact(cmd.Name(), "CR", crBytes)
		return q.IsEulaAcceptedInCrFile(bytes.NewBuffer(crBytes))
	}
}

func runLoadOrApplyCommandE(cmd *cobra.Command, callBack func(io.Reader) error) error {
	if crBytes := eulaPreRunHooks.getPostValidationArtifact(cmd.Name(), "CR"); crBytes != nil {
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
