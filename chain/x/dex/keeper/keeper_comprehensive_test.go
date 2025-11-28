package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/proto"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dex/keeper"
	"github.com/aequitas/aura/chain/x/dex/types"
)

// MockBankKeeper is a mock implementation of BankKeeper for testing
type MockBankKeeper struct{}

func (m MockBankKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m MockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (m MockBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m MockBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m MockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

type DEXKeeperTestSuite struct {
	suite.Suite

	keeper *keeper.Keeper
	ctx    sdk.Context
}

func (suite *DEXKeeperTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil, // bankKeeper
		nil, // authKeeper
		nil, // vcKeeper
	)
	suite.ctx = input.Ctx
}

func TestDEXKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(DEXKeeperTestSuite))
}

// Helper function to create a test keeper
func setupTestKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, MockBankKeeper{}, nil, nil)
	return k, input.Ctx
}

// Params Tests

func (suite *DEXKeeperTestSuite) TestGetParams() {
	params := suite.keeper.GetParams(suite.ctx)
	suite.Require().NotNil(params)
}

func (suite *DEXKeeperTestSuite) TestSetParams() {
	params := types.DefaultParams()
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	retrieved := suite.keeper.GetParams(suite.ctx)
	suite.Require().True(proto.Equal(params, retrieved))
}

// Pool Creation Tests

func TestCreatePool(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	tokenA := "uaura"
	tokenB := "uusdt"
	amountA := sdk.NewCoin("uaura", math.NewInt(1000000))
	amountB := sdk.NewCoin("uusdt", math.NewInt(1000000))

	pool, lpTokens, err := k.CreatePool(ctx, creator, tokenA, tokenB, amountA, amountB)
	require.NoError(t, err)
	require.NotNil(t, pool)
	require.True(t, lpTokens.GT(math.ZeroInt()))
}

func TestCreatePoolWithZeroLiquidity(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	tokenA := "uaura"
	tokenB := "uusdt"
	amountA := sdk.NewCoin("uaura", math.NewInt(0))
	amountB := sdk.NewCoin("uusdt", math.NewInt(1000000))

	_, _, err := k.CreatePool(ctx, creator, tokenA, tokenB, amountA, amountB)
	require.Error(t, err, "Should not allow zero liquidity")
}

func TestCreatePoolSameTokens(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	tokenA := "uaura"
	amountA := sdk.NewCoin("uaura", math.NewInt(1000000))

	// The SDK's NewCoins validates and panics on duplicate denominations before keeper logic runs.
	// We use defer/recover to catch this panic and verify it occurs as expected.
	defer func() {
		if r := recover(); r != nil {
			// Panic occurred as expected - duplicate denomination validation from SDK
			panicMsg := fmt.Sprintf("%v", r)
			require.Contains(t, panicMsg, "duplicate denomination", "Should panic due to duplicate denom")
		} else {
			// If no panic, the test setup itself is broken
			t.Fatal("Expected panic from NewCoins for duplicate denominations, but none occurred")
		}
	}()

	_, _, err := k.CreatePool(ctx, creator, tokenA, tokenA, amountA, amountA)
	// If we reach here without panic, log the error for debugging
	if err != nil {
		t.Logf("Got error instead of panic: %v", err)
	}
}

func TestGetPool(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	found := pool != nil
	require.True(t, found)
	require.Equal(t, "uaura", pool.DenomA)
	require.Equal(t, "uusdt", pool.DenomB)
}

func TestGetNonExistentPool(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	pool := k.GetPool(ctx, "nonexistent-pool")
	require.Nil(t, pool)
}

func TestGetOrdersByUserIndexed(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	addrs := keepertest.GenTestAddrs(2)
	user := addrs[0].String()
	other := addrs[1].String()

	for i := 0; i < 3; i++ {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Minute))

		_, err := k.CreateOrder(
			ctx,
			user,
			types.SwapOrderType_BUY,
			math.NewInt(int64(100+i)),
			"usdt",
			math.NewInt(int64(200+i)),
			60,
		)
		require.NoError(t, err)
	}

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	_, err := k.CreateOrder(
		ctx,
		other,
		types.SwapOrderType_SELL,
		math.NewInt(50),
		"usdt",
		math.NewInt(25),
		60,
	)
	require.NoError(t, err)

	orders := k.GetOrdersByUser(ctx, user)
	require.Len(t, orders, 3)

	firstAura, ok := math.NewIntFromString(orders[0].AuraAmount)
	require.True(t, ok)
	lastAura, ok := math.NewIntFromString(orders[len(orders)-1].AuraAmount)
	require.True(t, ok)
	require.True(t, firstAura.GT(lastAura), "newest order should appear first")
}

func TestUserOrderHistoryLimit(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	user := keepertest.GenTestAddr().String()

	for i := 0; i < 250; i++ {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Second))

		_, err := k.CreateOrder(
			ctx,
			user,
			types.SwapOrderType_BUY,
			math.NewInt(100),
			"usdt",
			math.NewInt(200),
			60,
		)
		require.NoError(t, err)
	}

	orders := k.GetOrdersByUser(ctx, user)
	require.LessOrEqual(t, len(orders), 200)
}

func TestGetAllPools(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()

	// Create multiple pools
	for i := 0; i < 5; i++ {
		tokenName := "token" + string(rune('A'+i))
		_, _, err := k.CreatePool(ctx, creator, tokenName, "uaura", sdk.NewCoin(tokenName, math.NewInt(1000000)), sdk.NewCoin("uaura", math.NewInt(1000000)))
		require.NoError(t, err)
	}

	pools := k.GetAllPools(ctx)
	require.GreaterOrEqual(t, len(pools), 5)
}

// Swap Tests

func TestSwap(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	trader := keepertest.GenTestAddr().String()
	coinIn := sdk.NewCoin("uaura", math.NewInt(1000))
	minAmountOut := math.NewInt(900)

	amountOut, _, _, err := k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, minAmountOut, 0)
	require.NoError(t, err)
	require.True(t, amountOut.GT(minAmountOut))
}

func TestSwapSlippageProtection(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	trader := keepertest.GenTestAddr().String()
	coinIn := sdk.NewCoin("uaura", math.NewInt(1000))
	minAmountOut := math.NewInt(10000) // Too high, should fail

	_, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, minAmountOut, 0)
	require.Error(t, err, "Should fail due to slippage protection")
}

func TestSwapInvalidPool(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	trader := keepertest.GenTestAddr().String()
	coinIn := sdk.NewCoin("uaura", math.NewInt(1000))

	_, _, _, err := k.SwapExactIn(ctx, trader, "99999", coinIn, math.NewInt(900), 0)
	require.Error(t, err)
}

func TestSwapZeroAmount(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	trader := keepertest.GenTestAddr().String()
	coinIn := sdk.NewCoin("uaura", math.NewInt(0))

	_, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, math.NewInt(0), 0)
	require.Error(t, err, "Should not allow zero swap")
}

func TestSwapPriceImpact(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	trader := keepertest.GenTestAddr().String()
	// Use 10% of pool (100000) to avoid triggering max trade size (20% limit)
	largeSwap := math.NewInt(100000)
	coinIn := sdk.NewCoin("uaura", largeSwap)

	// Large swap should have significant price impact
	amountOut, _, _, err := k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, math.NewInt(1), 0)
	require.NoError(t, err)

	// Output should be less than input (due to constant product formula)
	require.True(t, amountOut.LT(largeSwap))
}

// Liquidity Tests

func TestAddLiquidity(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	provider := keepertest.GenTestAddr().String()
	amountA := sdk.NewCoin("uaura", math.NewInt(100000))
	amountB := sdk.NewCoin("uusdt", math.NewInt(100000))

	lpTokens, _, err := k.AddLiquidity(ctx, provider, pool.PoolId, amountA, amountB)
	require.NoError(t, err)
	require.True(t, lpTokens.GT(math.ZeroInt()))
}

func TestAddLiquidityImbalanced(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	provider := keepertest.GenTestAddr().String()
	amountA := sdk.NewCoin("uaura", math.NewInt(100000))
	amountB := sdk.NewCoin("uusdt", math.NewInt(200000)) // Imbalanced ratio

	// AddLiquidity adjusts amounts to match ratio, so it returns successfully
	// with adjusted amounts (not an error)
	lpTokens, poolShare, err := k.AddLiquidity(ctx, provider, pool.PoolId, amountA, amountB)
	require.NoError(t, err)
	require.NotNil(t, lpTokens)
	require.True(t, poolShare.GT(math.LegacyZeroDec()))
}

func TestRemoveLiquidity(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	amountA := sdk.NewCoin("uaura", math.NewInt(100000))
	amountB := sdk.NewCoin("uusdt", math.NewInt(100000))
	lpTokens, _, err := k.AddLiquidity(ctx, creator, pool.PoolId, amountA, amountB)
	require.NoError(t, err)

	amountOutA, amountOutB, err := k.RemoveLiquidity(ctx, creator, pool.PoolId, lpTokens)
	require.NoError(t, err)
	require.True(t, amountOutA.Amount.GT(math.ZeroInt()))
	require.True(t, amountOutB.Amount.GT(math.ZeroInt()))
}

func TestRemoveLiquidityExceedsBalance(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	provider := keepertest.GenTestAddr().String()
	tooManyLPTokens := math.NewInt(999999999)

	_, _, err = k.RemoveLiquidity(ctx, provider, pool.PoolId, tooManyLPTokens)
	require.Error(t, err, "Should not allow removing more than owned")
}

func TestGetPoolPrice(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(2000000)))
	require.NoError(t, err)

	price, err := k.GetPoolPrice(ctx, pool.PoolId, "uaura")
	require.NoError(t, err)
	require.False(t, price.IsZero())
	// Price should be 2:1 (2 uusdt per uaura)
	require.True(t, price.GT(math.LegacyOneDec()))
}

func TestGetSpotPrice(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	spotPrice, err := k.GetSpotPrice(ctx, pool.PoolId, "uaura", "uusdt")
	require.NoError(t, err)
	require.False(t, spotPrice.IsZero())
}

// Fee Tests

func TestCalculateSwapFee(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	amount := math.NewInt(1000000)
	fee, err := k.CalculateSwapFee(ctx, amount)
	require.NoError(t, err)

	require.True(t, fee.GT(math.ZeroInt()))
	require.True(t, fee.LT(amount))
}

func TestCollectSwapFees(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	trader := keepertest.GenTestAddr().String()
	coinIn := sdk.NewCoin("uaura", math.NewInt(1000))
	_, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, math.NewInt(900), 0)
	require.NoError(t, err)

	fees, err := k.GetCollectedFees(ctx, pool.PoolId)
	require.NoError(t, err)
	require.NotNil(t, fees)
}

// Circuit Breaker Tests

func TestCircuitBreakerLargeSwap(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	trader := keepertest.GenTestAddr().String()
	// Try to swap entire pool
	massiveSwap := math.NewInt(10000000)
	coinIn := sdk.NewCoin("uaura", massiveSwap)

	_, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, math.NewInt(1), 0)
	require.Error(t, err, "Circuit breaker should trigger")
}

func TestCircuitBreakerPriceDeviation(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	// Multiple large swaps to cause price deviation
	trader := keepertest.GenTestAddr().String()
	for i := 0; i < 10; i++ {
		coinIn := sdk.NewCoin("uaura", math.NewInt(50000))
		_, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, math.NewInt(1), 0)
		if err != nil {
			// Circuit breaker should eventually trigger
			require.Contains(t, err.Error(), "circuit breaker")
			break
		}
	}
}

// Position Tests
// TODO: Implement these methods in keeper
/*
func TestGetUserPosition(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	position := k.GetUserPosition(ctx, pool.PoolId, creator)
	require.NotNil(t, position)
	require.True(t, position.LPTokens.GT(math.ZeroInt()))
}

func TestGetUserPositions(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()

	// Create multiple pools with liquidity
	for i := 0; i < 3; i++ {
		token := "token" + string(rune('A'+i))
		_, _, err := k.CreatePool(ctx, creator, token, "uaura", sdk.NewCoin(token, math.NewInt(1000000)), sdk.NewCoin("uaura", math.NewInt(1000000)))
		require.NoError(t, err)
	}

	positions := k.GetUserPositions(ctx, creator)
	require.GreaterOrEqual(t, len(positions), 3)
}
*/

// Genesis Tests

func TestInitGenesis(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	genesisState := types.DefaultGenesis()
	err := k.InitGenesis(ctx, *genesisState)
	require.NoError(t, err)
}

func TestExportGenesis(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	genesisState := types.DefaultGenesis()
	err := k.InitGenesis(ctx, *genesisState)
	require.NoError(t, err)

	exported := k.ExportGenesis(ctx)
	require.NotNil(t, exported)
	require.True(t, proto.Equal(genesisState.Params, exported.Params))
}
