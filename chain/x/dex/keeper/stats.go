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
	k.cdc.MustUnmarshal(bz, &stats)
	return &stats, true
}

func (k Keeper) GetMarketPrice(ctx sdk.Context, key string) (*types.MarketPrice, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.MarketPriceKey(key))
	if bz == nil {
		return nil, false
	}
	var price types.MarketPrice
	k.cdc.MustUnmarshal(bz, &price)
	return &price, true
}

func (k Keeper) GetAllSwapStats(ctx sdk.Context) []*types.SwapStats {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.SwapStatsPrefix)
	defer iter.Close()

	var stats []*types.SwapStats
	for ; iter.Valid(); iter.Next() {
		var entry types.SwapStats
		k.cdc.MustUnmarshal(iter.Value(), &entry)
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
		k.cdc.MustUnmarshal(iter.Value(), &price)
		prices = append(prices, &price)
	}
	return prices
}
