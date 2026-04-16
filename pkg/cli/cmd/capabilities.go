package cmd

import (
	"encoding/json"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	clischema "github.com/Escape-Technologies/cli/pkg/cli/schema"
	"github.com/spf13/cobra"
)

type CommandSchemas struct {
	Input  any
	Output any
}

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

func CommandSchemaRegistry() map[string]CommandSchemas {
	return map[string]CommandSchemas{
		"escape-cli me":                      {Output: v3.GetMe200Response{}},
		"escape-cli users me":                {Output: v3.GetMe200Response{}},
		"escape-cli users list":              {Output: []v3.ListUsers200ResponseInner{}},
		"escape-cli users get":               {Output: v3.GetUser200Response{}},
		"escape-cli users invite":            {Output: []v3.ListUsers200ResponseInner{}},
		"escape-cli roles list":              {Output: []v3.ListRoles200ResponseInner{}},
		"escape-cli roles get":               {Output: v3.CreateRole200Response{}},
		"escape-cli roles create":            {Input: v3.CreateRoleRequest{}, Output: v3.CreateRole200Response{}},
		"escape-cli roles update":            {Input: v3.UpdateRoleRequest{}, Output: v3.CreateRole200Response{}},
		"escape-cli projects list":           {Output: []v3.ListProjects200ResponseDataInner{}},
		"escape-cli projects get":            {Output: v3.CreateProject200Response{}},
		"escape-cli projects create":         {Input: v3.CreateProjectRequest{}, Output: v3.CreateProject200Response{}},
		"escape-cli projects update":         {Input: v3.UpdateProjectRequest{}, Output: v3.CreateProject200Response{}},
		"escape-cli integrations list":       {Output: []map[string]interface{}{}},
		"escape-cli integrations get":        {Output: map[string]interface{}{}},
		"escape-cli integrations create":     {Input: map[string]interface{}{}, Output: map[string]interface{}{}},
		"escape-cli integrations update":     {Input: map[string]interface{}{}, Output: map[string]interface{}{}},
		"escape-cli workflows list":          {Output: []v3.WorkflowSummarized{}},
		"escape-cli workflows get":           {Output: v3.CreateWorkflow200Response{}},
		"escape-cli workflows create":        {Input: v3.CreateWorkflowRequest{}, Output: v3.CreateWorkflow200Response{}},
		"escape-cli workflows update":        {Input: v3.UpdateWorkflowRequest{}, Output: v3.CreateWorkflow200Response{}},
		"escape-cli profiles list":           {Output: []v3.ProfileSummarized{}},
		"escape-cli profiles get":            {Output: v3.GetProfile200Response{}},
		"escape-cli profiles create-rest":    {Input: createRestProfileInput{}, Output: v3.GetProfile200Response{}},
		"escape-cli profiles create-webapp":  {Input: createWebappProfileInput{}, Output: v3.GetProfile200Response{}},
		"escape-cli profiles create-graphql": {Input: createGraphqlProfileInput{}, Output: v3.GetProfile200Response{}},
		"escape-cli profiles create-pentest-rest": {
			Input:  createPentestRestProfileInput{},
			Output: v3.GetProfile200Response{},
		},
		"escape-cli profiles create-pentest-graphql": {
			Input:  createPentestGraphqlProfileInput{},
			Output: v3.GetProfile200Response{},
		},
		"escape-cli profiles create-pentest-webapp": {
			Input:  createPentestWebappProfileInput{},
			Output: v3.GetProfile200Response{},
		},
		"escape-cli profiles update":               {Input: v3.UpdateProfileRequest{}, Output: v3.GetProfile200Response{}},
		"escape-cli profiles update-configuration": {Input: v3.UpdateProfileConfigurationRequest{}},
		"escape-cli issues list":                   {Output: []v3.IssueSummarized{}},
		"escape-cli issues get":                    {Output: v3.GetIssue200Response{}},
		"escape-cli issues list-activities":        {Output: []v3.ActivitySummarized{}},
		"escape-cli scans list":                    {Output: []v3.ScanSummarized{}},
		"escape-cli scans get":                     {Output: v3.StartScan200Response{}},
		"escape-cli scans start":                   {Output: v3.ScanDetailed1{}},
		"escape-cli scans watch":                   {Output: v3.ScanDetailed1{}},
		"escape-cli scans issues":                  {Output: []v3.IssueSummarized{}},
		"escape-cli scans targets":                 {Output: []v3.TargetDetailed{}},
		"escape-cli emails list":                   {Output: []v3.ScanEmailSummary{}},
		"escape-cli emails read":                   {Output: v3.ScanEmailDetails{}},
		"escape-cli emails wait":                   {Output: v3.ScanEmailDetails{}},
		"escape-cli authentications start":         {Input: v3.StartAuthenticationRequest{}, Output: v3.StartAuthentication200Response{}},
		"escape-cli authentications get":           {Output: v3.GetAuthentication200Response{}},
		"escape-cli jobs trigger-export":           {Output: v3.TriggerExport200Response{}},
		"escape-cli jobs get":                      {Output: v3.GetJob200Response{}},
		"escape-cli locations list":                {Output: []v3.LocationSummarized{}},
		"escape-cli locations get":                 {Output: v3.CreateLocation200Response{}},
		"escape-cli locations create":              {Input: v3.CreateLocationRequest{}, Output: v3.CreateLocation200Response{}},
		"escape-cli locations update":              {Input: v3.UpdateLocationRequest{}, Output: v3.CreateLocation200Response{}},
		"escape-cli assets list":                   {Output: []v3.AssetSummarized{}},
		"escape-cli assets get":                    {Output: v3.AssetDetailed{}},
		"escape-cli assets create":                 {Input: v3.CreateAssetRESTRequest{}, Output: v3.AssetDetailed{}},
		"escape-cli custom-rules list":             {Output: []v3.CustomRuleSummarized{}},
		"escape-cli custom-rules get":              {Output: v3.CreateCustomRule200Response{}},
		"escape-cli custom-rules create":           {Input: v3.CreateCustomRuleRequest{}, Output: v3.CreateCustomRule200Response{}},
		"escape-cli custom-rules update":           {Input: v3.UpdateCustomRuleRequest{}, Output: v3.CreateCustomRule200Response{}},
		"escape-cli tags list":                     {Output: []v3.TagDetail{}},
		"escape-cli tags get":                      {Output: v3.TagDetail{}},
		"escape-cli audit":                         {Output: []v3.AuditLogSummarized{}},
		"escape-cli issues comment":                {Input: v3.CreateAssetCommentRequest{}, Output: v3.CreateAssetComment200Response{}},
		"escape-cli capabilities":                  {Output: []commandCapabilitySchemaOutput{}},
	}
}

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
		capabilities = append(capabilities, CommandCapability{
			Path:         command.CommandPath(),
			Use:          command.Use,
			Short:        command.Short,
			Aliases:      command.Aliases,
			HasSub:       command.HasAvailableSubCommands(),
			HasFlags:     command.Flags().HasAvailableFlags(),
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
