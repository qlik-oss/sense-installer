package api

import "github.com/qlik-oss/k-apis/config"

// CommonConfig is exported
type CommonConfig struct {
	ApiVersion string    `json:"apiVersion" yaml:"apiVersion"`
	Kind       string    `json:"kind" yaml:"kind"`
	Metadata   *Metadata `json:"metadata" yaml:"metadata"`
}

// QliksenseConfig is exported
type QliksenseConfig struct {
	CommonConfig `json:",inline" yaml:",inline"`
	Spec         *ContextSpec `json:"spec" yaml:"spec"`
}

// QliksenseCR is exported
type QliksenseCR struct {
	CommonConfig `json:",inline" yaml:",inline"`
	Spec         *config.CRSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
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
