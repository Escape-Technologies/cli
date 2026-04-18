// Package escape provides the API client for the Escape Platform
package escape

import (
	"encoding/json"
	"errors"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// apiErrorBody mirrors the {message, details} shape returned by the public API
// for both validation failures (`Bad Request`) and pagination errors
// (`Invalid cursor`). Captured here so we can humanize errors without leaking
// internal anyOf wrapper structs from the generated client to callers.
type apiErrorBody struct {
	Message string `json:"message"`
	Details string `json:"details"`
}

// humanizeAPIError attaches the human-readable {message, details} payload to
// errors returned by the generated v3 client. The OpenAPI generator buries the
// useful fields inside an anyOf wrapper struct (e.g. ListProfiles400Response
// holds two ListProfiles400ResponseAnyOf* pointers), and the default
// formatErrorMessage only inspects RFC-7807 Title/Detail. We re-parse the raw
// body so callers see e.g. `Bad Request: severities.0: Invalid enum value`
// instead of the bare HTTP status.
func humanizeAPIError(err error) error {
	if err == nil {
		return nil
	}
	var apiErr v3.GenericOpenAPIError
	if !errors.As(err, &apiErr) {
		return err
	}
	if humanized := humanizeAPIErrorBody(apiErr.Body()); humanized != nil {
		return humanized
	}
	return err
}

// humanizeAPIErrorBody parses a public-API error response body and returns the
// formatted error string, or nil when the body cannot be interpreted as the
// shared {message, details} contract. Extracted from humanizeAPIError so it
// can be tested without constructing the generated GenericOpenAPIError type
// (whose fields are unexported).
func humanizeAPIErrorBody(body []byte) error {
	if len(body) == 0 {
		return nil
	}
	var parsed apiErrorBody
	if jsonErr := json.Unmarshal(body, &parsed); jsonErr != nil {
		return nil
	}
	if parsed.Message == "" && parsed.Details == "" {
		return nil
	}
	if parsed.Details == "" {
		return errors.New(parsed.Message)
	}
	if parsed.Message == "" {
		return errors.New(parsed.Details)
	}
	return errors.New(parsed.Message + ": " + parsed.Details)
}
