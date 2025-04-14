package scans

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"
)

func (s *Scan) String() string {
	var sb strings.Builder
	
	w := tabwriter.NewWriter(&sb, 0, 0, 3, ' ', tabwriter.AlignRight)
	
	scoreStr := "N/A"
	if s.Score != nil {
		scoreStr = fmt.Sprintf("%.2f", *s.Score)
	}
	
	finishedStr := "N/A"
	if s.FinishedAt != nil {
		finishedStr = s.FinishedAt.Format(time.RFC3339)
	}
	
	fmt.Fprintf(w, "ID\tStatus\tCreatedAt\tUpdatedAt\tFinishedAt\tProgress\tScore\tInitiator\t\n")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%.0f%%\t%s\t%s\t\n", 
		s.Id, 
		s.Status, 
		s.CreatedAt.Format(time.RFC3339), 
		s.UpdatedAt.Format(time.RFC3339), 
		finishedStr, 
		s.ProgressRatio*100,
		scoreStr, 
		s.Initiator)
	w.Flush()	
	
	return sb.String()
}

func FormatReportsTable(reports []*Report) string {
	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 3, ' ', tabwriter.AlignRight)
	
	fmt.Fprintf(w, "ID\tIssues\tSeverity\tIgnored\tCategory\tTitleOnFail\tType\t\n")
	
	for _, r := range reports {
		fmt.Fprintf(w, "%s\t%d\t%s\t%t\t%s\t%s\t%s\t\n", 
			r.Id, 
			len(r.Issues),
			r.Severity, 
			r.Ignored, 
			r.Test.Category, 
			r.Test.Meta.TitleOnFail, 
			r.Test.Meta.Type)
	}
	
	w.Flush()
	return sb.String()
}

func FormatIssuesTable(issues []*Issue) string {
	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 3, ' ', tabwriter.AlignRight)
	
	fmt.Fprintf(w, "ID\tIgnored\tFirstSeenScanId\tLastSeenScanId\tSeverity\tRisks\t\n")
	for _, issue := range issues {
		risks := make([]string, 0, len(issue.Risks))
		for _, risk := range issue.Risks {
			risks = append(risks, risk.Kind)
		}
		fmt.Fprintf(w, "%s\t%t\t%s\t%s\t%s\t%s\t\n", 
			issue.Id, 
			issue.Ignored, 
			issue.FirstSeenScanId, 
			issue.LastSeenScanId, 
			issue.Severity, 
			strings.Join(risks, ", "),
		)
	}
	w.Flush()
	return sb.String()
}

func (s *ScanExchangeArchive) String() string {
	return s.Archive
}

func (s *ScanEvent) String() string {
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s", 
		s.Id, 
		s.CreatedAt.Format(time.RFC3339),
		s.Level,
		s.Title,
		s.Description,
	)
}