package escape

import (
	"context"
	"fmt"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
	"github.com/Escape-Technologies/cli/pkg/log"
)

func ListLocations(ctx context.Context) ([]v2.ListLocations200ResponseInner, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.LocationsAPI.ListLocations(ctx)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get locations: %w", err)
	}
	return data, nil
}

func GetLocation(ctx context.Context, id string) (*v2.ListLocations200ResponseInner, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.LocationsAPI.GetLocation(ctx, id)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get location: %w", err)
	}
	return data, nil
}

func CreateLocation(ctx context.Context, name, sshPublicKey string) error {
	client, err := newAPIV2Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	req := client.LocationsAPI.CreateLocation(ctx).CreateLocationRequest(v2.CreateLocationRequest{
		Name:         name,
		SshPublicKey: sshPublicKey,
	})
	_, _, err = req.Execute()
	if err != nil {
		return fmt.Errorf("unable to create location: %w", err)
	}
	return nil
}

func UpdateLocation(ctx context.Context, id string, name, sshPublicKey string) error {
	client, err := newAPIV2Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	req := client.LocationsAPI.CreateLocation(ctx).CreateLocationRequest(v2.CreateLocationRequest{
		Name:         name,
		SshPublicKey: sshPublicKey,
	})
	_, _, err = req.Execute()
	if err != nil {
		return fmt.Errorf("unable to create location: %w", err)
	}
	return nil
}

func DeleteLocation(ctx context.Context, id string) error {
	client, err := newAPIV2Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	req := client.LocationsAPI.DeleteLocation(ctx, id)
	_, _, err = req.Execute()
	if err != nil {
		return fmt.Errorf("unable to delete location: %w", err)
	}
	return nil
}

func UpsertLocation(ctx context.Context, name, sshPublicKey string) error {
	err := CreateLocation(ctx, name, sshPublicKey)
	if err == nil {
		return nil
	}

	if oapiErr, ok := err.(v2.GenericOpenAPIError); ok {
		if conflict, ok := oapiErr.Model().(v2.CreateLocation409Response); ok {
			log.Debug("Location already exists, updating %s", conflict.InstanceId)
			return UpdateLocation(ctx, conflict.InstanceId, name, sshPublicKey)
		}
	}
	return err
}
