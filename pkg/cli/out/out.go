// Package out provides the output formatting for the CLI
package out

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Escape-Technologies/cli/pkg/log"
	"gopkg.in/yaml.v2"
)

type outputT string

const (
	outputPretty outputT = "pretty"
	outputJSON   outputT = "json"
	outputYAML   outputT = "yaml"
)

var output = outputPretty

func pprint(o outputT, data any, pretty string) {
	switch o {
	case outputJSON:
		json.NewEncoder(os.Stdout).Encode(data) //nolint:errcheck
	case outputYAML:
		yaml.NewEncoder(os.Stdout).Encode(data) //nolint:errcheck
	case outputPretty:
		fmt.Println(pretty)
	}
}

// Print prints the data in the output format
func Print(data any, pretty string) {
	pprint(output, data, pretty)
}

func getOutput(o string) *outputT {
	var res outputT
	switch o {
	case "", "pretty":
		res = outputPretty
	case "json", "jsonl":
		res = outputJSON
	case "yaml", "yml":
		res = outputYAML
	}
	return &res
}

// SetOutput sets the output format for the CLI
func SetOutput(o string) error {
	out := getOutput(o)
	if out == nil {
		return fmt.Errorf("invalid output format: %s", o)
	}
	output = *out
	log.Trace("Output format set to %s", output)
	return nil
}
