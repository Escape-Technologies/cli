package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// serverInstructionsLimit is the budget for ServerInstructions. Clients copy
// this text into the model context on initialize, so it stays a short rules
// card rather than a second system prompt.
const serverInstructionsLimit = 2000

// JSON-RPC ids for the in-process guidance catalog. They only need to be
// distinct so a response can be tied back to the request that produced it.
const (
	promptsListJSONRPCID = 1
	promptsGetJSONRPCID  = 2
)

// ServerInstructions is returned by MCP initialize. External clients never
// see the in-app Copilot system prompt, so the rules that stop a partial
// configuration write from wiping a profile live here. Tool names match
// buildMCPToolName.
const ServerInstructions = `Escape.tech API. Answer from tool results. Do not invent data, URLs, or errors.

Configuration:
- profiles_update_configuration replaces the whole configuration unless merge is true. This server defaults merge to true: send only the fields that change. Nested objects deep-merge; arrays and scalars replace; omitted keys stay. Set merge false only for a full replace: call profiles_get first, then send the complete configuration. A partial body with merge false resets omitted sections to defaults.
- Read-only, safe, or production-safe mode is scanner configuration, not RBAC. Call profiles_update_configuration with {"configuration":{"mode":"read_only"}} and leave merge true.
- Authentication, headers, users, credentials, scope, crawling, network, and security tests on a profile also go through profiles_update_configuration. Do not call authentications_start or authentications_get unless the user asks to run or inspect an authentication check.
- If a call returns a schema error, call escape_get_tool_spec for that tool and retry once.

Coverage:
- For coverage, scanner users, or whether requests succeeded, call scans_coverage. When complete is true, trust overall and byUser even if targetsTruncated is true. Route counts are overall.statuses out of overall.targets. A user's counts are byUser.<user>.statuses. Never sum users: a route OK for several users counts once per user. Do not count the targets sample, and do not use scans_reasoning or one scans_targets page for those totals.
- The scans_targets size argument is a total cap, not a page size. Omit size to fetch every page, or use scans_coverage.

Safety:
- Confirm with the user before deletes, role changes, revoking access, or removing users. Show the exact action and wait.
- Call scans_start, then poll scans_get until status is FINISHED, FAILED, or CANCELED.`

// RegisterPrompts adds the workflow prompts external MCP clients can render.
// Each prompt is a short tool sequence. Names and arguments are part of the
// server contract.
func RegisterPrompts(server *mcpserver.MCPServer) {
	server.AddPrompt(mcpgo.NewPrompt(
		"triage_findings",
		mcpgo.WithPromptDescription("Triage open findings: list issues, then read evidence before recommending action."),
	), triageFindings)

	server.AddPrompt(mcpgo.NewPrompt(
		"fix_finding",
		mcpgo.WithPromptDescription("Explain a finding and its fix from the issue and its scan events."),
		mcpgo.WithArgument("issue_id", mcpgo.RequiredArgument(), mcpgo.ArgumentDescription("Issue id from issues_list or issues_get.")),
	), fixFinding)

	server.AddPrompt(mcpgo.NewPrompt(
		"setup_dast_profile",
		mcpgo.WithPromptDescription("Create a DAST profile for a URL without wiping configuration."),
		mcpgo.WithArgument("url", mcpgo.RequiredArgument(), mcpgo.ArgumentDescription("Application URL to scan.")),
	), setupDASTProfile)

	server.AddPrompt(mcpgo.NewPrompt(
		"start_and_summarize_pentest",
		mcpgo.WithPromptDescription("Start a scan and summarize issues and coverage. Polls scans_get instead of watching."),
		mcpgo.WithArgument("profile_id", mcpgo.RequiredArgument(), mcpgo.ArgumentDescription("Profile id to scan.")),
	), startAndSummarizePentest)

	server.AddPrompt(mcpgo.NewPrompt(
		"explain_coverage",
		mcpgo.WithPromptDescription("Explain scan coverage from scans_coverage, not from a targets sample."),
		mcpgo.WithArgument("scan_id", mcpgo.RequiredArgument(), mcpgo.ArgumentDescription("Scan id to explain.")),
	), explainCoverage)
}

func triageFindings(_ context.Context, _ mcpgo.GetPromptRequest) (*mcpgo.GetPromptResult, error) {
	return userPrompt(`Triage open security findings.
1. Call issues_list. Filter with status OPEN, and severity HIGH or CRITICAL when the user asked for the important ones.
2. For each issue you will discuss, call issues_get_with_events for that issue so the reply includes the latest scan events.
3. Summarize severity, status, asset, and the evidence in those events. Do not invent findings.
4. Do not call issues_update or issues_comment unless the user asked to change the issue.`), nil
}

func fixFinding(_ context.Context, request mcpgo.GetPromptRequest) (*mcpgo.GetPromptResult, error) {
	issueID, err := requiredPromptArg(request, "issue_id")
	if err != nil {
		return nil, err
	}

	return userPrompt(fmt.Sprintf(`Explain how to fix issue %s.
1. Call issues_get_with_events for %q.
2. Describe the vulnerability, the proof in the events, and a concrete fix. Do not claim it is already fixed.
3. Only if the user asked to record the fix, call issues_comment or issues_update. Confirm before a status change that closes the issue or accepts the risk.`, issueID, issueID)), nil
}

func setupDASTProfile(_ context.Context, request mcpgo.GetPromptRequest) (*mcpgo.GetPromptResult, error) {
	targetURL, err := requiredPromptArg(request, "url")
	if err != nil {
		return nil, err
	}

	return userPrompt(fmt.Sprintf(`Set up a DAST profile for %s.
1. Call assets_list with search %q. If no asset matches, call assets_create. If that call returns a schema error, call escape_get_tool_spec for assets_create and retry once.
2. Create the profile: profiles_create_graphql for a GraphQL API, profiles_create_webapp for a browser app, otherwise profiles_create_rest. The body needs assetId and name. Set mode to read_only when the user asked for safe or production-safe testing. If that call returns a schema error, call escape_get_tool_spec for that tool and retry once.
3. Authentication, headers, scope, or security tests are configuration. Call profiles_update_configuration with merge true (the default) and only the configuration fields that change. Never send a partial body with merge false.
4. Do not call authentications_start unless the user asked to run an authentication check.`, targetURL, targetURL)), nil
}

func startAndSummarizePentest(_ context.Context, request mcpgo.GetPromptRequest) (*mcpgo.GetPromptResult, error) {
	profileID, err := requiredPromptArg(request, "profile_id")
	if err != nil {
		return nil, err
	}

	return userPrompt(fmt.Sprintf(`Start a scan for profile %s and summarize it.
1. Call profiles_get for profile %q and confirm the profile exists.
2. Call scans_start for profile %q.
3. Poll scans_get with the returned scan id until status is FINISHED, FAILED, or CANCELED.
4. On FINISHED, call scans_issues and scans_coverage for that scan id. Summarize status, issue counts, and coverage from scans_coverage overall.statuses out of overall.targets. Do not count a scans_targets sample.`, profileID, profileID, profileID)), nil
}

func explainCoverage(_ context.Context, request mcpgo.GetPromptRequest) (*mcpgo.GetPromptResult, error) {
	scanID, err := requiredPromptArg(request, "scan_id")
	if err != nil {
		return nil, err
	}

	return userPrompt(fmt.Sprintf(`Explain coverage for scan %s.
1. Call scans_coverage for scan %q.
2. When complete is true, answer from overall and byUser even if targetsTruncated is true.
3. How many routes are OK, covered, or failing: overall.statuses out of overall.targets.
4. For one scanner user: byUser.<user>.statuses. Never add users together. A route OK for several users counts once per user.
5. Do not count the targets sample. Do not call scans_reasoning, and do not answer from one scans_targets page. The size argument is a total cap, not a page size.`, scanID, scanID)), nil
}

func requiredPromptArg(request mcpgo.GetPromptRequest, name string) (string, error) {
	value := strings.TrimSpace(request.Params.Arguments[name])
	if value == "" {
		return "", fmt.Errorf("missing required argument %q", name)
	}

	return value, nil
}

// GuidanceTexts is the copy an MCP client shows the model: server instructions,
// each prompt description, each argument description, and every prompt rendered
// with sample arguments. A catalog test uses it to reject tool names that are
// not registered.
func GuidanceTexts() ([]string, error) {
	server := mcpserver.NewMCPServer("guidance", "test")
	RegisterPrompts(server)

	listed, err := guidanceCall(server, map[string]any{
		"jsonrpc": "2.0",
		"id":      promptsListJSONRPCID,
		"method":  "prompts/list",
	})
	if err != nil {
		return nil, err
	}

	var prompts struct {
		Result struct {
			Prompts []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Arguments   []struct {
					Name        string `json:"name"`
					Description string `json:"description"`
					Required    bool   `json:"required"`
				} `json:"arguments"`
			} `json:"prompts"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(listed, &prompts); err != nil {
		return nil, fmt.Errorf("parse prompts/list: %w", err)
	}

	if prompts.Error != nil {
		return nil, errors.New(prompts.Error.Message)
	}

	if len(prompts.Result.Prompts) == 0 {
		return nil, errors.New("no prompts registered")
	}

	texts := []string{ServerInstructions}
	for _, prompt := range prompts.Result.Prompts {
		texts = append(texts, prompt.Description)
		args := map[string]string{}
		for _, arg := range prompt.Arguments {
			texts = append(texts, arg.Description)
			if arg.Required {
				// Hyphenated so the placeholder itself is not a tool-shaped token.
				args[arg.Name] = "sample-value"
			}
		}

		params := map[string]any{"name": prompt.Name}
		if len(args) > 0 {
			params["arguments"] = args
		}

		body, err := guidanceCall(server, map[string]any{
			"jsonrpc": "2.0",
			"id":      promptsGetJSONRPCID,
			"method":  "prompts/get",
			"params":  params,
		})
		if err != nil {
			return nil, err
		}

		var rendered struct {
			Result struct {
				Messages []struct {
					Content struct {
						Text string `json:"text"`
					} `json:"content"`
				} `json:"messages"`
			} `json:"result"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &rendered); err != nil {
			return nil, fmt.Errorf("parse prompts/get %s: %w", prompt.Name, err)
		}

		if rendered.Error != nil {
			return nil, fmt.Errorf("prompts/get %s: %s", prompt.Name, rendered.Error.Message)
		}

		if len(rendered.Result.Messages) == 0 {
			return nil, fmt.Errorf("prompts/get %s returned no messages", prompt.Name)
		}

		for _, message := range rendered.Result.Messages {
			texts = append(texts, message.Content.Text)
		}
	}

	return texts, nil
}

func guidanceCall(server *mcpserver.MCPServer, payload map[string]any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode guidance request: %w", err)
	}

	resp := server.HandleMessage(context.Background(), json.RawMessage(raw))
	if resp == nil {
		return nil, errors.New("nil MCP response")
	}

	body, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("encode guidance response: %w", err)
	}

	return body, nil
}

func userPrompt(text string) *mcpgo.GetPromptResult {
	return mcpgo.NewGetPromptResult("", []mcpgo.PromptMessage{
		mcpgo.NewPromptMessage(mcpgo.RoleUser, mcpgo.NewTextContent(text)),
	})
}
