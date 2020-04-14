package qliksense

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	kapis_git "github.com/qlik-oss/k-apis/pkg/git"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

type FetchCommandOptions struct {
	GitUrl      string
	AccessToken string
	Version     string
	SecretName  string
	Overwrite   bool
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
		encKey, err := qConfig.GetEncryptionKeyFor(cr.GetName())
		if err != nil {
			return err
		}
		if err := cr.SetFetchAccessToken(opts.AccessToken, encKey); err != nil {
			return err
		}
	}
	if opts.SecretName != "" {
		cr.SetFetchAccessSecretName(opts.SecretName)
	}
	if opts.GitUrl != "" {
		cr.SetFetchUrl(opts.GitUrl)
	}
	v := getVersion(opts, cr)
	if v == "" {
		return errors.New("Cannot find gitref/tag/branch/version to fetch")
	}
	if qConfig.IsRepoExistForCurrent(v) {
		if opts.Overwrite || getVerionsOverwriteConfirmation(v) == "y" {
			if err := qConfig.DeleteRepoForCurrent(v); err != nil {
				return err
			}
		} else {
			// nothing to do
			return nil
		}
	}
	qConfig.WriteCR(cr)
	return fetchAndUpdateCR(qConfig, v)
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
	encKey, err := qConfig.GetEncryptionKeyFor(qcr.GetName())
	if err != nil {
		return err
	}
	// downlaod to temp first
	tempDest, err := fetchToTempDir(qcr.GetFetchUrl(), version, qcr.GetFetchAccessToken(encKey))
	if err != nil {
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

func getVersion(opts *FetchCommandOptions, qcr *qapi.QliksenseCR) string {
	if opts.Version == "" {
		if qcr.GetLabelFromCr("version") != "" {
			return qcr.GetLabelFromCr("version")
		}
	}
	return opts.Version
}

func getVerionsOverwriteConfirmation(version string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("The version  [" + version + "] already exist")
	cfm := "n"
	for {
		fmt.Print("Do you want to delete and fetch again [y/N]: ")
		cfm, _ = reader.ReadString('\n')
		cfm = strings.Replace(cfm, "\n", "", -1)
		cfm = strings.TrimSpace(cfm)
		if cfm == "" {
			cfm = "n"
		}
		cfm = strings.ToLower(cfm)
		if cfm == "y" || cfm == "n" {
			break
		}
	}
	return cfm
}
