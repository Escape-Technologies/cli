package locations

import (
	"context"
	"fmt"

	v1 "github.com/Escape-Technologies/cli/pkg/api/v1"
	"github.com/Escape-Technologies/cli/pkg/log"
)

func Upsert(ctx context.Context, client *v1.ClientWithResponses, input LocationSchemaInput) (any, func(), error) {
	log.Info("Upserting location %s", input.Name)
	location, err := client.UpsertLocationWithResponse(ctx, v1.UpsertLocationJSONRequestBody{
		Name:              input.Name,
		PrivateLocationV2: &input.PrivateLocationV2,
	})
	if err != nil {
		return nil, nil, err
	}
	if location.JSON200 != nil {
		return location.JSON200, func() {
			fmt.Println("Location:", *location.JSON200.Name)
			fmt.Println("Type:", *location.JSON200.Type)
			fmt.Println("Id:", *location.JSON200.Id)
		}, nil
	} else if location.JSON400 != nil {
		return nil, nil, fmt.Errorf("unable to create location: %s", location.JSON400.Message)
	} else if location.JSON500 != nil {
		return nil, nil, fmt.Errorf("unable to create location: %s", location.JSON500.Message)
	} else {
		return nil, nil, fmt.Errorf("unable to create location: Unknown error")
	}
}
