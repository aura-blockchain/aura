package keeper_test

import (
	"github.com/aequitas/aura/chain/testing/testutil"
	"math/big"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dex/keeper"
)

// LPInflationAttackTestSuite tests LP token inflation attack prevention
// Tests cover the classic "first depositor attack" and related vulnerabilities
// where an attacker manipulates LP token value to steal from subsequent depositors.
type LPInflationAttackTestSuite struct {
	suite.Suite

	Keeper     *keeper.Keeper
	SdkCtx     sdk.Context
	BankKeeper *MockBankKeeper
}

func TestLPInflationAttackTestSuite(t *testing.T) {
	suite.Run(t, new(LPInflationAttackTestSuite))
}

// SetupTest initializes the test suite before each test
func (suite *LPInflationAttackTestSuite) SetupTest() {
	// Configure SDK with Aura-specific prefixes
	keepertest.ConfigureSDK()

	input := keepertest.CreateTestInput(suite.T())
	suite.BankKeeper = NewMockBankKeeper()

	suite.Keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		suite.BankKeeper,
		testutil.NewMockAccountKeeper(),
		testutil.NewMockVCRegistryKeeper(),
		testutil.NewMockSecurityKeeper(),
	)
	suite.SdkCtx = input.Ctx
}

// Helper function to fund an account
func (suite *LPInflationAttackTestSuite) fundAccount(addr sdk.AccAddress, coins sdk.Coins) {
	for _, coin := range coins {
		suite.BankKeeper.SetBalance(addr, coin.Denom, coin.Amount)
	}
}

// TestFirstDepositorAttack_Prevention tests that the classic first depositor attack is prevented
//
// Classic Attack Scenario (WITHOUT minimum liquidity burn):
// 1. Attacker creates pool with 1 WEI of each token, receives 1 LP token
// 2. Attacker donates large amounts directly to pool (not via AddLiquidity)
// 3. Pool reserves are now imbalanced, but LP supply is still 1
// 4. Victim adds liquidity, receives 0 LP tokens due to rounding
// 5. Attacker withdraws, stealing victim's deposit
//
// Prevention: By burning MinimumLiquidity (1000) LP tokens on pool creation,
// the LP token value stays reasonable and rounding attacks become prohibitively expensive.
func (suite *LPInflationAttackTestSuite) TestFirstDepositorAttack_Prevention() {
	ctx := suite.SdkCtx

	// Generate test addresses
	attackerAddr := keepertest.GenTestAddrs(1)[0]
	attacker := attackerAddr.String()

	// ATTEMPT: Create pool with minimal amounts (1 WEI each)
	// This would normally give the attacker complete control over LP token value
	minimalAmount := math.NewInt(1)

	// Fund attacker with minimal amounts
	suite.fundAccount(attackerAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", minimalAmount),
		sdk.NewCoin("usdt", minimalAmount),
	))

	// CRITICAL: This should FAIL due to two protections:
	// 1. Minimum liquidity requirement ($1000 worth, ~10,000 AURA)
	// 2. Even if that passes, sqrt(1*1) = 1 LP token < MinimumLiquidity (1000)
	_, _, err := suite.Keeper.CreatePool(
		ctx,
		attacker,
		"uaura",
		"usdt",
		sdk.NewCoin("uaura", minimalAmount),
		sdk.NewCoin("usdt", minimalAmount),
	)

	// Verify attack is blocked
	suite.Require().Error(err, "creating pool with 1 WEI should fail")
	suite.Require().Contains(err.Error(), "minimum liquidity requirement not met",
		"error should indicate minimum liquidity requirement not met")
}

// TestFirstDepositorAttack_MinimumThreshold tests the LP burn threshold with realistic pool
func (suite *LPInflationAttackTestSuite) TestFirstDepositorAttack_MinimumThreshold() {
	ctx := suite.SdkCtx
	creatorAddr := keepertest.GenTestAddrs(1)[0]
	creator := creatorAddr.String()

	// Use minimum viable amounts that meet the $1000 requirement
	// At $0.10 per AURA, need 10,000 AURA minimum
	minAmount := math.NewInt(10_000)

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", minAmount),
		sdk.NewCoin("usdt", minAmount),
	))

	// This should succeed: sqrt(10000*10000) = 10,000 LP tokens
	// After burning 1000, creator gets 9,000 LP tokens
	pool, lpTokens, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"usdt",
		sdk.NewCoin("uaura", minAmount),
		sdk.NewCoin("usdt", minAmount),
	)

	suite.Require().NoError(err, "pool with minimum amounts should succeed")
	suite.Require().NotNil(pool)

	// Verify minimum liquidity was burned
	suite.Require().Equal(math.NewInt(10000).String(), pool.TotalLpTokens.String(), "total LP tokens should be 10000")
	suite.Require().Equal(math.NewInt(1000).String(), pool.LockedLiquidity.String(), "locked liquidity should be 1000")
	suite.Require().Equal("9000", lpTokens.String(), "creator should receive 9000 LP tokens (10000 - 1000)")

	// Verify LP token invariant holds
	err = suite.Keeper.ValidateLPTokenInvariantExported(pool)
	suite.Require().NoError(err, "pool should satisfy LP token invariant")
}

// TestFirstDepositorAttack_LargerPool tests minimum liquidity burn with realistic pool
func (suite *LPInflationAttackTestSuite) TestFirstDepositorAttack_LargerPool() {
	ctx := suite.SdkCtx
	creatorAddr := keepertest.GenTestAddrs(1)[0]
	creator := creatorAddr.String()

	// Create pool with realistic amounts
	amountA := math.NewInt(1_000_000) // 1M tokens
	amountB := math.NewInt(500_000)   // 500K tokens

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", amountA),
		sdk.NewCoin("dai", amountB),
	))

	pool, lpTokens, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"dai",
		sdk.NewCoin("uaura", amountA),
		sdk.NewCoin("dai", amountB),
	)

	suite.Require().NoError(err)
	suite.Require().NotNil(pool)

	// Calculate expected LP tokens: sqrt(1000000 * 500000) = 707,106
	expectedTotal := math.NewIntFromBigInt(new(big.Int).Sqrt(
		amountA.Mul(amountB).BigInt(),
	))
	suite.Require().Equal(expectedTotal.String(), pool.TotalLpTokens.String(),
		"total LP tokens should be sqrt(x*y)")

	// Verify minimum liquidity was locked
	suite.Require().Equal(math.NewInt(1000).String(), pool.LockedLiquidity.String())

	// Creator receives total minus locked
	expectedCreatorLP := expectedTotal.Sub(math.NewInt(1000))
	suite.Require().Equal(expectedCreatorLP.String(), lpTokens.String(),
		"creator should receive total minus locked liquidity")

	// Verify invariant
	err = suite.Keeper.ValidateLPTokenInvariantExported(pool)
	suite.Require().NoError(err)
}

// TestDustAttack_Prevention tests that dust deposits are rejected
//
// Dust Attack Scenario (after pool exists with large reserves):
// 1. Pool has large reserves (e.g., 1B tokens each)
// 2. Attacker donates massive amounts to inflate reserves
// 3. Victim tries to add small liquidity
// 4. Due to rounding, victim receives 0 LP tokens but tokens are taken
//
// Prevention: AddLiquidity rejects any deposit that would result in 0 LP tokens
// Additionally, minimum liquidity requirements prevent dust deposits
func (suite *LPInflationAttackTestSuite) TestDustAttack_Prevention() {
	ctx := suite.SdkCtx

	// Create pool with large reserves
	creatorAddr := keepertest.GenTestAddrs(1)[0]
	creator := creatorAddr.String()
	largeAmount := math.NewInt(1_000_000_000) // 1 billion

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", largeAmount),
		sdk.NewCoin("usdc", largeAmount),
	))

	pool, _, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"usdc",
		sdk.NewCoin("uaura", largeAmount),
		sdk.NewCoin("usdc", largeAmount),
	)
	suite.Require().NoError(err)

	// Simulate attacker donating to inflate reserves (bypassing AddLiquidity)
	// This would normally be done by sending tokens directly to the module account
	// For simplicity, we'll manually inflate the reserves
	inflatedReserve := largeAmount.Mul(math.NewInt(1000)) // 1000x inflation
	pool.ReserveA = inflatedReserve
	pool.ReserveB = inflatedReserve
	suite.Require().NoError(suite.Keeper.SetPool(ctx, pool))

	// Victim attempts to add liquidity with small amount
	victimAddr := keepertest.GenTestAddrs(2)[1]
	victim := victimAddr.String()
	dustAmount := math.NewInt(100) // Small amount

	suite.fundAccount(victimAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", dustAmount),
		sdk.NewCoin("usdc", dustAmount),
	))

	// CRITICAL: This should FAIL due to minimum liquidity requirements
	// The minimum liquidity check prevents new LPs from adding dust amounts
	_, _, err = suite.Keeper.AddLiquidity(
		ctx,
		victim,
		"uaura-usdc",
		sdk.NewCoin("uaura", dustAmount),
		sdk.NewCoin("usdc", dustAmount),
	)

	suite.Require().Error(err, "dust liquidity addition should fail")
	suite.Require().Contains(err.Error(), "minimum liquidity requirement not met",
		"error should indicate minimum liquidity requirement not met")
}

// TestDustAttack_SafeMinimumDeposit tests that existing LPs can add any amount
func (suite *LPInflationAttackTestSuite) TestDustAttack_SafeMinimumDeposit() {
	ctx := suite.SdkCtx

	// Create pool
	creatorAddr := keepertest.GenTestAddrs(1)[0]
	creator := creatorAddr.String()
	amount := math.NewInt(1_000_000)

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", amount.Mul(math.NewInt(2))),
		sdk.NewCoin("busd", amount.Mul(math.NewInt(2))),
	))

	_, _, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"busd",
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("busd", amount),
	)
	suite.Require().NoError(err)

	// Creator (existing LP) can add any amount, even small amounts
	// because they're grandfathered in
	addAmount := math.NewInt(100) // Small amount

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", addAmount),
		sdk.NewCoin("busd", addAmount),
	))

	lpTokens, _, err := suite.Keeper.AddLiquidity(
		ctx,
		creator,
		"busd-uaura",
		sdk.NewCoin("uaura", addAmount),
		sdk.NewCoin("busd", addAmount),
	)

	// This should succeed for existing LP
	suite.Require().NoError(err, "existing LP can add any amount")
	suite.Require().True(lpTokens.IsPositive(), "should receive positive LP tokens")
}

// TestDonationAttack_LPInvariantProtection tests that donated tokens don't break invariant
//
// Donation Attack Scenario:
// 1. Pool exists with normal reserves and LP tokens
// 2. Attacker donates large amounts directly to pool (not via AddLiquidity)
// 3. This inflates reserves without minting LP tokens
// 4. Next depositor gets unfairly diluted
//
// Protection: While donation itself can't be prevented, the minimum liquidity
// burn and zero-LP-token rejection make this attack economically infeasible
func (suite *LPInflationAttackTestSuite) TestDonationAttack_LPInvariantProtection() {
	ctx := suite.SdkCtx

	// Create pool normally
	creatorAddr := keepertest.GenTestAddrs(1)[0]
	creator := creatorAddr.String()
	initialAmount := math.NewInt(1_000_000)

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", initialAmount),
		sdk.NewCoin("usdt", initialAmount),
	))

	pool, creatorLP, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"usdt",
		sdk.NewCoin("uaura", initialAmount),
		sdk.NewCoin("usdt", initialAmount),
	)
	suite.Require().NoError(err)

	// Record initial state
	initialTotalLP := pool.TotalLpTokens
	initialLockedLP := pool.LockedLiquidity

	// Simulate donation attack: manually inflate reserves
	donationAmount := math.NewInt(10_000_000) // 10x donation
	newReserveA := initialAmount.Add(donationAmount)
	newReserveB := initialAmount.Add(donationAmount)
	pool.ReserveA = newReserveA
	pool.ReserveB = newReserveB
	// NOTE: TotalLpTokens and LockedLiquidity remain unchanged - this is the attack!
	suite.Require().NoError(suite.Keeper.SetPool(ctx, pool))

	// Verify LP token invariant still holds (donations don't affect it)
	// Invariant: TotalLP = sum(provider LP) + LockedLP
	err = suite.Keeper.ValidateLPTokenInvariantExported(pool)
	suite.Require().NoError(err, "LP invariant should hold even after donation")

	// Verify the values haven't changed
	suite.Require().Equal(initialTotalLP.String(), pool.TotalLpTokens.String(),
		"donation should not affect total LP tokens")
	suite.Require().Equal(initialLockedLP.String(), pool.LockedLiquidity.String(),
		"donation should not affect locked liquidity")
	suite.Require().Equal(creatorLP.String(), pool.Providers[0].LpTokens.String(),
		"donation should not affect provider LP tokens")
}

// TestMinimumLiquidityLocked_Permanent tests that locked liquidity is permanent
func (suite *LPInflationAttackTestSuite) TestMinimumLiquidityLocked_Permanent() {
	ctx := suite.SdkCtx

	// Create pool
	creatorAddr := keepertest.GenTestAddrs(1)[0]
	creator := creatorAddr.String()
	amount := math.NewInt(1_000_000)

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("dai", amount),
	))

	pool, creatorLP, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"dai",
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("dai", amount),
	)
	suite.Require().NoError(err)

	// Verify locked liquidity is set
	suite.Require().Equal(math.NewInt(1000).String(), pool.LockedLiquidity.String())

	// Total LP = sqrt(1000000 * 1000000) = 1000000
	totalLP := pool.TotalLpTokens
	lockedLP := pool.LockedLiquidity

	// Creator receives total minus locked
	suite.Require().Equal(totalLP.Sub(lockedLP).String(), creatorLP.String())

	// Attempt to remove ALL liquidity
	_, _, err = suite.Keeper.RemoveLiquidity(
		ctx,
		creator,
		"dai-uaura",
		creatorLP,
	)
	suite.Require().NoError(err)

	// Verify locked liquidity remains in pool
	pool = suite.Keeper.GetPool(ctx, "dai-uaura")
	suite.Require().NotNil(pool)
	suite.Require().Equal(math.NewInt(1000).String(), pool.LockedLiquidity.String(),
		"locked liquidity should remain even after all providers exit")
	suite.Require().Equal(math.NewInt(1000).String(), pool.TotalLpTokens.String(),
		"only locked liquidity should remain in total LP tokens")
}

// TestMinimumLiquidityEvent_Emitted tests that minimum liquidity lock event is emitted
func (suite *LPInflationAttackTestSuite) TestMinimumLiquidityEvent_Emitted() {
	ctx := suite.SdkCtx

	creatorAddr := keepertest.GenTestAddrs(1)[0]
	creator := creatorAddr.String()
	amount := math.NewInt(1_000_000)

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("usdc", amount),
	))

	_, _, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"usdc",
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("usdc", amount),
	)
	suite.Require().NoError(err)

	// Check that the minimum_liquidity_locked event was emitted
	events := ctx.EventManager().Events()
	foundEvent := false
	for _, event := range events {
		if event.Type == "minimum_liquidity_locked" {
			foundEvent = true
			// Verify event attributes
			for _, attr := range event.Attributes {
				if attr.Key == "pool_id" {
					suite.Require().Equal("uaura-usdc", attr.Value)
				}
				if attr.Key == "amount" {
					suite.Require().Equal("1000", attr.Value)
				}
				if attr.Key == "reason" {
					suite.Require().Equal("first_depositor_attack_prevention", attr.Value)
				}
			}
			break
		}
	}
	suite.Require().True(foundEvent, "minimum_liquidity_locked event should be emitted")
}

// TestMultipleProviders_AfterMinimumLock tests that subsequent providers work correctly
func (suite *LPInflationAttackTestSuite) TestMultipleProviders_AfterMinimumLock() {
	ctx := suite.SdkCtx

	// Create pool
	creatorAddr := keepertest.GenTestAddrs(3)[0]
	creator := creatorAddr.String()
	amount := math.NewInt(1_000_000)

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("busd", amount),
	))

	pool, creatorLP, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"busd",
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("busd", amount),
	)
	suite.Require().NoError(err)

	initialTotalLP := pool.TotalLpTokens

	// Second provider adds liquidity
	provider2Addr := keepertest.GenTestAddrs(3)[1]
	provider2 := provider2Addr.String()
	addAmount := math.NewInt(500_000)

	suite.fundAccount(provider2Addr, sdk.NewCoins(
		sdk.NewCoin("uaura", addAmount),
		sdk.NewCoin("busd", addAmount),
	))

	lp2, _, err := suite.Keeper.AddLiquidity(
		ctx,
		provider2,
		"busd-uaura",
		sdk.NewCoin("uaura", addAmount),
		sdk.NewCoin("busd", addAmount),
	)
	suite.Require().NoError(err)
	suite.Require().True(lp2.IsPositive())

	// Third provider adds liquidity
	provider3Addr := keepertest.GenTestAddrs(3)[2]
	provider3 := provider3Addr.String()

	suite.fundAccount(provider3Addr, sdk.NewCoins(
		sdk.NewCoin("uaura", addAmount),
		sdk.NewCoin("busd", addAmount),
	))

	lp3, _, err := suite.Keeper.AddLiquidity(
		ctx,
		provider3,
		"busd-uaura",
		sdk.NewCoin("uaura", addAmount),
		sdk.NewCoin("busd", addAmount),
	)
	suite.Require().NoError(err)
	suite.Require().True(lp3.IsPositive())

	// Verify invariant holds with multiple providers
	pool = suite.Keeper.GetPool(ctx, "busd-uaura")
	err = suite.Keeper.ValidateLPTokenInvariantExported(pool)
	suite.Require().NoError(err)

	// Verify total LP increased correctly
	finalTotalLP := pool.TotalLpTokens
	expectedIncrease := lp2.Add(lp3)
	actualIncrease := finalTotalLP.Sub(initialTotalLP)
	suite.Require().Equal(expectedIncrease.String(), actualIncrease.String(),
		"total LP should increase by sum of new LP tokens")

	// Verify locked liquidity unchanged
	suite.Require().Equal(math.NewInt(1000).String(), pool.LockedLiquidity.String(),
		"locked liquidity should never change after pool creation")

	// Verify provider balances sum correctly
	sumProviderLP := creatorLP.Add(lp2).Add(lp3)
	lockedLP := math.NewInt(1000)
	expectedTotal := sumProviderLP.Add(lockedLP)
	suite.Require().Equal(expectedTotal.String(), finalTotalLP.String(),
		"invariant: total LP = sum(provider LP) + locked LP")
}

// TestLargePoolCreation_MinimumLiquidityNegligible tests that for large pools, the minimum lock is negligible
func (suite *LPInflationAttackTestSuite) TestLargePoolCreation_MinimumLiquidityNegligible() {
	ctx := suite.SdkCtx

	creatorAddr := keepertest.GenTestAddrs(1)[0]
	creator := creatorAddr.String()

	// Create pool with very large amounts (simulating mainnet pools)
	largeAmount := math.NewInt(1_000_000_000_000) // 1 trillion tokens

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", largeAmount),
		sdk.NewCoin("usdt", largeAmount),
	))

	pool, creatorLP, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"usdt",
		sdk.NewCoin("uaura", largeAmount),
		sdk.NewCoin("usdt", largeAmount),
	)
	suite.Require().NoError(err)

	// Calculate expected LP: sqrt(10^24) = 10^12
	expectedTotalLP := math.NewInt(1_000_000_000_000)
	suite.Require().Equal(expectedTotalLP.String(), pool.TotalLpTokens.String())

	// Locked liquidity is 1000 (negligible compared to 10^12)
	lockedLP := math.NewInt(1000)
	expectedCreatorLP := expectedTotalLP.Sub(lockedLP)
	suite.Require().Equal(expectedCreatorLP.String(), creatorLP.String())

	// Verify the lock is less than 0.0001% of total
	lockPercentage := lockedLP.ToLegacyDec().Quo(expectedTotalLP.ToLegacyDec()).MulInt64(100)
	suite.Require().True(lockPercentage.LT(math.LegacyNewDecWithPrec(1, 4)),
		"locked liquidity should be less than 0.0001%% for large pools")
}

// TestEdgeCase_BelowMinimumLPTokenBurn tests that pools must generate more than MinimumLiquidity
func (suite *LPInflationAttackTestSuite) TestEdgeCase_BelowMinimumLPTokenBurn() {
	ctx := suite.SdkCtx

	creatorAddr := keepertest.GenTestAddrs(1)[0]
	creator := creatorAddr.String()

	// Use amount that would generate exactly 1000 LP tokens
	// sqrt(1000 * 1000) = 1000
	// But this will fail the minimum liquidity requirement check first (need 10,000 AURA)
	amount := math.NewInt(1000)

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("test", amount),
	))

	// This should FAIL due to minimum liquidity requirement
	_, _, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"test",
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("test", amount),
	)

	suite.Require().Error(err, "pool with amount below minimum should fail")
	suite.Require().Contains(err.Error(), "minimum liquidity requirement not met")
}

// TestEdgeCase_ExactMinimumBurn tests LP burn with exactly MinimumLiquidity threshold
func (suite *LPInflationAttackTestSuite) TestEdgeCase_ExactMinimumBurn() {
	ctx := suite.SdkCtx

	creatorAddr := keepertest.GenTestAddrs(1)[0]
	creator := creatorAddr.String()

	// Use minimum amount that meets requirements (10,000 AURA)
	// sqrt(10000 * 10000) = 10,000 LP tokens
	// After burning 1000, creator gets 9,000 LP tokens
	amount := math.NewInt(10_000)

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("test", amount),
	))

	pool, creatorLP, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"test",
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("test", amount),
	)

	suite.Require().NoError(err, "pool with minimum amounts should succeed")
	suite.Require().Equal(math.NewInt(10000).String(), pool.TotalLpTokens.String())
	suite.Require().Equal(math.NewInt(1000).String(), pool.LockedLiquidity.String())
	suite.Require().Equal("9000", creatorLP.String(), "creator gets total - locked")
}

// TestConcurrentDeposits_InvariantPreservation tests that concurrent deposits maintain invariant
func (suite *LPInflationAttackTestSuite) TestConcurrentDeposits_InvariantPreservation() {
	ctx := suite.SdkCtx

	// Create pool
	creatorAddr := keepertest.GenTestAddrs(1)[0]
	creator := creatorAddr.String()
	amount := math.NewInt(1_000_000)

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("dai", amount),
	))

	_, _, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"dai",
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("dai", amount),
	)
	suite.Require().NoError(err)

	// Simulate multiple NEW providers adding liquidity in sequence
	// Each new provider must meet minimum liquidity requirements
	numProviders := 5
	addrs := keepertest.GenTestAddrs(numProviders + 1) // +1 to skip creator

	for i := 0; i < numProviders; i++ {
		provider := addrs[i+1].String() // Skip index 0 (creator)
		addAmount := math.NewInt(100_000) // Well above 10,000 minimum

		suite.fundAccount(addrs[i+1], sdk.NewCoins(
			sdk.NewCoin("uaura", addAmount),
			sdk.NewCoin("dai", addAmount),
		))

		_, _, err := suite.Keeper.AddLiquidity(
			ctx,
			provider,
			"dai-uaura",
			sdk.NewCoin("uaura", addAmount),
			sdk.NewCoin("dai", addAmount),
		)
		suite.Require().NoError(err, "provider %d should successfully add liquidity", i)

		// Verify invariant after each deposit
		pool := suite.Keeper.GetPool(ctx, "dai-uaura")
		err = suite.Keeper.ValidateLPTokenInvariantExported(pool)
		suite.Require().NoError(err, "invariant should hold after provider %d deposit", i)
	}
}

// TestGenesisExportImport_LockedLiquidity tests that locked liquidity persists through genesis export/import
func (suite *LPInflationAttackTestSuite) TestGenesisExportImport_LockedLiquidity() {
	ctx := suite.SdkCtx

	// Create pool
	creatorAddr := keepertest.GenTestAddrs(1)[0]
	creator := creatorAddr.String()
	amount := math.NewInt(1_000_000)

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("usdc", amount),
	))

	originalPool, _, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"usdc",
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("usdc", amount),
	)
	suite.Require().NoError(err)

	// Verify locked liquidity is set
	suite.Require().Equal(math.NewInt(1000).String(), originalPool.LockedLiquidity.String())

	// Simulate genesis export by reading the pool
	exportedPool := suite.Keeper.GetPool(ctx, "uaura-usdc")
	suite.Require().NotNil(exportedPool)

	// Simulate genesis import by storing the pool in a new context
	// (In real genesis import, this would be a fresh chain state)
	suite.Require().NoError(suite.Keeper.SetPool(ctx, exportedPool))

	// Verify locked liquidity persisted
	importedPool := suite.Keeper.GetPool(ctx, "uaura-usdc")
	suite.Require().Equal(originalPool.LockedLiquidity.String(), importedPool.LockedLiquidity.String(),
		"locked liquidity should persist through export/import")
	suite.Require().Equal(math.NewInt(1000).String(), importedPool.LockedLiquidity.String())

	// Verify invariant still holds
	err = suite.Keeper.ValidateLPTokenInvariantExported(importedPool)
	suite.Require().NoError(err)
}

// TestInvariantCheck_AfterEveryOperation ensures invariant is validated after all operations
func (suite *LPInflationAttackTestSuite) TestInvariantCheck_AfterEveryOperation() {
	ctx := suite.SdkCtx

	// Create pool
	creatorAddr := keepertest.GenTestAddrs(3)[0]
	creator := creatorAddr.String()
	amount := math.NewInt(1_000_000)

	suite.fundAccount(creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", amount.Mul(math.NewInt(2))),
		sdk.NewCoin("busd", amount.Mul(math.NewInt(2))),
	))

	pool, creatorLP, err := suite.Keeper.CreatePool(
		ctx,
		creator,
		"uaura",
		"busd",
		sdk.NewCoin("uaura", amount),
		sdk.NewCoin("busd", amount),
	)
	suite.Require().NoError(err)

	// After CreatePool
	err = suite.Keeper.ValidateLPTokenInvariantExported(pool)
	suite.Require().NoError(err, "invariant should hold after CreatePool")

	// After AddLiquidity
	providerAddr := keepertest.GenTestAddrs(3)[1]
	provider := providerAddr.String()
	addAmount := math.NewInt(100_000)

	suite.fundAccount(providerAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", addAmount),
		sdk.NewCoin("busd", addAmount),
	))

	providerLP, _, err := suite.Keeper.AddLiquidity(
		ctx,
		provider,
		"busd-uaura",
		sdk.NewCoin("uaura", addAmount),
		sdk.NewCoin("busd", addAmount),
	)
	suite.Require().NoError(err)

	pool = suite.Keeper.GetPool(ctx, "busd-uaura")
	err = suite.Keeper.ValidateLPTokenInvariantExported(pool)
	suite.Require().NoError(err, "invariant should hold after AddLiquidity")

	// After RemoveLiquidity
	halfLP := creatorLP.QuoRaw(2)
	_, _, err = suite.Keeper.RemoveLiquidity(ctx, creator, "busd-uaura", halfLP)
	suite.Require().NoError(err)

	pool = suite.Keeper.GetPool(ctx, "busd-uaura")
	err = suite.Keeper.ValidateLPTokenInvariantExported(pool)
	suite.Require().NoError(err, "invariant should hold after RemoveLiquidity")

	// After Swap (shouldn't affect LP tokens, but we check anyway)
	traderAddr := keepertest.GenTestAddrs(3)[2]
	trader := traderAddr.String()
	swapAmount := math.NewInt(10_000)

	suite.fundAccount(traderAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", swapAmount),
	))

	_, _, _, err = suite.Keeper.SwapExactIn(
		ctx,
		trader,
		"busd-uaura",
		sdk.NewCoin("uaura", swapAmount),
		math.NewInt(1),
		0,
	)
	suite.Require().NoError(err)

	pool = suite.Keeper.GetPool(ctx, "busd-uaura")
	err = suite.Keeper.ValidateLPTokenInvariantExported(pool)
	suite.Require().NoError(err, "invariant should hold after Swap")

	// After complete liquidity removal
	remainingCreatorLP := creatorLP.Sub(halfLP)
	_, _, err = suite.Keeper.RemoveLiquidity(ctx, creator, "busd-uaura", remainingCreatorLP)
	suite.Require().NoError(err)

	_, _, err = suite.Keeper.RemoveLiquidity(ctx, provider, "busd-uaura", providerLP)
	suite.Require().NoError(err)

	pool = suite.Keeper.GetPool(ctx, "busd-uaura")
	err = suite.Keeper.ValidateLPTokenInvariantExported(pool)
	suite.Require().NoError(err, "invariant should hold even with only locked liquidity remaining")

	// Verify only locked liquidity remains
	suite.Require().Equal(math.NewInt(1000).String(), pool.TotalLpTokens.String())
	suite.Require().Equal(math.NewInt(1000).String(), pool.LockedLiquidity.String())
	suite.Require().Len(pool.Providers, 0, "no providers should remain")
}
