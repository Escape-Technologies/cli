package mcp

import (
	"strings"
	"testing"
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
	if got := buffer.Text(); got != "hello"+truncatedOutputSuffix {
		t.Fatalf("expected truncated text, got %q", got)
	}
	if !buffer.Truncated() {
		t.Fatal("expected buffer to report truncation")
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
