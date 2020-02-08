package qliksense

import (
	"fmt"
	kapis_git "github.com/qlik-oss/k-apis/pkg/git"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

const (
	QLIK_GIT_REPO = "https://github.com/qlik-oss/qliksense-k8s"
)

func (q *Qliksense) FetchQK8s(version string) error {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	return fetchAndUpdateCR(qConfig, version)
}

func fetchAndUpdateCR(qConfig *qapi.QliksenseConfig, version string) error {
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	}
	if qConfig.IsRepoExistForCurrent(version) {
		return nil
	}
	destDir := qConfig.BuildRepoPath(version)
	fmt.Printf("fetching version [%s] from %s\n", version, QLIK_GIT_REPO)

	if repo, err := kapis_git.CloneRepository(destDir, QLIK_GIT_REPO, nil); err != nil {
		return err
	} else if err = kapis_git.Checkout(repo, version, version, nil); err != nil {
		return err
	}
	qcr.Spec.ManifestsRoot = qConfig.BuildCurrentManifestsRoot(version)
	qcr.AddLabelToCr("version", version)
	return qConfig.WriteCurrentContextCR(qcr)
}
