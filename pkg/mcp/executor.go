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

type ExecutionOptions struct {
	Command      []string
	Body         any
	Auth         Auth
	PublicAPIURL string
}

type ExecutionResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Payload  any
}

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

func buildCommandEnv(options ExecutionOptions) []string {
	env := append([]string{}, os.Environ()...)
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

func marshalBody(body any) ([]byte, error) {
	if raw, ok := body.([]byte); ok {
		return raw, nil
	}

	return json.Marshal(body)
}

func errorAs(err error, target any) bool {
	return err != nil && target != nil && errors.As(err, target)
}
