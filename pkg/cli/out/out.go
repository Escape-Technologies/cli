// Package out provides the output formatting for the CLI
package out

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Escape-Technologies/cli/pkg/cli/schema"
	"github.com/Escape-Technologies/cli/pkg/log"
	"gopkg.in/yaml.v2"
)

type outputT string

const (
	outputPretty outputT = "pretty"
	outputJSON   outputT = "json"
	outputYAML   outputT = "yaml"
	outputSchema outputT = "schema"
)

var output = outputPretty

// jsonDocumentEmitted is set when JSON mode has already written its one
// document. A later Log is then a status line on stderr. Log itself writes
// the document when nothing else has, so message-only commands still print
// {"msg":"..."}.
var jsonDocumentEmitted bool

// IsSchemaMode returns true if the output mode is schema
func IsSchemaMode() bool {
	return output == outputSchema
}

// IsJSON reports whether stdout is reserved for a single JSON document.
func IsJSON() bool {
	return output == outputJSON
}

// IsPretty reports whether stdout is the human table.
func IsPretty() bool {
	return output == outputPretty
}

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

// Print prints the data in the output format.
// In JSON mode this is the command's one stdout document.
func Print(data any, pretty string) {
	pprint(output, data, pretty)
	noteJSONDocument()
}

func noteJSONDocument() {
	if output == outputJSON {
		jsonDocumentEmitted = true
	}
}

// Message is the JSON payload for commands that only report success
// (for example `scans cancel`). They declare it as their output schema and
// Print it so structuredContent stays one validated document.
type Message struct {
	Msg string `json:"msg"`
}

// Log prints a human-readable status line.
// JSON mode keeps a single stdout document. When Print or Table already
// wrote that document, the line goes to stderr. When Log is the only output,
// it prints {"msg":"..."} so `command -o json | jq` still works.
func Log(pretty string) {
	if output == outputJSON && jsonDocumentEmitted {
		fmt.Fprintln(os.Stderr, pretty)
		return
	}

	Print(Message{Msg: pretty}, pretty)
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
	case "schema":
		res = outputSchema
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
	jsonDocumentEmitted = false
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

// Schema prints the JSON Schema for the given type and returns true if in schema mode
// Commands should call this at the start and return early if it returns true
func Schema(v any) bool {
	if output != outputSchema {
		return false
	}

	_ = schema.Print(v)

	return true
}

// inputSchemaRequested tracks if --input-schema was requested
var inputSchemaRequested bool

// SetInputSchema sets the input schema flag
func SetInputSchema(v bool) {
	inputSchemaRequested = v
}

// InputSchema prints the JSON Schema for the input type and returns true if input-schema was requested
// Commands that accept stdin input should call this to show their expected input format
func InputSchema(v any) bool {
	if !inputSchemaRequested {
		return false
	}

	_ = schema.Print(v)

	return true
}
