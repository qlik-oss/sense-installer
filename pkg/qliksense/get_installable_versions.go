package qliksense

import (
	"errors"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/qlik-oss/k-apis/pkg/git"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

type LsRemoteCmdOptions struct {
	IncludeBranches bool
	Limit           int
}

func (q *Qliksense) GetInstallableVersions(opts *LsRemoteCmdOptions) error {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		return err
	}

	var repoPath string
	if qcr.Spec.GetManifestsRoot() != "" {
		repoPath = qcr.Spec.GetManifestsRoot()
	} else {
		repoPath, err = DownloadFromGitRepoToTmpDir(defaultConfigRepoGitUrl, "master")
		if err != nil {
			return err
		}
	}

	r, err := git.OpenRepository(repoPath)
	if err != nil {
		return err
	}

	remoteRefsList, err := git.GetRemoteRefs(r, nil,
		&git.RemoteRefConstraints{
			Include:   true,
			Sort:      true,
			SortOrder: git.RefSortOrderDescending,
		},
		&git.RemoteRefConstraints{
			Include:   opts.IncludeBranches,
			Sort:      true,
			SortOrder: git.RefSortOrderAscending,
		})
	if err != nil {
		return err
	}

	if len(remoteRefsList) < 1 {
		return errors.New("cannot find git remote information in the config repository")
	}

	var originRemoteRefs *git.RemoteRefs
	for _, remoteRefs := range remoteRefsList {
		if remoteRefs.Name == "origin" {
			originRemoteRefs = remoteRefs
			break
		}
	}
	if originRemoteRefs == nil {
		return errors.New(`cannot find git remote called "origin" in the config repository`)
	}

	tags := originRemoteRefs.Tags
	if len(tags) > opts.Limit {
		tags = tags[:opts.Limit]
	}
	fmt.Print("Versions:\n")
	for _, tag := range tags {
		fmt.Printf(" %s\n", tag)
	}
	if opts.IncludeBranches {
		branches := originRemoteRefs.Branches
		if len(branches) > opts.Limit {
			branches = branches[:opts.Limit]
		}
		fmt.Print("Branches:\n")
		for _, branch := range branches {
			fmt.Printf(" %s\n", branch)
		}
	}

	return nil
}

func getLatestTag(repoUrl, accessToken string) (string, error) {
	if repoUrl == "" {
		return "", errors.New("repo url is empty")
	}
	repoPath, err := fetchToTempDir(repoUrl, "master", accessToken)
	if err != nil {
		return "", err
	}

	r, err := git.OpenRepository(repoPath)
	if err != nil {
		return "", err
	}

	remoteRefsList, err := git.GetRemoteRefs(r, nil,
		&git.RemoteRefConstraints{
			Include:   true,
			Sort:      true,
			SortOrder: git.RefSortOrderDescending,
		},
		&git.RemoteRefConstraints{
			Include:   false,
			Sort:      true,
			SortOrder: git.RefSortOrderAscending,
		})
	if err != nil {
		return "", err
	}

	if len(remoteRefsList) < 1 {
		return "", errors.New("cannot find git remote information in the config repository")
	}

	var originRemoteRefs *git.RemoteRefs
	for _, remoteRefs := range remoteRefsList {
		if remoteRefs.Name == "origin" {
			originRemoteRefs = remoteRefs
			break
		}
	}
	if originRemoteRefs == nil {
		return "", errors.New(`cannot find git remote called "origin" in the config repository`)
	}

	tags := originRemoteRefs.Tags
	if len(tags) == 0 {
		return "", errors.New(("no tags exists in the repo: " + repoPath))
	}
	maxSem, _ := semver.NewVersion(tags[0])
	for _, sv := range tags[1:] {
		if sv == "" {
			continue
		}
		v, err := semver.NewVersion(sv)
		if err != nil {
			// it may happen, in the repo some tags may not conform to semver
			fmt.Print("Unconform tags: " + sv)
			continue
		}
		if maxSem == nil || maxSem.LessThan(v) {
			maxSem = v
		}
	}
	return maxSem.String(), nil
}
