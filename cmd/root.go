package cmd

import (
	"context"
	deindent "github.com/76creates/de-indent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"strings"
)

var (
	version = "0"
	ctx     = context.Background()
)

func Init() {
	rootCmd.PersistentFlags().Bool("github.actions", false, "tell script that you are running inside a runner")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "show debug messages")

	viper.BindPFlag("github.actions", rootCmd.PersistentFlags().Lookup("github.actions"))
	viper.BindEnv("github.actions", strings.ReplaceAll("github.actions", ".", "_"), "GITHUB_ACTIONS")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gh-runner-ctl",
	Short: "cmd tool for spawning GitHub runners",
	Long: deindent.DeIndent(`
	GitHub runner ctl is used to create runners on the demand
	`),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}
