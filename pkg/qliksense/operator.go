package qliksense

import (
	"fmt"
	_ "gopkg.in/yaml.v2"
	"strings"
)

func (q *Qliksense) ViewOperatorCrd() {
	for _, v := range q.getFileList("crd") {
		q.printYamlFile(v)
	}
	for _, v := range q.getFileList("crd-deploy") {
		q.printYamlFile(v)
	}
}

func (q *Qliksense) printYamlFile(packrFile string) {
	s, err := q.CrdBox.FindString(packrFile)
	if err != nil {
		fmt.Printf("Cannot read file %s", packrFile)
	}
	fmt.Println(s)
	fmt.Println("---")

}
func (q *Qliksense) getFileList(resourceType string) []string {
	var resList []string
	for _, v := range q.CrdBox.List() {
		if strings.Contains(v, resourceType+"/") {
			resList = append(resList, []string{v}...)
		}
	}
	return resList
}
