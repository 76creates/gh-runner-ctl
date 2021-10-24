package cmd

import (
	"errors"
	deindent "github.com/76creates/de-indent"
	"github.com/76creates/runner-cli/ghRunnerCtl"
	"github.com/76creates/runner-cli/log"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

func init() {
	runnerTendCmd.Flags().String("conf", "", "location of the runner configuration yaml")
	runnerTendCmd.Flags().String("github-workflow-run-id", "", "workflow run ID to †end to")
	runnerTendCmd.MarkFlagRequired("github-workflow-run-id")

	runnerCmd.AddCommand(runnerTendCmd)
}

var runnerTendCmd = &cobra.Command{
	Use:   "tend [WORKFLOW_ID]",
	Short: deindent.DeIndent(`
		tend watches the workflow run, it looks for the jobs waiting in the queue,
		checks the workflow file to see what runner type is used for a given job,
		and creates the appropriate object defined with the provider extracted from
		the runners configuration file
	`),
	//Args: cobra.MinimumNArgs(1), // workflow run ID
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debug("landed")
		// parse workflow run id to int64
		var workflowRunID int64
		workflowRunID, err := strconv.ParseInt(cmd.Flag("github-workflow-run-id").Value.String(), 10, 64)
		if err != nil {
			log.ErrorF("could not convert %q to int64", cmd.Flag("github-workflow-run-id").Value.String())
			return err
		}

		var runnerConfig *ghRunnerCtl.RunnerConfig
		if cmd.Flag("conf").Value.String() != "" {
			runnerConfigFile, err := os.Open(cmd.Flag("conf").Value.String())
			if err != nil {
				log.Error(err.Error())
				return err
			}
			runnerConfig = ghRunnerCtl.Parse(runnerConfigFile)
		} else {
			// read stdin if not empty
			in, err := os.Stdin.Stat()
			if err != nil {
				return err
			}
			if in.Size() > 0 {
				runnerConfig = ghRunnerCtl.Parse(os.Stdin)
			}
		}

		if runnerConfig == nil {
			return errors.New("could not find config")
		}

		tend := ghRunnerCtl.Tend{}
		return tend.Start(ctx, workflowRunID, runnerConfig)
	},
}
