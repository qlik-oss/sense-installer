package preflight

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"

	"github.com/qlik-oss/sense-installer/pkg/api"
)

const dnsCheckYAML = `
apiVersion: troubleshoot.replicated.com/v1beta1
kind: Preflight
metadata:
  name: cluster-preflight-checks
  namespace: {{ . }}
spec:
  collectors:
    - run: 
        collectorName: spin-up-pod
        args: ["-z", "-v", "-w 1", "qnginx001", "80"]
        command: ["nc"]
        image: subfuzion/netcat:latest
        imagePullPolicy: IfNotPresent
        name: spin-up-pod-check-dns
        namespace: {{ . }}
        timeout: 30s

  analyzers:
    - textAnalyze:
        checkName: DNS check
        collectorName: spin-up-pod-check-dns
        fileName: spin-up-pod.log
        regex: succeeded
        outcomes:
          - fail:
              message: DNS check failed
          - pass:
              message: DNS check passed
`

func (qp *QliksensePreflight) CheckDns() error {
	// retrieve namespace
	namespace := api.GetKubectlNamespace()

	api.LogDebugMessage("Namespace: %s\n", namespace)

	tmpl, err := template.New("dnsCheckYAML").Parse(dnsCheckYAML)
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

	// creating Kubectl resources
	appName := "qnginx001"
	const PreflightChecksDirName = "preflight_checks"

	fmt.Println("Creating resources to run preflight checks")

	// kubectl create deployment
	opr := fmt.Sprintf("create deployment %s --image=nginx", appName)
	err = initiateK8sOps(opr, namespace)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer func() {
		// Deleting deployment..
		opr = fmt.Sprintf("delete deployment %s", appName)
		// we want to delete the k8s resource here, we dont care a lot about an error here
		_ = initiateK8sOps(opr, namespace)
		api.LogDebugMessage("delete deployment executed")
	}()

	// create service
	opr = fmt.Sprintf("create service clusterip %s --tcp=80:80", appName)
	err = initiateK8sOps(opr, namespace)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer func() {
		// delete service
		opr = fmt.Sprintf("delete service %s", appName)
		// we want to delete the k8s resource here, we dont care a lot about an error here
		_ = initiateK8sOps(opr, namespace)
		api.LogDebugMessage("delete service executed")
	}()

	//kubectl -n $namespace wait --for=condition=ready pod -l app=$appName --timeout=120s
	opr = fmt.Sprintf("wait --for=condition=ready pod -l app=%s --timeout=120s", appName)
	err = initiateK8sOps(opr, namespace)
	if err != nil {
		fmt.Println(err)
		return err
	}
	api.LogDebugMessage("kubectl wait executed")

	// call preflight
	preflightCommand := filepath.Join(qp.Q.QliksenseHome, PreflightChecksDirName, preflightFileName)

	err = invokePreflight(preflightCommand, tempYaml)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("DNS check completed, cleaning up resources now")
	return nil
}
