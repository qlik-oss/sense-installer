package api

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func checkExists(filename string) os.FileInfo {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil
	}
	LogDebugMessage("File exists")
	return info
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	if fe := checkExists(filename); fe != nil && !fe.IsDir() {
		return true
	}
	return false
}

// DirExists checks if a directory exists
func DirExists(dirname string) bool {
	if fe := checkExists(dirname); fe != nil && fe.IsDir() {
		return true
	}
	return false
}

// LogDebugMessage logs a debug message
func LogDebugMessage(strMessage string, args ...interface{}) {
	if os.Getenv("QLIKSENSE_DEBUG") == "true" {
		log.Printf(strMessage, args...)
	}
}

// ReadKeys reads key file from disk
func ReadKeys(keyFile string) ([]byte, error) {
	keybyteArray, err := ioutil.ReadFile(keyFile)
	if err != nil {
		err = fmt.Errorf("There was an error reading from file: %s, %v", keyFile, err)
		log.Println(err)
		return nil, err
	}
	return keybyteArray, nil
}

// ProcessConfigArgs processes args and returns an service, key, value slice
func ProcessConfigArgs(args []string) ([]*ServiceKeyValue, error) {
	// prepare received args
	// split args[0] into key and value
	if len(args) == 0 {
		err := fmt.Errorf("No args were provided. Please provide args to configure the current context")
		return nil, err
	}
	resultSvcKV := make([]*ServiceKeyValue, len(args))
	re1 := regexp.MustCompile(`(\w{1,}).(\w{1,})=("*[\w\-?=_/:0-9\.]+"*)`)
	for i, arg := range args {
		LogDebugMessage("Arg received: %s", arg)
		result := re1.FindStringSubmatch(arg)
		// check if result array's length is == 4 (index 0 - is the full match & indices 1,2,3- are the fields we need)
		if len(result) != 4 {
			err := fmt.Errorf("Please provide valid args for this command")
			return nil, err
		}
		resultSvcKV[i] = &ServiceKeyValue{
			SvcName: result[1],
			Key:     result[2],
			Value:   strings.ReplaceAll(result[3], `"`, ""),
		}
	}
	return resultSvcKV, nil
}

func ExecuteTaskWithBlinkingStdoutFeedback(task func() (interface{}, error), feedback string) (result interface{}, err error) {
	taskDone := make(chan bool)
	go func() {
		result, err = task()
		taskDone <- true
	}()
	progressOnTicker := time.NewTicker(500 * time.Millisecond)
	progressOffTicker := time.NewTicker(1000 * time.Millisecond)
	printProgress := func(on bool) {
		if on {
			fmt.Printf("%s\r", feedback)
		} else {
			fmt.Printf("%s\r", strings.Repeat(" ", len(feedback)))
		}
	}
	for {
		select {
		case <-taskDone:
			progressOnTicker.Stop()
			progressOffTicker.Stop()
			printProgress(false)
			return result, err
		case <-progressOnTicker.C:
			printProgress(true)
		case <-progressOffTicker.C:
			printProgress(false)
		}
	}
}

func DownloadFile(url, baseFolder, installerName string) error {
	var (
		out  *os.File
		err  error
		resp *http.Response
	)
	// Create the file
	fileName := filepath.Join(baseFolder, installerName)
	LogDebugMessage("Installer Filename: %s\n", fileName)
	if out, err = os.Create(fileName); err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	if resp, err = http.Get(url); err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unable to download the file from URL: %s, status: %s", url, resp.Status)
		log.Println(err)
		return err
	}

	// Write the body to file
	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}
	err = os.Chmod(fileName, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func ExplodePackage(destination, fileToUntar string) error {
	LogDebugMessage("Destination: %s\n", destination)
	LogDebugMessage("fileToUntar: %s\n", fileToUntar)

	if strings.HasSuffix(fileToUntar, "zip") {
		LogDebugMessage("This is a windows file : %s", fileToUntar)
		err := UnZipFile(destination, fileToUntar)
		if err != nil {
			return nil
		}
	} else if strings.HasSuffix(fileToUntar, "tar.gz") {
		LogDebugMessage("This is a mac/linux file: %s", fileToUntar)
		err := UntarGzFile(destination, fileToUntar)
		if err != nil {
			return nil
		}
	}
	return nil
}

func UntarGzFile(destination, fileToUntar string) error {
	lFile, err := os.Open(fileToUntar)
	if err != nil {
		err = errors.Wrapf(err, "unable to read the local file %s", fileToUntar)
		log.Fatal(err)
		return err
	}

	gzReader, err := gzip.NewReader(lFile)
	if err != nil {
		err = errors.Wrap(err, "unable to load the file into a gz reader")
		log.Fatal(err)
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			err = errors.Wrap(err, "error during untar")
			log.Fatal(err)
			return err
		case header == nil:
			continue
		}

		fileInLoop := filepath.Join(destination, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(fileInLoop); err != nil {
				if err := os.MkdirAll(fileInLoop, 0755); err != nil {
					err = errors.Wrapf(err, "error creating directory %s", fileInLoop)
					log.Fatal(err)
					return err
				}
			}
		case tar.TypeReg:
			fileAtLoc, err := os.OpenFile(fileInLoop, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				err = errors.Wrapf(err, "error opening file %s", fileInLoop)
				log.Fatal(err)
				return err
			}

			if _, err := io.Copy(fileAtLoc, tarReader); err != nil {
				err = errors.Wrapf(err, "error writing file %s", fileInLoop)
				log.Fatal(err)
				return err
			}
			fileAtLoc.Close()
			fileAtLoc.Chmod(os.ModePerm)
		}
	}
}

func UnZipFile(destination, fileToUnzip string) error {
	zipReader, _ := zip.OpenReader(fileToUnzip)
	for _, file := range zipReader.Reader.File {

		zippedFile, err := file.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer zippedFile.Close()
		extractedFilePath := filepath.Join(
			destination,
			file.Name,
		)
		outputFile, err := os.OpenFile(
			extractedFilePath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			file.Mode(),
		)
		if err != nil {
			log.Fatal(err)
		}
		defer outputFile.Close()

		_, err = io.Copy(outputFile, zippedFile)
		if err != nil {
			log.Fatal(err)
		}
		LogDebugMessage("File extracted: %s, Extracted file path: %s\n", file.Name, extractedFilePath)
	}
	return nil
}
