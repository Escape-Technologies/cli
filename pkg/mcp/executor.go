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
	"time"
)

// commandContext is the process starter. Tests replace it to run a sleeper
// without invoking the CLI binary.
var commandContext = exec.CommandContext

const (
	maxStdoutBytes = 4 << 20
	maxStderrBytes = 1 << 20
	// truncatedOutputSuffix marks a capped stderr buffer. Stdout must not use it:
	// appending text to a JSON payload makes the document invalid.
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
	// Timeout is the budget the caller placed on ctx. The deadline error
	// reports this duration. Zero means defaultToolExecutionTimeout.
	Timeout time.Duration
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
// The caller is responsible for binding a timeout on ctx. When that deadline
// is exceeded, the error names the budget and the get/list tool to poll.
// Models that only see "signal: killed" retry the same call.
func ExecuteCLICommand(ctx context.Context, options ExecutionOptions) (*ExecutionResult, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve executable path: %w", err)
	}

	commandArgs := append([]string{}, options.Command...)
	commandArgs = append(commandArgs, "--output", "json")

	cmd := commandContext(ctx, executablePath, commandArgs...)
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
	stderrText := stderr.Text()
	if stderr.Truncated() {
		stderrText += truncatedOutputSuffix
	}

	result := &ExecutionResult{
		Stdout:          strings.TrimSpace(stdout.Text()),
		Stderr:          strings.TrimSpace(stderrText),
		StdoutTruncated: stdout.Truncated(),
		StderrTruncated: stderr.Truncated(),
	}
	finalizeCapturedOutput(result, stdout)

	if runErr == nil {
		return result, nil
	}

	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
	}

	// Detect the deadline on the context, not on the process error. A killed
	// child surfaces as exit code -1 / "signal: killed", which hides the cause.
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return result, executionTimeoutFailure(options)
	}

	commandLabel := describeCommand(options)
	if exitErr != nil {
		return result, fmt.Errorf("command %q failed with exit code %d", commandLabel, result.ExitCode)
	}

	return result, fmt.Errorf("command %q failed: %w", commandLabel, runErr)
}

// executionTimeoutError is the MCP-facing deadline failure. commandFailureText
// prints it alone so subprocess stderr cannot bury the poll instruction.
type executionTimeoutError struct {
	message string
}

func (e *executionTimeoutError) Error() string {
	return e.message
}

func executionTimeoutFailure(options ExecutionOptions) error {
	label := describeCommand(options)
	tool := toolNameFromLabel(label)
	seconds := timeoutBudgetSeconds(options.Timeout)
	if poll, ok := serverSidePollTool(commandWords(options)); ok {
		return &executionTimeoutError{
			message: fmt.Sprintf(
				"%s timed out after %ds; the operation may still be running server-side. Poll with %s instead of retrying.",
				tool,
				seconds,
				poll,
			),
		}
	}

	return &executionTimeoutError{
		message: fmt.Sprintf(
			"%s timed out after %ds. Retry with a smaller size or narrower filters.",
			tool,
			seconds,
		),
	}
}

func toolNameFromLabel(label string) string {
	name := strings.ReplaceAll(label, " ", "_")
	return strings.ReplaceAll(name, "-", "_")
}

func commandWords(options ExecutionOptions) []string {
	if len(options.DisplayCommand) > 0 {
		return options.DisplayCommand
	}

	return options.Command
}

// minServerSideCommandWords is a verb and a subcommand, as in "scans start".
const minServerSideCommandWords = 2

// serverSidePollTool is the get tool for a command that starts work the server
// keeps running after this process is gone. The match is the verb path, not
// the first word: scans list, scans issues, and scans reasoning are reads, and
// telling the model to poll scans_get sends it after a scan that was never started.
func serverSidePollTool(words []string) (string, bool) {
	if len(words) < minServerSideCommandWords {
		return "", false
	}

	switch words[0] + " " + words[1] {
	case "scans start":
		return "scans_get", true
	case "jobs trigger-export":
		return "jobs_get", true
	case "authentications start":
		return "authentications_get", true
	default:
		return "", false
	}
}

func timeoutBudgetSeconds(budget time.Duration) int {
	if budget <= 0 {
		budget = defaultToolExecutionTimeout
	}

	seconds := int(budget.Round(time.Second) / time.Second)
	if seconds < 1 {
		return 1
	}

	return seconds
}

// buildCommandEnv builds the subprocess env from a strict parent allowlist plus
// the request-scoped Escape vars needed by the child CLI.
func buildCommandEnv(options ExecutionOptions) []string {
	parentEnv := os.Environ()
	env := make([]string, 0, len(forwardedParentEnvPrefixes)+4) //nolint:mnd
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

// finalizeCapturedOutput parses a complete JSON stdout into Payload.
// When the cap was hit, the captured prefix is not JSON: it is replaced with
// a truncation error and Payload is {"truncated": true, "error": ...}.
func finalizeCapturedOutput(result *ExecutionResult, stdout *cappedBuffer) {
	if stdout == nil || !stdout.Truncated() {
		if result.Stdout != "" && stdout != nil {
			var payload any
			if err := json.Unmarshal(stdout.Bytes(), &payload); err == nil {
				result.Payload = payload
			}
		}

		return
	}

	message := truncatedStdoutMessage(stdout.limit)
	result.Stdout = message
	result.StdoutTruncated = true
	result.Payload = map[string]any{
		"truncated": true,
		"error":     message,
	}
}

func truncatedStdoutMessage(limit int) string {
	return fmt.Sprintf(
		"output truncated at %d bytes; the payload was cut before it could be valid JSON. Retry with a smaller request or narrower filters.",
		limit,
	)
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
		return 0, fmt.Errorf("write to buffer: %w", err)
	}

	return written, nil
}

func (buffer *cappedBuffer) Bytes() []byte {
	return buffer.buf.Bytes()
}

func (buffer *cappedBuffer) Text() string {
	return buffer.buf.String()
}

func (buffer *cappedBuffer) Truncated() bool {
	return buffer.truncated
}
