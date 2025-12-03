package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dex/keeper"
	"github.com/aequitas/aura/chain/x/dex/types"
)

// TestFrontRunningProtection tests the front-running protection mechanism
func TestFrontRunningProtection(t *testing.T) {
	ctx, keeper := setupTest(t)

	address := "aura1testaddress"
	poolID := "uaura-usdt"

	// First trade should succeed
	err := keeper.CheckFrontRunningProtection(ctx, address, poolID)
	require.NoError(t, err)

	// Record the trade
	keeper.RecordTradeBlock(ctx, address, poolID)

	// Immediate second trade should fail
	err = keeper.CheckFrontRunningProtection(ctx, address, poolID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "front-running detected")

	// Advance blocks
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// Should now succeed
	err = keeper.CheckFrontRunningProtection(ctx, address, poolID)
	require.NoError(t, err)
}

// TestTWAPOracle tests the TWAP price oracle
func TestTWAPOracle(t *testing.T) {
	ctx, keeper := setupTest(t)

	poolID := "uaura-usdt"

	// Create a pool
	pool := &types.LiquidityPool{
		PoolId:   poolID,
		DenomA:   "uaura",
		DenomB:   "usdt",
		ReserveA: "1000000",
		ReserveB: "200000",
	}
	keeper.SetPool(ctx, pool)

	// Record first observation
	err := keeper.RecordTWAPObservation(ctx, poolID)
	require.NoError(t, err)

	// Advance time and update reserves
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(60 * time.Second))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 10)

	pool.ReserveA = "1100000"
	pool.ReserveB = "220000"
	keeper.SetPool(ctx, pool)

	// Record second observation
	err = keeper.RecordTWAPObservation(ctx, poolID)
	require.NoError(t, err)

	// Calculate TWAP
	twap, err := keeper.GetTWAPPrice(ctx, poolID, 20)
	require.NoError(t, err)
	require.True(t, twap.GT(math.LegacyZeroDec()))

	// TWAP should be between the two spot prices
	spotPrice1 := math.LegacyNewDec(200000).Quo(math.LegacyNewDec(1000000))
	spotPrice2 := math.LegacyNewDec(220000).Quo(math.LegacyNewDec(1100000))
	minPrice := spotPrice1
	if spotPrice2.LT(minPrice) {
		minPrice = spotPrice2
	}
	maxPrice := spotPrice1
	if spotPrice2.GT(maxPrice) {
		maxPrice = spotPrice2
	}
	require.True(t, twap.GTE(minPrice))
	require.True(t, twap.LTE(maxPrice))
}

// TestFlashLoanProtection tests flash loan attack prevention
func TestFlashLoanProtection(t *testing.T) {
	ctx, keeper := setupTest(t)

	provider := "aura1provider"
	poolID := "uaura-usdt"

	// First add liquidity should succeed
	err := keeper.CheckFlashLoanProtection(ctx, provider, poolID, true)
	require.NoError(t, err)

	// Record liquidity operation
	keeper.RecordLiquidityBlock(ctx, provider, poolID)

	// Immediate remove should fail
	err = keeper.CheckFlashLoanProtection(ctx, provider, poolID, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "flash loan")

	// Advance blocks
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 6)

	// Should now succeed
	err = keeper.CheckFlashLoanProtection(ctx, provider, poolID, false)
	require.NoError(t, err)
}

// TestMEVProtection tests MEV mitigation
func TestMEVProtection(t *testing.T) {
	ctx, keeper := setupTest(t)

	address := "aura1trader"
	poolID := "uaura-usdt"

	// First few trades should succeed
	for i := 0; i < 5; i++ {
		err := keeper.CheckMEVProtection(ctx, address)
		require.NoError(t, err)
		keeper.RecordTradeBlock(ctx, address, poolID)
	}

	// 6th trade in same block should fail
	err := keeper.CheckMEVProtection(ctx, address)
	require.Error(t, err)
	require.Contains(t, err.Error(), "MEV")

	// New block should reset counter
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	err = keeper.CheckMEVProtection(ctx, address)
	require.NoError(t, err)
}

// TestPoolSlippageLimit tests slippage limits
func TestPoolSlippageLimit(t *testing.T) {
	ctx, keeper := setupTest(t)

	poolID := "uaura-usdt"

	// 5% price impact should succeed
	priceImpact := math.LegacyNewDecWithPrec(5, 0)
	err := keeper.CheckPoolSlippageLimit(ctx, poolID, priceImpact)
	require.NoError(t, err)

	// 15% price impact should fail (exceeds 10% threshold)
	priceImpact = math.LegacyNewDecWithPrec(15, 0)
	err = keeper.CheckPoolSlippageLimit(ctx, poolID, priceImpact)
	require.Error(t, err)
	require.Contains(t, err.Error(), "price impact")
}

// TestMaxTradeSize tests maximum trade size caps
func TestMaxTradeSize(t *testing.T) {
	ctx, keeper := setupTest(t)

	poolID := "uaura-usdt"

	// Create pool with 1M reserves
	pool := &types.LiquidityPool{
		PoolId:   poolID,
		ReserveA: "1000000",
		ReserveB: "200000",
	}
	keeper.SetPool(ctx, pool)

	// 10% of pool should succeed (max is 20%)
	tradeSize := math.NewInt(100000)
	err := keeper.CheckMaxTradeSize(ctx, poolID, tradeSize)
	require.NoError(t, err)

	// 25% of pool should fail
	tradeSize = math.NewInt(250000)
	err = keeper.CheckMaxTradeSize(ctx, poolID, tradeSize)
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeds maximum")
}

// TestPriceImpactRejection tests price impact thresholds
func TestPriceImpactRejection(t *testing.T) {
	ctx, keeper := setupTest(t)

	// 8% price impact should succeed
	priceImpact := math.LegacyNewDecWithPrec(8, 0)
	err := keeper.CheckPriceImpactThreshold(ctx, priceImpact)
	require.NoError(t, err)

	// 12% price impact should fail
	priceImpact = math.LegacyNewDecWithPrec(12, 0)
	err = keeper.CheckPriceImpactThreshold(ctx, priceImpact)
	require.Error(t, err)
	require.Contains(t, err.Error(), "price impact")
}

// TestLiquidityLockup tests liquidity lock-up periods
func TestLiquidityLockup(t *testing.T) {
	ctx, keeper := setupTest(t)

	provider := "aura1provider"
	poolID := "uaura-usdt"
	lpTokens := math.NewInt(1000)

	// Create lock
	err := keeper.CreateLiquidityLock(ctx, provider, poolID, lpTokens)
	require.NoError(t, err)

	// Immediate removal should fail
	err = keeper.CheckLiquidityLock(ctx, provider, poolID, lpTokens)
	require.Error(t, err)
	require.Contains(t, err.Error(), "locked")

	// Advance time past lockup period (24 hours)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(25 * time.Hour))

	// Should now succeed
	err = keeper.CheckLiquidityLock(ctx, provider, poolID, lpTokens)
	require.NoError(t, err)

	// Lock should be marked inactive
	lock := keeper.GetLiquidityLock(ctx, provider, poolID)
	require.NotNil(t, lock)
	require.False(t, lock.IsActive)
}

// TestOrderManipulationDetection tests orderbook manipulation detection
func TestOrderManipulationDetection(t *testing.T) {
	ctx, keeper := setupTest(t)

	address := "aura1manipulator"
	poolID := "uaura-usdt"

	// Normal order size (within variance)
	orderSize := math.NewInt(1000)
	err := keeper.DetectOrderManipulation(ctx, address, poolID, orderSize)
	require.NoError(t, err)

	// Extremely large order (potential spoofing)
	// This would fail if there's order history
	orderSize = math.NewInt(1000000)
	// First order has no history, so it passes
	err = keeper.DetectOrderManipulation(ctx, address, poolID, orderSize)
	require.NoError(t, err)
}

// TestWashTradingDetection tests wash trading detection
func TestWashTradingDetection(t *testing.T) {
	ctx, keeper := setupTest(t)

	address := "aura1washtrader"
	poolID := "uaura-usdt"

	// First trade should succeed
	err := keeper.DetectWashTrading(ctx, address, poolID)
	require.NoError(t, err)

	// Update trade history
	keeper.UpdateTradeHistory(ctx, address, poolID)

	// Immediate second trade (within min interval) should be flagged
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(30 * time.Second))
	err = keeper.DetectWashTrading(ctx, address, poolID)
	// First occurrence won't error, just increment counter
	require.NoError(t, err)

	// After multiple suspicious trades, should error
	for i := 0; i < 5; i++ {
		keeper.IncrementWashTradeDetection(ctx, address, poolID)
	}

	err = keeper.DetectWashTrading(ctx, address, poolID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "wash trading")
}

// TestDustAttackPrevention tests dust attack prevention
func TestDustAttackPrevention(t *testing.T) {
	ctx, keeper := setupTest(t)

	// Normal trade amount should succeed
	amount := math.NewInt(10_000000)
	err := keeper.CheckDustAttack(ctx, amount)
	require.NoError(t, err)

	// Dust amount should fail
	amount = math.NewInt(100)
	err = keeper.CheckDustAttack(ctx, amount)
	require.Error(t, err)
	require.Contains(t, err.Error(), "dust")
}

// TestPoolCreationLimits tests pool creation limits and validation
func TestPoolCreationLimits(t *testing.T) {
	ctx, keeper := setupTest(t)

	creator := "aura1creator"

	// Insufficient liquidity should fail
	liquidity := math.NewInt(100)
	err := keeper.CheckPoolCreationLimits(ctx, creator, liquidity)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient")

	// Sufficient liquidity should succeed
	liquidity = math.NewInt(2000_000000)
	err = keeper.CheckPoolCreationLimits(ctx, creator, liquidity)
	require.NoError(t, err)

	// Record pool creation
	keeper.RecordPoolCreation(ctx, creator, "pool1", "uaura", "usdt", liquidity, liquidity)

	// Immediate second pool should fail (cooldown)
	err = keeper.CheckPoolCreationLimits(ctx, creator, liquidity)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cooldown")

	// After cooldown, should succeed
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Hour))
	err = keeper.CheckPoolCreationLimits(ctx, creator, liquidity)
	require.NoError(t, err)

	// Create max pools
	for i := 1; i < 10; i++ {
		poolID := fmt.Sprintf("pool%d", i)
		keeper.RecordPoolCreation(ctx, creator, poolID, "uaura", "usdt", liquidity, liquidity)
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Hour))
	}

	// Should now hit max pools limit
	err = keeper.CheckPoolCreationLimits(ctx, creator, liquidity)
	require.Error(t, err)
	require.Contains(t, err.Error(), "maximum pools")
}

// TestCircuitBreaker tests emergency pause functionality
func TestCircuitBreaker(t *testing.T) {
	ctx, keeper := setupTest(t)

	poolID := "uaura-usdt"

	// Initially not active
	active := keeper.IsCircuitBreakerActive(ctx, poolID)
	require.False(t, active)

	// Activate circuit breaker
	keeper.ActivateCircuitBreaker(ctx, "governance", "emergency", []string{poolID})

	// Should now be active
	active = keeper.IsCircuitBreakerActive(ctx, poolID)
	require.True(t, active)

	// Deactivate
	keeper.DeactivateCircuitBreaker(ctx)

	// Should be inactive again
	active = keeper.IsCircuitBreakerActive(ctx, poolID)
	require.False(t, active)
}

// TestCircuitBreakerAllPools tests global pause
func TestCircuitBreakerAllPools(t *testing.T) {
	ctx, keeper := setupTest(t)

	// Activate for all pools (empty list)
	keeper.ActivateCircuitBreaker(ctx, "governance", "systemic risk", []string{})

	// Any pool should be affected
	active := keeper.IsCircuitBreakerActive(ctx, "pool1")
	require.True(t, active)

	active = keeper.IsCircuitBreakerActive(ctx, "pool2")
	require.True(t, active)
}

// TestTWAPPruning tests TWAP observation pruning
func TestTWAPPruning(t *testing.T) {
	ctx, keeper := setupTest(t)

	poolID := "uaura-usdt"

	pool := &types.LiquidityPool{
		PoolId:   poolID,
		ReserveA: "1000000",
		ReserveB: "200000",
	}
	keeper.SetPool(ctx, pool)

	// Record many observations
	for i := 0; i < 150; i++ {
		keeper.RecordTWAPObservation(ctx, poolID)
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(10 * time.Second))
	}

	// Prune old observations
	keeper.PruneTWAPObservations(ctx, poolID)

	// Should only have observations within window (100 blocks)
	observations := keeper.GetTWAPObservations(ctx, poolID, 100)
	require.LessOrEqual(t, len(observations), 100)
}

// TestSecurityParamsDefaults tests default security parameters
func TestSecurityParamsDefaults(t *testing.T) {
	params := types.DefaultSecurityParams()

	require.Equal(t, uint64(2), params.MinBlockDelay)
	require.Equal(t, math.LegacyNewDecWithPrec(20, 2).String(), params.MaxTradeSizePercent)
	require.Equal(t, math.LegacyNewDecWithPrec(10, 0).String(), params.MaxPriceImpactPercent)
	require.Equal(t, int64(86400), params.LiquidityLockupSeconds)
	require.Equal(t, int64(3600), params.PoolCreationCooldown)
	require.Equal(t, uint64(10), params.MaxPoolsPerCreator)
	require.Equal(t, uint64(100), params.TwapWindowBlocks)
	require.Equal(t, math.NewInt(1000_000000).String(), params.MinPoolCreationLiquidity)
	require.Equal(t, uint64(5), params.MinLiquidityBlocks)
	require.Equal(t, int64(60), params.WashTradeMinInterval)
	require.Equal(t, math.NewInt(1_000000).String(), params.MinTradeAmount)
	require.True(t, params.CircuitBreakerEnabled)
	require.True(t, params.MevProtectionEnabled)
	require.Equal(t, uint64(5), params.MaxSwapsPerBlock)
}

// Helper function to setup test environment
func setupTest(t *testing.T) (sdk.Context, *keeper.Keeper) {
	// Use proper test setup with keeper test utilities
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil, // bankKeeper
		nil, // authKeeper
		nil, // vcKeeper
	)
	return input.Ctx, k
}
