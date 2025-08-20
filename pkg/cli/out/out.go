// Package out provides the output formatting for the CLI
package out

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

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
	toPrint := pretty
	if o == outputJSON { //nolint:staticcheck
		buf := bytes.NewBuffer(nil)
		json.NewEncoder(buf).Encode(data) //nolint:errcheck
		toPrint = buf.String()
	} else if o == outputYAML {
		buf := bytes.NewBuffer(nil)
		yaml.NewEncoder(buf).Encode(data) //nolint:errcheck
		toPrint = buf.String()
	}
	fmt.Println(toPrint)
}

// Print prints the data in the output format
func Print(data any, pretty string) {
	pprint(output, data, pretty)
}

// Log prints the data in the output format
func Log(pretty string) {
	Print(struct {
		Msg string `json:"msg"`
	}{
		Msg: pretty,
	}, pretty)
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

// GetShortDate returns the short date format of the given date
func GetShortDate(date string) string {
	parsed, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return date
	}
	return parsed.Format("2006-01-02")
}
