package qliksense

import (
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/qlik-oss/k-apis/pkg/cr"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (q *Qliksense) RotateKeys() error {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	if qcr, err := qConfig.GetCurrentCR(); err != nil {
		return err
	} else if userHomeDir, err := homedir.Dir(); err != nil {
		return err
	} else if err := cr.DeleteKeysClusterBackup(&qcr.KApiCr, path.Join(userHomeDir, ".kube", "config")); err != nil {
		return err
	}
	return nil
}
