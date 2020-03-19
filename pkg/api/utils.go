package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
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
	re1 := regexp.MustCompile(`(\w{1,}).(\w{1,})=("*[\w\-?=_/:0-9]+"*)`)
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

// ProcessUnsetConfigArgs processes args and returns an service, key, nil slice
func ProcessUnsetConfigArgs(args []string) ([]*ServiceKeyValue, error) {
	if len(args) == 0 {
		err := fmt.Errorf("No args were provided. Please provide args to configure the current context")
		return nil, err
	}
	resultSvcKV := make([]*ServiceKeyValue, len(args))
	re1 := regexp.MustCompile(`(\w{1,}).(\w{1,})`)
	for i, arg := range args {
		LogDebugMessage("Arg received: %s", arg)
		result := re1.FindStringSubmatch(arg)
		// check if result array's length is == 3 (index 0 - is the full match & indices 1,2,- are the fields we need)
		if len(result) != 3 {
			err := fmt.Errorf("Please provide valid args for this command")
			return nil, err
		}
		resultSvcKV[i] = &ServiceKeyValue{
			SvcName: result[1],
			Key:     result[2],
			Value:   "",
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
