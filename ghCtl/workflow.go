package ghCtl

import (
	"context"
	"fmt"

	"github.com/76creates/runner-cli/log"
	"github.com/google/go-github/v39/github"
)

func GetWorkflowRunWithTheID(ctx context.Context, workflowRunID int64) (*github.WorkflowRun, error) {
	log.Debug("getting workflow run")

	c := getClient(ctx)

	run, resp, err := c.Actions.GetWorkflowRunByID(
		ctx, GetRepoOwner(ctx), GetRepoName(ctx), workflowRunID)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("didnt get expected status code(200), got %d", resp.StatusCode)
	}

	log.DebugF("successfully got the workflow run with id: %d", workflowRunID)
	return run, nil
}

func GetQueuedWorkflowRunJobs(ctx context.Context, run *github.WorkflowRun) (*github.Jobs, error) {
	// TODO: pagination
	log.Debug("getting queued workflow run jobs")

	c := getClient(ctx)

	jobsAll, resp, err := c.Actions.ListWorkflowJobs(
		ctx, GetRepoOwner(ctx), GetRepoName(ctx), run.GetID(), &github.ListWorkflowJobsOptions{})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("didnt get expected status code(200), got %d", resp.StatusCode)
	}

	jobsQueued := new(github.Jobs)
	jobsQueuedCount := 0
	for _ , job := range jobsAll.Jobs {
		if job.GetStatus() == "queued" {
			jobsQueued.Jobs = append(jobsQueued.Jobs, job)
			jobsQueuedCount ++
		}
	}
	jobsQueued.TotalCount = &jobsQueuedCount

	log.DebugF("successfully got the workflow runs from the workflow id: %d", run.GetID())
	return jobsQueued, nil
}

func GetWorkflow(ctx context.Context, run *github.WorkflowRun) (*github.Workflow, error) {
	log.Debug("getting workflow file path")

	c := getClient(ctx)

	workflow, resp, err := c.Actions.GetWorkflowByID(
		ctx, GetRepoOwner(ctx), GetRepoName(ctx), run.GetWorkflowID())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("didnt get expected status code(200), got %d", resp.StatusCode)
	}

	return workflow, nil
}
