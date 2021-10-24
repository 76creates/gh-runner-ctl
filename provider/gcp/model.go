package gcp

import (
	"context"
	"github.com/76creates/runner-cli/log"
	"github.com/google/uuid"
)

type Provider struct{}

func (r RunnerConfig) CreateInstance(ctx context.Context, runnerInstanceName string) error {
	// generate unique ID, this will be used to tag the runner so we can
	// have a easier time looking it up, and knowing if it initialized
	runnerID := uuid.New().String()

	cloudInit, err := r.parseCloudData(ctx, runnerInstanceName, runnerID)
	if err != nil {
		return err
	}

	op, err := r.createMachine(ctx, runnerInstanceName, cloudInit)
	if err != nil {
		return err
	}

	log.Debug("waiting for the operation to finish")
	err = r.waitForComputeOP(ctx, op)
	if err != nil {
		return err
	}

	log.DebugF("successfully created instance with the name '%q'", runnerInstanceName)

	return nil
}

func (r RunnerConfig) DestroyInstance(ctx context.Context, runnerInstanceName string) error {
	err := r.destroyMachine(ctx, runnerInstanceName)
	if err != nil {
		log.ErrorF("failed destroyed an instance with the name %q", runnerInstanceName)
		return err
	}
	log.DebugF("successfully destroyed an instance with the name %q", runnerInstanceName)
	return nil
}

func (r RunnerConfig) InstanceStatus(ctx context.Context, runnerInstanceName string) error {
	return nil
}