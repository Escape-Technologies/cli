package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func TestSnapshotLatestInboxStateCollectsAllNewestIDs(t *testing.T) {
	t.Parallel()

	latest := time.Date(2026, time.April, 15, 16, 47, 32, 0, time.UTC)
	older := latest.Add(-time.Minute)
	calls := 0

	list := func(_ context.Context, cursor string, filters *escape.ListInboxEmailsFilters) (*v3.ListInboxEmails200Response, error) {
		calls++
		if filters.Email != "test@scan.escape.tech" {
			t.Fatalf("expected email filter to be preserved, got %q", filters.Email)
		}

		switch cursor {
		case "":
			return &v3.ListInboxEmails200Response{
				Data: []v3.ScanEmailSummary{
					*v3.NewScanEmailSummary("id-1", latest, "a@escape.tech", "first"),
					*v3.NewScanEmailSummary("id-2", latest, "b@escape.tech", "second"),
				},
				NextCursor: stringPtr("page-2"),
			}, nil
		case "page-2":
			return &v3.ListInboxEmails200Response{
				Data: []v3.ScanEmailSummary{
					*v3.NewScanEmailSummary("id-3", latest, "c@escape.tech", "third"),
					*v3.NewScanEmailSummary("id-4", older, "d@escape.tech", "older"),
				},
			}, nil
		default:
			t.Fatalf("unexpected cursor %q", cursor)
			return nil, nil
		}
	}

	state, err := snapshotLatestInboxState(context.Background(), "test@scan.escape.tech", list)
	if err != nil {
		t.Fatalf("snapshotLatestInboxState returned error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 list calls, got %d", calls)
	}
	if !state.latest.Equal(latest) {
		t.Fatalf("expected latest timestamp %s, got %s", latest, state.latest)
	}

	for _, id := range []string{"id-1", "id-2", "id-3"} {
		if _, ok := state.ids[id]; !ok {
			t.Fatalf("expected snapshot to include %s", id)
		}
	}
	if _, ok := state.ids["id-4"]; ok {
		t.Fatalf("did not expect snapshot to include older email")
	}
}

func TestFindNextInboxEmailReturnsFirstUnseenEmail(t *testing.T) {
	t.Parallel()

	latest := time.Date(2026, time.April, 15, 16, 47, 32, 0, time.UTC)
	newest := latest.Add(time.Second)

	list := func(_ context.Context, cursor string, filters *escape.ListInboxEmailsFilters) (*v3.ListInboxEmails200Response, error) {
		if cursor != "" {
			t.Fatalf("expected a single page lookup, got cursor %q", cursor)
		}
		if filters.After == nil || !filters.After.Equal(latest) {
			t.Fatalf("expected polling filter after=%s, got %+v", latest, filters.After)
		}

		return &v3.ListInboxEmails200Response{
			Data: []v3.ScanEmailSummary{
				*v3.NewScanEmailSummary("id-1", latest, "a@escape.tech", "already-seen"),
				*v3.NewScanEmailSummary("id-9", newest, "z@escape.tech", "new"),
			},
		}, nil
	}

	next, err := findNextInboxEmail(context.Background(), "test@scan.escape.tech", inboxState{
		latest: latest,
		ids: map[string]struct{}{
			"id-1": {},
		},
	}, list)
	if err != nil {
		t.Fatalf("findNextInboxEmail returned error: %v", err)
	}
	if next == nil {
		t.Fatal("expected a new email to be returned")
	}
	if next.GetId() != "id-9" {
		t.Fatalf("expected id-9, got %s", next.GetId())
	}
}

func stringPtr(value string) *string {
	return &value
}
