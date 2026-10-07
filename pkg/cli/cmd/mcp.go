package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	clischema "github.com/Escape-Technologies/cli/pkg/cli/schema"
	"github.com/Escape-Technologies/cli/pkg/env"
	climcp "github.com/Escape-Technologies/cli/pkg/mcp"
	"github.com/Escape-Technologies/cli/pkg/version"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// maxMCPToolDescriptionRunes caps Short plus the first paragraph of Long.
// The catalog is what an MCP client keeps in context for every tool.
const maxMCPToolDescriptionRunes = 400

var (
	// Bracket lists are how Go prints a slice of enum constants inside a
	// flag usage string: `filter by status: [OPEN RESOLVED]`.
	bracketEnumPattern = regexp.MustCompile(`\[([A-Z0-9_ ]+)\]`)
	upperEnumToken     = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	// headlinePattern matches a Title-Case help line that restates Short,
	// such as "List Security Issues - Query Your Vulnerability Database".
	headlinePattern = regexp.MustCompile(`^[A-Z][^.]* - [A-Z]`)
)

// headlineRuneLimit is the length under which a single-line first paragraph
// of Long is treated as a headline rather than the tool description.
const headlineRuneLimit = 100

const defaultMCPServePort = 8080

// flagEnumAnnotation is a pflag annotation listing the allowed values of a
// string flag. schemaForFlag copies it into the MCP property enum.
const flagEnumAnnotation = "enum"

// mcpSkipAnnotation marks a flag the MCP catalog must not offer. The value
// is mcpSkipValue. Commands set it where the flag is declared, so a blocking
// flag (watch until a remote job finishes) stays off the tool without the
// catalog hard-coding that command's path.
const (
	mcpSkipAnnotation = "mcp"
	mcpSkipValue      = "skip"
)

// mcpRequiredAnnotation marks a flag the CLI leaves optional but the MCP tool
// must receive. Omitting it would widen the call to every resource. The
// executor refuses an empty value before it starts the subprocess.
const mcpRequiredAnnotation = "mcp-required"

// mcpDefaultAnnotation marks a boolean flag whose MCP schema default is true
// while the CLI default stays false. schemaForFlag copies it onto the
// property. buildMCPTool copies that schema default onto FlagBinding so an
// omitted argument still sends the flag.
const mcpDefaultAnnotation = "mcp-default"

var (
	mcpServePort             int
	mcpServePublicAPIURL     string
	mcpServeOAuthIssuerURL   string
	mcpServeOAuthResourceURL string
	mcpServeOAuthPrivateKey  string
)

var mcpCmd = &cobra.Command{
	Use:    "mcp",
	Short:  "MCP server commands",
	Hidden: true,
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the embedded MCP server",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		toolSpecs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
		if err != nil {
			return fmt.Errorf("failed to build MCP tool catalog: %w", err)
		}

		mode := climcp.ModeFromEnv(climcp.IntentModeCompactOnly)
		var classifier climcp.Classifier
		if mode == climcp.IntentModeOn {
			classifier, err = climcp.NewClassifierFromEnv()
			if err != nil {
				// Server documents IntentModeOn + nil Classifier as a valid
				// fallback to compact-only behavior. Don't trade a recoverable
				// config issue for downtime — log and continue.
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: MCP classifier disabled: %v\n", err)
				classifier = nil
			}
		}

		server := climcp.NewServer(climcp.ServerOptions{
			Version:             version.GetVersion().DisplayVersion(),
			Port:                mcpServePort,
			PublicAPIURL:        resolveMCPPublicAPIURL(),
			Tools:               toolSpecs,
			IntentMode:          mode,
			Classifier:          classifier,
			IssuerURL:           mcpServeOAuthIssuerURL,
			ResourceURL:         mcpServeOAuthResourceURL,
			OAuthPrivateKeyPath: mcpServeOAuthPrivateKey,
		})

		return server.Serve(cmd.Context())
	},
}

func init() {
	mcpServeCmd.Flags().IntVar(&mcpServePort, "port", defaultMCPServePort, "port to listen on")
	mcpServeCmd.Flags().StringVar(
		&mcpServePublicAPIURL,
		"public-api-url",
		"",
		"public API base URL (defaults to ESCAPE_API_URL)",
	)
	mcpServeCmd.Flags().StringVar(
		&mcpServeOAuthIssuerURL,
		"oauth-issuer-url",
		"",
		"OAuth 2.1 issuer (e.g. https://app.escape.tech); empty disables the OAuth flow",
	)
	mcpServeCmd.Flags().StringVar(
		&mcpServeOAuthResourceURL,
		"oauth-resource-url",
		"",
		"OAuth protected resource URL (e.g. https://mcp.escape.tech/mcp); empty disables the OAuth flow",
	)
	mcpServeCmd.Flags().StringVar(
		&mcpServeOAuthPrivateKey,
		"oauth-private-key",
		"",
		"path to the RSA private key PEM used to decrypt JWE authorization codes; generated at the path if missing, or ephemeral when empty",
	)
	mcpCmd.AddCommand(mcpServeCmd)
	rootCmd.AddCommand(mcpCmd)
}

// destructiveMCPCommands are CLI commands whose MCP tools delete data or
// bulk-mutate assets. Membership is explicit: appearing in
// CommandSchemaRegistry does not make a command destructive, and a command
// in this set is not exposed without the confirm gate applied in buildMCPTool.
var destructiveMCPCommands = map[string]struct{}{
	"escape-cli profiles delete":         {},
	"escape-cli assets delete":           {},
	"escape-cli assets bulk-delete":      {},
	"escape-cli assets bulk-update":      {},
	"escape-cli issues bulk-update":      {},
	"escape-cli tags delete":             {},
	"escape-cli custom-rules delete":     {},
	"escape-cli locations delete":        {},
	"escape-cli workflows delete":        {},
	"escape-cli roles unbind":            {},
	"escape-cli integrations delete":     {},
	"escape-cli regression-tests delete": {},
}

func isDestructiveMCPCommand(path string) bool {
	_, ok := destructiveMCPCommands[path]
	return ok
}

func buildMCPToolSpecs(
	root *cobra.Command,
	registry map[string]CommandSchemas,
) ([]climcp.ToolSpec, error) {
	capabilities := BuildCommandCapabilities(root, registry)
	commands := indexCommands(root)
	toolSpecs := make([]climcp.ToolSpec, 0, len(capabilities))

	for _, capability := range capabilities {
		if capability.HasSub || skipMCPCommand(capability.Path) {
			continue
		}

		if _, allowed := registry[capability.Path]; !allowed {
			continue
		}

		command := commands[capability.Path]
		if command == nil {
			return nil, fmt.Errorf("missing cobra command for %q", capability.Path)
		}

		description := mcpCatalogDescription(capability.Path, command)
		destructive := isDestructiveMCPCommand(capability.Path)
		tool, flagBindings, positionalArgs, bodyProperty, err := buildMCPTool(capability, command, description, destructive)
		if err != nil {
			return nil, err
		}

		spec := climcp.ToolSpec{
			Name:           tool.Name,
			Path:           capability.Path,
			Description:    description,
			Tool:           tool,
			Command:        strings.Fields(capability.Path)[1:],
			PositionalArgs: positionalArgs,
			FlagBindings:   flagBindings,
			BodyProperty:   bodyProperty,
			Destructive:    destructive,
		}
		if capability.Path == "escape-cli scans reasoning" ||
			capability.Path == "escape-cli scans coverage" {
			spec.ExecutionTimeout = climcp.ReasoningToolExecutionTimeout
		}

		if pagedListCommand(command) {
			spec.Description = pagedMCPDescription(description)
			spec.DefaultArgs = map[string]any{"size": climcp.DefaultListPageSize}
			spec.Tool.Description = spec.Description
		}

		applyBoundedMCPCommand(&spec)
		toolSpecs = append(toolSpecs, spec)
	}

	return toolSpecs, nil
}

func buildMCPTool(
	capability CommandCapability,
	command *cobra.Command,
	description string,
	destructive bool,
) (mcpgo.Tool, []climcp.FlagBinding, []string, string, error) {
	properties := map[string]any{}
	required := []string{}
	flagBindings := make([]climcp.FlagBinding, 0)
	bodyProperty := ""

	positionalArgs := parsePositionalArgs(capability.Use, capability.HasInSchema)
	for _, name := range positionalArgs {
		properties[name] = map[string]any{
			"type":        "string",
			"description": fmt.Sprintf("Positional argument %s.", name),
		}
		required = append(required, name)
	}

	// LocalFlags is this command's own flags, including ones registered with
	// PersistentFlags. Flags() omits those until parse time and, once merged,
	// also contains inherited root flags (verbose, output, input-schema).
	command.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if skipMCPFlag(command, flag.Name) {
			return
		}

		property := normalizePropertyName(flag.Name)
		// confirm is an MCP-only gate. A CLI flag of the same name must not
		// be forwarded, or the model could satisfy the guard by passing a flag.
		if destructive && property == climcp.ConfirmProperty {
			return
		}

		if _, exists := properties[property]; exists {
			property = "flag_" + property
		}

		schema, kind := schemaForFlag(flag)
		binding := climcp.FlagBinding{
			Property: property,
			FlagName: flag.Name,
			Kind:     kind,
			Required: flagMarkedMCPRequired(flag),
		}
		// A schema default is what hosted clients should send when they omit
		// the property. Copy it onto the binding so the executor does not
		// fall through to the CLI default. The CLI flag's own DefValue is
		// unchanged.
		if def, ok := schema["default"]; ok {
			binding.Default = def
		}

		properties[property] = schema
		flagBindings = append(flagBindings, binding)
		if isRequiredFlag(flag) || binding.Required {
			required = append(required, property)
		}
	})

	if destructive {
		properties[climcp.ConfirmProperty] = map[string]any{
			"type":        "boolean",
			"description": climcp.ConfirmPropertyDescription,
		}
		required = append(required, climcp.ConfirmProperty)
	}

	if capability.InputSchema != nil {
		bodyProperty = "body"
		properties[bodyProperty] = capability.InputSchema
	}

	input := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
	inputSchema, err := json.Marshal(input)
	if err != nil {
		return mcpgo.Tool{}, nil, nil, "", fmt.Errorf("failed to marshal input schema for %q: %w", capability.Path, err)
	}

	// Body schemas hoist large enums into their own $defs. A $ref of
	// "#/$defs/..." resolves at the tool document root, so those maps have
	// to move up or the full schema escape_get_tool_spec returns is invalid.
	inputSchema, err = liftNestedDefs(inputSchema)
	if err != nil {
		return mcpgo.Tool{}, nil, nil, "", fmt.Errorf("failed to hoist enums for %q: %w", capability.Path, err)
	}

	// Use NewToolWithRawSchema so the default structured InputSchema (Type:"object")
	// is not set alongside RawInputSchema — the library refuses to marshal a tool
	// that has both set (errToolSchemaConflict).
	tool := mcpgo.NewToolWithRawSchema(buildMCPToolName(capability.Path), description, inputSchema)
	// MCP clients require a top-level object outputSchema. Array commands are
	// returned as {"items":[...]} by wrapStructuredPayload, so advertise that
	// wrapped object. A bounded command can withhold its declared schema when
	// the success document is not the only payload it prints.
	if bounded, ok := boundedMCPCommands[capability.Path]; !ok || !bounded.skipOutputSchema {
		if outputSchema := mcpOutputSchema(capability.OutputSchema, pagedListCommand(command)); outputSchema != nil {
			encoded, err := json.Marshal(outputSchema)
			if err != nil {
				return mcpgo.Tool{}, nil, nil, "", fmt.Errorf("failed to marshal output schema for %q: %w", capability.Path, err)
			}

			encoded, err = liftNestedDefs(encoded)
			if err != nil {
				return mcpgo.Tool{}, nil, nil, "", fmt.Errorf("failed to hoist output enums for %q: %w", capability.Path, err)
			}

			tool.RawOutputSchema = encoded
		}
	}

	return tool, flagBindings, positionalArgs, bodyProperty, nil
}

// pagedListExtraProperties is nextCursor plus totalCount.
const pagedListExtraProperties = 2

// mcpOutputSchema is the JSON Schema advertised for one MCP tool.
// Clients require a top-level object. Array commands are returned as
// {"items":[...]} by wrapStructuredPayload, so the advertised schema is
// {type:object, properties:{items:{type:array, items:<item schema>}}}.
// Object schemas pass through. Nil means the command declares no output.
// A paged list returns {items, nextCursor, totalCount}. Those fields are
// added when the declared schema does not already carry them.
func mcpOutputSchema(schema *clischema.JSONSchema, paged bool) *clischema.JSONSchema {
	if schema == nil {
		return nil
	}

	if schema.Type == "array" {
		schema = &clischema.JSONSchema{
			Type: "object",
			// Items keep $ref into this schema's $defs. The wrapper is the
			// document root, so the map has to move with them.
			Defs: schema.Defs,
			Properties: map[string]*clischema.JSONSchema{
				"items": {
					Type:  "array",
					Items: schema.Items,
				},
			},
		}
	}

	if !paged || schema.Type != "object" {
		return schema
	}

	if schema.Properties["nextCursor"] != nil && schema.Properties["totalCount"] != nil {
		return schema
	}

	properties := make(map[string]*clischema.JSONSchema, len(schema.Properties)+pagedListExtraProperties)
	for name, property := range schema.Properties {
		properties[name] = property
	}

	if properties["nextCursor"] == nil {
		properties["nextCursor"] = &clischema.JSONSchema{Type: "string"}
	}

	if properties["totalCount"] == nil {
		properties["totalCount"] = &clischema.JSONSchema{Type: "integer"}
	}

	copied := *schema
	copied.Properties = properties

	return &copied
}

func indexCommands(root *cobra.Command) map[string]*cobra.Command {
	commands := map[string]*cobra.Command{}

	var walk func(command *cobra.Command)
	walk = func(command *cobra.Command) {
		if !command.IsAvailableCommand() || command.Hidden {
			return
		}

		commands[command.CommandPath()] = command
		for _, child := range command.Commands() {
			walk(child)
		}
	}

	walk(root)

	return commands
}

// pagedListCommand reports whether pageFlags.bind marked the command as a
// cursor page. Flag names are not enough: retests get has --size and --cursor
// and is not a list. scans targets has --size as a total cap and is unmarked.
func pagedListCommand(command *cobra.Command) bool {
	return command != nil && command.Annotations[pagedListAnnotation] == "true"
}

func skipMCPCommand(path string) bool {
	name := strings.TrimPrefix(path, "escape-cli ")

	return name == "help" ||
		name == "help-all" ||
		name == "completion" ||
		name == "capabilities" ||
		name == "version" ||
		name == "mcp"
}

// markMCPSkip hides a flag from the MCP catalog. The flag must already be
// registered on flags. A missing flag is a declaration bug, so init panics.
func markMCPSkip(flags *pflag.FlagSet, name string) {
	if err := flags.SetAnnotation(name, mcpSkipAnnotation, []string{mcpSkipValue}); err != nil {
		panic(err)
	}
}

// skipMCPFlag drops flags the MCP catalog must not offer. A flag annotated
// mcp=skip is hidden on whatever command declared it. Every flag named watch
// is hidden too: watch blocks until a remote job finishes and cannot complete
// inside the tool budget, and a new command must not have to join a list to
// stay safe. file, out, and output-file-type read or write arbitrary paths
// on the machine hosting the server, so they are hidden on every command,
// including ones not registered yet. Bounded commands hide the flags their
// catalog injects.
func skipMCPFlag(command *cobra.Command, name string) bool {
	switch name {
	case "verbose", "output", "input-schema", "help", "watch", "file", "out", "output-file-type":
		return true
	}

	if bounded, ok := boundedMCPCommands[command.CommandPath()]; ok && slices.Contains(bounded.hiddenFlags, name) {
		return true
	}

	return flagMarkedMCPSkip(commandFlag(command, name))
}

func commandFlag(command *cobra.Command, name string) *pflag.Flag {
	if flag := command.Flags().Lookup(name); flag != nil {
		return flag
	}

	if flag := command.LocalFlags().Lookup(name); flag != nil {
		return flag
	}

	return command.PersistentFlags().Lookup(name)
}

func flagMarkedMCPSkip(flag *pflag.Flag) bool {
	return flag != nil && slices.Contains(flag.Annotations[mcpSkipAnnotation], mcpSkipValue)
}

// markMCPRequired keeps the CLI flag optional and tells the MCP catalog the
// tool call must include a non-empty value. The flag must already be
// registered. A missing flag is a declaration bug, so init panics.
func markMCPRequired(flags *pflag.FlagSet, name string) {
	if err := flags.SetAnnotation(name, mcpRequiredAnnotation, []string{"true"}); err != nil {
		panic(err)
	}
}

func flagMarkedMCPRequired(flag *pflag.Flag) bool {
	return flag != nil && slices.Contains(flag.Annotations[mcpRequiredAnnotation], "true")
}

// markMCPDefault keeps the CLI flag default and tells schemaForFlag to
// advertise true as the MCP default. The flag must already be registered.
// A missing flag is a declaration bug, so init panics.
func markMCPDefault(flags *pflag.FlagSet, name string) {
	if err := flags.SetAnnotation(name, mcpDefaultAnnotation, []string{"true"}); err != nil {
		panic(err)
	}
}

// mcpPollHint is appended to the CLI short text so a model polls the bounded
// get tool instead of retrying a start/trigger call that already returned.
var mcpPollHint = map[string]string{
	"escape-cli scans start":           " Returns when the scan has been created. Poll scans_get with the returned scan id until it finishes.",
	"escape-cli scans get":             " Poll scans_get with a scan id until the scan reaches a terminal status.",
	"escape-cli jobs trigger-export":   " Returns when the export job has been created. Poll jobs_get with the returned job id until it finishes.",
	"escape-cli jobs get":              " Poll jobs_get with a job id until the job reaches a terminal status.",
	"escape-cli authentications start": " Returns when the authentication check has been created. Poll authentications_get with the returned id until it finishes.",
	"escape-cli authentications get":   " Poll authentications_get with an authentication id until the check reaches a terminal status.",
}

// mcpCatalogDescription is the tools/list text. It keeps Short plus the first
// paragraph of Long, then adds the operational sentence the model must not
// lose: a poll hint, or a bounded command's own description. Paged commands
// add their page contract separately so that sentence stays inside the rune cap.
func mcpCatalogDescription(path string, command *cobra.Command) string {
	if bounded, ok := boundedMCPCommands[path]; ok && bounded.describe != nil {
		return bounded.describe(command)
	}

	base := mcpToolDescription(command)
	if hint, ok := mcpPollHint[path]; ok {
		return joinWithinRunes(base, hint, maxMCPToolDescriptionRunes)
	}

	return base
}

// pagedMCPDescription keeps the help text and the one-page contract. The
// contract is the part that must survive the rune cap.
func pagedMCPDescription(help string) string {
	help = strings.TrimSpace(help)
	if help != "" && !strings.HasSuffix(help, ".") {
		help += "."
	}

	note := strings.TrimSpace(climcp.ListPageToolDescription(""))

	return joinWithinRunes(help, note, maxMCPToolDescriptionRunes)
}

// boundedMCPCommand is the whole MCP special case for one CLI path: flags the
// model must not set, args the catalog injects, whether the declared output
// schema is withheld, and the tools/list description. emails wait is the only
// entry. Another bounded tool belongs here, not in a fresh path compare.
type boundedMCPCommand struct {
	hiddenFlags      []string
	forcedArgs       []string
	skipOutputSchema bool
	describe         func(*cobra.Command) string
}

func emailsWaitCatalogDescription(command *cobra.Command) string {
	short := ""
	if command != nil {
		short = strings.TrimSpace(command.Short)
	}

	return fmt.Sprintf(
		"%s Waits at most %s. If no email has arrived, the result is %q with after set to the latest createdAt and seenIds listing the ids already seen at that time. Call emails_wait again and pass that after value (and each seenIds entry as seen_id), or poll emails_list.",
		short,
		climcp.EmailsWaitTimeout,
		emailWaitPendingMessage,
	)
}

var boundedMCPCommands = map[string]boundedMCPCommand{
	"escape-cli emails wait": {
		hiddenFlags: []string{"timeout", emailWaitPendingFlag},
		forcedArgs: []string{
			"--timeout", climcp.EmailsWaitTimeout.String(),
			"--" + emailWaitPendingFlag + "=true",
		},
		skipOutputSchema: true,
		describe:         emailsWaitCatalogDescription,
	},
}

// applyBoundedMCPCommand copies a bounded command's injected args onto the spec.
func applyBoundedMCPCommand(spec *climcp.ToolSpec) {
	bounded, ok := boundedMCPCommands[spec.Path]
	if !ok {
		return
	}

	spec.ForcedArgs = append([]string(nil), bounded.forcedArgs...)
}

// isRequiredFlag mirrors cobra.ValidateRequiredFlags' annotation check so flags
// marked with MarkFlagRequired are advertised as required in the MCP tool input
// schema instead of only failing in the spawned subprocess.
func isRequiredFlag(flag *pflag.Flag) bool {
	values, ok := flag.Annotations[cobra.BashCompOneRequiredFlag]
	return ok && len(values) > 0 && values[0] == "true"
}

func buildMCPToolName(path string) string {
	name := strings.TrimPrefix(path, "escape-cli ")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	return name
}

func parsePositionalArgs(use string, hasBody bool) []string {
	parts := strings.Fields(use)
	if len(parts) <= 1 {
		return nil
	}

	names := make([]string, 0, len(parts)-1)
	for _, part := range parts[1:] {
		candidate := strings.Trim(part, "<>[]")
		candidate = strings.TrimSuffix(candidate, "...")
		if candidate == "" || candidate == "flags" {
			continue
		}

		if hasBody && strings.Contains(candidate, ".json") {
			continue
		}

		names = append(names, normalizePropertyName(candidate))
	}

	return names
}

func normalizePropertyName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, ".", "_")

	return value
}

// mcpToolDescription joins Cobra Short with the first paragraph of Long.
// Later paragraphs (examples, flag tables) stay out of the catalog. The
// combined text is capped so one help page cannot dominate tools/list.
func mcpToolDescription(command *cobra.Command) string {
	if command == nil {
		return ""
	}

	short := strings.TrimSpace(command.Short)
	paragraph := firstParagraph(command.Long)
	switch {
	case paragraph == "" || strings.EqualFold(paragraph, short):
		return truncateRunes(short, maxMCPToolDescriptionRunes)
	case short == "":
		return truncateRunes(paragraph, maxMCPToolDescriptionRunes)
	default:
		return truncateRunes(short+"\n"+paragraph, maxMCPToolDescriptionRunes)
	}
}

// firstParagraph returns the first block of Long that tells the model something
// Short does not. A single-line headline (under headlineRuneLimit, or a
// "Title - Subtitle" line) is skipped so the following paragraph is used.
func firstParagraph(long string) string {
	paragraphs := helpParagraphs(long)
	if len(paragraphs) == 0 {
		return ""
	}

	if isHeadline(paragraphs[0]) {
		if len(paragraphs) == 1 {
			return ""
		}

		return paragraphs[1].text
	}

	return paragraphs[0].text
}

type helpParagraph struct {
	text   string
	single bool
}

func helpParagraphs(long string) []helpParagraph {
	long = strings.ReplaceAll(long, "\r\n", "\n")
	long = strings.TrimSpace(long)
	if long == "" {
		return nil
	}

	var paragraphs []helpParagraph
	var lines []string
	flush := func() {
		if len(lines) == 0 {
			return
		}

		paragraphs = append(paragraphs, helpParagraph{
			text:   strings.Join(strings.Fields(strings.Join(lines, " ")), " "),
			single: len(lines) == 1,
		})
		lines = nil
	}
	for _, line := range strings.Split(long, "\n") {
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}

		lines = append(lines, line)
	}

	flush()

	return paragraphs
}

func isHeadline(paragraph helpParagraph) bool {
	if !paragraph.single {
		return false
	}

	if len([]rune(paragraph.text)) < headlineRuneLimit {
		return true
	}

	return headlinePattern.MatchString(paragraph.text)
}

// truncatedEllipsis is appended when truncateRunes cuts a string. Reserving
// its length keeps the marker inside the rune budget.
const truncatedEllipsis = "..."

func truncateRunes(text string, maxRunes int) string {
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}

	if maxRunes <= len(truncatedEllipsis) {
		return string(runes[:maxRunes])
	}

	cut := maxRunes - len(truncatedEllipsis)
	for i := cut; i > cut-40 && i > 0; i-- {
		if runes[i] == ' ' {
			cut = i
			break
		}
	}

	if cut <= 0 {
		cut = maxRunes - len(truncatedEllipsis)
	}

	return strings.TrimRight(string(runes[:cut]), " ") + truncatedEllipsis
}

// joinWithinRunes appends suffix to prefix inside maxRunes. When they do not
// fit, prefix is shortened and suffix is kept. Poll hints and the page
// contract are suffixes: losing them hides the rule the model has to follow.
func joinWithinRunes(prefix, suffix string, maxRunes int) string {
	suffix = strings.TrimSpace(suffix)
	prefix = strings.TrimSpace(prefix)
	if suffix == "" {
		return truncateRunes(prefix, maxRunes)
	}

	if prefix == "" {
		return truncateRunes(suffix, maxRunes)
	}

	combined := prefix + " " + suffix
	if len([]rune(combined)) <= maxRunes {
		return combined
	}

	room := maxRunes - len([]rune(suffix)) - 1
	if room < 1 {
		return truncateRunes(suffix, maxRunes)
	}

	return truncateRunes(prefix, room) + " " + suffix
}

func schemaForFlag(flag *pflag.Flag) (map[string]any, string) {
	description := flag.Usage
	schema, kind := flagTypeSchema(flag.Value.Type(), description)
	if kind == "string" || kind == "stringSlice" {
		enum := flag.Annotations[flagEnumAnnotation]
		if len(enum) == 0 {
			enum = enumFromUsage(description)
		}

		applyFlagEnum(schema, kind, enum)
	}

	if kind == "bool" && slices.Contains(flag.Annotations[mcpDefaultAnnotation], "true") {
		schema["default"] = true
	}

	return schema, kind
}

func flagTypeSchema(valueType, description string) (map[string]any, string) {
	switch valueType {
	case "bool":
		return map[string]any{
			"type":        "boolean",
			"description": description,
		}, "bool"
	case "int", "int64":
		return map[string]any{
			"type":        "integer",
			"description": description,
		}, "int"
	case "stringSlice", "stringArray":
		return map[string]any{
			"type":        "array",
			"description": description,
			"items": map[string]any{
				"type": "string",
			},
		}, "stringSlice"
	default:
		return map[string]any{
			"type":        "string",
			"description": description,
		}, "string"
	}
}

func applyFlagEnum(schema map[string]any, kind string, enum []string) {
	if len(enum) == 0 {
		return
	}

	switch kind {
	case "stringSlice":
		items, _ := schema["items"].(map[string]any)
		if items == nil {
			items = map[string]any{"type": "string"}
			schema["items"] = items
		}

		items["enum"] = enum
	case "string":
		schema["enum"] = enum
	}
}

// liftNestedDefs moves every nested $defs map onto the document root.
// JSON Schema resolves "#/$defs/name" against that root. Schemas with no
// nested map are returned unchanged, so unrelated documents stay byte-stable.
func liftNestedDefs(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return raw, nil
	}

	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("decode schema document: %w", err)
	}

	root, ok := doc.(map[string]any)
	if !ok {
		return raw, nil
	}

	collected := map[string]any{}
	moved := false
	var walk func(any, bool)
	walk = func(node any, isRoot bool) {
		switch typed := node.(type) {
		case map[string]any:
			if defs, ok := typed["$defs"].(map[string]any); ok && !isRoot {
				for name, def := range defs {
					if _, exists := collected[name]; !exists {
						collected[name] = def
					}
				}

				delete(typed, "$defs")
				moved = true
			}

			for key, child := range typed {
				if key == "$defs" {
					continue
				}

				walk(child, false)
			}
		case []any:
			for _, child := range typed {
				walk(child, false)
			}
		}
	}
	walk(root, true)
	if !moved {
		return raw, nil
	}

	existing, _ := root["$defs"].(map[string]any)
	if existing == nil {
		existing = map[string]any{}
	}

	for name, def := range collected {
		if _, exists := existing[name]; !exists {
			existing[name] = def
		}
	}

	root["$defs"] = existing
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(root); err != nil {
		return nil, fmt.Errorf("encode schema document: %w", err)
	}

	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// enumFromUsage lifts a Go slice print (`[OPEN RESOLVED]`) out of a cobra usage
// string. That shape is how flags print a generated Allowed*EnumValues list.
// A colon-led example (`sort direction: asc, desc`, `scanner type: BLST_REST,
// ...`) is not a closed enum: the help text is often a sample, and a closed
// list would reject the rest of the API's values.
func enumFromUsage(usage string) []string {
	return bracketEnum(usage)
}

// bracketEnumSubmatches is the full match plus bracketEnumPattern's one group.
const bracketEnumSubmatches = 2

// minBracketEnumTokens is the smallest list treated as a closed enum.
const minBracketEnumTokens = 2

func bracketEnum(usage string) []string {
	var best []string
	for _, match := range bracketEnumPattern.FindAllStringSubmatch(usage, -1) {
		if len(match) < bracketEnumSubmatches {
			continue
		}

		tokens := strings.Fields(match[1])
		if len(tokens) < minBracketEnumTokens {
			continue
		}

		ok := true
		for _, token := range tokens {
			if !upperEnumToken.MatchString(token) {
				ok = false
				break
			}
		}

		if ok && len(tokens) > len(best) {
			best = tokens
		}
	}

	return best
}

func resolveMCPPublicAPIURL() string {
	if mcpServePublicAPIURL != "" {
		return mcpServePublicAPIURL
	}

	apiURL, err := env.GetAPIURL()
	if err != nil {
		return ""
	}

	return apiURL.String()
}
