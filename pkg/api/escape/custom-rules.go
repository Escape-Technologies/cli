package escape

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListCustomRulesFilters holds optional filters for listing custom rules.
// Empty fields are omitted. Both are applied on every drained page.
type ListCustomRulesFilters struct {
	Search  string
	Context string
}

// ListCustomRules lists every custom rule, draining every page the
// X-Next-Cursor header announces. A nil filters lists every rule.
func ListCustomRules(ctx context.Context, filters *ListCustomRulesFilters) ([]v3.CustomRuleSummarized, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	rules, err := drainCursorPages(func(cursor string) ([]v3.CustomRuleSummarized, *http.Response, error) {
		req := client.CustomRulesAPI.ListCustomRules(ctx).Size(maxRolePageSize)
		if cursor != "" {
			req = req.Cursor(cursor)
		}

		if filters != nil && filters.Search != "" {
			req = req.Search(filters.Search)
		}

		if filters != nil && filters.Context != "" {
			req = req.Context(filters.Context)
		}

		data, httpRes, err := req.Execute()
		if err != nil {
			return nil, nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
		}

		return data, httpRes, nil
	})
	if err != nil {
		return nil, err
	}

	return rules, nil
}

// GetCustomRule gets a custom rule
func GetCustomRule(ctx context.Context, id string) (*v3.CreateCustomRule200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.CustomRulesAPI.GetCustomRule(ctx, id)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// CreateCustomRule creates a custom rule from raw JSON
func CreateCustomRule(ctx context.Context, data []byte) (*v3.CreateCustomRule200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.CreateCustomRuleRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	req := client.CustomRulesAPI.CreateCustomRule(ctx)
	res, _, err := req.CreateCustomRuleRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return res, nil
}

// UpdateCustomRule updates a custom rule from raw JSON
func UpdateCustomRule(ctx context.Context, id string, data []byte) (*v3.CreateCustomRule200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.UpdateCustomRuleRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	req := client.CustomRulesAPI.UpdateCustomRule(ctx, id)
	res, httpRes, err := req.UpdateCustomRuleRequest(payload).Execute()
	if err != nil {
		if httpRes != nil && httpRes.Body != nil {
			body, _ := io.ReadAll(httpRes.Body)
			return nil, fmt.Errorf("api error: %s", string(body))
		}

		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return res, nil
}

// DeleteCustomRule deletes a custom rule
func DeleteCustomRule(ctx context.Context, id string) (*v3.DeleteCustomRule200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.CustomRulesAPI.DeleteCustomRule(ctx, id)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}
