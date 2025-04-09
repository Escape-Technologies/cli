package escape

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type Location struct {
	Id            openapi_types.UUID `json:"id"`
	Name          string             `json:"name"`
	SSHPublicKey  string             `json:"sshPublicKey"`
}

func (l *Location) String() string {
	return fmt.Sprintf("%s\t%s\t%s", l.Id, l.Name, l.SSHPublicKey)
}

func GetLocations(ctx context.Context) ([]Location, error) {
	client, err := v2.NewAPIClient()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	resp, err := client.GetLocationsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get locations: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unable to get location: status code %d", resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, errors.New("unable to get location: no data received")
	}
	locations := []Location{}
	for _, location := range *resp.JSON200 {
		locations = append(locations, Location{
			Id:            *location.Id,
			Name:          *location.Name,
			SSHPublicKey:  *location.SshPublicKey,
		})
	}
	return locations, nil
}

func GetLocation(ctx context.Context, id openapi_types.UUID) (*Location, error) {
	client, err := v2.NewAPIClient()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	resp, err := client.GetLocationWithResponse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get location: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unable to get location: status code %d", resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, errors.New("unable to get location: no data received")
	}
	return &Location{
		Id:            *resp.JSON200.Id,
		Name:          *resp.JSON200.Name,
		SSHPublicKey:  *resp.JSON200.SshPublicKey,
	}, nil
}

func CreateLocation(ctx context.Context, location Location) (*Location, error) {
	client, err := v2.NewAPIClient()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	resp, err := client.CreateLocationWithResponse(ctx, v2.CreateLocationJSONRequestBody{
		Name:          location.Name,
		SshPublicKey:  location.SSHPublicKey,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create location: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unable to create location: status code %d", resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, errors.New("unable to create location: no data received")
	}
	return &Location{
		Id:            *resp.JSON200.Id,
		Name:          *resp.JSON200.Name,
		SSHPublicKey:  *resp.JSON200.SshPublicKey,
	}, nil
}
