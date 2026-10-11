package escape

import (
	"context"
	"encoding/json"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// StartValidation asks the platform to validate an asset reachability and
// returns the validation handle used to poll its status.
func StartValidation(ctx context.Context, body []byte) (*v3.StartValidation200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.StartValidationRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON for StartValidationRequest: %w", err)
	}

	data, _, err := client.ValidationsAPI.StartValidation(ctx).StartValidationRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// GetValidation returns the status, progress and events of a validation.
func GetValidation(ctx context.Context, validationID string) (*v3.GetValidation200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.ValidationsAPI.GetValidation(ctx, validationID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}
