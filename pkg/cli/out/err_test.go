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
	stdout, stderr := captureOutErr(t, func() {
		PrintError(invalidAPIKeyError())
	})
	if stdout != "" {
		t.Fatalf("error leaked to stdout: %q", stdout)
	}

	want := "Error:\n  Invalid API Key.\n  Get your key here: https://app.escape.tech/user/profile/\n"
	if stderr != want {
		t.Fatalf("expected %q, got %q", want, stderr)
	}
}

func TestPrintErrorInvalidAPIKeyVerbose(t *testing.T) {
	log.SetLevel(logrus.DebugLevel)
	t.Cleanup(func() { log.SetLevel(logrus.WarnLevel) })

	stdout, stderr := captureOutErr(t, func() {
		PrintError(invalidAPIKeyError())
	})
	if stdout != "" {
		t.Fatalf("error leaked to stdout: %q", stdout)
	}

	if !strings.Contains(stderr, "Error:\n  Invalid API Key.\n") {
		t.Fatalf("expected friendly invalid API key message, got %q", stderr)
	}

	if !strings.Contains(stderr, "  unable to create location: \n") {
		t.Fatalf("expected verbose error chain, got %q", stderr)
	}
}

func captureOutErr(t *testing.T, fn func()) (string, string) {
	t.Helper()
	outReader, outWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	errReader, errWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	originalOut := os.Stdout
	originalErr := os.Stderr
	os.Stdout = outWriter
	os.Stderr = errWriter
	defer func() {
		os.Stdout = originalOut
		os.Stderr = originalErr
	}()

	fn()

	if err := outWriter.Close(); err != nil {
		t.Fatalf("failed to close pipe: %v", err)
	}

	if err := errWriter.Close(); err != nil {
		t.Fatalf("failed to close pipe: %v", err)
	}

	var outBuf, errBuf bytes.Buffer
	if _, err := io.Copy(&outBuf, outReader); err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	if _, err := io.Copy(&errBuf, errReader); err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	return outBuf.String(), errBuf.String()
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
