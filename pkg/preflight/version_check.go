package preflight

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/qlik-oss/sense-installer/pkg/api"
)

const minK8sVersion = "1.11.0"
const checkVersionYAML = `
apiVersion: troubleshoot.replicated.com/v1beta1
kind: Preflight
metadata:
  name: cluster-preflight-checks
  namespace: {{ .namespace }}
spec:
  analyzers:
    - clusterVersion:
        outcomes:
          - fail:
              when: "< {{ .minK8sVersion }}"
              message: The application requires at least Kubernetes {{ .minK8sVersion }} or later.
              uri: https://www.kubernetes.io
          - pass:
              when: ">= {{ .minK8sVersion }}"
              message: Good to go.
`

func (qp *QliksensePreflight) CheckK8sVersion() error {
	// retrieve namespace
	namespace := api.GetKubectlNamespace()

	api.LogDebugMessage("Namespace: %s\n", namespace)

	tmpl, err := template.New("checkVersionYAML").Parse(checkVersionYAML)
	if err != nil {
		fmt.Printf("cannot parse template: %v", err)
		return err
	}
	tempYaml, err := ioutil.TempFile("", "")
	if err != nil {
		fmt.Printf("cannot create file: %v", err)
		return err
	}
	api.LogDebugMessage("Temp Yaml file: %s\n", tempYaml.Name())

	b := bytes.Buffer{}
	err = tmpl.Execute(&b, map[string]string{
		"namespace":     namespace,
		"minK8sVersion": minK8sVersion,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	tempYaml.WriteString(b.String())
	//api.LogDebugMessage("Temp yaml contents: %s", b.String())
	fmt.Printf("Minimum Kubernetes version supported: %s\n", minK8sVersion)

	// current kubectl version
	opr := fmt.Sprintf("version")
	err = initiateK8sOps(opr, namespace)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// call preflight
	preflightCommand := filepath.Join(qp.Q.QliksenseHome, PreflightChecksDirName, preflightFileName)

	err = invokePreflight(preflightCommand, tempYaml)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("Minimum kubernetes version check completed")
	return nil
}
