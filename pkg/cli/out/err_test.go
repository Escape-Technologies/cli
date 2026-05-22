package out

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"unsafe"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/sirupsen/logrus"
)

func TestPrintErrorInvalidAPIKey(t *testing.T) {
	got := captureStdout(t, func() {
		PrintError(invalidAPIKeyError())
	})
	want := "Error:\n  Invalid API Key.\n  Get your key here: https://app.escape.tech/user/profile/\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestPrintErrorInvalidAPIKeyVerbose(t *testing.T) {
	log.SetLevel(logrus.DebugLevel)
	t.Cleanup(func() { log.SetLevel(logrus.WarnLevel) })

	got := captureStdout(t, func() {
		PrintError(invalidAPIKeyError())
	})
	if !strings.Contains(got, "Error:\n  Invalid API Key.\n") {
		t.Fatalf("expected friendly invalid API key message, got %q", got)
	}
	if !strings.Contains(got, "  unable to create location: \n") {
		t.Fatalf("expected verbose error chain, got %q", got)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = originalStdout
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe: %v", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	return buf.String()
}

func invalidAPIKeyError() error {
	apiErr := newTestGenericOpenAPIErrorWithStatus([]byte(`{"message":"Not authorized."}`), "401 Unauthorized")
	return fmt.Errorf("unable to create location: %w", apiErr)
}

func newTestGenericOpenAPIErrorWithStatus(body []byte, status string) *v3.GenericOpenAPIError {
	type genericOpenAPIError struct {
		body  []byte
		error string
		model interface{}
	}
	e := genericOpenAPIError{body: body, error: status}
	return (*v3.GenericOpenAPIError)(unsafe.Pointer(&e))
}

var _ error = &v3.GenericOpenAPIError{}
