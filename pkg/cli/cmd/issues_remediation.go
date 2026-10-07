package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

const (
	remediationKindRemediation = "remediation"
	remediationKindSummary     = "summary"
)

var (
	remediationGenerateKind string
	remediationFeedbackKind string
	remediationFeedback     string
)

// IssueRemediation is the stable CLI view of an issue's AI remediation and
// summary, returned by `issues remediation get`.
type IssueRemediation struct {
	IssueID             string  `json:"issueId"`
	Framework           string  `json:"framework"`
	Remediation         *string `json:"remediation,omitempty"`
	RemediationFeedback *bool   `json:"remediationFeedback,omitempty"`
	Summary             *string `json:"summary,omitempty"`
	SummaryFeedback     *bool   `json:"summaryFeedback,omitempty"`
}

var issuesRemediationCmd = &cobra.Command{
	Use:   "remediation",
	Short: "Generate, read and rate AI remediation for an issue",
	Long: `Manage AI Remediation - Generate and Rate Fix Guidance

AI remediation is generated asynchronously. Trigger it with generate, read the
result back with get, and record whether it was useful with feedback.

KINDS:
  remediation - the full issue-panel remediation (default)
  summary     - the short overview summary`,
}

var issuesRemediationGenerateCmd = &cobra.Command{
	Use:   "generate issue-id",
	Short: "Generate AI remediation for an issue",
	Long: `Generate AI Remediation - Schedule Fix Guidance

Trigger the asynchronous generation of AI remediation for an issue. Generation
only schedules a job: read the result back with 'escape-cli issues remediation
get <issue-id>'.

Generating the full remediation clears the previous remediation and its feedback
immediately. Generating the summary replaces the previous summary, and resets
its rating, only once the generation job succeeds. Identical requests repeated
within a short window are ignored and report generated=false.

The same request can be sent as JSON on stdin ({"kind":"remediation"}); the
--kind flag wins when both are set.`,
	Example: `  # Generate the full remediation
  escape-cli issues remediation generate <issue-id>

  # Generate the short overview summary
  escape-cli issues remediation generate <issue-id> --kind summary

  # Read the result back once generation finishes
  escape-cli issues remediation get <issue-id>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.InputSchema(v3.GenerateIssueAiRemediationRequest{}) {
			return nil
		}

		if out.Schema(v3.GenerateIssueAiRemediation200Response{}) {
			return nil
		}

		body, err := readPipedStdin(cmd.InOrStdin())
		if err != nil {
			return err
		}

		request, err := buildGenerateRemediationRequest(remediationGenerateKind, body)
		if err != nil {
			return err
		}

		issueID := args[0]
		result, err := escape.GenerateIssueAiRemediation(cmd.Context(), issueID, *request)
		if err != nil {
			return fmt.Errorf("unable to generate AI remediation for issue %s: %w", issueID, err)
		}

		out.Table(result, func() []string {
			res := []string{"KIND\tGENERATED\tCLEARED"}
			res = append(res, fmt.Sprintf("%s\t%t\t%t", result.GetKind(), result.GetGenerated(), result.GetCleared()))

			return res
		})

		pretty := fmt.Sprintf("Requested %s AI remediation for issue %s", result.GetKind(), issueID)
		if !result.GetGenerated() {
			pretty = fmt.Sprintf("An identical %s generation is already in flight for issue %s", result.GetKind(), issueID)
		}

		out.Log(pretty)

		return nil
	},
}

var issuesRemediationGetCmd = &cobra.Command{
	Use:   "get issue-id",
	Short: "Show the AI remediation and summary of an issue",
	Long: `Get AI Remediation - Read Fix Guidance

Read the AI remediation and summary already stored on an issue. There is no
dedicated GET endpoint: the fields come from GET /issues/{issueId}.`,
	Example: `  # Show the remediation and summary
  escape-cli issues remediation get <issue-id>

  # Machine-readable output
  escape-cli issues remediation get <issue-id> -o json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(IssueRemediation{}) {
			return nil
		}

		issueID := args[0]
		issue, err := escape.GetIssue(cmd.Context(), issueID)
		if err != nil {
			return fmt.Errorf("unable to get issue %s: %w", issueID, err)
		}

		if issue == nil {
			return fmt.Errorf("unable to get issue %s: empty response", issueID)
		}

		remediation := issueRemediationFromIssue(issue)

		out.Table(remediation, func() []string {
			res := []string{"FIELD\tVALUE"}
			res = append(res,
				"Issue ID\t"+remediation.IssueID,
				"Framework\t"+formatOptional(remediation.Framework),
				"Remediation\t"+formatOptionalPtr(remediation.Remediation),
				"Remediation feedback\t"+formatOptionalBool(remediation.RemediationFeedback),
				"Summary\t"+formatOptionalPtr(remediation.Summary),
				"Summary feedback\t"+formatOptionalBool(remediation.SummaryFeedback),
			)

			return res
		})

		return nil
	},
}

var issuesRemediationFeedbackCmd = &cobra.Command{
	Use:   "feedback issue-id",
	Short: "Rate the AI remediation of an issue",
	Long: `Rate AI Remediation - Send Usefulness Feedback

Record whether an AI remediation artefact was useful. Use --kind to choose the
full issue-panel remediation or the short overview summary. Pass
--feedback null to clear a previous rating.

The same request can be sent as JSON on stdin
({"kind":"remediation","feedback":true}); flags win when both are set.`,
	Example: `  # Rate the full remediation as useful
  escape-cli issues remediation feedback <issue-id> --feedback true

  # Rate the summary as not useful
  escape-cli issues remediation feedback <issue-id> --kind summary --feedback false

  # Clear a previous rating
  escape-cli issues remediation feedback <issue-id> --feedback null`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.InputSchema(v3.SaveIssueAiRemediationFeedbackRequest{}) {
			return nil
		}

		if out.Schema(v3.SaveIssueAiRemediationFeedback200Response{}) {
			return nil
		}

		body, err := readPipedStdin(cmd.InOrStdin())
		if err != nil {
			return err
		}

		request, err := buildSaveFeedbackRequest(remediationFeedbackKind, remediationFeedback, body)
		if err != nil {
			return err
		}

		issueID := args[0]
		result, err := escape.SaveIssueAiRemediationFeedback(cmd.Context(), issueID, *request)
		if err != nil {
			return fmt.Errorf("unable to save AI remediation feedback for issue %s: %w", issueID, err)
		}

		out.Table(result, func() []string {
			res := []string{"KIND\tSAVED"}
			res = append(res, fmt.Sprintf("%s\t%t", result.GetKind(), result.GetSaved()))

			return res
		})

		pretty := fmt.Sprintf("Saved %s feedback for issue %s", result.GetKind(), issueID)
		if request.Feedback == nil {
			pretty = fmt.Sprintf("Cleared %s feedback for issue %s", result.GetKind(), issueID)
		}

		out.Log(pretty)

		return nil
	},
}

// parseRemediationKind normalizes and validates an AI remediation kind. An
// empty value defaults to the full remediation, matching the API default.
func parseRemediationKind(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", remediationKindRemediation:
		return remediationKindRemediation, nil
	case remediationKindSummary:
		return remediationKindSummary, nil
	default:
		return "", fmt.Errorf("invalid kind %q; valid values: %s, %s", value, remediationKindRemediation, remediationKindSummary)
	}
}

// buildGenerateRemediationRequest merges the --kind flag with an optional JSON
// body from stdin. The flag wins. Unknown body fields are preserved on the
// generated request so a body produced by an MCP client is not dropped.
func buildGenerateRemediationRequest(kindFlag string, body []byte) (*v3.GenerateIssueAiRemediationRequest, error) {
	request := v3.NewGenerateIssueAiRemediationRequest()

	if len(bytes.TrimSpace(body)) > 0 {
		if err := json.Unmarshal(body, request); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
	}

	if err := applyRemediationKind(request, request.Kind); err != nil {
		return nil, err
	}

	if strings.TrimSpace(kindFlag) != "" {
		parsed, err := parseRemediationKind(kindFlag)
		if err != nil {
			return nil, err
		}

		request.SetKind(v3.ENUMPROPERTIESKIND(parsed))
	}

	return request, nil
}

// buildSaveFeedbackRequest merges the --kind/--feedback flags with an optional
// JSON body from stdin. Flags win. An explicit null feedback (from the flag or
// the body) clears the rating; a missing feedback is an error.
func buildSaveFeedbackRequest(kindFlag, feedbackFlag string, body []byte) (*v3.SaveIssueAiRemediationFeedbackRequest, error) {
	request := v3.NewSaveIssueAiRemediationFeedbackRequest()
	var feedback *bool
	feedbackSet := false

	if len(bytes.TrimSpace(body)) > 0 {
		if err := json.Unmarshal(body, request); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}

		// The generated struct cannot tell an explicit null from an absent
		// feedback: both leave Feedback nil. Read the raw key to detect a
		// clear request.
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(body, &raw); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}

		if rawFeedback, ok := raw["feedback"]; ok {
			parsed, set, err := parseRemediationFeedbackRaw(rawFeedback)
			if err != nil {
				return nil, err
			}

			feedback, feedbackSet = parsed, set
		}
	}

	if err := applyRemediationKind(request, request.Kind); err != nil {
		return nil, err
	}

	if strings.TrimSpace(kindFlag) != "" {
		parsed, err := parseRemediationKind(kindFlag)
		if err != nil {
			return nil, err
		}

		request.SetKind(v3.ENUMPROPERTIESKIND(parsed))
	}

	if flagFeedback, flagSet, err := parseRemediationFeedbackFlag(feedbackFlag); err != nil {
		return nil, err
	} else if flagSet {
		feedback, feedbackSet = flagFeedback, true
	}

	if !feedbackSet {
		return nil, errors.New("--feedback is required (true, false, or null)")
	}

	if feedback == nil {
		// A nil Feedback pointer is dropped by the generated ToMap, so the
		// explicit null that clears a rating is carried through
		// AdditionalProperties instead.
		if request.AdditionalProperties == nil {
			request.AdditionalProperties = map[string]any{}
		}

		request.AdditionalProperties["feedback"] = nil
	} else {
		request.SetFeedback(*feedback)
	}

	return request, nil
}

// applyRemediationKind validates the kind carried on a request body and
// defaults it to the full remediation when absent.
func applyRemediationKind(request interface{ SetKind(v3.ENUMPROPERTIESKIND) }, kind *v3.ENUMPROPERTIESKIND) error {
	if kind == nil {
		request.SetKind(v3.ENUMPROPERTIESKIND(remediationKindRemediation))

		return nil
	}

	parsed, err := parseRemediationKind(string(*kind))
	if err != nil {
		return err
	}

	request.SetKind(v3.ENUMPROPERTIESKIND(parsed))

	return nil
}

// parseRemediationFeedbackRaw parses a JSON feedback value. An absent value
// (empty raw message) returns set=false; an explicit null returns nil, true.
func parseRemediationFeedbackRaw(raw json.RawMessage) (*bool, bool, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, false, nil
	}

	if bytes.Equal(trimmed, []byte("null")) {
		return nil, true, nil
	}

	var value bool
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return nil, false, fmt.Errorf("invalid feedback %s; valid values: true, false, null", trimmed)
	}

	return v3.PtrBool(value), true, nil
}

// parseRemediationFeedbackFlag parses the --feedback flag. An empty flag
// returns set=false; "null" or "clear" returns nil, true.
func parseRemediationFeedbackFlag(value string) (*bool, bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return nil, false, nil
	case "true":
		return v3.PtrBool(true), true, nil
	case "false":
		return v3.PtrBool(false), true, nil
	case "null", "clear":
		return nil, true, nil
	default:
		return nil, false, fmt.Errorf("invalid --feedback %q; valid values: true, false, null", value)
	}
}

// issueRemediationFromIssue extracts the AI remediation fields from an issue
// detail response into the stable CLI DTO.
func issueRemediationFromIssue(issue *v3.GetIssue200Response) IssueRemediation {
	result := IssueRemediation{
		IssueID:   issue.GetId(),
		Framework: issue.GetAiRemediationFramework(),
	}

	if remediation, ok := issue.GetRemediationOk(); ok {
		result.Remediation = remediation
	}

	if feedback, ok := issue.GetAiRemediationFeedbackOk(); ok {
		result.RemediationFeedback = feedback
	}

	if summary, ok := issue.GetAiRemediationSummaryOk(); ok {
		result.Summary = summary
	}

	if feedback, ok := issue.GetAiRemediationSummaryFeedbackOk(); ok {
		result.SummaryFeedback = feedback
	}

	return result
}

func formatOptionalPtr(value *string) string {
	if value == nil {
		return "-"
	}

	return formatOptional(*value)
}

func formatOptionalBool(value *bool) string {
	if value == nil {
		return "-"
	}

	return strconv.FormatBool(*value)
}

func init() {
	issuesRemediationGenerateCmd.Flags().StringVar(&remediationGenerateKind, "kind", "", "AI remediation kind: remediation (full) or summary (default: remediation)")

	issuesRemediationFeedbackCmd.Flags().StringVar(&remediationFeedbackKind, "kind", "", "AI remediation kind: remediation (full) or summary (default: remediation)")
	issuesRemediationFeedbackCmd.Flags().StringVar(&remediationFeedback, "feedback", "", "usefulness rating: true, false, or null to clear (required)")

	issuesRemediationCmd.AddCommand(issuesRemediationGenerateCmd, issuesRemediationGetCmd, issuesRemediationFeedbackCmd)
	issuesCmd.AddCommand(issuesRemediationCmd)
}
