package gcp

import (
	compute "cloud.google.com/go/compute/apiv1"
	"context"
	"fmt"
	"github.com/76creates/runner-cli/ghCtl"
	"github.com/76creates/runner-cli/log"
	"github.com/76creates/runner-cli/provider"
	"google.golang.org/api/option"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

// getClientAuthOption return option for authentification client, contains json credentials
func (r *RunnerConfig)getClientAuthOption() option.ClientOption{
	return option.WithCredentialsJSON([]byte(*r.Access.JSON))
}

func (r *RunnerConfig)createMachine(ctx context.Context, runnerInstanceName string, cloudInit *string) (*compute.Operation, error) {
	log.Debug("creating machine")

	log.Debug("getting instance client")
	clientInstance, err := compute.NewInstancesRESTClient(ctx, r.getClientAuthOption())
	if err != nil {
		return nil, err
	}
	defer clientInstance.Close()

	req := &computepb.InsertInstanceRequest{
		Project: *r.Project,
		Zone:    *r.Zone,
		InstanceResource: &computepb.Instance{
			Name: proto.String(runnerInstanceName),
			Disks: []*computepb.AttachedDisk{
				{
					InitializeParams: &computepb.AttachedDiskInitializeParams{
						DiskSizeGb:  proto.Int64(10),
						SourceImage: r.Image,
					},
					AutoDelete: proto.Bool(true),
					Boot:       proto.Bool(true),
					Type:       computepb.AttachedDisk_PERSISTENT.Enum(),
				},
			},
			MachineType: proto.String(fmt.Sprintf("zones/%s/machineTypes/%s", *r.Zone, *r.MachineType)),
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					Name: r.NetworkName,
					AccessConfigs: []*computepb.AccessConfig{
						{
							Type: computepb.AccessConfig_ONE_TO_ONE_NAT.Enum(),
							Name: proto.String("External NAT"),
						},
					},
				},
			},
		},
	}


	// add cloud-init script to instance
	userDataKey := "user-data"
	userData := computepb.Items{
		Key: &userDataKey,
		Value: cloudInit,
	}
	metadata := new(computepb.Metadata)
	metadata.Items = append(metadata.Items, &userData)
	req.InstanceResource.Metadata = metadata

	log.Debug("making an request")
	return clientInstance.Insert(ctx, req)
}

func (r *RunnerConfig)destroyMachine(ctx context.Context, instanceName string) error {
	log.Debug("destroying machine")

	clientInstance, err := compute.NewInstancesRESTClient(ctx, r.getClientAuthOption())
	if err != nil {
		return err
	}
	defer clientInstance.Close()

	req := &computepb.DeleteInstanceRequest{
		Instance: instanceName,
		Project: *r.Project,
		Zone: *r.Zone,
	}

	log.DebugF("deleting a machine with ID %q", req.Instance)
	resp, err := clientInstance.Delete(ctx, req)
	if err != nil {
		return err
	}

	// TODO: check if resp needs to be debugged for an error as well

	log.DebugF("deleted the machine with ID %q with the code %q", req.Instance, resp.Proto().Status.String())
	return nil
}

func (r *RunnerConfig) parseCloudData(ctx context.Context, runnerName, runnerID string) (*string, error) {
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

	return cloudInitParsed, nil
}


func (r *RunnerConfig) waitForComputeOP(ctx context.Context, op *compute.Operation) error {
	log.DebugF("waiting for zone op %q", op.Proto().GetName())

	log.Debug("getting zone op client")
	zoneOperationsClient, err := compute.NewZoneOperationsRESTClient(ctx, r.getClientAuthOption())
	if err != nil {
		return err
	}
	defer zoneOperationsClient.Close()

	for {
		waitReq := &computepb.WaitZoneOperationRequest{
			Operation: op.Proto().GetName(),
			Project:   *r.Project,
			Zone:      *r.Zone,
		}
		zoneOp, err := zoneOperationsClient.Wait(ctx, waitReq)
		if err != nil {
			log.DebugF("operation %q had an error", op.Proto().GetName())
			return err
		}

		if *zoneOp.Status.Enum() == computepb.Operation_DONE {
			log.DebugF("operation %q is done", op.Proto().GetName())
			return nil
		}
	}
}