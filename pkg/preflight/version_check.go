package preflight

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/qlik-oss/sense-installer/pkg/api"
	"k8s.io/apimachinery/pkg/version"
)

func (qp *QliksensePreflight) CheckK8sVersion(namespace string, kubeConfigContents []byte) error {

	var currentVersion *semver.Version

	clientset, _, err := getK8SClientSet(kubeConfigContents, "")
	if err != nil {
		err = fmt.Errorf("Unable to create clientset: %v\n", err)
		return err
	}
	var serverVersion *version.Info
	if err := retryOnError(func() (err error) {
		serverVersion, err = clientset.ServerVersion()
		return err
	}); err != nil {
		err = fmt.Errorf("Unable to get server version: %v\n", err)
		//fmt.Println(err)
		return err
	}
	qp.P.LogVerboseMessage("Kubernetes API Server version: %s\n", serverVersion.String())

	// Compare K8s version on the cluster with minimum supported k8s version
	currentVersion, err = semver.NewVersion(serverVersion.String())
	if err != nil {
		err = fmt.Errorf("Unable to convert server version into semver version: %v\n", err)
		//fmt.Println(err)
		return err
	}
	api.LogDebugMessage("Current Kubernetes Version: %v\n", currentVersion)

	minK8sVersionSemver, err := semver.NewVersion(qp.GetPreflightConfigObj().GetMinK8sVersion())
	// minK8sVersionSemver, err := semver.NewVersion("v1.17.7")
	if err != nil {
		err = fmt.Errorf("Unable to convert minimum Kubernetes version into semver version:%v\n", err)
		fmt.Println(err)
		return err
	}

	if currentVersion.GreaterThan(minK8sVersionSemver) {
		//fmt.Printf("\n\nCurrent %s Component version: %s is less than minimum required version:%s\n", component, currentComponentVersion, componentVersionFromDependenciesYaml)
		qp.P.LogVerboseMessage("Current Kubernetes API Server version %s is greater than or equal to minimum required version: %s\n", currentVersion, minK8sVersionSemver)
		// qp.P.LogVerboseMessage("Preflight minimum kubernetes version check: PASSED\n")
	} else {
		// qp.P.LogVerboseMessage("Current %s is less than minimum required version:%s\n", currentVersion, minK8sVersionSemver)
		err = fmt.Errorf("Current Kubernetes API Server version %s is less than minimum required version: %s", currentVersion, minK8sVersionSemver)
		return err
		// qp.P.LogVerboseMessage("Preflight minimum kubernetes version check: FAILED\n")
	}
	// qp.P.LogVerboseMessage("Completed Preflight kubernetes minimum version check\n")
	return nil
}
