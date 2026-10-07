package cmd

import (
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func ptr[T any](v T) *T { return &v }

const (
	classSchema  = v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESCLASS_SCHEMA
	classNonSpec = v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESCLASS_HOST
)

func schemaAsset(id string, active bool, signed string) v3.ProfileExtraAsset {
	a := v3.ProfileExtraAsset{
		Id:       id,
		Name:     "schema-" + id,
		Class:    classSchema,
		IsActive: active,
	}
	if signed != "" {
		a.SignedUrl = ptr(signed)
	}

	return a
}

func TestApplyProfileUpdatePentestFlags(t *testing.T) {
	for name, value := range map[string]string{
		"mode":                   "STRICT",
		"rate-limit-req-per-sec": "5",
		"app-description":        "Payments API",
	} {
		resetFlag(t, profileUpdateCmd, name)
		if err := profileUpdateCmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
	}

	payload := map[string]interface{}{}
	applyProfileUpdatePentestFlags(profileUpdateCmd, payload)

	if payload["mode"] != "STRICT" {
		t.Fatalf("mode = %#v", payload["mode"])
	}

	if payload["rateLimitReqPerSec"] != 5 {
		t.Fatalf("rateLimitReqPerSec = %#v", payload["rateLimitReqPerSec"])
	}

	context, ok := payload["context"].(map[string]interface{})
	if !ok || context["appDescription"] != "Payments API" {
		t.Fatalf("context = %#v", payload["context"])
	}
}

func TestPrintProfilePentest(t *testing.T) {
	profile := &v3.GetProfile200Response{}
	profile.SetPentest(v3.ProfilePentest{
		Mode:               v3.ENUMPROPERTIESCONFIGURATIONPROPERTIESAUTOMATEDPENTESTINGPROPERTIESMULTIAGENTPENTESTPROPERTIESMODE_STRICT,
		RateLimitReqPerSec: 5,
		Context: v3.ProfilePentestContext{
			AppDescription: ptr("Payments API"),
		},
	})

	stdout, _, err := captureOutput(t, "pretty", func() error {
		printProfilePentest(profile)

		return nil
	})
	if err != nil {
		t.Fatalf("print: %v", err)
	}

	for _, want := range []string{"STRICT", "Payments API", "MODE"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestPickProfileSchema_NilProfile(t *testing.T) {
	t.Parallel()
	if _, err := pickProfileSchema(nil, ""); err == nil {
		t.Fatalf("expected error for nil profile, got nil")
	}
}

func TestPickProfileSchema_ActiveHappyPath(t *testing.T) {
	t.Parallel()
	profile := &v3.GetProfile200Response{
		ExtraAssets: []v3.ProfileExtraAsset{
			{Id: "a1", Class: classNonSpec},
			schemaAsset("s1", false, "https://old.example"),
			schemaAsset("s2", true, "https://active.example"),
		},
	}

	got, err := pickProfileSchema(profile, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil || got.Id != "s2" {
		t.Fatalf("expected active schema s2, got %+v", got)
	}

	if got.SignedUrl == nil || *got.SignedUrl != "https://active.example" {
		t.Fatalf("expected active signed URL to be preserved, got %+v", got.SignedUrl)
	}
}

func TestPickProfileSchema_NoActive(t *testing.T) {
	t.Parallel()
	profile := &v3.GetProfile200Response{
		ExtraAssets: []v3.ProfileExtraAsset{
			{Id: "a1", Class: classNonSpec},
			schemaAsset("s1", false, "https://old.example"),
		},
	}

	_, err := pickProfileSchema(profile, "")
	if err == nil {
		t.Fatalf("expected error when no active schema, got nil")
	}

	if !strings.Contains(err.Error(), "no active schema") {
		t.Fatalf("expected 'no active schema' error, got: %v", err)
	}
}

func TestPickProfileSchema_ByID(t *testing.T) {
	t.Parallel()
	profile := &v3.GetProfile200Response{
		ExtraAssets: []v3.ProfileExtraAsset{
			schemaAsset("s1", true, "https://active.example"),
			schemaAsset("s2", false, "https://old.example"),
		},
	}

	got, err := pickProfileSchema(profile, "s2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil || got.Id != "s2" {
		t.Fatalf("expected s2, got %+v", got)
	}
}

func TestPickProfileSchema_ByIDNoMatch(t *testing.T) {
	t.Parallel()
	profile := &v3.GetProfile200Response{
		ExtraAssets: []v3.ProfileExtraAsset{schemaAsset("s1", true, "")},
	}

	_, err := pickProfileSchema(profile, "nope")
	if err == nil {
		t.Fatalf("expected error for unknown schema id, got nil")
	}

	if !strings.Contains(err.Error(), "no schema asset with id nope") {
		t.Fatalf("expected 'no schema asset with id nope', got: %v", err)
	}
}

func TestPickProfileSchema_MultipleActive_FailsFast(t *testing.T) {
	t.Parallel()
	// Server invariant says "exactly one active SCHEMA". We explicitly refuse
	// to pick arbitrarily — caller should disambiguate via --schema-id.
	profile := &v3.GetProfile200Response{
		ExtraAssets: []v3.ProfileExtraAsset{
			schemaAsset("s1", true, "https://one.example"),
			schemaAsset("s2", true, "https://two.example"),
		},
	}

	_, err := pickProfileSchema(profile, "")
	if err == nil {
		t.Fatalf("expected error for multiple active schemas, got nil")
	}

	if !strings.Contains(err.Error(), "refusing to pick arbitrarily") {
		t.Fatalf("expected fail-fast error, got: %v", err)
	}
}

func TestPickProfileSchema_IgnoresNonSchemaClass(t *testing.T) {
	t.Parallel()
	// A non-SCHEMA asset that happens to be marked IsActive must be ignored
	// (defense against a regeneration / schema drift bug). The function
	// filters on class first, so we should still report "no active schema".
	profile := &v3.GetProfile200Response{
		ExtraAssets: []v3.ProfileExtraAsset{
			{Id: "a1", Class: classNonSpec, IsActive: true},
		},
	}

	_, err := pickProfileSchema(profile, "")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "no active schema") {
		t.Fatalf("expected 'no active schema' error, got: %v", err)
	}
}
