package preflight

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/qlik-oss/sense-installer/pkg/api"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
)

func (qp *QliksensePreflight) CheckCreateRole(namespace string, cleanup bool) error {
	// create a Role
	if !cleanup {
		qp.CG.LogVerboseMessage("Preflight role check: \n")
		qp.CG.LogVerboseMessage("--------------------- \n")
	}
	err := qp.checkCreateEntity(namespace, "Role", cleanup)
	if err != nil {
		return err
	}
	if !cleanup {
		qp.CG.LogVerboseMessage("Completed preflight role check\n")
	}
	return nil
}

func (qp *QliksensePreflight) CheckCreateRoleBinding(namespace string, cleanup bool) error {
	// create a RoleBinding
	if !cleanup {
		qp.CG.LogVerboseMessage("Preflight rolebinding check: \n")
		qp.CG.LogVerboseMessage("---------------------------- \n")
	}
	err := qp.checkCreateEntity(namespace, "RoleBinding", cleanup)
	if err != nil {
		return err
	}
	if !cleanup {
		qp.CG.LogVerboseMessage("Completed preflight rolebinding check\n")
	}
	return nil
}

func (qp *QliksensePreflight) CheckCreateServiceAccount(namespace string, cleanup bool) error {
	// create a service account
	if !cleanup {
		qp.CG.LogVerboseMessage("Preflight serviceaccount check: \n")
		qp.CG.LogVerboseMessage("------------------------------- \n")
	}
	err := qp.checkCreateEntity(namespace, "ServiceAccount", cleanup)
	if err != nil {
		return err
	}
	if !cleanup {
		qp.CG.LogVerboseMessage("Completed preflight serviceaccount check\n")
	}
	return nil
}
func (qp *QliksensePreflight) checkCreateEntity(namespace, entityToTest string, cleanup bool) error {
	qConfig := qapi.NewQConfig(qp.Q.QliksenseHome)
	var currentCR *qapi.QliksenseCR
	mfroot := ""
	kusDir := ""
	resultYamlBytes := []byte("")
	var err error
	currentCR, err = qConfig.GetCurrentCR()
	if err != nil {
		qp.CG.LogVerboseMessage("Unable to retrieve current CR: %v\n", err)
		return err
	}
	if currentCR.IsRepoExist() {
		mfroot = currentCR.Spec.GetManifestsRoot()
	} else if tempDownloadedDir, err := qliksense.DownloadFromGitRepoToTmpDir(qliksense.QLIK_GIT_REPO, "master"); err != nil {
		qp.CG.LogVerboseMessage("Unable to Download from git repo to tmp dir: %v\n", err)
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
			err := fmt.Errorf("Unable to retrieve manifests from executing kustomize: %s, error: %v", kusDir, err)
			return err
		}
	}
	sa := qliksense.GetYamlsFromMultiDoc(string(resultYamlBytes), entityToTest)
	if sa != "" {
		sa = strings.Replace(sa, "name: qliksense", "name: preflight", -1)
	} else {
		err := fmt.Errorf("Unable to retrieve yamls to apply on cluster from dir: %s, error: %v", kusDir, err)
		return err
	}
	namespace = "" // namespace is handled when generating the manifests

	// check if entity already exists in the cluster, if so - delete it
	api.KubectlDeleteVerbose(sa, namespace, qp.P.Verbose)
	if cleanup {
		return nil
	}

	defer func() {
		qp.CG.LogVerboseMessage("Cleaning up resources...\n")
		err := api.KubectlDeleteVerbose(sa, namespace, qp.P.Verbose)
		if err != nil {
			qp.CG.LogVerboseMessage("Preflight cleanup failed!\n")
		}
	}()

	err = api.KubectlApplyVerbose(sa, namespace, qp.P.Verbose)
	if err != nil {
		err := fmt.Errorf("Failed to create entity on the cluster: %v", err)
		return err
	}

	qp.CG.LogVerboseMessage("Preflight %s check: PASSED\n", entityToTest)
	return nil
}

func (qp *QliksensePreflight) CheckCreateRB(namespace string, kubeConfigContents []byte) error {

	// create a role
	qp.CG.LogVerboseMessage("Preflight createRole check: \n")
	qp.CG.LogVerboseMessage("--------------------------- \n")
	errStr := strings.Builder{}
	err1 := qp.checkCreateEntity(namespace, "Role", false)
	if err1 != nil {
		errStr.WriteString(err1.Error())
		errStr.WriteString("\n")
		qp.CG.LogVerboseMessage("%v\n", err1)
		qp.CG.LogVerboseMessage("Preflight role check: FAILED\n")
	}
	qp.CG.LogVerboseMessage("Completed preflight role check\n\n")

	// create a roleBinding
	qp.CG.LogVerboseMessage("Preflight rolebinding check: \n")
	qp.CG.LogVerboseMessage("---------------------------- \n")
	err2 := qp.checkCreateEntity(namespace, "RoleBinding", false)
	if err2 != nil {
		errStr.WriteString(err2.Error())
		errStr.WriteString("\n")
		qp.CG.LogVerboseMessage("%v\n", err2)
		qp.CG.LogVerboseMessage("Preflight rolebinding check: FAILED\n")
	}
	qp.CG.LogVerboseMessage("Completed preflight rolebinding check\n\n")

	// create a service account
	qp.CG.LogVerboseMessage("Preflight serviceaccount check: \n")
	qp.CG.LogVerboseMessage("------------------------------- \n")
	err3 := qp.checkCreateEntity(namespace, "ServiceAccount", false)
	if err3 != nil {
		errStr.WriteString(err3.Error())
		errStr.WriteString("\n")
		qp.CG.LogVerboseMessage("%v\n", err3)
		qp.CG.LogVerboseMessage("Preflight serviceaccount check: FAILED\n")
	}
	qp.CG.LogVerboseMessage("Completed preflight serviceaccount check\n\n")

	if err1 != nil || err2 != nil || err3 != nil {
		qp.CG.LogVerboseMessage("Preflight authcheck: FAILED\n")
		qp.CG.LogVerboseMessage("Completed preflight authcheck\n")
		return errors.New(errStr.String())
	}
	qp.CG.LogVerboseMessage("Preflight authcheck: PASSED\n")
	qp.CG.LogVerboseMessage("Completed preflight authcheck\n")
	return nil
}
