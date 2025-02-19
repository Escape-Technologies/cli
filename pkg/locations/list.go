package locations

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/Escape-Technologies/cli/pkg/log"
)

func List(ctx context.Context, client *api.ClientWithResponses) (any, func(), error) {
	log.Info("Listing locations")
	locations, err := client.ListLocationsWithResponse(ctx)
	if err != nil {
		return nil, nil, err
	}
	if locations.JSON200 == nil {
		return nil, nil, errors.New("no locations found")
	}
	return locations.JSON200, func() {
		for _, location := range *locations.JSON200 {
			typeString := ""
			if location.Type != nil {
				typeString = string(*location.Type)
			}
			nameString := ""
			if location.Name != nil {
				nameString = string(*location.Name)
			}
			fmt.Printf("%s%s%s\n", typeString, strings.Repeat(" ", 10-len(typeString)), nameString)
		}
	}, nil
}
