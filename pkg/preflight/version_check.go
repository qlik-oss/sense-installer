package preflight

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"

	"github.com/qlik-oss/sense-installer/pkg/api"
)

const checkVersionYAML = `
apiVersion: troubleshoot.replicated.com/v1beta1
kind: Preflight
metadata:
  name: cluster-preflight-checks
  namespace: {{ . }}
spec:
  analyzers:
     - clusterVersion:
        outcomes:
          - fail:
              when: "<= 1.13.0"
              message: The application requires at Kubernetes 1.13.0 or later, and recommends 1.15.0.
              uri: https://www.kubernetes.io
          - warn:
              when: "< 1.13.1"
              message: Your cluster meets the minimum version of Kubernetes, but we recommend you update to 1.15.0 or later.
              uri: https://kubernetes.io
          - pass:
              when: ">= 1.13.0"
              message: Good to go.
`

func (qp *QliksensePreflight) CheckK8sVersion() error {
	// retrieve namespace
	namespace := api.GetKubectlNamespace()

	api.LogDebugMessage("Namespace: %s\n", namespace)

	tmpl, err := template.New("test").Parse(checkVersionYAML)
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
	err = tmpl.Execute(&b, namespace)
	if err != nil {
		fmt.Println(err)
		return err
	}

	tempYaml.WriteString(b.String())

	// call preflight
	preflightCommand := filepath.Join(qp.Q.QliksenseHome, PreflightChecksDirName, preflightFileName)

	err = invokePreflight(preflightCommand, tempYaml)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
