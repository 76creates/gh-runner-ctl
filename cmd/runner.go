package cmd

import (
	"context"
	"github.com/spf13/cobra"
)

func init() {
	runnerCmd.PersistentFlags().String("github-token", "", "github token, must have repo scope")
	runnerCmd.MarkPersistentFlagRequired("github-token")
	// TODO: check if binary is in repo and extract automatically
	runnerCmd.PersistentFlags().String("github-repo-owner", "", "github repo owner string")
	runnerCmd.MarkPersistentFlagRequired("github-repo-owner")
	runnerCmd.PersistentFlags().String("github-repo-name", "", "github repo name string")
	runnerCmd.MarkPersistentFlagRequired("github-repo-name")

	rootCmd.AddCommand(runnerCmd)
}

var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "subcommand for runner actions",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ctx = context.WithValue(ctx, "github-token", cmd.Flag("github-token").Value.String())
		ctx = context.WithValue(ctx, "github-repo-owner", cmd.Flag("github-repo-owner").Value.String())
		ctx = context.WithValue(ctx, "github-repo-name", cmd.Flag("github-repo-name").Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}
