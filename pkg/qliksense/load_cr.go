package qliksense

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

//
func (q *Qliksense) LoadCr(reader io.Reader) error {
	for _, doc := range readMultipleYamlFromReader(reader) {
		if crName, err := q.loadCrStringIntoFileSystem(doc); err != nil {
			return err
		} else {
			fmt.Println("cr name: [ " + crName + " ] has been loaded")
		}
	}
	return nil
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
	if err = qapi.WriteToFile(cr, qConfig.BuildCrFilePath(cr.GetName())); err != nil {
		return "", err
	}
	qConfig.AddToContextsRaw(cr.GetName(), qConfig.BuildCrFilePath(cr.GetName()))
	qConfig.SetCurrentContextName(cr.GetName())
	qConfig.Write()

	return cr.GetName(), nil
}

func readMultipleYamlFromReader(reader io.Reader) []string {
	docs := make([]string, 0)
	scanner := bufio.NewScanner(bufio.NewReader(reader))
	adoc := ""
	for scanner.Scan() {
		s := scanner.Text()
		if s == "---" {
			docs = append(docs, adoc)
			adoc = ""
			s = ""
		}
		adoc = adoc + "\n" + s
	}
	if adoc != "" {
		docs = append(docs, adoc)
	}
	return docs
}
