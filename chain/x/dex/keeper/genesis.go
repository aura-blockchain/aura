// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// InitGenesis initializes the dex module state from genesis.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) error {
	if err := types.ValidateGenesis(&data); err != nil {
		return fmt.Errorf("error in InitGenesis for ValidateGenesis: %w", err)
	}

	if err := k.SetParams(ctx, &data.Params); err != nil {
		return fmt.Errorf("error in InitGenesis for ValidateGenesis: %w", err)
	}

	for i := range data.LiquidityPools {
		pool := &data.LiquidityPools[i]
		if err := k.SetPool(ctx, pool); err != nil {
			return fmt.Errorf("error in InitGenesis for LiquidityPools: %w", err)
		}
	}

	for i := range data.SwapOrders {
		order := &data.SwapOrders[i]
		if err := k.SetOrder(ctx, order); err != nil {
			return fmt.Errorf("error in InitGenesis for LiquidityPools: %w", err)
		}
		if order.Status == types.SwapOrderStatus_PENDING {
			k.AddToOrderbook(ctx, order)
		}
	}

	for i := range data.Orderbooks {
		book := &data.Orderbooks[i]
		for j := range book.BuyOrders {
			k.addPendingOrderToIndex(ctx, &book.BuyOrders[j])
		}
		for j := range book.SellOrders {
			k.addPendingOrderToIndex(ctx, &book.SellOrders[j])
		}
	}

	for i := range data.SwapStats {
		stats := &data.SwapStats[i]
		if err := k.setSwapStats(ctx, stats); err != nil {
			return fmt.Errorf("error in InitGenesis: %w", err)
		}
	}

	for i := range data.MarketPrices {
		price := &data.MarketPrices[i]
		if err := k.setMarketPrice(ctx, price); err != nil {
			return fmt.Errorf("error in InitGenesis: %w", err)
		}
	}

	// Import pool creation records (audit trail)
	for i := range data.PoolCreationRecords {
		record := &data.PoolCreationRecords[i]
		store := ctx.KVStore(k.storeKey)
		key := types.PoolCreationKey(record.Creator)
		bz, err := k.cdc.Marshal(record)
		if err != nil {
			return types.ErrMarshalFailed.Wrapf("failed to marshal pool creation record for creator %s: %v", record.Creator, err)
		}
		store.Set(key, bz)
	}

	// Import order commitments (commit-reveal scheme)
	for i := range data.OrderCommitments {
		commitment := &data.OrderCommitments[i]
		if err := k.SetOrderCommitment(ctx, commitment); err != nil {
			return fmt.Errorf("failed to marshal: %w", err)
		}
	}

	// Import queued orders (batch execution)
	for i := range data.QueuedOrders {
		queuedOrder := &data.QueuedOrders[i]
		if err := k.QueueOrderForBatch(ctx, &queuedOrder.Order, queuedOrder.Salt); err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}
	}

	return nil
}

// ExportGenesis exports the dex module state to genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		ctx.Logger().Error("failed to get params for genesis export, using defaults", "error", err)
		params = types.DefaultParams()
	}
	poolsPtrs := k.GetAllPools(ctx)
	ordersPtrs := k.GetAllOrders(ctx)
	orderbooksPtrs := k.exportOrderbooks(ctx)
	swapStatsPtrs := k.GetAllSwapStats(ctx)
	marketPricesPtrs := k.GetAllMarketPrices(ctx)
	poolCreationRecordsPtrs := k.GetAllPoolCreationRecords(ctx)
	orderCommitmentsPtrs := k.GetAllOrderCommitments(ctx)
	queuedOrdersPtrs := k.GetAllQueuedOrders(ctx)

	// Convert pointer slices to value slices for genesis state
	pools := make([]types.LiquidityPool, len(poolsPtrs))
	for i, p := range poolsPtrs {
		pools[i] = *p
	}

	orders := make([]types.SwapOrder, len(ordersPtrs))
	for i, o := range ordersPtrs {
		orders[i] = *o
	}

	orderbooks := make([]types.Orderbook, len(orderbooksPtrs))
	for i, ob := range orderbooksPtrs {
		orderbooks[i] = *ob
	}

	swapStats := make([]types.SwapStats, len(swapStatsPtrs))
	for i, s := range swapStatsPtrs {
		swapStats[i] = *s
	}

	marketPrices := make([]types.MarketPrice, len(marketPricesPtrs))
	for i, mp := range marketPricesPtrs {
		marketPrices[i] = *mp
	}

	poolCreationRecords := make([]types.PoolCreationRecord, len(poolCreationRecordsPtrs))
	for i, pcr := range poolCreationRecordsPtrs {
		poolCreationRecords[i] = *pcr
	}

	orderCommitments := make([]types.OrderCommitment, len(orderCommitmentsPtrs))
	for i, oc := range orderCommitmentsPtrs {
		orderCommitments[i] = *oc
	}

	queuedOrders := make([]types.QueuedOrder, len(queuedOrdersPtrs))
	for i, qo := range queuedOrdersPtrs {
		queuedOrders[i] = *qo
	}

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
