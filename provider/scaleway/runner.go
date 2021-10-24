package scaleway

import "github.com/76creates/runner-cli/provider"

// RunnerConfig configuration for the runner runner creation, it contains access/credentials for the
// given provider thus you can use multiple keys/accounts for multiple different runner types
type RunnerConfig struct {
	provider.BaseProvider

	Access *AccessConfig `mapstructure:"access" yaml:"access"`

	Image *string `mapstructure:"image" yaml:"image"`
	InstanceType *string `mapstructure:"instance-type" yaml:"instance-type"`
	Zone *string `mapstructure:"zone" yaml:"zone"`
	SecurityGroup *string `mapstructure:"security-group" yaml:"security-group"`
	Tags *[]string `mapstructure:"tags" yaml:"tags"`
	CloudInit *string `mapstructure:"cloud-init" yaml:"cloud-init"`
}

// AccessConfig if needed provides access info for the provider to authentification, etc.
// if "access method" is not needed for the provider this struct should be left empty
type AccessConfig struct {
	KeyID *string `mapstructure:"key_id" yaml:"key_id"`
	KeySecret *string `mapstructure:"key_secret" yaml:"key_secret"`
	ProjectID *string `mapstructure:"project" yaml:"project"`
	OrgID *string `mapstructure:"organisation" yaml:"organisation"`
}