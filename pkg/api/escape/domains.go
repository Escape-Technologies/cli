package escape

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
)

type Domain struct {
	CreatedAt     time.Time          `json:"createdAt"`
	Id            openapi_types.UUID `json:"id"`
	Fqdn          string             `json:"fqdn"`
	ServicesCount int                `json:"servicesCount"`
}

func (d *Domain) String() string {
	return fmt.Sprintf("%s\t%s\t%s\t%d", d.Id, d.CreatedAt, d.Fqdn, d.ServicesCount)
}

func GetDomains(ctx context.Context) ([]Domain, error) {
	client, err := v2.NewAPIClient()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	resp, err := client.GetDomainsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get domains: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unable to get domains: status code %d", resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, errors.New("unable to get domains: no data received")
	}
	domains := []Domain{}
	for _, domain := range *resp.JSON200 {
		domains = append(domains, Domain{
			CreatedAt:     domain.CreatedAt,
			Id:            domain.Id,
			Fqdn:          domain.Fqdn,
			ServicesCount: domain.ServicesCount,
		})
	}
	return domains, nil
}
