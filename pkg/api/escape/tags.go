package escape

import (
	"context"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListTags lists all tags
func ListTags(ctx context.Context) ([]v3.TagDetail, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.TagsAPI.ListTags(ctx)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// CreateTag creates a tag
func CreateTag(ctx context.Context, name string, color string) (*v3.TagDetail, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.TagsAPI.CreateTag(ctx)
	data, _, err := req.CreateTagRequest(v3.CreateTagRequest{
		Name: name,
		Color: color,
	}).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// DeleteTag deletes a tag
func DeleteTag(ctx context.Context, id string) error {
	client, err := newAPIV3Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	req := client.TagsAPI.DeleteTag(ctx, id)
	_, _, err = req.Execute()
	if err != nil {
		return fmt.Errorf("unable to delete tag: %w", err)
	}
	return nil
}