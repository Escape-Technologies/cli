package cmd

import (
	"encoding/json"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func TestParseRemediationKind(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "", want: remediationKindRemediation},
		{in: "remediation", want: remediationKindRemediation},
		{in: "SUMMARY", want: remediationKindSummary},
		{in: " summary ", want: remediationKindSummary},
		{in: "other", wantErr: true},
	}

	for _, tc := range cases {
		got, err := parseRemediationKind(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseRemediationKind(%q) expected an error", tc.in)
			}

			continue
		}

		if err != nil {
			t.Errorf("parseRemediationKind(%q): %v", tc.in, err)
			continue
		}

		if got != tc.want {
			t.Errorf("parseRemediationKind(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBuildGenerateRemediationRequest(t *testing.T) {
	t.Run("defaults to remediation", func(t *testing.T) {
		request, err := buildGenerateRemediationRequest("", nil)
		if err != nil || request.GetKind() != v3.ENUMPROPERTIESKIND(remediationKindRemediation) {
			t.Fatalf("kind = %q err = %v", request.GetKind(), err)
		}
	})

	t.Run("flag wins over body", func(t *testing.T) {
		request, err := buildGenerateRemediationRequest("remediation", []byte(`{"kind":"summary"}`))
		if err != nil || request.GetKind() != v3.ENUMPROPERTIESKIND(remediationKindRemediation) {
			t.Fatalf("kind = %q err = %v", request.GetKind(), err)
		}
	})

	t.Run("body used when flag empty", func(t *testing.T) {
		request, err := buildGenerateRemediationRequest("", []byte(`{"kind":"summary"}`))
		if err != nil || request.GetKind() != v3.ENUMPROPERTIESKIND(remediationKindSummary) {
			t.Fatalf("kind = %q err = %v", request.GetKind(), err)
		}
	})

	t.Run("unknown body fields are preserved", func(t *testing.T) {
		request, err := buildGenerateRemediationRequest("", []byte(`{"stdinProbe":"body"}`))
		if err != nil {
			t.Fatalf("err = %v", err)
		}

		payload := marshalToMap(t, request)
		if payload["stdinProbe"] != "body" {
			t.Fatalf("additional property dropped: %v", payload)
		}

		if payload["kind"] != remediationKindRemediation {
			t.Fatalf("kind = %v, want the default", payload["kind"])
		}
	})

	t.Run("invalid body kind", func(t *testing.T) {
		if _, err := buildGenerateRemediationRequest("", []byte(`{"kind":"other"}`)); err == nil {
			t.Fatal("expected an invalid kind error")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		if _, err := buildGenerateRemediationRequest("", []byte(`{`)); err == nil {
			t.Fatal("expected an invalid JSON error")
		}
	})
}

func TestBuildSaveFeedbackRequest(t *testing.T) {
	t.Run("flag true", func(t *testing.T) {
		request, err := buildSaveFeedbackRequest("", "true", nil)
		if err != nil || request.GetKind() != v3.ENUMPROPERTIESKIND(remediationKindRemediation) {
			t.Fatalf("kind = %q err = %v", request.GetKind(), err)
		}

		if !request.HasFeedback() || !request.GetFeedback() {
			t.Fatalf("feedback = %v, want true", request.Feedback)
		}
	})

	t.Run("flag false", func(t *testing.T) {
		request, err := buildSaveFeedbackRequest("", "false", nil)
		if err != nil {
			t.Fatalf("err = %v", err)
		}

		if !request.HasFeedback() || request.GetFeedback() {
			t.Fatalf("feedback = %v, want false", request.Feedback)
		}
	})

	t.Run("flag null clears", func(t *testing.T) {
		request, err := buildSaveFeedbackRequest("", "null", nil)
		if err != nil {
			t.Fatalf("err = %v", err)
		}

		if request.Feedback != nil {
			t.Fatalf("feedback = %v, want nil (clear)", request.Feedback)
		}

		payload := marshalToMap(t, request)
		value, present := payload["feedback"]
		if !present || value != nil {
			t.Fatalf("feedback = %#v, want an explicit null", payload["feedback"])
		}
	})

	t.Run("flag wins over body", func(t *testing.T) {
		request, err := buildSaveFeedbackRequest("summary", "false", []byte(`{"kind":"remediation","feedback":true}`))
		if err != nil || request.GetKind() != v3.ENUMPROPERTIESKIND(remediationKindSummary) {
			t.Fatalf("kind = %q err = %v", request.GetKind(), err)
		}

		if !request.HasFeedback() || request.GetFeedback() {
			t.Fatalf("feedback = %v, want false", request.Feedback)
		}
	})

	t.Run("body feedback used when flag empty", func(t *testing.T) {
		request, err := buildSaveFeedbackRequest("", "", []byte(`{"feedback":true}`))
		if err != nil {
			t.Fatalf("err = %v", err)
		}

		if !request.HasFeedback() || !request.GetFeedback() {
			t.Fatalf("feedback = %v, want true", request.Feedback)
		}
	})

	t.Run("body null clears", func(t *testing.T) {
		request, err := buildSaveFeedbackRequest("", "", []byte(`{"feedback":null}`))
		if err != nil {
			t.Fatalf("err = %v", err)
		}

		payload := marshalToMap(t, request)
		value, present := payload["feedback"]
		if !present || value != nil {
			t.Fatalf("feedback = %#v, want an explicit null", payload["feedback"])
		}
	})

	t.Run("unknown body fields are preserved", func(t *testing.T) {
		request, err := buildSaveFeedbackRequest("", "true", []byte(`{"stdinProbe":"body"}`))
		if err != nil {
			t.Fatalf("err = %v", err)
		}

		payload := marshalToMap(t, request)
		if payload["stdinProbe"] != "body" {
			t.Fatalf("additional property dropped: %v", payload)
		}
	})

	t.Run("missing feedback is rejected", func(t *testing.T) {
		if _, err := buildSaveFeedbackRequest("", "", nil); err == nil {
			t.Fatal("expected a --feedback requirement error")
		}
	})

	t.Run("invalid flag feedback", func(t *testing.T) {
		if _, err := buildSaveFeedbackRequest("", "maybe", nil); err == nil {
			t.Fatal("expected an invalid --feedback error")
		}
	})
}

func TestIssueRemediationFromIssue(t *testing.T) {
	issue := &v3.GetIssue200Response{
		Id:                           "00000000-0000-0000-0000-000000000001",
		AiRemediationFramework:       "OPENAI",
		Remediation:                  v3.PtrString("Patch the query"),
		AiRemediationSummary:         v3.PtrString("Short summary"),
		AiRemediationFeedback:        v3.PtrBool(true),
		AiRemediationSummaryFeedback: v3.PtrBool(false),
	}

	got := issueRemediationFromIssue(issue)
	if got.IssueID != issue.Id {
		t.Errorf("IssueID = %q", got.IssueID)
	}

	if got.Framework != "OPENAI" {
		t.Errorf("Framework = %q", got.Framework)
	}

	if got.Remediation == nil || *got.Remediation != "Patch the query" {
		t.Errorf("Remediation = %v", got.Remediation)
	}

	if got.Summary == nil || *got.Summary != "Short summary" {
		t.Errorf("Summary = %v", got.Summary)
	}

	if got.RemediationFeedback == nil || !*got.RemediationFeedback {
		t.Errorf("RemediationFeedback = %v", got.RemediationFeedback)
	}

	if got.SummaryFeedback == nil || *got.SummaryFeedback {
		t.Errorf("SummaryFeedback = %v", got.SummaryFeedback)
	}
}

func TestIssueRemediationFromIssueOmitsUnsetFields(t *testing.T) {
	got := issueRemediationFromIssue(&v3.GetIssue200Response{Id: "00000000-0000-0000-0000-000000000001"})

	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, key := range []string{"remediation", "remediationFeedback", "summary", "summaryFeedback"} {
		if _, ok := payload[key]; ok {
			t.Errorf("expected %q to be omitted, got %v", key, payload[key])
		}
	}
}
