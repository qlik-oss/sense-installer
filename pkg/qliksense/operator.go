package qliksense

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/markbates/pkger"
)

func (q *Qliksense) ViewOperator() error {
	io.WriteString(os.Stdout, q.GetOperatorCRDString())
	return nil
}

func (q *Qliksense) ViewOperatorController() error {
	io.WriteString(os.Stdout, q.GetOperatorControllerString())
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
	fmt.Println(packrFile)
	s, err := pkger.Info(packrFile)
	fmt.Println(s.Name)
	if err != nil {
		fmt.Printf("Cannot read file %s", packrFile)
	}
	return fmt.Sprintln("#soruce: " + packrFile + "\n\n" + s.Name + "\n---")
}

func (q *Qliksense) getFileList(resourceType string) []string {
	var resList []string
	var keys []string
	pkger.Walk(q.CrdPkger, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			keys = append(keys, path)
		}
		return nil
	})
	sort.Strings(keys)
	for _, v := range keys {
		if strings.Contains(v, filepath.Join(resourceType, "")) {
			resList = append(resList, []string{v}...)
		}
	}
	return resList
}
