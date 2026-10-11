package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	clischema "github.com/Escape-Technologies/cli/pkg/cli/schema"
	"github.com/spf13/cobra"
)

const (
	exportFormatCSV   = "csv"
	exportFormatSARIF = "sarif"
	csvContentType    = "text/csv"
	// exportFilePerm keeps a written export readable only by the user who ran
	// the command. The file is the raw CSV or SARIF, not the JSON wrapper.
	exportFilePerm os.FileMode = 0o600
)

// csvExport is the JSON document for a CSV export.
// Pretty mode prints Content as the raw file instead of this object.
type csvExport struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

// sarifExport is the minimum SARIF object the issues export schema accepts.
// The command prints the server document unchanged; this type is only the schema.
type sarifExport struct {
	Version string `json:"version"`
	Runs    []any  `json:"runs"`
}

// issueExportOutput is the JSON stdout of `issues export`.
// CSV is csvExport. SARIF is the SARIF document itself, so the schema is a oneOf.
type issueExportOutput struct{}

func (issueExportOutput) JSONSchema() *clischema.JSONSchema {
	csvSchema := clischema.Generate(csvExport{})
	sarifSchema := clischema.Generate(sarifExport{})
	// Generate stamps the draft URI on each root, then returns a schemaProvider
	// document without stamping it again. Nested oneOf branches are not
	// documents, so the URI stays off them. The union root is what schema mode
	// prints, so it keeps the draft URI.
	csvSchema.Schema = ""
	sarifSchema.Schema = ""

	// Both documents are objects. The catalog rejects a schema with no type.
	return &clischema.JSONSchema{
		Schema: jsonSchemaDraft,
		Type:   "object",
		OneOf:  []*clischema.JSONSchema{csvSchema, sarifSchema},
	}
}

var (
	issueExportFormat string
	issueExportWhere  string
	issueExportOut    string

	assetExportFormat string
	assetExportWhere  string
	assetExportOut    string
)

var issueExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export issues as CSV or SARIF (Beta)",
	Long: `Download the issues that match --where as CSV or SARIF.

JSON mode prints SARIF as the document itself and CSV as {"contentType":"text/csv","content":"..."}.
Pretty mode prints the raw file. --out also writes that raw file.
Pass a JSON IssueWhereInput to --where, the same shape as jobs trigger-export --issue-where.
A payload that does not fit the tool output limit should be retried with a narrower where.`,
	Example: `  escape-cli issues export --format csv --where '{"status":"OPEN"}'
  escape-cli issues export --format sarif -o json
  escape-cli issues export --format csv --out issues.csv`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(issueExportOutput{}) {
			return nil
		}

		format, err := normalizeExportFormat(issueExportFormat, exportFormatCSV, exportFormatSARIF)
		if err != nil {
			return err
		}

		where, err := parseWhereFlag(cmd, "where", issueExportWhere)
		if err != nil {
			return err
		}

		var document escape.Export
		if format == exportFormatSARIF {
			document, err = escape.ExportIssuesSARIF(cmd.Context(), where)
		} else {
			document, err = escape.ExportIssuesCSV(cmd.Context(), where)
		}

		if err != nil {
			return fmt.Errorf("unable to export issues: %w", err)
		}

		return emitExport(issueExportOut, format, document.Body)
	},
}

var assetExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export assets as CSV (Beta)",
	Long: `Download the assets that match --where as CSV.

JSON mode prints {"contentType":"text/csv","content":"..."}. Pretty mode prints the raw file.
--out also writes that raw file. Pass a JSON AssetWhereInput to --where, the same
shape as jobs trigger-export --asset-where.
A payload that does not fit the tool output limit should be retried with a narrower where.`,
	Example: `  escape-cli assets export --format csv --where '{"statuses":["MONITORED"]}'
  escape-cli assets export --format csv --out assets.csv`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(csvExport{}) {
			return nil
		}

		format, err := normalizeExportFormat(assetExportFormat, exportFormatCSV)
		if err != nil {
			return err
		}

		where, err := parseWhereFlag(cmd, "where", assetExportWhere)
		if err != nil {
			return err
		}

		document, err := escape.ExportAssetsCSV(cmd.Context(), where)
		if err != nil {
			return fmt.Errorf("unable to export assets: %w", err)
		}

		return emitExport(assetExportOut, format, document.Body)
	},
}

func init() {
	issuesCmd.AddCommand(issueExportCmd)
	issueExportCmd.Flags().StringVar(&issueExportFormat, "format", "", "export format: csv or sarif")
	setEnumAnnotation(issueExportCmd.Flags(), "format", []string{exportFormatCSV, exportFormatSARIF})
	// Not MarkFlagRequired: cobra checks that before RunE, so `issues export -o schema`
	// would fail before the schema is printed. MCP still requires the property.
	// RunE checks the value after out.Schema returns.
	markMCPRequired(issueExportCmd.Flags(), "format")
	issueExportCmd.Flags().StringVar(&issueExportWhere, "where", "", "JSON IssueWhereInput, the same shape as jobs trigger-export --issue-where")
	issueExportCmd.Flags().StringVar(&issueExportOut, "out", "", "write the raw export to this file")

	assetsCmd.AddCommand(assetExportCmd)
	assetExportCmd.Flags().StringVar(&assetExportFormat, "format", exportFormatCSV, "export format: csv")
	setEnumAnnotation(assetExportCmd.Flags(), "format", []string{exportFormatCSV})
	assetExportCmd.Flags().StringVar(&assetExportWhere, "where", "", "JSON AssetWhereInput, the same shape as jobs trigger-export --asset-where")
	assetExportCmd.Flags().StringVar(&assetExportOut, "out", "", "write the raw export to this file")
}

// parseWhereFlag accepts one JSON object, the where-input the export routes forward.
// An empty value is omitted. A flag that was set to blank is rejected, the same
// rule as jobs trigger-export --issue-where.
func parseWhereFlag(cmd *cobra.Command, name, value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		if cmd.Flags().Changed(name) {
			return "", fmt.Errorf("--%s must not be blank", name)
		}

		return "", nil
	}

	var decoded any
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return "", fmt.Errorf("--%s must be a JSON object: %w", name, err)
	}

	if _, ok := decoded.(map[string]any); !ok {
		return "", fmt.Errorf("--%s must be a JSON object", name)
	}

	return value, nil
}

func normalizeExportFormat(value string, allowed ...string) (string, error) {
	format := strings.ToLower(strings.TrimSpace(value))
	for _, candidate := range allowed {
		if format == candidate {
			return format, nil
		}
	}

	return "", fmt.Errorf("--format must be one of: %s", strings.Join(allowed, ", "))
}

func emitExport(path, format string, body []byte) error {
	if format == exportFormatSARIF && !json.Valid(body) {
		return errors.New("sarif export is not a JSON document")
	}

	if path != "" {
		if err := os.WriteFile(path, body, exportFilePerm); err != nil {
			return fmt.Errorf("write export to %s: %w", path, err)
		}
	}

	// Pretty mode is the file itself. out.Print appends a newline, which adds
	// a blank CSV record when the body already ends in "\n" and makes stdout
	// differ from --out. json and yaml keep the structured document.
	if out.IsPretty() {
		if _, err := os.Stdout.Write(body); err != nil {
			return fmt.Errorf("write export: %w", err)
		}

		return nil
	}

	if format == exportFormatSARIF {
		out.Print(json.RawMessage(body), string(body))

		return nil
	}

	text := string(body)
	out.Print(csvExport{ContentType: csvContentType, Content: text}, text)

	return nil
}
