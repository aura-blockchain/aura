package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================
// HTLC (Hash Time-Locked Contracts) for Atomic Swaps
// Enables trustless cross-chain swaps
// ============================

// CreateHTLC creates a new Hash Time-Locked Contract
// This locks funds that can only be claimed with the secret preimage
func (k Keeper) CreateHTLC(
	ctx sdk.Context,
	sender string,
	recipient string,
	amount sdk.Coin,
	secretHash string,
	timelockMinutes uint64,
) (*types.HTLCData, error) {
	// Validate inputs
	if amount.Amount.LTE(sdk.ZeroInt()) {
		return nil, fmt.Errorf("amount must be positive")
	}
	if secretHash == "" {
		return nil, fmt.Errorf("secret hash cannot be empty")
	}
	if timelockMinutes == 0 {
		return nil, fmt.Errorf("timelock must be positive")
	}

	// Calculate timelock expiration
	timelock := ctx.BlockTime().Add(time.Duration(timelockMinutes) * time.Minute)

	// Generate HTLC ID
	htlcID := k.GenerateHTLCID(ctx, sender, secretHash)

	// Lock sender's funds
	senderAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		senderAddr,
		types.ModuleName,
		sdk.NewCoins(amount),
	); err != nil {
		return nil, fmt.Errorf("failed to lock funds: %w", err)
	}

	// Create HTLC
	htlc := &types.HTLCData{
		HtlcId:     htlcID,
		Sender:     sender,
		Recipient:  recipient,
		Amount:     amount.String(),
		SecretHash: secretHash,
		Timelock:   timelock,
		Status:     types.HTLCStatus_PENDING,
		CreatedAt:  ctx.BlockTime(),
		Secret:     "", // Will be revealed when claimed
	}

	// Store HTLC
	k.SetHTLC(ctx, htlc)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"htlc_created",
			sdk.NewAttribute("htlc_id", htlcID),
			sdk.NewAttribute("sender", sender),
			sdk.NewAttribute("recipient", recipient),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("secret_hash", secretHash),
			sdk.NewAttribute("timelock", timelock.String()),
		),
	)

	return htlc, nil
}

// ClaimHTLC claims an HTLC by revealing the secret preimage
func (k Keeper) ClaimHTLC(
	ctx sdk.Context,
	htlcID string,
	secret string,
	claimer string,
) error {
	// Get HTLC
	htlc := k.GetHTLC(ctx, htlcID)
	if htlc == nil {
		return fmt.Errorf("HTLC not found: %s", htlcID)
	}

	// Validate status
	if htlc.Status != types.HTLCStatus_PENDING {
		return fmt.Errorf("HTLC is not pending: %s", htlc.Status.String())
	}

	// Check if timelock has expired
	if ctx.BlockTime().After(htlc.Timelock) {
		return fmt.Errorf("HTLC has expired, use RefundHTLC instead")
	}

	// Verify claimer is the recipient
	if claimer != htlc.Recipient {
		return fmt.Errorf("only recipient can claim HTLC")
	}

	// Verify secret matches hash
	if !k.VerifySecret(secret, htlc.SecretHash) {
		return fmt.Errorf("invalid secret: hash does not match")
	}

	// Parse amount
	amount, err := sdk.ParseCoinNormalized(htlc.Amount)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	// Transfer funds to recipient
	recipientAddr, err := sdk.AccAddressFromBech32(htlc.Recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		recipientAddr,
		sdk.NewCoins(amount),
	); err != nil {
		return fmt.Errorf("failed to transfer funds: %w", err)
	}

	// Update HTLC
	htlc.Status = types.HTLCStatus_COMPLETED
	htlc.Secret = secret
	htlc.ClaimedAt = ctx.BlockTime()
	k.SetHTLC(ctx, htlc)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"htlc_claimed",
			sdk.NewAttribute("htlc_id", htlcID),
			sdk.NewAttribute("claimer", claimer),
			sdk.NewAttribute("secret", secret),
		),
	)

	return nil
}

// RefundHTLC refunds an expired HTLC back to the sender
func (k Keeper) RefundHTLC(
	ctx sdk.Context,
	htlcID string,
	refunder string,
) error {
	// Get HTLC
	htlc := k.GetHTLC(ctx, htlcID)
	if htlc == nil {
		return fmt.Errorf("HTLC not found: %s", htlcID)
	}

	// Validate status
	if htlc.Status != types.HTLCStatus_PENDING {
		return fmt.Errorf("HTLC is not pending: %s", htlc.Status.String())
	}

	// Check if timelock has expired
	if !ctx.BlockTime().After(htlc.Timelock) {
		return fmt.Errorf("HTLC has not expired yet, cannot refund")
	}

	// Verify refunder is the sender
	if refunder != htlc.Sender {
		return fmt.Errorf("only sender can refund expired HTLC")
	}

	// Parse amount
	amount, err := sdk.ParseCoinNormalized(htlc.Amount)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	// Refund to sender
	senderAddr, err := sdk.AccAddressFromBech32(htlc.Sender)
	if err != nil {
		return fmt.Errorf("invalid sender address: %w", err)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		senderAddr,
		sdk.NewCoins(amount),
	); err != nil {
		return fmt.Errorf("failed to refund funds: %w", err)
	}

	// Update HTLC
	htlc.Status = types.HTLCStatus_REFUNDED
	htlc.RefundedAt = ctx.BlockTime()
	k.SetHTLC(ctx, htlc)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"htlc_refunded",
			sdk.NewAttribute("htlc_id", htlcID),
			sdk.NewAttribute("refunder", refunder),
		),
	)

	return nil
}

// ============================
// ATOMIC SWAP WORKFLOW
// ============================

// InitiateAtomicSwap initiates a cross-chain atomic swap
// This is the first step where Alice creates HTLC on chain A
func (k Keeper) InitiateAtomicSwap(
	ctx sdk.Context,
	initiator string,
	counterparty string,
	sendAmount sdk.Coin,
	receiveAmount sdk.Coin,
	receiveDenom string,
	timelockMinutes uint64,
) (*types.AtomicSwap, error) {
	// Generate random secret
	secret := k.GenerateSecret(ctx, initiator)
	secretHash := k.HashSecret(secret)

	// Create HTLC on this chain
	htlc, err := k.CreateHTLC(
		ctx,
		initiator,
		counterparty,
		sendAmount,
		secretHash,
		timelockMinutes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTLC: %w", err)
	}

	// Create atomic swap record
	swap := &types.AtomicSwap{
		SwapId:        k.GenerateSwapID(ctx, initiator),
		Initiator:     initiator,
		Counterparty:  counterparty,
		SendAmount:    sendAmount.String(),
		ReceiveAmount: receiveAmount.String(),
		ReceiveDenom:  receiveDenom,
		SecretHash:    secretHash,
		Secret:        secret, // Stored temporarily, will be revealed when claiming
		Status:        types.AtomicSwapStatus_INITIATED,
		HtlcId:        htlc.HtlcId,
		CreatedAt:     ctx.BlockTime(),
	}

	// Store swap
	k.SetAtomicSwap(ctx, swap)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"atomic_swap_initiated",
			sdk.NewAttribute("swap_id", swap.SwapId),
			sdk.NewAttribute("initiator", initiator),
			sdk.NewAttribute("counterparty", counterparty),
			sdk.NewAttribute("secret_hash", secretHash),
			sdk.NewAttribute("htlc_id", htlc.HtlcId),
		),
	)

	return swap, nil
}

// CompleteAtomicSwap completes an atomic swap by claiming counterparty's HTLC
func (k Keeper) CompleteAtomicSwap(
	ctx sdk.Context,
	swapID string,
	secret string,
	claimer string,
) error {
	// Get swap
	swap := k.GetAtomicSwap(ctx, swapID)
	if swap == nil {
		return fmt.Errorf("atomic swap not found: %s", swapID)
	}

	// Verify secret
	if !k.VerifySecret(secret, swap.SecretHash) {
		return fmt.Errorf("invalid secret")
	}

	// Claim HTLC
	if err := k.ClaimHTLC(ctx, swap.HtlcId, secret, claimer); err != nil {
		return fmt.Errorf("failed to claim HTLC: %w", err)
	}

	// Update swap status
	swap.Status = types.AtomicSwapStatus_COMPLETED
	swap.CompletedAt = ctx.BlockTime()
	k.SetAtomicSwap(ctx, swap)

	return nil
}

// ============================
// HTLC STORAGE
// ============================

// GetHTLC returns an HTLC by ID
func (k Keeper) GetHTLC(ctx sdk.Context, htlcID string) *types.HTLCData {
	store := ctx.KVStore(k.storeKey)
	key := types.HTLCKey(htlcID)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var htlc types.HTLCData
	k.cdc.MustUnmarshal(bz, &htlc)
	return &htlc
}

// SetHTLC stores an HTLC
func (k Keeper) SetHTLC(ctx sdk.Context, htlc *types.HTLCData) {
	store := ctx.KVStore(k.storeKey)
	key := types.HTLCKey(htlc.HtlcId)

	bz := k.cdc.MustMarshal(htlc)
	store.Set(key, bz)
}

// GetAtomicSwap returns an atomic swap by ID
func (k Keeper) GetAtomicSwap(ctx sdk.Context, swapID string) *types.AtomicSwap {
	store := ctx.KVStore(k.storeKey)
	key := types.AtomicSwapKey(swapID)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var swap types.AtomicSwap
	k.cdc.MustUnmarshal(bz, &swap)
	return &swap
}

// SetAtomicSwap stores an atomic swap
func (k Keeper) SetAtomicSwap(ctx sdk.Context, swap *types.AtomicSwap) {
	store := ctx.KVStore(k.storeKey)
	key := types.AtomicSwapKey(swap.SwapId)

	bz := k.cdc.MustMarshal(swap)
	store.Set(key, bz)
}

// ============================
// HELPERS
// ============================

// GenerateSecret generates a random secret for HTLC
func (k Keeper) GenerateSecret(ctx sdk.Context, initiator string) string {
	// In production, use cryptographically secure random
	data := fmt.Sprintf("%s-%d-%d", initiator, ctx.BlockHeight(), ctx.BlockTime().Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// HashSecret hashes a secret using SHA-256
func (k Keeper) HashSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}

// VerifySecret verifies a secret matches the hash
func (k Keeper) VerifySecret(secret, secretHash string) bool {
	return k.HashSecret(secret) == secretHash
}

// GenerateHTLCID generates a unique HTLC ID
func (k Keeper) GenerateHTLCID(ctx sdk.Context, sender, secretHash string) string {
	return fmt.Sprintf("htlc-%s-%s", sender, secretHash[:8])
}

// GenerateSwapID generates a unique atomic swap ID
func (k Keeper) GenerateSwapID(ctx sdk.Context, initiator string) string {
	return fmt.Sprintf("swap-%s-%d", initiator, ctx.BlockHeight())
}

// CleanupExpiredHTLCs automatically refunds expired HTLCs (called in EndBlocker)
func (k Keeper) CleanupExpiredHTLCs(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.HTLCPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var htlc types.HTLCData
		k.cdc.MustUnmarshal(iterator.Value(), &htlc)

		if htlc.Status == types.HTLCStatus_PENDING && ctx.BlockTime().After(htlc.Timelock) {
			// Auto-refund expired HTLC
			k.RefundHTLC(ctx, htlc.HtlcId, htlc.Sender)
		}
	}
}
