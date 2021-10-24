package scaleway

import (
	"context"
	"errors"
	"fmt"
	"github.com/76creates/runner-cli/ghCtl"
	"github.com/76creates/runner-cli/log"
	"github.com/76creates/runner-cli/provider"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"io"
	"strings"
	"time"
)

// getClient initialize and get Scaleway API
func (r *RunnerConfig) getClient() (*scw.Client, error) {
	log.Debug("authenticating Scaleway client")
	client, err := scw.NewClient(
		// Get your credentials at https://console.scaleway.com/project/credentials
		scw.WithDefaultOrganizationID(*r.Access.OrgID),
		scw.WithAuth(*r.Access.KeyID, *r.Access.KeySecret),
	)
	if err != nil {
		return nil, err
	}

	log.Debug("successfully authenticated Scaleway client")
	return client, nil
}

// createInstance creates new scaleway instance, this instance is bare and stopped after this action
func (r *RunnerConfig) createInstance(ctx context.Context, client *scw.Client, name string) (*instance.Server, error) {
	log.Debug("creating scaleway instance")

	api := instance.NewAPI(client)

	project := r.Access.ProjectID
	if project != nil {
		project = r.Access.OrgID
	}
	dynamicIP := true

	request := instance.CreateServerRequest{
		Zone:              scw.Zone(*r.Zone),
		Name:              name,
		DynamicIPRequired: &dynamicIP,
		CommercialType:    *r.InstanceType,
		Image:             *r.Image,
		Project:           project,
	}
	if r.SecurityGroup != nil {
		request.SecurityGroup = r.SecurityGroup
	}
	if r.Tags != nil {
		request.Tags = *r.Tags
	}

	resp, err := api.CreateServer(&request, scw.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp.Server, nil
}

// attachPublicIPv4 adds cloud init user data to be ran at the instance startup
func (r *RunnerConfig) attachPublicIPv4(ctx context.Context, client *scw.Client, serverID string) (*instance.IP, error) {
	log.Debug("attaching IPv4 to the instance instance")

	api := instance.NewAPI(client)

	project := r.Access.ProjectID
	if project != nil {
		project = r.Access.OrgID
	}
	request := instance.CreateIPRequest{
		Zone:         scw.Zone(*r.Zone),
		Project:      project,
		Server:       &serverID,
	}
	if r.Tags != nil {
		request.Tags = *r.Tags
	}

	resp, err := api.CreateIP(&request, scw.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp.IP, nil
}

// addCloudInit adds cloud init user data to be ran at the instance startup
func (r *RunnerConfig) addCloudInit(ctx context.Context, client *scw.Client, serverID string, cloudInit *io.Reader) error {
	log.Debug("adding cloud-init to the instance instance")

	api := instance.NewAPI(client)

	request := instance.SetServerUserDataRequest{
		Zone:     scw.Zone(*r.Zone),
		ServerID: serverID,
		Key:      "cloud-init",
		Content:  *cloudInit,
	}

	return api.SetServerUserData(&request, scw.WithContext(ctx))
}

func (r *RunnerConfig) parseCloudData(ctx context.Context, runnerName, runnerID string) (*io.Reader, error) {
	if r.CloudInit == nil {
		log.Warning("cloud init is null, nothing to parse")
		return nil, nil
	}
	log.Debug("parsing cloud init")

	cloudInitData := provider.CloudInitData {
		GithubRepo: fmt.Sprintf("%s/%s", ghCtl.GetRepoOwner(ctx), ghCtl.GetRepoName(ctx)),
		GithubRunnerName: runnerName,
		GithubRunnerToken: r.GithubRegistrationToken,
		GithubRunnerType: r.RunnerType,
		GithubRunnerUniqueID: runnerID,
	}
	cloudInitParsed, err := provider.ParseCloudInit(*r.CloudInit, cloudInitData)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	var reader io.Reader
	reader = strings.NewReader(*cloudInitParsed)

	return &reader, nil
}

// startServerAndWait sends power-on call and waits for it to execute
// it will timeout after 5 minutes
func (r *RunnerConfig) startServerAndWait(client *scw.Client, serverID string) error {
	log.Debug("powering on instance and waiting for action to complete")

	api := instance.NewAPI(client)

	timeout := time.Minute * 5
	request := instance.ServerActionAndWaitRequest{
		ServerID:      serverID,
		Zone:          scw.Zone(*r.Zone),
		Action:        instance.ServerAction("poweron"),
		Timeout:       &timeout,
	}

	return api.ServerActionAndWait(&request)
}

// getServerByName lists servers by name, it will then check if there is a name that is an
// exact match since on the Scaleway side it works more like "contains" than "is"
func (r *RunnerConfig) getServerByName(ctx context.Context, client *scw.Client, name string) (*instance.Server, error) {
	log.DebugF("looking up server object with name %q", name)

	api := instance.NewAPI(client)

	project := r.Access.ProjectID
	if project != nil {
		project = r.Access.OrgID
	}
	var perPage uint32 = 100
	request := instance.ListServersRequest{
		Zone:           scw.Zone(*r.Zone),
		PerPage:        &perPage,
		// TODO: enable pagination
		//Page:           nil,
		Project:        project,
		Name:           &name,
	}
	if r.Tags != nil {
		request.Tags = *r.Tags
	}

	resp, err := api.ListServers(&request, scw.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	if resp.TotalCount == 0 {
		log.WarningF("found 0 servers containing %q", name)
		return nil, nil
	}

	if resp.TotalCount == 1 {
		if resp.Servers[0].Name == name {
			return resp.Servers[0], nil
		}
		return nil, errors.New(fmt.Sprintf("failed exact matching server with name %q", name))
	}

	// here we check for the exact name match as per Scaleway API:
	// "server1" will return "server100" and "server1"
	log.Warning("matched multiple servers, attempting to find the right one")
	found := 0
	var server *instance.Server
	for _, s := range resp.Servers {
		if s.Name == name {
			if found > 1 {
				return nil, errors.New(fmt.Sprintf("matched multiple servers with the name %q", name))
			}
			server = s
		}
	}

	log.Debug("successfully matched one server by name")
	return server, nil
}

// terminateInstance sends terminate call and waits for it to execute
// it will timeout after 2 minutes
func (r *RunnerConfig) terminateInstance(client *scw.Client, server *instance.Server) error {
	log.DebugF("sending terminate action to the instance %q", server.ID)

	api := instance.NewAPI(client)
	timeout := time.Minute * 2
	request := instance.ServerActionAndWaitRequest{
		ServerID:      server.ID,
		Zone:          scw.Zone(*r.Zone),
		Action:        instance.ServerAction("terminate"),
		Timeout:       &timeout,
	}

	err := api.ServerActionAndWait(&request)
	if err != nil {
		// scaleway sdk for action and wait does not handle termination requests well
		// generally speaking these requests for termination dont even support stopped instances
		// still rather do this than create own loop with deleteInstance and getInstance to
		// check if its deleted or not, this one is a safe shot
		if instanceNotFoundError(err, server.ID) {
			return nil
		}
	}

	return err
}

// instanceNotFoundError since scw sdk does not have a proper error type for this error as far as I see
// we will use this primitive error check, if error type is found in future please replace this
func instanceNotFoundError(err error, instanceID string) bool {
	if err.Error() == fmt.Sprintf("scaleway-sdk-go: waiting for server failed: scaleway-sdk-go: resource  with ID %s is not found", instanceID) {
		return true
	}
	return false
}