package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// InitGenesis initializes the dex module state from genesis.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) error {
	if data.Params == nil {
		data.Params = types.DefaultParams()
	}
	if err := types.ValidateGenesis(&data); err != nil {
		return err
	}

	if err := k.SetParams(ctx, data.Params); err != nil {
		return err
	}

	for _, pool := range data.LiquidityPools {
		if pool == nil {
			continue
		}
		k.SetPool(ctx, pool)
	}

	for _, order := range data.SwapOrders {
		if order == nil {
			continue
		}
		k.SetOrder(ctx, order)
		if order.Status == types.SwapOrderStatus_PENDING {
			k.AddToOrderbook(ctx, order)
		}
	}

	for _, book := range data.Orderbooks {
		if book == nil {
			continue
		}
		for _, entry := range book.BuyOrders {
			k.addPendingOrderToIndex(ctx, entry)
		}
		for _, entry := range book.SellOrders {
			k.addPendingOrderToIndex(ctx, entry)
		}
	}

	for _, stats := range data.SwapStats {
		if stats == nil {
			continue
		}
		k.setSwapStats(ctx, stats)
	}

	for _, price := range data.MarketPrices {
		if price == nil {
			continue
		}
		k.setMarketPrice(ctx, price)
	}

	// Import pool creation records (audit trail)
	for _, record := range data.PoolCreationRecords {
		if record == nil {
			continue
		}
		store := ctx.KVStore(k.storeKey)
		key := types.PoolCreationKey(record.Creator)
		bz := k.cdc.MustMarshal(record)
		store.Set(key, bz)
	}

	// Import order commitments (commit-reveal scheme)
	for _, commitment := range data.OrderCommitments {
		if commitment == nil {
			continue
		}
		k.SetOrderCommitment(ctx, commitment)
	}

	// Import queued orders (batch execution)
	for _, queuedOrder := range data.QueuedOrders {
		if queuedOrder == nil || queuedOrder.Order == nil {
			continue
		}
		k.QueueOrderForBatch(ctx, queuedOrder.Order, queuedOrder.Salt)
	}

	return nil
}

// ExportGenesis exports the dex module state to genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	params := k.GetParams(ctx)
	pools := k.GetAllPools(ctx)
	orders := k.GetAllOrders(ctx)
	orderbooks := k.exportOrderbooks(ctx)
	swapStats := k.GetAllSwapStats(ctx)
	marketPrices := k.GetAllMarketPrices(ctx)
	poolCreationRecords := k.GetAllPoolCreationRecords(ctx)
	orderCommitments := k.GetAllOrderCommitments(ctx)
	queuedOrders := k.GetAllQueuedOrders(ctx)

	return types.GenesisState{
		Params:              params,
		LiquidityPools:      pools,
		SwapOrders:          orders,
		Orderbooks:          orderbooks,
		MarketPrices:        marketPrices,
		SwapStats:           swapStats,
		PoolCreationRecords: poolCreationRecords,
		OrderCommitments:    orderCommitments,
		QueuedOrders:        queuedOrders,
	}
}
