package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================
// P2P ORDERBOOK (from simple_swap_gui.py)
// ============================

// CreateOrder creates a new swap order in the orderbook
// Adapted from simple_swap_gui.py lines 120-180
func (k Keeper) CreateOrder(
	ctx sdk.Context,
	creator string,
	orderType types.SwapOrderType,
	auraAmount sdk.Int,
	otherCoin string,
	otherAmount sdk.Int,
	expirationMinutes uint64,
) (*types.SwapOrder, error) {
	// Validate inputs
	if auraAmount.LTE(sdk.ZeroInt()) {
		return nil, fmt.Errorf("AURA amount must be positive")
	}
	if otherAmount.LTE(sdk.ZeroInt()) {
		return nil, fmt.Errorf("other coin amount must be positive")
	}

	// Generate order ID
	orderID := k.GenerateOrderID(ctx, creator)

	// Calculate expiration time
	expirationTime := ctx.BlockTime().Add(time.Duration(expirationMinutes) * time.Minute)

	// Create order
	order := &types.SwapOrder{
		OrderId:     orderID,
		OrderType:   orderType,
		AuraAmount:  auraAmount.String(),
		OtherCoin:   otherCoin,
		OtherAmount: otherAmount.String(),
		UserAddress: creator,
		Status:      types.SwapOrderStatus_PENDING,
		CreatedAt:   ctx.BlockTime(),
		ExpiresAt:   expirationTime,
	}

	// Lock user's funds based on order type
	if orderType == types.SwapOrderType_SELL {
		// Selling AURA, lock AURA
		if err := k.LockFundsForOrder(ctx, creator, "uaura", auraAmount); err != nil {
			return nil, fmt.Errorf("failed to lock AURA: %w", err)
		}
	} else {
		// Buying AURA, lock other coin
		if err := k.LockFundsForOrder(ctx, creator, otherCoin, otherAmount); err != nil {
			return nil, fmt.Errorf("failed to lock %s: %w", otherCoin, err)
		}
	}

	// Store order
	k.SetOrder(ctx, order)

	// Add to orderbook index
	k.AddToOrderbook(ctx, order)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"order_created",
			sdk.NewAttribute("order_id", orderID),
			sdk.NewAttribute("creator", creator),
			sdk.NewAttribute("type", orderType.String()),
			sdk.NewAttribute("aura_amount", auraAmount.String()),
			sdk.NewAttribute("other_coin", otherCoin),
			sdk.NewAttribute("other_amount", otherAmount.String()),
		),
	)

	return order, nil
}

// MatchOrder attempts to match and execute an order
// Adapted from simple_swap_gui.py lines 220-280
func (k Keeper) MatchOrder(
	ctx sdk.Context,
	matcher string,
	orderID string,
) error {
	// Get order
	order := k.GetOrder(ctx, orderID)
	if order == nil {
		return fmt.Errorf("order not found: %s", orderID)
	}

	// Validate order status
	if order.Status != types.SwapOrderStatus_PENDING {
		return fmt.Errorf("order is not pending: %s", order.Status.String())
	}

	// Check expiration
	if ctx.BlockTime().After(order.ExpiresAt) {
		// Order expired, cancel it
		k.CancelOrder(ctx, orderID, "expired")
		return fmt.Errorf("order has expired")
	}

	// Parse amounts
	auraAmount, ok := sdk.NewIntFromString(order.AuraAmount)
	if !ok {
		return fmt.Errorf("invalid AURA amount")
	}
	otherAmount, ok := sdk.NewIntFromString(order.OtherAmount)
	if !ok {
		return fmt.Errorf("invalid other amount")
	}

	// Lock matcher's funds
	if order.OrderType == types.SwapOrderType_SELL {
		// Order is selling AURA, matcher needs to provide other coin
		if err := k.LockFundsForOrder(ctx, matcher, order.OtherCoin, otherAmount); err != nil {
			return fmt.Errorf("failed to lock matcher funds: %w", err)
		}
	} else {
		// Order is buying AURA, matcher needs to provide AURA
		if err := k.LockFundsForOrder(ctx, matcher, "uaura", auraAmount); err != nil {
			return fmt.Errorf("failed to lock matcher funds: %w", err)
		}
	}

	// Update order status
	order.Status = types.SwapOrderStatus_MATCHED
	order.MatcherAddress = matcher
	order.MatchedAt = ctx.BlockTime()
	k.SetOrder(ctx, order)

	// Execute swap
	if err := k.ExecuteSwap(ctx, order); err != nil {
		// Swap failed, unlock funds and revert
		k.UnlockFundsForOrder(ctx, order.UserAddress, order)
		k.UnlockFundsForOrder(ctx, matcher, order)
		order.Status = types.SwapOrderStatus_PENDING
		k.SetOrder(ctx, order)
		return fmt.Errorf("failed to execute swap: %w", err)
	}

	// Mark as completed
	order.Status = types.SwapOrderStatus_COMPLETED
	k.SetOrder(ctx, order)

	// Remove from orderbook
	k.RemoveFromOrderbook(ctx, orderID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"order_matched",
			sdk.NewAttribute("order_id", orderID),
			sdk.NewAttribute("creator", order.UserAddress),
			sdk.NewAttribute("matcher", matcher),
		),
	)

	return nil
}

// CancelOrder cancels an order and unlocks funds
func (k Keeper) CancelOrder(
	ctx sdk.Context,
	orderID string,
	reason string,
) error {
	// Get order
	order := k.GetOrder(ctx, orderID)
	if order == nil {
		return fmt.Errorf("order not found: %s", orderID)
	}

	// Can only cancel pending orders
	if order.Status != types.SwapOrderStatus_PENDING {
		return fmt.Errorf("order is not pending")
	}

	// Unlock funds
	if err := k.UnlockFundsForOrder(ctx, order.UserAddress, order); err != nil {
		return fmt.Errorf("failed to unlock funds: %w", err)
	}

	// Update status
	order.Status = types.SwapOrderStatus_CANCELLED
	k.SetOrder(ctx, order)

	// Remove from orderbook
	k.RemoveFromOrderbook(ctx, orderID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"order_cancelled",
			sdk.NewAttribute("order_id", orderID),
			sdk.NewAttribute("reason", reason),
		),
	)

	return nil
}

// ExecuteSwap executes the actual token swap for a matched order
func (k Keeper) ExecuteSwap(ctx sdk.Context, order *types.SwapOrder) error {
	auraAmount, _ := sdk.NewIntFromString(order.AuraAmount)
	otherAmount, _ := sdk.NewIntFromString(order.OtherAmount)

	if order.OrderType == types.SwapOrderType_SELL {
		// Creator sells AURA, receives other coin
		// Matcher buys AURA, pays other coin

		// Transfer AURA: creator → matcher
		if err := k.TransferLockedFunds(ctx, order.UserAddress, order.MatcherAddress, "uaura", auraAmount); err != nil {
			return fmt.Errorf("failed to transfer AURA: %w", err)
		}

		// Transfer other coin: matcher → creator
		if err := k.TransferLockedFunds(ctx, order.MatcherAddress, order.UserAddress, order.OtherCoin, otherAmount); err != nil {
			return fmt.Errorf("failed to transfer %s: %w", order.OtherCoin, err)
		}

	} else {
		// Creator buys AURA, pays other coin
		// Matcher sells AURA, receives other coin

		// Transfer AURA: matcher → creator
		if err := k.TransferLockedFunds(ctx, order.MatcherAddress, order.UserAddress, "uaura", auraAmount); err != nil {
			return fmt.Errorf("failed to transfer AURA: %w", err)
		}

		// Transfer other coin: creator → matcher
		if err := k.TransferLockedFunds(ctx, order.UserAddress, order.MatcherAddress, order.OtherCoin, otherAmount); err != nil {
			return fmt.Errorf("failed to transfer %s: %w", order.OtherCoin, err)
		}
	}

	return nil
}

// ============================
// ORDER STORAGE
// ============================

// GetOrder returns an order by ID
func (k Keeper) GetOrder(ctx sdk.Context, orderID string) *types.SwapOrder {
	store := ctx.KVStore(k.storeKey)
	key := types.OrderKey(orderID)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var order types.SwapOrder
	k.cdc.MustUnmarshal(bz, &order)
	return &order
}

// SetOrder stores an order
func (k Keeper) SetOrder(ctx sdk.Context, order *types.SwapOrder) {
	store := ctx.KVStore(k.storeKey)
	key := types.OrderKey(order.OrderId)

	bz := k.cdc.MustMarshal(order)
	store.Set(key, bz)
}

// DeleteOrder removes an order
func (k Keeper) DeleteOrder(ctx sdk.Context, orderID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.OrderKey(orderID)
	store.Delete(key)
}

// GetAllOrders returns all orders
func (k Keeper) GetAllOrders(ctx sdk.Context) []*types.SwapOrder {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.OrderPrefix)
	defer iterator.Close()

	var orders []*types.SwapOrder
	for ; iterator.Valid(); iterator.Next() {
		var order types.SwapOrder
		k.cdc.MustUnmarshal(iterator.Value(), &order)
		orders = append(orders, &order)
	}

	return orders
}

// GetOrdersByStatus returns orders filtered by status
func (k Keeper) GetOrdersByStatus(ctx sdk.Context, status types.SwapOrderStatus) []*types.SwapOrder {
	allOrders := k.GetAllOrders(ctx)

	filtered := []*types.SwapOrder{}
	for _, order := range allOrders {
		if order.Status == status {
			filtered = append(filtered, order)
		}
	}

	return filtered
}

// GetOrdersByUser returns all orders created by a user
func (k Keeper) GetOrdersByUser(ctx sdk.Context, userAddress string) []*types.SwapOrder {
	allOrders := k.GetAllOrders(ctx)

	userOrders := []*types.SwapOrder{}
	for _, order := range allOrders {
		if order.UserAddress == userAddress {
			userOrders = append(userOrders, order)
		}
	}

	return userOrders
}

// ============================
// ORDERBOOK INDEXING
// ============================

// AddToOrderbook adds an order to the orderbook index
func (k Keeper) AddToOrderbook(ctx sdk.Context, order *types.SwapOrder) {
	// Index by trading pair
	pairKey := fmt.Sprintf("%s-%s", "uaura", order.OtherCoin)

	store := ctx.KVStore(k.storeKey)
	key := types.OrderbookKey(pairKey, order.OrderId)

	// Just store the order ID (actual order is in main storage)
	store.Set(key, []byte(order.OrderId))
}

// RemoveFromOrderbook removes an order from the orderbook index
func (k Keeper) RemoveFromOrderbook(ctx sdk.Context, orderID string) {
	order := k.GetOrder(ctx, orderID)
	if order == nil {
		return
	}

	pairKey := fmt.Sprintf("%s-%s", "uaura", order.OtherCoin)

	store := ctx.KVStore(k.storeKey)
	key := types.OrderbookKey(pairKey, orderID)
	store.Delete(key)
}

// GetOrderbookForPair returns all pending orders for a trading pair
func (k Keeper) GetOrderbookForPair(ctx sdk.Context, coinA, coinB string) []*types.SwapOrder {
	pairKey := fmt.Sprintf("%s-%s", coinA, coinB)

	store := ctx.KVStore(k.storeKey)
	prefix := types.OrderbookPairPrefix(pairKey)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var orders []*types.SwapOrder
	for ; iterator.Valid(); iterator.Next() {
		orderID := string(iterator.Value())
		order := k.GetOrder(ctx, orderID)
		if order != nil && order.Status == types.SwapOrderStatus_PENDING {
			orders = append(orders, order)
		}
	}

	return orders
}

// ============================
// FUND LOCKING
// ============================

// LockFundsForOrder locks funds for an order
func (k Keeper) LockFundsForOrder(ctx sdk.Context, address string, denom string, amount sdk.Int) error {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}

	coins := sdk.NewCoins(sdk.NewCoin(denom, amount))

	// Send from user to DEX module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, coins); err != nil {
		return err
	}

	return nil
}

// UnlockFundsForOrder unlocks funds for a cancelled order
func (k Keeper) UnlockFundsForOrder(ctx sdk.Context, address string, order *types.SwapOrder) error {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}

	auraAmount, _ := sdk.NewIntFromString(order.AuraAmount)
	otherAmount, _ := sdk.NewIntFromString(order.OtherAmount)

	var coins sdk.Coins
	if order.OrderType == types.SwapOrderType_SELL {
		coins = sdk.NewCoins(sdk.NewCoin("uaura", auraAmount))
	} else {
		coins = sdk.NewCoins(sdk.NewCoin(order.OtherCoin, otherAmount))
	}

	// Send back from DEX module to user
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, coins); err != nil {
		return err
	}

	return nil
}

// TransferLockedFunds transfers locked funds between users
func (k Keeper) TransferLockedFunds(ctx sdk.Context, from, to, denom string, amount sdk.Int) error {
	toAddr, err := sdk.AccAddressFromBech32(to)
	if err != nil {
		return err
	}

	coins := sdk.NewCoins(sdk.NewCoin(denom, amount))

	// Funds are already in module account, just send to recipient
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, toAddr, coins); err != nil {
		return err
	}

	return nil
}

// ============================
// HELPERS
// ============================

// GenerateOrderID generates a unique order ID
func (k Keeper) GenerateOrderID(ctx sdk.Context, creator string) string {
	return fmt.Sprintf("order-%s-%d", creator, ctx.BlockHeight())
}

// CleanupExpiredOrders removes expired orders (should be called in EndBlocker)
func (k Keeper) CleanupExpiredOrders(ctx sdk.Context) {
	orders := k.GetOrdersByStatus(ctx, types.SwapOrderStatus_PENDING)

	for _, order := range orders {
		if ctx.BlockTime().After(order.ExpiresAt) {
			k.CancelOrder(ctx, order.OrderId, "expired")
		}
	}
}
