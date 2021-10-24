package ghCtl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/76creates/runner-cli/log"
	"github.com/google/go-github/v39/github"
)

type Runner struct {
}

// GenerateRunnerToken generates registration token which is used on the self hosted runner
// in order to register it, returns token string
func GenerateRunnerToken(ctx context.Context) (string, error) {
	log.Debug("generating gitlab runner registration token")
	c := getClient(ctx)

	token, resp, err := c.Actions.CreateRegistrationToken(
		ctx, GetRepoOwner(ctx), GetRepoName(ctx),
	)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 201 {
		return "", errors.New(
			fmt.Sprintf("Didnt get expected status code(201), got %d", resp.StatusCode),
		)
	}

	return token.GetToken(), nil
}

// WaitForRunnerToBecomeActive waits for runner to spawn, and the waits for it to
// exit the offline state, each action timeouts after 2 minutes
func WaitForRunnerToBecomeActive(ctx context.Context, label string) error {
	runner, err := waitForLabeledRunnerToSpawn(ctx, label)
	if err != nil {
		return err
	}

	err = waitForRunnerStateActive(ctx, runner.GetID())
	if err != nil {
		return err
	}

	log.DebugF("runner %q is active", runner.GetID())
	return nil
}

// WaitForRunnerToBecomeOffline waits for the runner to enter offline status
// we would use this to know when the runner has ended its execution with the
// ephemeral flag on
func WaitForRunnerToBecomeOffline(ctx context.Context, label string) error {
	runner, err := getOneRunnerByLabel(ctx, label)
	if err != nil {
		return err
	}

	retryWait := time.Second*30
	err = waitForRunnerStateEqual(ctx, "offline", runner.GetID(), 20, retryWait)
	if err != nil {
		return err
	}

	log.DebugF("runner %q is offline", runner.GetID())
	return nil
}

// WaitForRunnerToBeDeRegistered waits for the runner to de-register itself,
// that is we wait for the runner to go missing
func WaitForRunnerToBeDeRegistered(ctx context.Context, label string, retryCount int, waitRetry time.Duration) error {
		log.DebugF("wait for runner with label '%q' to de-register", label)

		// TODO: this here is a bit racy, runner in theory could finish faster than this
		runner, err := getOneRunnerByLabel(ctx, label)
		if err != nil {
			return err
		}

		for retry := 0; retry < retryCount; retry++ {
		_, err := getRunnerByID(ctx, runner.GetID())
		if err != nil {
			if _, ok := err.(*RunnerNotFound); !ok {
				log.Debug("runner not found, ergo de-registered")
				return nil
			}
			return err
		}
		log.DebugF("runner %d still registered", runner.GetID())
		time.Sleep(waitRetry)
	}

	return fmt.Errorf("runner %d still registered after the timeout", runner.GetID())
}


func ListRunnersLabeled(ctx context.Context) error {
	runners, err := listRunnersLabeled(ctx, "on-demand")
	if err != nil {
		return err
	}

	for _, runner := range runners {
		fmt.Printf("%s / %s\n", *runner.Name, *runner.Status)
	}

	return nil
}

// listRunners return full list of runners
// TODO: implement pagination
func listRunners(ctx context.Context) ([]*github.Runner, error) {
	c := getClient(ctx)

	opts := github.ListOptions{PerPage: 100}
	runners, resp, err := c.Actions.ListRunners(
		ctx, GetRepoOwner(ctx), GetRepoName(ctx), &opts)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(
			fmt.Sprintf("Didnt get expected status code(200), got %d", resp.StatusCode),
		)
	}

	return runners.Runners, nil
}

// listRunnersLabeled return list of runners that contain a label
func listRunnersLabeled(ctx context.Context, label string) ([]*github.Runner, error) {
	var runnersLabeled []*github.Runner

	runners, err := listRunners(ctx)
	if err != nil {
		return nil, err
	}

	for _, runner := range runners {
		for _, runnerLabel := range runner.Labels {
			if runnerLabel.GetName() == label {
				runnersLabeled = append(runnersLabeled, runner)
				break
			}
		}
	}

	return runnersLabeled, nil
}

// removeRunner removes the github runner, this does not de-register the runner
func removeRunner(ctx context.Context, runner *github.Runner) error {
	c := getClient(ctx)

	log.DebugF("attempting to remove runner with name/id: %q/%q", runner.GetName(), runner.GetID())
	resp, err := c.Actions.RemoveRunner(
		ctx, GetRepoOwner(ctx), GetRepoName(ctx), runner.GetID())
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(
			fmt.Sprintf("Didnt get expected status code(200), got %d", resp.StatusCode),
		)
	}

	log.DebugF("successfully removed runner with name/id: %q/%q", runner.GetName(), runner.GetID())
	return nil
}

// removeRunnerLabeled removes the github runner with a label, this does not de-register the runners
func removeRunnerLabeled(ctx context.Context, label string) error {
	runners, err := listRunnersLabeled(ctx, label)
	if err != nil {
		return err
	}

	if len(runners) == 0 {
		log.WarningF("found 0 runners with label %q", label)
		return nil
	}
	log.DebugF("found %d runners with label %q", len(runners), label)

	for _, runner := range runners {
		removeRunner(ctx, runner)
	}

	return nil
}

// getRunnerByID fetches runner with by the ID
func getRunnerByID(ctx context.Context, id int64) (*github.Runner, error) {
	c := getClient(ctx)

	log.DebugF("attempting to get the runner with the id: %d", id)
	runner, resp, err := c.Actions.GetRunner(
		ctx, GetRepoOwner(ctx), GetRepoName(ctx), id)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return nil, &RunnerNotFound{id: id}
		}
		return nil, errors.New(
			fmt.Sprintf("Didnt get expected status code(200), got %d", resp.StatusCode),
		)
	}

	log.DebugF("successfully got the runner with id: %d", id)
	return runner, nil
}

// getOneRunnerByLabel tries to get one runner by label provided, this is useful
// when getting a runner with a unique ID
func getOneRunnerByLabel(ctx context.Context, label string) (*github.Runner, error) {
	runners, err := listRunnersLabeled(ctx, label)
	if err != nil {
		return nil, err
	}

	if len(runners) == 0 {
		return nil, &RunnerNotFound{label: label}
	}
	if len(runners) > 1 {
		return nil, &MultipleRunnersFound{count: len(runners), label: label}
	}

	return runners[0], nil
}

type RunnerNotFound struct {
	id int64
	label string
}

func (e *RunnerNotFound) Error() string {
	if e.label != "" {
		e.label = fmt.Sprintf(" labeled %q", e.label)
	}
	return fmt.Sprintf("[ RunnerNotFound ] couldnt find a runner%s", e.label)
}

type MultipleRunnersFound struct {
	count int
	label string
}

func (e *MultipleRunnersFound) Error() string {
	if e.label != "" {
		e.label = fmt.Sprintf(" labeled %q", e.label)
	}
	return fmt.Sprintf("[ MultipleRunnersFound ] found %d runners%s", e.count, e.label)
}

// waitForLabeledRunnerToSpawn wait for runner to appear on the GH
func waitForLabeledRunnerToSpawn(ctx context.Context, label string) (*github.Runner, error) {
	log.DebugF("wait for runner labeled %q to spawn", label)

	// try getting runner for 2 minutes
	for retry := 0; retry < 12; retry++ {
		runner, err := getOneRunnerByLabel(ctx, label)
		if err != nil {
			log.Error(err.Error())
			if _, ok := err.(*RunnerNotFound); !ok {
				return nil, err
			}

			time.Sleep(time.Second * 5)
			continue
		}

		log.Warning("success")
		return runner, nil
	}

	return nil, &OperationTimeout{operation: "wait for runner to be created"}
}

type OperationTimeout struct {
	operation string
}

func (e *OperationTimeout) Error() string {
	return fmt.Sprintf("[ OperationTimeout ] timeout while executing operation: %s", e.operation)
}

// waitForRunnerStateActive wait for a runner to enter active state, meaning its not offline
func waitForRunnerStateActive(ctx context.Context, id int64) error {
	log.DebugF("wait for runner '%d' to exit 'offline' status", id)

	// wait for runner to exit offline status for 2 minutes
	for retry := 0; retry < 12; retry++ {
		runner, err := getRunnerByID(ctx, id)
		if err != nil {
			return err
		}

		if runner.GetStatus() != "offline" {
			return nil
		}

		log.DebugF("runner id '%d' status is %q", id, runner.GetStatus())
		time.Sleep(time.Second * 10)
	}

	return &OperationTimeout{operation: "wait for the runner to exit offline state"}
}

// waitForRunnerStateActive wait for a runner to enter a certain state state
func waitForRunnerStateEqual(ctx context.Context, state string, id int64, retryCount int, waitRetry time.Duration) error {
	log.DebugF("wait for runner '%d' to transition to 'offline' state", id)

	for retry := 0; retry < retryCount; retry++ {
		runner, err := getRunnerByID(ctx, id)
		if err != nil {
			return err
		}

		if runner.GetStatus() == state {
			log.DebugF("runner entered state %q", state)
			return nil
		}

		log.DebugF("runner id '%d' status is %q", id, runner.GetStatus())
		time.Sleep(waitRetry)
	}

	return &OperationTimeout{operation: fmt.Sprintf("wait for the runner to enter %q state", state)}
}
