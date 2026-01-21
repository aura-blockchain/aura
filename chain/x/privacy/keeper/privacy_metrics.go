// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/privacy/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PrivacyMetrics tracks privacy-related metrics
// DETERMINISM: All percentage/decimal fields use basis points (0-10000) instead of float64
// to ensure cross-platform consistency. Float64 operations can produce different results
// on different CPU architectures, causing consensus failures.
type PrivacyMetrics struct {
	TotalShieldedTxs       int64
	TotalRingSignatures    int64
	TotalMixingRounds      int64
	AveragePrivacyScoreBps uint64 // Basis points (0-10000)
	AverageRingSizeBps     uint64 // Ring size scaled by 100 (e.g., 750 = 7.5 ring members)
	AnonymitySetSize       int64
	LastUpdated            time.Time
}

// TrackPrivacyMetrics tracks privacy metrics
func (k Keeper) TrackPrivacyMetrics(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	metrics := &PrivacyMetrics{
		LastUpdated: sdkCtx.BlockTime(),
	}

	// Count metrics
	metrics.TotalShieldedTxs = k.countShieldedTransactions(ctx)
	metrics.TotalRingSignatures = k.countRingSignatures(ctx)
	metrics.TotalMixingRounds = k.countMixingRounds(ctx)
	metrics.AverageRingSizeBps = k.calculateAverageRingSizeBps(ctx)
	metrics.AnonymitySetSize = k.getAnonymitySetSize(ctx)

	// Store metrics
	k.storePrivacyMetrics(ctx, metrics)

	return nil
}

func (k Keeper) countShieldedTransactions(ctx context.Context) int64 {
	store := k.getStore(ctx)
	count := int64(0)

	iterator := storetypes.KVStorePrefixIterator(store, types.ShieldedTxPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		count++
	}

	return count
}

func (k Keeper) countRingSignatures(ctx context.Context) int64 {
	store := k.getStore(ctx)
	count := int64(0)

	iterator := storetypes.KVStorePrefixIterator(store, types.KeyImagePrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		count++
	}

	return count
}

func (k Keeper) countMixingRounds(ctx context.Context) int64 {
	pools := k.GetAllMixingPools(ctx)
	count := int64(0)

	for _, pool := range pools {
		if pool.Status == "completed" {
			count++
		}
	}

	return count
}

// calculateAverageRingSizeBps calculates average ring size scaled by 100 for precision.
// DETERMINISM: Uses integer arithmetic instead of float64 division.
// Returns ring size * 100, e.g., if min=5, max=10, returns (5+10)*100/2 = 750 (representing 7.5)
func (k Keeper) calculateAverageRingSizeBps(ctx context.Context) uint64 {
	params := k.GetParams(ctx)
	// Scale by 100 before division to maintain precision
	return uint64((params.MinRingSize + params.MaxRingSize) * 100 / 2)
}

func (k Keeper) getAnonymitySetSize(ctx context.Context) int64 {
	store := k.getStore(ctx)
	count := int64(0)

	iterator := storetypes.KVStorePrefixIterator(store, types.RingMemberPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		count++
	}

	return count
}

func (k Keeper) storePrivacyMetrics(ctx context.Context, metrics *PrivacyMetrics) {
	store := k.getStore(ctx)
	key := []byte("privacy_metrics")

	// DETERMINISM: Store all values as integers - no float formatting which can vary by platform
	data := []byte(fmt.Sprintf("%d,%d,%d,%d,%d",
		metrics.TotalShieldedTxs,
		metrics.TotalRingSignatures,
		metrics.TotalMixingRounds,
		metrics.AverageRingSizeBps,
		metrics.AnonymitySetSize,
	))

	store.Set(key, data)
}

// GetPrivacyMetrics retrieves current privacy metrics
func (k Keeper) GetPrivacyMetrics(ctx context.Context) *PrivacyMetrics {
	if err := k.TrackPrivacyMetrics(ctx); err != nil {
		// If tracking fails, return zeroed metrics to avoid panics for callers.
		return &PrivacyMetrics{}
	}

	metrics := &PrivacyMetrics{
		TotalShieldedTxs:    k.countShieldedTransactions(ctx),
		TotalRingSignatures: k.countRingSignatures(ctx),
		TotalMixingRounds:   k.countMixingRounds(ctx),
		AverageRingSizeBps:  k.calculateAverageRingSizeBps(ctx),
		AnonymitySetSize:    k.getAnonymitySetSize(ctx),
	}

	return metrics
}
