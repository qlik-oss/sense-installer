package qliksense

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (q *Qliksense) LoadCr(reader io.Reader) error {
	if crBytes, err := ioutil.ReadAll(reader); err != nil {
		return err
	} else if crName, err := q.loadCrStringIntoFileSystem(string(crBytes)); err != nil {
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

func (q *Qliksense) loadCrStringIntoFileSystem(crstr string) (string, error) {
	cr, err := qapi.CreateCRObjectFromString(crstr)
	if err != nil {
		return "", err
	}
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	if qConfig.IsContextExist(cr.GetName()) {
		return "", errors.New("Context Name: " + cr.GetName() + " already exist. please delete the existing context first using delete-context command")
	}
	qConfig.CreateContextDirs(cr.GetName())

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

	// write to disk

	if err = qConfig.CreateOrWriteCrAndContext(cr); err != nil {
		return "", err
	}
	qConfig.SetCurrentContextName(cr.GetName())
	qConfig.Write()

	return cr.GetName(), nil
}
