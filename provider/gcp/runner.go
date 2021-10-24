package gcp

import "github.com/76creates/runner-cli/provider"

// RunnerConfig configuration for the runner runner creation, it contains access/credentials for the
// given provider thus you can use multiple keys/accounts for multiple different runner types
type RunnerConfig struct {
	provider.BaseProvider

	Access *AccessConfig `mapstructure:"access" yaml:"access"`

	Zone *string `mapstructure:"zone" yaml:"zone"`
	Project *string `mapstructure:"project" yaml:"project"`
	MachineType *string `mapstructure:"machine-type" yaml:"machine-type"`
	NetworkName *string `mapstructure:"network-name" yaml:"network-name"`
	Image *string `mapstructure:"image" yaml:"image"`
	CloudInit *string `mapstructure:"cloud-init" yaml:"cloud-init"`
}

// AccessConfig if needed provides access info for the provider to authentification, etc.
// if "access method" is not needed for the provider this struct should be left empty
type AccessConfig struct {
	JSON *string `mapstructure:"json-key" yaml:"json-key"`
}