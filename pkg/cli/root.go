package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type outputT string

const (
	outputPretty outputT = "pretty"
	outputJSON   outputT = "json"
	outputYAML   outputT = "yaml"
)

var output outputT = outputPretty
var verbose bool

func print(data any, pretty func()) error {
	switch output {
	case outputJSON:
		return json.NewEncoder(os.Stdout).Encode(data)
	case outputYAML:
		return yaml.NewEncoder(os.Stdout).Encode(data)
	default:
		pretty()
	}
	return nil
}

func setupTerminalLog() {
	if verbose {
		// Already logging to terminal
		return
	}

	switch output {
	case outputJSON:
		log.AddHook("termlog", func(log log.LogItem) {
			json.NewEncoder(os.Stdout).Encode(log)
		})
	case outputYAML:
		log.AddHook("termlog", func(log log.LogItem) {
			// JSON is a valid YAML but multiline readable
			json.NewEncoder(os.Stdout).Encode(log)
		})
	default:
		log.AddHook("termlog", func(log log.LogItem) {
			if log.Level <= logrus.InfoLevel {
				fmt.Printf("[%s] %s\n", log.Level, log.Message)
			}
		})
	}
}

func Run() error {
	var outputStr string
	rootCmd := &cobra.Command{
		Use:   "escape-cli",
		Short: "CLI to interact with Escape API",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				log.SetLevel(logrus.TraceLevel)
			}
			switch outputStr {
			case "":
				output = outputPretty
			case "json":
				output = outputJSON
			case "yaml":
				output = outputYAML
			case "yml":
				output = outputYAML
			case "pretty":
				output = outputPretty
			default:
				return fmt.Errorf("invalid output format: %s", outputStr)
			}
			log.Trace("Output format set to %s", output)
			return nil
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			log.Trace("Main cli done, exiting")
		},
	}
	// Flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&outputStr, "output", "o", "pretty", "Output format (pretty|json|yaml)")

	// Version
	rootCmd.AddCommand(versionCmd)

	// Locations
	rootCmd.AddCommand(locationsCmd)
	locationsCmd.AddCommand(locationsListCmd)
	locationsCmd.AddCommand(locationsGetCmd)
	locationsCmd.AddCommand(locationsCreateCmd)
	locationsCmd.AddCommand(locationsStartCmd)

	// Integrations
	rootCmd.AddCommand(integrationsCmd)
	integrationsCmd.AddCommand(integrationsAkamaiCmd)
	integrationsAkamaiCmd.AddCommand(integrationsAkamaiList)
	integrationsCmd.AddCommand(integrationsKubernetesCmd)
	integrationsKubernetesCmd.AddCommand(integrationsKubernetesList)
	integrationsKubernetesCmd.AddCommand(integrationsKubernetesDelete)
	//v2
	integrationsCmd.AddCommand(integrationsListCmd)
	integrationsCmd.AddCommand(integrationsGetCmd)
	integrationsCmd.AddCommand(integrationsCreateCmd)
	
	// Scan
	rootCmd.AddCommand(startScanCmd)

	// Domains
	rootCmd.AddCommand(domainsCmd)
	domainsCmd.AddCommand(domainsList)

	// Subdomains
	rootCmd.AddCommand(subdomainsCmd)
	subdomainsCmd.AddCommand(subdomainsList)

	return rootCmd.Execute()
}
