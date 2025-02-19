package locations

import (
	"context"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/Escape-Technologies/cli/pkg/privatelocation"
)


func Start(ctx context.Context, client *api.ClientWithResponses, name string) error {
	sshPublicKey, sshPrivateKey := privatelocation.GenSSHKeys(name)
	log.Info("Generated public SSH Key: %s", sshPublicKey)

	log.Info("Upserting location %s with public key %s", name, sshPublicKey)
	y := true
	location, err := client.UpsertLocationWithResponse(ctx, api.UpsertLocationJSONRequestBody{
		Name:              name,
		PrivateLocationV2: &y,
		Key:               &sshPublicKey,
	})
	if err != nil {
		return err
	}

	if location.JSON200 != nil {
		return privatelocation.StartLocation(ctx,location.JSON200.Id.String(), sshPrivateKey)
	} else if location.JSON400 != nil {
		return fmt.Errorf("unable to start location: %s", location.JSON400.Message)
	} else if location.JSON500 != nil {
		return fmt.Errorf("unable to start location: %s", location.JSON500.Message)
	} else {
		return fmt.Errorf("unable to start location: Unknown error")
	}
}
