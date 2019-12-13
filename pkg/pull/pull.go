package mixin

import (
	"get.porter.sh/porter/pkg/context"
)

// MixinProvider handles searching, listing and communicating with the mixins.
type MixinProvider interface {
	List() ([]Metadata, error)
	GetSchema(Metadata) (string, error)

	// GetVersion is the obsolete form of retrieving mixin version, e.g. exec version, which returned an unstructured
	// version string. It will be deprecated soon and is replaced by GetVersionMetadata.
	GetVersion(Metadata) (string, error)

	// GetVersionMetadata is the new form of retrieving mixin version, e.g. exec version --output json, which returns
	// a structured version string. It replaces GetVersion.
	GetVersionMetadata(Metadata) (*VersionInfo, error)
	Install(InstallOptions) (*Metadata, error)
	Uninstall(UninstallOptions) (*Metadata, error)

	// Run a command against the specified mixin
	Run(mixinContext *context.Context, mixinName string, commandOpts CommandOptions) error
}

type CommandOptions struct {
	Runtime bool
	Command string
	Input   string
	File    string
}