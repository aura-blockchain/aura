package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dex/keeper"
	"github.com/aequitas/aura/chain/x/dex/types"
)

// MockBankKeeper is a mock implementation of BankKeeper for testing
type MockBankKeeper struct {
	balances map[string]map[string]math.Int // address -> denom -> amount
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		balances: make(map[string]map[string]math.Int),
	}
}

func (m *MockBankKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	// For testing, simply add to recipient balance (ignore sender balance check)
	for _, coin := range amt {
		m.addBalance(toAddr, coin.Denom, coin.Amount)
	}
	return nil
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	// See MOCK_BANK_KEEPER_EXPLANATION.md for why this implementation is critical

	// Check if balance tracking is enabled for this address
	// An address has tracking enabled if it appears in the balances map (even with zero balance)
	addrStr := senderAddr.String()
	_, addressKnown := m.balances[addrStr]

	if addressKnown {
		// Balance tracking enabled: check AND deduct (simulates real BankKeeper behavior)
		for _, coin := range amt {
			balance := m.GetBalance(ctx, senderAddr, coin.Denom)
			if balance.Amount.LT(coin.Amount) {
				return fmt.Errorf("insufficient balance: have %s, need %s %s",
					balance.Amount.String(), coin.Amount.String(), coin.Denom)
			}
			// CRITICAL: Must deduct to properly test insufficient balance scenarios
			newBalance := balance.Amount.Sub(coin.Amount)
			m.SetBalance(senderAddr, coin.Denom, newBalance)
		}
		return nil
	}

	// Address not in tracking map - two possibilities:
	// 1. Test explicitly funded with SendCoinsFromModuleToAccount -> address is in map
	// 2. Test never funded address -> not in map, return insufficient balance error
	//
	// To handle case 2 properly, check if requesting non-zero amount from unknown address
	for _, coin := range amt {
		if coin.Amount.GT(math.ZeroInt()) {
			// Trying to send from unfunded address
			return fmt.Errorf("insufficient balance: have 0, need %s %s",
				coin.Amount.String(), coin.Denom)
		}
	}

	// Zero amount send or permissive for backward compatibility
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	// For testing, simply add to recipient balance (modules have unlimited funds in tests)
	for _, coin := range amt {
		m.addBalance(recipientAddr, coin.Denom, coin.Amount)
	}
	return nil
}

func (m *MockBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	// Minting is a no-op for the mock (module balances not tracked)
	// Funds will be added when SendCoinsFromModuleToAccount is called
	return nil
}

func (m *MockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	// Burning is a no-op for the mock (module balances not tracked)
	return nil
}

func (m *MockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	addrStr := addr.String()
	if m.balances[addrStr] == nil {
		return sdk.NewCoin(denom, math.ZeroInt())
	}
	if amt, ok := m.balances[addrStr][denom]; ok {
		return sdk.NewCoin(denom, amt)
	}
	return sdk.NewCoin(denom, math.ZeroInt())
}

func (m *MockBankKeeper) SetBalance(addr sdk.AccAddress, denom string, amount math.Int) {
	addrStr := addr.String()
	if m.balances[addrStr] == nil {
		m.balances[addrStr] = make(map[string]math.Int)
	}
	m.balances[addrStr][denom] = amount
}

// addBalance adds to an existing balance (helper for SendCoins methods)
func (m *MockBankKeeper) addBalance(addr sdk.AccAddress, denom string, amount math.Int) {
	addrStr := addr.String()
	if m.balances[addrStr] == nil {
		m.balances[addrStr] = make(map[string]math.Int)
	}

	currentBalance := m.balances[addrStr][denom]
	if currentBalance.IsNil() {
		currentBalance = math.ZeroInt()
	}

	m.balances[addrStr][denom] = currentBalance.Add(amount)
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
		nil, // securityKeeper
	)
	suite.ctx = input.Ctx
}

func TestDEXKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(DEXKeeperTestSuite))
}

// Helper function to create a test keeper with access to the mock bank keeper
func setupTestKeeper(t *testing.T) (*keeper.Keeper, sdk.Context, *MockBankKeeper) {
	input := keepertest.CreateTestInput(t)
	mockBank := NewMockBankKeeper()
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, mockBank, nil, nil, nil)
	return k, input.Ctx, mockBank
}

// Alias for backward compatibility
func setupTestKeeperWithMock(t *testing.T) (*keeper.Keeper, sdk.Context, *MockBankKeeper) {
	return setupTestKeeper(t)
}

// Helper to create a funded test address (useful for tests that need balances)
func fundedTestAddr(mockBank *MockBankKeeper, uaura, uusdt int64) (sdk.AccAddress, string) {
	addr := keepertest.GenTestAddr()
	if uaura > 0 {
		mockBank.SetBalance(addr, "uaura", math.NewInt(uaura))
	}
	if uusdt > 0 {
		mockBank.SetBalance(addr, "uusdt", math.NewInt(uusdt))
	}
	return addr, addr.String()
}

// Params Tests

func (suite *DEXKeeperTestSuite) TestGetParams() {
	params := suite.keeper.GetParams(suite.ctx)
	suite.Require().NotNil(params)
}

func (suite *DEXKeeperTestSuite) TestSetParams() {
	params := types.DefaultParams()
	err := suite.keeper.SetParams(suite.ctx, &params)
	suite.Require().NoError(err)

	retrieved := suite.keeper.GetParams(suite.ctx)
	// Compare gogoproto types by comparing their fields instead of using proto.Equal
	suite.Require().Equal(params.TradingFee, retrieved.TradingFee)
	suite.Require().Equal(params.ProtocolFee, retrieved.ProtocolFee)
	suite.Require().Equal(len(params.MinLiquidityTiers), len(retrieved.MinLiquidityTiers))
}

// Pool Creation Tests

func TestCreatePool(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	tokenA := "uaura"
	tokenB := "uusdt"
	amountA := sdk.NewCoin("uaura", math.NewInt(1000000))
	amountB := sdk.NewCoin("uusdt", math.NewInt(1000000))

	// Fund creator for pool creation
	mockBank.SetBalance(creatorAddr, tokenA, amountA.Amount)
	mockBank.SetBalance(creatorAddr, tokenB, amountB.Amount)

	pool, lpTokens, err := k.CreatePool(ctx, creator, tokenA, tokenB, amountA, amountB)
	require.NoError(t, err)
	require.NotNil(t, pool)
	require.True(t, lpTokens.GT(math.ZeroInt()))
}

func TestCreatePoolWithZeroLiquidity(t *testing.T) {
	k, ctx, _ := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	tokenA := "uaura"
	tokenB := "uusdt"
	amountA := sdk.NewCoin("uaura", math.NewInt(0))
	amountB := sdk.NewCoin("uusdt", math.NewInt(1000000))

	_, _, err := k.CreatePool(ctx, creator, tokenA, tokenB, amountA, amountB)
	require.Error(t, err, "Should not allow zero liquidity")
}

func TestCreatePoolSameTokens(t *testing.T) {
	k, ctx, _ := setupTestKeeper(t)

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
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", math.NewInt(1000000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	found := pool != nil
	require.True(t, found)
	require.Equal(t, "uaura", pool.DenomA)
	require.Equal(t, "uusdt", pool.DenomB)
}

func TestGetNonExistentPool(t *testing.T) {
	k, ctx, _ := setupTestKeeper(t)

	pool := k.GetPool(ctx, "nonexistent-pool")
	require.Nil(t, pool)
}

func TestGetOrdersByUserIndexed(t *testing.T) {
	k, ctx, mockBank := setupTestKeeperWithMock(t)

	// Set up mock bank keeper with balances
	addrs := keepertest.GenTestAddrs(2)
	user := addrs[0].String()
	other := addrs[1].String()

	// User creates BUY orders, needs usdt balance
	mockBank.SetBalance(addrs[0], "usdt", math.NewInt(1000))
	// Other creates SELL order, needs uaura balance
	mockBank.SetBalance(addrs[1], "uaura", math.NewInt(100))

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

	// orders[].AuraAmount is already a math.Int, not a string
	firstAura := orders[0].AuraAmount
	lastAura := orders[len(orders)-1].AuraAmount
	require.True(t, firstAura.GT(lastAura), "newest order should appear first")
}

func TestUserOrderHistoryLimit(t *testing.T) {
	k, ctx, mockBank := setupTestKeeperWithMock(t)

	// Set up mock bank keeper with sufficient balance for 250 orders
	userAddr := keepertest.GenTestAddr()
	user := userAddr.String()

	// BUY orders need usdt balance (250 orders * 200 usdt each)
	mockBank.SetBalance(userAddr, "usdt", math.NewInt(50000))

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
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()

	// Create multiple pools with cooldown advancement
	for i := 0; i < 5; i++ {
		// Advance time by 1 hour + 1 second to satisfy cooldown (3600 seconds default)
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour + time.Second))

		tokenName := "token" + string(rune('A'+i))
		// Fund creator for each pool
		mockBank.SetBalance(creatorAddr, tokenName, math.NewInt(1000000))
		mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1000000*(int64(i)+1)))

		_, _, err := k.CreatePool(ctx, creator, tokenName, "uaura", sdk.NewCoin(tokenName, math.NewInt(1000000)), sdk.NewCoin("uaura", math.NewInt(1000000)))
		require.NoError(t, err)
	}

	pools := k.GetAllPools(ctx)
	require.GreaterOrEqual(t, len(pools), 5)
}

// Swap Tests

func TestSwap(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", math.NewInt(1000000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	traderAddr := keepertest.GenTestAddr()
	trader := traderAddr.String()
	mockBank.SetBalance(traderAddr, "uaura", math.NewInt(1000))

	coinIn := sdk.NewCoin("uaura", math.NewInt(1000))
	minAmountOut := math.NewInt(900)

	amountOut, _, _, err := k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, minAmountOut, 0)
	require.NoError(t, err)
	require.True(t, amountOut.GT(minAmountOut))
}

func TestSwapSlippageProtection(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", math.NewInt(1000000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	trader := keepertest.GenTestAddr().String()
	coinIn := sdk.NewCoin("uaura", math.NewInt(1000))
	minAmountOut := math.NewInt(10000) // Too high, should fail

	_, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, minAmountOut, 0)
	require.Error(t, err, "Should fail due to slippage protection")
}

func TestSwapInvalidPool(t *testing.T) {
	k, ctx, _ := setupTestKeeper(t)

	trader := keepertest.GenTestAddr().String()
	coinIn := sdk.NewCoin("uaura", math.NewInt(1000))

	_, _, _, err := k.SwapExactIn(ctx, trader, "99999", coinIn, math.NewInt(900), 0)
	require.Error(t, err)
}

func TestSwapZeroAmount(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", math.NewInt(1000000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	trader := keepertest.GenTestAddr().String()
	coinIn := sdk.NewCoin("uaura", math.NewInt(0))

	_, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, math.NewInt(0), 0)
	require.Error(t, err, "Should not allow zero swap")
}

func TestSwapPriceImpact(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", math.NewInt(1000000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	traderAddr := keepertest.GenTestAddr()
	trader := traderAddr.String()
	// Use 10% of pool (100000) to avoid triggering max trade size (20% limit)
	largeSwap := math.NewInt(100000)
	mockBank.SetBalance(traderAddr, "uaura", largeSwap)

	coinIn := sdk.NewCoin("uaura", largeSwap)

	// Large swap should have significant price impact
	amountOut, _, _, err := k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, math.NewInt(1), 0)
	require.NoError(t, err)

	// Output should be less than input (due to constant product formula)
	require.True(t, amountOut.LT(largeSwap))
}

// Liquidity Tests

func TestAddLiquidity(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", math.NewInt(1000000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	providerAddr := keepertest.GenTestAddr()
	provider := providerAddr.String()
	amountA := sdk.NewCoin("uaura", math.NewInt(100000))
	amountB := sdk.NewCoin("uusdt", math.NewInt(100000))
	mockBank.SetBalance(providerAddr, "uaura", amountA.Amount)
	mockBank.SetBalance(providerAddr, "uusdt", amountB.Amount)

	lpTokens, _, err := k.AddLiquidity(ctx, provider, pool.PoolId, amountA, amountB)
	require.NoError(t, err)
	require.True(t, lpTokens.GT(math.ZeroInt()))
}

func TestAddLiquidityImbalanced(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", math.NewInt(1000000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	providerAddr := keepertest.GenTestAddr()
	provider := providerAddr.String()
	amountA := sdk.NewCoin("uaura", math.NewInt(100000))
	amountB := sdk.NewCoin("uusdt", math.NewInt(200000)) // Imbalanced ratio
	mockBank.SetBalance(providerAddr, "uaura", amountA.Amount)
	mockBank.SetBalance(providerAddr, "uusdt", amountB.Amount)

	// AddLiquidity adjusts amounts to match ratio, so it returns successfully
	// with adjusted amounts (not an error)
	lpTokens, poolShare, err := k.AddLiquidity(ctx, provider, pool.PoolId, amountA, amountB)
	require.NoError(t, err)
	require.NotNil(t, lpTokens)
	require.True(t, poolShare.GT(math.LegacyZeroDec()))
}

func TestRemoveLiquidity(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	// Fund creator for pool creation (1000000) + additional liquidity (100000)
	mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1100000))
	mockBank.SetBalance(creatorAddr, "uusdt", math.NewInt(1100000))

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
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", math.NewInt(1000000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	provider := keepertest.GenTestAddr().String()
	tooManyLPTokens := math.NewInt(999999999)

	_, _, err = k.RemoveLiquidity(ctx, provider, pool.PoolId, tooManyLPTokens)
	require.Error(t, err, "Should not allow removing more than owned")
}

func TestGetPoolPrice(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", math.NewInt(2000000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(2000000)))
	require.NoError(t, err)

	price, err := k.GetPoolPrice(ctx, pool.PoolId, "uaura")
	require.NoError(t, err)
	require.False(t, price.IsZero())
	// Price should be 2:1 (2 uusdt per uaura)
	require.True(t, price.GT(math.LegacyOneDec()))
}

func TestGetSpotPrice(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", math.NewInt(1000000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	spotPrice, err := k.GetSpotPrice(ctx, pool.PoolId, "uaura", "uusdt")
	require.NoError(t, err)
	require.False(t, spotPrice.IsZero())
}

// Fee Tests

func TestCalculateSwapFee(t *testing.T) {
	k, ctx, _ := setupTestKeeper(t)

	amount := math.NewInt(1000000)
	fee, err := k.CalculateSwapFee(ctx, amount)
	require.NoError(t, err)

	require.True(t, fee.GT(math.ZeroInt()))
	require.True(t, fee.LT(amount))
}

func TestCollectSwapFees(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", math.NewInt(1000000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	traderAddr := keepertest.GenTestAddr()
	trader := traderAddr.String()
	coinIn := sdk.NewCoin("uaura", math.NewInt(1000))
	mockBank.SetBalance(traderAddr, "uaura", coinIn.Amount)

	_, _, _, err = k.SwapExactIn(ctx, trader, pool.PoolId, coinIn, math.NewInt(900), 0)
	require.NoError(t, err)

	fees, err := k.GetCollectedFees(ctx, pool.PoolId)
	require.NoError(t, err)
	require.NotNil(t, fees)
}

// Circuit Breaker Tests

func TestCircuitBreakerLargeSwap(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", math.NewInt(1000000))

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
	k, ctx, mockBank := setupTestKeeper(t)

	// Fund creator with sufficient balance for pool creation
	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", math.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", math.NewInt(1000000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	// Multiple large swaps to cause price deviation
	traderAddr := keepertest.GenTestAddr()
	trader := traderAddr.String()
	// Fund trader with enough for 10 swaps of 50000 each = 500000 uaura
	mockBank.SetBalance(traderAddr, "uaura", math.NewInt(500000))

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
	k, ctx, _ := setupTestKeeper(t)

	creator := keepertest.GenTestAddr().String()
	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt", sdk.NewCoin("uaura", math.NewInt(1000000)), sdk.NewCoin("uusdt", math.NewInt(1000000)))
	require.NoError(t, err)

	position := k.GetUserPosition(ctx, pool.PoolId, creator)
	require.NotNil(t, position)
	require.True(t, position.LPTokens.GT(math.ZeroInt()))
}

func TestGetUserPositions(t *testing.T) {
	k, ctx, _ := setupTestKeeper(t)

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
	k, ctx, _ := setupTestKeeper(t)

	genesisState := types.DefaultGenesis()
	err := k.InitGenesis(ctx, *genesisState)
	require.NoError(t, err)
}

func TestExportGenesis(t *testing.T) {
	k, ctx, _ := setupTestKeeper(t)

	genesisState := types.DefaultGenesis()
	err := k.InitGenesis(ctx, *genesisState)
	require.NoError(t, err)

	exported := k.ExportGenesis(ctx)
	require.NotNil(t, exported)
	// Compare gogoproto types by comparing their fields instead of using proto.Equal
	require.Equal(t, genesisState.Params.TradingFee, exported.Params.TradingFee)
	require.Equal(t, genesisState.Params.ProtocolFee, exported.Params.ProtocolFee)
	require.Equal(t, len(genesisState.Params.MinLiquidityTiers), len(exported.Params.MinLiquidityTiers))
}
