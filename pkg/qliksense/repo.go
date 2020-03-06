package qliksense

import (
	"errors"

	kapis_git "github.com/qlik-oss/k-apis/pkg/git"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (q *Qliksense) DiscardAllUnstagedChangesFromGitRepo(qConfig *qapi.QliksenseConfig) error {
	if qcr, err := qConfig.GetCurrentCR(); err != nil {
		return err
	} else if version := qcr.GetLabelFromCr("version"); version == "" {
		return errors.New("version label is not set in CR")
	} else if qcr.Spec.ManifestsRoot == qConfig.BuildRepoPath(version) {
		if repo, err := kapis_git.OpenRepository(qcr.Spec.ManifestsRoot); err != nil {
			return err
		} else if err = kapis_git.DiscardAllUnstagedChanges(repo); err != nil {
			return err
		}
	}
	return nil
}
