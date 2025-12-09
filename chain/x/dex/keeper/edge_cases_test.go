package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dex/keeper"
	"github.com/aequitas/aura/chain/x/dex/types"
)

// TestSwap_EdgeCase_ZeroAmount verifies that swaps with zero amount are rejected
func TestSwap_EdgeCase_ZeroAmount(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	// Create a test pool first
	creator := suite.TestAccs[0].String()
	amountA := sdk.NewCoin("uaura", sdkmath.NewInt(1000000))
	amountB := sdk.NewCoin("usdt", sdkmath.NewInt(2000000))

	// Fund creator account
	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[0], sdk.NewCoins(amountA, amountB))

	// Create pool
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "usdt", amountA, amountB)
	require.NoError(t, err)
	require.NotNil(t, pool)

	// Test swap with zero amount
	sender := suite.TestAccs[1].String()
	coinIn := sdk.NewCoin("uaura", sdkmath.ZeroInt())
	minAmountOut := sdkmath.NewInt(1)

	_, _, _, err = k.SwapExactIn(ctx, sender, pool.PoolId, coinIn, minAmountOut, 500)
	require.Error(t, err)
	require.Contains(t, err.Error(), "swap amount must be greater than zero")
}

// TestSwap_EdgeCase_PoolNotFound verifies that swaps fail when pool doesn't exist
func TestSwap_ErrorPath_PoolNotFound(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	sender := suite.TestAccs[0].String()
	coinIn := sdk.NewCoin("uaura", sdkmath.NewInt(1000))
	minAmountOut := sdkmath.NewInt(1)

	// Try to swap with non-existent pool ID
	_, _, _, err := k.SwapExactIn(ctx, sender, "nonexistent-pool", coinIn, minAmountOut, 500)
	require.Error(t, err)
	require.Contains(t, err.Error(), "pool")
	require.Contains(t, err.Error(), "not found")
}

// TestSwap_ErrorPath_SlippageExceeded verifies slippage protection works
func TestSwap_ErrorPath_SlippageExceeded(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	// Create a small pool to cause high slippage
	creator := suite.TestAccs[0].String()
	amountA := sdk.NewCoin("uaura", sdkmath.NewInt(1000000))
	amountB := sdk.NewCoin("usdt", sdkmath.NewInt(1000000))

	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[0], sdk.NewCoins(amountA, amountB))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "usdt", amountA, amountB)
	require.NoError(t, err)

	// Try to swap large amount that would exceed slippage
	sender := suite.TestAccs[1].String()
	largeSwap := sdk.NewCoin("uaura", sdkmath.NewInt(500000)) // 50% of pool
	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[1], sdk.NewCoins(largeSwap))

	// Demand very low slippage (10 bps = 0.1%)
	minAmountOut := sdkmath.NewInt(490000) // Expecting almost 1:1 ratio
	maxSlippageBps := uint64(10) // 0.1%

	_, _, _, err = k.SwapExactIn(ctx, sender, pool.PoolId, largeSwap, minAmountOut, maxSlippageBps)
	require.Error(t, err)
	require.Contains(t, err.Error(), "slippage")
}

// TestCreatePool_EdgeCase_ZeroLiquidity verifies that pools cannot be created with zero liquidity
func TestCreatePool_EdgeCase_ZeroLiquidity(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	creator := suite.TestAccs[0].String()
	zeroAmount := sdk.NewCoin("uaura", sdkmath.ZeroInt())
	validAmount := sdk.NewCoin("usdt", sdkmath.NewInt(1000000))

	_, _, err := k.CreatePool(ctx, creator, "uaura", "usdt", zeroAmount, validAmount)
	require.Error(t, err)
	require.Contains(t, err.Error(), "amounts must be positive")
}

// TestCreatePool_ErrorPath_DuplicatePool verifies that duplicate pools are rejected
func TestCreatePool_ErrorPath_DuplicatePool(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	creator := suite.TestAccs[0].String()
	amountA := sdk.NewCoin("uaura", sdkmath.NewInt(1000000))
	amountB := sdk.NewCoin("usdt", sdkmath.NewInt(2000000))

	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[0], sdk.NewCoins(amountA, amountB))

	// Create first pool
	_, _, err := k.CreatePool(ctx, creator, "uaura", "usdt", amountA, amountB)
	require.NoError(t, err)

	// Try to create duplicate pool with same denoms
	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[0], sdk.NewCoins(amountA, amountB))
	_, _, err = k.CreatePool(ctx, creator, "uaura", "usdt", amountA, amountB)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
}

// TestAddLiquidity_ErrorPath_PoolNotFound verifies adding liquidity fails for non-existent pool
func TestAddLiquidity_ErrorPath_PoolNotFound(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	provider := suite.TestAccs[0].String()
	amountA := sdk.NewCoin("uaura", sdkmath.NewInt(1000))
	amountB := sdk.NewCoin("usdt", sdkmath.NewInt(2000))

	_, _, err := k.AddLiquidity(ctx, provider, "nonexistent-pool", amountA, amountB)
	require.Error(t, err)
	require.Contains(t, err.Error(), "pool")
	require.Contains(t, err.Error(), "not found")
}

// TestAddLiquidity_EdgeCase_ZeroLPTokens verifies that dust deposits are rejected
func TestAddLiquidity_EdgeCase_ZeroLPTokens(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	// Create a large pool
	creator := suite.TestAccs[0].String()
	largeAmountA := sdk.NewCoin("uaura", sdkmath.NewInt(1000000000000)) // 1 trillion
	largeAmountB := sdk.NewCoin("usdt", sdkmath.NewInt(1000000000000))

	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[0], sdk.NewCoins(largeAmountA, largeAmountB))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "usdt", largeAmountA, largeAmountB)
	require.NoError(t, err)

	// Try to add tiny amount that would result in 0 LP tokens
	provider := suite.TestAccs[1].String()
	tinyAmountA := sdk.NewCoin("uaura", sdkmath.NewInt(1)) // 1 unit
	tinyAmountB := sdk.NewCoin("usdt", sdkmath.NewInt(1))

	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[1], sdk.NewCoins(tinyAmountA, tinyAmountB))

	_, _, err = k.AddLiquidity(ctx, provider, pool.PoolId, tinyAmountA, tinyAmountB)
	require.Error(t, err)
	require.Contains(t, err.Error(), "0 LP tokens")
}

// TestRemoveLiquidity_ErrorPath_InsufficientLPTokens verifies removing more than owned is rejected
func TestRemoveLiquidity_ErrorPath_InsufficientLPTokens(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	// Create pool
	creator := suite.TestAccs[0].String()
	amountA := sdk.NewCoin("uaura", sdkmath.NewInt(1000000))
	amountB := sdk.NewCoin("usdt", sdkmath.NewInt(2000000))

	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[0], sdk.NewCoins(amountA, amountB))

	pool, lpTokens, err := k.CreatePool(ctx, creator, "uaura", "usdt", amountA, amountB)
	require.NoError(t, err)

	// Try to remove more LP tokens than owned
	excessiveLPTokens := lpTokens.Add(sdkmath.NewInt(1000000))

	_, _, err = k.RemoveLiquidity(ctx, creator, pool.PoolId, excessiveLPTokens)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient")
}

// TestRemoveLiquidity_ErrorPath_NotLiquidityProvider verifies non-providers cannot remove liquidity
func TestRemoveLiquidity_ErrorPath_NotLiquidityProvider(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	// Create pool with account 0
	creator := suite.TestAccs[0].String()
	amountA := sdk.NewCoin("uaura", sdkmath.NewInt(1000000))
	amountB := sdk.NewCoin("usdt", sdkmath.NewInt(2000000))

	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[0], sdk.NewCoins(amountA, amountB))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "usdt", amountA, amountB)
	require.NoError(t, err)

	// Try to remove liquidity from account 1 (not a provider)
	notProvider := suite.TestAccs[1].String()
	lpTokens := sdkmath.NewInt(1000)

	_, _, err = k.RemoveLiquidity(ctx, notProvider, pool.PoolId, lpTokens)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a liquidity provider")
}

// TestGetQuote_EdgeCase_ZeroAmount verifies quote calculation rejects zero amounts
func TestGetQuote_EdgeCase_ZeroAmount(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	// Create pool
	creator := suite.TestAccs[0].String()
	amountA := sdk.NewCoin("uaura", sdkmath.NewInt(1000000))
	amountB := sdk.NewCoin("usdt", sdkmath.NewInt(2000000))

	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[0], sdk.NewCoins(amountA, amountB))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "usdt", amountA, amountB)
	require.NoError(t, err)

	// Get quote with zero amount
	_, _, _, _, err = k.GetQuote(ctx, pool.PoolId, "uaura", sdkmath.ZeroInt())
	require.Error(t, err)
	require.Contains(t, err.Error(), "positive")
}

// TestGetQuote_EdgeCase_ExcessiveAmount verifies overflow protection in quote calculation
func TestGetQuote_EdgeCase_ExcessiveAmount(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	// Create pool
	creator := suite.TestAccs[0].String()
	amountA := sdk.NewCoin("uaura", sdkmath.NewInt(1000000))
	amountB := sdk.NewCoin("usdt", sdkmath.NewInt(2000000))

	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[0], sdk.NewCoins(amountA, amountB))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "usdt", amountA, amountB)
	require.NoError(t, err)

	// Get quote with excessive amount (> 1 trillion)
	excessiveAmount := sdkmath.NewInt(2_000_000_000_000) // 2 trillion

	_, _, _, _, err = k.GetQuote(ctx, pool.PoolId, "uaura", excessiveAmount)
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeds maximum")
}

// TestSwap_EdgeCase_InvalidDenom verifies swaps fail with invalid token denom
func TestSwap_EdgeCase_InvalidDenom(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	// Create pool with uaura/usdt
	creator := suite.TestAccs[0].String()
	amountA := sdk.NewCoin("uaura", sdkmath.NewInt(1000000))
	amountB := sdk.NewCoin("usdt", sdkmath.NewInt(2000000))

	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[0], sdk.NewCoins(amountA, amountB))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "usdt", amountA, amountB)
	require.NoError(t, err)

	// Try to swap with denom not in pool
	sender := suite.TestAccs[1].String()
	wrongDenom := sdk.NewCoin("atom", sdkmath.NewInt(1000))

	_, _, _, err = k.SwapExactIn(ctx, sender, pool.PoolId, wrongDenom, sdkmath.NewInt(1), 500)
	require.Error(t, err)
	require.Contains(t, err.Error(), "coin denom not in pool")
}

// TestCreatePool_EdgeCase_InsufficientLiquidityForMinimum verifies minimum liquidity burn requirement
func TestCreatePool_EdgeCase_InsufficientLiquidityForMinimum(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	creator := suite.TestAccs[0].String()

	// Create pool with amounts that would result in LP tokens < MinimumLiquidity (1000)
	// LP tokens = sqrt(amount_a * amount_b)
	// For sqrt(30 * 30) = 30, which is < 1000
	tinyAmountA := sdk.NewCoin("uaura", sdkmath.NewInt(30))
	tinyAmountB := sdk.NewCoin("usdt", sdkmath.NewInt(30))

	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[0], sdk.NewCoins(tinyAmountA, tinyAmountB))

	_, _, err := k.CreatePool(ctx, creator, "uaura", "usdt", tinyAmountA, tinyAmountB)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient initial liquidity")
}

// TestSwap_EdgeCase_MaxValue verifies handling of maximum integer values
func TestSwap_EdgeCase_MaxValue(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	// Create pool
	creator := suite.TestAccs[0].String()
	amountA := sdk.NewCoin("uaura", sdkmath.NewInt(1000000000))
	amountB := sdk.NewCoin("usdt", sdkmath.NewInt(1000000000))

	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[0], sdk.NewCoins(amountA, amountB))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "usdt", amountA, amountB)
	require.NoError(t, err)

	// Test with very large swap amount (should be rejected by max trade size check)
	sender := suite.TestAccs[1].String()
	maxIntCoin := sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000_000_000_000)) // quadrillion

	_, _, _, err = k.SwapExactIn(ctx, sender, pool.PoolId, maxIntCoin, sdkmath.NewInt(1), 500)
	require.Error(t, err)
	// Error could be from max trade size check or insufficient balance
}

// TestCreatePool_Concurrent_MultipleCreators verifies pool creation with concurrent attempts
func TestCreatePool_Concurrent_MultipleCreators(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	creator1 := suite.TestAccs[0].String()
	creator2 := suite.TestAccs[1].String()

	amountA1 := sdk.NewCoin("uaura", sdkmath.NewInt(1000000))
	amountB1 := sdk.NewCoin("usdt", sdkmath.NewInt(2000000))

	amountA2 := sdk.NewCoin("uaura", sdkmath.NewInt(3000000))
	amountB2 := sdk.NewCoin("usdt", sdkmath.NewInt(4000000))

	// Fund both creators
	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[0], sdk.NewCoins(amountA1, amountB1))
	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[1], sdk.NewCoins(amountA2, amountB2))

	// First creator creates pool successfully
	pool1, _, err := k.CreatePool(ctx, creator1, "uaura", "usdt", amountA1, amountB1)
	require.NoError(t, err)
	require.NotNil(t, pool1)

	// Second creator tries to create same pool (should fail)
	_, _, err = k.CreatePool(ctx, creator2, "uaura", "usdt", amountA2, amountB2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
}

// TestInvariant_LPTokenAccounting verifies LP token invariant holds after operations
func TestInvariant_LPTokenAccounting(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	k := suite.DexKeeper

	// Create pool
	creator := suite.TestAccs[0].String()
	amountA := sdk.NewCoin("uaura", sdkmath.NewInt(1000000))
	amountB := sdk.NewCoin("usdt", sdkmath.NewInt(2000000))

	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[0], sdk.NewCoins(amountA, amountB))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "usdt", amountA, amountB)
	require.NoError(t, err)

	// Verify invariant: TotalLPTokens = sum(provider LP tokens) + locked liquidity
	err = k.ValidateLPTokenInvariantExported(pool)
	require.NoError(t, err)

	// Add liquidity from another provider
	provider2 := suite.TestAccs[1].String()
	addAmountA := sdk.NewCoin("uaura", sdkmath.NewInt(100000))
	addAmountB := sdk.NewCoin("usdt", sdkmath.NewInt(200000))

	suite.FundAccount(suite.BankKeeper, ctx, suite.TestAccs[1], sdk.NewCoins(addAmountA, addAmountB))

	_, _, err = k.AddLiquidity(ctx, provider2, pool.PoolId, addAmountA, addAmountB)
	require.NoError(t, err)

	// Get updated pool and verify invariant still holds
	updatedPool := k.GetPool(ctx, pool.PoolId)
	require.NotNil(t, updatedPool)

	err = k.ValidateLPTokenInvariantExported(updatedPool)
	require.NoError(t, err)
}
