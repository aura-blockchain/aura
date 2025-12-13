package keeper

import (
	"context"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/aequitas/aura/chain/x/privacy/types"
)

// PrivacyMetrics tracks privacy-related metrics
type PrivacyMetrics struct {
	TotalShieldedTxs      int64
	TotalRingSignatures   int64
	TotalMixingRounds     int64
	AveragePrivacyScore   float64
	AverageRingSize       float64
	AnonymitySetSize      int64
	LastUpdated           time.Time
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
	metrics.AverageRingSize = k.calculateAverageRingSize(ctx)
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

func (k Keeper) calculateAverageRingSize(ctx context.Context) float64 {
	params := k.GetParams(ctx)
	return float64(params.MinRingSize+params.MaxRingSize) / 2.0
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

	data := []byte(fmt.Sprintf("%d,%d,%d,%.2f,%d",
		metrics.TotalShieldedTxs,
		metrics.TotalRingSignatures,
		metrics.TotalMixingRounds,
		metrics.AverageRingSize,
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
		AverageRingSize:     k.calculateAverageRingSize(ctx),
		AnonymitySetSize:    k.getAnonymitySetSize(ctx),
	}

	return metrics
}
