package cmd

// betaMCPToolNote is the sentence appended to the tools/list description of a Beta
// command. The catalog is what a model keeps in context for every tool, so a Beta
// surface has to say so there, not only in `--help`.
const betaMCPToolNote = "Beta: the public API operation behind this tool is in beta and may change."

// betaMCPTools lists the MCP tools whose underlying public-API operations carry the
// OpenAPI "Beta" tag. The list mirrors the spec surfaces declared in
// services/public-api/src/v3/routes/beta-policy.ts: a command wrapping a Beta operation
// is added here, and the CLI help text of the matching command group says Beta too.
var betaMCPTools = map[string]string{
	// Regression testing, gated by REGRESSION_TESTING.
	"escape-cli regression-tests list":    betaMCPToolNote,
	"escape-cli regression-tests get":     betaMCPToolNote,
	"escape-cli regression-tests create":  betaMCPToolNote,
	"escape-cli regression-tests update":  betaMCPToolNote,
	"escape-cli regression-tests delete":  betaMCPToolNote,
	"escape-cli regression-tests run":     betaMCPToolNote,
	"escape-cli regression-tests stop":    betaMCPToolNote,
	"escape-cli regression-tests history": betaMCPToolNote,
	"escape-cli regression-tests answer":  betaMCPToolNote,

	// Async jobs, exports and their artefacts.
	"escape-cli jobs trigger-export": betaMCPToolNote,
	"escape-cli jobs get":            betaMCPToolNote,
	"escape-cli jobs list":           betaMCPToolNote,

	// Continuous pentest, gated by CONTINUOUS_PENTEST.
	"escape-cli profiles continuous-pentest create":  betaMCPToolNote,
	"escape-cli profiles continuous-pentest update":  betaMCPToolNote,
	"escape-cli profiles continuous-pentest enable":  betaMCPToolNote,
	"escape-cli profiles continuous-pentest disable": betaMCPToolNote,

	// Target reachability validation: the release notes list it as coming
	// next, not as shipped.
	"escape-cli validations start": betaMCPToolNote,
	"escape-cli validations get":   betaMCPToolNote,

	// Copilot chats: announced, but the API contract is still moving.
	"escape-cli ai-chats list":     betaMCPToolNote,
	"escape-cli ai-chats get":      betaMCPToolNote,
	"escape-cli ai-chats messages": betaMCPToolNote,
	"escape-cli ai-chats create":   betaMCPToolNote,
	"escape-cli ai-chats message":  betaMCPToolNote,

	// AI pentest profile creation.
	"escape-cli profiles create-ai-pentest": betaMCPToolNote,

	// Only APK creation is Beta (MOBILE_APP_PENTEST): the other asset types are not.
	"escape-cli assets create": "Beta: creating an APK asset is in beta and may change.",

	// Scan inbox emails.
	"escape-cli emails list": betaMCPToolNote,
	"escape-cli emails read": betaMCPToolNote,
	"escape-cli emails wait": betaMCPToolNote,

	// The variables vault, gated by VARIABLES_VAULT.
	"escape-cli vault variables list":   betaMCPToolNote,
	"escape-cli vault variables create": betaMCPToolNote,
	"escape-cli vault variables delete": betaMCPToolNote,
	"escape-cli vault variables rotate": betaMCPToolNote,
	"escape-cli vault list":             betaMCPToolNote,
	"escape-cli vault link":             betaMCPToolNote,
	"escape-cli vault entries":          betaMCPToolNote,

	// The knowledge file library, not GA.
	"escape-cli knowledge-files list":   betaMCPToolNote,
	"escape-cli knowledge-files upload": betaMCPToolNote,
	"escape-cli knowledge-files delete": betaMCPToolNote,

	// Issue retests, gated by ISSUE_RETEST.
	"escape-cli retests list":  betaMCPToolNote,
	"escape-cli retests get":   betaMCPToolNote,
	"escape-cli retests start": betaMCPToolNote,

	// AI pentest agents and their reasoning logs.
	"escape-cli scans agents":    betaMCPToolNote,
	"escape-cli scans reasoning": betaMCPToolNote,

	// Organization settings writes. GET settings is not Beta.
	"escape-cli org-settings update": betaMCPToolNote,
}

// betaMCPToolNoteFor returns the Beta sentence of a command path, or "" when the
// command does not wrap a Beta API operation.
func betaMCPToolNoteFor(path string) string {
	return betaMCPTools[path]
}
