package escape

import "testing"

func TestHumanizeAPIErrorBodyExtractsBadRequest(t *testing.T) {
	t.Parallel()
	got := humanizeAPIErrorBody([]byte(`{"message":"Bad Request","details":"severities.0: Invalid enum value"}`))
	if got == nil {
		t.Fatal("expected error, got nil")
	}
	const want = "Bad Request: severities.0: Invalid enum value"
	if got.Error() != want {
		t.Fatalf("expected %q, got %q", want, got.Error())
	}
}

func TestHumanizeAPIErrorBodyExtractsInvalidCursor(t *testing.T) {
	t.Parallel()
	got := humanizeAPIErrorBody([]byte(`{"message":"Invalid cursor","details":"cursor expired"}`))
	if got == nil {
		t.Fatal("expected error, got nil")
	}
	const want = "Invalid cursor: cursor expired"
	if got.Error() != want {
		t.Fatalf("expected %q, got %q", want, got.Error())
	}
}

func TestHumanizeAPIErrorBodyMessageOnly(t *testing.T) {
	t.Parallel()
	got := humanizeAPIErrorBody([]byte(`{"message":"Not Found"}`))
	if got == nil || got.Error() != "Not Found" {
		t.Fatalf("expected just message, got %v", got)
	}
}

func TestHumanizeAPIErrorBodyDetailsOnly(t *testing.T) {
	t.Parallel()
	got := humanizeAPIErrorBody([]byte(`{"details":"raw details"}`))
	if got == nil || got.Error() != "raw details" {
		t.Fatalf("expected just details, got %v", got)
	}
}

func TestHumanizeAPIErrorBodyReturnsNilOnEmpty(t *testing.T) {
	t.Parallel()
	if got := humanizeAPIErrorBody(nil); got != nil {
		t.Fatalf("expected nil for nil body, got %v", got)
	}
	if got := humanizeAPIErrorBody([]byte("")); got != nil {
		t.Fatalf("expected nil for empty body, got %v", got)
	}
}

func TestHumanizeAPIErrorBodyReturnsNilOnUnparseable(t *testing.T) {
	t.Parallel()
	if got := humanizeAPIErrorBody([]byte("not json")); got != nil {
		t.Fatalf("expected nil for non-json body, got %v", got)
	}
}

func TestHumanizeAPIErrorBodyReturnsNilOnUnknownShape(t *testing.T) {
	t.Parallel()
	if got := humanizeAPIErrorBody([]byte(`{"foo":"bar"}`)); got != nil {
		t.Fatalf("expected nil for unknown body shape, got %v", got)
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

type stubError struct{ msg string }

func (e *stubError) Error() string { return e.msg }
