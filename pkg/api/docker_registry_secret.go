package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type k8sDockerConfigJsonMapType struct {
	Auths map[string]k8sDockerConfigJsonType `json:"auths"`
}

type k8sDockerConfigJsonType struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
	Auth     string `json:"auth"`
}

func (kdcjt *k8sDockerConfigJsonType) GenerateAuth() {
	kdcjt.Auth = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", kdcjt.Username, kdcjt.Password)))
}

type DockerConfigJsonSecret struct {
	Name      string
	Namespace string
	Uri       string
	Username  string
	Password  string
	Email     string
}

func (d *DockerConfigJsonSecret) ToYaml() ([]byte, error) {
	k8sDockerConfigJson := k8sDockerConfigJsonType{
		Username: d.Username,
		Password: d.Password,
		Email:    d.Email,
	}
	k8sDockerConfigJson.GenerateAuth()
	k8sDockerConfigJsonMap := k8sDockerConfigJsonMapType{
		Auths: map[string]k8sDockerConfigJsonType{
			d.Uri: k8sDockerConfigJson,
		},
	}
	k8sDockerConfigJsonMapBytes, err := json.Marshal(k8sDockerConfigJsonMap)
	if err != nil {
		return nil, err
	}

	k8sSecret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.Name,
			Namespace: d.Namespace,
		},
		Type: v1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": k8sDockerConfigJsonMapBytes,
		},
	}

	//need this to remove the unnecessary metadata.creationTimestamp field:
	k8sSecretYamlMap := map[string]interface{}{}
	if k8sSecretYamlBytes, err := yaml.Marshal(k8sSecret); err != nil {
		return nil, err
	} else if err := yaml.Unmarshal(k8sSecretYamlBytes, &k8sSecretYamlMap); err != nil {
		return nil, err
	} else {
		delete(k8sSecretYamlMap["metadata"].(map[string]interface{}), "creationTimestamp")
		return yaml.Marshal(k8sSecretYamlMap)
	}
}

func (d *DockerConfigJsonSecret) FromYaml(secretBytes []byte) error {
	k8sSecret := v1.Secret{}
	k8sDockerConfigJsonMap := k8sDockerConfigJsonMapType{}
	if err := yaml.UnmarshalStrict(secretBytes, &k8sSecret); err != nil {
		return err
	} else if k8sSecret.TypeMeta.Kind != "Secret" {
		return errors.New("not a Secret kind")
	} else if k8sSecret.Type != v1.SecretTypeDockerConfigJson {
		return errors.New("not a kubernetes.io/dockerconfigjson type")
	} else if k8sDockerConfigJsonMapBytes, ok := k8sSecret.Data[".dockerconfigjson"]; !ok {
		return errors.New("secret data is missing a value for the .dockerconfigjson key")
	} else if err := json.Unmarshal(k8sDockerConfigJsonMapBytes, &k8sDockerConfigJsonMap); err != nil {
		return err
	} else {
		d.Name = k8sSecret.ObjectMeta.Name
		d.Namespace = k8sSecret.ObjectMeta.Namespace
		for registry, k8sDockerConfigJson := range k8sDockerConfigJsonMap.Auths {
			d.Uri = registry
			d.Username = k8sDockerConfigJson.Username
			d.Password = k8sDockerConfigJson.Password
			d.Email = k8sDockerConfigJson.Email
			break
		}
		return nil
	}
}
