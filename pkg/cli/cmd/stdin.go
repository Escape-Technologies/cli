package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// readPipedStdin reads a JSON body from r when the caller is not a terminal.
// A terminal returns (nil, nil) so flag-only commands do not block. Any other
// reader, including a pipe from the MCP executor, is read in full. Empty
// input returns (nil, nil).
func readPipedStdin(r io.Reader) ([]byte, error) {
	if r == nil {
		return nil, nil
	}

	if file, ok := r.(*os.File); ok {
		stat, err := file.Stat()
		if err != nil {
			return nil, fmt.Errorf("stat stdin: %w", err)
		}

		if stat.Mode()&os.ModeCharDevice != 0 {
			return nil, nil
		}
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}

	return data, nil
}
