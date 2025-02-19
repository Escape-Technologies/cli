package locations

import (
	"context"
	"fmt"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/Escape-Technologies/cli/pkg/log"
)

func startLocation(ctx context.Context, key string) error {
	log.Info("Location started with %s", key)
	time.Sleep(10 * time.Second)
	return nil
}

func genSSHKeys() (string, string) {
	return "aaa", "bbb"
}

func Start(ctx context.Context, client *api.ClientWithResponses, name string) error {
	privateSSHKey, publicSSHKey := genSSHKeys()

	log.Info("Upserting location %s", name)
	y := true
	location, err := client.UpsertLocationWithResponse(ctx, api.UpsertLocationJSONRequestBody{
		Name:              name,
		PrivateLocationV2: &y,
		Key:               &publicSSHKey,
	})
	if err != nil {
		return err
	}

	if location.JSON200 != nil {
		return startLocation(ctx, privateSSHKey)
	} else if location.JSON400 != nil {
		return fmt.Errorf("unable to start location: %s", location.JSON400.Message)
	} else if location.JSON500 != nil {
		return fmt.Errorf("unable to start location: %s", location.JSON500.Message)
	} else {
		return fmt.Errorf("unable to start location: Unknown error")
	}
}
