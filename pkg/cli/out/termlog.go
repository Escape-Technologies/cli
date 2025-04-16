package out

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/sirupsen/logrus"
)

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
