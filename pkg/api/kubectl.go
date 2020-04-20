package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// KubectlApply create resoruces in the provided namespace,
// if namespace="" then use whatever the kubectl default is
func KubectlApply(manifests, namespace string) error {
	return kubectlOperation(manifests, "apply", namespace)
}

func KubectlApplyVerbose(manifests, namespace string, verbose bool) error {
	return kubectlOperationVerbose(manifests, "apply", namespace, verbose)
}

// KubectlDelete delete resoruces in the provided namespace,
// if namespace="" then use whatever the kubectl default is
func KubectlDelete(manifests, namespace string) error {
	return kubectlOperation(manifests, "delete", namespace)
}

func KubectlDeleteVerbose(manifests, namespace string, verbose bool) error {
	return kubectlOperationVerbose(manifests, "delete", namespace, verbose)
}

func GetKubectlNamespace() string {
	namespace := ""
	cmd := exec.Command("kubectl", "config", "current-context")
	var out, out2 bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("kubectl config current-context %q\n", err)
		return namespace
	}
	if out.String() == "" {
		fmt.Println("kubectl config current-context does not return anything")
		return namespace
	}

	cmd = exec.Command("kubectl", "config", "view", "-o", `jsonpath={.contexts[?(@.name == "`+strings.TrimSpace(out.String())+`")].context.namespace}`)
	cmd.Stdout = &out2
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("kubectl config view failed with %q\n", err)
		return namespace
	}
	namespace = out2.String()
	return namespace
}

func SetKubectlNamespace(ns string) {
	cmd := exec.Command("kubectl", "config", "set-context", "--namespace="+ns, "--current")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("kubectl config set-context --namespace failed with %q\n", err)
	}
}

func kubectlOperation(manifests string, oprName string, namespace string) error {
	return kubectlOperationVerbose(manifests, oprName, namespace, true)
}

func kubectlOperationVerbose(manifests string, oprName string, namespace string, verbose bool) error {
	tempYaml, err := ioutil.TempFile("", "")
	if err != nil {
		fmt.Println("cannot create file ", err)
		return err
	}
	tempYaml.WriteString(manifests)

	arguments := make([]string, 0)
	arguments = append(arguments, oprName)
	arguments = append(arguments, "-f")
	arguments = append(arguments, tempYaml.Name())

	if oprName == "apply" {
		arguments = append(arguments, "--validate=false")
	}
	if namespace != "" {
		arguments = append(arguments, "-n")
		arguments = append(arguments, namespace)
	}
	var cmd *exec.Cmd
	if oprName == "apply" {
		cmd = exec.Command("kubectl", arguments...)
	} else {
		cmd = exec.Command("kubectl", arguments...)
	}

	sterrBuffer := &bytes.Buffer{}
	stoutBuffer := &bytes.Buffer{}
	cmd.Stdout = stoutBuffer
	cmd.Stderr = sterrBuffer
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("kubectl %v failed with: %v, %v, temp k8s yaml file:%v\n", oprName, err, sterrBuffer.String(), tempYaml.Name())
	}
	if verbose {
		fmt.Println(stoutBuffer.String())
	}
	os.Remove(tempYaml.Name())
	return nil
}

func KubectlDirectOps(opr []string, namespace string) (string, error) {
	arguments := []string{}
	if namespace != "" {
		arguments = append(arguments, "-n", namespace)
	}
	arguments = append(arguments, opr...)
	var out bytes.Buffer
	cmd := exec.Command("kubectl", arguments...)
	LogDebugMessage("Kubectl command: %s %v\n", "kubectl", arguments)
	sterrBuffer := &bytes.Buffer{}
	cmd.Stderr = sterrBuffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("kubectl %v failed with: %v, %v\n", opr, err, sterrBuffer.String())
	}
	s := out.String()
	return s, nil
}
