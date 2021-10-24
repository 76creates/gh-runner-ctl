package ghRunnerCtl

import (
	"context"
	"fmt"
	"github.com/76creates/runner-cli/ghCtl"
	"github.com/76creates/runner-cli/log"
	"github.com/google/go-github/v39/github"
	"time"
)

type Tend struct {
	ctx context.Context
	workflowRunID int64
}

// Start detects jobs that are in need of handling and spawns a dedicated coroutine
func (t *Tend)Start(ctx context.Context, workflowRunID int64, runnerConfig *RunnerConfig) error {
	t.ctx = ctx
	t.workflowRunID = workflowRunID

	jobs := make(map[string]*tendJob)
	// init gh client
	t.ctx = context.WithValue(ctx, "client", ghCtl.InitClient(t.ctx))

	// validate that the workflow exists
	workflowRun, err := ghCtl.GetWorkflowRunWithTheID(t.ctx, t.workflowRunID)
	if err != nil {
		return err
	}

	// run as long as workflow run is not completed
	// TODO: move this to coroutine
	for !workflowRunIsComplete(workflowRun) {
		// update workflow run
		workflowRun, err = ghCtl.GetWorkflowRunWithTheID(t.ctx, workflowRunID)
		if err != nil {
			return err
		}

		workflowJobs, err := ghCtl.GetQueuedWorkflowRunJobs(t.ctx, workflowRun)
		if err != nil {
			// TODO: implement retry function
			log.Warning("failed getting workflow run jobs")
			return err
		}
		// log.DebugF("number of queued jobs: %d", workflowJobs.GetTotalCount())
		// no workflow running at the moment
		if workflowJobs.GetTotalCount() == 0 {
			time.Sleep(time.Second * 20)
			continue
		}

		for _, job := range workflowJobs.Jobs {

			jobName := fmt.Sprintf("%d-%s",workflowRunID, job.GetName())
			if _, ok := jobs[jobName]; ok {
				if jobs[jobName].status == jobStatusFailed {
					return fmt.Errorf("job %q failed, exiting", jobName)
				}

				// TODO: handle other status codes here
				// log.DebugF("job %q is already being handled with status %q", jobName, jobs[jobName].status)
				// disabled to tone down spam messages in debug
				continue
			}

			// iter trough job labels and use first runner config that matches a label
			// https://docs.github.com/en/rest/reference/actions#get-a-job-for-a-workflow-run
			var runner *RunnerType
			var label string
			for _, l := range job.Labels {
				if r, ok := runnerConfig.Runners[l]; ok {
					runner = r
					label = l
					break
				}
			}
			if runner == nil {
				// TODO: if its not like "ubuntu|windows|mac" exit with error
				log.WarningF("could not find workflow config for the job name %q", job.GetName())
			}

			provider := runner.provider
			provider.WithRunnerType(label)

			j := new(tendJob)
			jobs[jobName] = j
			j.status = jobStatusQueued
			j.maxRetry = 2

			// TODO: handle error
			go j.run(t.ctx, provider, workflowRunID)
		}

		time.Sleep(time.Second * 20)
	}
	log.Debug("waiting for all jobs to finish")
allJobs:
	for {
		for name, job := range jobs {
			if job.status != jobStatusFinished {
				if job.status == jobStatusFailed {
					// TODO: should we "log" error that job produced and print it here?
					return fmt.Errorf("job %q failed", name)
				}
				log.DebugF("job %q not finished, current status: %s", name, job.status)
				time.Sleep(time.Second * 10)
				continue allJobs
			}
		}
		break
	}

	log.Debug("all done")
	return nil
}

func workflowRunIsComplete(workflowRun *github.WorkflowRun) bool {
	return workflowRun.GetStatus() == "completed"
}