package qliksense

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"

	kapis_git "github.com/qlik-oss/k-apis/pkg/git"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/src-d/go-git/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

type FetchCommandOptions struct {
	GitUrl      string
	AccessToken string
	Version     string
	SecretName  string
}

const (
	QLIK_GIT_REPO = "https://github.com/qlik-oss/qliksense-k8s"
)

func (q *Qliksense) FetchQK8s(version string) error {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	return fetchAndUpdateCR(qConfig, version)
}

func (q *Qliksense) FetchK8sWithOpts(opts *FetchCommandOptions) error {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	cr, err := qConfig.GetCurrentCR()
	if err != nil {
		return err
	}
	if opts.AccessToken != "" {
		cr.SetFetchAccessToken(opts.AccessToken)
	}
	if opts.SecretName != "" {
		cr.SetFetchAccessSecretName(opts.SecretName)
	}
	if opts.GitUrl != "" {
		cr.SetFetchUrl(opts.GitUrl)
	}
	qConfig.WriteCR(cr)
	return fetchAndUpdateCR(qConfig, opts.Version)
}

// fetchAndUpdateCR fetch
func fetchAndUpdateCR(qConfig *qapi.QliksenseConfig, version string) error {
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		fmt.Println("cannot get the current-context cr", err)
		return err
	}
	if version == "" {
		if qcr.GetLabelFromCr("version") == "" {
			return errors.New("Cannot find gitref/tag/branch/version to fetch")
		}
		version = qcr.GetLabelFromCr("version")
	}
	// downlaod to temp first
	tempDest, err := fetchToTempDir(qcr.GetFetchUrl(), version, qcr.GetFetchAccessToken())
	if err != nil {
		return err
	}

	if err := qConfig.DeleteRepoForCurrent(version); err != nil {
		return err
	}

	destDir := qConfig.BuildRepoPath(version)
	fmt.Printf("fetching version [%s] from %s\n", version, qcr.GetFetchUrl())
	if err := qapi.CopyDirectory(tempDest, destDir); err != nil {
		return nil
	}
	qcr.Spec.ManifestsRoot = qConfig.BuildCurrentManifestsRoot(version)
	qcr.AddLabelToCr("version", version)
	return qConfig.WriteCurrentContextCR(qcr)
}

func fetchToTempDir(gitUrl, gitRef, accessToken string) (string, error) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	downloadPath := path.Join(tmpDir, "repo")
	var auth transport.AuthMethod
	if accessToken != "" {
		auth = &http.BasicAuth{
			Username: "something",
			Password: accessToken,
		}
	}
	if repo, err := kapis_git.CloneRepository(downloadPath, gitUrl, auth); err != nil {
		return "", err
	} else if err := kapis_git.Checkout(repo, gitRef, "", auth); err != nil {
		return "", err
	} else {
		return downloadPath, nil
	}
}
