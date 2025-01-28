package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Injected by ldflags
var (
	version = "dev"
	commit  = "none"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version of the CLI",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Printing version")
		switch output {
		case outputJSON:
			json.NewEncoder(os.Stdout).Encode(map[string]string{"version": version, "commit": commit})
		case outputYAML:
			yaml.NewEncoder(os.Stdout).Encode(map[string]string{"version": version, "commit": commit})
		default:
			fmt.Println("Version:", version, commit)
		}
		log.Trace("Done")
	},
}
