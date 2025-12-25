// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"cosmossdk.io/math"

	v1beta1 "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

// Note: Params module is deprecated in Cosmos SDK v0.50+
// Params are now stored directly in module state

// DefaultParams returns default parameters
func DefaultParams() v1beta1.Params {
	return v1beta1.Params{
		TradingFee:                math.LegacyMustNewDecFromStr("0.003"),
		ProtocolFee:               math.LegacyMustNewDecFromStr("0.0005"),
		MinLiquidityTiers:         []v1beta1.MinLiquidityTier{},
		MaxSlippageBps:            10000,
		MinSwapAmount:             math.NewInt(1000000),
		IrBoostEnabled:            true,
		IrBoostPercent:            40,
		BondingCurveEnabled:       true,
		BuybackBurnEnabled:        true,
		BuybackPercent:            100,
		CommitRevealThreshold:     math.NewInt(10000000000), // 10,000 AURA (large orders require commit-reveal)
		CommitRevealWindow:        60,                       // 60 seconds to reveal
		BatchExecutionEnabled:     true,
		BatchExecutionInterval:    5,                                       // Execute batch every 5 blocks
		GovernanceFallbackPrice:   math.LegacyNewDecWithPrec(10, 2),        // $0.10 USD default fallback price
	}
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *v1beta1.GenesisState {
	params := DefaultParams()
	return &v1beta1.GenesisState{
		Params:                params,
		LiquidityPools:        []v1beta1.LiquidityPool{},
		SwapOrders:            []v1beta1.SwapOrder{},
		Orderbooks:            []v1beta1.Orderbook{},
		MarketPrices:          []v1beta1.MarketPrice{},
		SwapStats:             []v1beta1.SwapStats{},
		OrderCommitments:      []v1beta1.OrderCommitment{},
		QueuedOrders:          []v1beta1.QueuedOrder{},
		PoolCreationRecords:   []v1beta1.PoolCreationRecord{},
	}
}
