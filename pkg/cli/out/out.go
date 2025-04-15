package out

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type outputT string

const (
	outputPretty outputT = "pretty"
	outputJSON   outputT = "json"
	outputYAML   outputT = "yaml"
)

var output outputT = outputPretty

func print(o outputT, data any, pretty string) {
	switch o {
	case outputJSON:
		json.NewEncoder(os.Stdout).Encode(data)
	case outputYAML:
		yaml.NewEncoder(os.Stdout).Encode(data)
	default:
		fmt.Println(pretty)
	}
}

func Print(data any, pretty string) {
	print(output, data, pretty)
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

func SetOutput(o string) error {
	out := getOutput(o)
	if out == nil {
		return fmt.Errorf("invalid output format: %s", o)
	}
	output = *out
	log.Trace("Output format set to %s", output)
	return nil
}

var termlog = "termlog"

func SetupTerminalLog() {
	switch output {
	case outputJSON:
		log.AddHook(termlog, func(log log.LogItem) {
			json.NewEncoder(os.Stdout).Encode(log)
		})
	case outputYAML:
		log.AddHook(termlog, func(log log.LogItem) {
			// JSON is a valid YAML but multiline readable
			json.NewEncoder(os.Stdout).Encode(log)
		})
	default:
		log.AddHook(termlog, func(log log.LogItem) {
			if log.Level <= logrus.InfoLevel {
				fmt.Printf("[%s] %s\n", log.Level, log.Message)
			}
		})
	}
}

func StopTerminalLog() {
	log.RemoveHook(termlog)
}
