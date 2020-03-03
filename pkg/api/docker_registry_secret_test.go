package api

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDockerConfigJsonSecret(t *testing.T) {
	dockerConfigJsonSecret := DockerConfigJsonSecret{
		Name:      "some-name",
		Uri:       "some-uri",
		Username:  "some-username",
		Password:  "some-password",
		Email:     "some-email",
	}
	dockerConfigJsonSecretFromYaml := DockerConfigJsonSecret{}
	validYamlMap := map[string]interface{}{}

	privateKey, err := rsa.GenerateKey(rand.Reader, RSA_KEY_LENGTH)
	if err != nil {
		t.Fatalf("error generating RSA private key: %v\n", err)
	}

	dockerConfigJsonSecretYamlBytes, err := dockerConfigJsonSecret.ToYaml(&privateKey.PublicKey)
	dockerConfigJsonMap := map[string]interface{}{}
	if err != nil {
		t.Fatalf("error converting secret to yaml: %v", err)
	} else if err := yaml.Unmarshal(dockerConfigJsonSecretYamlBytes, &validYamlMap); err != nil {
		t.Fatalf("error unmarshalling yaml string: %v, error: %v", string(dockerConfigJsonSecretYamlBytes), err)
	} else if validYamlMap["apiVersion"] != "v1" ||
		validYamlMap["kind"] != "Secret" ||
		validYamlMap["metadata"].(map[string]interface{})["name"] != dockerConfigJsonSecret.Name ||
		validYamlMap["type"] != "kubernetes.io/dockerconfigjson" {
		t.Fatalf("error verifying validity of secret yaml: %v", string(dockerConfigJsonSecretYamlBytes))
	} else if dockerConfigJsonBytesBase64, ok := validYamlMap["data"].(map[string]interface{})[".dockerconfigjson"]; !ok {
		t.Fatalf("no .dockerconfigjson data key in the secret yaml: %v", string(dockerConfigJsonSecretYamlBytes))
	} else if dockerConfigJsonEncryptedBytes, err := base64.StdEncoding.DecodeString(dockerConfigJsonBytesBase64.(string)); err != nil {
		t.Fatalf("error decoding dockerConfigJsonBytes from base64: %v", err)
	} else if dockerConfigJsonBytes, err := Decrypt(dockerConfigJsonEncryptedBytes, privateKey); err != nil {
		t.Fatalf("error decrypting dockerConfigJsonBytes: %v", err)
	} else if err := json.Unmarshal(dockerConfigJsonBytes, &dockerConfigJsonMap); err != nil {
		t.Fatalf("error unmarshalling dockerConfigJson from json: %v", err)
	} else if dockerConfigJson, ok := dockerConfigJsonMap["auths"].(map[string]interface{})[dockerConfigJsonSecret.Uri]; !ok {
		t.Fatalf("dockerConfigJson map does not contain data for the registry: %v", dockerConfigJsonSecret.Uri)
	} else if dockerConfigJson.(map[string]interface{})["username"] != dockerConfigJsonSecret.Username ||
		dockerConfigJson.(map[string]interface{})["password"] != dockerConfigJsonSecret.Password ||
		dockerConfigJson.(map[string]interface{})["email"] != dockerConfigJsonSecret.Email {
		t.Fatal("dockerConfigJson map does not contain expected values")
	} else {
		authBase64 := dockerConfigJson.(map[string]interface{})["auth"]
		if auth, err := base64.StdEncoding.DecodeString(authBase64.(string)); err != nil {
			t.Fatal("error base64 decoding auth value")
		} else if string(auth) != fmt.Sprintf("%s:%s", dockerConfigJsonSecret.Username, dockerConfigJsonSecret.Password) {
			t.Fatal("auth value was not what we expected")
		}
	}

	t.Logf("dockerConfigJsonSecretYaml: \n%v\n", string(dockerConfigJsonSecretYamlBytes))
	if err := dockerConfigJsonSecretFromYaml.FromYaml(dockerConfigJsonSecretYamlBytes, privateKey); err != nil {
		t.Fatalf("error reading secret in from yaml: %v", err)
	} else if !reflect.DeepEqual(dockerConfigJsonSecret, dockerConfigJsonSecretFromYaml) {
		t.Fatalf("secret: %v does not equal secret: %v", dockerConfigJsonSecret, dockerConfigJsonSecretFromYaml)
	}
}
