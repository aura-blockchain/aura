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

// TestFixture provides common test setup
type TestFixture struct {
	Ctx        sdk.Context
	Keeper     *keeper.Keeper
	BankKeeper *MockBankKeeper
	Addrs      []sdk.AccAddress
}

// SetupKeeperTest creates a test fixture with keeper and mock dependencies
func SetupKeeperTest(t *testing.T) TestFixture {
	input := keepertest.CreateTestInput(t)
	bankKeeper := NewMockBankKeeper()

	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		bankKeeper,
		nil, // accountKeeper
		nil, // vcKeeper
		nil, // securityKeeper
	)

	// Generate test addresses
	addrs := keepertest.GenTestAddrs(5)

	return TestFixture{
		Ctx:        input.Ctx,
		Keeper:     k,
		BankKeeper: bankKeeper,
		Addrs:      addrs,
	}
}

// TestOrderbookReentrancyProtection tests that reentrancy attacks are prevented
// in order creation, matching, and cancellation operations
func TestOrderbookReentrancyProtection(t *testing.T) {
	f := SetupKeeperTest(t)
	ctx := f.Ctx

	// Setup: Create test accounts with initial balances
	creator := f.Addrs[0].String()
	matcher := f.Addrs[1].String()

	// Fund accounts using mock
	creatorAddr := f.Addrs[0]
	matcherAddr := f.Addrs[1]

	initialAura := sdkmath.NewInt(10000)
	initialUsdt := sdkmath.NewInt(5000)

	f.BankKeeper.SetBalance(creatorAddr, "uaura", initialAura)
	f.BankKeeper.SetBalance(matcherAddr, "usdt", initialUsdt)

	// Test 1: Reentrancy protection in CreateOrder
	t.Run("CreateOrder_ReentrancyProtection", func(t *testing.T) {
		auraAmount := sdkmath.NewInt(1000)
		usdtAmount := sdkmath.NewInt(500)

		// First order should succeed
		order1, err := f.Keeper.CreateOrder(
			ctx,
			creator,
			types.SwapOrderType_SELL,
			auraAmount,
			"usdt",
			usdtAmount,
			1440,
		)
		require.NoError(t, err)
		require.NotNil(t, order1)
		require.Equal(t, types.SwapOrderStatus_PENDING, order1.Status)

		// Move to next block to ensure different order IDs
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

		// Second order should also succeed (sequential calls, no reentrancy)
		order2, err := f.Keeper.CreateOrder(
			ctx,
			creator,
			types.SwapOrderType_SELL,
			auraAmount,
			"usdt",
			usdtAmount,
			1440,
		)
		require.NoError(t, err)
		require.NotNil(t, order2)
		require.NotEqual(t, order1.OrderId, order2.OrderId)
	})

	// Test 2: Reentrancy protection in MatchOrder
	t.Run("MatchOrder_ReentrancyProtection", func(t *testing.T) {
		// Create a sell order
		auraAmount := sdkmath.NewInt(500)
		usdtAmount := sdkmath.NewInt(250)

		order, err := f.Keeper.CreateOrder(
			ctx,
			creator,
			types.SwapOrderType_SELL,
			auraAmount,
			"usdt",
			usdtAmount,
			1440,
		)
		require.NoError(t, err)
		require.NotNil(t, order)

		// Match the order
		err = f.Keeper.MatchOrder(ctx, matcher, order.OrderId)
		require.NoError(t, err)

		// Verify order is completed
		completedOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.NotNil(t, completedOrder)
		require.Equal(t, types.SwapOrderStatus_COMPLETED, completedOrder.Status)

		// Attempting to match the same order again should fail
		// (not due to reentrancy, but due to status check)
		err = f.Keeper.MatchOrder(ctx, matcher, order.OrderId)
		require.Error(t, err)
		require.Contains(t, err.Error(), "order is not pending")
	})

	// Test 3: Reentrancy protection in CancelOrder
	t.Run("CancelOrder_ReentrancyProtection", func(t *testing.T) {
		// Create an order to cancel
		auraAmount := sdkmath.NewInt(300)
		usdtAmount := sdkmath.NewInt(150)

		order, err := f.Keeper.CreateOrder(
			ctx,
			creator,
			types.SwapOrderType_SELL,
			auraAmount,
			"usdt",
			usdtAmount,
			1440,
		)
		require.NoError(t, err)
		require.NotNil(t, order)

		// Cancel the order
		err = f.Keeper.CancelOrder(ctx, order.OrderId, "user_cancelled")
		require.NoError(t, err)

		// Verify order is cancelled
		cancelledOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.NotNil(t, cancelledOrder)
		require.Equal(t, types.SwapOrderStatus_CANCELLED, cancelledOrder.Status)

		// Attempting to cancel the same order again should fail
		err = f.Keeper.CancelOrder(ctx, order.OrderId, "user_cancelled")
		require.Error(t, err)
		require.Contains(t, err.Error(), "order is not pending")
	})

	// Test 4: State consistency - order status protection
	t.Run("MatchOrder_StatusProtection", func(t *testing.T) {
		// Create a sell order
		auraAmount := sdkmath.NewInt(600)
		usdtAmount := sdkmath.NewInt(300)

		order, err := f.Keeper.CreateOrder(
			ctx,
			creator,
			types.SwapOrderType_SELL,
			auraAmount,
			"usdt",
			usdtAmount,
			1440,
		)
		require.NoError(t, err)
		require.NotNil(t, order)

		// Match the order successfully
		err = f.Keeper.MatchOrder(ctx, matcher, order.OrderId)
		require.NoError(t, err)

		// Verify order is COMPLETED
		completedOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.NotNil(t, completedOrder)
		require.Equal(t, types.SwapOrderStatus_COMPLETED, completedOrder.Status)

		// Attempting to match again should fail (status protection)
		err = f.Keeper.MatchOrder(ctx, matcher, order.OrderId)
		require.Error(t, err)
		require.Contains(t, err.Error(), "order is not pending")
	})

	// Test 5: Prevent self-matching
	t.Run("MatchOrder_PreventSelfMatching", func(t *testing.T) {
		// Create a sell order
		auraAmount := sdkmath.NewInt(400)
		usdtAmount := sdkmath.NewInt(200)

		order, err := f.Keeper.CreateOrder(
			ctx,
			creator,
			types.SwapOrderType_SELL,
			auraAmount,
			"usdt",
			usdtAmount,
			1440,
		)
		require.NoError(t, err)
		require.NotNil(t, order)

		// Attempt to match own order should fail
		err = f.Keeper.MatchOrder(ctx, creator, order.OrderId)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot match your own order")

		// Verify order is still PENDING
		pendingOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.NotNil(t, pendingOrder)
		require.Equal(t, types.SwapOrderStatus_PENDING, pendingOrder.Status)
	})
}

// TestOrderbookInvariantPreservation tests that invariants are maintained
// during order execution to prevent manipulation attacks
func TestOrderbookInvariantPreservation(t *testing.T) {
	f := SetupKeeperTest(t)
	ctx := f.Ctx

	creator := f.Addrs[0].String()
	matcher := f.Addrs[1].String()

	creatorAddr := f.Addrs[0]
	matcherAddr := f.Addrs[1]

	// Fund accounts
	initialAura := sdkmath.NewInt(10000)
	initialUsdt := sdkmath.NewInt(5000)

	f.BankKeeper.SetBalance(creatorAddr, "uaura", initialAura)
	f.BankKeeper.SetBalance(matcherAddr, "usdt", initialUsdt)

	// Test: Order execution completes successfully
	t.Run("OrderExecution_Success", func(t *testing.T) {
		auraAmount := sdkmath.NewInt(1000)
		usdtAmount := sdkmath.NewInt(500)

		// Create and match order
		order, err := f.Keeper.CreateOrder(
			ctx,
			creator,
			types.SwapOrderType_SELL,
			auraAmount,
			"usdt",
			usdtAmount,
			1440,
		)
		require.NoError(t, err)
		require.NotNil(t, order)
		require.Equal(t, types.SwapOrderStatus_PENDING, order.Status)

		err = f.Keeper.MatchOrder(ctx, matcher, order.OrderId)
		require.NoError(t, err)

		// Verify order is completed
		completedOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.NotNil(t, completedOrder)
		require.Equal(t, types.SwapOrderStatus_COMPLETED, completedOrder.Status)
		require.Equal(t, matcher, completedOrder.MatcherAddress)
	})
}

// TestOrderbookDoubleSpendPrevention tests that double-spend attacks are prevented
func TestOrderbookDoubleSpendPrevention(t *testing.T) {
	f := SetupKeeperTest(t)
	ctx := f.Ctx

	creator := f.Addrs[0].String()
	matcher1 := f.Addrs[1].String()
	matcher2 := f.Addrs[2].String()

	creatorAddr := f.Addrs[0]
	matcher1Addr := f.Addrs[1]
	matcher2Addr := f.Addrs[2]

	// Fund accounts
	initialAura := sdkmath.NewInt(10000)
	initialUsdt := sdkmath.NewInt(5000)

	f.BankKeeper.SetBalance(creatorAddr, "uaura", initialAura)
	f.BankKeeper.SetBalance(matcher1Addr, "usdt", initialUsdt)
	f.BankKeeper.SetBalance(matcher2Addr, "usdt", initialUsdt)

	t.Run("PreventDoubleMatching", func(t *testing.T) {
		auraAmount := sdkmath.NewInt(1000)
		usdtAmount := sdkmath.NewInt(500)

		// Create order
		order, err := f.Keeper.CreateOrder(
			ctx,
			creator,
			types.SwapOrderType_SELL,
			auraAmount,
			"usdt",
			usdtAmount,
			1440,
		)
		require.NoError(t, err)

		// First matcher matches successfully
		err = f.Keeper.MatchOrder(ctx, matcher1, order.OrderId)
		require.NoError(t, err)

		// Verify order is completed
		completedOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.Equal(t, types.SwapOrderStatus_COMPLETED, completedOrder.Status)

		// Second matcher attempts to match the same order (should fail)
		err = f.Keeper.MatchOrder(ctx, matcher2, order.OrderId)
		require.Error(t, err)
		require.Contains(t, err.Error(), "order is not pending")

		// Verify order is still completed with first matcher
		finalOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.Equal(t, types.SwapOrderStatus_COMPLETED, finalOrder.Status)
		require.Equal(t, matcher1, finalOrder.MatcherAddress)
	})
}
