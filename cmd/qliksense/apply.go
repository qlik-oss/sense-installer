package main

import (
	"bytes"
	"errors"
	"fmt"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"

	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

func applyCmd(q *qliksense.Qliksense) *cobra.Command {
	opts := &qliksense.InstallCommandOptions{}
	filePath := ""
	keepPatchFiles, pull, push := false, false, false
	c := &cobra.Command{
		Use:     "apply",
		Short:   "install qliksense based on provided cr file",
		Long:    `install qliksense based on provided cr file`,
		Example: `qliksense apply -f file_name or cat cr_file | qliksense apply -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLoadOrApplyCommandE(cmd, func(crBytes []byte) error {
				if cr, crBytesWithEula, err := getCrWithEulaInserted(crBytes); err != nil {
					return err
				} else if err := validatePullPushFlagsOnApply(cr, pull, push); err != nil {
					return err
				} else {
					return q.ApplyCRFromReader(bytes.NewReader(crBytesWithEula), opts, keepPatchFiles, true, pull, push)
				}
			})
		},
	}

	f := c.Flags()
	f.StringVarP(&filePath, "file", "f", "", "Install from a CR file")
	c.MarkFlagRequired("file")
	f.StringVarP(&opts.StorageClass, "storageClass", "s", "", "Storage class for qliksense")
	f.StringVarP(&opts.MongodbUri, "mongodbUri", "m", "", "mongodbUri for qliksense (i.e. mongodb://qlik-default-mongodb:27017/qliksense?ssl=false)")
	f.StringVarP(&opts.RotateKeys, "rotateKeys", "r", "", "Rotate JWT keys for qliksense (yes:rotate keys/ no:use exising keys from cluster/ None: use default EJSON_KEY from env")
	f.BoolVar(&keepPatchFiles, keepPatchFilesFlagName, keepPatchFiles, keepPatchFilesFlagUsage)
	f.BoolVarP(&pull, pullFlagName, pullFlagShorthand, pull, pullFlagUsage)
	f.BoolVarP(&push, pushFlagName, pushFlagShorthand, push, pushFlagUsage)

	eulaPreRunHooks.addValidator(fmt.Sprintf("%v %v", rootCommandName, c.Name()), loadOrApplyCommandEulaPreRunHook)

	return c
}

func validatePullPushFlagsOnApply(cr *qapi.QliksenseCR, pull, push bool) error {
	if pull && !push {
		fmt.Printf("WARNING: pulling images without pushing them")
	}
	if push {
		if registry := cr.Spec.GetImageRegistry(); registry == "" {
			return errors.New("no image registry set in the CR; to set it use: qliksense config set-image-registry")
		}
	}
	return nil
}

func getCrWithEulaInserted(crBytes []byte) (*qapi.QliksenseCR, []byte, error) {
	if cr, err := qapi.CreateCRObjectFromString(string(crBytes)); err != nil {
		return nil, nil, err
	} else {
		cr.SetEULA("yes")
		if crBytesWithEula, err := qapi.K8sToYaml(cr); err != nil {
			return nil, nil, err
		} else {
			return cr, crBytesWithEula, nil
		}
	}
}
