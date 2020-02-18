package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func checkExists(filename string, isFile bool) os.FileInfo {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		if isFile {
			LogDebugMessage("File does not exist")
		} else {
			LogDebugMessage("Dir does not exist")
		}
		return nil
	}
	LogDebugMessage("File exists")
	return info
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	if fe := checkExists(filename, true); fe != nil && !fe.IsDir() {
		return true
	}
	return false
}

// DirExists checks if a directory exists
func DirExists(dirname string) bool {
	if fe := checkExists(dirname, false); fe != nil && fe.IsDir() {
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

// ReadKeys - reads key file from disk
func ReadKeys(keyFile string) ([]byte, error) {
	keybyteArray, err := ioutil.ReadFile(keyFile)
	if err != nil {
		err = fmt.Errorf("There was an error reading from file: %s, %v", keyFile, err)
		log.Println(err)
		return nil, err
	} else {
		LogDebugMessage("Read key as byte[]: %+v", keybyteArray)
	}
	return keybyteArray, nil
}
