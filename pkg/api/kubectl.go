package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func KubectlApply(manifests, namespace string) error {
	return kubectlOperation(manifests, "apply", namespace)
}

func KubectlDelete(manifests, namespace string) error {
	return kubectlOperation(manifests, "delete", namespace)
}

func kubectlOperation(manifests string, oprName string, namespace string) error {
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

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("kubectl apply failed with %s\n", err)
		fmt.Println("temp CRD file: " + tempYaml.Name())
		return err
	}
	os.Remove(tempYaml.Name())
	return nil
}
