package qliksense

import (
	"io"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (q *Qliksense) ApplyCRFromReader(r io.Reader) error {
	if err := q.LoadCr(r); err != nil {
		return err
	}
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	cr, err := qConfig.GetCurrentCR()
	if err != nil {
		return err
	}
	opts := &InstallCommandOptions{}
	if err := q.InstallQK8s(cr.GetLabelFromCr("version"), opts, true); err != nil {
		return err
	}
	return nil
}
