package locations

import (
	"context"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/google/uuid"
)

func Delete(ctx context.Context, client *api.ClientWithResponses, id uuid.UUID) (any, func(), error) {
	log.Info("Deleting location %s", id.String())
	deleted, err := client.DeleteLocationWithResponse(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if deleted.JSON200 != nil {
		return deleted.JSON200, func() { fmt.Println("Location deleted") }, nil
	} else if deleted.JSON404 != nil {
		return nil, nil, fmt.Errorf("location %s not found: %s", id.String(), deleted.JSON404.Message)
	} else if deleted.JSON400 != nil {
		return nil, nil, fmt.Errorf("unable to delete location: %s", deleted.JSON400.Message)
	} else {
		return nil, nil, fmt.Errorf("unable to delete location: Unknown error")
	}
}
