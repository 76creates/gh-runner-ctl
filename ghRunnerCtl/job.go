package ghRunnerCtl

import (
	"context"
	"errors"
	"fmt"
	"github.com/76creates/runner-cli/ghCtl"
	"github.com/76creates/runner-cli/log"
	"github.com/76creates/runner-cli/provider"
	"github.com/google/uuid"
	"time"
)

// job is a object holder that tells us the job status and existence of certain jobName which serves as a ID essentially
type tendJob struct {
	status string
	maxRetry int
}

var (
	jobStatusQueued = "queued"
	jobStatusRunning = "running"
	jobStatusFailed = "failed"
	jobStatusFinished= "finished"
)

func (j *tendJob) run(ctx context.Context, p provider.Provider, workflowRunID int64) error {
	// name is the unique name given to GH runner and runner instance which we create here
	// we append bit of randomness to the workflow id in order to support
	// runners for multiple jobs within same workflow
	name := fmt.Sprintf("runner-%d-%s", workflowRunID, uuid.NewString()[0:8])
	log.DebugF("[%s] running the job", name)

	j.status = jobStatusRunning

	// creating a runner
	log.DebugF("[%s] creating the runner", name)
	done := false
	for try := 0; try < j.maxRetry; try++ {
		if p.WantGithubRegistrationToken() {
			// generate github runner registration token
			token, err := ghCtl.GenerateRunnerToken(ctx)
			if err != nil {
				log.Error(err.Error())
				return err
			}
			p.WithGithubRegistrationToken(token)
		}

		err := p.CreateInstance(ctx, name)
		if err != nil {
			// TODO: cleanup here
			log.ErrorF("[%s] error while creating the instance: %s", name, err.Error())
			time.Sleep(time.Second * 5)
			continue
		}

		// TODO: create a logging child function to integrate job name into all lines ran by it
		log.DebugF("[%s] created instance successfully", name)
		done = true
		break
	}
	if ! done {
		j.status = jobStatusFailed
		return errors.New("failed completing the job")
	}

	// waiting for runner to become active
	if p.WantGithubRegistrationToken() {
		log.DebugF("[%s] waiting for a runner to become active", name)
		err := ghCtl.WaitForRunnerToBecomeActive(ctx, name)
		if err != nil {
			log.ErrorF("[%s] error while waiting for runner to become active: %s", err.Error() )
			j.status = jobStatusFailed
			return err
		}
	}

	// waiting for runner to finish executing
	if p.WantGithubRegistrationToken() {
		log.DebugF("[%s] waiting for a runner to finish executing", name)
		// TODO: allow for custom wait time
		err := ghCtl.WaitForRunnerToBeDeRegistered(ctx, name, 20,time.Second*30)
		if err != nil {
			log.ErrorF("[%s] error while waiting for runner to de-register: %s", name, err.Error() )
			j.status = jobStatusFailed
			return err
		}
	}

	// deleting a runner
	log.DebugF("[%s] deleting the runner", name)
	done = false
	for try := 0; try < j.maxRetry; try++ {
		err := p.DestroyInstance(ctx, name)
		if err != nil {
			// TODO: cleanup here
			log.ErrorF("[%s] error while deleting the instance: %s", name, err.Error())
			time.Sleep(time.Second * 5)
			continue
		}

		// TODO: create a logging child function to integrate job name into all lines ran by it
		log.DebugF("[%s] deleted the instance successfully", name)
		done = true
		break
	}
	if ! done {
		j.status = jobStatusFailed
		return errors.New("failed completing the job")
	}

	log.DebugF("[%s] finished successfully", name)
	j.status = jobStatusFinished
	return nil
}