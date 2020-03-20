package qliksense

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"text/template"

	"github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/spf13/cobra"
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
    - deploymentStatus:
        checkName: check for deploymentStatus
        name: qnginx001
        namespace: {{ . }}
        outcomes:
          - fail:
              when: "= 0"
              message: deployment not found
          - pass:
              when: "> 0"
              message: deployment found
    - textAnalyze:
        checkName: DNS check
        collectorName: spin-up-pod-check-dns
        fileName: spin-up-pod.txt
        regex: succeeded
        outcomes:
          - fail:
              message: DNS check failed
          - pass:
              message: DNS check passed
`

// PerformDnsCheck
func PerformDnsCheck(q *Qliksense) *cobra.Command {
	var (
		cmd *cobra.Command
	)

	cmd = &cobra.Command{
		Use:     "preflight dns",
		Short:   "Perform preflight check on dns ",
		Example: `qliksense preflight --dns`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return q.checkDns()
		},
	}
	return cmd
}

func (q *Qliksense) checkDns() error {
	// retrieve namespace
	namespace := api.GetKubectlNamespace()
	api.LogDebugMessage("Namespace here: %s", namespace)

	tmpl, err := template.New("test").Parse(dnsCheckYAML)
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
	const preflightFileName = "preflight"

	// kubectl create deployment
	opr := fmt.Sprintf("create deployment %s --image=nginx", appName)
	err = initiateK8sOps(opr, namespace)
	if err != nil {
		fmt.Println(err)
		return err
	}
	api.LogDebugMessage("create deployment executed")

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
	api.LogDebugMessage("create service executed")

	defer func() {
		// Deleting service..
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

	// calling preflight here..
	preflightCommand := filepath.Join(q.QliksenseHome, PreflightChecksDirName, preflightFileName)
	trackSuccess, err := invokePreflight(preflightCommand, tempYaml)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if trackSuccess {
		fmt.Println("PREFLIGHT DNS CHECK PASSED")
	} else {
		fmt.Println("PREFLIGHT DNS CHECK FAILED")
	}

	return nil
}

func initiateK8sOps(opr, namespace string) error {
	opr1 := strings.Fields(opr)
	err := api.KubectlDirectOps(opr1, namespace)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func invokePreflight(preflightCommand string, yamlFile *os.File) (bool, error) {
	arguments := []string{}
	arguments = append(arguments, yamlFile.Name(), "--interactive=false")
	cmd := exec.Command(preflightCommand, arguments...)

	sterrBuffer := &bytes.Buffer{}
	cmd.Stdout = sterrBuffer
	cmd.Stderr = sterrBuffer
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("Error when running preflight command: %v\n", err)
	}
	ind := strings.Index(sterrBuffer.String(), "---")
	output := sterrBuffer.String()
	if ind > -1 {
		output = fmt.Sprintf("%s\n%s", output[:ind], output[ind:])
	}
	fmt.Printf("%v\n", output)
	outputArr := strings.Fields(strings.TrimSpace(output))
	trackSuccess := false
	trackPrg := false

	// We are only checking the overall "PASS" or "FAIL"
	// We are going to look for the first occurance of PASS or FAIL from the end
	// there are also some space-like deceiving characters
	for i := len(outputArr) - 1; i >= 0; i-- {
		if strings.TrimSpace(outputArr[i]) != "" {
			if outputArr[i] == "PASS" {
				trackSuccess = true
				trackPrg = true
			} else if outputArr[i] == "FAIL" {
				trackPrg = true
			}
		}
		if trackPrg {
			break
		}
	}
	return trackSuccess, nil
}
