package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

const (
	inboxPageSize         = 100
	emailWaitPollInterval = 2 * time.Second
)

type inboxState struct {
	latest time.Time
	ids    map[string]struct{}
}

type inboxListFn func(context.Context, string, *escape.ListInboxEmailsFilters) (*v3.ListInboxEmails200Response, error)

var (
	emailsTarget           string
	emailsBefore           string
	emailsAfter            string
	emailsLimit            int
	emailsWaitTimeout      time.Duration
	emailsWaitPollInterval time.Duration
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
	Example: `  escape-cli emails list --email test.00000000-0000-0000-0000-000000000000@scan.escape.tech
  escape-cli emails list --email test.00000000-0000-0000-0000-000000000000@scan.escape.tech --limit 10
  escape-cli emails list --email test.00000000-0000-0000-0000-000000000000@scan.escape.tech --after 2026-04-15T16:47:32Z`,
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
	Example: `  escape-cli emails wait --email test.00000000-0000-0000-0000-000000000000@scan.escape.tech
  escape-cli emails wait --email test.00000000-0000-0000-0000-000000000000@scan.escape.tech --timeout 2m
  escape-cli emails wait --email test.00000000-0000-0000-0000-000000000000@scan.escape.tech --poll-interval 1s`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.ScanEmailDetails{}) {
			return nil
		}
		if strings.TrimSpace(emailsTarget) == "" {
			return errors.New("--email is required")
		}
		if emailsWaitPollInterval <= 0 {
			return errors.New("--poll-interval must be greater than 0")
		}

		waitCtx := cmd.Context()
		if emailsWaitTimeout > 0 {
			var cancel context.CancelFunc
			waitCtx, cancel = context.WithTimeout(waitCtx, emailsWaitTimeout)
			defer cancel()
		}

		state, err := snapshotLatestInboxState(waitCtx, emailsTarget, escape.ListInboxEmails)
		if err != nil {
			return fmt.Errorf("unable to snapshot inbox state: %w", err)
		}

		for {
			next, err := findNextInboxEmail(waitCtx, emailsTarget, state, escape.ListInboxEmails)
			if err != nil {
				return fmt.Errorf("unable to wait for inbox email: %w", err)
			}
			if next != nil {
				item, err := escape.ReadInboxEmail(waitCtx, next.GetId())
				if err != nil {
					return fmt.Errorf("unable to read new inbox email: %w", err)
				}
				out.Print(item, prettyInboxEmail(item))
				return nil
			}

			select {
			case <-waitCtx.Done():
				if errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
					return errors.New("timed out waiting for a new email")
				}
				return fmt.Errorf("email wait canceled: %w", waitCtx.Err())
			case <-time.After(emailsWaitPollInterval):
			}
		}
	},
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
	emailsWaitCmd.Flags().DurationVar(&emailsWaitPollInterval, "poll-interval", emailWaitPollInterval, "delay between inbox polls")
	emailsWaitCmd.Flags().DurationVar(&emailsWaitTimeout, "timeout", 0, "maximum time to wait before failing (e.g. 30s, 2m)")

	rootCmd.AddCommand(emailsCmd)
}
