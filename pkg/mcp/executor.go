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

// strippedParentEnvPrefixes lists env var prefixes that must not leak from the
// MCP server process into the child CLI subprocess. Request-scoped values are
// appended by buildCommandEnv after the strip loop.
var strippedParentEnvPrefixes = []string{
	"ESCAPE_API_URL=",
	"ESCAPE_API_KEY=",
	"ESCAPE_AUTHORIZATION=",
	"ESCAPE_COLOR_DISABLED=",
}

// ExecutionOptions carries the request-scoped inputs ExecuteCLICommand needs
// to spawn one CLI subprocess.
type ExecutionOptions struct {
	Command      []string
	Body         any
	Auth         Auth
	PublicAPIURL string
}

// ExecutionResult is the captured stdout/stderr/exit-code plus the parsed
// JSON payload (when the CLI produced valid JSON on stdout).
type ExecutionResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Payload  any
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

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	result := &ExecutionResult{
		Stdout: strings.TrimSpace(stdout.String()),
		Stderr: strings.TrimSpace(stderr.String()),
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

	var exitErr *exec.ExitError
	if ok := errorAs(runErr, &exitErr); ok {
		result.ExitCode = exitErr.ExitCode()
		return result, fmt.Errorf("command %q failed with exit code %d", strings.Join(commandArgs, " "), result.ExitCode)
	}

	return result, fmt.Errorf("command %q failed: %w", strings.Join(commandArgs, " "), runErr)
}

// buildCommandEnv builds the subprocess env, stripping inherited auth and
// color vars from the parent process so the child CLI only uses values from
// the current request. Prevents stale server-scoped credentials or duplicate
// keys from leaking when a request omits one of them.
func buildCommandEnv(options ExecutionOptions) []string {
	parentEnv := os.Environ()
	env := make([]string, 0, len(parentEnv)+len(strippedParentEnvPrefixes))
	for _, entry := range parentEnv {
		if hasAnyPrefix(entry, strippedParentEnvPrefixes) {
			continue
		}
		env = append(env, entry)
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

func errorAs(err error, target any) bool {
	return err != nil && target != nil && errors.As(err, target)
}
