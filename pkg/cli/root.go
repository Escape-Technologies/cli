// Package cli run the cli
package cli

import (
	"context"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/cli/cmd"
)

// Run the CLI
func Run(ctx context.Context) error {
	err := cmd.Execute(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}
	return nil
}
