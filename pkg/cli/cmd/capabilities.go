package cmd

import (
	"encoding/json"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	clischema "github.com/Escape-Technologies/cli/pkg/cli/schema"
	"github.com/spf13/cobra"
)

type commandSchemas struct {
	input  any
	output any
}

type commandCapability struct {
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

func schemaFor(v any) *clischema.JSONSchema {
	if v == nil {
		return nil
	}
	return clischema.Generate(v)
}

func commandSchemaRegistry() map[string]commandSchemas {
	return map[string]commandSchemas{
		"escape-cli me":                      {output: v3.GetMe200Response{}},
		"escape-cli users me":                {output: v3.GetMe200Response{}},
		"escape-cli users list":              {output: []v3.ListUsers200ResponseInner{}},
		"escape-cli users get":               {output: v3.GetUser200Response{}},
		"escape-cli users invite":            {output: []v3.ListUsers200ResponseInner{}},
		"escape-cli roles list":              {output: []v3.ListRoles200ResponseInner{}},
		"escape-cli roles get":               {output: v3.CreateRole200Response{}},
		"escape-cli roles create":            {input: v3.CreateRoleRequest{}, output: v3.CreateRole200Response{}},
		"escape-cli roles update":            {input: v3.UpdateRoleRequest{}, output: v3.CreateRole200Response{}},
		"escape-cli projects list":           {output: []v3.ListProjects200ResponseDataInner{}},
		"escape-cli projects get":            {output: v3.CreateProject200Response{}},
		"escape-cli projects create":         {input: v3.CreateProjectRequest{}, output: v3.CreateProject200Response{}},
		"escape-cli projects update":         {input: v3.UpdateProjectRequest{}, output: v3.CreateProject200Response{}},
		"escape-cli integrations list":       {output: []map[string]interface{}{}},
		"escape-cli integrations get":        {output: map[string]interface{}{}},
		"escape-cli integrations create":     {input: map[string]interface{}{}, output: map[string]interface{}{}},
		"escape-cli integrations update":     {input: map[string]interface{}{}, output: map[string]interface{}{}},
		"escape-cli workflows list":          {output: []v3.WorkflowSummarized{}},
		"escape-cli workflows get":           {output: v3.CreateWorkflow200Response{}},
		"escape-cli workflows create":        {input: v3.CreateWorkflowRequest{}, output: v3.CreateWorkflow200Response{}},
		"escape-cli workflows update":        {input: v3.UpdateWorkflowRequest{}, output: v3.CreateWorkflow200Response{}},
		"escape-cli profiles list":           {output: []v3.ProfileSummarized{}},
		"escape-cli profiles get":            {output: v3.GetProfile200Response{}},
		"escape-cli profiles create-rest":    {input: createRestProfileInput{}, output: v3.GetProfile200Response{}},
		"escape-cli profiles create-webapp":  {input: createWebappProfileInput{}, output: v3.GetProfile200Response{}},
		"escape-cli profiles create-graphql": {input: createGraphqlProfileInput{}, output: v3.GetProfile200Response{}},
		"escape-cli profiles create-pentest-rest": {
			input:  createPentestRestProfileInput{},
			output: v3.GetProfile200Response{},
		},
		"escape-cli profiles create-pentest-graphql": {
			input:  createPentestGraphqlProfileInput{},
			output: v3.GetProfile200Response{},
		},
		"escape-cli profiles create-pentest-webapp": {
			input:  createPentestWebappProfileInput{},
			output: v3.GetProfile200Response{},
		},
		"escape-cli profiles update":               {input: v3.UpdateProfileRequest{}, output: v3.GetProfile200Response{}},
		"escape-cli profiles update-configuration": {input: v3.UpdateProfileConfigurationRequest{}},
		"escape-cli issues list":                   {output: []v3.IssueSummarized{}},
		"escape-cli issues get":                    {output: v3.GetIssue200Response{}},
		"escape-cli issues list-activities":        {output: []v3.ActivitySummarized{}},
		"escape-cli scans list":                    {output: []v3.ScanSummarized{}},
		"escape-cli scans get":                     {output: v3.StartScan200Response{}},
		"escape-cli scans start":                   {output: v3.ScanDetailed1{}},
		"escape-cli scans watch":                   {output: v3.ScanDetailed1{}},
		"escape-cli scans issues":                  {output: []v3.IssueSummarized{}},
		"escape-cli scans targets":                 {output: []v3.TargetDetailed{}},
		"escape-cli emails list":                   {output: []v3.ScanEmailSummary{}},
		"escape-cli emails read":                   {output: v3.ScanEmailDetails{}},
		"escape-cli emails wait":                   {output: v3.ScanEmailDetails{}},
		"escape-cli authentications start":         {input: v3.StartAuthenticationRequest{}, output: v3.StartAuthentication200Response{}},
		"escape-cli authentications get":           {output: v3.GetAuthentication200Response{}},
		"escape-cli jobs trigger-export":           {output: v3.TriggerExport200Response{}},
		"escape-cli jobs get":                      {output: v3.GetJob200Response{}},
		"escape-cli locations list":                {output: []v3.LocationSummarized{}},
		"escape-cli locations get":                 {output: v3.CreateLocation200Response{}},
		"escape-cli locations create":              {input: v3.CreateLocationRequest{}, output: v3.CreateLocation200Response{}},
		"escape-cli locations update":              {input: v3.UpdateLocationRequest{}, output: v3.CreateLocation200Response{}},
		"escape-cli assets list":                   {output: []v3.AssetSummarized{}},
		"escape-cli assets get":                    {output: v3.AssetDetailed{}},
		"escape-cli assets create":                 {input: v3.CreateAssetRESTRequest{}, output: v3.AssetDetailed{}},
		"escape-cli custom-rules list":             {output: []v3.CustomRuleSummarized{}},
		"escape-cli custom-rules get":              {output: v3.CreateCustomRule200Response{}},
		"escape-cli custom-rules create":           {input: v3.CreateCustomRuleRequest{}, output: v3.CreateCustomRule200Response{}},
		"escape-cli custom-rules update":           {input: v3.UpdateCustomRuleRequest{}, output: v3.CreateCustomRule200Response{}},
		"escape-cli tags list":                     {output: []v3.TagDetail{}},
		"escape-cli tags get":                      {output: v3.TagDetail{}},
		"escape-cli audit":                         {output: []v3.AuditLogSummarized{}},
		"escape-cli issues comment":                {input: v3.CreateAssetCommentRequest{}, output: v3.CreateAssetComment200Response{}},
		"escape-cli capabilities":                  {output: []commandCapability{}},
	}
}

var capabilitiesCmd = &cobra.Command{
	Use:   "capabilities",
	Short: "Describe CLI commands in a machine-readable format",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		if out.Schema([]commandCapability{}) {
			return nil
		}

		registry := commandSchemaRegistry()
		capabilities := make([]commandCapability, 0)
		var walk func(command *cobra.Command)
		walk = func(command *cobra.Command) {
			if !command.IsAvailableCommand() || command.Hidden {
				return
			}
			schemas := registry[command.CommandPath()]
			inputSchema := schemaFor(schemas.input)
			outputSchema := schemaFor(schemas.output)
			capabilities = append(capabilities, commandCapability{
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
		walk(rootCmd)

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
