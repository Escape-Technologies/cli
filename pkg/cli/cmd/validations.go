package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	validationsWait        bool
	validationsWaitTimeout time.Duration
)

// validationPollInterval is the delay between two validation polls. It is a
// variable so a test drives the wait without real sleeps.
var validationPollInterval = 3 * time.Second

// defaultValidationWaitTimeout bounds --wait: a validation that never settles
// must not hang the command.
const defaultValidationWaitTimeout = 10 * time.Minute

// terminalValidationStatuses end a wait. STARTING and RUNNING do not.
var terminalValidationStatuses = map[string]bool{
	"FINISHED": true,
	"FAILED":   true,
	"CANCELED": true,
}

var validationsCmd = &cobra.Command{
	Use:     "validations",
	Aliases: []string{"validation"},
	Short:   "Check that an asset is reachable (Beta)",
	Long: `Validations - Check Asset Reachability

Start a validation to have the platform reach an asset from an Escape or a
private location and report what it found. A validation is asynchronous: it
starts, runs, and ends on a status, and its events narrate each step.

BETA: target validation is listed as coming next in the release notes, not
shipped, so this group can still change.

COMMON WORKFLOWS:
  • Validate an asset and wait for the verdict:
    $ escape-cli validations start < body.json --wait

  • Poll the validation the start returned:
    $ escape-cli validations get <validation-id>`,
}

var validationsStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a validation of an asset (Beta)",
	Long: `Start Validation - Ask the Platform to Reach an Asset

Send the JSON body on stdin: the asset to reach and an optional location
override. --input-schema prints the full body schema.

The asset is a union of the asset types the platform knows, so "asset_type"
selects the fields the body may carry. --wait polls the validation until it
reaches a terminal status and prints that final state.`,
	Example: `  echo '{"asset":{"asset_class":"API_SERVICE","asset_type":"REST","url":"https://api.example.com"}}' | \
    escape-cli validations start

  # The same, from a private location, waiting for the verdict
  echo '{"asset":{"asset_class":"API_SERVICE","asset_type":"REST","url":"https://api.example.com"},"proxyId":"00000000-0000-0000-0000-000000000000"}' | \
    escape-cli validations start --wait`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.InputSchema(v3.StartValidationRequest{}) {
			return nil
		}

		if out.Schema(v3.StartValidation200Response{}) {
			return nil
		}

		if validationsWaitTimeout < 0 {
			return errors.New("--timeout must not be negative (0 disables the timeout)")
		}

		body, err := readRequiredBody(cmd)
		if err != nil {
			return err
		}

		started, err := escape.StartValidation(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("unable to start validation: %w", err)
		}

		// With --wait, JSON and YAML output stays one document: the terminal state
		// printed by the wait. Pretty output still announces the start.
		if !validationsWait || out.IsPretty() {
			out.Print(started, fmt.Sprintf("Validation %s started", started.GetId()))
		}

		if !validationsWait {
			return nil
		}

		return waitValidation(cmd, started.GetId())
	},
}

var validationsGetCmd = &cobra.Command{
	Use:     "get validation-id",
	Aliases: []string{"describe", "show"},
	Short:   "Get the status and events of a validation (Beta)",
	Long: `Get Validation - Read the Verdict and Its Events

Fetch a validation by ID: its status, how far it progressed, the output it
produced, and the chronological events it emitted. This is the poll target of
"validations start".`,
	Example: `  escape-cli validations get 00000000-0000-0000-0000-000000000000
  escape-cli validations get <validation-id> -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.GetValidation200Response{}) {
			return nil
		}

		validation, err := escape.GetValidation(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get validation: %w", err)
		}

		printValidation(validation)

		return nil
	},
}

// validationStatusFn reads one validation. Injectable so a wait test drives
// statuses without the API.
type validationStatusFn func(ctx context.Context, validationID string) (*v3.GetValidation200Response, error)

// waitValidation polls the validation until it ends, the timeout fires, or the
// context is cancelled. It prints the final state once.
func waitValidation(cmd *cobra.Command, validationID string) error {
	return runValidationWait(cmd, validationID, escape.GetValidation)
}

func runValidationWait(cmd *cobra.Command, validationID string, getValidation validationStatusFn) error {
	waitCtx := cmd.Context()
	if validationsWaitTimeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(waitCtx, validationsWaitTimeout)
		defer cancel()
	}

	var last *v3.GetValidation200Response
	for {
		// The deadline can fire while a poll was in flight: catch it before
		// the next one, or a timeout would surface as an API error.
		if waitCtx.Err() != nil {
			return validationWaitStopError(waitCtx, validationID, last)
		}

		validation, err := getValidation(waitCtx, validationID)
		if err != nil {
			if waitCtx.Err() != nil {
				return validationWaitStopError(waitCtx, validationID, last)
			}

			return fmt.Errorf("unable to get validation: %w", err)
		}

		last = validation

		status := string(validation.GetStatus())
		if terminalValidationStatuses[status] {
			printValidation(validation)
			if status != "FINISHED" {
				return fmt.Errorf("validation %s ended with status %s", validationID, status)
			}

			return nil
		}

		select {
		case <-waitCtx.Done():
			return validationWaitStopError(waitCtx, validationID, last)
		case <-time.After(validationPollInterval):
		}
	}
}

// validationWaitStopError tells a timeout apart from an interrupt: only the
// deadline means the validation was still running when the wait ran out.
func validationWaitStopError(ctx context.Context, validationID string, last *v3.GetValidation200Response) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return validationWaitTimeoutError(validationID, last)
	}

	return fmt.Errorf("validation wait canceled: %w", ctx.Err())
}

func validationWaitTimeoutError(validationID string, last *v3.GetValidation200Response) error {
	if last == nil {
		return fmt.Errorf("no validation status after %s: poll 'escape-cli validations get %s'", validationsWaitTimeout, validationID)
	}

	return fmt.Errorf(
		"validation %s is still %s after %s: poll 'escape-cli validations get %s'",
		validationID,
		last.GetStatus(),
		validationsWaitTimeout,
		validationID,
	)
}

func printValidation(validation *v3.GetValidation200Response) {
	out.Table(validation, func() []string {
		return []string{
			"ID\tSTATUS\tPROGRESS\tCREATED AT",
			fmt.Sprintf("%s\t%s\t%.0f%%\t%s",
				validation.GetId(),
				validation.GetStatus(),
				validation.GetProgressRatio()*100, //nolint:mnd
				out.GetShortDate(validation.GetCreatedAt()),
			),
		}
	})

	// JSON and YAML already carry the whole document: the events are a second
	// table for the human reader, so they only print in pretty mode.
	if !out.IsPretty() {
		return
	}

	if events := validation.GetEvents(); len(events) > 0 {
		out.Table(events, func() []string {
			rows := []string{"CREATED AT\tLEVEL\tSTAGE\tTITLE"}
			for _, event := range events {
				rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s",
					event.GetCreatedAt(),
					event.GetLevel(),
					event.GetStage(),
					event.GetTitle(),
				))
			}

			return rows
		})
	}
}

func init() {
	validationsStartCmd.Flags().BoolVarP(&validationsWait, "wait", "w", false, "poll the validation until it reaches a terminal status")
	validationsStartCmd.Flags().DurationVar(&validationsWaitTimeout, "timeout", defaultValidationWaitTimeout, "maximum time to wait before failing (e.g. 5m; 0 disables the timeout)")
	// wait blocks until the remote validation finishes, which a hosted tool call
	// cannot guarantee. The model polls validations get instead.
	markMCPSkip(validationsStartCmd.Flags(), "wait", "timeout")

	validationsCmd.AddCommand(validationsStartCmd, validationsGetCmd)
	rootCmd.AddCommand(validationsCmd)
}
