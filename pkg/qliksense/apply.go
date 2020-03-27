package qliksense

import (
	"io"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (q *Qliksense) ApplyCRFromReader(r io.Reader, opts *InstallCommandOptions, keepPatchFiles, overwriteExistingContext bool) error {
	if err := q.LoadCr(r, overwriteExistingContext); err != nil {
		return err
	}
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	cr, err := qConfig.GetCurrentCR()
	if err != nil {
		return err
	}
	if err := q.InstallQK8s(cr.GetLabelFromCr("version"), opts, keepPatchFiles); err != nil {
		return err
	}
	return nil
}
