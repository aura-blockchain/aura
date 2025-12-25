// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	lru "github.com/hashicorp/golang-lru/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

// transferCacheKey generates a cache key for a transfer lookup
func transferCacheKey(transferID string) string {
	return transferID
}

// getTransferWithCache retrieves a transfer from cache or store
func (k *Keeper) getTransferWithCache(ctx sdk.Context, transferID string) (*types.CrossChainTransfer, bool) {
	if transferID == "" {
		return nil, false
	}

	// Check cache first
	if k.transferCache != nil {
		cacheKey := transferCacheKey(transferID)
		if cached, ok := k.transferCache.Get(cacheKey); ok {
			// Cache hit - return cached value
			return cached, true
		}
	}

	// Cache miss - fetch from store
	store := k.store(ctx)
	bz := store.Get(types.TransferKey(transferID))
	if bz == nil {
		return nil, false
	}

	var transfer types.CrossChainTransfer
	if err := k.cdc.Unmarshal(bz, &transfer); err != nil {
		return nil, false
	}

	// Update cache
	if k.transferCache != nil {
		k.transferCache.Add(transferCacheKey(transferID), &transfer)
	}

	return &transfer, true
}

// setTransferWithCache updates a transfer in store and invalidates cache
func (k *Keeper) setTransferWithCache(ctx sdk.Context, transfer *types.CrossChainTransfer) error {
	if transfer == nil || transfer.TransferId == "" {
		return nil
	}

	// Marshal and store
	bz, err := k.cdc.Marshal(transfer)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal transfer %s: %v", transfer.TransferId, err)
	}
	k.store(ctx).Set(types.TransferKey(transfer.TransferId), bz)

	// Invalidate cache entry to ensure consistency
	if k.transferCache != nil {
		k.transferCache.Remove(transferCacheKey(transfer.TransferId))
	}

	return nil
}

// deleteTransferWithCache deletes a transfer from store and cache
func (k *Keeper) deleteTransferWithCache(ctx sdk.Context, transferID string) {
	if transferID == "" {
		return
	}

	// Delete from store
	k.store(ctx).Delete(types.TransferKey(transferID))

	// Invalidate cache
	if k.transferCache != nil {
		k.transferCache.Remove(transferCacheKey(transferID))
	}
}

// initTransferCache initializes the transfer LRU cache
func (k *Keeper) initTransferCache(size int) error {
	if size <= 0 {
		size = DefaultTransferCacheSize
	}

	cache, err := lru.New[string, *types.CrossChainTransfer](size)
	if err != nil {
		return fmt.Errorf("failed to create transfer cache with size %d: %w", size, err)
	}

	k.transferCache = cache
	return nil
}

// ClearTransferCache clears all entries from the transfer cache
func (k *Keeper) ClearTransferCache() {
	if k.transferCache != nil {
		k.transferCache.Purge()
	}
}

// GetCacheStats returns cache statistics (hits, misses, size)
// This can be used for monitoring and metrics
func (k *Keeper) GetCacheStats() (size int, capacity int) {
	if k.transferCache == nil {
		return 0, 0
	}
	return k.transferCache.Len(), DefaultTransferCacheSize
}
