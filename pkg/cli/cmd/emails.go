package cmd

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

const (
	inboxPageSize           = 100
	emailWaitPollInterval   = 2 * time.Second
	emailWaitPendingFlag    = "pending-on-timeout"
	emailWaitPendingMessage = "no email yet, call again"
)

type inboxState struct {
	latest time.Time
	ids    map[string]struct{}
	// observedAt is when an empty-inbox snapshot ran. It is only the pending
	// payload's fallback baseline. It is never sent as the polling After
	// filter: a client clock ahead of the server would hide a mail that
	// lands during the wait.
	observedAt time.Time
}

// emailWaitPending is the document a bounded wait prints when no new message
// arrived. after and seenIds are the baseline the next call must pass back:
// a fresh snapshot would treat mail that landed between calls as already seen.
type emailWaitPending struct {
	Status  string   `json:"status"`
	After   string   `json:"after"`
	SeenIDs []string `json:"seenIds"`
	Message string   `json:"message"`
}

type inboxListFn func(context.Context, string, *escape.ListInboxEmailsFilters) (*v3.ListInboxEmails200Response, error)

// readInboxEmail loads one message. Tests replace it; production calls the API.
var readInboxEmail = escape.ReadInboxEmail

var (
	emailsTarget               string
	emailsBefore               string
	emailsAfter                string
	emailsLimit                int
	emailsWaitTimeout          time.Duration
	emailsWaitPollInterval     time.Duration
	emailsWaitPendingOnTimeout bool
	emailsWaitAfter            string
	emailsWaitSeenIDs          []string
)

var emailsCmd = &cobra.Command{
	Use:   "emails",
	Short: "List and read scan inbox emails",
	Long: `Read emails sent to Escape scan inboxes.

Use these commands to inspect inbox traffic for scan addresses and block until the
next new email is received.`,
}

var emailsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List inbox emails for a target scan address",
	Example: `  escape-cli emails list --email test.abc123@scan.escape.tech
  escape-cli emails list --email test.abc123@scan.escape.tech --limit 10
  escape-cli emails list --email test.abc123@scan.escape.tech --after 2026-04-15T16:47:32Z`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]v3.ScanEmailSummary{}) {
			return nil
		}

		if strings.TrimSpace(emailsTarget) == "" {
			return errors.New("--email is required")
		}

		if emailsLimit < 0 {
			return errors.New("--limit must be greater than or equal to 0")
		}

		before, err := parseEmailTimeFlag("--before", emailsBefore)
		if err != nil {
			return err
		}

		after, err := parseEmailTimeFlag("--after", emailsAfter)
		if err != nil {
			return err
		}

		items, err := listInboxEmailSummaries(cmd.Context(), emailsTarget, before, after, emailsLimit, escape.ListInboxEmails)
		if err != nil {
			return fmt.Errorf("unable to list inbox emails: %w", err)
		}

		out.Table(items, func() []string {
			rows := []string{"ID\tCREATED AT\tFROM\tSUBJECT"}
			for _, item := range items {
				rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s",
					item.GetId(),
					item.GetCreatedAt().Format(time.RFC3339),
					item.GetFrom(),
					item.GetSubject(),
				))
			}

			return rows
		})

		return nil
	},
}

var emailsReadCmd = &cobra.Command{
	Use:     "read email-id",
	Aliases: []string{"get", "show"},
	Short:   "Read a full inbox email by ID",
	Example: `  escape-cli emails read 1126f550-e77b-49e8-9952-e74ac3014825
  escape-cli emails read 1126f550-e77b-49e8-9952-e74ac3014825 -o json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("email ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.ScanEmailDetails{}) {
			return nil
		}

		item, err := escape.ReadInboxEmail(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to read inbox email: %w", err)
		}

		out.Print(item, prettyInboxEmail(item))

		return nil
	},
}

var emailsWaitCmd = &cobra.Command{
	Use:   "wait",
	Short: "Poll until the next new inbox email is received",
	Example: `  escape-cli emails wait --email test.abc123@scan.escape.tech
  escape-cli emails wait --email test.abc123@scan.escape.tech --timeout 2m
  escape-cli emails wait --email test.abc123@scan.escape.tech --after 2026-04-15T16:47:32Z --seen-id 1126f550-e77b-49e8-9952-e74ac3014825`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.ScanEmailDetails{}) {
			return nil
		}

		return runEmailWait(cmd.Context(), emailsTarget, escape.ListInboxEmails)
	},
}

// runEmailWait polls until a new inbox message arrives, the wait context ends,
// or emailsWaitTimeout elapses. list is injectable so tests can force a deadline
// without calling the public API.
//
// Without --after the first call snapshots the newest mail and waits for
// anything newer. A timeout prints that baseline. The next call must pass
// --after (and each seen id) so mail that arrived between the two calls is
// not absorbed into a new snapshot. Re-snapshotting a large inbox also burns
// the whole tool budget before any waiting happens.
func runEmailWait(ctx context.Context, email string, list inboxListFn) error {
	if strings.TrimSpace(email) == "" {
		return errors.New("--email is required")
	}

	if emailsWaitPollInterval <= 0 {
		return errors.New("--poll-interval must be greater than 0")
	}

	resumed, err := emailWaitResumeBaseline()
	if err != nil {
		return err
	}

	waitCtx := ctx
	if emailsWaitTimeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(waitCtx, emailsWaitTimeout)
		defer cancel()
	}

	state := resumed.state
	if !resumed.ok {
		observedAt := time.Now().UTC()
		state, err = snapshotLatestInboxState(waitCtx, email, list)
		// An empty inbox has no createdAt to resume from. Polling keeps no
		// After filter, so any mail counts as new. A timeout hands back the
		// observation time as the next call's --after.
		state.observedAt = observedAt
		if err != nil {
			return failEmailWait(waitCtx, err, "unable to snapshot inbox state", state)
		}
	}

	for {
		next, err := findNextInboxEmail(waitCtx, email, state, list)
		if err != nil {
			return failEmailWait(waitCtx, err, "unable to wait for inbox email", state)
		}

		if next != nil {
			item, err := readInboxEmail(waitCtx, next.GetId())
			if err != nil {
				return failEmailWait(waitCtx, err, "unable to read new inbox email", state)
			}

			out.Print(item, prettyInboxEmail(item))

			return nil
		}

		select {
		case <-waitCtx.Done():
			return finishEmailWait(waitCtx, state)
		case <-time.After(emailsWaitPollInterval):
		}
	}
}

type emailWaitResume struct {
	state inboxState
	ok    bool
}

// emailWaitResumeBaseline is the caller's previous pending result.
// ok is false when --after was omitted and the inbox must be snapshotted.
func emailWaitResumeBaseline() (emailWaitResume, error) {
	if strings.TrimSpace(emailsWaitAfter) == "" {
		return emailWaitResume{}, nil
	}

	parsed, err := parseEmailTimeFlag("--after", emailsWaitAfter)
	if err != nil {
		return emailWaitResume{}, err
	}

	return emailWaitResume{
		ok: true,
		state: inboxState{
			latest: *parsed,
			ids:    seenIDSet(emailsWaitSeenIDs),
		},
	}, nil
}

// failEmailWait prefers the wait deadline over the in-flight API error. The
// deadline is why the call stopped; surfacing a transport error here makes the
// bounded wait look like a crash. state is the baseline already known, so a
// snapshot that dies on a huge inbox still hands the model a cursor.
func failEmailWait(waitCtx context.Context, err error, message string, state inboxState) error {
	if waitCtx.Err() != nil {
		return finishEmailWait(waitCtx, state)
	}

	return fmt.Errorf("%s: %w", message, err)
}

// finishEmailWait maps the end of the wait context to a CLI result.
// MCP sets pending-on-timeout so a deadline is a successful "call again"
// payload. A killed subprocess looks like a crash, and models retry it
// blindly. Human timeouts stay a non-zero error unless that flag is set.
func finishEmailWait(waitCtx context.Context, state inboxState) error {
	if errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
		if emailsWaitPendingOnTimeout {
			if state.latest.IsZero() {
				state.latest = state.observedAt
			}

			if state.latest.IsZero() {
				state.latest = time.Now().UTC()
			}

			out.Print(pendingFromState(state), emailWaitPendingMessage)

			return nil
		}

		return errors.New("timed out waiting for a new email")
	}

	if err := waitCtx.Err(); err != nil {
		return fmt.Errorf("email wait canceled: %w", err)
	}

	return errors.New("timed out waiting for a new email")
}

func pendingFromState(state inboxState) emailWaitPending {
	return emailWaitPending{
		Status:  "pending",
		After:   state.latest.UTC().Format(time.RFC3339Nano),
		SeenIDs: seenIDList(state.ids),
		Message: emailWaitPendingMessage,
	}
}

func seenIDList(ids map[string]struct{}) []string {
	list := make([]string, 0, len(ids))
	for id := range ids {
		list = append(list, id)
	}

	slices.Sort(list)

	return list
}

func seenIDSet(ids []string) map[string]struct{} {
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}

		set[id] = struct{}{}
	}

	return set
}

func parseEmailTimeFlag(flag string, value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("%s must be a valid RFC3339 timestamp: %w", flag, err)
	}

	return &parsed, nil
}

func listInboxEmailSummaries(
	ctx context.Context,
	email string,
	before *time.Time,
	after *time.Time,
	limit int,
	list inboxListFn,
) ([]v3.ScanEmailSummary, error) {
	all := make([]v3.ScanEmailSummary, 0)
	cursor := ""
	pageSize := inboxPageSize
	if limit > 0 && limit < pageSize {
		pageSize = limit
	}

	filters := &escape.ListInboxEmailsFilters{
		Email:  email,
		Before: before,
		After:  after,
		Size:   pageSize,
	}

	for {
		page, err := list(ctx, cursor, filters)
		if err != nil {
			return nil, err
		}

		all = append(all, page.GetData()...)
		if limit > 0 && len(all) >= limit {
			return all[:limit], nil
		}

		cursor = page.GetNextCursor()
		if cursor == "" {
			return all, nil
		}
	}
}

func snapshotLatestInboxState(ctx context.Context, email string, list inboxListFn) (inboxState, error) {
	state := inboxState{
		ids: map[string]struct{}{},
	}
	cursor := ""
	filters := &escape.ListInboxEmailsFilters{
		Email: email,
		Size:  inboxPageSize,
	}

	for {
		page, err := list(ctx, cursor, filters)
		if err != nil {
			return state, err
		}

		for _, item := range page.GetData() {
			createdAt := item.GetCreatedAt()
			if state.latest.IsZero() {
				state.latest = createdAt
			}

			if !createdAt.Equal(state.latest) {
				return state, nil
			}

			state.ids[item.GetId()] = struct{}{}
		}

		cursor = page.GetNextCursor()
		if cursor == "" {
			return state, nil
		}
	}
}

func findNextInboxEmail(
	ctx context.Context,
	email string,
	state inboxState,
	list inboxListFn,
) (*v3.ScanEmailSummary, error) {
	cursor := ""
	filters := &escape.ListInboxEmailsFilters{
		Email: email,
		Size:  inboxPageSize,
	}
	if !state.latest.IsZero() {
		filters.After = &state.latest
	}

	for {
		page, err := list(ctx, cursor, filters)
		if err != nil {
			return nil, err
		}

		for _, item := range page.GetData() {
			if _, seen := state.ids[item.GetId()]; seen {
				continue
			}

			email := item

			return &email, nil
		}

		cursor = page.GetNextCursor()
		if cursor == "" {
			return nil, nil
		}
	}
}

func prettyInboxEmail(item *v3.ScanEmailDetails) string {
	return fmt.Sprintf(
		"ID: %s\nCREATED AT: %s\nFROM: %s\nSUBJECT: %s\n\nBODY:\n%s",
		item.GetId(),
		item.GetCreatedAt().Format(time.RFC3339),
		item.GetFrom(),
		item.GetSubject(),
		item.GetBody(),
	)
}

func init() {
	emailsCmd.AddCommand(emailsListCmd, emailsReadCmd, emailsWaitCmd)

	emailsListCmd.Flags().StringVar(&emailsTarget, "email", "", "target inbox email address (required)")
	emailsListCmd.Flags().StringVar(&emailsBefore, "before", "", "only include emails created before this RFC3339 timestamp")
	emailsListCmd.Flags().StringVar(&emailsAfter, "after", "", "only include emails created after this RFC3339 timestamp")
	emailsListCmd.Flags().IntVar(&emailsLimit, "limit", 0, "limit total number of emails returned")

	emailsWaitCmd.Flags().StringVar(&emailsTarget, "email", "", "target inbox email address (required)")
	emailsWaitCmd.Flags().StringVar(&emailsWaitAfter, "after", "", "RFC3339 baseline from a previous pending result; pass that after value back so mail that arrived between calls is not missed")
	emailsWaitCmd.Flags().StringArrayVar(&emailsWaitSeenIDs, "seen-id", nil, "inbox email id already seen at the after timestamp; repeat for each seenIds entry from the pending result")
	emailsWaitCmd.Flags().DurationVar(&emailsWaitPollInterval, "poll-interval", emailWaitPollInterval, "delay between inbox polls")
	emailsWaitCmd.Flags().DurationVar(&emailsWaitTimeout, "timeout", 0, "maximum time to wait before failing (e.g. 30s, 2m)")
	emailsWaitCmd.Flags().BoolVar(&emailsWaitPendingOnTimeout, emailWaitPendingFlag, false, "on timeout, print a pending result and exit successfully")

	rootCmd.AddCommand(emailsCmd)
}
