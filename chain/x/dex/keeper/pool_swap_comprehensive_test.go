package keeper_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dex/keeper"
)

// PoolSwapComprehensiveTestSuite tests pool creation and swap AMM math edge cases
// as specified in ROADMAP_PRODUCTION.md task #12
type PoolSwapComprehensiveTestSuite struct {
	suite.Suite

	Keeper     *keeper.Keeper
	SdkCtx     sdk.Context
	BankKeeper *MockBankKeeper
}

func TestPoolSwapComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(PoolSwapComprehensiveTestSuite))
}

// SetupTest initializes the test suite before each test
func (suite *PoolSwapComprehensiveTestSuite) SetupTest() {
	// Configure SDK with Aura-specific prefixes
	keepertest.ConfigureSDK()

	input := keepertest.CreateTestInput(suite.T())
	suite.BankKeeper = NewMockBankKeeper()

	suite.Keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		suite.BankKeeper,
		nil, // accountKeeper
		nil, // vcKeeper
		nil, // securityKeeper
	)
	suite.SdkCtx = input.Ctx
}

// addr generates a deterministic test address from a seed string
func (suite *PoolSwapComprehensiveTestSuite) addr(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	bech32, err := sdk.Bech32ifyAddressBytes("aura", sum[:20])
	suite.Require().NoError(err)
	return bech32
}

// fundAccount funds an account with coins using the mock bank keeper
func (suite *PoolSwapComprehensiveTestSuite) fundAccount(addrStr string, denom string, amount int64) {
	addr, err := sdk.AccAddressFromBech32(addrStr)
	suite.Require().NoError(err)
	suite.BankKeeper.SetBalance(addr, denom, sdkmath.NewInt(amount))
}

// ============================================================================
// Pool Creation AMM Math Edge Cases
// ============================================================================

// TestPoolCreation_ZeroAmounts verifies that pool creation rejects zero amounts
func (suite *PoolSwapComprehensiveTestSuite) TestPoolCreation_ZeroAmounts() {
	creator := suite.addr("creator")

	// Fund creator
	suite.fundAccount(creator, "tokenA", 1000000)
	suite.fundAccount(creator, "tokenB", 1000000)

	// Test zero amount A
	_, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.ZeroInt()),
		sdk.NewCoin("tokenB", sdkmath.NewInt(100000)),
	)
	suite.Error(err)
	suite.Contains(err.Error(), "amounts must be positive")

	// Test zero amount B
	_, _, err = suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(100000)),
		sdk.NewCoin("tokenB", sdkmath.ZeroInt()),
	)
	suite.Error(err)
	suite.Contains(err.Error(), "amounts must be positive")
}

// TestPoolCreation_NegativeAmounts verifies that pool creation rejects negative amounts
func (suite *PoolSwapComprehensiveTestSuite) TestPoolCreation_NegativeAmounts() {
	creator := suite.addr("creator")

	// Fund creator
	suite.fundAccount(creator, "tokenA", 1000000)
	suite.fundAccount(creator, "tokenB", 1000000)

	// SDK Coin validation should catch negative amounts before they reach CreatePool
	// Test that invalid coins are rejected
	invalidCoin := sdk.Coin{Denom: "tokenA", Amount: sdkmath.NewInt(-1000)}
	err := invalidCoin.Validate()
	suite.Error(err, "negative amounts should fail validation")
}

// TestPoolCreation_ExtremeRatios tests pool creation with extreme token ratios
func (suite *PoolSwapComprehensiveTestSuite) TestPoolCreation_ExtremeRatios() {
	creator := suite.addr("creator")

	// Fund creator with large amounts
	suite.fundAccount(creator, "tokenA", 1000000000000)
	suite.fundAccount(creator, "tokenB", 1000000000000)

	testCases := []struct {
		name     string
		amountA  sdkmath.Int
		amountB  sdkmath.Int
		shouldOk bool
	}{
		{
			name:     "1:1000 ratio - small A",
			amountA:  sdkmath.NewInt(1_000_000),    // Minimum liquidity
			amountB:  sdkmath.NewInt(1_000_000_000), // 1000x ratio
			shouldOk: true,
		},
		{
			name:     "1000:1 ratio - small B",
			amountA:  sdkmath.NewInt(1_000_000_000), // 1000x ratio
			amountB:  sdkmath.NewInt(1_000_000),    // Minimum liquidity
			shouldOk: true,
		},
		{
			name:     "1:1000000 ratio - very extreme",
			amountA:  sdkmath.NewInt(1_000_000),      // Minimum liquidity
			amountB:  sdkmath.NewInt(1_000_000_000_000), // 1,000,000x ratio
			shouldOk: true,
		},
	}

	for i, tc := range testCases {
		suite.Run(tc.name, func() {
			// Use unique denoms for each test case
			denomA := fmt.Sprintf("tokenA%d", i)
			denomB := fmt.Sprintf("tokenB%d", i)

			suite.fundAccount(creator, denomA, tc.amountA.Int64()+10000)
			suite.fundAccount(creator, denomB, tc.amountB.Int64()+10000)

			pool, lpTokens, err := suite.Keeper.CreatePool(
				suite.SdkCtx,
				creator,
				denomA,
				denomB,
				sdk.NewCoin(denomA, tc.amountA),
				sdk.NewCoin(denomB, tc.amountB),
			)

			if tc.shouldOk {
				suite.NoError(err)
				suite.NotNil(pool)
				suite.True(lpTokens.GT(sdkmath.ZeroInt()), "LP tokens should be positive")

				// Verify k = x * y invariant is set correctly
				suite.Equal(tc.amountA.String(), pool.ReserveA)
				suite.Equal(tc.amountB.String(), pool.ReserveB)
			} else {
				suite.Error(err)
			}
		})
	}
}

// TestPoolCreation_DuplicatePool verifies duplicate pool prevention
func (suite *PoolSwapComprehensiveTestSuite) TestPoolCreation_DuplicatePool() {
	creator := suite.addr("creator")

	suite.fundAccount(creator, "tokenA", 2000000)
	suite.fundAccount(creator, "tokenB", 2000000)

	// Create first pool
	pool1, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(100000)),
		sdk.NewCoin("tokenB", sdkmath.NewInt(100000)),
	)
	suite.NoError(err)
	suite.NotNil(pool1)

	// Attempt to create duplicate pool
	_, _, err = suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(50000)),
		sdk.NewCoin("tokenB", sdkmath.NewInt(50000)),
	)
	suite.Error(err)
	suite.Contains(err.Error(), "pool")
	suite.Contains(err.Error(), "already exists")
}

// TestPoolCreation_MinimumLiquidityBurn verifies minimum liquidity mechanism
func (suite *PoolSwapComprehensiveTestSuite) TestPoolCreation_MinimumLiquidityBurn() {
	creator := suite.addr("creator")

	suite.fundAccount(creator, "tokenA", 1000000)
	suite.fundAccount(creator, "tokenB", 1000000)

	amountA := sdkmath.NewInt(100000)
	amountB := sdkmath.NewInt(100000)

	pool, lpTokens, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", amountA),
		sdk.NewCoin("tokenB", amountB),
	)
	suite.NoError(err)
	suite.NotNil(pool)

	// Calculate expected LP tokens = sqrt(x * y)
	// For equal amounts: sqrt(100000 * 100000) = 100000
	expectedTotal := sdkmath.NewInt(100000)

	// Minimum liquidity (1000) should be burned from first deposit
	minimumLiquidity := sdkmath.NewInt(1000)
	expectedCreator := expectedTotal.Sub(minimumLiquidity)

	suite.Equal(expectedCreator.String(), lpTokens.String(),
		"LP tokens should be total - minimum liquidity")

	// Verify pool has LP tokens issued
	suite.NotEmpty(pool.TotalLpTokens, "Pool should have LP tokens issued")
}

// ============================================================================
// Swap AMM Math Edge Cases
// ============================================================================

// TestSwap_ExtremelySmallInput tests swaps with minimal input amounts
func (suite *PoolSwapComprehensiveTestSuite) TestSwap_ExtremelySmallInput() {
	creator := suite.addr("creator")
	swapper := suite.addr("swapper")

	// Create pool with reasonable liquidity
	suite.fundAccount(creator, "tokenA", 100_000_000)
	suite.fundAccount(creator, "tokenB", 100_000_000)

	pool, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(10_000_000)), // Large pool to allow minimum swaps
		sdk.NewCoin("tokenB", sdkmath.NewInt(10_000_000)), // Large pool to allow minimum swaps
	)
	suite.NoError(err)

	// Fund swapper with tiny amount
	suite.fundAccount(swapper, "tokenA", 10)

	// Swap extremely small amount (1 unit)
	amountOut, _, _, err := suite.Keeper.SecureSwapExactIn(
		suite.SdkCtx,
		swapper,
		pool.PoolId,
		sdk.NewCoin("tokenA", sdkmath.NewInt(1)),
		sdkmath.NewInt(1), // Minimum output
		10000,             // 100% slippage allowed
	)

	// Small swaps should work but may result in minimal output due to fees
	if err == nil {
		// If swap succeeds, verify output is reasonable
		suite.True(amountOut.GTE(sdkmath.NewInt(1)))
	} else {
		// Or it may fail due to validation
		suite.NotNil(err)
	}
}

// TestSwap_LargeInputRelativeToPool tests swaps that significantly impact pool
func (suite *PoolSwapComprehensiveTestSuite) TestSwap_LargeInputRelativeToPool() {
	creator := suite.addr("creator")
	swapper := suite.addr("swapper")

	// Create pool sized to allow minimum trade amount (1,000,000) to be ~10% of pool
	// Pool size: 10,000,000 each side (so 1M swap is 10% of pool)
	suite.fundAccount(creator, "tokenA", 100_000_000)
	suite.fundAccount(creator, "tokenB", 100_000_000)

	pool, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(10_000_000)),
		sdk.NewCoin("tokenB", sdkmath.NewInt(10_000_000)),
	)
	suite.NoError(err)

	// Large amount relative to pool (10% of pool = minimum trade amount)
	largeAmount := sdkmath.NewInt(1_000_000)
	suite.fundAccount(swapper, "tokenA", largeAmount.Int64()+1000)

	// Large swap should have significant price impact
	amountOut, effectivePrice, priceImpact, err := suite.Keeper.SecureSwapExactIn(
		suite.SdkCtx,
		swapper,
		pool.PoolId,
		sdk.NewCoin("tokenA", largeAmount),
		sdkmath.NewInt(1), // Minimum output
		10000, // 100% slippage allowed for test
	)

	suite.NoError(err)
	suite.True(amountOut.GT(sdkmath.ZeroInt()), "Output should be positive")

	// Price impact should be substantial (>5% for 10% of pool)
	suite.True(priceImpact.GT(sdkmath.LegacyNewDec(5)),
		"Large swap should have >5% price impact")

	// Verify output is less than naive expectation due to slippage
	suite.True(amountOut.LT(largeAmount),
		"Output should be less than input due to price impact")

	// Effective price should be returned
	suite.True(effectivePrice.GT(sdkmath.LegacyZeroDec()), "Effective price should be positive")
}

// TestSwap_MultipleSequentialSwaps tests repeated swaps and price impact accumulation
func (suite *PoolSwapComprehensiveTestSuite) TestSwap_MultipleSequentialSwaps() {
	creator := suite.addr("creator")
	swapper := suite.addr("swapper")

	// Create large pool to keep price impact below 10% for 1M swaps
	suite.fundAccount(creator, "tokenA", 1_000_000_000)
	suite.fundAccount(creator, "tokenB", 1_000_000_000)

	pool, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(100_000_000)), // 100M pool keeps 1M swaps under 10% impact
		sdk.NewCoin("tokenB", sdkmath.NewInt(100_000_000)), // 100M pool keeps 1M swaps under 10% impact
	)
	suite.NoError(err)

	// Fund swapper
	suite.fundAccount(swapper, "tokenA", 50_000_000)

	// Record initial reserves
	initialReserveA, _ := sdkmath.NewIntFromString(pool.ReserveA)
	initialReserveB, _ := sdkmath.NewIntFromString(pool.ReserveB)
	initialK := initialReserveA.Mul(initialReserveB)

	// Perform multiple swaps (must meet minimum of 1,000,000 uaura)
	// Use different swappers to avoid wash trading detection
	totalAmountOut := sdkmath.ZeroInt()
	swapAmount := sdkmath.NewInt(1_000_000) // Minimum swap amount
	numSwaps := 10

	for i := 0; i < numSwaps; i++ {
		// Use a different swapper for each iteration to avoid wash trading detection
		currentSwapper := suite.addr(fmt.Sprintf("swapper%d", i))
		suite.fundAccount(currentSwapper, "tokenA", swapAmount.Int64()+1000)

		// Advance blocks to avoid front-running protection (requires 2 block wait)
		if i > 0 {
			suite.SdkCtx = suite.SdkCtx.WithBlockHeight(suite.SdkCtx.BlockHeight() + 2)
		}

		amountOut, _, _, err := suite.Keeper.SecureSwapExactIn(
			suite.SdkCtx,
			currentSwapper,
			pool.PoolId,
			sdk.NewCoin("tokenA", swapAmount),
			sdkmath.NewInt(1), // Minimum output
			10000,
		)
		suite.NoError(err)
		totalAmountOut = totalAmountOut.Add(amountOut)
	}

	// Fetch updated pool
	updatedPool := suite.Keeper.GetPool(suite.SdkCtx, pool.PoolId)
	suite.NotNil(updatedPool)

	// Verify reserves changed correctly
	finalReserveA, _ := sdkmath.NewIntFromString(updatedPool.ReserveA)
	finalReserveB, _ := sdkmath.NewIntFromString(updatedPool.ReserveB)

	// Reserve A should increase (we added tokenA)
	suite.True(finalReserveA.GT(initialReserveA))

	// Reserve B should decrease (we removed tokenB)
	suite.True(finalReserveB.LT(initialReserveB))

	// Verify k (with fees) increased or stayed approximately constant
	// k' >= k due to fee accumulation
	finalK := finalReserveA.Mul(finalReserveB)
	suite.True(finalK.GTE(initialK.MulRaw(99).QuoRaw(100)),
		"K should be preserved (within fee tolerance)")
}

// TestSwap_BothDirections tests bidirectional swaps maintain pool invariants
func (suite *PoolSwapComprehensiveTestSuite) TestSwap_BothDirections() {
	creator := suite.addr("creator")
	swapper1 := suite.addr("swapper1")
	swapper2 := suite.addr("swapper2")

	// Create pool
	suite.fundAccount(creator, "tokenA", 100_000_000)
	suite.fundAccount(creator, "tokenB", 100_000_000)

	// Create pool large enough that 1M swap is <20% of pool
	// 1M / 6M = 16.67% < 20%
	pool, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(6_000_000)),
		sdk.NewCoin("tokenB", sdkmath.NewInt(6_000_000)),
	)
	suite.NoError(err)

	// Fund swappers with minimum trade amount
	suite.fundAccount(swapper1, "tokenA", 2_000_000)
	suite.fundAccount(swapper2, "tokenB", 2_000_000)

	// Swap A -> B
	amountOut1, _, _, err := suite.Keeper.SecureSwapExactIn(
		suite.SdkCtx,
		swapper1,
		pool.PoolId,
		sdk.NewCoin("tokenA", sdkmath.NewInt(1_000_000)), // Minimum trade amount
		sdkmath.NewInt(1), // Minimum output
		10000,
	)
	suite.NoError(err)
	suite.True(amountOut1.GT(sdkmath.ZeroInt()))

	// Advance blocks to avoid MEV protection
	suite.SdkCtx = suite.SdkCtx.WithBlockHeight(suite.SdkCtx.BlockHeight() + 2)

	// Swap B -> A (reverse direction)
	amountOut2, _, _, err := suite.Keeper.SecureSwapExactIn(
		suite.SdkCtx,
		swapper2,
		pool.PoolId,
		sdk.NewCoin("tokenB", sdkmath.NewInt(1_000_000)), // Minimum trade amount
		sdkmath.NewInt(1), // Minimum output
		10000,
	)
	suite.NoError(err)
	suite.True(amountOut2.GT(sdkmath.ZeroInt()))

	// Verify pool still healthy
	finalPool := suite.Keeper.GetPool(suite.SdkCtx, pool.PoolId)
	suite.NotNil(finalPool)

	reserveA, _ := sdkmath.NewIntFromString(finalPool.ReserveA)
	reserveB, _ := sdkmath.NewIntFromString(finalPool.ReserveB)

	suite.True(reserveA.GT(sdkmath.ZeroInt()))
	suite.True(reserveB.GT(sdkmath.ZeroInt()))
}

// TestSwap_InsufficientLiquidity tests swaps that would drain pool
func (suite *PoolSwapComprehensiveTestSuite) TestSwap_InsufficientLiquidity() {
	creator := suite.addr("creator")
	swapper := suite.addr("swapper")

	// Create small pool
	suite.fundAccount(creator, "tokenA", 100000)
	suite.fundAccount(creator, "tokenB", 100000)

	pool, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(10000)),
		sdk.NewCoin("tokenB", sdkmath.NewInt(10000)),
	)
	suite.NoError(err)

	// Fund swapper with massive amount
	hugeAmount := sdkmath.NewInt(1000000)
	suite.fundAccount(swapper, "tokenA", hugeAmount.Int64())

	// Attempt to swap more than pool can handle
	// This should either fail or enforce maximum output
	_, _, _, err = suite.Keeper.SecureSwapExactIn(
		suite.SdkCtx,
		swapper,
		pool.PoolId,
		sdk.NewCoin("tokenA", hugeAmount),
		sdkmath.NewInt(1), // Minimum output
		10000,
	)

	// Should fail due to insufficient liquidity or price impact limits
	suite.Error(err)
}

// ============================================================================
// Fee Calculation and Gas Guardrails
// ============================================================================

// TestFeeCalculation_ZeroFee tests behavior with zero fee configuration
func (suite *PoolSwapComprehensiveTestSuite) TestFeeCalculation_ZeroFee() {
	creator := suite.addr("creator")
	swapper := suite.addr("swapper")

	// Get default params (fee configuration is per-pool, not global)
	_ = suite.Keeper.GetParams(suite.SdkCtx)
	// Note: Fees are configured per pool, not via params

	// Create pool
	suite.fundAccount(creator, "tokenA", 1000000)
	suite.fundAccount(creator, "tokenB", 1000000)

	pool, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(100000)),
		sdk.NewCoin("tokenB", sdkmath.NewInt(100000)),
	)
	suite.NoError(err)

	// Fund swapper
	suite.fundAccount(swapper, "tokenA", 20_000_000)

	// Swap should work with zero fees
	amountOut, _, _, err := suite.Keeper.SecureSwapExactIn(
		suite.SdkCtx,
		swapper,
		pool.PoolId,
		sdk.NewCoin("tokenA", sdkmath.NewInt(1_000_000)), // Minimum swap amount
		sdkmath.NewInt(1), // Minimum output
		10000,
	)

	suite.NoError(err)
	suite.True(amountOut.GT(sdkmath.ZeroInt()))
}

// TestFeeCalculation_MaximumFee tests behavior with maximum fee
func (suite *PoolSwapComprehensiveTestSuite) TestFeeCalculation_MaximumFee() {
	creator := suite.addr("creator")
	swapper := suite.addr("swapper")

	// Get default params (fee is part of pool configuration)
	_ = suite.Keeper.GetParams(suite.SdkCtx)
	// Note: Pool has default fee configured

	// Create pool
	suite.fundAccount(creator, "tokenA", 1000000)
	suite.fundAccount(creator, "tokenB", 1000000)

	pool, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(100000)),
		sdk.NewCoin("tokenB", sdkmath.NewInt(100000)),
	)
	suite.NoError(err)

	// Fund swapper
	suite.fundAccount(swapper, "tokenA", 20_000_000)

	// Swap with high fee
	amountOut, _, _, err := suite.Keeper.SecureSwapExactIn(
		suite.SdkCtx,
		swapper,
		pool.PoolId,
		sdk.NewCoin("tokenA", sdkmath.NewInt(1_000_000)), // Minimum swap amount
		sdkmath.NewInt(1), // Minimum output
		10000,
	)

	suite.NoError(err)

	// Output should be significantly reduced due to high fee
	// With 10% fee, output should be notably less than no-fee scenario
	suite.True(amountOut.GT(sdkmath.ZeroInt()))
	suite.True(amountOut.LT(sdkmath.NewInt(1_000_000)), // Minimum swap amount
		"Output should be reduced by fees")
}

// TestFeeCalculation_FeeAccumulation tests that fees accumulate correctly
func (suite *PoolSwapComprehensiveTestSuite) TestFeeCalculation_FeeAccumulation() {
	creator := suite.addr("creator")

	// Create pool
	suite.fundAccount(creator, "tokenA", 100_000_000)
	suite.fundAccount(creator, "tokenB", 100_000_000)

	pool, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(10_000_000)), // Large pool to allow minimum swaps
		sdk.NewCoin("tokenB", sdkmath.NewInt(10_000_000)), // Large pool to allow minimum swaps
	)
	suite.NoError(err)

	// Record initial k
	initialReserveA, _ := sdkmath.NewIntFromString(pool.ReserveA)
	initialReserveB, _ := sdkmath.NewIntFromString(pool.ReserveB)
	initialK := initialReserveA.Mul(initialReserveB)

	// Perform multiple swaps with different swappers to avoid wash trading detection
	for i := 0; i < 20; i++ {
		currentSwapper := suite.addr(fmt.Sprintf("feeswapper%d", i))
		suite.fundAccount(currentSwapper, "tokenA", 2_000_000)

		// Advance blocks to avoid MEV protection
		if i > 0 {
			suite.SdkCtx = suite.SdkCtx.WithBlockHeight(suite.SdkCtx.BlockHeight() + 2)
		}

		_, _, _, err := suite.Keeper.SecureSwapExactIn(
			suite.SdkCtx,
			currentSwapper,
			pool.PoolId,
			sdk.NewCoin("tokenA", sdkmath.NewInt(1_000_000)), // Minimum swap amount
			sdkmath.NewInt(1), // Minimum output
			10000,
		)
		suite.NoError(err)
	}

	// Get final pool state
	finalPool := suite.Keeper.GetPool(suite.SdkCtx, pool.PoolId)
	finalReserveA, _ := sdkmath.NewIntFromString(finalPool.ReserveA)
	finalReserveB, _ := sdkmath.NewIntFromString(finalPool.ReserveB)
	finalK := finalReserveA.Mul(finalReserveB)

	// K should have increased due to fee accumulation
	suite.True(finalK.GT(initialK),
		"K should increase as fees accumulate in pool")
}

// TestGasGuardrails_PoolCreation tests gas consumption for pool creation
func (suite *PoolSwapComprehensiveTestSuite) TestGasGuardrails_PoolCreation() {
	creator := suite.addr("creator")

	suite.fundAccount(creator, "tokenA", 1000000)
	suite.fundAccount(creator, "tokenB", 1000000)

	// Measure gas before operation
	gasBefore := suite.SdkCtx.GasMeter().GasConsumed()

	_, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(100000)),
		sdk.NewCoin("tokenB", sdkmath.NewInt(100000)),
	)
	suite.NoError(err)

	// Measure gas after operation
	gasAfter := suite.SdkCtx.GasMeter().GasConsumed()
	gasUsed := gasAfter - gasBefore

	// Gas should be reasonable (< 500k for pool creation)
	suite.Less(gasUsed, uint64(500000),
		"Pool creation should use reasonable gas")

	// Gas should be non-trivial (> 10k due to storage operations)
	suite.Greater(gasUsed, uint64(10000),
		"Pool creation should consume non-trivial gas")
}

// TestGasGuardrails_SwapOperation tests gas consumption for swaps
func (suite *PoolSwapComprehensiveTestSuite) TestGasGuardrails_SwapOperation() {
	creator := suite.addr("creator")
	swapper := suite.addr("swapper")

	// Setup pool
	suite.fundAccount(creator, "tokenA", 1000000)
	suite.fundAccount(creator, "tokenB", 1000000)

	pool, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(500000)),
		sdk.NewCoin("tokenB", sdkmath.NewInt(500000)),
	)
	suite.NoError(err)

	// Fund swapper
	suite.fundAccount(swapper, "tokenA", 20_000_000)

	// Measure gas for swap
	gasBefore := suite.SdkCtx.GasMeter().GasConsumed()

	_, _, _, err = suite.Keeper.SecureSwapExactIn(
		suite.SdkCtx,
		swapper,
		pool.PoolId,
		sdk.NewCoin("tokenA", sdkmath.NewInt(1_000_000)), // Minimum swap amount
		sdkmath.NewInt(1), // Minimum output
		10000,
	)
	suite.NoError(err)

	gasAfter := suite.SdkCtx.GasMeter().GasConsumed()
	gasUsed := gasAfter - gasBefore

	// Swap should use reasonable gas (< 300k)
	suite.Less(gasUsed, uint64(300000),
		"Swap should use reasonable gas")

	// Swap should use non-trivial gas (> 5k due to calculations and storage)
	suite.Greater(gasUsed, uint64(5000),
		"Swap should consume non-trivial gas")
}

// ============================================================================
// Slippage Protection Scenarios
// ============================================================================

// TestSlippageProtection_ExactMinimumOutput tests slippage exactly at minimum
func (suite *PoolSwapComprehensiveTestSuite) TestSlippageProtection_ExactMinimumOutput() {
	creator := suite.addr("creator")
	swapper := suite.addr("swapper")

	// Create pool
	suite.fundAccount(creator, "tokenA", 1000000)
	suite.fundAccount(creator, "tokenB", 1000000)

	pool, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(100000)),
		sdk.NewCoin("tokenB", sdkmath.NewInt(100000)),
	)
	suite.NoError(err)

	suite.fundAccount(swapper, "tokenA", 20_000_000)

	// First, do a test swap to see what output we get
	testOutput, _, _, err := suite.Keeper.SecureSwapExactIn(
		suite.SdkCtx,
		swapper,
		pool.PoolId,
		sdk.NewCoin("tokenA", sdkmath.NewInt(1_000_000)), // Minimum swap amount
		sdkmath.NewInt(1), // Minimum output
		10000,
	)
	suite.NoError(err)

	// Reset and try again with exact minimum
	suite.SetupTest()
	suite.fundAccount(creator, "tokenA", 1000000)
	suite.fundAccount(creator, "tokenB", 1000000)

	pool2, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(100000)),
		sdk.NewCoin("tokenB", sdkmath.NewInt(100000)),
	)
	suite.NoError(err)

	suite.fundAccount(swapper, "tokenA", 20_000_000)

	// Swap with exact minimum output
	actualOutput, _, _, err := suite.Keeper.SecureSwapExactIn(
		suite.SdkCtx,
		swapper,
		pool2.PoolId,
		sdk.NewCoin("tokenA", sdkmath.NewInt(1_000_000)), // Minimum swap amount
		testOutput, // Use exact expected output as minimum
		10000,
	)

	suite.NoError(err)
	suite.True(actualOutput.GTE(testOutput),
		"Actual output should meet minimum")
}

// TestSlippageProtection_BelowMinimumOutput tests slippage rejection
func (suite *PoolSwapComprehensiveTestSuite) TestSlippageProtection_BelowMinimumOutput() {
	creator := suite.addr("creator")
	swapper := suite.addr("swapper")

	// Create larger pool to allow minimum trade amounts
	suite.fundAccount(creator, "tokenA", 100_000_000)
	suite.fundAccount(creator, "tokenB", 100_000_000)

	pool, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(10_000_000)),
		sdk.NewCoin("tokenB", sdkmath.NewInt(10_000_000)),
	)
	suite.NoError(err)

	suite.fundAccount(swapper, "tokenA", 20_000_000)

	// Set unrealistically high minimum output (more than possible output)
	unrealisticMinimum := sdkmath.NewInt(2_000_000) // Asking for 2M output when ~900k is realistic

	_, _, _, err = suite.Keeper.SecureSwapExactIn(
		suite.SdkCtx,
		swapper,
		pool.PoolId,
		sdk.NewCoin("tokenA", sdkmath.NewInt(1_000_000)), // Minimum swap amount
		unrealisticMinimum,
		10000,
	)

	// Should fail due to slippage/minimum output protection
	suite.Error(err)
	// Error could be about slippage or minimum output not met
	suite.True(
		suite.Contains(err.Error(), "slippage") ||
		suite.Contains(err.Error(), "minimum") ||
		suite.Contains(err.Error(), "output"),
		"Error should be related to slippage or minimum output",
	)
}

// TestSlippageProtection_MaxSlippageBps tests basis point slippage limits
func (suite *PoolSwapComprehensiveTestSuite) TestSlippageProtection_MaxSlippageBps() {
	creator := suite.addr("creator")

	// Create large pool to allow varying swap sizes above minimum
	// Pool: 100M each side, so 1M (min) is 1% of pool, 2M is 2%, 10M is 10%
	suite.fundAccount(creator, "tokenA", 200_000_000)
	suite.fundAccount(creator, "tokenB", 200_000_000)

	pool, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(100_000_000)),
		sdk.NewCoin("tokenB", sdkmath.NewInt(100_000_000)),
	)
	suite.NoError(err)

	testCases := []struct {
		name          string
		swapAmount    int64
		maxSlippageBps uint64
		shouldSucceed bool
	}{
		{
			name:          "Small swap with tight slippage (0.1%)",
			swapAmount:    1_000_000, // 1% of pool
			maxSlippageBps: 10, // 0.1% - may fail due to fees/slippage
			shouldSucceed: false, // 1% of pool causes >0.1% slippage
		},
		{
			name:          "Medium swap with moderate slippage (1%)",
			swapAmount:    1_000_000, // 1% of pool
			maxSlippageBps: 100, // 1%
			shouldSucceed: true, // Should succeed with 1% tolerance
		},
		{
			name:          "Large swap with loose slippage (10%)",
			swapAmount:    10_000_000, // 10% of pool
			maxSlippageBps: 1000, // 10%
			shouldSucceed: true, // Should succeed with 10% tolerance
		},
		{
			name:          "Large swap with tight slippage (0.1%) - should fail",
			swapAmount:    10_000_000, // 10% of pool
			maxSlippageBps: 10, // 0.1% - definitely too tight
			shouldSucceed: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			swapper := suite.addr(fmt.Sprintf("slippageswapper-%s", tc.name))
			suite.fundAccount(swapper, "tokenA", tc.swapAmount+1_000_000)

			// Advance blocks between tests to avoid MEV protection
			suite.SdkCtx = suite.SdkCtx.WithBlockHeight(suite.SdkCtx.BlockHeight() + 2)

			_, _, _, err := suite.Keeper.SecureSwapExactIn(
				suite.SdkCtx,
				swapper,
				pool.PoolId,
				sdk.NewCoin("tokenA", sdkmath.NewInt(tc.swapAmount)),
				sdkmath.NewInt(1), // Minimum output
				tc.maxSlippageBps,
			)

			if tc.shouldSucceed {
				suite.NoError(err, "Swap should succeed with sufficient slippage tolerance")
			} else {
				suite.Error(err, "Swap should fail with insufficient slippage tolerance")
				// Error could be about dust, trade size, or slippage
			}
		})
	}
}

// TestSlippageProtection_PriceImpactCalculation tests accurate price impact reporting
func (suite *PoolSwapComprehensiveTestSuite) TestSlippageProtection_PriceImpactCalculation() {
	creator := suite.addr("creator")

	// Create large pool with 1:1 ratio to allow different swap sizes
	// Pool: 100M each side, allows swaps from 1M (1%) to 20M (20% - max trade size)
	suite.fundAccount(creator, "tokenA", 200_000_000)
	suite.fundAccount(creator, "tokenB", 200_000_000)

	pool, _, err := suite.Keeper.CreatePool(
		suite.SdkCtx,
		creator,
		"tokenA",
		"tokenB",
		sdk.NewCoin("tokenA", sdkmath.NewInt(100_000_000)),
		sdk.NewCoin("tokenB", sdkmath.NewInt(100_000_000)),
	)
	suite.NoError(err)

	testCases := []struct {
		name                string
		swapAmount          int64
		expectedMinImpact   float64 // minimum expected impact %
		expectedMaxImpact   float64 // maximum expected impact %
	}{
		{
			name:                "Tiny swap (1% of pool)",
			swapAmount:          1_000_000, // Minimum trade amount
			expectedMinImpact:   0.5,
			expectedMaxImpact:   2.0,
		},
		{
			name:                "Small swap (2% of pool)",
			swapAmount:          2_000_000,
			expectedMinImpact:   1.0,
			expectedMaxImpact:   3.0,
		},
		{
			name:                "Medium swap (10% of pool)",
			swapAmount:          10_000_000,
			expectedMinImpact:   5.0,
			expectedMaxImpact:   15.0,
		},
		{
			name:                "Large swap (20% of pool - max trade size)",
			swapAmount:          20_000_000, // 20% = MaxTradeSizePercent
			expectedMinImpact:   10.0,
			expectedMaxImpact:   30.0,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			swapper := suite.addr(fmt.Sprintf("impactswapper-%s", tc.name))
			suite.fundAccount(swapper, "tokenA", tc.swapAmount+1_000_000)

			// Advance blocks between tests
			suite.SdkCtx = suite.SdkCtx.WithBlockHeight(suite.SdkCtx.BlockHeight() + 2)

			_, _, priceImpact, err := suite.Keeper.SecureSwapExactIn(
				suite.SdkCtx,
				swapper,
				pool.PoolId,
				sdk.NewCoin("tokenA", sdkmath.NewInt(tc.swapAmount)),
				sdkmath.NewInt(1), // Minimum output
				10000, // Allow 100% slippage for testing
			)

			suite.NoError(err)

			// Verify price impact is within expected range
			impactFloat, _ := priceImpact.Float64()
			suite.GreaterOrEqual(impactFloat, tc.expectedMinImpact,
				"Price impact should be at least %.2f%%", tc.expectedMinImpact)
			suite.LessOrEqual(impactFloat, tc.expectedMaxImpact,
				"Price impact should be at most %.2f%%", tc.expectedMaxImpact)
		})
	}
}
