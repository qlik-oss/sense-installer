package qliksense

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/markbates/pkger"
)

func init() {
	pkger.Include("/pkg/qliksense/operator-yaml")
}

func (q *Qliksense) ViewOperator() error {
	if operatorCRDString, err := q.GetOperatorCRDString(); err != nil {
		return err
	} else if _, err := io.WriteString(os.Stdout, operatorCRDString); err != nil {
		return err
	}
	return nil
}

func (q *Qliksense) ViewOperatorController() error {
	if operatorControllerString, err := q.GetOperatorControllerString(); err != nil {
		return err
	} else if _, err := io.WriteString(os.Stdout, operatorControllerString); err != nil {
		return err
	}
	return nil
}

func (q *Qliksense) GetOperatorCRDString() (string, error) {
	return getYamlFromPkgerDir("/pkg/qliksense/operator-yaml/crds")
}

func (q *Qliksense) GetOperatorControllerString() (string, error) {
	return getYamlFromPkgerDir("/pkg/qliksense/operator-yaml/deploy")
}

func getYamlFromPkgerDir(dir string) (string, error) {
	result := ""
	pkgingFile, err := pkger.Open(dir)
	if err != nil {
		return "", err
	}
	defer pkgingFile.Close()
	if fileInfos, err := pkgingFile.Readdir(-1); err != nil {
		return "", err
	} else {
		for _, fileInfo := range fileInfos {
			if yaml, err := getYamlFromPkgerFile(path.Join(pkgingFile.Path().Name, fileInfo.Name())); err != nil {
				return "", err
			} else {
				result = result + yaml
			}
		}
	}
	return result, nil
}

func getYamlFromPkgerFile(filePath string) (string, error) {
	f, err := pkger.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if fBytes, err := ioutil.ReadAll(f); err != nil {
		return "", err
	} else {
		return fmt.Sprintln("#source: " + path.Base(filePath) + "\n\n" + string(fBytes) + "\n---"), nil
	}
}
