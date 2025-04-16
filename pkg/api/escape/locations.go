package escape

import (
	"context"
	"fmt"
	"net/http"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
)

func ListLocations(ctx context.Context) ([]v2.ListLocations200ResponseInner, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.LocationsAPI.ListLocations(ctx)
	data, resp, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get locations: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to get location: status code %d", resp.StatusCode)
	}
	return data, nil
}

func GetLocation(ctx context.Context, id string) (*v2.ListLocations200ResponseInner, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.LocationsAPI.GetLocation(ctx, id)
	data, resp, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get location: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to get location: status code %d", resp.StatusCode)
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
	_, resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("unable to create location: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to create location: status code %d", resp.StatusCode)
	}
	return nil
}
