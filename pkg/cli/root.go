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

func Run() error {
	var verbose bool
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
	locationsCmd.AddCommand(locationsDeleteCmd)
	locationsCmd.AddCommand(locationsGetCmd)
	locationsCreateCmd.Flags().BoolVar(&locationsCreateInput.PrivateLocationV2, "new", false, "Opt in to the new private location feature.")
	locationsCmd.AddCommand(locationsCreateCmd)
	locationsUpsertCmd.Flags().BoolVar(&locationsUpsertInput.PrivateLocationV2, "new", false, "Opt in to the new private location feature.")
	locationsCmd.AddCommand(locationsUpsertCmd)
	locationsCmd.AddCommand(locationsStartCmd)

	// Integrations
	rootCmd.AddCommand(integrationsCmd)
	integrationsCmd.AddCommand(integrationsAkamaiCmd)
	integrationsAkamaiCmd.AddCommand(integrationsAkamaiList)
	integrationsCmd.AddCommand(integrationsKubernetesCmd)
	integrationsKubernetesCmd.AddCommand(integrationsKubernetesList)
	integrationsKubernetesCmd.AddCommand(integrationsKubernetesDelete)

	// Scan
	rootCmd.AddCommand(startScanCmd)

	return rootCmd.Execute()
}
