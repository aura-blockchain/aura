package keeper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const userOrderHistoryLimit = 200

// ============================
// P2P ORDERBOOK (from simple_swap_gui.py)
// ============================

// CreateOrder creates a new swap order in the orderbook
// Adapted from simple_swap_gui.py lines 120-180
//
// SECURITY: Protected with reentrancy guard and follows Checks-Effects-Interactions pattern
// to prevent manipulation through callbacks during external calls
func (k Keeper) CreateOrder(
	ctx sdk.Context,
	creator string,
	orderType types.SwapOrderType,
	auraAmount sdkmath.Int,
	otherCoin string,
	otherAmount sdkmath.Int,
	expirationMinutes uint64,
) (*types.SwapOrder, error) {
	// Derive pool identifier for scoped reentrancy protection
	poolID := k.GeneratePoolID("uaura", otherCoin)
	scope := fmt.Sprintf("orderbook:%s", poolID)

	// Acquire reentrancy lock for this orderbook
	// NOTE: Cosmos SDK context is deterministic and single-threaded per block execution,
	// making reentrancy impossible in the traditional sense. However, we maintain this
	// check for cross-module call protection and future-proofing.
	if err := k.securityKeeper.EnterNoReentrant(ctx, scope); err != nil {
		return nil, fmt.Errorf("reentrancy detected: %w", err)
	}
	defer k.securityKeeper.ExitNoReentrant(ctx, scope)

	// === 1. CHECKS - Validate all inputs and state ===

	// Validate amounts
	if auraAmount.LTE(sdkmath.ZeroInt()) {
		return nil, fmt.Errorf("AURA amount must be positive")
	}
	if otherAmount.LTE(sdkmath.ZeroInt()) {
		return nil, fmt.Errorf("other coin amount must be positive")
	}

	// Detect order manipulation
	if err := k.DetectOrderManipulation(ctx, creator, poolID, auraAmount); err != nil {
		return nil, err
	}

	// Validate user has sufficient balance
	addr, err := sdk.AccAddressFromBech32(creator)
	if err != nil {
		return nil, fmt.Errorf("invalid creator address: %w", err)
	}

	var requiredDenom string
	var requiredAmount sdkmath.Int
	if orderType == types.SwapOrderType_SELL {
		requiredDenom = "uaura"
		requiredAmount = auraAmount
	} else {
		requiredDenom = otherCoin
		requiredAmount = otherAmount
	}

	balance := k.bankKeeper.GetBalance(ctx, addr, requiredDenom)
	if balance.Amount.LT(requiredAmount) {
		return nil, fmt.Errorf("insufficient balance: have %s, need %s %s",
			balance.Amount.String(), requiredAmount.String(), requiredDenom)
	}

	// === 2. EFFECTS - Update state BEFORE external calls ===

	// Generate order ID
	orderID := k.GenerateOrderID(ctx, creator)

	// Calculate expiration time
	maxExpirationMinutes := uint64(math.MaxInt64 / int64(time.Minute))
	if expirationMinutes > maxExpirationMinutes {
		expirationMinutes = maxExpirationMinutes
	}
	expirationMinutesInt := int64(expirationMinutes)
	expirationTime := ctx.BlockTime().Add(time.Duration(expirationMinutesInt) * time.Minute)

	pricePerAura := otherAmount.ToLegacyDec().Quo(auraAmount.ToLegacyDec())

	// Create order struct
	order := &types.SwapOrder{
		OrderId:      orderID,
		OrderType:    orderType,
		AuraAmount:   auraAmount,
		OtherCoin:    otherCoin,
		OtherAmount:  otherAmount,
		UserAddress:  creator,
		Status:       types.SwapOrderStatus_PENDING,
		Timestamp:    ctx.BlockTime(),
		ExpiresAt:    expirationTime,
		PricePerAura: pricePerAura,
	}

	// Store order BEFORE external call
	if err := k.SetOrder(ctx, order); err != nil {
		return nil, err
	}

	// Add to orderbook index BEFORE external call
	k.AddToOrderbook(ctx, order)

	// === 3. INTERACTIONS - External calls LAST ===

	// Lock user's funds based on order type
	if orderType == types.SwapOrderType_SELL {
		// Selling AURA, lock AURA
		if err := k.LockFundsForOrder(ctx, creator, "uaura", auraAmount); err != nil {
			// Rollback state changes
			k.RemoveFromOrderbook(ctx, orderID)
			k.DeleteOrder(ctx, orderID)
			return nil, fmt.Errorf("failed to lock AURA: %w", err)
		}
	} else {
		// Buying AURA, lock other coin
		if err := k.LockFundsForOrder(ctx, creator, otherCoin, otherAmount); err != nil {
			// Rollback state changes
			k.RemoveFromOrderbook(ctx, orderID)
			k.DeleteOrder(ctx, orderID)
			return nil, fmt.Errorf("failed to lock %s: %w", otherCoin, err)
		}
	}

	// Emit event (safe - no state changes after this)
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
//
// SECURITY: Protected with reentrancy guard and follows Checks-Effects-Interactions pattern
// State changes (order status, indexing) occur BEFORE external transfers to prevent
// reentrancy attacks where callbacks could manipulate order state during execution
func (k Keeper) MatchOrder(
	ctx sdk.Context,
	matcher string,
	orderID string,
) error {
	// Get order for scoped reentrancy protection
	order := k.GetOrder(ctx, orderID)
	if order == nil {
		return fmt.Errorf("order not found: %s", orderID)
	}

	// Derive scope from order
	poolID := k.GeneratePoolID("uaura", order.OtherCoin)
	scope := fmt.Sprintf("orderbook:%s", poolID)

	// Acquire reentrancy lock for this orderbook
	// NOTE: Cosmos SDK context is deterministic and single-threaded per block execution,
	// making reentrancy impossible in the traditional sense. However, we maintain this
	// check for cross-module call protection and future-proofing.
	if err := k.securityKeeper.EnterNoReentrant(ctx, scope); err != nil {
		return fmt.Errorf("reentrancy detected: %w", err)
	}
	defer k.securityKeeper.ExitNoReentrant(ctx, scope)

	// === 1. CHECKS - Validate all inputs and state ===

	// Validate order status
	if order.Status != types.SwapOrderStatus_PENDING {
		return fmt.Errorf("order is not pending: %s", order.Status.String())
	}

	// Check expiration
	if ctx.BlockTime().After(order.ExpiresAt) {
		// Order expired, cancel it
		if err := k.CancelOrder(ctx, orderID, "expired"); err != nil {
			return fmt.Errorf("failed to cancel expired order: %w", err)
		}
		return fmt.Errorf("order has expired")
	}

	// AuraAmount and OtherAmount are already math.Int types (customtype in proto)
	auraAmount := order.AuraAmount
	otherAmount := order.OtherAmount

	// Validate matcher is not the order creator (can't match own order)
	if matcher == order.UserAddress {
		return fmt.Errorf("cannot match your own order")
	}

	// Validate matcher has sufficient balance
	matcherAddr, err := sdk.AccAddressFromBech32(matcher)
	if err != nil {
		return fmt.Errorf("invalid matcher address: %w", err)
	}

	var requiredDenom string
	var requiredAmount sdkmath.Int
	if order.OrderType == types.SwapOrderType_SELL {
		// Order is selling AURA, matcher needs to provide other coin
		requiredDenom = order.OtherCoin
		requiredAmount = otherAmount
	} else {
		// Order is buying AURA, matcher needs to provide AURA
		requiredDenom = "uaura"
		requiredAmount = auraAmount
	}

	matcherBalance := k.bankKeeper.GetBalance(ctx, matcherAddr, requiredDenom)
	if matcherBalance.Amount.LT(requiredAmount) {
		return fmt.Errorf("matcher has insufficient balance: have %s, need %s %s",
			matcherBalance.Amount.String(), requiredAmount.String(), requiredDenom)
	}

	// === 2. EFFECTS - Update state BEFORE external calls ===

	// Update order status to MATCHED BEFORE external calls
	order.Status = types.SwapOrderStatus_MATCHED
	order.MatcherAddress = matcher
	matchedTime := ctx.BlockTime()
	order.MatchedAt = &matchedTime
	if err := k.SetOrder(ctx, order); err != nil {
		return err
	}

	// Remove from orderbook BEFORE external calls (prevents double-matching)
	k.RemoveFromOrderbook(ctx, orderID)

	// === 3. INTERACTIONS - External calls LAST ===

	// Lock matcher's funds
	if order.OrderType == types.SwapOrderType_SELL {
		// Order is selling AURA, matcher needs to provide other coin
		if err := k.LockFundsForOrder(ctx, matcher, order.OtherCoin, otherAmount); err != nil {
			// Rollback state changes
			k.AddToOrderbook(ctx, order)
			order.Status = types.SwapOrderStatus_PENDING
			order.MatcherAddress = ""
			order.MatchedAt = nil
			_ = k.SetOrder(ctx, order) // Best effort rollback
			return fmt.Errorf("failed to lock matcher funds: %w", err)
		}
	} else {
		// Order is buying AURA, matcher needs to provide AURA
		if err := k.LockFundsForOrder(ctx, matcher, "uaura", auraAmount); err != nil {
			// Rollback state changes
			k.AddToOrderbook(ctx, order)
			order.Status = types.SwapOrderStatus_PENDING
			order.MatcherAddress = ""
			order.MatchedAt = nil
			_ = k.SetOrder(ctx, order) // Best effort rollback
			return fmt.Errorf("failed to lock matcher funds: %w", err)
		}
	}

	// Execute swap (transfers funds between parties)
	if err := k.ExecuteSwap(ctx, order); err != nil {
		// Swap failed, unlock funds and revert
		if err := k.UnlockFundsForOrder(ctx, order.UserAddress, order); err != nil {
			return fmt.Errorf("failed to unlock order funds: %w", err)
		}
		if err := k.UnlockFundsForOrder(ctx, matcher, order); err != nil {
			return fmt.Errorf("failed to unlock matcher funds: %w", err)
		}
		k.AddToOrderbook(ctx, order)
		order.Status = types.SwapOrderStatus_PENDING
		order.MatcherAddress = ""
		order.MatchedAt = nil
		_ = k.SetOrder(ctx, order) // Best effort rollback
		return fmt.Errorf("failed to execute swap: %w", err)
	}

	// Mark as completed AFTER successful swap
	order.Status = types.SwapOrderStatus_COMPLETED
	if err := k.SetOrder(ctx, order); err != nil {
		return err
	}

	// Emit event (safe - no state changes after this)
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
//
// SECURITY: Protected with reentrancy guard and follows Checks-Effects-Interactions pattern
// State changes occur BEFORE external calls to prevent reentrancy attacks
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

	// Derive scope for reentrancy protection
	poolID := k.GeneratePoolID("uaura", order.OtherCoin)
	scope := fmt.Sprintf("orderbook:%s", poolID)

	// Acquire reentrancy lock for this orderbook
	// NOTE: Cosmos SDK context is deterministic and single-threaded per block execution,
	// making reentrancy impossible in the traditional sense. However, we maintain this
	// check for cross-module call protection and future-proofing.
	if err := k.securityKeeper.EnterNoReentrant(ctx, scope); err != nil {
		return fmt.Errorf("reentrancy detected: %w", err)
	}
	defer k.securityKeeper.ExitNoReentrant(ctx, scope)

	// === 1. CHECKS - Validate state ===

	// Can only cancel pending orders
	if order.Status != types.SwapOrderStatus_PENDING {
		return fmt.Errorf("order is not pending")
	}

	// === 2. EFFECTS - Update state BEFORE external calls ===

	// Update status BEFORE external call
	order.Status = types.SwapOrderStatus_CANCELLED
	if err := k.SetOrder(ctx, order); err != nil {
		return err
	}

	// Remove from orderbook BEFORE external call
	k.RemoveFromOrderbook(ctx, orderID)

	// === 3. INTERACTIONS - External calls LAST ===

	// Unlock funds
	if err := k.UnlockFundsForOrder(ctx, order.UserAddress, order); err != nil {
		// Rollback state changes
		k.AddToOrderbook(ctx, order)
		order.Status = types.SwapOrderStatus_PENDING
		_ = k.SetOrder(ctx, order) // Best effort rollback
		return fmt.Errorf("failed to unlock funds: %w", err)
	}

	// Emit event (safe - no state changes after this)
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
//
// SECURITY: This function is called from MatchOrder which already holds the reentrancy lock.
// We validate invariants and execute transfers in the correct order to prevent manipulation.
//
// Note: This function should ONLY be called from within MatchOrder's reentrancy-protected section.
func (k Keeper) ExecuteSwap(ctx sdk.Context, order *types.SwapOrder) error {
	// === 1. CHECKS - Parse and validate amounts ===

	// AuraAmount and OtherAmount are already math.Int types (customtype in proto)
	auraAmount := order.AuraAmount
	otherAmount := order.OtherAmount

	// Validate order is in MATCHED status
	if order.Status != types.SwapOrderStatus_MATCHED {
		return fmt.Errorf("order must be in MATCHED status, got: %s", order.Status.String())
	}

	// Validate matcher is set
	if order.MatcherAddress == "" {
		return fmt.Errorf("order has no matcher address")
	}

	// Validate amounts are positive
	if auraAmount.LTE(sdkmath.ZeroInt()) {
		return fmt.Errorf("AURA amount must be positive")
	}
	if otherAmount.LTE(sdkmath.ZeroInt()) {
		return fmt.Errorf("other amount must be positive")
	}

	// === 2. EFFECTS - No additional state changes needed here ===
	// (State was already updated in MatchOrder before calling this function)

	// === 3. INTERACTIONS - Execute transfers ===

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
	if err := k.cdc.Unmarshal(bz, &order); err != nil {
		ctx.Logger().Error("failed to unmarshal swap order",
			"order_id", orderID,
			"error", err,
			"data_len", len(bz))
		return nil
	}
	return &order
}

// SetOrder stores an order
func (k Keeper) SetOrder(ctx sdk.Context, order *types.SwapOrder) error {
	store := ctx.KVStore(k.storeKey)
	key := types.OrderKey(order.OrderId)

	isNew := store.Get(key) == nil

	bz, err := k.cdc.Marshal(order)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal order %s: %v", order.OrderId, err)
	}
	store.Set(key, bz)

	if isNew {
		k.indexUserOrder(ctx, order)
	}
	return nil
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
	iterator := storetypes.KVStorePrefixIterator(store, types.OrderPrefix)
	defer iterator.Close()

	orders := make([]*types.SwapOrder, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var order types.SwapOrder
		if err := k.cdc.Unmarshal(iterator.Value(), &order); err != nil {
			ctx.Logger().Error("failed to unmarshal order in GetAllOrders, skipping",
				"error", err,
				"data_len", len(iterator.Value()))
			continue
		}
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
	store := ctx.KVStore(k.storeKey)
	prefix := types.UserOrderAddressPrefix(userAddress)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	userOrders := []*types.SwapOrder{}
	count := 0

	for ; iterator.Valid(); iterator.Next() {
		orderID := string(iterator.Value())
		order := k.GetOrder(ctx, orderID)
		if order != nil {
			userOrders = append(userOrders, order)
		}
		count++
		if count >= userOrderHistoryLimit {
			break
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

	// Add to expiration index for efficient cleanup
	k.addToExpirationIndex(ctx, order)
}

// addToExpirationIndex adds an order to the expiration time index
func (k Keeper) addToExpirationIndex(ctx sdk.Context, order *types.SwapOrder) {
	if order == nil || order.Status != types.SwapOrderStatus_PENDING {
		return
	}

	store := ctx.KVStore(k.storeKey)
	expirationKey := types.OrderExpirationKey(order.ExpiresAt.Unix(), order.OrderId)
	// Store orderID as value for easy retrieval
	store.Set(expirationKey, []byte(order.OrderId))
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

	// Remove from expiration index
	k.removeFromExpirationIndex(ctx, order)
}

// removeFromExpirationIndex removes an order from the expiration time index
func (k Keeper) removeFromExpirationIndex(ctx sdk.Context, order *types.SwapOrder) {
	if order == nil {
		return
	}

	store := ctx.KVStore(k.storeKey)
	expirationKey := types.OrderExpirationKey(order.ExpiresAt.Unix(), order.OrderId)
	store.Delete(expirationKey)
}

// GetOrderbookForPair returns all pending orders for a trading pair
func (k Keeper) GetOrderbookForPair(ctx sdk.Context, coinA, coinB string) []*types.SwapOrder {
	pairKey := fmt.Sprintf("%s-%s", coinA, coinB)

	store := ctx.KVStore(k.storeKey)
	prefix := types.OrderbookPairPrefix(pairKey)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	orders := make([]*types.SwapOrder, 0, 64)
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
func (k Keeper) LockFundsForOrder(ctx sdk.Context, address string, denom string, amount sdkmath.Int) error {
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

	// AuraAmount and OtherAmount are already math.Int types (customtype in proto)
	auraAmount := order.AuraAmount
	otherAmount := order.OtherAmount

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
func (k Keeper) TransferLockedFunds(ctx sdk.Context, from, to, denom string, amount sdkmath.Int) error {
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
// DEPRECATED: Use CleanupExpiredOrdersBatched for production to prevent consensus failure
func (k Keeper) CleanupExpiredOrders(ctx sdk.Context) {
	orders := k.GetOrdersByStatus(ctx, types.SwapOrderStatus_PENDING)

	for _, order := range orders {
		if ctx.BlockTime().After(order.ExpiresAt) {
			if err := k.CancelOrder(ctx, order.OrderId, "expired"); err != nil {
				ctx.Logger().Error("failed to cancel expired order", "order_id", order.OrderId, "err", err)
			}
		}
	}
}

// CleanupExpiredOrdersBatched removes up to 'limit' expired orders per call.
// Uses cursor-based pagination to track progress across blocks, ensuring
// all expired orders are eventually processed without blocking consensus.
//
// SECURITY: This function is designed to prevent consensus failure by limiting
// the number of operations per block. With 10,000+ orders, unbounded cleanup
// causes block production to exceed timeout, halting the chain.
//
// DEPRECATED: Use CleanupExpiredOrdersOptimized for better performance.
// This function loads ALL pending orders into memory, which is inefficient.
//
// Returns: number of orders processed in this batch
func (k Keeper) CleanupExpiredOrdersBatched(ctx sdk.Context, limit int) int {
	if limit <= 0 {
		limit = types.MaxOrdersCleanupPerBlock
	}

	store := ctx.KVStore(k.storeKey)

	// Get cursor (last processed order ID)
	cursorKey := types.OrderCleanupCursorKey()
	cursorBytes := store.Get(cursorKey)
	cursor := cursorBytes

	// Get all pending orders starting from cursor
	orders := k.GetOrdersByStatus(ctx, types.SwapOrderStatus_PENDING)
	if len(orders) == 0 {
		// No orders to process, reset cursor
		store.Delete(cursorKey)
		return 0
	}

	// Find start position based on cursor
	startIdx := 0
	if cursor != nil {
		cursorOrderID := string(cursor)
		for i, order := range orders {
			if order.OrderId == cursorOrderID {
				// Start from next order after cursor
				startIdx = i + 1
				break
			}
		}
	}

	// If cursor points beyond end of list, reset and start from beginning
	if startIdx >= len(orders) {
		startIdx = 0
	}

	// Process up to 'limit' orders
	processed := 0
	var lastProcessedOrderID string

	for i := startIdx; i < len(orders) && processed < limit; i++ {
		order := orders[i]
		lastProcessedOrderID = order.OrderId

		// Check if order expired
		if ctx.BlockTime().After(order.ExpiresAt) {
			// Cancel expired order
			if err := k.CancelOrder(ctx, order.OrderId, "expired"); err != nil {
				// Log error but continue processing other orders
				ctx.Logger().Error("failed to cancel expired order",
					"order_id", order.OrderId,
					"error", err)
			}
		}

		processed++
	}

	// Update cursor for next block
	if processed > 0 {
		if startIdx+processed >= len(orders) {
			// Completed full pass, reset cursor for next iteration
			store.Delete(cursorKey)
		} else {
			// Save cursor pointing to last processed order
			store.Set(cursorKey, []byte(lastProcessedOrderID))
		}
	}

	return processed
}

// CleanupExpiredOrdersOptimized efficiently removes up to 'limit' expired orders
// using the expiration time index. This function ONLY loads orders that have actually
// expired, preventing memory exhaustion and consensus timeouts.
//
// PERFORMANCE: With 100,000+ orders, this function performs O(limit) operations
// instead of O(total_orders), reducing cleanup time from seconds to milliseconds.
//
// SECURITY: Prevents consensus failure by:
// 1. Only iterating over expired orders (via time index)
// 2. Limiting batch size to prevent long-running operations
// 3. Not loading all pending orders into memory
//
// Returns: number of orders processed in this batch
func (k Keeper) CleanupExpiredOrdersOptimized(ctx sdk.Context, limit int) int {
	if limit <= 0 {
		limit = types.MaxOrdersCleanupPerBlock
	}

	store := ctx.KVStore(k.storeKey)
	currentTime := ctx.BlockTime().Unix()

	// Create iterator over expiration index up to current time
	// This only iterates over orders that have actually expired
	iterator := storetypes.KVStorePrefixIterator(store, types.OrderExpirationPrefix)
	defer iterator.Close()

	processed := 0
	for ; iterator.Valid() && processed < limit; iterator.Next() {
		// Extract expiration time from key
		key := iterator.Key()
		if len(key) < len(types.OrderExpirationPrefix)+8 {
			// Invalid key, skip
			continue
		}

		// Parse expiration timestamp (8 bytes after prefix)
		expiresAtBytes := key[len(types.OrderExpirationPrefix) : len(types.OrderExpirationPrefix)+8]
		expiresAt := int64(binary.BigEndian.Uint64(expiresAtBytes))

		// Stop if we've reached orders that haven't expired yet
		// Since the index is time-sorted, all subsequent orders are also not expired
		if expiresAt > currentTime {
			break
		}

		// Extract order ID from value
		orderID := string(iterator.Value())

		// Get the actual order to verify it's still pending
		order := k.GetOrder(ctx, orderID)
		if order == nil {
			// Order no longer exists, remove stale index entry
			store.Delete(key)
			continue
		}

		// Double-check order is still pending (status might have changed)
		if order.Status != types.SwapOrderStatus_PENDING {
			// Order status changed, remove from expiration index
			store.Delete(key)
			continue
		}

		// Cancel expired order
		if err := k.CancelOrder(ctx, order.OrderId, "expired"); err != nil {
			// Log error but continue processing other orders
			ctx.Logger().Error("failed to cancel expired order",
				"order_id", order.OrderId,
				"expires_at", order.ExpiresAt,
				"current_time", ctx.BlockTime(),
				"error", err)
		}

		processed++
	}

	return processed
}

func (k Keeper) indexUserOrder(ctx sdk.Context, order *types.SwapOrder) {
	if order == nil || order.UserAddress == "" || order.OrderId == "" {
		return
	}

	// Timestamp is time.Time, not *timestamppb.Timestamp
	ts := order.Timestamp.Unix()
	if ts < 0 {
		ts = 0
	}

	store := ctx.KVStore(k.storeKey)
	key := types.UserOrderKey(order.UserAddress, uint64(ts), order.OrderId)
	store.Set(key, []byte(order.OrderId))

	k.enforceUserOrderLimit(ctx, order.UserAddress)
}

func (k Keeper) enforceUserOrderLimit(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	prefix := types.UserOrderAddressPrefix(address)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	count := 0
	for ; iterator.Valid(); iterator.Next() {
		count++
		if count > userOrderHistoryLimit {
			key := append([]byte(nil), iterator.Key()...)
			store.Delete(key)
		}
	}
}

func (k Keeper) addPendingOrderToIndex(ctx sdk.Context, order *types.SwapOrder) {
	if order == nil || order.OrderId == "" {
		return
	}

	stored := k.GetOrder(ctx, order.OrderId)
	if stored == nil {
		return
	}
	if stored.Status != types.SwapOrderStatus_PENDING {
		return
	}

	k.AddToOrderbook(ctx, stored)
}

func (k Keeper) exportOrderbooks(ctx sdk.Context) []*types.Orderbook {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.OrderbookPrefix)
	defer iterator.Close()

	pairs := make(map[string][]*types.SwapOrder)
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()[len(types.OrderbookPrefix):]
		sep := bytes.IndexByte(key, 0x00)
		if sep < 0 {
			continue
		}
		pairKey := string(key[:sep])
		orderID := string(key[sep+1:])

		order := k.GetOrder(ctx, orderID)
		if order == nil || order.Status != types.SwapOrderStatus_PENDING {
			continue
		}

		pairs[pairKey] = append(pairs[pairKey], order)
	}

	if len(pairs) == 0 {
		return nil
	}

	pairKeys := make([]string, 0, len(pairs))
	for pair := range pairs {
		pairKeys = append(pairKeys, pair)
	}
	sort.Strings(pairKeys)

	orderbooks := make([]*types.Orderbook, 0, len(pairKeys))
	for _, pair := range pairKeys {
		orders := pairs[pair]
		book := &types.Orderbook{
			Pair:         pair,
			BuyOrders:    []types.SwapOrder{},
			SellOrders:   []types.SwapOrder{},
			TotalPending: uint64(len(orders)),
		}

		bestBid := sdkmath.LegacyZeroDec()
		bestAsk := sdkmath.LegacyZeroDec()
		firstBid := true
		firstAsk := true

		for _, order := range orders {
			price := orderPriceDec(order)

			if order.OrderType == types.SwapOrderType_BUY {
				// Dereference pointer to append value
				book.BuyOrders = append(book.BuyOrders, *order)
				if firstBid || price.GT(bestBid) {
					bestBid = price
					firstBid = false
				}
			} else {
				// Dereference pointer to append value
				book.SellOrders = append(book.SellOrders, *order)
				if firstAsk || price.LT(bestAsk) || bestAsk.IsZero() {
					bestAsk = price
					firstAsk = false
				}
			}
		}

		sort.Slice(book.BuyOrders, func(i, j int) bool {
			return orderPriceDec(&book.BuyOrders[i]).GT(orderPriceDec(&book.BuyOrders[j]))
		})
		sort.Slice(book.SellOrders, func(i, j int) bool {
			return orderPriceDec(&book.SellOrders[i]).LT(orderPriceDec(&book.SellOrders[j]))
		})

		book.BestBid = bestBid
		book.BestAsk = bestAsk

		if !bestAsk.IsZero() && !bestBid.IsZero() {
			book.SpreadPercent = bestAsk.Sub(bestBid).Quo(bestAsk).MulInt64(100)
		}

		orderbooks = append(orderbooks, book)
	}

	return orderbooks
}
