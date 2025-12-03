package types

// Note: Params module is deprecated in Cosmos SDK v0.50+
// Params are now stored directly in module state

// DefaultParams returns default parameters
func DefaultParams() *Params {
	return &Params{
		TradingFee:             "0.003",
		ProtocolFee:            "0.0005",
		MinLiquidityTiers:      []*MinLiquidityTier{},
		MaxSlippageBps:         10000,
		MinSwapAmount:          "1000000",
		IrBoostEnabled:         true,
		IrBoostPercent:         40,
		BondingCurveEnabled:    true,
		BuybackBurnEnabled:     true,
		BuybackPercent:         100,
		CommitRevealThreshold:  "10000000000", // 10,000 AURA (large orders require commit-reveal)
		CommitRevealWindow:     60,            // 60 seconds to reveal
		BatchExecutionEnabled:  true,
		BatchExecutionInterval: 5, // Execute batch every 5 blocks
	}
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:         DefaultParams(),
		LiquidityPools: []*LiquidityPool{},
		SwapOrders:     []*SwapOrder{},
		Orderbooks:     []*Orderbook{},
		MarketPrices:   []*MarketPrice{},
		SwapStats:      []*SwapStats{},
	}
}
