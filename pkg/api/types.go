package api

import (
	kapi_config "github.com/qlik-oss/k-apis/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// QliksenseConfig is exported
type QliksenseConfig struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec              *ContextSpec `json:"spec" yaml:"spec"`
	QliksenseHomePath string       `json:"-" yaml:"-"`
}

/*type CommonConfig struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}
*/
// QliksenseCR is exported
type QliksenseCR struct {
	kapi_config.KApiCr `json:",inline" yaml:",inline"`
}

// ContextSpec is exported
type ContextSpec struct {
	Contexts       []Context `json:"contexts" yaml:"contexts"`
	CurrentContext string    `json:"currentContext" yaml:"currentContext"`
}

// Context is exported
type Context struct {
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
	CrFile string `json:"crFile,omitempty" yaml:"crFile,omitempty"`
}

// Metadata is exported
type Metadata struct {
	Name   string            `json:"name,omitempty" yaml:"name,omitempty"`
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// ServiceKeyValue holds the combination of service, key and value
type ServiceKeyValue struct {
	SvcName string
	Key     string
	Value   string
}
