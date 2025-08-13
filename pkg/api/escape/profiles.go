package escape

import (
	"context"
	"errors"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListProfiles lists all profiles
func ListProfiles(ctx context.Context, next string) ([]v3.ProfileSummarized, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.ProfilesAPI.ListProfiles(ctx)
	if next != "" {
		req = req.Cursor(next)
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("api error: %w", err)
	}
	return data.Data, data.NextCursor, nil
}

// GetProfile gets a profile by ID
func GetProfile(ctx context.Context, profileID string) (*v3.ProfileDetailed, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.ProfilesAPI.GetProfile(ctx, profileID)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// DeleteProfile deletes a profile by ID
func DeleteProfile(ctx context.Context, profileID string) error {
	client, err := newAPIV3Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	_, _, err = client.ProfilesAPI.DeleteProfile(ctx, profileID).Execute()
	if err != nil {
		return fmt.Errorf("api error: %w", err)
	}
	return nil
}

// UpdateProfile updates a profile by ID
func UpdateProfile(
	ctx context.Context,
	profileID string,
	profileName *string,
	profileCron *string,
) error {
	client, err := newAPIV3Client()
	req := client.ProfilesAPI.UpdateProfile(ctx, profileID)
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}

	updateProfileRequest := v3.UpdateProfileRequest{}
	if profileName != nil {
		updateProfileRequest.Name = profileName
	}
	if profileCron != nil {
		updateProfileRequest.Cron = profileCron
	}
	req = req.UpdateProfileRequest(updateProfileRequest)

	_, _, err = req.Execute()
	if err != nil {
		return fmt.Errorf("api error: %w", err)
	}
	return nil
}

// UpdateProfileSchema updates a profile schema by profileID
func UpdateProfileSchema(
	ctx context.Context,
	profileID string,
	newProfileSchemaURL *string,
) error {
	client, err := newAPIV3Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}

	if newProfileSchemaURL != nil {
		updateProfileSchemaRequest := v3.UpdateProfileSchemaRequest{
			ProfileId: profileID,
			SchemaUrl: *newProfileSchemaURL,
		}
		req := client.ProfilesAPI.UpdateProfileSchema(ctx, profileID).UpdateProfileSchemaRequest(updateProfileSchemaRequest)

		_, _, err := req.Execute()
		if err != nil {
			return fmt.Errorf("api error: %w", err)
		}

		return nil
	}

	return errors.New("no new profile schema url provided")
}

// UpdateProfileConfiguration updates a profile schema by profileID
func UpdateProfileConfiguration(
	ctx context.Context,
	profileID string,
	configuration *map[string]string,
) error {
	client, err := newAPIV3Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}

	if configuration != nil {
		cfg := make(map[string]interface{}, len(*configuration))
		for k, v := range *configuration {
			cfg[k] = v
		}
		updateProfileConfigurationRequest := v3.UpdateProfileConfigurationRequest{
			Configuration: cfg,
		}
		req := client.ProfilesAPI.UpdateProfileConfiguration(ctx, profileID).UpdateProfileConfigurationRequest(updateProfileConfigurationRequest)

		_, apiResp, err := req.Execute()
		fmt.Println(apiResp)
		if err != nil {
			return fmt.Errorf("api error: %w", err)
		}

		return nil
	}

	return errors.New("no configuration provided")
}
