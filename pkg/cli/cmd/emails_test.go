package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
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

func TestRunEmailWaitInFlightDeadlineReturnsPending(t *testing.T) {
	withEmailWait(t, 50*time.Millisecond, true)
	useJSONOutput(t)

	var err error
	stdout := captureStdout(t, func() {
		err = runEmailWait(context.Background(), "test@scan.escape.tech", func(ctx context.Context, _ string, _ *escape.ListInboxEmailsFilters) (*v3.ListInboxEmails200Response, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		})
	})
	if err != nil {
		t.Fatalf("expected pending result, got %v", err)
	}

	assertPendingEmail(t, stdout)
}

func TestRunEmailWaitIdleDeadlineReturnsPending(t *testing.T) {
	withEmailWait(t, 50*time.Millisecond, true)
	useJSONOutput(t)

	var err error
	stdout := captureStdout(t, func() {
		err = runEmailWait(context.Background(), "test@scan.escape.tech", func(context.Context, string, *escape.ListInboxEmailsFilters) (*v3.ListInboxEmails200Response, error) {
			return &v3.ListInboxEmails200Response{}, nil
		})
	})
	if err != nil {
		t.Fatalf("expected pending result, got %v", err)
	}

	assertPendingEmail(t, stdout)
}

func TestRunEmailWaitDeadlineWithoutPendingFlagIsAnError(t *testing.T) {
	withEmailWait(t, 50*time.Millisecond, false)

	err := runEmailWait(context.Background(), "test@scan.escape.tech", func(ctx context.Context, _ string, _ *escape.ListInboxEmailsFilters) (*v3.ListInboxEmails200Response, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})
	if err == nil || err.Error() != "timed out waiting for a new email" {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestRunEmailWaitCancelIsNotPending(t *testing.T) {
	withEmailWait(t, 0, true)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runEmailWait(ctx, "test@scan.escape.tech", func(context.Context, string, *escape.ListInboxEmailsFilters) (*v3.ListInboxEmails200Response, error) {
		return nil, errors.New("stopped")
	})
	if err == nil || !strings.Contains(err.Error(), "email wait canceled") {
		t.Fatalf("expected cancel error, got %v", err)
	}
}

func withEmailWait(t *testing.T, timeout time.Duration, pending bool) {
	t.Helper()

	prevTimeout := emailsWaitTimeout
	prevPending := emailsWaitPendingOnTimeout
	prevInterval := emailsWaitPollInterval
	prevAfter := emailsWaitAfter
	prevSeen := append([]string(nil), emailsWaitSeenIDs...)
	emailsWaitTimeout = timeout
	emailsWaitPendingOnTimeout = pending
	emailsWaitPollInterval = time.Hour
	emailsWaitAfter = ""
	emailsWaitSeenIDs = nil
	t.Cleanup(func() {
		emailsWaitTimeout = prevTimeout
		emailsWaitPendingOnTimeout = prevPending
		emailsWaitPollInterval = prevInterval
		emailsWaitAfter = prevAfter
		emailsWaitSeenIDs = prevSeen
	})
}

func useJSONOutput(t *testing.T) {
	t.Helper()
	if err := out.SetOutput("json"); err != nil {
		t.Fatalf("set output: %v", err)
	}

	t.Cleanup(func() {
		_ = out.SetOutput("pretty")
	})
}

func assertPendingEmail(t *testing.T, stdout string) emailWaitPending {
	t.Helper()

	var payload emailWaitPending
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("expected JSON pending payload, got %q (%v)", stdout, err)
	}

	if payload.Message != emailWaitPendingMessage {
		t.Fatalf("expected %q, got %#v", emailWaitPendingMessage, payload)
	}

	if payload.Status != "pending" {
		t.Fatalf("expected pending status, got %#v", payload)
	}

	if _, err := time.Parse(time.RFC3339, payload.After); err != nil {
		t.Fatalf("pending after %q: %v", payload.After, err)
	}

	if payload.SeenIDs == nil {
		t.Fatal("pending seenIds must be present")
	}

	return payload
}

func TestRunEmailWaitPendingReturnsSnapshotBaseline(t *testing.T) {
	withEmailWait(t, 50*time.Millisecond, true)
	useJSONOutput(t)

	latest := time.Date(2026, time.April, 15, 16, 47, 32, 0, time.UTC)
	var err error
	stdout := captureStdout(t, func() {
		err = runEmailWait(context.Background(), "test@scan.escape.tech", func(_ context.Context, _ string, filters *escape.ListInboxEmailsFilters) (*v3.ListInboxEmails200Response, error) {
			if filters.After != nil {
				return &v3.ListInboxEmails200Response{}, nil
			}

			return &v3.ListInboxEmails200Response{
				Data: []v3.ScanEmailSummary{
					*v3.NewScanEmailSummary("id-1", latest, "a@escape.tech", "seen"),
				},
			}, nil
		})
	})
	if err != nil {
		t.Fatalf("expected pending result, got %v", err)
	}

	payload := assertPendingEmail(t, stdout)
	gotAfter, err := time.Parse(time.RFC3339, payload.After)
	if err != nil {
		t.Fatalf("parse after: %v", err)
	}

	if !gotAfter.Equal(latest) {
		t.Fatalf("after = %s, want %s", payload.After, latest.Format(time.RFC3339Nano))
	}

	if len(payload.SeenIDs) != 1 || payload.SeenIDs[0] != "id-1" {
		t.Fatalf("seenIds = %#v", payload.SeenIDs)
	}
}

func TestRunEmailWaitAfterFindsMailThatArrivedBetweenCalls(t *testing.T) {
	withEmailWait(t, time.Minute, true)
	useJSONOutput(t)

	latest := time.Date(2026, time.April, 15, 16, 47, 32, 0, time.UTC)
	newer := latest.Add(time.Second)
	emailsWaitAfter = latest.Format(time.RFC3339Nano)
	emailsWaitSeenIDs = []string{"id-1"}

	prevRead := readInboxEmail
	readInboxEmail = func(_ context.Context, id string) (*v3.ScanEmailDetails, error) {
		if id != "id-9" {
			t.Fatalf("read %s, want id-9", id)
		}

		return v3.NewScanEmailDetails(id, newer, "z@escape.tech", "new", "body"), nil
	}
	t.Cleanup(func() { readInboxEmail = prevRead })

	calls := 0
	var err error
	stdout := captureStdout(t, func() {
		err = runEmailWait(context.Background(), "test@scan.escape.tech", func(_ context.Context, _ string, filters *escape.ListInboxEmailsFilters) (*v3.ListInboxEmails200Response, error) {
			calls++
			if filters.After == nil || !filters.After.Equal(latest) {
				t.Fatalf("expected resume filter after=%s, got %+v", latest, filters.After)
			}

			return &v3.ListInboxEmails200Response{
				Data: []v3.ScanEmailSummary{
					*v3.NewScanEmailSummary("id-1", latest, "a@escape.tech", "already-seen"),
					*v3.NewScanEmailSummary("id-9", newer, "z@escape.tech", "arrived-between-calls"),
				},
			}, nil
		})
	})
	if err != nil {
		t.Fatalf("expected the mail that arrived between calls, got %v", err)
	}

	if calls != 1 {
		t.Fatalf("resume must not snapshot the inbox, list calls = %d", calls)
	}

	var details v3.ScanEmailDetails
	if err := json.Unmarshal([]byte(stdout), &details); err != nil {
		t.Fatalf("stdout %q: %v", stdout, err)
	}

	if details.GetId() != "id-9" {
		t.Fatalf("printed %s, want id-9", details.GetId())
	}
}

// An empty inbox has no server timestamp to resume from. The local clock must
// not become the polling After filter: with a client clock ahead of the
// server, a mail that lands during the wait would be filtered out.
func TestRunEmailWaitEmptyInboxPollsWithoutLocalClockFilter(t *testing.T) {
	withEmailWait(t, time.Minute, true)
	useJSONOutput(t)

	serverCreatedAt := time.Now().UTC().Add(-time.Hour)
	prevRead := readInboxEmail
	readInboxEmail = func(_ context.Context, id string) (*v3.ScanEmailDetails, error) {
		return v3.NewScanEmailDetails(id, serverCreatedAt, "z@escape.tech", "verify", "body"), nil
	}
	t.Cleanup(func() { readInboxEmail = prevRead })

	calls := 0
	var err error
	stdout := captureStdout(t, func() {
		err = runEmailWait(context.Background(), "test@scan.escape.tech", func(_ context.Context, _ string, filters *escape.ListInboxEmailsFilters) (*v3.ListInboxEmails200Response, error) {
			calls++
			if filters.After != nil {
				t.Fatalf("call %d sent after=%s, want no After filter for an empty inbox", calls, filters.After)
			}

			if calls == 1 {
				return &v3.ListInboxEmails200Response{}, nil
			}

			return &v3.ListInboxEmails200Response{
				Data: []v3.ScanEmailSummary{
					*v3.NewScanEmailSummary("id-skewed", serverCreatedAt, "z@escape.tech", "verify"),
				},
			}, nil
		})
	})
	if err != nil {
		t.Fatalf("expected the new mail, got %v", err)
	}

	var details v3.ScanEmailDetails
	if err := json.Unmarshal([]byte(stdout), &details); err != nil {
		t.Fatalf("stdout %q: %v", stdout, err)
	}

	if details.GetId() != "id-skewed" {
		t.Fatalf("printed %s, want id-skewed", details.GetId())
	}
}

// The pending payload of an empty inbox still carries a baseline: the time the
// snapshot ran, not the later timeout.
func TestRunEmailWaitEmptyInboxPendingUsesObservationTime(t *testing.T) {
	withEmailWait(t, 50*time.Millisecond, true)
	useJSONOutput(t)

	before := time.Now().UTC()
	var err error
	stdout := captureStdout(t, func() {
		err = runEmailWait(context.Background(), "test@scan.escape.tech", func(_ context.Context, _ string, filters *escape.ListInboxEmailsFilters) (*v3.ListInboxEmails200Response, error) {
			if filters.After != nil {
				t.Fatalf("empty inbox poll sent after=%s", filters.After)
			}

			return &v3.ListInboxEmails200Response{}, nil
		})
	})
	if err != nil {
		t.Fatalf("expected pending result, got %v", err)
	}

	payload := assertPendingEmail(t, stdout)
	gotAfter, err := time.Parse(time.RFC3339Nano, payload.After)
	if err != nil {
		t.Fatalf("parse after: %v", err)
	}

	if gotAfter.Before(before) || gotAfter.After(before.Add(40*time.Millisecond)) {
		t.Fatalf("after = %s, want the observation time near %s", gotAfter, before)
	}
}
