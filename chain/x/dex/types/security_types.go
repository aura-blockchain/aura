package types

import (
	"cosmossdk.io/math"
	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

// DefaultSecurityParams returns default security parameters
func DefaultSecurityParams() *dexpb.SecurityParams {
	return &dexpb.SecurityParams{
		MinBlockDelay:            2,                                         // 2 blocks between trades
		MaxTradeSizePercent:      math.LegacyNewDecWithPrec(20, 2).String(), // 20% of pool
		MaxPriceImpactPercent:    math.LegacyNewDecWithPrec(10, 0).String(), // 10%
		LiquidityLockupSeconds:   86400,                                     // 24 hours
		PoolCreationCooldown:     3600,                                      // 1 hour
		MaxPoolsPerCreator:       10,                                        // Max 10 pools per address
		TwapWindowBlocks:         100,                                       // 100 block TWAP window
		MinPoolCreationLiquidity: math.NewInt(1000_000000).String(),         // 1000 tokens minimum
		MinLiquidityBlocks:       5,                                         // 5 blocks between add/remove
		WashTradeMinInterval:     60,                                        // 60 seconds between trades
		MinTradeAmount:           math.NewInt(1_000000).String(),            // 1 token minimum
		MaxOrderVariance:         math.LegacyNewDecWithPrec(50, 2).String(), // 50% variance allowed
		CircuitBreakerEnabled:    true,                                      // Enable emergency pause
		MevProtectionEnabled:     true,                                      // Enable MEV protection
		MaxSwapsPerBlock:         5,                                         // Max 5 swaps per block
	}
}
