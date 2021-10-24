package cmd

import (
	"github.com/spf13/viper"
	"strings"
)

// initFlags do initialization steps for flag/s if needed here
func initFlags() {
	// split github repo string
	s := strings.SplitN(viper.GetString("github.repo.path"), "/", 2)
	viper.Set("github.repo.owner", s[0])
	viper.Set("github.repo.name", s[1])
}
