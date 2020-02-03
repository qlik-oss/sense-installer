package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func KubectlApply(manifests string) error {
	tempYaml, err := ioutil.TempFile("", "")
	if err != nil {
		fmt.Println("cannot create file ", err)
		return err
	}
	tempYaml.WriteString(manifests)

	cmd := exec.Command("kubectl", "apply", "-f", tempYaml.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("kubectl apply failed with %s\n", err)
		return err
	}
	os.Remove(tempYaml.Name())
	return nil
}
