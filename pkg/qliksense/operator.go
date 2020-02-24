package qliksense

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (q *Qliksense) ViewOperator() error {
	io.WriteString(os.Stdout, q.GetOperatorCRDString())
	return nil
}

// this will return crd,deployment,role, rolebinding,serviceaccount for operator
func (q *Qliksense) GetOperatorCRDString() string {
	result := ""
	for _, v := range q.getFileList("crd") {
		result = q.getYamlFromPackrFile(v)
	}

	return result
}

func (q *Qliksense) GetOperatorControllerString() string {
	result := ""
	for _, v := range q.getFileList("crd-deploy") {
		result = result + q.getYamlFromPackrFile(v)
	}
	return result
}

func (q *Qliksense) getYamlFromPackrFile(packrFile string) string {
	s, err := q.CrdBox.FindString(packrFile)
	if err != nil {
		fmt.Printf("Cannot read file %s", packrFile)
	}
	return fmt.Sprintln("#soruce: " + packrFile + "\n\n" + s + "\n---")
}
func (q *Qliksense) getFileList(resourceType string) []string {
	var resList []string
	for _, v := range q.CrdBox.List() {
		if strings.Contains(v, filepath.Join(resourceType, "")) {
			resList = append(resList, []string{v}...)
		}
	}
	return resList
}
