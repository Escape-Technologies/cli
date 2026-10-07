package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	clischema "github.com/Escape-Technologies/cli/pkg/cli/schema"
	"github.com/spf13/cobra"
)

// CommandSchemas captures optional runtime zero-values for a command's input
// and output Go structs so JSON schemas can be generated and surfaced both in
// `capabilities` output and by the embedded MCP server.
type CommandSchemas struct {
	Input  any
	Output any
}

// CommandCapability is the JSON-friendly view of one CLI command used by the
// `capabilities` command and consumed by the embedded MCP server to build its
// tool catalog.
type CommandCapability struct {
	Path                string                `json:"path"`
	Use                 string                `json:"use"`
	Short               string                `json:"short,omitempty"`
	Aliases             []string              `json:"aliases,omitempty"`
	HasSub              bool                  `json:"hasSubcommands"`
	HasFlags            bool                  `json:"hasFlags"`
	HasInSchema         bool                  `json:"hasInputSchema"`
	HasOutSchema        bool                  `json:"hasOutputSchema"`
	InputSchema         *clischema.JSONSchema `json:"inputSchema,omitempty"`
	OutputSchema        *clischema.JSONSchema `json:"outputSchema,omitempty"`
	InputSchemaCommand  string                `json:"inputSchemaCommand,omitempty"`
	OutputSchemaCommand string                `json:"outputSchemaCommand,omitempty"`
}

type commandCapabilitySchemaOutput struct {
	Path                string         `json:"path"`
	Use                 string         `json:"use"`
	Short               string         `json:"short,omitempty"`
	Aliases             []string       `json:"aliases,omitempty"`
	HasSub              bool           `json:"hasSubcommands"`
	HasFlags            bool           `json:"hasFlags"`
	HasInSchema         bool           `json:"hasInputSchema"`
	HasOutSchema        bool           `json:"hasOutputSchema"`
	InputSchema         map[string]any `json:"inputSchema,omitempty"`
	OutputSchema        map[string]any `json:"outputSchema,omitempty"`
	InputSchemaCommand  string         `json:"inputSchemaCommand,omitempty"`
	OutputSchemaCommand string         `json:"outputSchemaCommand,omitempty"`
}

func schemaFor(v any) *clischema.JSONSchema {
	if v == nil {
		return nil
	}

	return clischema.Generate(v)
}

// CommandSchemaRegistry returns the explicit allowlist mapping Cobra command
// paths to their input/output Go structs. New commands exposed to MCP must be
// added here so the Cobra-to-MCP surface stays reviewed rather than derived.
// scans watch is omitted: it blocks until the scan ends, which an MCP tool
// call cannot finish inside the tool budget. Models poll scans get instead.
func CommandSchemaRegistry() map[string]CommandSchemas {
	return map[string]CommandSchemas{
		"escape-cli me":                            {Output: v3.GetMe200Response{}},
		"escape-cli users me":                      {Output: v3.GetMe200Response{}},
		"escape-cli users list":                    {Output: []v3.ListUsers200ResponseInner{}},
		"escape-cli users get":                     {Output: v3.GetUser200Response{}},
		"escape-cli users invite":                  {Output: []v3.ListUsers200ResponseInner{}},
		"escape-cli roles list":                    {Output: []v3.ListRoles200ResponseInner{}},
		"escape-cli roles get":                     {Output: v3.CreateRole200Response{}},
		"escape-cli roles create":                  {Input: v3.CreateRoleRequest{}, Output: v3.CreateRole200Response{}},
		"escape-cli roles update":                  {Input: v3.UpdateRoleRequest{}, Output: v3.CreateRole200Response{}},
		"escape-cli roles bind":                    {Output: []v3.CreateRoleBindings200ResponseInner{}},
		"escape-cli roles unbind":                  {Output: v3.DeleteCustomRule200Response{}},
		"escape-cli projects list":                 {Output: Page[v3.ListProjects200ResponseDataInner]{}},
		"escape-cli projects get":                  {Output: v3.CreateProject200Response{}},
		"escape-cli projects create":               {Input: v3.CreateProjectRequest{}, Output: v3.CreateProject200Response{}},
		"escape-cli projects update":               {Input: v3.UpdateProjectRequest{}, Output: v3.CreateProject200Response{}},
		"escape-cli integrations list":             {Output: Page[map[string]interface{}]{}},
		"escape-cli integrations get":              {Output: map[string]interface{}{}},
		"escape-cli integrations create":           {Input: mustInputSchema(integrationCreateInputSchema), Output: map[string]interface{}{}},
		"escape-cli integrations update":           {Input: mustInputSchema(integrationUpdateInputSchema), Output: map[string]interface{}{}},
		"escape-cli integrations delete":           {Output: v3.CreateakamaiIntegration200Response{}},
		"escape-cli workflows list":                {Output: Page[v3.WorkflowSummarized]{}},
		"escape-cli workflows get":                 {Output: v3.CreateWorkflow200Response{}},
		"escape-cli workflows create":              {Input: v3.CreateWorkflowRequest{}, Output: v3.CreateWorkflow200Response{}},
		"escape-cli workflows update":              {Input: v3.UpdateWorkflowRequest{}, Output: v3.CreateWorkflow200Response{}},
		"escape-cli workflows delete":              {Output: v3.CreateWorkflow200Response{}},
		"escape-cli profiles list":                 {Output: Page[v3.ProfileSummarized]{}},
		"escape-cli profiles get":                  {Output: v3.GetProfile200Response{}},
		"escape-cli profiles get-schema":           {Output: v3.ProfileExtraAsset{}},
		"escape-cli profiles create-rest":          {Input: createRestProfileInput{}, Output: v3.GetProfile200Response{}},
		"escape-cli profiles create-webapp":        {Input: createWebappProfileInput{}, Output: v3.GetProfile200Response{}},
		"escape-cli profiles create-graphql":       {Input: createGraphqlProfileInput{}, Output: v3.GetProfile200Response{}},
		"escape-cli profiles create-ai-pentest":    {Input: createAiPentestProfileInput{}, Output: v3.GetProfile200Response{}},
		"escape-cli profiles update":               {Input: v3.UpdateProfileRequest{}, Output: v3.GetProfile200Response{}},
		"escape-cli profiles update-configuration": {Input: v3.UpdateProfileConfigurationRequest{}},
		"escape-cli profiles problems":             {Output: Page[v3.ProfileScanProblemsRow]{}},
		"escape-cli profiles delete":               {Output: v3.DeleteProfile200Response{}},
		"escape-cli issues list":                   {Output: Page[v3.IssueSummarized]{}},
		"escape-cli issues get":                    {Output: v3.GetIssue200Response{}},
		"escape-cli issues get-with-events":        {Output: IssueWithEvents{}},
		"escape-cli issues list-activities":        {Output: []v3.ActivitySummarized{}},
		"escape-cli issues funnel":                 {Output: []escape.IssueFunnelStep{}},
		"escape-cli issues trends":                 {Output: []escape.IssueTrendPoint{}},
		"escape-cli issues update":                 {Output: v3.UpdateIssue200Response{}},
		"escape-cli issues bulk-update":            {Output: v3.BulkUpdateIssues200Response{}},
		"escape-cli issues notify":                 {Output: v3.NotifyIssueOwners200Response{}},
		"escape-cli issues trigger-workflow":       {Output: v3.TriggerIssueManualWorkflow200Response{}},
		"escape-cli issues remediation generate":   {Input: v3.GenerateIssueAiRemediationRequest{}, Output: v3.GenerateIssueAiRemediation200Response{}},
		"escape-cli issues remediation get":        {Output: IssueRemediation{}},
		"escape-cli issues remediation feedback":   {Input: v3.SaveIssueAiRemediationFeedbackRequest{}, Output: v3.SaveIssueAiRemediationFeedback200Response{}},
		"escape-cli scans list":                    {Output: []v3.ScanSummarized{}},
		"escape-cli scans get":                     {Output: v3.StartScan200Response{}},
		"escape-cli scans start":                   {Output: v3.ScanDetailed1{}},
		"escape-cli scans cancel":                  {Output: out.Message{}},
		"escape-cli scans ignore":                  {Output: out.Message{}},
		"escape-cli scans issues":                  {Output: Page[v3.IssueSummarized]{}},
		"escape-cli scans targets":                 {Output: []v3.TargetDetailed{}},
		"escape-cli scans coverage":                {Output: ScanCoverage{}},
		"escape-cli scans reasoning":               {Output: ScanReasoningLogs{}},
		"escape-cli scans agents":                  {Output: ScanAgents{}},
		"escape-cli scans problems":                {Output: Page[v3.ScanSummarizedWithProblems2]{}},
		"escape-cli events list":                   {Output: Page[v3.EventSummarized]{}},
		"escape-cli events get":                    {Output: v3.GetEvent200Response{}},
		"escape-cli emails list":                   {Output: []v3.ScanEmailSummary{}},
		"escape-cli emails read":                   {Output: v3.ScanEmailDetails{}},
		"escape-cli emails wait":                   {Output: v3.ScanEmailDetails{}},
		"escape-cli authentications start":         {Input: v3.StartAuthenticationRequest{}, Output: v3.StartAuthentication200Response{}},
		"escape-cli authentications get":           {Output: v3.GetAuthentication200Response{}},
		"escape-cli jobs trigger-export":           {Output: v3.TriggerExport200Response{}},
		"escape-cli jobs get":                      {Output: v3.GetJob200Response{}},
		"escape-cli locations list":                {Output: Page[v3.LocationSummarized]{}},
		"escape-cli locations get":                 {Output: v3.CreateLocation200Response{}},
		"escape-cli locations create":              {Input: v3.CreateLocationRequest{}, Output: v3.CreateLocation200Response{}},
		"escape-cli locations update":              {Input: v3.UpdateLocationRequest{}, Output: v3.CreateLocation200Response{}},
		"escape-cli locations delete":              {Output: v3.DeleteLocation200Response{}},
		"escape-cli assets list":                   {Output: Page[v3.AssetSummarized]{}},
		"escape-cli assets get":                    {Output: v3.AssetDetailed{}},
		"escape-cli assets list-activities":        {Output: []v3.ActivitySummarized{}},
		"escape-cli assets create":                 {Input: mustInputSchema(assetCreateInputSchema), Output: v3.AssetDetailed{}},
		"escape-cli assets update":                 {Output: v3.UpdateAsset200Response{}},
		"escape-cli assets comment":                {Output: v3.CreateAssetComment200Response{}},
		"escape-cli assets delete":                 {Output: v3.DeleteProfile200Response{}}, // generated client reuses this {"message": string} schema
		"escape-cli assets bulk-update":            {Output: v3.BulkUpdateAssets200Response{}},
		"escape-cli assets bulk-delete":            {Output: v3.BulkUpdateAssets200Response{}},
		"escape-cli assets bulk-import":            {Input: v3.BulkImportAssets{}, Output: v3.BulkImportAssets200Response{}},
		"escape-cli asm trigger":                   {Output: v3.TriggerAsmScans200Response{}},
		"escape-cli custom-rules list":             {Output: []v3.CustomRuleSummarized{}},
		"escape-cli custom-rules get":              {Output: v3.CreateCustomRule200Response{}},
		"escape-cli custom-rules create":           {Input: v3.CreateCustomRuleRequest{}, Output: v3.CreateCustomRule200Response{}},
		"escape-cli custom-rules update":           {Input: v3.UpdateCustomRuleRequest{}, Output: v3.CreateCustomRule200Response{}},
		"escape-cli custom-rules delete":           {Output: v3.DeleteCustomRule200Response{}},
		"escape-cli tags list":                     {Output: []v3.TagDetail{}},
		"escape-cli tags get":                      {Output: v3.TagDetail{}},
		"escape-cli tags create":                   {Output: v3.CreateTag200Response{}},
		"escape-cli tags update":                   {Output: v3.CreateTag200Response{}},
		"escape-cli tags delete":                   {Output: v3.DeleteProfile200Response{}}, // same generated schema as assets delete
		"escape-cli audit list":                    {Output: []v3.AuditLogSummarized{}},
		"escape-cli stats":                         {Output: v3.GetStatistics200Response{}},
		"escape-cli problems":                      {Output: problemsCommandOutput},
		"escape-cli issues comment":                {Input: v3.CreateAssetCommentRequest{}, Output: v3.CreateAssetComment200Response{}},
		"escape-cli retests list":                  {Output: Page[v3.RetestDetailed]{}},
		"escape-cli retests get":                   {Output: v3.GetRetest200Response{}},
		"escape-cli retests start":                 {Input: v3.StartRetestRequest{}, Output: v3.StartRetest200Response{}},
		"escape-cli capabilities":                  {Output: []commandCapabilitySchemaOutput{}},
	}
}

// BuildCommandCapabilities walks the Cobra command tree and returns the
// materialized capability descriptors enriched with JSON schemas from the
// supplied registry. Consumed by both the `capabilities` command and the MCP
// server's tool catalog builder.
func BuildCommandCapabilities(root *cobra.Command, registry map[string]CommandSchemas) []CommandCapability {
	capabilities := make([]CommandCapability, 0)

	var walk func(command *cobra.Command)
	walk = func(command *cobra.Command) {
		if !command.IsAvailableCommand() || command.Hidden {
			return
		}

		schemas := registry[command.CommandPath()]
		inputSchema := schemaFor(schemas.Input)
		outputSchema := schemaFor(schemas.Output)
		// LocalFlags includes this command's PersistentFlags and excludes
		// inherited parent flags. Flags() alone drops leaf persistent
		// flags until the command is parsed.
		hasFlags := command.LocalFlags().HasAvailableFlags()
		capabilities = append(capabilities, CommandCapability{
			Path:         command.CommandPath(),
			Use:          command.Use,
			Short:        command.Short,
			Aliases:      command.Aliases,
			HasSub:       command.HasAvailableSubCommands(),
			HasFlags:     hasFlags,
			HasInSchema:  inputSchema != nil,
			HasOutSchema: outputSchema != nil,
			InputSchema:  inputSchema,
			OutputSchema: outputSchema,
			InputSchemaCommand: func() string {
				if inputSchema == nil {
					return ""
				}

				return command.CommandPath() + " --input-schema"
			}(),
			OutputSchemaCommand: func() string {
				if outputSchema == nil {
					return ""
				}

				return command.CommandPath() + " --output schema"
			}(),
		})
		for _, child := range command.Commands() {
			walk(child)
		}
	}

	walk(root)

	return capabilities
}

var capabilitiesCmd = &cobra.Command{
	Use:   "capabilities",
	Short: "Describe CLI commands in a machine-readable format",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		if out.Schema([]commandCapabilitySchemaOutput{}) {
			return nil
		}

		capabilities := BuildCommandCapabilities(rootCmd, CommandSchemaRegistry())

		pretty, err := json.MarshalIndent(capabilities, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal capabilities output: %w", err)
		}

		out.Print(capabilities, string(pretty))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(capabilitiesCmd)
}
