package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/aequitas/aura/chain/x/dataregistry/ipfs"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
)

// StoreDataItem stores a new data item (legacy - uses pre-computed hash and CID)
func (k *Keeper) StoreDataItem(ctx sdk.Context, 
	ownerAddress string,
	dataType types.DataItemType,
	title string,
	description string,
	contentHash []byte,
	storageLocation string,
	isEncrypted bool,
	geoLocation *types.GeoLocation,
	metadata map[string]string,
	accessPolicy *types.AccessPolicy,
	tags []string,
) (string, error) {
	// Validate inputs
	if ownerAddress == "" {
		return "", types.ErrInvalidOwner
	}
	if len(contentHash) == 0 {
		return "", types.ErrInvalidContentHash
	}
	if storageLocation == "" {
		return "", types.ErrInvalidStorageLocation
	}

	// Check user limits
	params := k.GetParams()
	userItems := k.ListUserDataItems(ctx, ownerAddress, types.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED, types.DataItemStatus_DATA_ITEM_STATUS_UNSPECIFIED)
	if uint64(len(userItems)) >= params.MaxDataItemsPerUser {
		return "", types.ErrMaxDataItemsExceeded
	}

	// Generate unique data ID
	dataID := k.GenerateDataID(ctx, ownerAddress, dataType)

	// Create data item
	item := types.DataItem{
		DataId:            dataID,
		OwnerAddress:      ownerAddress,
		DataType:          dataType,
		Status:            types.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
		VerificationLevel: types.VerificationLevel_VERIFICATION_LEVEL_SELF_ATTESTED,
		ContentHash:       contentHash,
		StorageLocation:   storageLocation,
		IsEncrypted:       isEncrypted,
		Title:             title,
		Description:       description,
		Metadata:          metadata,
		Tags:              tags,
		CreatedAt:         timestampFromTime(ctx.BlockTime()),
		GeoLocation:       geoLocation,
		Verifications:     []*types.Verification{},
		AccessPolicy:      accessPolicy,
		Version:           1,
	}

	// Store the item
	if err := k.SetDataItem(ctx, item); err != nil {
		return "", err
	}

	return dataID, nil
}

// timestampFromTime converts time.Time to gogoproto timestamp
func timestampFromTime(t time.Time) *gogotypes.Timestamp {
	if t.IsZero() {
		return nil
	}
	seconds := t.Unix()
	nanos := int32(t.UnixNano() - (seconds * 1000000000))
	return &gogotypes.Timestamp{
		Seconds: seconds,
		Nanos:   nanos,
	}
}

// StoreDataItemWithContent stores a new data item with content uploaded to IPFS
func (k *Keeper) StoreDataItemWithContent(sdkCtx sdk.Context,
	ownerAddress string,
	dataType types.DataItemType,
	title string,
	description string,
	content []byte,
	isEncrypted bool,
	geoLocation *types.GeoLocation,
	metadata map[string]string,
	accessPolicy *types.AccessPolicy,
	tags []string,
) (string, error) {
	// Validate inputs
	if ownerAddress == "" {
		return "", types.ErrInvalidOwner
	}
	if len(content) == 0 {
		return "", fmt.Errorf("content cannot be empty")
	}

	// Check user limits
	params := k.GetParams()
	userItems := k.ListUserDataItems(sdkCtx, ownerAddress, types.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED, types.DataItemStatus_DATA_ITEM_STATUS_UNSPECIFIED)
	if uint64(len(userItems)) >= params.MaxDataItemsPerUser {
		return "", types.ErrMaxDataItemsExceeded
	}

	// Validate content size
	if err := ipfs.ValidateDataSize(int64(len(content)), int64(params.MaxStorageBytes)); err != nil {
		return "", fmt.Errorf("content size validation failed: %w", err)
	}

	// Calculate content hash
	contentHash := k.ipfsClient.CalculateHash(content)

	// Create context for IPFS operations
	goCtx := context.Background()

	// Upload to IPFS
	cid, err := k.ipfsClient.Upload(goCtx, content)
	if err != nil {
		return "", fmt.Errorf("failed to upload content to IPFS: %w", err)
	}

	// Generate unique data ID
	dataID := k.GenerateDataID(sdkCtx, ownerAddress, dataType)

	// Create data item
	item := types.DataItem{
		DataId:            dataID,
		OwnerAddress:      ownerAddress,
		DataType:          dataType,
		Status:            types.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
		VerificationLevel: types.VerificationLevel_VERIFICATION_LEVEL_SELF_ATTESTED,
		ContentHash:       contentHash,
		StorageLocation:   cid, // Store IPFS CID
		IsEncrypted:       isEncrypted,
		Title:             title,
		Description:       description,
		Metadata:          metadata,
		Tags:              tags,
		CreatedAt:         timestampFromTime(sdkCtx.BlockTime()),
		GeoLocation:       geoLocation,
		Verifications:     []*types.Verification{},
		AccessPolicy:      accessPolicy,
		Version:           1,
	}

	// Store the item
	if err := k.SetDataItem(sdkCtx, item); err != nil {
		// Unpin from IPFS if storage fails
		_ = k.ipfsClient.Unpin(goCtx, cid)
		return "", err
	}

	return dataID, nil
}

// RetrieveDataItemContent downloads content from IPFS for a data item
func (k *Keeper) RetrieveDataItemContent(sdkCtx sdk.Context, dataID string, requester string) ([]byte, error) {
	// Get data item
	item, ok := k.GetDataItem(sdkCtx, dataID)
	if !ok {
		return nil, types.ErrDataItemNotFound
	}

	// Check access
	if !k.CheckAccess(sdkCtx, dataID, requester) {
		return nil, types.ErrUnauthorized
	}

	// Validate CID format
	if !ipfs.IsValidCID(item.StorageLocation) {
		return nil, fmt.Errorf("invalid storage location (not a valid IPFS CID): %s", item.StorageLocation)
	}

	// Create context for IPFS operations
	goCtx := context.Background()

	// Download from IPFS
	content, err := k.ipfsClient.Download(goCtx, item.StorageLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to download content from IPFS: %w", err)
	}

	// Verify content hash
	if !k.ipfsClient.VerifyHash(content, item.ContentHash) {
		return nil, fmt.Errorf("content hash mismatch: downloaded content does not match stored hash")
	}

	return content, nil
}

// UpdateDataItem updates an existing data item
func (k *Keeper) UpdateDataItem(ctx sdk.Context,
	dataID string,
	ownerAddress string,
	title string,
	description string,
	metadata map[string]string,
	accessPolicy *types.AccessPolicy,
	tags []string,
) error {
	// Get existing item
	item, ok := k.GetDataItem(ctx, dataID)
	if !ok {
		return types.ErrDataItemNotFound
	}

	// Check ownership
	if item.OwnerAddress != ownerAddress {
		return types.ErrUnauthorized
	}

	// Check status
	if item.Status == types.DataItemStatus_DATA_ITEM_STATUS_REVOKED {
		return types.ErrDataItemRevoked
	}

	// Update fields
	if title != "" {
		item.Title = title
	}
	if description != "" {
		item.Description = description
	}
	if metadata != nil {
		item.Metadata = metadata
	}
	if accessPolicy != nil {
		item.AccessPolicy = accessPolicy
	}
	if tags != nil {
		item.Tags = tags
	}

	// Store updated item
	return k.SetDataItem(ctx, item)
}

// VerifyDataItem adds a verification to a data item
func (k *Keeper) VerifyDataItem(ctx sdk.Context,
	dataID string,
	verifierAddress string,
	level types.VerificationLevel,
	confidenceScore uint64,
	notes string,
	verificationMethod string,
	proof []byte,
) error {
	// Get existing item
	item, ok := k.GetDataItem(ctx, dataID)
	if !ok {
		return types.ErrDataItemNotFound
	}

	// Check status
	if item.Status == types.DataItemStatus_DATA_ITEM_STATUS_REVOKED {
		return types.ErrDataItemRevoked
	}

	// Create verification record
	verification := &types.Verification{
		VerifierAddress:    verifierAddress,
		Level:              level,
		VerifiedAt:         timestampFromTime(ctx.BlockTime()),
		VerificationMethod: verificationMethod,
		ConfidenceScore:    confidenceScore,
		Notes:              notes,
		Proof:              proof,
	}

	// Add verification
	item.Verifications = append(item.Verifications, verification)

	// Update verification level and status
	if level > item.VerificationLevel {
		item.VerificationLevel = level
	}
	if level >= types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED {
		item.Status = types.DataItemStatus_DATA_ITEM_STATUS_VERIFIED
		item.VerifiedAt = timestampFromTime(ctx.BlockTime())
		item.VerifiedBy = verifierAddress
	}

	// Store updated item
	return k.SetDataItem(ctx, item)
}

// RevokeDataItem revokes a data item
func (k *Keeper) RevokeDataItem(ctx sdk.Context, dataID string, authority string, reason string) error {
	// Get existing item
	item, ok := k.GetDataItem(ctx, dataID)
	if !ok {
		return types.ErrDataItemNotFound
	}

	// Check if already revoked
	if item.Status == types.DataItemStatus_DATA_ITEM_STATUS_REVOKED {
		return types.ErrDataItemRevoked
	}

	// Update status
	item.Status = types.DataItemStatus_DATA_ITEM_STATUS_REVOKED

	// Store updated item
	return k.SetDataItem(ctx, item)
}

// GetDataItemVerifications returns all verifications for a data item
func (k *Keeper) GetDataItemVerifications(ctx sdk.Context, dataID string) ([]*types.Verification, error) {
	item, ok := k.GetDataItem(ctx, dataID)
	if !ok {
		return nil, types.ErrDataItemNotFound
	}

	return item.Verifications, nil
}
