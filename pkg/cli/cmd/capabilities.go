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

// outputWithDryRun is the output schema of a command whose --dry-run prints
// a different document from the mutation result. anyOf, not oneOf: a preview
// that also matches the result schema (a loose object, or extra fields the
// result schema allows) must still validate. oneOf would reject it.
func outputWithDryRun(result, preview any) providedSchema {
	branches := []*clischema.JSONSchema{
		dryRunBranch(result),
		dryRunBranch(preview),
	}
	defs := map[string]*clischema.JSONSchema{}
	for _, branch := range branches {
		for name, def := range branch.Defs {
			if _, exists := defs[name]; !exists {
				defs[name] = def
			}
		}

		// $ref resolves at the document root. Nested $defs would make the
		// MCP schema and schemaFor disagree after liftNestedDefs rewrites
		// only one of them.
		branch.Defs = nil
		branch.Schema = ""
	}

	schema := &clischema.JSONSchema{
		Schema:      jsonSchemaDraft,
		Type:        "object",
		Description: "Either the command result or the document --dry-run prints.",
		AnyOf:       branches,
	}
	if len(defs) > 0 {
		schema.Defs = defs
	}

	return providedSchema{schema: schema}
}

// dryRunBranch copies the generated root so clearing $schema and $defs does
// not change a schema another command still holds. Generate returns a
// schemaProvider document directly, without cloning.
func dryRunBranch(value any) *clischema.JSONSchema {
	generated := clischema.Generate(value)
	cloned := *generated

	return &cloned
}

// Dry-run previews are the document confirm.go prints when --dry-run is set
// and the command can fetch the target. Message-only deletes print the same
// {"msg"} document either way, so they stay out.Message.
var (
	assetDeleteOutput          = outputWithDryRun(v3.DeleteProfile200Response{}, v3.AssetDetailed1{})
	assetBulkUpdateOutput      = outputWithDryRun(escape.BulkOperationResult{}, out.Message{})
	assetBulkDeleteOutput      = outputWithDryRun(escape.BulkOperationResult{}, out.Message{})
	tagDeleteOutput            = outputWithDryRun(v3.DeleteProfile200Response{}, v3.CreateTag200Response{})
	profileDeleteOutput        = outputWithDryRun(v3.DeleteProfile200Response{}, v3.GetProfile200Response{})
	locationDeleteOutput       = outputWithDryRun(v3.DeleteLocation200Response{}, v3.GetLocation200Response{})
	customRuleDeleteOutput     = outputWithDryRun(v3.DeleteCustomRule200Response{}, v3.CreateCustomRule200Response{})
	integrationDeleteOutput    = outputWithDryRun(v3.CreateakamaiIntegration200Response{}, map[string]interface{}{})
	regressionTestDeleteOutput = outputWithDryRun(v3.DeleteProfile200Response{}, v3.CreateRegressionTest200Response{})
	userDeactivateOutput       = outputWithDryRun(v3.ListUsers200ResponseInner{}, v3.GetUser200Response{})
	// roles delete and projects delete are not MCP tools. --dry-run still
	// prints the fetched resource, so the schema guard declares the same union.
	roleDeleteOutput    = outputWithDryRun(out.Message{}, v3.CreateRole200Response{})
	projectDeleteOutput = outputWithDryRun(out.Message{}, v3.CreateProject200Response{})
)

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
		"escape-cli integrations delete":           {Output: integrationDeleteOutput},
		"escape-cli integrations test":             {Output: v3.SyncAkamaiIntegration200Response{}},
		"escape-cli integrations sync":             {Output: v3.SyncAkamaiIntegration200Response{}},
		"escape-cli integrations runs":             {Output: Page[map[string]interface{}]{}},
		"escape-cli views list":                    {Output: []v3.ViewDetail{}},
		"escape-cli views issues":                  {Output: Page[v3.IssueSummarized]{}},
		"escape-cli views create":                  {Input: v3.CreateViewRequest{}, Output: v3.CreateView200Response{}},
		"escape-cli views update":                  {Input: v3.UpdateViewRequest{}, Output: v3.CreateView200Response{}},
		"escape-cli views reorder":                 {Input: v3.ReorderViewsRequest{}, Output: out.Message{}},
		"escape-cli views delete":                  {Output: out.Message{}},
		"escape-cli vault variables list":          {Output: Page[v3.VaultVariableSummary]{}},
		"escape-cli vault variables create":        {Output: v3.CreateVaultVariable200Response{}},
		"escape-cli vault variables delete":        {Output: out.Message{}},
		"escape-cli vault variables rotate":        {Output: v3.CreateVaultVariable200Response{}},
		"escape-cli vault list":                    {Output: Page[v3.VaultSummary1]{}},
		"escape-cli vault link":                    {Output: v3.LinkVault200Response{}},
		"escape-cli vault entries":                 {Output: []v3.VaultEntry{}},
		"escape-cli knowledge-files list":          {Output: Page[v3.KnowledgeFileSummary]{}},
		"escape-cli knowledge-files delete":        {Output: out.Message{}},
		"escape-cli knowledge-files upload":        {Output: v3.UploadKnowledgeFile200Response{}},
		"escape-cli org-settings get":              {Output: v3.GetOrganizationSettings200Response{}},
		"escape-cli org-settings update":           {Input: v3.UpdateOrganizationSettingsRequest{}, Output: v3.GetOrganizationSettings200Response{}},
		"escape-cli users deactivate":              {Output: userDeactivateOutput},
		"escape-cli workflows list":                {Output: Page[v3.WorkflowSummarized]{}},
		"escape-cli workflows get":                 {Output: v3.CreateWorkflow200Response{}},
		"escape-cli workflows create":              {Input: v3.CreateWorkflowRequest{}, Output: v3.CreateWorkflow200Response{}},
		"escape-cli workflows update":              {Input: v3.UpdateWorkflowRequest{}, Output: v3.CreateWorkflow200Response{}},
		"escape-cli workflows delete":              {Output: v3.CreateWorkflow200Response{}},
		"escape-cli workflows pause":               {Output: v3.CreateWorkflow200Response{}},
		"escape-cli workflows executions":          {Output: Page[v3.WorkflowExecution]{}},
		"escape-cli locations logs":                {Output: Page[v3.LocationLog]{}},
		"escape-cli locations tags":                {Output: v3.GetLocation200Response{}},
		"escape-cli profiles list":                 {Output: Page[v3.ProfileSummarized]{}},
		"escape-cli profiles get":                  {Output: v3.GetProfile200Response{}},
		"escape-cli profiles get-schema":           {Output: v3.ProfileExtraAsset{}},
		"escape-cli profiles create-rest":          {Input: createRestProfileInput{}, Output: v3.GetProfile200Response{}},
		"escape-cli profiles create-webapp":        {Input: createWebappProfileInput{}, Output: v3.GetProfile200Response{}},
		"escape-cli profiles create-graphql":       {Input: createGraphqlProfileInput{}, Output: v3.GetProfile200Response{}},
		"escape-cli profiles create-ai-pentest":    {Input: createAiPentestProfileInput{}, Output: v3.GetProfile200Response{}},
		"escape-cli profiles update":               {Input: v3.UpdateProfileRequest{}, Output: v3.GetProfile200Response{}},
		"escape-cli profiles update-configuration": {Input: v3.UpdateProfileConfigurationRequest{}, Output: map[string]interface{}{}},
		"escape-cli profiles continuous-pentest create": {
			Input:  v3.ContinuousPentestCreateInput{},
			Output: v3.UpdateContinuousPentest200Response{},
		},
		"escape-cli profiles continuous-pentest update": {
			Input:  v3.ContinuousPentestUpdateInput{},
			Output: v3.UpdateContinuousPentest200Response{},
		},
		"escape-cli profiles continuous-pentest enable": {
			Input:  v3.ContinuousPentestEnableInput{},
			Output: v3.UpdateContinuousPentest200Response{},
		},
		"escape-cli profiles continuous-pentest disable": {Output: v3.UpdateContinuousPentest200Response{}},
		"escape-cli profiles problems":                   {Output: Page[v3.ProfileScanProblemsRow]{}},
		"escape-cli profiles delete":                     {Output: profileDeleteOutput},
		"escape-cli issues list":                         {Output: Page[v3.IssueSummarized]{}},
		"escape-cli issues get":                          {Output: v3.GetIssue200Response{}},
		"escape-cli issues get-with-events":              {Output: IssueWithEvents{}},
		"escape-cli issues list-activities":              {Output: []v3.ActivitySummarized{}},
		"escape-cli issues funnel":                       {Output: []escape.IssueFunnelStep{}},
		"escape-cli issues trends":                       {Output: []escape.IssueTrendPoint{}},
		"escape-cli issues update":                       {Output: v3.UpdateIssue200Response{}},
		"escape-cli issues bulk-update":                  {Output: v3.BulkUpdateIssues200Response{}},
		"escape-cli issues notify":                       {Output: v3.NotifyIssueOwners200Response{}},
		"escape-cli issues trigger-workflow":             {Output: v3.TriggerIssueManualWorkflow200Response{}},
		"escape-cli issues remediation generate":         {Input: v3.GenerateIssueAiRemediationRequest{}, Output: v3.GenerateIssueAiRemediation200Response{}},
		"escape-cli issues remediation get":              {Output: IssueRemediation{}},
		"escape-cli issues remediation feedback":         {Input: v3.SaveIssueAiRemediationFeedbackRequest{}, Output: v3.SaveIssueAiRemediationFeedback200Response{}},
		"escape-cli scans list":                          {Output: []v3.ScanSummarized{}},
		"escape-cli scans get":                           {Output: v3.StartScan200Response{}},
		"escape-cli scans configuration":                 {Output: v3.GetScanConfiguration200Response{}},
		"escape-cli scans statistics":                    {Output: v3.GetScanStatistics200Response{}},
		"escape-cli scans start":                         {Output: v3.ScanDetailed1{}},
		"escape-cli scans cancel":                        {Output: out.Message{}},
		"escape-cli scans ignore":                        {Output: out.Message{}},
		"escape-cli scans issues":                        {Output: Page[v3.IssueSummarized]{}},
		"escape-cli scans targets":                       {Output: []v3.TargetDetailed{}},
		"escape-cli scans coverage":                      {Output: ScanCoverage{}},
		"escape-cli scans reasoning":                     {Output: ScanReasoningLogs{}},
		"escape-cli scans agents":                        {Output: ScanAgents{}},
		"escape-cli scans problems":                      {Output: Page[v3.ScanSummarizedWithProblems2]{}},
		"escape-cli events list":                         {Output: Page[v3.EventSummarized]{}},
		"escape-cli events get":                          {Output: v3.GetEvent200Response{}},
		"escape-cli emails list":                         {Output: []v3.ScanEmailSummary{}},
		"escape-cli emails read":                         {Output: v3.ScanEmailDetails{}},
		"escape-cli emails wait":                         {Output: v3.ScanEmailDetails{}},
		"escape-cli authentications start":               {Input: v3.StartAuthenticationRequest{}, Output: v3.StartAuthentication200Response{}},
		"escape-cli authentications get":                 {Output: v3.GetAuthentication200Response{}},
		"escape-cli jobs trigger-export":                 {Output: v3.TriggerExport200Response{}},
		"escape-cli jobs get":                            {Output: v3.GetJob200Response{}},
		"escape-cli locations list":                      {Output: Page[v3.LocationSummarized]{}},
		"escape-cli locations get":                       {Output: v3.GetLocation200Response{}},
		"escape-cli locations create":                    {Input: v3.CreateLocationRequest{}, Output: v3.CreateLocation200Response{}},
		"escape-cli locations update":                    {Input: v3.UpdateLocationRequest{}, Output: v3.GetLocation200Response{}},
		"escape-cli locations delete":                    {Output: locationDeleteOutput},
		"escape-cli assets list":                         {Output: Page[v3.AssetSummarized]{}},
		"escape-cli assets get":                          {Output: v3.AssetDetailed{}},
		"escape-cli assets list-activities":              {Output: []v3.ActivitySummarized{}},
		"escape-cli assets create":                       {Input: mustInputSchema(assetCreateInputSchema), Output: v3.AssetDetailed{}},
		"escape-cli assets update":                       {Output: v3.UpdateAsset200Response{}},
		"escape-cli assets comment":                      {Output: v3.CreateAssetComment200Response{}},
		"escape-cli assets delete":                       {Output: assetDeleteOutput},
		"escape-cli assets bulk-update":                  {Output: assetBulkUpdateOutput},
		"escape-cli assets bulk-delete":                  {Output: assetBulkDeleteOutput},
		"escape-cli assets bulk-import":                  {Input: v3.BulkImportAssets{}, Output: v3.BulkImportAssets200Response{}},
		"escape-cli asm trigger":                         {Output: v3.TriggerAsmScans200Response{}},
		"escape-cli custom-rules list":                   {Output: []v3.CustomRuleSummarized{}},
		"escape-cli custom-rules get":                    {Output: v3.CreateCustomRule200Response{}},
		"escape-cli custom-rules create":                 {Input: v3.CreateCustomRuleRequest{}, Output: v3.CreateCustomRule200Response{}},
		"escape-cli custom-rules update":                 {Input: v3.UpdateCustomRuleRequest{}, Output: v3.CreateCustomRule200Response{}},
		"escape-cli custom-rules delete":                 {Output: customRuleDeleteOutput},
		"escape-cli tags list":                           {Output: []v3.TagDetail{}},
		"escape-cli tags get":                            {Output: v3.TagDetail{}},
		"escape-cli tags create":                         {Output: v3.CreateTag200Response{}},
		"escape-cli tags update":                         {Output: v3.CreateTag200Response{}},
		"escape-cli tags delete":                         {Output: tagDeleteOutput},
		"escape-cli audit list":                          {Output: []v3.AuditLogSummarized{}},
		"escape-cli stats":                               {Output: v3.GetStatistics200Response{}},
		"escape-cli problems":                            {Output: problemsCommandOutput},
		"escape-cli issues comment":                      {Input: v3.CreateAssetCommentRequest{}, Output: v3.CreateAssetComment200Response{}},
		"escape-cli retests list":                        {Output: Page[v3.RetestDetailed]{}},
		"escape-cli retests get":                         {Output: v3.GetRetest200Response{}},
		"escape-cli retests start":                       {Input: v3.StartRetestRequest{}, Output: v3.StartRetest200Response{}},
		"escape-cli regression-tests list":               {Output: Page[v3.RegressionTestSummary]{}},
		"escape-cli regression-tests get":                {Output: v3.CreateRegressionTest200Response{}},
		"escape-cli regression-tests create":             {Input: v3.CreateRegressionTestInput{}, Output: v3.CreateRegressionTest200Response{}},
		"escape-cli regression-tests update":             {Input: v3.UpdateRegressionTestRequest{}, Output: v3.CreateRegressionTest200Response{}},
		"escape-cli regression-tests delete":             {Output: regressionTestDeleteOutput},
		"escape-cli regression-tests run":                {Output: v3.CreateRegressionTest200Response{}},
		"escape-cli regression-tests stop":               {Output: v3.CreateRegressionTest200Response{}},
		"escape-cli regression-tests history":            {Output: Page[v3.RegressionTestRun]{}},
		"escape-cli regression-tests answer": {
			Input:  v3.ReplyRegressionTestClarificationInput{},
			Output: v3.ReplyRegressionTestClarification200Response{},
		},
		"escape-cli security-tests list":        {Output: Page[v3.SecurityTest]{}},
		"escape-cli security-tests get":         {Output: v3.GetSecurityTest200Response{}},
		"escape-cli compliance list":            {Output: v3.GetCompliancePosture200Response{}},
		"escape-cli compliance enabled":         {Output: v3.GetEnabledCompliances200Response{}},
		"escape-cli compliance per-application": {Output: Page[v3.GetCompliancePerApplication200ResponseDataInner]{}},
		"escape-cli scans compliance":           {Output: v3.GetScanCompliance200Response{}},
		"escape-cli statistics endpoints":       {Output: v3.GetStatisticsEndpoints200Response{}},
		"escape-cli statistics issues":          {Output: v3.GetStatisticsIssues200Response{}},
		"escape-cli statistics trends":          {Output: v3.GetStatisticsTrends200Response{}},
		"escape-cli permissions list":           {Output: []v3.PermissionCatalogEntry{}},
		"escape-cli validations start": {
			Input:  v3.StartValidationRequest{},
			Output: v3.StartValidation200Response{},
		},
		"escape-cli validations get":   {Output: v3.GetValidation200Response{}},
		"escape-cli ai-chats list":     {Output: Page[v3.AIChatSummary]{}},
		"escape-cli ai-chats get":      {Output: v3.GetAIChat200Response{}},
		"escape-cli ai-chats messages": {Output: Page[v3.AIChatMessage]{}},
		"escape-cli ai-chats create":   {Output: v3.CreateAIChat200Response{}},
		"escape-cli ai-chats message":  {Output: v3.CreateAIChatMessage200Response{}},
		"escape-cli jobs list":         {Output: Page[v3.JobSummary]{}},
		"escape-cli capabilities":      {Output: []commandCapabilitySchemaOutput{}},
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
