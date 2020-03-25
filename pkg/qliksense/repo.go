package qliksense

import (
	"errors"
	"path/filepath"

	kapis_git "github.com/qlik-oss/k-apis/pkg/git"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (q *Qliksense) DiscardAllUnstagedChangesFromGitRepo(qConfig *qapi.QliksenseConfig) error {
	if qcr, err := qConfig.GetCurrentCR(); err != nil {
		return err
	} else if version := qcr.GetLabelFromCr("version"); version == "" {
		return errors.New("version label is not set in CR")
	} else if qConfig.GetCrManifestRoot(qcr) == qConfig.BuildRepoPath(version) {
		if repo, err := kapis_git.OpenRepository(qConfig.GetCrManifestRoot(qcr)); err != nil {
			return err
		} else if err = kapis_git.DiscardAllUnstagedChanges(repo); err != nil {
			return err
		}
	}
	return nil
}

func (q *Qliksense) GetCrManifestRoot(cr *qapi.QliksenseCR) string {
	if filepath.IsAbs(cr.Spec.GetManifestsRoot()) {
		return cr.Spec.GetManifestsRoot()
	}
	return filepath.Join(q.QliksenseHome, cr.Spec.GetManifestsRoot())
}
