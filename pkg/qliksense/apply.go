package qliksense

import (
	"fmt"
	"io"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (q *Qliksense) ApplyCRFromReader(r io.Reader, opts *InstallCommandOptions, cleanPatchFiles, overwriteExistingContext, pull, push bool) error {
	if err := q.LoadCr(r, overwriteExistingContext); err != nil {
		return err
	}
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	cr, err := qConfig.GetCurrentCR()
	if err != nil {
		return err
	}
	version := cr.GetLabelFromCr("version")

	if pull {
		fmt.Println("Pulling images...")
		if err := q.PullImages(version, ""); err != nil {
			return err
		}
	}
	if push {
		fmt.Println("Pushing images...")
		if err := q.PushImagesForCurrentCR(); err != nil {
			return err
		}
	}

	if IsQliksenseInstalled(cr.GetName()) {
		// it is needed in case want to upgrade from one version to another
		if cr.Spec.ManifestsRoot == "" && cr.Spec.Git == nil {
			if !qConfig.IsRepoExistForCurrent(version) {
				if err := q.FetchQK8s(version); err != nil {
					return err
				}
			}
		}
		return q.UpgradeQK8s(cleanPatchFiles)
	}
	return q.InstallQK8s(version, opts, cleanPatchFiles)
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
