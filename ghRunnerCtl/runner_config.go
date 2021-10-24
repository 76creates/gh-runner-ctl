package ghRunnerCtl

import (
	"fmt"
	"github.com/76creates/runner-cli/provider"
	"github.com/76creates/runner-cli/provider/gcp"
	"github.com/76creates/runner-cli/provider/scaleway"
	"gopkg.in/yaml.v2"
	"io"
	"strings"
)


// TODO: this should be a file where there is a list of possible "access" objects, and we should inject them if not overriden per type itself

// TODO: dont use viper get, use these objects to run "CREATE" or "DELETE" and such

// RunnerProvider holds the runner configuration per provider, configs should be declared in the providers namespace
type RunnerProvider struct {
	Scaleway *scaleway.RunnerConfig `mapstructure:"scaleway,omitempty" yaml:"scaleway"`
	GCP *gcp.RunnerConfig `mapstructure:"gcp,omitempty" yaml:"gcp"`
}

type RunnerType struct {
	// Provider name of the provider declared in the providers object
	Provider string `mapstructure:"provider" yaml:"provider"`

	provider provider.Provider
}

type ConfigYaml struct{
	Types map[string]RunnerType `mapstructure:"runners" yaml:"runners"`
	Providers map[string]RunnerProvider `mapstructure:"providers" yaml:"providers"`
}

type RunnerConfig struct {
	Runners map[string]*RunnerType
}

func (rt RunnerType)GetProvider() (p provider.Provider) {
	p = rt.provider
	return p
}

// Parse the yaml runner config file into the object, panics if config cannot be decoded
func Parse(file io.Reader) *RunnerConfig {
	c := new(RunnerConfig)
	c.Runners = make(map[string]*RunnerType)

	runnerConf := new(ConfigYaml)
	err := yaml.NewDecoder(file).Decode(&runnerConf)
	if err != nil {
		panic(err)
	}

	mp := getProviderMap(runnerConf.Providers)

	for k, v := range runnerConf.Types {
		if _, ok := mp[v.Provider]; !ok {
			panic(fmt.Sprintf("could not find the %q provider in the providers object", v.Provider))
		}

		c.Runners[k] = &RunnerType{provider: mp[v.Provider]}
	}

	return c
}

// ParseString the yaml runner config string into the object, panics if config cannot be decoded
func ParseString(conf string) *RunnerConfig {
	return Parse(strings.NewReader(conf))
}

// getProvider attempts to extract the provider from the RunnerProvider, on failure it panics
// if it finds two providers under same config it will panic
func getProviderMap(in map[string]RunnerProvider) map[string]provider.Provider {
	if len(in) == 0 {
		panic("no provider defined")
	}

	mp := make(map[string]provider.Provider)

	for providerName, providerMap := range in {
		// this needs to be repeated for every provider that will be added in the future as is atm
		if providerMap.Scaleway != nil {
			if _, ok := mp[providerName]; ok {
				panic(fmt.Sprintf("more than one provider for name %s", providerName))
			}
			mp[providerName] = providerMap.Scaleway
		}
		if providerMap.GCP != nil {
			if _, ok := mp[providerName]; ok {
				panic(fmt.Sprintf("more than one provider for name %s", providerName))
			}
			mp[providerName] = providerMap.GCP
		}
		// panic if no provider has been matched
		if _, ok := mp[providerName]; !ok {
			panic(fmt.Sprintf("no provider matched for name %s", providerName))
		}
	}

	return mp
}


