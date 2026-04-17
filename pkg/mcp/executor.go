package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	maxStdoutBytes        = 4 << 20
	maxStderrBytes        = 1 << 20
	truncatedOutputSuffix = "\n...[truncated]"
)

// forwardedParentEnvPrefixes lists the minimal parent env the child CLI keeps.
// Request-scoped Escape auth/config values are appended explicitly below.
var forwardedParentEnvPrefixes = []string{
	"TMPDIR=",
}

// ExecutionOptions carries the request-scoped inputs ExecuteCLICommand needs
// to spawn one CLI subprocess.
type ExecutionOptions struct {
	Command        []string
	DisplayCommand []string
	Body           any
	Auth           Auth
	PublicAPIURL   string
}

// ExecutionResult is the captured stdout/stderr/exit-code plus the parsed
// JSON payload (when the CLI produced valid JSON on stdout).
type ExecutionResult struct {
	Stdout          string
	Stderr          string
	StdoutTruncated bool
	StderrTruncated bool
	ExitCode        int
	Payload         any
}

// ExecuteCLICommand spawns the current escape-cli binary with the supplied
// command + arguments, forwards authentication through a sanitized environment,
// pipes the optional request body to stdin, and returns the structured result.
// The caller is responsible for binding a timeout on ctx.
func ExecuteCLICommand(ctx context.Context, options ExecutionOptions) (*ExecutionResult, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve executable path: %w", err)
	}

	commandArgs := append([]string{}, options.Command...)
	commandArgs = append(commandArgs, "--output", "json")

	cmd := exec.CommandContext(ctx, executablePath, commandArgs...)
	cmd.Env = buildCommandEnv(options)

	if options.Body != nil {
		body, err := marshalBody(options.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode stdin body: %w", err)
		}
		cmd.Stdin = bytes.NewReader(body)
	}

	stdout := newCappedBuffer(maxStdoutBytes)
	stderr := newCappedBuffer(maxStderrBytes)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	runErr := cmd.Run()
	result := &ExecutionResult{
		Stdout:          strings.TrimSpace(stdout.Text()),
		Stderr:          strings.TrimSpace(stderr.Text()),
		StdoutTruncated: stdout.Truncated(),
		StderrTruncated: stderr.Truncated(),
	}

	if runErr == nil {
		if result.Stdout != "" {
			var payload any
			if err := json.Unmarshal(stdout.Bytes(), &payload); err == nil {
				result.Payload = payload
			}
		}
		return result, nil
	}

	commandLabel := describeCommand(options)
	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
		return result, fmt.Errorf("command %q failed with exit code %d", commandLabel, result.ExitCode)
	}

	return result, fmt.Errorf("command %q failed: %w", commandLabel, runErr)
}

// buildCommandEnv builds the subprocess env from a strict parent allowlist plus
// the request-scoped Escape vars needed by the child CLI.
func buildCommandEnv(options ExecutionOptions) []string {
	parentEnv := os.Environ()
	env := make([]string, 0, len(forwardedParentEnvPrefixes)+4)
	for _, entry := range parentEnv {
		if hasAnyPrefix(entry, forwardedParentEnvPrefixes) {
			env = append(env, entry)
		}
	}
	env = append(env, "ESCAPE_COLOR_DISABLED=true")

	if options.PublicAPIURL != "" {
		env = append(env, "ESCAPE_API_URL="+options.PublicAPIURL)
	}
	if options.Auth.APIKey != "" {
		env = append(env, "ESCAPE_API_KEY="+options.Auth.APIKey)
	}
	if options.Auth.Authorization != "" {
		env = append(env, "ESCAPE_AUTHORIZATION="+options.Auth.Authorization)
	}

	return env
}

func hasAnyPrefix(entry string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(entry, prefix) {
			return true
		}
	}
	return false
}

func marshalBody(body any) ([]byte, error) {
	if raw, ok := body.([]byte); ok {
		return raw, nil
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}
	return encoded, nil
}

func describeCommand(options ExecutionOptions) string {
	switch {
	case len(options.DisplayCommand) > 0:
		return strings.Join(options.DisplayCommand, " ")
	case len(options.Command) > 0:
		return options.Command[0]
	default:
		return "command"
	}
}

type cappedBuffer struct {
	buf       bytes.Buffer
	limit     int
	truncated bool
}

func newCappedBuffer(limit int) *cappedBuffer {
	return &cappedBuffer{limit: limit}
}

func (buffer *cappedBuffer) Write(data []byte) (int, error) {
	written := len(data)
	if buffer.limit <= 0 {
		buffer.truncated = buffer.truncated || written > 0
		return written, nil
	}

	remaining := buffer.limit - buffer.buf.Len()
	if remaining <= 0 {
		buffer.truncated = buffer.truncated || written > 0
		return written, nil
	}

	if written > remaining {
		buffer.truncated = true
		data = data[:remaining]
	}

	_, err := buffer.buf.Write(data)
	if err != nil {
		return 0, err
	}

	return written, nil
}

func (buffer *cappedBuffer) Bytes() []byte {
	return buffer.buf.Bytes()
}

func (buffer *cappedBuffer) Text() string {
	if !buffer.truncated {
		return buffer.buf.String()
	}

	return buffer.buf.String() + truncatedOutputSuffix
}

func (buffer *cappedBuffer) Truncated() bool {
	return buffer.truncated
}
