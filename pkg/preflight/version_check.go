package preflight

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"k8s.io/apimachinery/pkg/version"
)

const minK8sVersion = "1.11.0"

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
	fmt.Printf("Kubernetes API Server version: %s\n", serverVersion.String())

	// Compare K8s version on the cluster with minimum supported k8s version
	currentVersion, err = semver.NewVersion(serverVersion.String())
	if err != nil {
		err = fmt.Errorf("Unable to convert server version into semver version: %v\n", err)
		//fmt.Println(err)
		return err
	}
	fmt.Printf("Current K8s Version: %v\n", currentVersion)

	minK8sVersionSemver, err := semver.NewVersion(minK8sVersion)
	if err != nil {
		err = fmt.Errorf("Unable to convert minimum Kubernetes version into semver version:%v\n", err)
		fmt.Println(err)
		return err
	}

	if currentVersion.GreaterThan(minK8sVersionSemver) {
		//fmt.Printf("\n\nCurrent %s Component version: %s is less than minimum required version:%s\n", component, currentComponentVersion, componentVersionFromDependenciesYaml)
		fmt.Printf("Current %s is greater than minimum required version:%s, hence good to go\n", currentVersion, minK8sVersionSemver)
		fmt.Println("Preflight minimum kubernetes version check: PASS")
	} else {
		fmt.Printf("Current %s is less than minimum required version:%s\n", currentVersion, minK8sVersionSemver)
		fmt.Println("Preflight minimum kubernetes version check: FAIL")
	}
	fmt.Printf("Completed Preflight kubernetes minimum version check\n\n")
	return nil
}
