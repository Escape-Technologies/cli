package escape

import (
	"errors"
	"fmt"
	"testing"
	"unsafe"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func TestHumanizeAPIErrorBodyExtractsBadRequest(t *testing.T) {
	t.Parallel()
	got := humanizeAPIErrorBody([]byte(`{"message":"Bad Request","details":"severities.0: Invalid enum value"}`))
	const want = "Bad Request: severities.0: Invalid enum value"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestHumanizeAPIErrorBodyExtractsInvalidCursor(t *testing.T) {
	t.Parallel()
	got := humanizeAPIErrorBody([]byte(`{"message":"Invalid cursor","details":"cursor expired"}`))
	const want = "Invalid cursor: cursor expired"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestHumanizeAPIErrorBodyMessageOnly(t *testing.T) {
	t.Parallel()
	if got := humanizeAPIErrorBody([]byte(`{"message":"Not Found"}`)); got != "Not Found" {
		t.Fatalf("expected just message, got %q", got)
	}
}

func TestHumanizeAPIErrorBodyDetailsOnly(t *testing.T) {
	t.Parallel()
	if got := humanizeAPIErrorBody([]byte(`{"details":"raw details"}`)); got != "raw details" {
		t.Fatalf("expected just details, got %q", got)
	}
}

func TestHumanizeAPIErrorBodyReturnsEmptyOnEmpty(t *testing.T) {
	t.Parallel()
	if got := humanizeAPIErrorBody(nil); got != "" {
		t.Fatalf("expected empty for nil body, got %q", got)
	}
	if got := humanizeAPIErrorBody([]byte("")); got != "" {
		t.Fatalf("expected empty for empty body, got %q", got)
	}
}

func TestHumanizeAPIErrorBodyReturnsEmptyOnUnparseable(t *testing.T) {
	t.Parallel()
	if got := humanizeAPIErrorBody([]byte("not json")); got != "" {
		t.Fatalf("expected empty for non-json body, got %q", got)
	}
}

func TestHumanizeAPIErrorBodyReturnsEmptyOnUnknownShape(t *testing.T) {
	t.Parallel()
	if got := humanizeAPIErrorBody([]byte(`{"foo":"bar"}`)); got != "" {
		t.Fatalf("expected empty for unknown body shape, got %q", got)
	}
}

func TestHumanizeAPIErrorPassesThroughNonAPIError(t *testing.T) {
	t.Parallel()
	original := &stubError{msg: "boom"}
	if got := humanizeAPIError(original); got != original {
		t.Fatalf("expected pass-through, got %v", got)
	}
}

func TestHumanizeAPIErrorPassesThroughNil(t *testing.T) {
	t.Parallel()
	if got := humanizeAPIError(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

// TestHumanizeAPIErrorHumanizesGeneratedAPIError exercises the full
// errors.As + Body() path against the generated v3.GenericOpenAPIError type
// that the CLI actually receives from the public API client. The body is
// poked in via unsafe because the generated struct's fields are unexported
// and there is no exported constructor.
func TestHumanizeAPIErrorHumanizesGeneratedAPIError(t *testing.T) {
	t.Parallel()
	apiErr := newTestGenericOpenAPIError(
		[]byte(`{"message":"Bad Request","details":"severities.0: Invalid enum value"}`),
	)
	got := humanizeAPIError(apiErr)
	if got == nil {
		t.Fatal("expected humanized error, got nil")
	}
	const want = "Bad Request: severities.0: Invalid enum value"
	if got.Error() != want {
		t.Fatalf("expected %q, got %q", want, got.Error())
	}
}

// TestHumanizeAPIErrorUnwrapsWrappedGeneratedAPIError verifies that when the
// generated pointer error is wrapped (as callers typically do with fmt.Errorf
// + %w), errors.As still reaches it and Body() is parsed.
func TestHumanizeAPIErrorUnwrapsWrappedGeneratedAPIError(t *testing.T) {
	t.Parallel()
	apiErr := newTestGenericOpenAPIError([]byte(`{"message":"Invalid cursor","details":"cursor expired"}`))
	wrapped := fmt.Errorf("api error: %w", apiErr)
	got := humanizeAPIError(wrapped)
	if got == nil {
		t.Fatal("expected humanized error, got nil")
	}
	const want = "Invalid cursor: cursor expired"
	if got.Error() != want {
		t.Fatalf("expected %q, got %q", want, got.Error())
	}
}

// TestHumanizeAPIErrorReturnsOriginalOnUnparseableBody verifies the function
// returns the original generated error (not a humanized one) when the body
// does not match the shared {message, details} contract, so callers still see
// the upstream status information.
func TestHumanizeAPIErrorReturnsOriginalOnUnparseableBody(t *testing.T) {
	t.Parallel()
	apiErr := newTestGenericOpenAPIError([]byte("not-json"))
	if got := humanizeAPIError(apiErr); got != error(apiErr) {
		t.Fatalf("expected pass-through of original error, got %v", got)
	}
}

// TestHumanizeAPIErrorReturnsOriginalOnEmptyBody verifies the empty-body path
// through a real GenericOpenAPIError so Body() is invoked.
func TestHumanizeAPIErrorReturnsOriginalOnEmptyBody(t *testing.T) {
	t.Parallel()
	apiErr := newTestGenericOpenAPIError(nil)
	if got := humanizeAPIError(apiErr); got != error(apiErr) {
		t.Fatalf("expected pass-through of original error, got %v", got)
	}
}

// TestHumanizeAPIErrorPreservesChain ensures that callers like
// extractConflict can still walk the wrapped error and reach the original
// *v3.GenericOpenAPIError. Regression: a previous implementation returned a
// detached errors.New(...), breaking UpsertLocation's conflict recovery.
func TestHumanizeAPIErrorPreservesChain(t *testing.T) {
	t.Parallel()
	apiErr := newTestGenericOpenAPIError([]byte(`{"message":"Bad Request","details":"x"}`))
	wrapped := fmt.Errorf("api error: %w", apiErr)
	humanized := humanizeAPIError(wrapped)
	var unwrapped *v3.GenericOpenAPIError
	if !errors.As(humanized, &unwrapped) || unwrapped != apiErr {
		t.Fatalf("expected errors.As to reach the original *v3.GenericOpenAPIError, got %#v", humanized)
	}
}

// newTestGenericOpenAPIError fabricates a v3.GenericOpenAPIError with the
// given body. The struct is copied from v3 so its unexported fields can be
// populated via unsafe; the layout must mirror the generated definition
// exactly. This is only ever used in tests.
func newTestGenericOpenAPIError(body []byte) *v3.GenericOpenAPIError {
	type genericOpenAPIError struct {
		body  []byte
		error string
		model interface{}
	}
	e := genericOpenAPIError{body: body}
	return (*v3.GenericOpenAPIError)(unsafe.Pointer(&e))
}

type stubError struct{ msg string }

func (e *stubError) Error() string { return e.msg }

// Compile-time assertion that v3.GenericOpenAPIError still implements error
// via a value receiver, so *v3.GenericOpenAPIError is assignable to the
// error interface. If this breaks, the tests above may need updating.
var _ error = v3.GenericOpenAPIError{}
var _ error = &v3.GenericOpenAPIError{}
