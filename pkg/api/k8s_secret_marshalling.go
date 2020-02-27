package api

import (
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func K8sSecretToYaml(k8sSecret v1.Secret) ([]byte, error) {
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

func K8sSecretFromYaml(k8sSecretBytes []byte) (v1.Secret, error) {
	k8sSecret := v1.Secret{}
	if err := yaml.UnmarshalStrict(k8sSecretBytes, &k8sSecret); err != nil {
		return k8sSecret, err
	}
	return k8sSecret, nil
}
