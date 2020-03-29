package qliksense

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (q *Qliksense) LoadCr(reader io.Reader, overwriteExistingContext bool) error {
	if crBytes, err := ioutil.ReadAll(reader); err != nil {
		return err
	} else if crName, err := q.loadCrStringIntoFileSystem(string(crBytes), overwriteExistingContext); err != nil {
		return err
	} else {
		fmt.Println("cr name: [ " + crName + " ] has been loaded")
	}
	return nil
}

func (q *Qliksense) IsEulaAcceptedInCrFile(reader io.Reader) (bool, error) {
	if crBytes, err := ioutil.ReadAll(reader); err != nil {
		return false, err
	} else if cr, err := qapi.CreateCRObjectFromString(string(crBytes)); err != nil {
		return false, err
	} else {
		return cr.IsEULA(), nil
	}
}

func (q *Qliksense) loadCrStringIntoFileSystem(crstr string, overwriteExistingContext bool) (string, error) {
	cr, err := qapi.CreateCRObjectFromString(crstr)
	if err != nil {
		return "", err
	}
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	if qConfig.IsContextExist(cr.GetName()) {
		if !overwriteExistingContext {
			return "", errors.New("Context with name: " + cr.GetName() + " already exists. " +
				"Please delete the existing context first using the delete-context command or specify the --overwrite flag.")
		}
		// else if err := os.RemoveAll(qConfig.GetContextPath(cr.GetName())); err != nil {
		// 	return "", err
		// }
	}
	if err := qConfig.CreateContextDirs(cr.GetName()); err != nil {
		return "", err
	}

	// encrypt the secrets and do base64 then update the CR
	rsaPublicKey, _, err := qConfig.GetContextEncryptionKeyPair(cr.GetName())
	if err != nil {
		return "", err
	}
	for svc, nvs := range cr.Spec.Secrets {
		for _, nv := range nvs {
			if nv.ValueFrom == nil {
				skv := &qapi.ServiceKeyValue{
					Key:     nv.Name,
					Value:   nv.Value,
					SvcName: svc,
				}
				if err := q.processSecret(skv, rsaPublicKey, cr, false); err != nil {
					return cr.GetName(), err
				}
			}
		}
	}

	// update manifestsRoot in case already exist
	if existingCr, err := qConfig.GetCR(cr.GetName()); err == nil {
		// cr exists, so update the manifestsRoot if version exist
		newV := cr.GetLabelFromCr("version")
		if strings.HasSuffix(existingCr.Spec.ManifestsRoot, newV) {
			cr.Spec.ManifestsRoot = existingCr.Spec.ManifestsRoot
		}
	}
	// write to disk
	if err = qConfig.CreateOrWriteCrAndContext(cr); err != nil {
		return "", err
	}
	qConfig.SetCurrentContextName(cr.GetName())
	qConfig.Write()

	return cr.GetName(), nil
}
