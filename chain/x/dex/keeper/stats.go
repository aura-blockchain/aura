// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

func (k Keeper) setSwapStats(ctx sdk.Context, stats *types.SwapStats) error {
	if stats == nil {
		return nil
	}
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(stats)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal swap stats for pool %s: %v", stats.PoolId, err)
	}
	store.Set(types.SwapStatsKey(stats.PoolId), bz)
	return nil
}

func (k Keeper) setMarketPrice(ctx sdk.Context, price *types.MarketPrice) error {
	if price == nil {
		return nil
	}
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(price)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal market price for coin %s: %v", price.Coin, err)
	}
	store.Set(types.MarketPriceKey(price.Coin), bz)
	return nil
}

func (k Keeper) GetSwapStats(ctx sdk.Context, poolID string) (*types.SwapStats, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.SwapStatsKey(poolID))
	if bz == nil {
		return nil, false
	}
	var stats types.SwapStats
	if err := k.cdc.Unmarshal(bz, &stats); err != nil {
		ctx.Logger().Error("failed to unmarshal swap stats",
			"pool_id", poolID,
			"error", err,
			"data_len", len(bz))
		return nil, false
	}
	return &stats, true
}

func (k Keeper) GetMarketPrice(ctx sdk.Context, key string) (*types.MarketPrice, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.MarketPriceKey(key))
	if bz == nil {
		return nil, false
	}
	var price types.MarketPrice
	if err := k.cdc.Unmarshal(bz, &price); err != nil {
		ctx.Logger().Error("failed to unmarshal market price",
			"key", key,
			"error", err,
			"data_len", len(bz))
		return nil, false
	}
	return &price, true
}

func (k Keeper) GetAllSwapStats(ctx sdk.Context) []*types.SwapStats {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.SwapStatsPrefix)
	defer iter.Close()

	stats := make([]*types.SwapStats, 0, 64)
	for ; iter.Valid(); iter.Next() {
		var entry types.SwapStats
		if err := k.cdc.Unmarshal(iter.Value(), &entry); err != nil {
			ctx.Logger().Error("failed to unmarshal swap stats entry, skipping",
				"error", err,
				"data_len", len(iter.Value()))
			continue
		}
		stats = append(stats, &entry)
	}
	return stats
}

func (k Keeper) GetAllMarketPrices(ctx sdk.Context) []*types.MarketPrice {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.MarketPricePrefix)
	defer iter.Close()

	prices := make([]*types.MarketPrice, 0, 64)
	for ; iter.Valid(); iter.Next() {
		var price types.MarketPrice
		if err := k.cdc.Unmarshal(iter.Value(), &price); err != nil {
			ctx.Logger().Error("failed to unmarshal market price entry, skipping",
				"error", err,
				"data_len", len(iter.Value()))
			continue
		}
		prices = append(prices, &price)
	}
	return prices
}
