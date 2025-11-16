package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/dataregistry/ipfs"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
)

// StoreDataItem stores a new data item (legacy - uses pre-computed hash and CID)
func (k *Keeper) StoreDataItem(
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
	userItems := k.ListUserDataItems(ownerAddress, types.DataItemTypeUnspecified, types.DataItemStatusUnspecified)
	if uint64(len(userItems)) >= params.MaxDataItemsPerUser {
		return "", types.ErrMaxDataItemsExceeded
	}

	// Generate unique data ID
	dataID := k.GenerateDataID(ownerAddress, dataType)

	// Create data item
	item := types.DataItem{
		DataID:            dataID,
		OwnerAddress:      ownerAddress,
		DataType:          dataType,
		Status:            types.DataItemStatusPendingVerification,
		VerificationLevel: types.VerificationLevelSelfAttested,
		ContentHash:       contentHash,
		StorageLocation:   storageLocation,
		IsEncrypted:       isEncrypted,
		Title:             title,
		Description:       description,
		Metadata:          metadata,
		Tags:              tags,
		CreatedAt:         time.Unix(k.currentTime, 0),
		GeoLocation:       geoLocation,
		Verifications:     []types.Verification{},
		AccessPolicy:      accessPolicy,
		Version:           1,
	}

	// Store the item
	if err := k.SetDataItem(item); err != nil {
		return "", err
	}

	return dataID, nil
}

// StoreDataItemWithContent stores a new data item with content uploaded to IPFS
func (k *Keeper) StoreDataItemWithContent(
	ctx context.Context,
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
	userItems := k.ListUserDataItems(ownerAddress, types.DataItemTypeUnspecified, types.DataItemStatusUnspecified)
	if uint64(len(userItems)) >= params.MaxDataItemsPerUser {
		return "", types.ErrMaxDataItemsExceeded
	}

	// Validate content size
	if err := ipfs.ValidateDataSize(int64(len(content)), int64(params.MaxStorageBytes)); err != nil {
		return "", fmt.Errorf("content size validation failed: %w", err)
	}

	// Calculate content hash
	contentHash := k.ipfsClient.CalculateHash(content)

	// Upload to IPFS
	cid, err := k.ipfsClient.Upload(ctx, content)
	if err != nil {
		return "", fmt.Errorf("failed to upload content to IPFS: %w", err)
	}

	// Generate unique data ID
	dataID := k.GenerateDataID(ownerAddress, dataType)

	// Create data item
	item := types.DataItem{
		DataID:            dataID,
		OwnerAddress:      ownerAddress,
		DataType:          dataType,
		Status:            types.DataItemStatusPendingVerification,
		VerificationLevel: types.VerificationLevelSelfAttested,
		ContentHash:       contentHash,
		StorageLocation:   cid, // Store IPFS CID
		IsEncrypted:       isEncrypted,
		Title:             title,
		Description:       description,
		Metadata:          metadata,
		Tags:              tags,
		CreatedAt:         time.Unix(k.currentTime, 0),
		GeoLocation:       geoLocation,
		Verifications:     []types.Verification{},
		AccessPolicy:      accessPolicy,
		Version:           1,
	}

	// Store the item
	if err := k.SetDataItem(item); err != nil {
		// Unpin from IPFS if storage fails
		_ = k.ipfsClient.Unpin(ctx, cid)
		return "", err
	}

	return dataID, nil
}

// RetrieveDataItemContent downloads content from IPFS for a data item
func (k *Keeper) RetrieveDataItemContent(ctx context.Context, dataID string, requester string) ([]byte, error) {
	// Get data item
	item, ok := k.GetDataItem(dataID)
	if !ok {
		return nil, types.ErrDataItemNotFound
	}

	// Check access
	if !k.CheckAccess(dataID, requester) {
		return nil, types.ErrUnauthorized
	}

	// Validate CID format
	if !ipfs.IsValidCID(item.StorageLocation) {
		return nil, fmt.Errorf("invalid storage location (not a valid IPFS CID): %s", item.StorageLocation)
	}

	// Download from IPFS
	content, err := k.ipfsClient.Download(ctx, item.StorageLocation)
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
func (k *Keeper) UpdateDataItem(
	dataID string,
	ownerAddress string,
	title string,
	description string,
	metadata map[string]string,
	accessPolicy *types.AccessPolicy,
	tags []string,
) error {
	// Get existing item
	item, ok := k.GetDataItem(dataID)
	if !ok {
		return types.ErrDataItemNotFound
	}

	// Check ownership
	if item.OwnerAddress != ownerAddress {
		return types.ErrUnauthorized
	}

	// Check status
	if item.Status == types.DataItemStatusRevoked {
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
	return k.SetDataItem(item)
}

// VerifyDataItem adds a verification to a data item
func (k *Keeper) VerifyDataItem(
	dataID string,
	verifierAddress string,
	level types.VerificationLevel,
	confidenceScore uint64,
	notes string,
	verificationMethod string,
	proof []byte,
) error {
	// Get existing item
	item, ok := k.GetDataItem(dataID)
	if !ok {
		return types.ErrDataItemNotFound
	}

	// Check status
	if item.Status == types.DataItemStatusRevoked {
		return types.ErrDataItemRevoked
	}

	// Create verification record
	verification := types.Verification{
		VerifierAddress:    verifierAddress,
		Level:              level,
		VerifiedAt:         time.Unix(k.currentTime, 0),
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
	if level >= types.VerificationLevelPeerVerified {
		item.Status = types.DataItemStatusVerified
		item.VerifiedAt = time.Unix(k.currentTime, 0)
		item.VerifiedBy = verifierAddress
	}

	// Store updated item
	return k.SetDataItem(item)
}

// RevokeDataItem revokes a data item
func (k *Keeper) RevokeDataItem(dataID string, authority string, reason string) error {
	// Get existing item
	item, ok := k.GetDataItem(dataID)
	if !ok {
		return types.ErrDataItemNotFound
	}

	// Check if already revoked
	if item.Status == types.DataItemStatusRevoked {
		return types.ErrDataItemRevoked
	}

	// Update status
	item.Status = types.DataItemStatusRevoked

	// Store updated item
	return k.SetDataItem(item)
}

// GetDataItemVerifications returns all verifications for a data item
func (k *Keeper) GetDataItemVerifications(dataID string) ([]types.Verification, error) {
	item, ok := k.GetDataItem(dataID)
	if !ok {
		return nil, types.ErrDataItemNotFound
	}

	return item.Verifications, nil
}
