package keeper

import (
	"crypto/sha256"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/testing/testutil"
	"github.com/aequitas/aura/chain/x/dex/types"
)

// LPTokenInvariantTestSuite tests LP token supply invariant validation
type LPTokenInvariantTestSuite struct {
	suite.Suite

	Keeper     *Keeper
	SdkCtx     sdk.Context
	BankKeeper *testutil.MockBankKeeper
}

func TestLPTokenInvariantTestSuite(t *testing.T) {
	suite.Run(t, new(LPTokenInvariantTestSuite))
}

// SetupTest initializes the test suite before each test
func (suite *LPTokenInvariantTestSuite) SetupTest() {
	// Configure SDK with Aura-specific prefixes
	keepertest.ConfigureSDK()

	input := keepertest.CreateTestInput(suite.T())
	suite.BankKeeper = testutil.NewMockBankKeeper()

	suite.Keeper = NewKeeper(
		input.Cdc,
		input.StoreKey,
		suite.BankKeeper,
		testutil.NewMockAccountKeeper(),
		testutil.NewMockVCRegistryKeeper(),
		testutil.NewMockSecurityKeeper(),
	)
	suite.SdkCtx = input.Ctx
}

// Helper function to generate test addresses
func (suite *LPTokenInvariantTestSuite) testAddr(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	bech32, err := sdk.Bech32ifyAddressBytes("aura", sum[:20])
	suite.Require().NoError(err)
	return bech32
}

// TestValidateLPTokenInvariant_ValidPool tests invariant validation with a valid pool
func (suite *LPTokenInvariantTestSuite) TestValidateLPTokenInvariant_ValidPool() {
	pool := &types.LiquidityPool{
		PoolId:          "uaura-usdt",
		DenomA:          "uaura",
		DenomB:          "usdt",
		ReserveA:        sdkmath.NewInt(1000000),
		ReserveB:        sdkmath.NewInt(500000),
		TotalLpTokens:   sdkmath.NewInt(707106), // sqrt(1000000 * 500000)
		LockedLiquidity: sdkmath.NewInt(1000),   // MinimumLiquidity locked
		Providers: []types.LiquidityProvider{
			{
				Address:  suite.testAddr("provider1"),
				LpTokens: sdkmath.NewInt(706106), // Total - Locked = 707106 - 1000
			},
		},
	}

	err := suite.Keeper.validateLPTokenInvariant(pool)
	suite.NoError(err, "valid pool should pass invariant check")
}

// TestValidateLPTokenInvariant_MultipleProviders tests invariant with multiple providers
func (suite *LPTokenInvariantTestSuite) TestValidateLPTokenInvariant_MultipleProviders() {
	pool := &types.LiquidityPool{
		PoolId:          "uaura-usdc",
		DenomA:          "uaura",
		DenomB:          "usdc",
		ReserveA:        sdkmath.NewInt(2000000),
		ReserveB:        sdkmath.NewInt(1000000),
		TotalLpTokens:   sdkmath.NewInt(1415213), // sqrt(2000000 * 1000000) + locked
		LockedLiquidity: sdkmath.NewInt(1000),
		Providers: []types.LiquidityProvider{
			{
				Address:  suite.testAddr("provider1"),
				LpTokens: sdkmath.NewInt(500000),
			},
			{
				Address:  suite.testAddr("provider2"),
				LpTokens: sdkmath.NewInt(500000),
			},
			{
				Address:  suite.testAddr("provider3"),
				LpTokens: sdkmath.NewInt(414213), // Remainder
			},
		},
	}

	err := suite.Keeper.validateLPTokenInvariant(pool)
	suite.NoError(err, "pool with multiple providers should pass invariant")
}

// TestValidateLPTokenInvariant_MismatchedTotals tests detection of mismatched totals
func (suite *LPTokenInvariantTestSuite) TestValidateLPTokenInvariant_MismatchedTotals() {
	pool := &types.LiquidityPool{
		PoolId:          "uaura-dai",
		DenomA:          "uaura",
		DenomB:          "dai",
		ReserveA:        sdkmath.NewInt(1000000),
		ReserveB:        sdkmath.NewInt(500000),
		TotalLpTokens:   sdkmath.NewInt(1000000), // WRONG: doesn't match provider sum + locked
		LockedLiquidity: sdkmath.NewInt(1000),
		Providers: []types.LiquidityProvider{
			{
				Address:  suite.testAddr("provider1"),
				LpTokens: sdkmath.NewInt(500000), // Sum + locked = 500000 + 1000 = 501000 != 1000000
			},
		},
	}

	err := suite.Keeper.validateLPTokenInvariant(pool)
	suite.Error(err, "mismatched totals should fail invariant")
	suite.Contains(err.Error(), "CRITICAL: LP token invariant violated")
	suite.Contains(err.Error(), "1000000")
	suite.Contains(err.Error(), "500000")
}

// TestValidateLPTokenInvariant_InflatedSupply tests detection of inflated supply
func (suite *LPTokenInvariantTestSuite) TestValidateLPTokenInvariant_InflatedSupply() {
	pool := &types.LiquidityPool{
		PoolId:          "uaura-busd",
		DenomA:          "uaura",
		DenomB:          "busd",
		ReserveA:        sdkmath.NewInt(1000000),
		ReserveB:        sdkmath.NewInt(500000),
		TotalLpTokens:   sdkmath.NewInt(500000), // WRONG: less than provider tokens
		LockedLiquidity: sdkmath.NewInt(1000),
		Providers: []types.LiquidityProvider{
			{
				Address:  suite.testAddr("provider1"),
				LpTokens: sdkmath.NewInt(600000), // Provider has more than total!
			},
		},
	}

	err := suite.Keeper.validateLPTokenInvariant(pool)
	suite.Error(err, "inflated provider balance should fail invariant")
	suite.Contains(err.Error(), "CRITICAL: LP token invariant violated")
}

// TestValidateLPTokenInvariant_NoLockedLiquidity tests pools without locked liquidity
func (suite *LPTokenInvariantTestSuite) TestValidateLPTokenInvariant_NoLockedLiquidity() {
	// This could happen in tests or edge cases where pool is created differently
	pool := &types.LiquidityPool{
		PoolId:          "uaura-test",
		DenomA:          "uaura",
		DenomB:          "test",
		ReserveA:        sdkmath.NewInt(1000000),
		ReserveB:        sdkmath.NewInt(500000),
		TotalLpTokens:   sdkmath.NewInt(707106),
		LockedLiquidity: sdkmath.ZeroInt(), // No locked liquidity
		Providers: []types.LiquidityProvider{
			{
				Address:  suite.testAddr("provider1"),
				LpTokens: sdkmath.NewInt(707106), // All tokens to provider
			},
		},
	}

	err := suite.Keeper.validateLPTokenInvariant(pool)
	suite.NoError(err, "pool without locked liquidity should still validate correctly")
}

// TestValidateLPTokenInvariant_InvalidProviderTokens tests detection of invalid provider tokens
func (suite *LPTokenInvariantTestSuite) TestValidateLPTokenInvariant_InvalidProviderTokens() {
	pool := &types.LiquidityPool{
		PoolId:          "uaura-invalid",
		DenomA:          "uaura",
		DenomB:          "invalid",
		ReserveA:        sdkmath.NewInt(1000000),
		ReserveB:        sdkmath.NewInt(500000),
		TotalLpTokens:   sdkmath.NewInt(707106),
		LockedLiquidity: sdkmath.NewInt(1000),
		Providers: []types.LiquidityProvider{
			{
				Address:  suite.testAddr("provider1"),
				LpTokens: sdkmath.NewInt(-1), // Invalid - negative tokens should fail validation
			},
		},
	}

	err := suite.Keeper.validateLPTokenInvariant(pool)
	suite.Error(err, "invalid provider LP tokens should fail")
	suite.Contains(err.Error(), "LP token invariant violated")
}

// TestValidateLPTokenInvariant_InvalidTotalTokens tests detection of invalid total tokens
func (suite *LPTokenInvariantTestSuite) TestValidateLPTokenInvariant_InvalidTotalTokens() {
	pool := &types.LiquidityPool{
		PoolId:          "uaura-invalid2",
		DenomA:          "uaura",
		DenomB:          "invalid2",
		ReserveA:        sdkmath.NewInt(1000000),
		ReserveB:        sdkmath.NewInt(500000),
		TotalLpTokens:   sdkmath.NewInt(-1), // Invalid - negative total tokens should fail validation
		LockedLiquidity: sdkmath.NewInt(1000),
		Providers: []types.LiquidityProvider{
			{
				Address:  suite.testAddr("provider1"),
				LpTokens: sdkmath.NewInt(706106),
			},
		},
	}

	err := suite.Keeper.validateLPTokenInvariant(pool)
	suite.Error(err, "invalid total LP tokens should fail")
	suite.Contains(err.Error(), "LP token invariant violated")
}

// TestCreatePool_InvariantValidation tests that CreatePool validates invariant
func (suite *LPTokenInvariantTestSuite) TestCreatePool_InvariantValidation() {
	ctx := suite.SdkCtx
	creator := suite.testAddr("creator")

	// Fund creator
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)
	suite.fundAccount(ctx, creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(2000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(1000000)),
	))

	// Create pool
	pool, lpTokens, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(500000)),
	)

	suite.Require().NoError(err, "CreatePool should succeed and pass invariant validation")
	suite.Require().NotNil(pool)
	suite.Require().True(lpTokens.IsPositive())

	// Verify invariant holds
	err = suite.Keeper.validateLPTokenInvariant(pool)
	suite.NoError(err, "created pool should maintain invariant")
}

// TestAddLiquidity_InvariantValidation tests that AddLiquidity validates invariant
func (suite *LPTokenInvariantTestSuite) TestAddLiquidity_InvariantValidation() {
	ctx := suite.SdkCtx
	creator := suite.testAddr("creator")
	provider := suite.testAddr("provider")

	// Fund accounts
	creatorAddr, _ := sdk.AccAddressFromBech32(creator)
	providerAddr, _ := sdk.AccAddressFromBech32(provider)
	suite.fundAccount(ctx, creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(2000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(1000000)),
	))
	suite.fundAccount(ctx, providerAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(2000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(1000000)),
	))

	// Create pool
	_, _, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(500000)),
	)
	suite.Require().NoError(err)

	// Add liquidity
	lpTokens, _, err := suite.Keeper.AddLiquidity(
		ctx,
		provider,
		"uaura-usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(500000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(250000)),
	)

	suite.Require().NoError(err, "AddLiquidity should succeed and pass invariant validation")
	suite.Require().True(lpTokens.IsPositive())

	// Verify invariant holds
	pool := suite.Keeper.GetPool(ctx, "uaura-usdt")
	err = suite.Keeper.validateLPTokenInvariant(pool)
	suite.NoError(err, "pool after adding liquidity should maintain invariant")
}

// TestRemoveLiquidity_InvariantValidation tests that RemoveLiquidity validates invariant
func (suite *LPTokenInvariantTestSuite) TestRemoveLiquidity_InvariantValidation() {
	ctx := suite.SdkCtx
	creator := suite.testAddr("creator")

	// Fund creator
	creatorAddr, _ := sdk.AccAddressFromBech32(creator)
	suite.fundAccount(ctx, creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(2000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(1000000)),
	))

	// Create pool
	_, lpTokens, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(500000)),
	)
	suite.Require().NoError(err)

	// Remove half the liquidity
	halfTokens := lpTokens.QuoRaw(2)
	coinA, coinB, err := suite.Keeper.RemoveLiquidity(
		ctx,
		creator,
		"uaura-usdt",
		halfTokens,
	)

	suite.Require().NoError(err, "RemoveLiquidity should succeed and pass invariant validation")
	suite.Require().True(coinA.Amount.IsPositive())
	suite.Require().True(coinB.Amount.IsPositive())

	// Verify invariant holds
	pool := suite.Keeper.GetPool(ctx, "uaura-usdt")
	err = suite.Keeper.validateLPTokenInvariant(pool)
	suite.NoError(err, "pool after removing liquidity should maintain invariant")
}

// TestSwap_InvariantValidation tests that Swap validates invariant
func (suite *LPTokenInvariantTestSuite) TestSwap_InvariantValidation() {
	ctx := suite.SdkCtx
	creator := suite.testAddr("creator")
	trader := suite.testAddr("trader")

	// Fund accounts
	creatorAddr, _ := sdk.AccAddressFromBech32(creator)
	traderAddr, _ := sdk.AccAddressFromBech32(trader)
	suite.fundAccount(ctx, creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(2000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(1000000)),
	))
	suite.fundAccount(ctx, traderAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(100000)),
	))

	// Create pool
	_, _, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(500000)),
	)
	suite.Require().NoError(err)

	// Execute swap
	amountOut, _, _, err := suite.Keeper.SwapExactIn(
		ctx,
		trader,
		"uaura-usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(10000)),
		sdkmath.NewInt(1), // minAmountOut
		0,                 // no slippage limit
	)

	suite.Require().NoError(err, "Swap should succeed and pass invariant validation")
	suite.Require().True(amountOut.IsPositive())

	// Verify invariant holds (swaps don't affect LP tokens, but we check anyway)
	pool := suite.Keeper.GetPool(ctx, "uaura-usdt")
	err = suite.Keeper.validateLPTokenInvariant(pool)
	suite.NoError(err, "pool after swap should maintain invariant")
}

// TestInvariant_DetectsManualCorruption tests that the invariant detects manual state corruption
func (suite *LPTokenInvariantTestSuite) TestInvariant_DetectsManualCorruption() {
	ctx := suite.SdkCtx
	creator := suite.testAddr("creator")

	// Fund creator
	creatorAddr, _ := sdk.AccAddressFromBech32(creator)
	suite.fundAccount(ctx, creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(2000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(1000000)),
	))

	// Create pool
	pool, _, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(500000)),
	)
	suite.Require().NoError(err)

	// Manually corrupt the pool state (simulating a bug)
	pool.Providers[0].LpTokens = sdkmath.NewInt(999999999) // Way more than total
	suite.Require().NoError(suite.Keeper.SetPool(ctx, pool))

	// Verify the invariant now fails
	err = suite.Keeper.validateLPTokenInvariant(pool)
	suite.Error(err, "corrupted pool should fail invariant validation")
	suite.Contains(err.Error(), "CRITICAL: LP token invariant violated")
}

// TestInvariant_ZeroProviders tests invariant with no providers (edge case)
func (suite *LPTokenInvariantTestSuite) TestInvariant_ZeroProviders() {
	pool := &types.LiquidityPool{
		PoolId:          "uaura-empty",
		DenomA:          "uaura",
		DenomB:          "empty",
		ReserveA:        sdkmath.NewInt(1000000),
		ReserveB:        sdkmath.NewInt(500000),
		TotalLpTokens:   sdkmath.NewInt(1000), // Only locked liquidity
		LockedLiquidity: sdkmath.NewInt(1000),
		Providers: []types.LiquidityProvider{}, // No providers
	}

	err := suite.Keeper.validateLPTokenInvariant(pool)
	suite.NoError(err, "pool with only locked liquidity and no providers should pass")
}

// TestInvariant_AccountingErrorPrevention tests that the invariant prevents accounting errors
func (suite *LPTokenInvariantTestSuite) TestInvariant_AccountingErrorPrevention() {
	// Simulate an accounting error scenario
	pool := &types.LiquidityPool{
		PoolId:          "uaura-error",
		DenomA:          "uaura",
		DenomB:          "error",
		ReserveA:        sdkmath.NewInt(1000000),
		ReserveB:        sdkmath.NewInt(500000),
		TotalLpTokens:   sdkmath.NewInt(707106),
		LockedLiquidity: sdkmath.NewInt(1000),
		Providers: []types.LiquidityProvider{
			{
				Address:  suite.testAddr("provider1"),
				LpTokens: sdkmath.NewInt(300000),
			},
			{
				Address:  suite.testAddr("provider2"),
				LpTokens: sdkmath.NewInt(400000), // Sum = 700000, but should be 706106 (707106 - 1000)
			},
			// Missing 6106 tokens - accounting error!
		},
	}

	err := suite.Keeper.validateLPTokenInvariant(pool)
	suite.Error(err, "accounting error should be detected")
	suite.Contains(err.Error(), "CRITICAL: LP token invariant violated")
	suite.Contains(err.Error(), "707106") // Total
	suite.Contains(err.Error(), "700000") // Provider sum
}

// TestModuleLevelInvariant_AllPools tests the module-level invariant across all pools
func (suite *LPTokenInvariantTestSuite) TestModuleLevelInvariant_AllPools() {
	ctx := suite.SdkCtx
	creator1 := suite.testAddr("creator1")
	creator2 := suite.testAddr("creator2")

	// Fund creators
	creator1Addr, _ := sdk.AccAddressFromBech32(creator1)
	creator2Addr, _ := sdk.AccAddressFromBech32(creator2)
	suite.fundAccount(ctx, creator1Addr, sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(5000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(2500000)),
	))
	suite.fundAccount(ctx, creator2Addr, sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(5000000)),
		sdk.NewCoin("usdc", sdkmath.NewInt(2500000)),
	))

	// Create multiple pools
	_, _, err := suite.Keeper.CreatePool(
		ctx,
		creator1,
		"uaura",
		"usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(500000)),
	)
	suite.Require().NoError(err)

	_, _, err = suite.Keeper.CreatePool(
		ctx,
		creator2,
		"uaura",
		"usdc",
		sdk.NewCoin("uaura", sdkmath.NewInt(2000000)),
		sdk.NewCoin("usdc", sdkmath.NewInt(1000000)),
	)
	suite.Require().NoError(err)

	// Run the module-level invariant
	invariantFn := LiquidityProviderConsistencyInvariant(suite.Keeper)
	msg, broken := invariantFn(ctx)

	suite.False(broken, "module-level invariant should pass for all pools")
	suite.Empty(msg, "no error message should be returned")
}

// TestModuleLevelInvariant_DetectsViolation tests that module-level invariant detects violations
func (suite *LPTokenInvariantTestSuite) TestModuleLevelInvariant_DetectsViolation() {
	ctx := suite.SdkCtx
	creator := suite.testAddr("creator")

	// Fund creator
	creatorAddr, _ := sdk.AccAddressFromBech32(creator)
	suite.fundAccount(ctx, creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(2000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(1000000)),
	))

	// Create pool
	pool, _, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(500000)),
	)
	suite.Require().NoError(err)

	// Manually corrupt the pool (simulate accounting error)
	pool.TotalLpTokens = sdkmath.NewInt(999999999) // Inflate total
	suite.Require().NoError(suite.Keeper.SetPool(ctx, pool))

	// Run the module-level invariant
	invariantFn := LiquidityProviderConsistencyInvariant(suite.Keeper)
	msg, broken := invariantFn(ctx)

	suite.True(broken, "module-level invariant should detect violation")
	suite.Contains(msg, "CRITICAL")
	suite.Contains(msg, "invariant violated")
	suite.Contains(msg, pool.PoolId)
}

// TestModuleLevelInvariant_WithLockedLiquidity tests invariant accounts for locked liquidity
func (suite *LPTokenInvariantTestSuite) TestModuleLevelInvariant_WithLockedLiquidity() {
	ctx := suite.SdkCtx

	// Manually create a pool with locked liquidity
	pool := &types.LiquidityPool{
		PoolId:          "uaura-test",
		DenomA:          "uaura",
		DenomB:          "test",
		ReserveA:        sdkmath.NewInt(1000000),
		ReserveB:        sdkmath.NewInt(500000),
		TotalLpTokens:   sdkmath.NewInt(707106),
		LockedLiquidity: sdkmath.NewInt(1000),
		Providers: []types.LiquidityProvider{
			{
				Address:  suite.testAddr("provider1"),
				LpTokens: sdkmath.NewInt(706106), // Total - Locked = 707106 - 1000
			},
		},
	}
	suite.Require().NoError(suite.Keeper.SetPool(ctx, pool))

	// Run the module-level invariant
	invariantFn := LiquidityProviderConsistencyInvariant(suite.Keeper)
	msg, broken := invariantFn(ctx)

	suite.False(broken, "invariant should pass when locked liquidity is correctly accounted")
	suite.Empty(msg)
}

// TestModuleLevelInvariant_MissingLockedLiquidity tests detection of missing locked liquidity
func (suite *LPTokenInvariantTestSuite) TestModuleLevelInvariant_MissingLockedLiquidity() {
	ctx := suite.SdkCtx

	// Create pool where provider tokens + locked != total (accounting error)
	pool := &types.LiquidityPool{
		PoolId:          "uaura-error",
		DenomA:          "uaura",
		DenomB:          "error",
		ReserveA:        sdkmath.NewInt(1000000),
		ReserveB:        sdkmath.NewInt(500000),
		TotalLpTokens:   sdkmath.NewInt(707106),
		LockedLiquidity: sdkmath.NewInt(1000),
		Providers: []types.LiquidityProvider{
			{
				Address:  suite.testAddr("provider1"),
				LpTokens: sdkmath.NewInt(700000), // Sum + locked = 701000, but total is 707106!
			},
		},
	}
	suite.Require().NoError(suite.Keeper.SetPool(ctx, pool))

	// Run the module-level invariant
	invariantFn := LiquidityProviderConsistencyInvariant(suite.Keeper)
	msg, broken := invariantFn(ctx)

	suite.True(broken, "invariant should detect mismatch")
	suite.Contains(msg, "CRITICAL")
	suite.Contains(msg, "707106")
	suite.Contains(msg, "700000")
	suite.Contains(msg, "1000")
}

// TestModuleLevelInvariant_InvalidLockedLiquidity tests detection of invalid locked liquidity
func (suite *LPTokenInvariantTestSuite) TestModuleLevelInvariant_InvalidLockedLiquidity() {
	ctx := suite.SdkCtx

	// Create pool with invalid locked liquidity value
	pool := &types.LiquidityPool{
		PoolId:          "uaura-invalid",
		DenomA:          "uaura",
		DenomB:          "invalid",
		ReserveA:        sdkmath.NewInt(1000000),
		ReserveB:        sdkmath.NewInt(500000),
		TotalLpTokens:   sdkmath.NewInt(707106),
		LockedLiquidity: sdkmath.NewInt(-1), // Invalid - negative locked liquidity should fail validation
		Providers: []types.LiquidityProvider{
			{
				Address:  suite.testAddr("provider1"),
				LpTokens: sdkmath.NewInt(706106),
			},
		},
	}
	suite.Require().NoError(suite.Keeper.SetPool(ctx, pool))

	// Run the module-level invariant
	invariantFn := LiquidityProviderConsistencyInvariant(suite.Keeper)
	msg, broken := invariantFn(ctx)

	suite.True(broken, "invariant should detect invalid locked liquidity")
	suite.Contains(msg, "invalid locked liquidity")
}

// TestModuleLevelInvariant_EmptyLockedLiquidity tests pools without locked liquidity
func (suite *LPTokenInvariantTestSuite) TestModuleLevelInvariant_EmptyLockedLiquidity() {
	ctx := suite.SdkCtx

	// Create pool without locked liquidity (legacy pool or test scenario)
	pool := &types.LiquidityPool{
		PoolId:          "uaura-legacy",
		DenomA:          "uaura",
		DenomB:          "legacy",
		ReserveA:        sdkmath.NewInt(1000000),
		ReserveB:        sdkmath.NewInt(500000),
		TotalLpTokens:   sdkmath.NewInt(707106),
		LockedLiquidity: sdkmath.ZeroInt(), // No locked liquidity
		Providers: []types.LiquidityProvider{
			{
				Address:  suite.testAddr("provider1"),
				LpTokens: sdkmath.NewInt(707106), // All tokens to provider
			},
		},
	}
	suite.Require().NoError(suite.Keeper.SetPool(ctx, pool))

	// Run the module-level invariant
	invariantFn := LiquidityProviderConsistencyInvariant(suite.Keeper)
	msg, broken := invariantFn(ctx)

	suite.False(broken, "invariant should handle empty locked liquidity (treats as zero)")
	suite.Empty(msg)
}

// TestAllInvariants tests that AllInvariants function includes LP token invariant
func (suite *LPTokenInvariantTestSuite) TestAllInvariants() {
	ctx := suite.SdkCtx
	creator := suite.testAddr("creator")

	// Fund creator
	creatorAddr, _ := sdk.AccAddressFromBech32(creator)
	suite.fundAccount(ctx, creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(2000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(1000000)),
	))

	// Create pool
	_, _, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(500000)),
	)
	suite.Require().NoError(err)

	// Run all invariants
	allInvariantsFn := AllInvariants(suite.Keeper)
	msg, broken := allInvariantsFn(ctx)

	suite.False(broken, "all invariants should pass for valid pool")
	suite.Empty(msg)
}

// Helper function to fund an account
func (suite *LPTokenInvariantTestSuite) fundAccount(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	err := suite.BankKeeper.MintCoins(ctx, types.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, coins)
	suite.Require().NoError(err)
}
