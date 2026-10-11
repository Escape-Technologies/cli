package cmd

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/spf13/cobra"
)

const validationID = "00000000-0000-0000-0000-0000000000e0"

func validationWithStatus(status string, events int) *v3.GetValidation200Response {
	eventList := make([]v3.ValidationEvent, 0, events)
	for range events {
		eventList = append(eventList, *v3.NewValidationEvent(
			"2026-10-01T10:00:05Z",
			v3.ENUMPROPERTIESEVENTSITEMSPROPERTIESLEVEL("INFO"),
			v3.ENUMPROPERTIESEVENTSITEMSPROPERTIESSTAGE("AGENT_REASONING"),
			"Reaching the asset",
			"GET https://api.example.com",
		))
	}

	return v3.NewGetValidation200Response(
		validationID,
		v3.ENUMPROPERTIESSTATUS(status),
		0.5,
		"2026-10-01T10:00:00Z",
		eventList,
	)
}

// watchValidationFunc builds a status source answering calls in order, and
// fails the test when the wait polls more often than the fixture allows.
func watchValidationFunc(t *testing.T, statuses ...string) validationStatusFn {
	t.Helper()

	calls := 0

	return func(_ context.Context, _ string) (*v3.GetValidation200Response, error) {
		if calls >= len(statuses) {
			t.Fatalf("wait polled %d times, fixture has %d statuses", calls, len(statuses))
		}

		status := statuses[calls]
		calls++

		return validationWithStatus(status, 0), nil
	}
}

func resetValidationWaitFlags(t *testing.T) {
	t.Helper()

	previousWait, previousTimeout, previousInterval := validationsWait, validationsWaitTimeout, validationPollInterval
	t.Cleanup(func() {
		validationsWait, validationsWaitTimeout, validationPollInterval = previousWait, previousTimeout, previousInterval
	})
	validationsWait = false
	validationsWaitTimeout = time.Minute
	validationPollInterval = time.Millisecond
}

func TestRunValidationWaitEndsOnFinished(t *testing.T) {
	resetValidationWaitFlags(t)

	command := &cobra.Command{Use: "validations get"}
	command.SetContext(context.Background())

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return runValidationWait(command, validationID, watchValidationFunc(t, "STARTING", "RUNNING", "FINISHED"))
	})
	if err != nil {
		t.Fatalf("wait ended with error: %v\nstderr: %s", err, stderr)
	}

	var validation v3.GetValidation200Response
	mustUnmarshalOne(t, stdout, &validation)
	if validation.GetStatus() != "FINISHED" {
		t.Fatalf("final status = %s", validation.GetStatus())
	}
}

func TestRunValidationWaitFailsOnAFailedValidation(t *testing.T) {
	resetValidationWaitFlags(t)

	command := &cobra.Command{Use: "validations get"}
	command.SetContext(context.Background())

	err := runValidationWait(command, validationID, watchValidationFunc(t, "FAILED"))
	if err == nil || !strings.Contains(err.Error(), "FAILED") {
		t.Fatalf("err = %v", err)
	}
}

func TestRunValidationWaitStopsOnTimeout(t *testing.T) {
	resetValidationWaitFlags(t)
	validationsWaitTimeout = 10 * time.Millisecond

	command := &cobra.Command{Use: "validations get"}
	command.SetContext(context.Background())

	stuck := func(_ context.Context, _ string) (*v3.GetValidation200Response, error) {
		return validationWithStatus("RUNNING", 0), nil
	}

	err := runValidationWait(command, validationID, stuck)
	if err == nil || !strings.Contains(err.Error(), "still RUNNING") {
		t.Fatalf("err = %v", err)
	}
}

func TestRunValidationWaitReportsACancellation(t *testing.T) {
	resetValidationWaitFlags(t)
	validationsWaitTimeout = time.Hour

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	command := &cobra.Command{Use: "validations get"}
	command.SetContext(ctx)

	stuck := func(_ context.Context, _ string) (*v3.GetValidation200Response, error) {
		return validationWithStatus("RUNNING", 0), nil
	}

	err := runValidationWait(command, validationID, stuck)
	if err == nil || !strings.Contains(err.Error(), "canceled") || strings.Contains(err.Error(), "still RUNNING") {
		t.Fatalf("err = %v", err)
	}
}

// TestValidationsStartPrintsDeclaredDocument checks the start command forwards
// the piped body and returns the validation handle.
func TestValidationsStartPrintsDeclaredDocument(t *testing.T) {
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/validations" {
			http.NotFound(w, r)

			return
		}

		writeJSON(t, w, map[string]any{"id": validationID, "status": "RUNNING"})
	})

	resetValidationWaitFlags(t)
	validationsStartCmd.SetContext(context.Background())
	validationsStartCmd.SetIn(strings.NewReader(`{"asset":{"asset_class":"HOST","asset_type":"DNS","address":"api.example.com"}}`))

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return validationsStartCmd.RunE(validationsStartCmd, nil)
	})
	if err != nil {
		t.Fatalf("start: %v\nstderr: %s", err, stderr)
	}

	var started v3.StartValidation200Response
	mustUnmarshalOne(t, stdout, &started)
	if started.GetId() != validationID || started.GetStatus() != "RUNNING" {
		t.Fatalf("started = %#v", started)
	}
}

func TestValidationsStartRejectsAnEmptyBody(t *testing.T) {
	serveJSON(t, func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no request must reach the API when the body is missing")
	})

	resetValidationWaitFlags(t)
	validationsStartCmd.SetContext(context.Background())
	validationsStartCmd.SetIn(strings.NewReader("   "))

	err := validationsStartCmd.RunE(validationsStartCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "no input provided") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidationsStartRejectsANegativeTimeout(t *testing.T) {
	serveJSON(t, func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no request must reach the API when the timeout is invalid")
	})

	resetValidationWaitFlags(t)
	validationsWaitTimeout = -time.Second
	validationsStartCmd.SetContext(context.Background())

	err := validationsStartCmd.RunE(validationsStartCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "must not be negative") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidationsGetPrintsDeclaredDocument(t *testing.T) {
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/validations/"+validationID {
			http.NotFound(w, r)

			return
		}

		writeJSON(t, w, map[string]any{
			"id":            validationID,
			"status":        "FINISHED",
			"progressRatio": 1,
			"createdAt":     "2026-10-01T10:00:00Z",
			"events": []map[string]any{{
				"createdAt":   "2026-10-01T10:00:05Z",
				"level":       "INFO",
				"stage":       "AGENT_REASONING",
				"title":       "Reaching the asset",
				"description": "GET https://api.example.com",
			}},
		})
	})

	validationsGetCmd.SetContext(context.Background())

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return validationsGetCmd.RunE(validationsGetCmd, []string{validationID})
	})
	if err != nil {
		t.Fatalf("get: %v\nstderr: %s", err, stderr)
	}

	var validation v3.GetValidation200Response
	mustUnmarshalOne(t, stdout, &validation)
	if validation.GetStatus() != "FINISHED" || len(validation.GetEvents()) != 1 {
		t.Fatalf("validation = %#v", validation)
	}

	if validation.GetEvents()[0].GetTitle() != "Reaching the asset" {
		t.Fatalf("event = %#v", validation.GetEvents()[0])
	}
}

// TestValidationsWaitIsHiddenFromMCP guards the blocking flags: a hosted tool
// call cannot outlive its budget, so the model polls validations get.
func TestValidationsWaitIsHiddenFromMCP(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	spec := findToolSpecByName(specs, "validations_start")
	if spec == nil {
		t.Fatal("missing validations_start")
	}

	for _, name := range []string{"wait", "timeout"} {
		for _, binding := range spec.FlagBindings {
			if binding.FlagName == name {
				t.Fatalf("validations_start still binds --%s", name)
			}
		}
	}
}
