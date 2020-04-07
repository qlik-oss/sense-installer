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
	if IsQliksenseInstalled(cr.GetName()) {
		// it is needed in case want to upgrade from one version to another
		if cr.Spec.ManifestsRoot == "" && cr.Spec.Git == nil {
			v := cr.GetLabelFromCr("version")
			if !qConfig.IsRepoExistForCurrent(v) {
				if err := q.FetchQK8s(v); err != nil {
					return err
				}
			}
		}
		return q.UpgradeQK8s(keepPatchFiles)
	}
	return q.InstallQK8s(cr.GetLabelFromCr("version"), opts, keepPatchFiles)
}

func IsQliksenseInstalled(crName string) bool {
	args := []string{
		"get",
		"qliksense",
		crName,
		"-ogo-template",
		`--template='{{ .metadata.name}}'`,
	}
	_, err := qapi.KubectlDirectOps(args, "")
	if err != nil {
		return false
	}
	return true
}
