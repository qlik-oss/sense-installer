package qliksense

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (q *Qliksense) ExportContext(context string) error {
	qliksenseContextsDir := filepath.Join(q.QliksenseHome, QliksenseContextsDir)
	qliksenseContextFile := filepath.Join(qliksenseContextsDir, context, context+".yaml")
	qliksenseSecretsDir := filepath.Join(q.QliksenseHome, QliksenseSecretsDir, QliksenseContextsDir)
	qliksenseSecretsFile := filepath.Join(qliksenseSecretsDir, context)
	// files := []string{qliksenseContextFile, qliksenseSecretsFile}

	fmt.Println(q.QliksenseHome)
	fmt.Println(qliksenseSecretsFile)
	fmt.Println(qliksenseContextFile)

	if err := RecursiveZip("result.zip", qliksenseSecretsFile, q.QliksenseHome); err != nil {
		return err
	}
	if err := RecursiveZip("result.zip", qliksenseContextFile, q.QliksenseHome); err != nil {
		return err
	}
	return nil
}

func RecursiveZip(filename, pathToZip, destinationPath string) error {
	destinationFile, err := os.Create(destinationPath + "/" + filename)
	if err != nil {
		return err
	}
	myZip := zip.NewWriter(destinationFile)
	err = filepath.Walk(pathToZip, func(filePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(filePath, pathToZip)
		zipFile, err := myZip.Create(relPath)
		if err != nil {
			return err
		}
		fsFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, fsFile)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = myZip.Close()
	if err != nil {
		return err
	}
	return nil
}
