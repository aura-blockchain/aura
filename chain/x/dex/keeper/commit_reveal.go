package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================
// COMMIT-REVEAL SCHEME
// Front-Running Protection
// ============================

// CommitOrder creates a commitment for a large order (phase 1 of commit-reveal)
// The order details are hashed to prevent validators/MEV bots from front-running
//
// SECURITY: Commit-reveal prevents front-running by hiding order details until reveal.
// Large orders (>= threshold) must use this scheme to protect users from MEV extraction.
func (k Keeper) CommitOrder(
	ctx sdk.Context,
	sender string,
	commitHash []byte,
) (string, error) {
	// === 1. CHECKS - Validate inputs ===

	// Validate sender address
	_, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return "", fmt.Errorf("invalid sender address: %w", err)
	}

	// Validate commit hash (must be SHA-256, 32 bytes)
	if len(commitHash) != 32 {
		return "", fmt.Errorf("invalid commit hash length: expected 32 bytes, got %d", len(commitHash))
	}

	// Get params
	params := k.GetParams(ctx)

	// Generate unique commitment ID
	commitID := k.GenerateCommitmentID(ctx, sender)

	// Check if commitment already exists for this sender (prevent spam)
	existingCommitments := k.GetCommitmentsBySender(ctx, sender)
	if len(existingCommitments) > 0 {
		// Allow only one pending commitment per sender
		return "", types.ErrCommitmentAlreadyExists
	}

	// === 2. EFFECTS - Update state BEFORE external calls ===

	// Calculate reveal deadline
	revealDeadline := ctx.BlockTime().Add(time.Duration(params.CommitRevealWindow) * time.Second)

	// Create commitment
	commitment := &types.OrderCommitment{
		CommitId:       commitID,
		Sender:         sender,
		CommitHash:     commitHash,
		CommittedAt:    timestamppb.New(ctx.BlockTime()),
		RevealDeadline: timestamppb.New(revealDeadline),
		CommitHeight:   ctx.BlockHeight(),
	}

	// Store commitment
	k.SetOrderCommitment(ctx, commitment)

	// === 3. INTERACTIONS - External calls (none needed) ===

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOrderCommitted,
			sdk.NewAttribute("commit_id", commitID),
			sdk.NewAttribute("sender", sender),
			sdk.NewAttribute("reveal_deadline", revealDeadline.Format(time.RFC3339)),
		),
	)

	return commitID, nil
}

// RevealOrder reveals a committed order (phase 2 of commit-reveal)
// Verifies the hash matches, then queues the order for batch execution
//
// SECURITY: Hash verification ensures the revealed order matches the commitment.
// This prevents users from changing their order after seeing other orders.
func (k Keeper) RevealOrder(
	ctx sdk.Context,
	commitID string,
	sender string,
	orderType types.SwapOrderType,
	auraAmount sdkmath.Int,
	otherCoin string,
	otherAmount sdkmath.Int,
	salt []byte,
) (string, error) {
	// === 1. CHECKS - Validate all inputs and state ===

	// Get commitment
	commitment, found := k.GetOrderCommitment(ctx, commitID)
	if !found {
		return "", types.ErrCommitmentNotFound
	}

	// Verify sender matches commitment
	if commitment.Sender != sender {
		return "", fmt.Errorf("sender mismatch: expected %s, got %s", commitment.Sender, sender)
	}

	// Verify reveal deadline
	if ctx.BlockTime().After(commitment.RevealDeadline.AsTime()) {
		// Cleanup expired commitment
		k.DeleteOrderCommitment(ctx, commitID)
		return "", types.ErrRevealDeadlineExpired
	}

	// Compute hash of revealed order
	revealedHash := k.ComputeOrderHash(orderType, auraAmount, otherCoin, otherAmount, salt)

	// Verify commitment hash matches revealed hash
	if hex.EncodeToString(commitment.CommitHash) != hex.EncodeToString(revealedHash) {
		return "", types.ErrHashMismatch
	}

	// Validate order parameters
	if auraAmount.LTE(sdkmath.ZeroInt()) {
		return "", fmt.Errorf("AURA amount must be positive")
	}
	if otherAmount.LTE(sdkmath.ZeroInt()) {
		return "", fmt.Errorf("other coin amount must be positive")
	}

	// Parse sender address (needed for balance check and later operations)
	addr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return "", fmt.Errorf("invalid sender address: %w", err)
	}

	// === 2. EFFECTS - Update state BEFORE external calls ===

	// Generate order ID
	orderID := k.GenerateOrderID(ctx, sender)

	// Calculate expiration time (24 hours from reveal, not commit)
	expirationTime := ctx.BlockTime().Add(24 * time.Hour)

	pricePerAura := otherAmount.ToLegacyDec().Quo(auraAmount.ToLegacyDec())

	// Create order struct
	order := &types.SwapOrder{
		OrderId:      orderID,
		OrderType:    orderType,
		AuraAmount:   auraAmount.String(),
		OtherCoin:    otherCoin,
		OtherAmount:  otherAmount.String(),
		UserAddress:  sender,
		Status:       types.SwapOrderStatus_PENDING,
		Timestamp:    timestamppb.New(ctx.BlockTime()),
		ExpiresAt:    timestamppb.New(expirationTime),
		PricePerAura: pricePerAura.String(),
	}

	// Delete commitment (prevent reuse)
	k.DeleteOrderCommitment(ctx, commitID)

	// Check if batch execution is enabled
	params := k.GetParams(ctx)
	if params.BatchExecutionEnabled {
		// Queue order for batch execution
		k.QueueOrderForBatch(ctx, order, salt)

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeOrderRevealed,
				sdk.NewAttribute("commit_id", commitID),
				sdk.NewAttribute("order_id", orderID),
				sdk.NewAttribute("sender", sender),
				sdk.NewAttribute("status", "queued_for_batch"),
			),
		)

		return orderID, nil
	}

	// === 3. INTERACTIONS - External calls for immediate execution ===

	// If batch execution is disabled, execute order immediately

	// Check balance before immediate execution (batch execution will check later)
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
		return "", fmt.Errorf("insufficient balance: have %s, need %s %s",
			balance.Amount.String(), requiredAmount.String(), requiredDenom)
	}

	// Store order
	k.SetOrder(ctx, order)

	// Add to orderbook
	k.AddToOrderbook(ctx, order)

	// Lock funds
	if orderType == types.SwapOrderType_SELL {
		if err := k.LockFundsForOrder(ctx, sender, "uaura", auraAmount); err != nil {
			k.RemoveFromOrderbook(ctx, orderID)
			k.DeleteOrder(ctx, orderID)
			return "", fmt.Errorf("failed to lock AURA: %w", err)
		}
	} else {
		if err := k.LockFundsForOrder(ctx, sender, otherCoin, otherAmount); err != nil {
			k.RemoveFromOrderbook(ctx, orderID)
			k.DeleteOrder(ctx, orderID)
			return "", fmt.Errorf("failed to lock %s: %w", otherCoin, err)
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOrderRevealed,
			sdk.NewAttribute("commit_id", commitID),
			sdk.NewAttribute("order_id", orderID),
			sdk.NewAttribute("sender", sender),
			sdk.NewAttribute("status", "executed_immediately"),
		),
	)

	return orderID, nil
}

// RequiresCommitReveal determines if an order requires commit-reveal scheme
// based on the configured threshold.
func (k Keeper) RequiresCommitReveal(ctx sdk.Context, amount sdkmath.Int) bool {
	params := k.GetParams(ctx)

	// Parse threshold from string (proto customtype not yet applied)
	threshold, ok := sdkmath.NewIntFromString(params.CommitRevealThreshold)
	if !ok {
		// If parsing fails, disable commit-reveal (threshold = max)
		return false
	}

	return amount.GTE(threshold)
}

// ExecuteBatch executes all queued orders in a batch
// Orders are sorted by price priority (not arrival time) to prevent front-running
//
// SECURITY: Batch execution with price-priority sorting eliminates time-based ordering
// advantages, making front-running ineffective. This should be called in EndBlocker
// at regular intervals (e.g., every N blocks).
func (k Keeper) ExecuteBatch(ctx sdk.Context) error {
	// Get all queued orders
	queuedOrders := k.GetAllQueuedOrders(ctx)

	if len(queuedOrders) == 0 {
		return nil // Nothing to execute
	}

	// Sort orders by price priority (best prices first)
	// Buy orders: highest price first (willing to pay more)
	// Sell orders: lowest price first (willing to accept less)
	sort.Slice(queuedOrders, func(i, j int) bool {
		return k.compareOrderPriority(queuedOrders[i].Order, queuedOrders[j].Order)
	})

	// Execute each order
	successCount := 0
	for _, queuedOrder := range queuedOrders {
		order := queuedOrder.Order

		// Parse amounts
		auraAmount, ok := sdkmath.NewIntFromString(order.AuraAmount)
		if !ok {
			continue
		}
		otherAmount, ok := sdkmath.NewIntFromString(order.OtherAmount)
		if !ok {
			continue
		}

		// Store order
		k.SetOrder(ctx, order)

		// Add to orderbook
		k.AddToOrderbook(ctx, order)

		// Lock funds
		var lockErr error
		if order.OrderType == types.SwapOrderType_SELL {
			lockErr = k.LockFundsForOrder(ctx, order.UserAddress, "uaura", auraAmount)
		} else {
			lockErr = k.LockFundsForOrder(ctx, order.UserAddress, order.OtherCoin, otherAmount)
		}

		if lockErr != nil {
			// Failed to lock funds, remove from orderbook
			k.RemoveFromOrderbook(ctx, order.OrderId)
			k.DeleteOrder(ctx, order.OrderId)
			continue
		}

		successCount++

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeBatchOrderExecuted,
				sdk.NewAttribute("order_id", order.OrderId),
				sdk.NewAttribute("user", order.UserAddress),
			),
		)
	}

	// Clear the queue
	k.ClearOrderQueue(ctx)

	// Emit batch summary event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBatchExecuted,
			sdk.NewAttribute("total_orders", fmt.Sprintf("%d", len(queuedOrders))),
			sdk.NewAttribute("successful_orders", fmt.Sprintf("%d", successCount)),
		),
	)

	return nil
}

// ============================
// STORAGE OPERATIONS
// ============================

// SetOrderCommitment stores an order commitment
func (k Keeper) SetOrderCommitment(ctx sdk.Context, commitment *types.OrderCommitment) {
	store := ctx.KVStore(k.storeKey)
	key := types.OrderCommitmentKey(commitment.CommitId)
	bz := k.cdc.MustMarshal(commitment)
	store.Set(key, bz)
}

// GetOrderCommitment retrieves an order commitment by ID
func (k Keeper) GetOrderCommitment(ctx sdk.Context, commitID string) (*types.OrderCommitment, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.OrderCommitmentKey(commitID)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var commitment types.OrderCommitment
	k.cdc.MustUnmarshal(bz, &commitment)
	return &commitment, true
}

// DeleteOrderCommitment removes an order commitment
func (k Keeper) DeleteOrderCommitment(ctx sdk.Context, commitID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.OrderCommitmentKey(commitID)
	store.Delete(key)
}

// GetAllOrderCommitments returns all order commitments
func (k Keeper) GetAllOrderCommitments(ctx sdk.Context) []*types.OrderCommitment {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.OrderCommitmentPrefix)
	defer iterator.Close()

	var commitments []*types.OrderCommitment
	for ; iterator.Valid(); iterator.Next() {
		var commitment types.OrderCommitment
		k.cdc.MustUnmarshal(iterator.Value(), &commitment)
		commitments = append(commitments, &commitment)
	}

	return commitments
}

// GetCommitmentsBySender returns all commitments for a sender
func (k Keeper) GetCommitmentsBySender(ctx sdk.Context, sender string) []*types.OrderCommitment {
	allCommitments := k.GetAllOrderCommitments(ctx)

	var senderCommitments []*types.OrderCommitment
	for _, commitment := range allCommitments {
		if commitment.Sender == sender {
			senderCommitments = append(senderCommitments, commitment)
		}
	}

	return senderCommitments
}

// QueueOrderForBatch adds an order to the batch execution queue
func (k Keeper) QueueOrderForBatch(ctx sdk.Context, order *types.SwapOrder, salt []byte) {
	store := ctx.KVStore(k.storeKey)

	queuedOrder := &types.QueuedOrder{
		Order:    order,
		Salt:     salt,
		QueuedAt: timestamppb.New(ctx.BlockTime()),
	}

	key := types.QueuedOrderKey(order.OrderId)
	bz := k.cdc.MustMarshal(queuedOrder)
	store.Set(key, bz)
}

// GetAllQueuedOrders returns all orders queued for batch execution
func (k Keeper) GetAllQueuedOrders(ctx sdk.Context) []*types.QueuedOrder {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.QueuedOrderPrefix)
	defer iterator.Close()

	var queuedOrders []*types.QueuedOrder
	for ; iterator.Valid(); iterator.Next() {
		var queuedOrder types.QueuedOrder
		k.cdc.MustUnmarshal(iterator.Value(), &queuedOrder)
		queuedOrders = append(queuedOrders, &queuedOrder)
	}

	return queuedOrders
}

// ClearOrderQueue removes all queued orders (called after batch execution)
func (k Keeper) ClearOrderQueue(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.QueuedOrderPrefix)
	defer iterator.Close()

	keysToDelete := [][]byte{}
	for ; iterator.Valid(); iterator.Next() {
		keysToDelete = append(keysToDelete, append([]byte(nil), iterator.Key()...))
	}

	for _, key := range keysToDelete {
		store.Delete(key)
	}
}

// CleanupExpiredCommitments removes commitments past their reveal deadline
// Should be called in EndBlocker
func (k Keeper) CleanupExpiredCommitments(ctx sdk.Context) {
	commitments := k.GetAllOrderCommitments(ctx)

	for _, commitment := range commitments {
		if ctx.BlockTime().After(commitment.RevealDeadline.AsTime()) {
			k.DeleteOrderCommitment(ctx, commitment.CommitId)

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeCommitmentExpired,
					sdk.NewAttribute("commit_id", commitment.CommitId),
					sdk.NewAttribute("sender", commitment.Sender),
				),
			)
		}
	}
}

// ============================
// HELPER FUNCTIONS
// ============================

// GenerateCommitmentID generates a unique commitment ID
func (k Keeper) GenerateCommitmentID(ctx sdk.Context, sender string) string {
	return fmt.Sprintf("commit-%s-%d", sender, ctx.BlockHeight())
}

// ComputeOrderHash computes SHA-256 hash of order details and salt
// Hash = SHA256(order_type || aura_amount || other_coin || other_amount || salt)
func (k Keeper) ComputeOrderHash(
	orderType types.SwapOrderType,
	auraAmount sdkmath.Int,
	otherCoin string,
	otherAmount sdkmath.Int,
	salt []byte,
) []byte {
	hasher := sha256.New()

	// Write order type
	hasher.Write([]byte{byte(orderType)})

	// Write AURA amount
	hasher.Write([]byte(auraAmount.String()))

	// Write other coin
	hasher.Write([]byte(otherCoin))

	// Write other amount
	hasher.Write([]byte(otherAmount.String()))

	// Write salt
	hasher.Write(salt)

	return hasher.Sum(nil)
}

// compareOrderPriority compares two orders for batch execution priority
// Returns true if order i should be executed before order j
func (k Keeper) compareOrderPriority(orderI, orderJ *types.SwapOrder) bool {
	priceI, _ := sdkmath.LegacyNewDecFromStr(orderI.PricePerAura)
	priceJ, _ := sdkmath.LegacyNewDecFromStr(orderJ.PricePerAura)

	// Same order type: sort by price
	if orderI.OrderType == orderJ.OrderType {
		if orderI.OrderType == types.SwapOrderType_BUY {
			// Buy orders: highest price first (willing to pay more)
			return priceI.GT(priceJ)
		} else {
			// Sell orders: lowest price first (willing to accept less)
			return priceI.LT(priceJ)
		}
	}

	// Different order types: prioritize buy orders
	return orderI.OrderType == types.SwapOrderType_BUY
}
