package locations

import (
	"context"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/google/uuid"
)

func Get(ctx context.Context, client *api.ClientWithResponses, id uuid.UUID) (any, func(), error) {
	log.Info("Getting location %s", id.String())
	location, err := client.GetLocationWithResponse(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if location.JSON200 != nil {
		return location.JSON200, func() {
			fmt.Println("Location:", *location.JSON200.Name)
			fmt.Println("Type:", *location.JSON200.Type)
			fmt.Println("Id:", *location.JSON200.Id)
		}, nil
	} else if location.JSON404 != nil {
		return nil, nil, fmt.Errorf("location %s not found: %s", id.String(), location.JSON404.Message)
	} else {
		return nil, nil, fmt.Errorf("unable to get location: Unknown error")
	}
}
