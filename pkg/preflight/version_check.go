package preflight

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/qlik-oss/sense-installer/pkg/api"
	"k8s.io/apimachinery/pkg/version"
)

func (qp *QliksensePreflight) CheckK8sVersion(namespace string, kubeConfigContents []byte) error {
	qp.CG.LogVerboseMessage("Preflight kubernetes version check: \n")
	qp.CG.LogVerboseMessage("----------------------------------- \n")
	var currentVersion *semver.Version

	clientset, _, err := api.GetK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("Unable to create clientset: %v\n", err)
		return err
	}
	var serverVersion *version.Info
	if err := api.RetryOnError(func() (err error) {
		serverVersion, err = clientset.ServerVersion()
		return err
	}); err != nil {
		err = fmt.Errorf("Unable to get server version: %v\n", err)
		return err
	}
	qp.CG.LogVerboseMessage("Kubernetes API Server version: %s\n", serverVersion.String())

	// Compare K8s version on the cluster with minimum supported k8s version
	currentVersion, err = semver.NewVersion(serverVersion.String())
	if err != nil {
		err = fmt.Errorf("Unable to convert server version into semver version: %v\n", err)
		return err
	}
	api.LogDebugMessage("Current Kubernetes Version: %v\n", currentVersion)

	minK8sVersionSemver, err := semver.NewVersion(qp.GetPreflightConfigObj().GetMinK8sVersion())
	if err != nil {
		err = fmt.Errorf("Unable to convert minimum Kubernetes version into semver version:%v\n", err)
		return err
	}

	if currentVersion.GreaterThan(minK8sVersionSemver) {
		qp.CG.LogVerboseMessage("Current Kubernetes API Server version %s is greater than or equal to minimum required version: %s\n", currentVersion, minK8sVersionSemver)
	} else {
		err = fmt.Errorf("Current Kubernetes API Server version %s is less than minimum required version: %s", currentVersion, minK8sVersionSemver)
		return err
	}
	return nil
}
