package escape

import (
	"context"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// StartContinuousScan starts a CONTINUOUS_PENTEST on a profile that already has
// continuous pentest enabled. The returned id is the scan id.
func StartContinuousScan(
	ctx context.Context,
	profileID string,
	commitHash string,
	continuousCommitHash string,
	repositoryURL string,
	initiator v3.ENUMPROPERTIESINITIATOR,
) (*v3.StartScan200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := v3.NewStartContinuousScanRequest(profileID)
	if commitHash != "" {
		req.SetCommitHash(commitHash)
	}

	if continuousCommitHash != "" || repositoryURL != "" {
		tip := v3.NewStartContinuousScanRequestContinuous()
		if continuousCommitHash != "" {
			tip.SetCommitHash(continuousCommitHash)
		}

		if repositoryURL != "" {
			tip.SetRepositoryUrl(repositoryURL)
		}

		req.SetContinuous(*tip)
	}

	if initiator != "" {
		req.SetInitiator(initiator)
	}

	data, _, err := client.ContinuousScansAPI.StartContinuousScan(ctx).StartContinuousScanRequest(*req).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to start continuous scan: %w", humanizeAPIError(err))
	}

	return data, nil
}
