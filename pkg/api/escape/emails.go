package escape

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListInboxEmailsFilters holds the supported inbox email list filters.
type ListInboxEmailsFilters struct {
	Email  string
	IDs    []string
	Before *time.Time
	After  *time.Time
	Size   int
}

// ListInboxEmails lists inbox emails for a target scan email address.
func ListInboxEmails(ctx context.Context, cursor string, filters *ListInboxEmailsFilters) (*v3.ListInboxEmails200Response, error) {
	if filters == nil || strings.TrimSpace(filters.Email) == "" {
		return nil, errors.New("email is required")
	}

	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.EmailsAPI.ListInboxEmails(ctx).Email(filters.Email)
	if cursor != "" {
		req = req.Cursor(cursor)
	}
	if filters.Size > 0 {
		req = req.Size(filters.Size)
	}
	if len(filters.IDs) > 0 {
		req = req.Ids(strings.Join(filters.IDs, ","))
	}
	if filters.Before != nil {
		req = req.Before(*filters.Before)
	}
	if filters.After != nil {
		req = req.After(*filters.After)
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to list inbox emails: %w", err)
	}
	return data, nil
}

// ReadInboxEmail reads the full raw content of an inbox email by ID.
func ReadInboxEmail(ctx context.Context, id string) (*v3.ScanEmailDetails, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.EmailsAPI.ReadInboxEmail(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to read inbox email: %w", err)
	}
	return data, nil
}
