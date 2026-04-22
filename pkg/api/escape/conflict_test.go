package escape

import (
	"errors"
	"fmt"
	"testing"
	"unsafe"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func newTestConflictAPIError(instanceID string) *v3.GenericOpenAPIError {
	type genericOpenAPIError struct {
		body  []byte
		error string
		model interface{}
	}
	e := genericOpenAPIError{
		body:  []byte(`{"message":"Conflict on the following field","field":"name","instanceId":"` + instanceID + `"}`),
		error: "409 Conflict",
		model: v3.Conflict{InstanceId: instanceID, Field: "name"},
	}
	return (*v3.GenericOpenAPIError)(unsafe.Pointer(&e))
}

func TestExtractConflictDirect(t *testing.T) {
	t.Parallel()
	apiErr := newTestConflictAPIError("abc-123")
	id, err := extractConflict(apiErr)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if id != "abc-123" {
		t.Fatalf("expected abc-123, got %q", id)
	}
}

func TestExtractConflictThroughHumanizedWrapper(t *testing.T) {
	t.Parallel()
	apiErr := newTestConflictAPIError("abc-123")
	wrapped := fmt.Errorf("unable to create location: %w", humanizeAPIError(apiErr))
	id, err := extractConflict(wrapped)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if id != "abc-123" {
		t.Fatalf("expected abc-123, got %q", id)
	}
}

func TestExtractConflictReturnsErrorOnNonConflict(t *testing.T) {
	t.Parallel()
	if _, err := extractConflict(errors.New("boom")); err == nil {
		t.Fatal("expected error for non-conflict, got nil")
	}
	if _, err := extractConflict(nil); err == nil {
		t.Fatal("expected error for nil, got nil")
	}
}

func TestExtractConflictRejectsEmptyInstanceID(t *testing.T) {
	t.Parallel()
	apiErr := newTestConflictAPIError("")
	if _, err := extractConflict(apiErr); err == nil {
		t.Fatal("expected error when instanceId is empty, got nil")
	}
}
