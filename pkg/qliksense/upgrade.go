package qliksense

import (
	"fmt"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (q *Qliksense) UpgradeQK8s(keepPatchFiles bool) error {

	// step1: get CR
	// step2: run kustomize
	// step3: run kubectl apply

	// fetch the version
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	if !keepPatchFiles {
		defer func() {
			if err := q.DiscardAllUnstagedChangesFromGitRepo(qConfig); err != nil {
				fmt.Printf("error removing temporary changes to the config: %v\n", err)
			}
		}()
	}

	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	}
	qcr.Spec.RotateKeys = "no"

	dcr, err := qConfig.GetDecryptedCr(qcr)
	if err != nil {
		return err
	}
	if dcr.Spec.Git != nil && dcr.Spec.Git.Repository != "" {
		// fetching and applying manifest will be in the operator controller
		// get decrypted cr
		return q.applyCR(dcr)
	}
	err = q.applyConfigToK8s(dcr)
	if err != nil {
		fmt.Println("cannot do kubectl apply on manifests")
		return err
	}
	return q.applyCR(dcr)
}
