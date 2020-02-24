package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	secretAPIVersion = "v1"
	secretKind       = "Secret"
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

// ReadKeys reads key file from disk
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

// ProcessConfigArgs processes args and returns an service, key, value slice
func ProcessConfigArgs(args []string) ([]*ServiceKeyValue, error) {
	// prepare received args
	// split args[0] into key and value
	if len(args) == 0 {
		err := fmt.Errorf("No args were provided. Please provide args to configure the current context")
		log.Println(err)
		return nil, err
	}
	resultSvcKV := make([]*ServiceKeyValue, len(args))
	re1 := regexp.MustCompile(`(\w{1,})\[name=(\w{1,})\]=("*[\w\-_/:0-9]+"*)`)
	for i, arg := range args {
		LogDebugMessage("Arg received: %s", arg)
		result := re1.FindStringSubmatch(arg)
		// check if result array's length is == 4 (index 0 - is the full match & indices 1,2,3- are the fields we need)
		if len(result) != 4 {
			err := fmt.Errorf("Please provide valid args for this command")
			log.Println(err)
			return nil, err
		}
		resultSvcKV[i] = &ServiceKeyValue{
			SvcName: result[1],
			Key:     result[2],
			Value:   result[3],
		}
	}
	return resultSvcKV, nil
}

// GenerateUUID generates a random number
func GenerateUUID() string {
	id := uuid.New()
	fmt.Println(id.String())
	return id.String()
}

// ConstructK8sSecretStructure constructs a K8s Secret struct
func ConstructK8sSecretStructure(secretName, namespace string, dataMap map[string][]byte) v1.Secret {
	secret := v1.Secret{
		TypeMeta: metaV1.TypeMeta{
			APIVersion: secretAPIVersion,
			Kind:       secretKind,
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Type: v1.SecretTypeOpaque,
		Data: dataMap,
	}
	return secret
}
