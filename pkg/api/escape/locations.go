package escape

import (
	"context"
	"errors"
	"fmt"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
)

// ListLocations lists all locations
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

// GetLocation gets a location by ID
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

// CreateLocation creates a location
func CreateLocation(ctx context.Context, name, sshPublicKey string) (string, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return "", fmt.Errorf("unable to init client: %w", err)
	}
	req := client.LocationsAPI.CreateLocation(ctx).CreateLocationRequest(v2.CreateLocationRequest{
		Name:         name,
		SshPublicKey: sshPublicKey,
	})
	data, _, err := req.Execute()
	if err != nil {
		return "", fmt.Errorf("unable to create location: %w", err)
	}
	if data == nil || data.Id == nil {
		return "", errors.New("location created but unable to get location id")
	}
	return *data.Id, nil
}

// UpdateLocation updates a location
func UpdateLocation(ctx context.Context, id string, name, sshPublicKey string) error {
	client, err := newAPIV2Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	req := client.LocationsAPI.UpdateLocation(ctx, id).UpdateLocationRequest(v2.UpdateLocationRequest{
		Name:         &name,
		SshPublicKey: &sshPublicKey,
	})
	_, _, err = req.Execute()
	if err != nil {
		return fmt.Errorf("unable to update location: %w", err)
	}
	return nil
}

// DeleteLocation deletes a location
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

// UpsertLocation Creates or updates a location
func UpsertLocation(ctx context.Context, name, sshPublicKey string) (string, error) {
	id, err := CreateLocation(ctx, name, sshPublicKey)
	if err == nil {
		return id, nil
	}
	id, err = extractConflict(err)
	if err != nil {
		return "", fmt.Errorf("unable to extract conflict: %w", err)
	}
	return id, UpdateLocation(ctx, id, name, sshPublicKey)
}
