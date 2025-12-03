package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/testutil/testdata"
	"github.com/aequitas/aura/chain/x/dex/keeper"
	"github.com/aequitas/aura/chain/x/dex/types"
)

// Helper to create keeper with mock bank keeper for slippage tests
func setupSlippageTestKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	input := keepertest.CreateTestInput(t)

	// Use the existing MockBankKeeper from keeper_comprehensive_test.go
	mockBank := NewMockBankKeeper()
	// Set up large balances for all addresses so tests don't fail due to insufficient funds
	// The mock will return zero by default, but swaps check balances

	k := keeper.NewKeeper(input.Cdc, input.StoreKey, mockBank, nil, nil)
	return k, input.Ctx
}

// SlippageProtectionTestSuite tests comprehensive slippage protection
type SlippageProtectionTestSuite struct {
	suite.Suite
	keeper *keeper.Keeper
	ctx    sdk.Context
}

func TestSlippageProtectionSuite(t *testing.T) {
	suite.Run(t, new(SlippageProtectionTestSuite))
}

func (suite *SlippageProtectionTestSuite) SetupTest() {
	suite.keeper, suite.ctx = setupSlippageTestKeeper(suite.T())
}

// TestSlippageEnforcementBeforeStateChanges verifies slippage check happens before any state modifications
func (suite *SlippageProtectionTestSuite) TestSlippageEnforcementBeforeStateChanges() {
	ctx := suite.ctx
	k := suite.keeper

	// Create pool
	creator := testdata.GenTestAddr().String()
	pool, _, err := k.CreatePool(
		ctx,
		creator,
		"uaura",
		"uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(1000000)),
	)
	suite.Require().NoError(err)

	// Get initial state
	initialPool := k.GetPool(ctx, pool.PoolId)
	suite.Require().NotNil(initialPool)
	initialReserveA := initialPool.ReserveA
	initialReserveB := initialPool.ReserveB
	initialSwapCount := initialPool.SwapCount

	// Attempt swap with impossible minAmountOut
	trader := testdata.GenTestAddr().String()
	coinIn := sdk.NewCoin("uaura", sdkmath.NewInt(1000))
	impossibleMinOut := sdkmath.NewInt(1000000) // Way more than possible

	// Execute swap - should fail
	_, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, impossibleMinOut, 0)
	suite.Require().Error(err, "swap should fail due to slippage")
	suite.Require().ErrorIs(err, types.ErrSlippageTooHigh)

	// Verify state was NOT changed
	afterFailPool := k.GetPool(ctx, pool.PoolId)
	suite.Require().NotNil(afterFailPool)
	suite.Require().Equal(initialReserveA, afterFailPool.ReserveA, "reserves should not change on slippage failure")
	suite.Require().Equal(initialReserveB, afterFailPool.ReserveB, "reserves should not change on slippage failure")
	suite.Require().Equal(initialSwapCount, afterFailPool.SwapCount, "swap count should not change on slippage failure")
}

// TestMinAmountOutEnforcement tests that minAmountOut is strictly enforced
func (suite *SlippageProtectionTestSuite) TestMinAmountOutEnforcement() {
	ctx := suite.ctx
	k := suite.keeper

	// Create pool with 1:1 ratio
	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(
		ctx,
		creator,
		"uaura",
		"uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(1000000)),
	)
	suite.Require().NoError(err)

	trader := testdata.GenTestAddr().String()
	coinIn := sdk.NewCoin("uaura", sdkmath.NewInt(1000))

	// First, get a quote to see what output we'd actually get
	estimatedOutput, _, _, _, err := k.GetQuote(ctx, pool.PoolId, "uaura", sdkmath.NewInt(1000))
	suite.Require().NoError(err)
	suite.Require().True(estimatedOutput.IsPositive())

	// Test 1: Set minAmountOut exactly equal to estimated - should succeed
	amountOut, _, _, err := k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, estimatedOutput, 0)
	suite.Require().NoError(err)
	suite.Require().True(amountOut.GTE(estimatedOutput), "output should be >= minAmountOut")

	// Test 2: Set minAmountOut slightly higher than possible - should fail
	tooHighMin := estimatedOutput.Add(sdkmath.NewInt(100))
	_, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, tooHighMin, 0)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrSlippageTooHigh)

	// Test 3: Set minAmountOut to 1 - should succeed (very loose slippage)
	looseMin := sdkmath.NewInt(1)
	amountOut, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, looseMin, 0)
	suite.Require().NoError(err)
	suite.Require().True(amountOut.GTE(looseMin))
}

// TestMaxSlippageBpsEnforcement tests price impact slippage limits
func (suite *SlippageProtectionTestSuite) TestMaxSlippageBpsEnforcement() {
	ctx := suite.ctx
	k := suite.keeper

	// Create small pool to make price impact significant
	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(
		ctx,
		creator,
		"uaura",
		"uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(100000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(100000)),
	)
	suite.Require().NoError(err)

	trader := testdata.GenTestAddr().String()

	// Large swap relative to pool size (5% of pool for lower price impact)
	largeSwap := sdk.NewCoin("uaura", sdkmath.NewInt(5000))

	// First get quote to see price impact
	_, _, priceImpact, _, err := k.GetQuote(ctx, pool.PoolId, "uaura", largeSwap.Amount)
	suite.Require().NoError(err)
	suite.Require().True(priceImpact.IsPositive(), "large swap should have price impact")

	// Test 1: Allow reasonable slippage that covers the actual price impact
	// Use 20% = 2000 bps to ensure it covers the actual impact (which will be <10% for 5% pool trade)
	reasonableSlippageBps := uint64(2000) // 20%
	amountOut, _, _, err := k.SwapExactIn(ctx, trader, pool.PoolId, largeSwap, sdkmath.NewInt(1), reasonableSlippageBps)
	suite.Require().NoError(err)
	suite.Require().True(amountOut.GT(sdkmath.ZeroInt()))

	// Test 2: Set very tight slippage limit (should fail)
	tightSlippageBps := uint64(10) // 0.1%
	_, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, largeSwap, sdkmath.NewInt(1), tightSlippageBps)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrSlippageTooHigh)

	// Test 3: maxSlippageBps = 0 means no price impact limit
	// Get fresh quote after first swap changed the pool
	freshOutput, _, _, _, err := k.GetQuote(ctx, pool.PoolId, "uaura", largeSwap.Amount)
	suite.Require().NoError(err)
	// Use slightly less than the fresh quote as minimum
	minOut := freshOutput.MulRaw(99).QuoRaw(100) // 99% of estimated
	amountOut, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, largeSwap, minOut, 0)
	suite.Require().NoError(err)
	suite.Require().True(amountOut.IsPositive())
	suite.Require().True(amountOut.GTE(minOut))
}

// TestSlippageProtectionAgainstSandwichAttack simulates a sandwich attack scenario
func (suite *SlippageProtectionTestSuite) TestSlippageProtectionAgainstSandwichAttack() {
	ctx := suite.ctx
	k := suite.keeper

	// Create pool
	creator := testdata.GenTestAddr().String()
	pool, _, err := k.CreatePool(
		ctx,
		creator,
		"uaura",
		"uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(1000000)),
	)
	suite.Require().NoError(err)

	victim := testdata.GenTestAddr().String()
	attacker := testdata.GenTestAddr().String()

	// Victim wants to swap 10000 uaura
	victimSwapAmount := sdkmath.NewInt(10000)

	// Victim gets quote BEFORE attack
	expectedOutput, _, _, _, err := k.GetQuote(ctx, pool.PoolId, "uaura", victimSwapAmount)
	suite.Require().NoError(err)

	// Victim sets reasonable slippage tolerance (1% = 100 bps)
	minAcceptableOutput := expectedOutput.MulRaw(99).QuoRaw(100) // 99% of expected

	// ATTACK PHASE 1: Attacker front-runs with large buy
	attackerBuyAmount := sdkmath.NewInt(100000) // 10x victim's trade
	attackerCoinIn := sdk.NewCoin("uaura", attackerBuyAmount)
	_, _, _, err = k.SwapExactIn(ctx, attacker, pool.PoolId, attackerCoinIn, sdkmath.NewInt(1), 0)
	suite.Require().NoError(err)

	// Now pool reserves are imbalanced - price moved against victim

	// VICTIM'S TRANSACTION: Should fail due to slippage protection
	victimCoinIn := sdk.NewCoin("uaura", victimSwapAmount)
	_, _, _, err = k.SwapExactIn(ctx, victim, pool.PoolId, victimCoinIn, minAcceptableOutput, 100)

	// This SHOULD fail because the attacker's front-run moved the price
	suite.Require().Error(err, "victim's transaction should fail due to slippage protection")
	suite.Require().ErrorIs(err, types.ErrSlippageTooHigh)

	// This demonstrates that slippage protection successfully prevented
	// the victim from getting sandwiched
}

// TestSlippageProtectionWithMultipleSwaps tests slippage across multiple swaps
func (suite *SlippageProtectionTestSuite) TestSlippageProtectionWithMultipleSwaps() {
	ctx := suite.ctx
	k := suite.keeper

	// Create pool
	creator := testdata.GenTestAddr().String()
	pool, _, err := k.CreatePool(
		ctx,
		creator,
		"uaura",
		"uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(1000000)),
	)
	suite.Require().NoError(err)

	trader := testdata.GenTestAddr().String()
	swapAmount := sdkmath.NewInt(1000)

	// Execute multiple swaps, each should enforce slippage
	for i := 0; i < 5; i++ {
		// Get quote for current pool state
		expectedOutput, _, _, _, err := k.GetQuote(ctx, pool.PoolId, "uaura", swapAmount)
		suite.Require().NoError(err)

		// Set minAmountOut to 95% of expected (5% slippage tolerance)
		minOutput := expectedOutput.MulRaw(95).QuoRaw(100)

		// Execute swap
		actualOutput, _, _, err := k.SwapExactIn(
			ctx,
			trader,
			pool.PoolId,
			sdk.NewCoin("uaura", swapAmount),
			minOutput,
			500, // 5% max slippage
		)
		suite.Require().NoError(err)
		suite.Require().True(actualOutput.GTE(minOutput), "actual output should meet minimum")

		// Verify output is close to quote (within expected fee range)
		// Quote doesn't account for fees, so actual should be slightly less
		suite.Require().True(actualOutput.LTE(expectedOutput), "actual should be <= quote due to fees")
	}
}

// TestSlippageProtectionEdgeCases tests edge cases in slippage protection
func (suite *SlippageProtectionTestSuite) TestSlippageProtectionEdgeCases() {
	ctx := suite.ctx
	k := suite.keeper

	// Create pool
	creator := testdata.GenTestAddr().String()
	pool, _, err := k.CreatePool(
		ctx,
		creator,
		"uaura",
		"uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(1000000)),
	)
	suite.Require().NoError(err)

	trader := testdata.GenTestAddr().String()

	// Edge Case 1: Very small swap (dust)
	tinySwap := sdk.NewCoin("uaura", sdkmath.NewInt(1))
	expectedOutput, _, _, _, err := k.GetQuote(ctx, pool.PoolId, "uaura", tinySwap.Amount)
	suite.Require().NoError(err)

	if expectedOutput.IsPositive() {
		// If output is positive, slippage should be enforced
		_, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, tinySwap, expectedOutput.AddRaw(1), 0)
		suite.Require().Error(err, "should fail when minAmountOut > possible output")
	}

	// Edge Case 2: minAmountOut exactly equals output (boundary condition)
	normalSwap := sdk.NewCoin("uaura", sdkmath.NewInt(1000))
	expectedOutput, _, _, _, err = k.GetQuote(ctx, pool.PoolId, "uaura", normalSwap.Amount)
	suite.Require().NoError(err)

	// This should succeed (output should be >= min)
	actualOutput, _, _, err := k.SwapExactIn(ctx, trader, pool.PoolId, normalSwap, expectedOutput, 0)
	suite.Require().NoError(err)
	suite.Require().True(actualOutput.GTE(expectedOutput))

	// Edge Case 3: minAmountOut = 1 (minimal slippage protection)
	minimalMin := sdkmath.NewInt(1)
	actualOutput, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, normalSwap, minimalMin, 0)
	suite.Require().NoError(err)
	suite.Require().True(actualOutput.GTE(minimalMin))
}

// TestSlippageProtectionBidirectional tests slippage in both swap directions
func (suite *SlippageProtectionTestSuite) TestSlippageProtectionBidirectional() {
	ctx := suite.ctx
	k := suite.keeper

	// Create pool
	creator := testdata.GenTestAddr().String()
	pool, _, err := k.CreatePool(
		ctx,
		creator,
		"uaura",
		"uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(1000000)),
	)
	suite.Require().NoError(err)

	trader := testdata.GenTestAddr().String()
	swapAmount := sdkmath.NewInt(1000)

	// Test swap A -> B
	{
		expectedOutput, _, _, _, err := k.GetQuote(ctx, pool.PoolId, "uaura", swapAmount)
		suite.Require().NoError(err)

		minOutput := expectedOutput.MulRaw(95).QuoRaw(100)
		actualOutput, _, _, err := k.SwapExactIn(
			ctx,
			trader,
			pool.PoolId,
			sdk.NewCoin("uaura", swapAmount),
			minOutput,
			0,
		)
		suite.Require().NoError(err)
		suite.Require().True(actualOutput.GTE(minOutput))

		// Test rejection with too high minAmountOut
		tooHighMin := expectedOutput.MulRaw(105).QuoRaw(100)
		_, _, _, err = k.SwapExactIn(
			ctx,
			trader,
			pool.PoolId,
			sdk.NewCoin("uaura", swapAmount),
			tooHighMin,
			0,
		)
		suite.Require().Error(err)
		suite.Require().ErrorIs(err, types.ErrSlippageTooHigh)
	}

	// Test swap B -> A (reverse direction)
	{
		expectedOutput, _, _, _, err := k.GetQuote(ctx, pool.PoolId, "uusdt", swapAmount)
		suite.Require().NoError(err)

		minOutput := expectedOutput.MulRaw(95).QuoRaw(100)
		actualOutput, _, _, err := k.SwapExactIn(
			ctx,
			trader,
			pool.PoolId,
			sdk.NewCoin("uusdt", swapAmount),
			minOutput,
			0,
		)
		suite.Require().NoError(err)
		suite.Require().True(actualOutput.GTE(minOutput))

		// Test rejection
		tooHighMin := expectedOutput.MulRaw(105).QuoRaw(100)
		_, _, _, err = k.SwapExactIn(
			ctx,
			trader,
			pool.PoolId,
			sdk.NewCoin("uusdt", swapAmount),
			tooHighMin,
			0,
		)
		suite.Require().Error(err)
		suite.Require().ErrorIs(err, types.ErrSlippageTooHigh)
	}
}

// TestSlippageRejectionErrorMessages tests that error messages are clear and informative
func (suite *SlippageProtectionTestSuite) TestSlippageRejectionErrorMessages() {
	ctx := suite.ctx
	k := suite.keeper

	// Create pool
	creator := testdata.GenTestAddr().String()
	pool, _, err := k.CreatePool(
		ctx,
		creator,
		"uaura",
		"uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(1000000)),
	)
	suite.Require().NoError(err)

	trader := testdata.GenTestAddr().String()
	coinIn := sdk.NewCoin("uaura", sdkmath.NewInt(1000))
	impossibleMin := sdkmath.NewInt(1000000)

	// Execute swap that will fail
	_, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, impossibleMin, 0)
	suite.Require().Error(err)

	// Verify error is specific and includes useful information
	suite.Require().Contains(err.Error(), "output amount")
	suite.Require().Contains(err.Error(), "less than minimum")
	suite.Require().ErrorIs(err, types.ErrSlippageTooHigh)
}
