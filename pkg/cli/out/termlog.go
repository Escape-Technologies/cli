package out

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/sirupsen/logrus"
)

var termlog = "termlog"

// SetupTerminalLog sets up the terminal log
func SetupTerminalLog() {
	if output == outputJSON || output == outputYAML {
		log.AddHook(termlog, func(log log.Entry) {
			// JSON is a valid YAML but multiline readable
			json.NewEncoder(os.Stdout).Encode(log) //nolint:errcheck
		})
	} else {
		log.AddHook(termlog, func(log log.Entry) {
			if log.Level <= logrus.InfoLevel {
				fmt.Printf("[%s] %s\n", log.Level, log.Message)
			}
		})
	}
}

// StopTerminalLog stops the terminal log
func StopTerminalLog() {
	log.RemoveHook(termlog)
}
