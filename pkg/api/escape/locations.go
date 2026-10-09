package escape

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListLocationsFilters holds optional filters for listing locations
type ListLocationsFilters struct {
	Search        string
	Enabled       bool
	LocationTypes []string
	SortType      string
	SortDirection string
}

// ListLocations lists one page of locations.
// size 0 keeps the API default page size. The returned total is totalCount.
func ListLocations(ctx context.Context, next string, filters *ListLocationsFilters, size int) ([]v3.LocationSummarized, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.LocationsAPI.ListLocations(ctx)
	if next != "" {
		req = req.Cursor(next)
	}

	if size > 0 {
		req = req.Size(size)
	}

	if filters != nil {
		if filters.SortType != "" {
			req = req.SortType(filters.SortType)
		}

		if filters.SortDirection != "" {
			req = req.SortDirection(filters.SortDirection)
		}

		if filters.Search != "" {
			req = req.Search(filters.Search)
		}

		if filters.Enabled {
			req = req.Enabled(strconv.FormatBool(filters.Enabled))
		}

		if len(filters.LocationTypes) > 0 {
			normalized := make([]string, 0, len(filters.LocationTypes))
			for _, locationType := range filters.LocationTypes {
				trimmed := strings.TrimSpace(locationType)
				if trimmed == "" {
					continue
				}

				normalized = append(normalized, strings.ToUpper(trimmed))
			}

			if len(normalized) > 0 {
				req = req.Type_(normalized)
			}
		}
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to get locations: %w", humanizeAPIError(err))
	}

	return data.Data, data.NextCursor, data.GetTotalCount(), nil
}

// GetLocation gets a location by ID
func GetLocation(ctx context.Context, id string) (*v3.GetLocation200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.LocationsAPI.GetLocation(ctx, id)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get location: %w", humanizeAPIError(err))
	}

	return data, nil
}

// CreateLocation creates a location and returns the created location.
func CreateLocation(ctx context.Context, name, sshPublicKey string) (*v3.CreateLocation200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.LocationsAPI.CreateLocation(ctx).CreateLocationRequest(v3.CreateLocationRequest{
		Name:         name,
		SshPublicKey: sshPublicKey,
	})
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to create location: %w", humanizeAPIError(err))
	}

	if data == nil || data.Id == nil {
		return nil, errors.New("location created but unable to get location id")
	}

	return data, nil
}

// UpdateLocation updates a location. Only non-nil fields are sent.
// The returned value is the location after the update.
func UpdateLocation(ctx context.Context, id string, name, sshPublicKey *string, enabled *bool) (*v3.GetLocation200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	body := v3.UpdateLocationRequest{}
	if name != nil {
		body.Name = name
	}

	if sshPublicKey != nil {
		body.SshPublicKey = sshPublicKey
	}

	if enabled != nil {
		body.Enabled = enabled
	}

	req := client.LocationsAPI.UpdateLocation(ctx, id).UpdateLocationRequest(body)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to update location: %w", humanizeAPIError(err))
	}

	if data == nil {
		return nil, errors.New("location updated but response was empty")
	}

	return data, nil
}

// DeleteLocation deletes a location and returns the API's acknowledgement.
func DeleteLocation(ctx context.Context, id string) (*v3.DeleteLocation200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.LocationsAPI.DeleteLocation(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to delete location: %w", humanizeAPIError(err))
	}

	return data, nil
}

// UpsertLocation Creates or updates a location and returns its ID.
func UpsertLocation(ctx context.Context, name, sshPublicKey string) (string, error) {
	created, createErr := CreateLocation(ctx, name, sshPublicKey)
	if createErr == nil {
		return created.GetId(), nil
	}

	id, extractErr := extractConflict(createErr)
	if extractErr != nil {
		return "", createErr
	}

	if _, err := UpdateLocation(ctx, id, &name, &sshPublicKey, nil); err != nil {
		return "", err
	}

	return id, nil
}
