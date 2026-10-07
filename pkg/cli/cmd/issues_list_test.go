package cmd

import (
	"testing"
)

// saveIssueListFlags snapshots the globals read by issueListFilters and restores them.
func saveIssueListFlags(t *testing.T) {
	t.Helper()
	prevStatus := issueStatus
	prevSeverities := issueSeverity
	prevProfileIDs := profileIDs
	prevAssetIDs := assetIDs
	prevDomains := domains
	prevIssueIDs := issueIDs
	prevScanIDs := scanIDs
	prevTagsIDs := tagsIDs
	prevSearch := search
	prevJiraTicket := jiraTicket
	prevRisks := risks
	prevAssetClasses := assetClasses
	prevScannerKinds := issueScannerKinds
	prevNames := issueNames
	prevProjectIDs := issueProjectIDs
	prevTargetIDs := issueTargetIDs
	prevCategories := issueCategories
	prevAssetTypes := issueAssetTypes
	prevAssetStatuses := issueAssetStatuses
	prevSecurityTestUids := issueSecurityTestUids
	prevBlacklistedIDs := issueBlacklistedIDs
	prevBlacklistedNames := issueBlacklistedNames
	prevAiFalsePositive := issueAiFalsePositive
	prevAgentic := issueAgentic
	prevNoTags := issueNoTags
	prevDnf := issueDnf
	t.Cleanup(func() {
		issueStatus = prevStatus
		issueSeverity = prevSeverities
		profileIDs = prevProfileIDs
		assetIDs = prevAssetIDs
		domains = prevDomains
		issueIDs = prevIssueIDs
		scanIDs = prevScanIDs
		tagsIDs = prevTagsIDs
		search = prevSearch
		jiraTicket = prevJiraTicket
		risks = prevRisks
		assetClasses = prevAssetClasses
		issueScannerKinds = prevScannerKinds
		issueNames = prevNames
		issueProjectIDs = prevProjectIDs
		issueTargetIDs = prevTargetIDs
		issueCategories = prevCategories
		issueAssetTypes = prevAssetTypes
		issueAssetStatuses = prevAssetStatuses
		issueSecurityTestUids = prevSecurityTestUids
		issueBlacklistedIDs = prevBlacklistedIDs
		issueBlacklistedNames = prevBlacklistedNames
		issueAiFalsePositive = prevAiFalsePositive
		issueAgentic = prevAgentic
		issueNoTags = prevNoTags
		issueDnf = prevDnf
	})

	issueStatus = nil
	issueSeverity = nil
	profileIDs = nil
	assetIDs = nil
	domains = nil
	issueIDs = nil
	scanIDs = nil
	tagsIDs = nil
	search = ""
	jiraTicket = ""
	risks = nil
	assetClasses = nil
	issueScannerKinds = nil
	issueNames = nil
	issueProjectIDs = nil
	issueTargetIDs = nil
	issueCategories = nil
	issueAssetTypes = nil
	issueAssetStatuses = nil
	issueSecurityTestUids = nil
	issueBlacklistedIDs = nil
	issueBlacklistedNames = nil
	issueAiFalsePositive = ""
	issueAgentic = ""
	issueNoTags = false
	issueDnf = ""
}

func TestIssueListFiltersDefaultsToNilBooleans(t *testing.T) {
	saveIssueListFlags(t)

	filters := issueListFilters()
	if filters.AiFalsePositive != "" {
		t.Errorf("AiFalsePositive = %q, want empty when --ai-false-positive is unset", filters.AiFalsePositive)
	}

	if filters.Agentic != "" {
		t.Errorf("Agentic = %q, want empty when --agentic is unset", filters.Agentic)
	}

	if filters.NoTags != nil {
		t.Errorf("NoTags = %v, want nil when --no-tags is unset", *filters.NoTags)
	}

	if filters.Dnf != "" {
		t.Errorf("Dnf = %q, want empty when --dnf is unset", filters.Dnf)
	}
}

func TestIssueListFiltersCarryNewFlags(t *testing.T) {
	saveIssueListFlags(t)

	issueProjectIDs = []string{"00000000-0000-0000-0000-000000000001"}
	issueTargetIDs = []string{"00000000-0000-0000-0000-000000000002"}
	issueCategories = []string{"INJECTION"}
	issueAssetTypes = []string{"API"}
	issueAssetStatuses = []string{"EXPOSED"}
	issueSecurityTestUids = []string{"ISSUE_SQL_INJECTION"}
	issueBlacklistedIDs = []string{"00000000-0000-0000-0000-000000000003"}
	issueBlacklistedNames = []string{"SQL injection"}
	issueAiFalsePositive = "false"
	issueAgentic = "true"
	issueNoTags = true
	issueDnf = `{"and":[{"severity":"HIGH"}]}`

	filters := issueListFilters()
	if len(filters.ProjectIDs) != 1 || filters.ProjectIDs[0] != issueProjectIDs[0] {
		t.Errorf("ProjectIDs = %v, want %v", filters.ProjectIDs, issueProjectIDs)
	}

	if len(filters.TargetIDs) != 1 || filters.TargetIDs[0] != issueTargetIDs[0] {
		t.Errorf("TargetIDs = %v, want %v", filters.TargetIDs, issueTargetIDs)
	}

	if len(filters.Categories) != 1 || filters.Categories[0] != "INJECTION" {
		t.Errorf("Categories = %v, want [INJECTION]", filters.Categories)
	}

	if len(filters.AssetTypes) != 1 || filters.AssetTypes[0] != "API" {
		t.Errorf("AssetTypes = %v, want [API]", filters.AssetTypes)
	}

	if len(filters.AssetStatuses) != 1 || filters.AssetStatuses[0] != "EXPOSED" {
		t.Errorf("AssetStatuses = %v, want [EXPOSED]", filters.AssetStatuses)
	}

	if len(filters.SecurityTestUids) != 1 || filters.SecurityTestUids[0] != "ISSUE_SQL_INJECTION" {
		t.Errorf("SecurityTestUids = %v, want [ISSUE_SQL_INJECTION]", filters.SecurityTestUids)
	}

	if len(filters.BlacklistedIDs) != 1 || filters.BlacklistedIDs[0] != issueBlacklistedIDs[0] {
		t.Errorf("BlacklistedIDs = %v, want %v", filters.BlacklistedIDs, issueBlacklistedIDs)
	}

	if len(filters.BlacklistedNames) != 1 || filters.BlacklistedNames[0] != "SQL injection" {
		t.Errorf("BlacklistedNames = %v, want [SQL injection]", filters.BlacklistedNames)
	}

	if filters.AiFalsePositive != "false" {
		t.Errorf("AiFalsePositive = %q, want false", filters.AiFalsePositive)
	}

	if filters.Agentic != "true" {
		t.Errorf("Agentic = %q, want true", filters.Agentic)
	}

	if filters.NoTags == nil || !*filters.NoTags {
		t.Errorf("NoTags = %v, want true", filters.NoTags)
	}

	if filters.Dnf != issueDnf {
		t.Errorf("Dnf = %q, want %q", filters.Dnf, issueDnf)
	}
}

func TestIssueListFiltersTrimDnf(t *testing.T) {
	saveIssueListFlags(t)

	issueDnf = `  {"and":[{"severity":"HIGH"}]}  `

	filters := issueListFilters()
	if filters.Dnf != `{"and":[{"severity":"HIGH"}]}` {
		t.Errorf("Dnf = %q, want the trimmed expression", filters.Dnf)
	}
}
