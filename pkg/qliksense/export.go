package qliksense

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (q *Qliksense) ExportContext(context string, output string) error {
	qliksenseContextsDir := filepath.Join(q.QliksenseHome, QliksenseContextsDir)
	qliksenseContextFile := filepath.Join(qliksenseContextsDir, context)
	qliksenseSecretsDir := filepath.Join(q.QliksenseHome, QliksenseSecretsDir, QliksenseContextsDir)
	qliksenseSecretsFile := filepath.Join(qliksenseSecretsDir, context)
	// files := []string{qliksenseContextFile, qliksenseSecretsFile}

	fmt.Println(q.QliksenseHome)
	fmt.Println(qliksenseSecretsFile)
	fmt.Println(qliksenseContextFile)

	filename := "result.zip"
	destinationFile, err := os.Create(output + "/" + filename)
	var folders []string
	if err != nil {
		return err
	}
	folders = append(folders, qliksenseContextFile, qliksenseSecretsFile)
	if err := RecursiveZip(folders, destinationFile); err != nil {
		return err
	}

	return nil
}

func RecursiveZip(pathToZip []string, destinationFile *os.File) error {

	myZip := tar.NewWriter(destinationFile)
	for _, element := range pathToZip {
		err := filepath.Walk(element, func(filePath string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if err != nil {
				return err
			}
			relPath := strings.TrimPrefix(filePath, element)
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
	}
	err := myZip.Close()
	if err != nil {
		return err
	}
	return nil
}
