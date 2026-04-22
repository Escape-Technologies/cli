// Package escape provides the API client for the Escape Platform
package escape

import (
	"encoding/json"
	"errors"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// apiErrorBody mirrors the {message, details} shape returned by the public API
// for both validation failures (`Bad Request`) and pagination errors
// (`Invalid cursor`). Captured here so we can humanize errors without coupling
// to the generated client's per-endpoint response structs.
type apiErrorBody struct {
	Message string `json:"message"`
	Details string `json:"details"`
}

// humanizedAPIError pairs a human-readable message parsed from the API
// response body with the original error returned by the generated v3 client.
// Preserving the underlying error in the chain is required so callers like
// extractConflict can still type-assert to *v3.GenericOpenAPIError and reach
// the typed model (e.g. Conflict) through errors.As.
type humanizedAPIError struct {
	msg string
	err error
}

func (e *humanizedAPIError) Error() string { return e.msg }
func (e *humanizedAPIError) Unwrap() error { return e.err }

// humanizeAPIError attaches the human-readable {message, details} payload to
// errors returned by the generated v3 client. The default formatErrorMessage
// only inspects RFC-7807 Title/Detail, so we re-parse the raw body to surface
// e.g. `Bad Request: severities.0: Invalid enum value` instead of the bare
// HTTP status. The returned error wraps the original so the chain stays
// walkable via errors.As / errors.Unwrap.
func humanizeAPIError(err error) error {
	if err == nil {
		return nil
	}
	// The generated v3 client returns *v3.GenericOpenAPIError, so target the
	// pointer type — targeting the value type silently fails to match and
	// drops all humanization (see TestHumanizeAPIErrorHumanizesGeneratedAPIError).
	var apiErr *v3.GenericOpenAPIError
	if !errors.As(err, &apiErr) || apiErr == nil {
		return err
	}
	msg := humanizeAPIErrorBody(apiErr.Body())
	if msg == "" {
		return err
	}
	return &humanizedAPIError{msg: msg, err: err}
}

// humanizeAPIErrorBody parses a public-API error response body and returns the
// formatted error message, or an empty string when the body cannot be
// interpreted as the shared {message, details} contract. Extracted from
// humanizeAPIError so it can be tested without constructing the generated
// GenericOpenAPIError type (whose fields are unexported).
func humanizeAPIErrorBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var parsed apiErrorBody
	if jsonErr := json.Unmarshal(body, &parsed); jsonErr != nil {
		return ""
	}
	if parsed.Message == "" && parsed.Details == "" {
		return ""
	}
	if parsed.Details == "" {
		return parsed.Message
	}
	if parsed.Message == "" {
		return parsed.Details
	}
	return parsed.Message + ": " + parsed.Details
}
