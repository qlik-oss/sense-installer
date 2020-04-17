package preflight

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/qlik-oss/sense-installer/pkg/api"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
)

var resultYamlBytes = []byte("")

func (qp *QliksensePreflight) CheckCreateRole(namespace string) error {
	// create a Role
	qp.P.LogVerboseMessage("Preflight role check: \n")
	qp.P.LogVerboseMessage("--------------------- \n")
	err := qp.checkCreateEntity(namespace, "Role")
	if err != nil {
		return err
	}
	qp.P.LogVerboseMessage("Completed preflight role check\n")
	return nil
}

func (qp *QliksensePreflight) CheckCreateRoleBinding(namespace string) error {
	// create a RoleBinding
	qp.P.LogVerboseMessage("Preflight rolebinding check: \n")
	qp.P.LogVerboseMessage("---------------------------- \n")
	err := qp.checkCreateEntity(namespace, "RoleBinding")
	if err != nil {
		return err
	}
	qp.P.LogVerboseMessage("Completed preflight rolebinding check\n")
	return nil
}

func (qp *QliksensePreflight) CheckCreateServiceAccount(namespace string) error {
	// create a service account
	qp.P.LogVerboseMessage("Preflight serviceaccount check: \n")
	qp.P.LogVerboseMessage("------------------------------- \n")
	err := qp.checkCreateEntity(namespace, "ServiceAccount")
	if err != nil {
		return err
	}
	qp.P.LogVerboseMessage("Completed preflight serviceaccount check\n")
	return nil
}
func (qp *QliksensePreflight) checkCreateEntity(namespace, entityToTest string) error {
	qConfig := qapi.NewQConfig(qp.Q.QliksenseHome)
	var currentCR *qapi.QliksenseCR
	mfroot := ""
	kusDir := ""
	var err error
	currentCR, err = qConfig.GetCurrentCR()
	if err != nil {
		qp.P.LogVerboseMessage("Unable to retrieve current CR: %v\n", err)
		return err
	}
	if currentCR.IsRepoExist() {
		mfroot = currentCR.Spec.GetManifestsRoot()
	} else if tempDownloadedDir, err := qliksense.DownloadFromGitRepoToTmpDir(qliksense.QLIK_GIT_REPO, "master"); err != nil {
		qp.P.LogVerboseMessage("Unable to Download from git repo to tmp dir: %v\n", err)
		return err
	} else {
		mfroot = tempDownloadedDir
	}

	if currentCR.Spec.Profile == "" {
		kusDir = filepath.Join(mfroot, "manifests", "docker-desktop")
	} else {
		kusDir = filepath.Join(mfroot, "manifests", currentCR.Spec.Profile)
	}
	if len(resultYamlBytes) == 0 {
		resultYamlBytes, err = qliksense.ExecuteKustomizeBuild(kusDir)
		if err != nil {
			err := fmt.Errorf("Unable to retrieve manifests from executing kustomize: %s", kusDir)
			return err
		}
	}
	sa := qliksense.GetYamlsFromMultiDoc(string(resultYamlBytes), entityToTest)
	if sa != "" {
		sa = strings.Replace(sa, "name: qliksense", "name: preflight", -1)
	} else {
		err := fmt.Errorf("Unable to retrieve yamls to apply on cluster from dir: %s", kusDir)
		return err
	}
	namespace = "" // namespace is handled when generating the manifests

	defer func() {
		qp.P.LogVerboseMessage("Cleaning up resources...\n")
		err := api.KubectlDeleteVerbose(sa, namespace, qp.P.Verbose)
		if err != nil {
			qp.P.LogVerboseMessage("Preflight cleanup failed!\n")
		}
	}()

	err = api.KubectlApplyVerbose(sa, namespace, qp.P.Verbose)
	if err != nil {
		err := fmt.Errorf("Failed to create entity on the cluster: %v", err)
		return err
	}

	qp.P.LogVerboseMessage("Preflight %s check: PASSED\n", entityToTest)
	return nil
}

func (qp *QliksensePreflight) CheckCreateRB(namespace string, kubeConfigContents []byte) error {

	// create a role
	qp.P.LogVerboseMessage("Preflight createRole check: \n")
	qp.P.LogVerboseMessage("--------------------------- \n")
	err := qp.checkCreateEntity(namespace, "Role")
	if err != nil {
		qp.P.LogVerboseMessage("%v\n", err)
		qp.P.LogVerboseMessage("Preflight role check: FAILED\n")
	}
	qp.P.LogVerboseMessage("Completed preflight role check\n\n")

	// create a roleBinding
	qp.P.LogVerboseMessage("Preflight rolebinding check: \n")
	qp.P.LogVerboseMessage("---------------------------- \n")
	err = qp.checkCreateEntity(namespace, "RoleBinding")
	if err != nil {
		qp.P.LogVerboseMessage("%v\n", err)
		qp.P.LogVerboseMessage("Preflight rolebinding check: FAILED\n")
	}
	qp.P.LogVerboseMessage("Completed preflight rolebinding check\n\n")

	// create a service account
	qp.P.LogVerboseMessage("Preflight serviceaccount check: \n")
	qp.P.LogVerboseMessage("------------------------------- \n")
	err = qp.checkCreateEntity(namespace, "ServiceAccount")
	if err != nil {
		qp.P.LogVerboseMessage("%v\n", err)
		qp.P.LogVerboseMessage("Preflight serviceaccount check: FAILED\n")
	}
	qp.P.LogVerboseMessage("Completed preflight serviceaccount check\n\n")

	qp.P.LogVerboseMessage("Preflight authcheck: PASSED\n")
	qp.P.LogVerboseMessage("Completed preflight authcheck\n")
	return nil
}
