package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

func (k Keeper) setSwapStats(ctx sdk.Context, stats *types.SwapStats) {
	if stats == nil {
		return
	}
	store := ctx.KVStore(k.storeKey)
	store.Set(types.SwapStatsKey(stats.PoolId), k.cdc.MustMarshal(stats))
}

func (k Keeper) setMarketPrice(ctx sdk.Context, price *types.MarketPrice) {
	if price == nil {
		return
	}
	store := ctx.KVStore(k.storeKey)
	store.Set(types.MarketPriceKey(price.Coin), k.cdc.MustMarshal(price))
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

	var stats []*types.SwapStats
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

	var prices []*types.MarketPrice
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
