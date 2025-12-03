package keeper

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	if err := k.reentrancyGuard.EnterScoped(scope); err != nil {
		return nil, fmt.Errorf("reentrancy detected: %w", err)
	}
	defer k.reentrancyGuard.ExitScoped(scope)

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
	expirationTime := ctx.BlockTime().Add(time.Duration(expirationMinutes) * time.Minute)

	pricePerAura := otherAmount.ToLegacyDec().Quo(auraAmount.ToLegacyDec())

	// Create order struct
	order := &types.SwapOrder{
		OrderId:      orderID,
		OrderType:    orderType,
		AuraAmount:   auraAmount.String(),
		OtherCoin:    otherCoin,
		OtherAmount:  otherAmount.String(),
		UserAddress:  creator,
		Status:       types.SwapOrderStatus_PENDING,
		Timestamp:    timestamppb.New(ctx.BlockTime()),
		ExpiresAt:    timestamppb.New(expirationTime),
		PricePerAura: pricePerAura.String(),
	}

	// Store order BEFORE external call
	k.SetOrder(ctx, order)

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
	if err := k.reentrancyGuard.EnterScoped(scope); err != nil {
		return fmt.Errorf("reentrancy detected: %w", err)
	}
	defer k.reentrancyGuard.ExitScoped(scope)

	// === 1. CHECKS - Validate all inputs and state ===

	// Validate order status
	if order.Status != types.SwapOrderStatus_PENDING {
		return fmt.Errorf("order is not pending: %s", order.Status.String())
	}

	// Check expiration
	if ctx.BlockTime().After(order.ExpiresAt.AsTime()) {
		// Order expired, cancel it
		k.CancelOrder(ctx, orderID, "expired")
		return fmt.Errorf("order has expired")
	}

	// Parse amounts
	auraAmount, ok := sdkmath.NewIntFromString(order.AuraAmount)
	if !ok {
		return fmt.Errorf("invalid AURA amount")
	}
	otherAmount, ok := sdkmath.NewIntFromString(order.OtherAmount)
	if !ok {
		return fmt.Errorf("invalid other amount")
	}

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
	order.MatchedAt = timestamppb.New(ctx.BlockTime())
	k.SetOrder(ctx, order)

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
			k.SetOrder(ctx, order)
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
			k.SetOrder(ctx, order)
			return fmt.Errorf("failed to lock matcher funds: %w", err)
		}
	}

	// Execute swap (transfers funds between parties)
	if err := k.ExecuteSwap(ctx, order); err != nil {
		// Swap failed, unlock funds and revert
		k.UnlockFundsForOrder(ctx, order.UserAddress, order)
		k.UnlockFundsForOrder(ctx, matcher, order)
		k.AddToOrderbook(ctx, order)
		order.Status = types.SwapOrderStatus_PENDING
		order.MatcherAddress = ""
		order.MatchedAt = nil
		k.SetOrder(ctx, order)
		return fmt.Errorf("failed to execute swap: %w", err)
	}

	// Mark as completed AFTER successful swap
	order.Status = types.SwapOrderStatus_COMPLETED
	k.SetOrder(ctx, order)

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
	if err := k.reentrancyGuard.EnterScoped(scope); err != nil {
		return fmt.Errorf("reentrancy detected: %w", err)
	}
	defer k.reentrancyGuard.ExitScoped(scope)

	// === 1. CHECKS - Validate state ===

	// Can only cancel pending orders
	if order.Status != types.SwapOrderStatus_PENDING {
		return fmt.Errorf("order is not pending")
	}

	// === 2. EFFECTS - Update state BEFORE external calls ===

	// Update status BEFORE external call
	order.Status = types.SwapOrderStatus_CANCELLED
	k.SetOrder(ctx, order)

	// Remove from orderbook BEFORE external call
	k.RemoveFromOrderbook(ctx, orderID)

	// === 3. INTERACTIONS - External calls LAST ===

	// Unlock funds
	if err := k.UnlockFundsForOrder(ctx, order.UserAddress, order); err != nil {
		// Rollback state changes
		k.AddToOrderbook(ctx, order)
		order.Status = types.SwapOrderStatus_PENDING
		k.SetOrder(ctx, order)
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

	auraAmount, ok := sdkmath.NewIntFromString(order.AuraAmount)
	if !ok {
		return fmt.Errorf("invalid AURA amount in order")
	}
	otherAmount, ok := sdkmath.NewIntFromString(order.OtherAmount)
	if !ok {
		return fmt.Errorf("invalid other amount in order")
	}

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
func (k Keeper) SetOrder(ctx sdk.Context, order *types.SwapOrder) {
	store := ctx.KVStore(k.storeKey)
	key := types.OrderKey(order.OrderId)

	isNew := store.Get(key) == nil

	bz := k.cdc.MustMarshal(order)
	store.Set(key, bz)

	if isNew {
		k.indexUserOrder(ctx, order)
	}
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

	var orders []*types.SwapOrder
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
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
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

	auraAmount, _ := sdkmath.NewIntFromString(order.AuraAmount)
	otherAmount, _ := sdkmath.NewIntFromString(order.OtherAmount)

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
func (k Keeper) CleanupExpiredOrders(ctx sdk.Context) {
	orders := k.GetOrdersByStatus(ctx, types.SwapOrderStatus_PENDING)

	for _, order := range orders {
		if ctx.BlockTime().After(order.ExpiresAt.AsTime()) {
			k.CancelOrder(ctx, order.OrderId, "expired")
		}
	}
}

func (k Keeper) indexUserOrder(ctx sdk.Context, order *types.SwapOrder) {
	if order == nil || order.UserAddress == "" || order.OrderId == "" {
		return
	}

	ts := ctx.BlockTime().Unix()
	if order.Timestamp != nil {
		ts = order.Timestamp.AsTime().Unix()
	}
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
			BuyOrders:    []*types.SwapOrder{},
			SellOrders:   []*types.SwapOrder{},
			TotalPending: uint64(len(orders)),
		}

		bestBid := sdkmath.LegacyZeroDec()
		bestAsk := sdkmath.LegacyZeroDec()
		firstBid := true
		firstAsk := true

		for _, order := range orders {
			price := orderPriceDec(order)

			if order.OrderType == types.SwapOrderType_BUY {
				book.BuyOrders = append(book.BuyOrders, order)
				if firstBid || price.GT(bestBid) {
					bestBid = price
					firstBid = false
				}
			} else {
				book.SellOrders = append(book.SellOrders, order)
				if firstAsk || price.LT(bestAsk) || bestAsk.IsZero() {
					bestAsk = price
					firstAsk = false
				}
			}
		}

		sort.Slice(book.BuyOrders, func(i, j int) bool {
			return orderPriceDec(book.BuyOrders[i]).GT(orderPriceDec(book.BuyOrders[j]))
		})
		sort.Slice(book.SellOrders, func(i, j int) bool {
			return orderPriceDec(book.SellOrders[i]).LT(orderPriceDec(book.SellOrders[j]))
		})

		book.BestBid = bestBid.String()
		book.BestAsk = bestAsk.String()

		if !bestAsk.IsZero() && !bestBid.IsZero() {
			book.SpreadPercent = bestAsk.Sub(bestBid).Quo(bestAsk).MulInt64(100).String()
		}

		orderbooks = append(orderbooks, book)
	}

	return orderbooks
}
