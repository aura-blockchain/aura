package keeper

import (
	"fmt"
	"strings"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

const (
	htlcStatusActive   = "active"
	htlcStatusClaimed  = "claimed"
	htlcStatusRefunded = "refunded"
)

// CreateHTLC locks funds in a hash time-lock contract and returns its ID.
func (k Keeper) CreateHTLC(
	ctx sdk.Context,
	sender string,
	recipient string,
	amount sdk.Coin,
	secretHash string,
	timelockSeconds uint64,
) (string, error) {
	if !amount.Amount.IsPositive() {
		return "", fmt.Errorf("amount must be positive")
	}
	if recipient == "" || sender == "" {
		return "", fmt.Errorf("sender and recipient must be provided")
	}
	if secretHash == "" {
		return "", fmt.Errorf("secret hash cannot be empty")
	}
	if timelockSeconds == 0 {
		return "", fmt.Errorf("timelock must be greater than zero")
	}

	senderAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return "", err
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, sdk.NewCoins(amount)); err != nil {
		return "", err
	}

	idPayload := fmt.Sprintf("%s|%s|%s|%d", sender, recipient, secretHash, ctx.BlockHeight())
	htlcID := fmt.Sprintf("htlc-%s", k.GenerateSecureHash([]byte(idPayload)))

	expiry := ctx.BlockTime().Add(time.Duration(timelockSeconds) * time.Second)
	htlc := &types.HTLCData{
		SecretHash:    secretHash,
		Secret:        "",
		Timelock:      expiry,
		Status:        htlcStatusActive,
		Sender:        sender,
		Recipient:     recipient,
		Amount:        amount.Amount,
		Denom:         amount.Denom,
		RefundAddress: sender,
	}

	if err := k.setHTLC(ctx, htlcID, htlc); err != nil {
		return "", err
	}
	return htlcID, nil
}

// ClaimHTLC releases funds to the recipient when the secret is revealed.
func (k Keeper) ClaimHTLC(ctx sdk.Context, recipient string, htlcID string, secret string) error {
	htlc, found := k.GetHTLC(ctx, htlcID)
	if !found {
		return types.ErrHTLCNotFound
	}

	if htlc.Status != htlcStatusActive {
		return types.ErrHTLCAlreadyClaimed
	}
	if htlc.Recipient != recipient {
		return fmt.Errorf("unauthorized recipient")
	}
	if ctx.BlockTime().After(htlc.Timelock) {
		return types.ErrHTLCExpired
	}

	hash := k.GenerateSecureHash([]byte(secret))
	if !strings.EqualFold(hash, htlc.SecretHash) {
		return fmt.Errorf("secret does not match hash")
	}

	recipientAddr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, sdk.NewCoins(sdk.NewCoin(htlc.Denom, htlc.Amount))); err != nil {
		return err
	}

	htlc.Secret = secret
	htlc.Status = htlcStatusClaimed
	if err := k.setHTLC(ctx, htlcID, htlc); err != nil {
		return err
	}
	return nil
}

// RefundHTLC refunds funds to the sender after expiry.
func (k Keeper) RefundHTLC(ctx sdk.Context, sender string, htlcID string) error {
	htlc, found := k.GetHTLC(ctx, htlcID)
	if !found {
		return types.ErrHTLCNotFound
	}

	if htlc.Status != htlcStatusActive {
		return types.ErrHTLCAlreadyClaimed
	}
	if htlc.Sender != sender {
		return fmt.Errorf("unauthorized sender")
	}
	if ctx.BlockTime().Before(htlc.Timelock) {
		return fmt.Errorf("htlc not yet expired")
	}

	senderAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddr, sdk.NewCoins(sdk.NewCoin(htlc.Denom, htlc.Amount))); err != nil {
		return err
	}

	htlc.Status = htlcStatusRefunded
	if err := k.setHTLC(ctx, htlcID, htlc); err != nil {
		return err
	}
	return nil
}

// GetHTLC retrieves HTLC data by ID.
func (k Keeper) GetHTLC(ctx sdk.Context, htlcID string) (*types.HTLCData, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.HTLCKey(htlcID))
	if bz == nil {
		return nil, false
	}

	var data types.HTLCData
	if err := k.cdc.Unmarshal(bz, &data); err != nil {
		ctx.Logger().Error("failed to unmarshal HTLC data",
			"htlc_id", htlcID,
			"error", err,
			"data_len", len(bz))
		return nil, false
	}
	return &data, true
}

func (k Keeper) setHTLC(ctx sdk.Context, htlcID string, htlc *types.HTLCData) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(htlc)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal HTLC %s: %v", htlcID, err)
	}
	store.Set(types.HTLCKey(htlcID), bz)
	return nil
}

// CleanupExpiredHTLCs refunds expired HTLCs during EndBlock.
// DEPRECATED: Use CleanupExpiredHTLCsBatched for production to prevent consensus failure
func (k Keeper) CleanupExpiredHTLCs(ctx sdk.Context) {
	if k.bankKeeper == nil {
		return
	}

	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.HTLCPrefix)
	defer iterator.Close()

	now := ctx.BlockTime()

	for ; iterator.Valid(); iterator.Next() {
		var data types.HTLCData
		if err := k.cdc.Unmarshal(iterator.Value(), &data); err != nil {
			ctx.Logger().Error("failed to unmarshal HTLC during cleanup, skipping",
				"error", err,
				"data_len", len(iterator.Value()))
			continue
		}

		if data.Status != htlcStatusActive {
			continue
		}
		if data.Timelock.IsZero() || now.Before(data.Timelock) {
			continue
		}

		if data.Amount.IsZero() {
			ctx.Logger().Error("htlc cleanup skipped due to invalid amount", "amount", data.Amount.String())
			continue
		}

		refundAddr := data.RefundAddress
		if refundAddr == "" {
			refundAddr = data.Sender
		}
		addr, err := sdk.AccAddressFromBech32(refundAddr)
		if err != nil {
			ctx.Logger().Error("htlc cleanup skipped due to invalid refund address", "address", refundAddr, "error", err)
			continue
		}

		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(data.Denom, data.Amount))); err != nil {
			ctx.Logger().Error("failed to refund expired htlc", "address", refundAddr, "error", err)
			continue
		}

		data.Status = htlcStatusRefunded
		htlcID := string(iterator.Key()[len(types.HTLCPrefix):])
		if err := k.setHTLC(ctx, htlcID, &data); err != nil {
			ctx.Logger().Error("failed to update HTLC status after refund", "htlc_id", htlcID, "error", err)
			continue
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"htlc_refunded",
				sdk.NewAttribute("htlc_id", htlcID),
				sdk.NewAttribute("recipient", refundAddr),
			),
		)
	}
}

// CleanupExpiredHTLCsBatched refunds up to 'limit' expired HTLCs per call.
// Uses cursor-based iteration to track progress across blocks, ensuring
// all expired HTLCs are eventually processed without blocking consensus.
//
// SECURITY: This function is designed to prevent consensus failure by limiting
// the number of operations per block. With many HTLCs, unbounded cleanup
// causes block production to exceed timeout, halting the chain.
//
// Returns: number of HTLCs processed in this batch
func (k Keeper) CleanupExpiredHTLCsBatched(ctx sdk.Context, limit int) int {
	if limit <= 0 {
		limit = types.MaxHTLCsCleanupPerBlock
	}

	if k.bankKeeper == nil {
		return 0
	}

	store := ctx.KVStore(k.storeKey)
	now := ctx.BlockTime()

	// Get cursor (last processed HTLC key)
	cursorKey := types.HTLCCleanupCursorKey()
	cursorBytes := store.Get(cursorKey)

	// Create iterator starting from cursor position
	var iterator storetypes.Iterator
	if cursorBytes != nil && len(cursorBytes) > 0 {
		// Start from cursor position (next key after cursor)
		iterator = store.Iterator(cursorBytes, nil)
		// Skip the cursor key itself if it exists
		if iterator.Valid() && string(iterator.Key()) == string(cursorBytes) {
			iterator.Next()
		}
	} else {
		// Start from beginning of HTLC prefix
		iterator = storetypes.KVStorePrefixIterator(store, types.HTLCPrefix)
	}
	defer iterator.Close()

	processed := 0
	var lastProcessedKey []byte

	for ; iterator.Valid() && processed < limit; iterator.Next() {
		// Only process HTLCs (skip if we've moved beyond HTLC prefix)
		if !hasPrefix(iterator.Key(), types.HTLCPrefix) {
			break
		}

		lastProcessedKey = append([]byte(nil), iterator.Key()...)

		var data types.HTLCData
		if err := k.cdc.Unmarshal(iterator.Value(), &data); err != nil {
			ctx.Logger().Error("failed to unmarshal HTLC during batched cleanup, skipping",
				"error", err,
				"data_len", len(iterator.Value()))
			processed++
			continue
		}

		// Skip if not active or not expired
		if data.Status != htlcStatusActive {
			processed++
			continue
		}
		if data.Timelock.IsZero() || now.Before(data.Timelock) {
			processed++
			continue
		}

		// Validate amount
		if data.Amount.IsZero() {
			ctx.Logger().Error("htlc cleanup skipped due to invalid amount",
				"amount", data.Amount.String())
			processed++
			continue
		}

		// Get refund address
		refundAddr := data.RefundAddress
		if refundAddr == "" {
			refundAddr = data.Sender
		}
		addr, err := sdk.AccAddressFromBech32(refundAddr)
		if err != nil {
			ctx.Logger().Error("htlc cleanup skipped due to invalid refund address",
				"address", refundAddr,
				"error", err)
			processed++
			continue
		}

		// Perform refund
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr,
			sdk.NewCoins(sdk.NewCoin(data.Denom, data.Amount))); err != nil {
			ctx.Logger().Error("failed to refund expired htlc",
				"address", refundAddr,
				"error", err)
			processed++
			continue
		}

		// Update HTLC status
		data.Status = htlcStatusRefunded
		htlcID := string(iterator.Key()[len(types.HTLCPrefix):])
		if err := k.setHTLC(ctx, htlcID, &data); err != nil {
			ctx.Logger().Error("failed to update HTLC status after refund",
				"htlc_id", htlcID,
				"error", err)
			processed++
			continue
		}

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"htlc_refunded",
				sdk.NewAttribute("htlc_id", htlcID),
				sdk.NewAttribute("recipient", refundAddr),
			),
		)

		processed++
	}

	// Update cursor for next block
	if processed > 0 {
		if !iterator.Valid() || !hasPrefix(iterator.Key(), types.HTLCPrefix) {
			// Completed full pass through all HTLCs, reset cursor
			store.Delete(cursorKey)
		} else {
			// Save cursor to continue from this position next block
			store.Set(cursorKey, lastProcessedKey)
		}
	}

	return processed
}

// hasPrefix checks if key starts with prefix
func hasPrefix(key, prefix []byte) bool {
	if len(key) < len(prefix) {
		return false
	}
	for i := range prefix {
		if key[i] != prefix[i] {
			return false
		}
	}
	return true
}
