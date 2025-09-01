package escape

import (
	"context"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListCustomRules lists all custom rules
func ListCustomRules(ctx context.Context) ([]v3.CustomRuleSummarized, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.CustomRulesAPI.ListCustomRules(ctx)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// GetCustomRule gets a custom rule
func GetCustomRule(ctx context.Context, id string) (*v3.CustomRule, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.CustomRulesAPI.GetCustomRule(ctx, id)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}
