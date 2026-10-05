package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestBuildCommandEnvUsesStrictAllowlist(t *testing.T) {
	t.Setenv("TMPDIR", "/tmp/escape")
	t.Setenv("PATH", "/bin")
	t.Setenv("HOME", "/home/tester")
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("SSL_CERT_FILE", "/tmp/certs.pem")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "aws-secret")
	t.Setenv("OPENAI_API_KEY", "openai-secret")
	t.Setenv("UNRELATED_VAR", "keep-me")
	t.Setenv("ESCAPE_API_URL", "https://parent.example.com")
	t.Setenv("ESCAPE_API_KEY", "parent-key")
	t.Setenv("ESCAPE_AUTHORIZATION", "Key parent")
	t.Setenv("ESCAPE_FOO", "bar")

	env := envMap(buildCommandEnv(ExecutionOptions{
		PublicAPIURL: "https://request.example.com",
		Auth: Auth{
			APIKey:        "request-key",
			Authorization: "Bearer request-token",
		},
	}))

	if got := env["TMPDIR"]; got != "/tmp/escape" {
		t.Fatalf("expected TMPDIR to be preserved, got %q", got)
	}

	droppedKeys := []string{
		"PATH",
		"HOME",
		"LANG",
		"SSL_CERT_FILE",
		"AWS_SECRET_ACCESS_KEY",
		"OPENAI_API_KEY",
		"UNRELATED_VAR",
		"ESCAPE_FOO",
	}
	for _, key := range droppedKeys {
		if _, ok := env[key]; ok {
			t.Fatalf("expected %s to be stripped, got %#v", key, env)
		}
	}

	if got := env["ESCAPE_COLOR_DISABLED"]; got != "true" {
		t.Fatalf("expected ESCAPE_COLOR_DISABLED=true, got %q", got)
	}

	if got := env["ESCAPE_API_URL"]; got != "https://request.example.com" {
		t.Fatalf("expected request ESCAPE_API_URL, got %q", got)
	}

	if got := env["ESCAPE_API_KEY"]; got != "request-key" {
		t.Fatalf("expected request ESCAPE_API_KEY, got %q", got)
	}

	if got := env["ESCAPE_AUTHORIZATION"]; got != "Bearer request-token" {
		t.Fatalf("expected request ESCAPE_AUTHORIZATION, got %q", got)
	}
}

func TestBuildCommandEnvWithoutRequestAuthDoesNotLeakParent(t *testing.T) {
	t.Setenv("TMPDIR", "/tmp/escape")
	t.Setenv("ESCAPE_API_URL", "https://parent.example.com")
	t.Setenv("ESCAPE_API_KEY", "parent-key")
	t.Setenv("ESCAPE_AUTHORIZATION", "Key parent")
	t.Setenv("ESCAPE_FOO", "bar")

	env := envMap(buildCommandEnv(ExecutionOptions{}))

	if got := env["TMPDIR"]; got != "/tmp/escape" {
		t.Fatalf("expected TMPDIR to be preserved, got %q", got)
	}

	for _, key := range []string{
		"ESCAPE_API_URL",
		"ESCAPE_API_KEY",
		"ESCAPE_AUTHORIZATION",
		"ESCAPE_FOO",
	} {
		if _, ok := env[key]; ok {
			t.Fatalf("expected %s to be stripped when request is empty, got %#v", key, env)
		}
	}
}

func TestCappedBufferTruncatesOutput(t *testing.T) {
	t.Parallel()

	buffer := newCappedBuffer(5)
	written, err := buffer.Write([]byte("hello-world"))
	if err != nil {
		t.Fatalf("expected write to succeed, got %v", err)
	}

	if written != len("hello-world") {
		t.Fatalf("expected write count %d, got %d", len("hello-world"), written)
	}

	if got := string(buffer.Bytes()); got != "hello" {
		t.Fatalf("expected capped bytes, got %q", got)
	}

	if got := buffer.Text(); got != "hello" {
		t.Fatalf("expected raw capped text without a suffix, got %q", got)
	}

	if strings.Contains(buffer.Text(), truncatedOutputSuffix) {
		t.Fatalf("stdout text must stay free of %q", truncatedOutputSuffix)
	}

	if !buffer.Truncated() {
		t.Fatal("expected buffer to report truncation")
	}
}

func TestFinalizeCapturedOutputReplacesTruncatedStdout(t *testing.T) {
	t.Parallel()

	buffer := newCappedBuffer(8)
	if _, err := buffer.Write([]byte(`{"items":[1,2,3]}`)); err != nil {
		t.Fatalf("write: %v", err)
	}

	result := &ExecutionResult{
		Stdout:          strings.TrimSpace(buffer.Text()),
		StdoutTruncated: buffer.Truncated(),
	}
	finalizeCapturedOutput(result, buffer)

	if !result.StdoutTruncated {
		t.Fatal("expected truncated result")
	}

	if strings.Contains(result.Stdout, truncatedOutputSuffix) || strings.Contains(result.Stdout, "items") {
		t.Fatalf("stdout must be a truncation message, got %q", result.Stdout)
	}

	if !strings.Contains(result.Stdout, "8 bytes") || !strings.Contains(result.Stdout, "narrower filters") {
		t.Fatalf("message = %q", result.Stdout)
	}

	if strings.Contains(result.Stdout, "cursor") || strings.Contains(result.Stdout, "--size") {
		t.Fatalf("truncation message must not assume a paged list, got %q", result.Stdout)
	}

	payload, ok := result.Payload.(map[string]any)
	if !ok || payload["truncated"] != true || payload["error"] != result.Stdout {
		t.Fatalf("payload = %#v", result.Payload)
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}

	if decoded["truncated"] != true {
		t.Fatalf("decoded = %#v", decoded)
	}
}

func TestFinalizeCapturedOutputParsesCompleteJSON(t *testing.T) {
	t.Parallel()

	buffer := newCappedBuffer(1024)
	if _, err := buffer.Write([]byte(`{"items":[],"nextCursor":"","totalCount":0}`)); err != nil {
		t.Fatalf("write: %v", err)
	}

	result := &ExecutionResult{Stdout: strings.TrimSpace(buffer.Text())}
	finalizeCapturedOutput(result, buffer)

	if result.StdoutTruncated {
		t.Fatal("complete JSON must not be marked truncated")
	}

	if result.Stdout != `{"items":[],"nextCursor":"","totalCount":0}` {
		t.Fatalf("stdout = %q", result.Stdout)
	}

	payload, ok := result.Payload.(map[string]any)
	if !ok {
		t.Fatalf("payload = %#v", result.Payload)
	}

	items, ok := payload["items"].([]any)
	if !ok || len(items) != 0 || payload["nextCursor"] != "" || payload["totalCount"] != float64(0) {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestExecuteCLICommandRedactsCallerControlledArgs(t *testing.T) {
	_, err := ExecuteCLICommand(t.Context(), ExecutionOptions{
		Command:        []string{"--definitely-not-a-real-cli-command", "secret-value"},
		DisplayCommand: []string{"safe-command"},
	})
	if err == nil {
		t.Fatal("expected command failure")
	}

	if !strings.Contains(err.Error(), `safe-command`) {
		t.Fatalf("expected redacted command label, got %q", err.Error())
	}

	if strings.Contains(err.Error(), "secret-value") {
		t.Fatalf("expected error to hide caller-controlled args, got %q", err.Error())
	}

	if strings.Contains(err.Error(), "--output") {
		t.Fatalf("expected injected wrapper args to stay hidden, got %q", err.Error())
	}
}

func TestExecutionTimeoutFailureNamesPollTool(t *testing.T) {
	t.Parallel()

	seconds := timeoutBudgetSeconds(defaultToolExecutionTimeout)
	running := func(tool, poll string) string {
		return fmt.Sprintf(
			"%s timed out after %ds; the operation may still be running server-side. Poll with %s instead of retrying.",
			tool,
			seconds,
			poll,
		)
	}
	retry := func(tool string) string {
		return fmt.Sprintf(
			"%s timed out after %ds. Retry with a smaller size or narrower filters.",
			tool,
			seconds,
		)
	}
	tests := []struct {
		display []string
		want    string
	}{
		{[]string{"scans", "start"}, running("scans_start", "scans_get")},
		{[]string{"scans", "start", "profile-id"}, running("scans_start_profile_id", "scans_get")},
		{[]string{"jobs", "trigger-export"}, running("jobs_trigger_export", "jobs_get")},
		{[]string{"authentications", "start"}, running("authentications_start", "authentications_get")},
		{[]string{"scans", "list"}, retry("scans_list")},
		{[]string{"scans", "issues"}, retry("scans_issues")},
		{[]string{"scans", "reasoning"}, retry("scans_reasoning")},
		{[]string{"authentications", "get"}, retry("authentications_get")},
		{[]string{"emails", "wait"}, retry("emails_wait")},
		{[]string{"profiles", "list"}, retry("profiles_list")},
	}
	for _, test := range tests {
		err := executionTimeoutFailure(ExecutionOptions{
			DisplayCommand: test.display,
			Timeout:        defaultToolExecutionTimeout,
		})
		if err.Error() != test.want {
			t.Fatalf("expected %q, got %q", test.want, err.Error())
		}
	}
}

func TestTimeoutBudgetSeconds(t *testing.T) {
	t.Parallel()

	if got := timeoutBudgetSeconds(0); got != int(defaultToolExecutionTimeout/time.Second) {
		t.Fatalf("zero budget reported %ds", got)
	}

	if got := timeoutBudgetSeconds(EmailsWaitTimeout); got != int(EmailsWaitTimeout/time.Second) {
		t.Fatalf("emails budget reported %ds", got)
	}

	if got := timeoutBudgetSeconds(time.Millisecond); got != 1 {
		t.Fatalf("sub-second budget reported %ds", got)
	}
}

func TestExecuteCLICommandReportsDeadline(t *testing.T) {
	original := commandContext
	commandContext = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sleep", "30")
	}
	t.Cleanup(func() { commandContext = original })

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := ExecuteCLICommand(ctx, ExecutionOptions{
		Command:        []string{"scans", "start", "secret-profile"},
		DisplayCommand: []string{"scans", "start"},
		Timeout:        defaultToolExecutionTimeout,
	})
	if err == nil {
		t.Fatal("expected deadline error")
	}

	want := executionTimeoutFailure(ExecutionOptions{
		DisplayCommand: []string{"scans", "start"},
		Timeout:        defaultToolExecutionTimeout,
	}).Error()
	if err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}

	if strings.Contains(err.Error(), "secret-profile") || strings.Contains(err.Error(), "signal: killed") || strings.Contains(err.Error(), "exit code") {
		t.Fatalf("deadline error leaked kill details: %q", err.Error())
	}

	var timedOut *executionTimeoutError
	if !errors.As(err, &timedOut) {
		t.Fatalf("expected executionTimeoutError, got %T", err)
	}

	if result == nil {
		t.Fatal("expected a result alongside the deadline error")
	}
}

func TestExecuteCLICommandCancelIsNotADeadline(t *testing.T) {
	original := commandContext
	commandContext = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sleep", "30")
	}
	t.Cleanup(func() { commandContext = original })

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := ExecuteCLICommand(ctx, ExecutionOptions{
		DisplayCommand: []string{"scans", "get"},
		Timeout:        defaultToolExecutionTimeout,
	})
	if err == nil {
		t.Fatal("expected cancel failure")
	}

	if strings.Contains(err.Error(), "timed out after") {
		t.Fatalf("cancel was reported as a deadline: %q", err.Error())
	}
}

func TestExecuteCLICommandSuccessIsNotADeadline(t *testing.T) {
	original := commandContext
	commandContext = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "true")
	}
	t.Cleanup(func() { commandContext = original })

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := ExecuteCLICommand(ctx, ExecutionOptions{
		DisplayCommand: []string{"scans", "get"},
		Timeout:        time.Second,
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func envMap(entries []string) map[string]string {
	env := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}

		env[key] = value
	}

	return env
}
