package escape

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type Subdomain struct {
	Id            openapi_types.UUID `json:"id"`
	Fqdn          string             `json:"fqdn"`
	ServicesCount int                `json:"servicesCount"`
}

func (s *Subdomain) String() string {
	return fmt.Sprintf("%s\t%s\t%d", s.Id, s.Fqdn, s.ServicesCount)
}

func GetSubdomains(ctx context.Context, count *int, after *string) ([]Subdomain, *string, error) {
	client, err := v2.NewAPIClient()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	resp, err := client.GetSubdomainsWithResponse(ctx, &v2.GetSubdomainsParams{
		Count: count,
		After: after,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get subdomains: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, nil, fmt.Errorf("unable to get subdomains: status code %d", resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, nil, errors.New("unable to get subdomains: no data received")
	}
	subdomains := []Subdomain{}
	for _, subdomain := range (*resp.JSON200).Data {
		subdomains = append(subdomains, Subdomain{
			Id:            subdomain.Id,
			Fqdn:          subdomain.Fqdn,
			ServicesCount: subdomain.ServicesCount,
		})
	}
	return subdomains, resp.JSON200.NextCursor, nil
}
