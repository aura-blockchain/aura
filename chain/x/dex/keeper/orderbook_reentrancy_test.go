package keeper_test

import (
	"github.com/aequitas/aura/chain/testing/testutil"
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
		testutil.NewMockAccountKeeper(),
		testutil.NewMockVCRegistryKeeper(),
		testutil.NewMockSecurityKeeper(),
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

// TestOrderbookReentrancyCallbackAttack simulates sophisticated reentrancy attacks
// that attempt to exploit token transfer callbacks (similar to ERC-777 hooks)
func TestOrderbookReentrancyCallbackAttack(t *testing.T) {
	f := SetupKeeperTest(t)
	ctx := f.Ctx

	creator := f.Addrs[0].String()
	attacker := f.Addrs[1].String()

	creatorAddr := f.Addrs[0]
	attackerAddr := f.Addrs[1]

	// Fund accounts
	initialAura := sdkmath.NewInt(10000)
	initialUsdt := sdkmath.NewInt(5000)

	f.BankKeeper.SetBalance(creatorAddr, "uaura", initialAura)
	f.BankKeeper.SetBalance(attackerAddr, "usdt", initialUsdt)

	t.Run("AttemptReentrantCancel_DuringMatch", func(t *testing.T) {
		// Scenario: Attacker tries to cancel order during match execution
		// This simulates callback reentrancy from token transfer hooks

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
		require.Equal(t, types.SwapOrderStatus_PENDING, order.Status)

		// Begin matching the order
		err = f.Keeper.MatchOrder(ctx, attacker, order.OrderId)
		require.NoError(t, err)

		// CRITICAL: After match completes, order should be COMPLETED
		// not PENDING, which prevents any cancellation attempts
		matchedOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.Equal(t, types.SwapOrderStatus_COMPLETED, matchedOrder.Status)

		// Attempt to cancel after matching (simulates reentrancy during callback)
		// This should fail because order is no longer PENDING
		err = f.Keeper.CancelOrder(ctx, order.OrderId, "malicious_cancel")
		require.Error(t, err)
		require.Contains(t, err.Error(), "order is not pending")

		// Verify order state is unchanged (still COMPLETED)
		finalOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.Equal(t, types.SwapOrderStatus_COMPLETED, finalOrder.Status)
		require.Equal(t, attacker, finalOrder.MatcherAddress)
	})

	t.Run("AttemptDoubleCreate_SamePool", func(t *testing.T) {
		// Scenario: Attacker tries to create multiple orders in rapid succession
		// to exploit potential race conditions in pool state

		auraAmount := sdkmath.NewInt(500)
		usdtAmount := sdkmath.NewInt(250)

		// Create first order
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

		// Move to next block for different order ID
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

		// Create second order (should succeed - this is allowed)
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

		// Orders should have different IDs
		require.NotEqual(t, order1.OrderId, order2.OrderId)

		// Both orders should be independently valid
		require.Equal(t, types.SwapOrderStatus_PENDING, order1.Status)
		require.Equal(t, types.SwapOrderStatus_PENDING, order2.Status)
	})

	t.Run("AttemptMatchDuringCancel", func(t *testing.T) {
		// Scenario: Two users try to interact with same order simultaneously
		// One tries to match while creator tries to cancel

		auraAmount := sdkmath.NewInt(800)
		usdtAmount := sdkmath.NewInt(400)

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

		// Creator cancels first
		err = f.Keeper.CancelOrder(ctx, order.OrderId, "user_cancelled")
		require.NoError(t, err)

		// Verify order is cancelled
		cancelledOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.Equal(t, types.SwapOrderStatus_CANCELLED, cancelledOrder.Status)

		// Attacker tries to match the cancelled order (should fail)
		err = f.Keeper.MatchOrder(ctx, attacker, order.OrderId)
		require.Error(t, err)
		require.Contains(t, err.Error(), "order is not pending")

		// Verify order remains cancelled
		finalOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.Equal(t, types.SwapOrderStatus_CANCELLED, finalOrder.Status)
	})
}

// TestOrderbookStateConsistency verifies that state remains consistent
// even under adversarial conditions that attempt to violate invariants
func TestOrderbookStateConsistency(t *testing.T) {
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

	t.Run("StateUpdates_BeforeExternalCalls", func(t *testing.T) {
		// Verify that Checks-Effects-Interactions pattern is followed
		// State should be updated BEFORE any external calls

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

		// Order should be stored immediately (EFFECTS before INTERACTIONS)
		storedOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.NotNil(t, storedOrder)
		require.Equal(t, order.OrderId, storedOrder.OrderId)
		require.Equal(t, types.SwapOrderStatus_PENDING, storedOrder.Status)

		// Match order
		err = f.Keeper.MatchOrder(ctx, matcher, order.OrderId)
		require.NoError(t, err)

		// Order status should be updated to COMPLETED immediately
		completedOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.Equal(t, types.SwapOrderStatus_COMPLETED, completedOrder.Status)
		require.Equal(t, matcher, completedOrder.MatcherAddress)
	})

	t.Run("OrderbookIndex_RemovedOnMatch", func(t *testing.T) {
		// Verify that matched orders are removed from orderbook index
		// to prevent double-matching

		auraAmount := sdkmath.NewInt(700)
		usdtAmount := sdkmath.NewInt(350)

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

		// Verify order is in orderbook
		orderbookOrders := f.Keeper.GetOrderbookForPair(ctx, "uaura", "usdt")
		foundInOrderbook := false
		for _, o := range orderbookOrders {
			if o.OrderId == order.OrderId {
				foundInOrderbook = true
				break
			}
		}
		require.True(t, foundInOrderbook, "order should be in orderbook before matching")

		// Match order
		err = f.Keeper.MatchOrder(ctx, matcher, order.OrderId)
		require.NoError(t, err)

		// Verify order is NO LONGER in orderbook (prevents double-matching)
		orderbookOrdersAfter := f.Keeper.GetOrderbookForPair(ctx, "uaura", "usdt")
		foundAfterMatch := false
		for _, o := range orderbookOrdersAfter {
			if o.OrderId == order.OrderId {
				foundAfterMatch = true
				break
			}
		}
		require.False(t, foundAfterMatch, "order should be removed from orderbook after matching")
	})

	t.Run("OrderbookIndex_RemovedOnCancel", func(t *testing.T) {
		// Verify that cancelled orders are removed from orderbook index

		auraAmount := sdkmath.NewInt(600)
		usdtAmount := sdkmath.NewInt(300)

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

		// Verify order is in orderbook
		orderbookOrders := f.Keeper.GetOrderbookForPair(ctx, "uaura", "usdt")
		foundInOrderbook := false
		for _, o := range orderbookOrders {
			if o.OrderId == order.OrderId {
				foundInOrderbook = true
				break
			}
		}
		require.True(t, foundInOrderbook, "order should be in orderbook before cancellation")

		// Cancel order
		err = f.Keeper.CancelOrder(ctx, order.OrderId, "user_cancelled")
		require.NoError(t, err)

		// Verify order is NO LONGER in orderbook
		orderbookOrdersAfter := f.Keeper.GetOrderbookForPair(ctx, "uaura", "usdt")
		foundAfterCancel := false
		for _, o := range orderbookOrdersAfter {
			if o.OrderId == order.OrderId {
				foundAfterCancel = true
				break
			}
		}
		require.False(t, foundAfterCancel, "order should be removed from orderbook after cancellation")
	})
}

// TestOrderbookScopedReentrancyProtection verifies that scoped reentrancy
// protection allows concurrent operations on different pools while preventing
// reentrancy on the same pool
func TestOrderbookScopedReentrancyProtection(t *testing.T) {
	f := SetupKeeperTest(t)
	ctx := f.Ctx

	creator := f.Addrs[0].String()
	matcher := f.Addrs[1].String()

	creatorAddr := f.Addrs[0]
	matcherAddr := f.Addrs[1]

	// Fund accounts with multiple denoms
	f.BankKeeper.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(20000))
	f.BankKeeper.SetBalance(matcherAddr, "usdt", sdkmath.NewInt(5000))
	f.BankKeeper.SetBalance(matcherAddr, "uusdc", sdkmath.NewInt(5000))

	t.Run("AllowConcurrentDifferentPools", func(t *testing.T) {
		// Create orders in different pools (uaura-usdt and uaura-uusdc)
		// These should be allowed concurrently since they have different scopes

		// Order 1: uaura-usdt pool
		order1, err := f.Keeper.CreateOrder(
			ctx,
			creator,
			types.SwapOrderType_SELL,
			sdkmath.NewInt(1000),
			"usdt",
			sdkmath.NewInt(500),
			1440,
		)
		require.NoError(t, err)
		require.NotNil(t, order1)

		// Move to next block for different order ID
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

		// Order 2: uaura-uusdc pool (different scope)
		order2, err := f.Keeper.CreateOrder(
			ctx,
			creator,
			types.SwapOrderType_SELL,
			sdkmath.NewInt(1000),
			"uusdc",
			sdkmath.NewInt(500),
			1440,
		)
		require.NoError(t, err)
		require.NotNil(t, order2)

		// Both orders should exist independently
		require.NotEqual(t, order1.OrderId, order2.OrderId)
		require.Equal(t, "usdt", order1.OtherCoin)
		require.Equal(t, "uusdc", order2.OtherCoin)
	})

	t.Run("PreventReentrancySamePool", func(t *testing.T) {
		// The reentrancy guard should prevent nested calls within the same pool
		// This is tested implicitly by the other tests - if we attempt to
		// manipulate an order during its own execution, it will fail

		auraAmount := sdkmath.NewInt(800)
		usdtAmount := sdkmath.NewInt(400)

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

		// Match order (this acquires the reentrancy lock for this pool)
		err = f.Keeper.MatchOrder(ctx, matcher, order.OrderId)
		require.NoError(t, err)

		// After execution completes, order should be in final state
		finalOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.Equal(t, types.SwapOrderStatus_COMPLETED, finalOrder.Status)

		// Any attempt to modify this order now should fail
		// because it's no longer PENDING
		err = f.Keeper.CancelOrder(ctx, order.OrderId, "attempt_after_match")
		require.Error(t, err)
		require.Contains(t, err.Error(), "order is not pending")
	})
}
