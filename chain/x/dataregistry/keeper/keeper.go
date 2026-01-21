// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"encoding/binary"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dataregistry/ipfs"
	"github.com/aequitas/aura/chain/x/dataregistry/params"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
)

// BankKeeper defines the expected interface for the bank keeper
type BankKeeper interface {
	MintCoins(ctx context.Context, moduleName string, amt interface{}) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr interface{}, amt interface{}) error
}

// Keeper manages the state of the dataregistry module using persistent KV store.
// All state is stored deterministically in the KV store to ensure consensus safety.
type Keeper struct {
	storeService store.KVStoreService
	cdc          codec.BinaryCodec
	paramsStore  *params.Store
	ipfsClient   ipfs.IPFSClient // IPFS client for content storage
	bankKeeper   BankKeeper
	authority    string
	logger       log.Logger
}

// NewKeeper creates a new Keeper instance with persistent KV store.
// All state is persisted to the KV store - no in-memory maps are used.
func NewKeeper(
	storeService store.KVStoreService,
	cdc codec.BinaryCodec,
	paramsStore *params.Store,
	authority string,
	logger log.Logger,
) *Keeper {
	if paramsStore == nil {
		paramsStore = params.NewStore(types.DefaultParams())
	}

	return &Keeper{
		storeService: storeService,
		cdc:          cdc,
		paramsStore:  paramsStore,
		ipfsClient:   ipfs.NewMockClient(), // Default to mock client
		authority:    authority,
		logger:       logger,
	}
}

// SetBankKeeper sets the bank keeper
func (k *Keeper) SetBankKeeper(bankKeeper BankKeeper) {
	k.bankKeeper = bankKeeper
}

// SetAuthority sets the governance authority
func (k *Keeper) SetAuthority(authority string) {
	k.authority = authority
}

// GetAuthority returns the governance module account address
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetIPFSClient sets the IPFS client (useful for testing)
func (k *Keeper) SetIPFSClient(client ipfs.IPFSClient) {
	k.ipfsClient = client
}

// GetIPFSClient returns the IPFS client
func (k *Keeper) GetIPFSClient() ipfs.IPFSClient {
	return k.ipfsClient
}

// GetParams returns the current module parameters
func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	if k.paramsStore != nil {
		return k.paramsStore.GetParams(), nil
	}
	return types.DefaultParams(), nil
}

// SetParams sets new module parameters
func (k *Keeper) SetParams(params types.Params) error {
	if k.paramsStore == nil {
		return types.ErrUnauthorized
	}
	return k.paramsStore.SetParams(params)
}

// ============================
// DATA ITEM COUNTER MANAGEMENT
// ============================

// GetNextDataID retrieves and increments the next data ID from KV store
func (k *Keeper) GetNextDataID(ctx sdk.Context) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.DataItemCounterKey)
	if err != nil || bz == nil {
		return 1
	}

	if len(bz) != 8 {
		return 1
	}

	return binary.BigEndian.Uint64(bz)
}

// SetNextDataID stores the next data ID to KV store
func (k *Keeper) SetNextDataID(ctx sdk.Context, id uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)

	if err := store.Set(types.DataItemCounterKey, bz); err != nil {
		return fmt.Errorf("failed to set next data ID: %w", err)
	}

	return nil
}

// GenerateDataID generates a unique data ID using counter
func (k *Keeper) GenerateDataID(ctx sdk.Context, ownerAddress string, dataType types.DataItemType) string {
	nextID := k.GetNextDataID(ctx)
	if err := k.SetNextDataID(ctx, nextID+1); err != nil {
		panic(fmt.Sprintf("failed to update data ID counter: %v", err))
	}

	return fmt.Sprintf("data:%s:%d:%d", ownerAddress, dataType, nextID)
}

// ============================
// DATA ITEM MANAGEMENT
// ============================

// GetDataItem retrieves a data item by ID from KV store
func (k *Keeper) GetDataItem(ctx sdk.Context, dataID string) (types.DataItem, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.DataItemKey(dataID))
	if err != nil || bz == nil {
		return types.DataItem{}, false
	}

	var item types.DataItem
	if err := k.cdc.Unmarshal(bz, &item); err != nil {
		return types.DataItem{}, false
	}

	return item, true
}

// SetDataItem stores a data item to KV store
func (k *Keeper) SetDataItem(ctx sdk.Context, item types.DataItem) error {
	if item.DataId == "" {
		return types.ErrInvalidDataID
	}
	if item.OwnerAddress == "" {
		return types.ErrInvalidOwner
	}

	store := k.storeService.OpenKVStore(ctx)

	// Marshal and store the data item
	bz, err := k.cdc.Marshal(&item)
	if err != nil {
		return fmt.Errorf("failed to marshal data item: %w", err)
	}

	if err := store.Set(types.DataItemKey(item.DataId), bz); err != nil {
		return fmt.Errorf("failed to store data item: %w", err)
	}

	// Index by user (store data ID in user index)
	if err := k.addToUserIndex(ctx, item.OwnerAddress, item.DataId); err != nil {
		return fmt.Errorf("failed to marshal for DataId: %w", err)
	}

	// Index by type
	if err := k.addToTypeIndex(ctx, item.DataType, item.DataId); err != nil {
		return fmt.Errorf("error in SetDataItem for DataId: %w", err)
	}

	return nil
}

// DeleteDataItem removes a data item from KV store
func (k *Keeper) DeleteDataItem(ctx sdk.Context, dataID string) error {
	item, ok := k.GetDataItem(ctx, dataID)
	if !ok {
		return types.ErrDataItemNotFound
	}

	// Unpin from IPFS if storage location is a valid CID
	if ipfs.IsValidCID(item.StorageLocation) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		if err := k.ipfsClient.Unpin(sdkCtx, item.StorageLocation); err != nil {
			// Log error but continue with deletion
			k.logger.Error("failed to unpin IPFS content", "error", err)
		}
	}

	store := k.storeService.OpenKVStore(ctx)

	// Remove from main store
	if err := store.Delete(types.DataItemKey(dataID)); err != nil {
		return fmt.Errorf("failed to delete data item: %w", err)
	}

	// Remove from user index
	if err := k.removeFromUserIndex(ctx, item.OwnerAddress, dataID); err != nil {
		return fmt.Errorf("error in DeleteDataItem: %w", err)
	}

	// Remove from type index
	if err := k.removeFromTypeIndex(ctx, item.DataType, dataID); err != nil {
		return fmt.Errorf("error in DeleteDataItem: %w", err)
	}

	return nil
}

// ============================
// USER INDEX MANAGEMENT
// ============================

// getUserDataIDs gets user data IDs by iterating through all items
// Note: For production, consider adding a separate index message type in proto
func (k *Keeper) getUserDataIDs(ctx sdk.Context, ownerAddress string) []string {
	store := k.storeService.OpenKVStore(ctx)
	dataIDs := make([]string, 0, 64)

	prefix := types.DataItemKeyPrefix
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return dataIDs
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var item types.DataItem
		if err := k.cdc.Unmarshal(iterator.Value(), &item); err != nil {
			continue
		}
		if item.OwnerAddress == ownerAddress {
			dataIDs = append(dataIDs, item.DataId)
		}
	}

	return dataIDs
}

// setUserDataIDs is no longer needed with iteration-based approach
func (k *Keeper) setUserDataIDs(ctx sdk.Context, ownerAddress string, dataIDs []string) error {
	// This is now handled automatically by storing DataItems
	return nil
}

// addToUserIndex adds a data ID to user's index
func (k *Keeper) addToUserIndex(ctx sdk.Context, ownerAddress, dataID string) error {
	dataIDs := k.getUserDataIDs(ctx, ownerAddress)

	// Check if already exists
	for _, id := range dataIDs {
		if id == dataID {
			return nil // Already indexed
		}
	}

	dataIDs = append(dataIDs, dataID)
	return k.setUserDataIDs(ctx, ownerAddress, dataIDs)
}

// removeFromUserIndex removes a data ID from user's index
func (k *Keeper) removeFromUserIndex(ctx sdk.Context, ownerAddress, dataID string) error {
	dataIDs := k.getUserDataIDs(ctx, ownerAddress)

	newDataIDs := []string{}
	for _, id := range dataIDs {
		if id != dataID {
			newDataIDs = append(newDataIDs, id)
		}
	}

	return k.setUserDataIDs(ctx, ownerAddress, newDataIDs)
}

// ListUserDataItems returns all data items for a user from KV store
func (k *Keeper) ListUserDataItems(ctx sdk.Context, ownerAddress string, typeFilter types.DataItemType, statusFilter types.DataItemStatus) []types.DataItem {
	dataIDs := k.getUserDataIDs(ctx, ownerAddress)

	items := []types.DataItem{}
	for _, dataID := range dataIDs {
		if item, ok := k.GetDataItem(ctx, dataID); ok {
			// Apply filters
			if typeFilter != types.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED && item.DataType != typeFilter {
				continue
			}
			if statusFilter != types.DataItemStatus_DATA_ITEM_STATUS_UNSPECIFIED && item.Status != statusFilter {
				continue
			}
			items = append(items, item)
		}
	}

	return items
}

// ============================
// TYPE INDEX MANAGEMENT
// ============================

// getTypeDataIDs gets data IDs by type by iterating through all items
// Note: For production, consider adding a separate index message type in proto
func (k *Keeper) getTypeDataIDs(ctx sdk.Context, dataType types.DataItemType) []string {
	store := k.storeService.OpenKVStore(ctx)
	dataIDs := make([]string, 0, 64)

	prefix := types.DataItemKeyPrefix
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return dataIDs
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var item types.DataItem
		if err := k.cdc.Unmarshal(iterator.Value(), &item); err != nil {
			continue
		}
		if item.DataType == dataType {
			dataIDs = append(dataIDs, item.DataId)
		}
	}

	return dataIDs
}

// setTypeDataIDs is no longer needed with iteration-based approach
func (k *Keeper) setTypeDataIDs(ctx sdk.Context, dataType types.DataItemType, dataIDs []string) error {
	// This is now handled automatically by storing DataItems
	return nil
}

// addToTypeIndex adds a data ID to type index
func (k *Keeper) addToTypeIndex(ctx sdk.Context, dataType types.DataItemType, dataID string) error {
	dataIDs := k.getTypeDataIDs(ctx, dataType)

	// Check if already exists
	for _, id := range dataIDs {
		if id == dataID {
			return nil // Already indexed
		}
	}

	dataIDs = append(dataIDs, dataID)
	return k.setTypeDataIDs(ctx, dataType, dataIDs)
}

// removeFromTypeIndex removes a data ID from type index
func (k *Keeper) removeFromTypeIndex(ctx sdk.Context, dataType types.DataItemType, dataID string) error {
	dataIDs := k.getTypeDataIDs(ctx, dataType)

	newDataIDs := []string{}
	for _, id := range dataIDs {
		if id != dataID {
			newDataIDs = append(newDataIDs, id)
		}
	}

	return k.setTypeDataIDs(ctx, dataType, newDataIDs)
}

// ============================
// ACCESS CONTROL
// ============================

// CheckAccess checks if requester has access to a data item
func (k *Keeper) CheckAccess(ctx sdk.Context, dataID string, requester string) bool {
	item, ok := k.GetDataItem(ctx, dataID)
	if !ok {
		return false
	}

	// Owner always has access
	if item.OwnerAddress == requester {
		return true
	}

	// Check access policy
	if item.AccessPolicy == nil {
		return false
	}

	switch item.AccessPolicy.Mode {
	case types.AccessMode_ACCESS_MODE_PRIVATE:
		return false
	case types.AccessMode_ACCESS_MODE_PUBLIC:
		return true
	case types.AccessMode_ACCESS_MODE_WHITELIST:
		for _, addr := range item.AccessPolicy.AllowedAddresses {
			if addr == requester {
				return true
			}
		}
		return false
	case types.AccessMode_ACCESS_MODE_VERIFIED_USERS:
		// In production, check if requester has AURA VC
		return true
	default:
		return false
	}
}

// ============================
// SEARCH FUNCTIONALITY
// ============================

// SearchDataItems searches for data items
func (k *Keeper) SearchDataItems(ctx sdk.Context, query string, tags []string, typeFilter types.DataItemType, geoLocation *types.GeoLocation, radiusKM float64, requester string) []types.DataItem {
	store := k.storeService.OpenKVStore(ctx)

	results := []types.DataItem{}

	// Iterate over all data items
	prefix := types.DataItemKeyPrefix
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return results
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var item types.DataItem
		if err := k.cdc.Unmarshal(iterator.Value(), &item); err != nil {
			continue
		}

		// Check access
		if !k.CheckAccess(ctx, item.DataId, requester) {
			continue
		}

		// Type filter
		if typeFilter != types.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED && item.DataType != typeFilter {
			continue
		}

		// Tag filter
		if len(tags) > 0 {
			hasTag := false
			for _, searchTag := range tags {
				for _, itemTag := range item.Tags {
					if itemTag == searchTag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		// Geo filter (simple distance check)
		if geoLocation != nil && item.GeoLocation != nil && radiusKM > 0 {
			distance := calculateDistance(
				geoLocation.Latitude, geoLocation.Longitude,
				item.GeoLocation.Latitude, item.GeoLocation.Longitude,
			)
			if distance > radiusKM {
				continue
			}
		}

		results = append(results, item)
	}

	return results
}

// calculateDistance calculates distance between two points (simplified)
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Simplified distance calculation (in production, use Haversine formula)
	dlat := lat2 - lat1
	dlon := lon2 - lon1
	return dlat*dlat + dlon*dlon // Approximate distance squared
}

// ============================
// STATISTICS
// ============================

// GetStats returns registry statistics from KV store
func (k *Keeper) GetStats(ctx sdk.Context) types.RegistryStats {
	store := k.storeService.OpenKVStore(ctx)

	stats := types.RegistryStats{
		TotalDataItems: 0,
		ItemsByType:    make(map[string]uint64),
	}

	// Iterate over all data items
	prefix := types.DataItemKeyPrefix
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return stats
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var item types.DataItem
		if err := k.cdc.Unmarshal(iterator.Value(), &item); err != nil {
			continue
		}

		stats.TotalDataItems++

		if item.Status == types.DataItemStatus_DATA_ITEM_STATUS_VERIFIED {
			stats.TotalVerifiedItems++
		}

		stats.TotalVerifications += uint64(len(item.Verifications))
		stats.TotalStorageBytes += calculateItemStorageBytes(item)

		typeName := fmt.Sprintf("%d", item.DataType)
		stats.ItemsByType[typeName]++
	}

	return stats
}

func calculateItemStorageBytes(item types.DataItem) uint64 {
	size := uint64(len(item.ContentHash))
	size += uint64(len(item.StorageLocation))
	size += uint64(len(item.Title))
	size += uint64(len(item.Description))

	for key, value := range item.Metadata {
		size += uint64(len(key) + len(value))
	}

	for _, tag := range item.Tags {
		size += uint64(len(tag))
	}

	if item.AccessPolicy != nil {
		for _, addr := range item.AccessPolicy.AllowedAddresses {
			size += uint64(len(addr))
		}
		for _, addr := range item.AccessPolicy.DeniedAddresses {
			size += uint64(len(addr))
		}
		if item.AccessPolicy.RequireVerifiedIdentity {
			size++
		}
	}

	if item.GeoLocation != nil {
		size += 24
	}

	return size
}

// ============================
// GENESIS MANAGEMENT
// ============================

// Genesis methods are implemented in genesis.go
