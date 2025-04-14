package scans

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func GetScan(ctx context.Context, id openapi_types.UUID) (*Scan, error) {
	client, err := v2.NewAPIClient()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	resp, err := client.GetScanWithResponse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get scans: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unable to get scans: status code %d", resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, errors.New("unable to get scans: no data received")
	}

	return &Scan{
		Id:            resp.JSON200.Id,
		Status:        string(resp.JSON200.Status),
		CreatedAt:     resp.JSON200.CreatedAt,
		UpdatedAt:     resp.JSON200.UpdatedAt,
		FinishedAt:    resp.JSON200.FinishedAt,
		ProgressRatio: float64(resp.JSON200.ProgressRatio),
		Score:     resp.JSON200.Score,
		Initiator: resp.JSON200.Initiator,
	}, nil
}

func GetScanIssues(ctx context.Context, id openapi_types.UUID) ([]*Report, error) {
	client, err := v2.NewAPIClient()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	resp, err := client.GetScanIssuesWithResponse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get scan issues: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unable to get scan issues: status code %d", resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, errors.New("unable to get scan issues: no data received")
	}

	reports := make([]*Report, 0, len(*resp.JSON200))
	for _, issue := range *resp.JSON200 {
		reports = append(reports, &Report{
			Id:       issue.Id,
			Issues:   []IssuesLite{
				{
					Id:       issue.Id,
					Ignored:  issue.Ignored,
				},
			},
			Severity: string(issue.Severity),
			Ignored:  issue.Ignored,
			Test: Test{
				Category:        string(issue.Test.Category),
				SecurityTestUid: issue.Test.SecurityTestUid,
				Meta: Meta{
					TitleOnFail: issue.Test.Meta.TitleOnFail,
					Type:        issue.Test.Meta.Type,
				},
			},
			Meta: Meta{
				TitleOnFail: issue.Test.Meta.TitleOnFail,
				Type:        issue.Test.Meta.Type,
			},
		})
	}

	return reports, nil
}

func GetScanIssue(ctx context.Context, scanId openapi_types.UUID, issueId openapi_types.UUID) ([]*Issue, error) {
	client, err := v2.NewAPIClient()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	resp, err := client.GetScanIssueWithResponse(ctx, scanId, issueId)
	if err != nil {
		return nil, fmt.Errorf("unable to get scan issue: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unable to get scan issue: status code %d", resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, errors.New("unable to get scan issue: no data received")
	}

	issues := make([]*Issue, 0, len(*resp.JSON200))

	for _, issue := range *resp.JSON200 {
		risks := make([]Risks, 0, len(issue.Risks))
		for _, risk := range issue.Risks {
			risks = append(risks, Risks{
				Id: risk.Id,
				Kind: string(risk.Kind),
			})
		}
		issues = append(issues, &Issue{
			Id: issue.Id,
			Ignored: issue.Ignored,
			FirstSeenScanId: issue.FirstSeenScanId,
			LastSeenScanId: issue.LastSeenScanId,
			Severity: string(issue.Severity),
			Risks: risks,
		})
	}

	return issues, nil
}

func GetScanExchangeArchive(ctx context.Context, id openapi_types.UUID) (*ScanExchangeArchive, error) {
	client, err := v2.NewAPIClient()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	resp, err := client.GetScanExchangesArchiveWithResponse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get scan exchange archive: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unable to get scan exchange archive: status code %d", resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, errors.New("unable to get scan exchange archive: no data received")
	}

	return &ScanExchangeArchive{
		Archive: resp.JSON200.Archive,
	}, nil
}

func GetScanEvents(ctx context.Context, id openapi_types.UUID, count *int, after *string) ([]*ScanEvent, *string, error) {
	client, err := v2.NewAPIClient()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	resp, err := client.GetScanEventsWithResponse(ctx, id, &v2.GetScanEventsParams{
		Count: count,
		After: after,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get scan events: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, nil, fmt.Errorf("unable to get scan events: status code %d", resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, nil, errors.New("unable to get scan events: no data received")
	}

	events := []*ScanEvent{}
	for _, event := range resp.JSON200.Data {
		events = append(events, &ScanEvent{
			Id: event.Id,
			CreatedAt: event.CreatedAt,
			Description: event.Description,
			Level: string(event.Level),
			Title: event.Title,
		})
	}

	return events, resp.JSON200.NextCursor, nil
}
