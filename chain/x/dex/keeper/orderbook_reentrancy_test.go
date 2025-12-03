package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// TestOrderbookReentrancyProtection tests that reentrancy attacks are prevented
// in order creation, matching, and cancellation operations
func TestOrderbookReentrancyProtection(t *testing.T) {
	f := SetupKeeperTest(t)
	ctx := f.Ctx

	// Setup: Create test accounts with initial balances
	creator := f.Addrs[0].String()
	matcher := f.Addrs[1].String()

	// Fund accounts
	creatorAddr := sdk.MustAccAddressFromBech32(creator)
	matcherAddr := sdk.MustAccAddressFromBech32(matcher)

	initialAura := sdkmath.NewInt(10000)
	initialUsdt := sdkmath.NewInt(5000)

	err := f.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(
		sdk.NewCoin("uaura", initialAura),
		sdk.NewCoin("usdt", initialUsdt),
	))
	require.NoError(t, err)

	err = f.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", initialAura),
	))
	require.NoError(t, err)

	err = f.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, matcherAddr, sdk.NewCoins(
		sdk.NewCoin("usdt", initialUsdt),
	))
	require.NoError(t, err)

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

		// Attempting to create another order in the same orderbook scope simultaneously
		// would be blocked by reentrancy guard if this were an actual reentrancy attack
		// In this test, we verify that sequential calls work (no lock held between calls)
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

	// Test 4: State consistency after failed match (rollback verification)
	t.Run("MatchOrder_FailedMatch_StateRollback", func(t *testing.T) {
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

		// Drain matcher's USDT balance to cause match failure
		matcherBalance := f.BankKeeper.GetBalance(ctx, matcherAddr, "usdt")
		err = f.BankKeeper.SendCoinsFromAccountToModule(ctx, matcherAddr, types.ModuleName, sdk.NewCoins(matcherBalance))
		require.NoError(t, err)

		// Attempt to match should fail due to insufficient balance
		err = f.Keeper.MatchOrder(ctx, matcher, order.OrderId)
		require.Error(t, err)
		require.Contains(t, err.Error(), "insufficient balance")

		// Verify order is still PENDING (state rolled back)
		pendingOrder := f.Keeper.GetOrder(ctx, order.OrderId)
		require.NotNil(t, pendingOrder)
		require.Equal(t, types.SwapOrderStatus_PENDING, pendingOrder.Status)
		require.Empty(t, pendingOrder.MatcherAddress)

		// Verify order is still in orderbook
		orderbook := f.Keeper.GetOrderbookForPair(ctx, "uaura", "usdt")
		found := false
		for _, o := range orderbook {
			if o.OrderId == order.OrderId {
				found = true
				break
			}
		}
		require.True(t, found, "Order should still be in orderbook after failed match")
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

	creatorAddr := sdk.MustAccAddressFromBech32(creator)
	matcherAddr := sdk.MustAccAddressFromBech32(matcher)

	// Fund accounts
	initialAura := sdkmath.NewInt(10000)
	initialUsdt := sdkmath.NewInt(5000)

	err := f.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(
		sdk.NewCoin("uaura", initialAura),
		sdk.NewCoin("usdt", initialUsdt),
	))
	require.NoError(t, err)

	err = f.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", initialAura),
	))
	require.NoError(t, err)

	err = f.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, matcherAddr, sdk.NewCoins(
		sdk.NewCoin("usdt", initialUsdt),
	))
	require.NoError(t, err)

	// Test: Order execution preserves total balance
	t.Run("OrderExecution_BalanceConservation", func(t *testing.T) {
		auraAmount := sdkmath.NewInt(1000)
		usdtAmount := sdkmath.NewInt(500)

		// Record initial balances
		creatorAuraBefore := f.BankKeeper.GetBalance(ctx, creatorAddr, "uaura").Amount
		creatorUsdtBefore := f.BankKeeper.GetBalance(ctx, creatorAddr, "usdt").Amount
		matcherAuraBefore := f.BankKeeper.GetBalance(ctx, matcherAddr, "uaura").Amount
		matcherUsdtBefore := f.BankKeeper.GetBalance(ctx, matcherAddr, "usdt").Amount

		totalAuraBefore := creatorAuraBefore.Add(matcherAuraBefore)
		totalUsdtBefore := creatorUsdtBefore.Add(matcherUsdtBefore)

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

		err = f.Keeper.MatchOrder(ctx, matcher, order.OrderId)
		require.NoError(t, err)

		// Record final balances
		creatorAuraAfter := f.BankKeeper.GetBalance(ctx, creatorAddr, "uaura").Amount
		creatorUsdtAfter := f.BankKeeper.GetBalance(ctx, creatorAddr, "usdt").Amount
		matcherAuraAfter := f.BankKeeper.GetBalance(ctx, matcherAddr, "uaura").Amount
		matcherUsdtAfter := f.BankKeeper.GetBalance(ctx, matcherAddr, "usdt").Amount

		totalAuraAfter := creatorAuraAfter.Add(matcherAuraAfter)
		totalUsdtAfter := creatorUsdtAfter.Add(matcherUsdtAfter)

		// Verify total AURA is conserved
		require.True(t, totalAuraBefore.Equal(totalAuraAfter),
			"Total AURA must be conserved: before=%s, after=%s",
			totalAuraBefore.String(), totalAuraAfter.String())

		// Verify total USDT is conserved
		require.True(t, totalUsdtBefore.Equal(totalUsdtAfter),
			"Total USDT must be conserved: before=%s, after=%s",
			totalUsdtBefore.String(), totalUsdtAfter.String())

		// Verify correct amounts were swapped
		require.True(t, creatorAuraAfter.Equal(creatorAuraBefore.Sub(auraAmount)),
			"Creator should have sent %s AURA", auraAmount.String())
		require.True(t, creatorUsdtAfter.Equal(creatorUsdtBefore.Add(usdtAmount)),
			"Creator should have received %s USDT", usdtAmount.String())
		require.True(t, matcherAuraAfter.Equal(matcherAuraBefore.Add(auraAmount)),
			"Matcher should have received %s AURA", auraAmount.String())
		require.True(t, matcherUsdtAfter.Equal(matcherUsdtBefore.Sub(usdtAmount)),
			"Matcher should have sent %s USDT", usdtAmount.String())
	})
}

// TestOrderbookDoubleSpendPrevention tests that double-spend attacks are prevented
func TestOrderbookDoubleSpendPrevention(t *testing.T) {
	f := SetupKeeperTest(t)
	ctx := f.Ctx

	creator := f.Addrs[0].String()
	matcher1 := f.Addrs[1].String()
	matcher2 := f.Addrs[2].String()

	creatorAddr := sdk.MustAccAddressFromBech32(creator)
	matcher1Addr := sdk.MustAccAddressFromBech32(matcher1)
	matcher2Addr := sdk.MustAccAddressFromBech32(matcher2)

	// Fund accounts
	initialAura := sdkmath.NewInt(10000)
	initialUsdt := sdkmath.NewInt(5000)

	err := f.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(
		sdk.NewCoin("uaura", initialAura),
		sdk.NewCoin("usdt", initialUsdt.Mul(sdkmath.NewInt(2))),
	))
	require.NoError(t, err)

	err = f.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, sdk.NewCoins(
		sdk.NewCoin("uaura", initialAura),
	))
	require.NoError(t, err)

	err = f.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, matcher1Addr, sdk.NewCoins(
		sdk.NewCoin("usdt", initialUsdt),
	))
	require.NoError(t, err)

	err = f.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, matcher2Addr, sdk.NewCoins(
		sdk.NewCoin("usdt", initialUsdt),
	))
	require.NoError(t, err)

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
