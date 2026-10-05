package cmd

import "testing"

func TestAssetBulkCommandsRequireSelection(t *testing.T) {
	prevIDs := bulkAssetIDs
	prevTypes := bulkAssetTypes
	prevStatuses := bulkAssetStatuses
	prevTagIDs := bulkAssetTagIDs
	prevProjIDs := bulkAssetProjIDs
	prevStatus := bulkAssetStatus
	t.Cleanup(func() {
		bulkAssetIDs = prevIDs
		bulkAssetTypes = prevTypes
		bulkAssetStatuses = prevStatuses
		bulkAssetTagIDs = prevTagIDs
		bulkAssetProjIDs = prevProjIDs
		bulkAssetStatus = prevStatus
	})

	bulkAssetIDs = nil
	bulkAssetTypes = nil
	bulkAssetStatuses = nil
	bulkAssetTagIDs = nil
	bulkAssetProjIDs = nil
	bulkAssetStatus = ""

	const want = "at least one of --asset-id, --type, or --asset-status is required"

	if err := assetBulkDeleteCmd.RunE(assetBulkDeleteCmd, nil); err == nil || err.Error() != want {
		t.Fatalf("bulk-delete without a filter: %v", err)
	}

	// Assignment flags are not a selection. An empty where still matches every asset.
	bulkAssetStatus = "ARCHIVED"
	bulkAssetTagIDs = []string{"tag-1"}
	bulkAssetProjIDs = []string{"proj-1"}
	if err := assetBulkUpdateCmd.RunE(assetBulkUpdateCmd, nil); err == nil || err.Error() != want {
		t.Fatalf("bulk-update with only mutation flags: %v", err)
	}
}
