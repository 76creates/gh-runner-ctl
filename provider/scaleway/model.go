package scaleway

import (
	"context"

	"github.com/76creates/runner-cli/log"
	"github.com/google/uuid"
)

type Provider struct{}

func (r RunnerConfig) CreateInstance(ctx context.Context, runnerInstanceName string) error {
	log.Debug("creating and running scaleway instance")

	// generate unique ID, this will be used to tag the runner so we can
	// have a easier time looking it up, and knowing if it initialized
	runnerID := uuid.New().String()

	log.Debug("parsing cloud init script")
	cloudInit, err := r.parseCloudData(ctx, runnerInstanceName, runnerID)
	if err != nil {
		return err
	}

	c, err := r.getClient()
	if err != nil {
		return err
	}

	srv, err := r.createInstance(ctx, c, runnerInstanceName)
	if err != nil {
		log.Error("failed creating scaleway instance")
		return err
	}
	log.DebugF("created server %q", srv.ID)

	ip, err := r.attachPublicIPv4(ctx, c, srv.ID)
	if err != nil {
		log.ErrorF("failed attaching IP to the %q instance", srv.ID)
		return err
	}
	log.DebugF("attached IP %q to the server %q", ip.Address.String(), srv.ID)

	err = r.addCloudInit(ctx, c, srv.ID, cloudInit)
	if err != nil {
		log.ErrorF("failed adding user data to the %q instance", srv.ID)
		return err
	}

	err = r.startServerAndWait(c, srv.ID)
	if err != nil {
		log.ErrorF("failed powering on the %q instance", srv.ID)
		return err
	}

	log.DebugF("successfully created and ran scaleway instance %q with public IP %q", srv.ID, ip.Address.String())
	return nil
}

func (r RunnerConfig) DestroyInstance(ctx context.Context, runnerInstanceName string) error {
	// TODO: reserved IP is not getting deleted
	log.Debug("destroying scaleway instance")

	c, err := r.getClient()
	if err != nil {
		return err
	}

	server, err := r.getServerByName(ctx, c, runnerInstanceName)
	if err != nil {
		return err
	}
	if server == nil {
		log.WarningF("instance with name %q not found, assuming its already deleted", runnerInstanceName)
	}

	err = r.terminateInstance(c, server)
	if err != nil {
		return err
	}

	log.Debug("successfully destroyed scaleway instance")
	return nil
}

func (r RunnerConfig) InstanceStatus(ctx context.Context, runnerInstanceName string) error {
	return nil
}
