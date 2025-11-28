package keeper

import (
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

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
		Timelock:      timestamppb.New(expiry),
		Status:        htlcStatusActive,
		Sender:        sender,
		Recipient:     recipient,
		Amount:        amount.Amount.String(),
		Denom:         amount.Denom,
		RefundAddress: sender,
	}

	k.setHTLC(ctx, htlcID, htlc)
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
	if ctx.BlockTime().After(htlc.Timelock.AsTime()) {
		return types.ErrHTLCExpired
	}

	hash := k.GenerateSecureHash([]byte(secret))
	if !strings.EqualFold(hash, htlc.SecretHash) {
		return fmt.Errorf("secret does not match hash")
	}

	amount, ok := sdkmath.NewIntFromString(htlc.Amount)
	if !ok {
		return fmt.Errorf("invalid HTLC amount")
	}

	recipientAddr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, sdk.NewCoins(sdk.NewCoin(htlc.Denom, amount))); err != nil {
		return err
	}

	htlc.Secret = secret
	htlc.Status = htlcStatusClaimed
	k.setHTLC(ctx, htlcID, htlc)
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
	if ctx.BlockTime().Before(htlc.Timelock.AsTime()) {
		return fmt.Errorf("htlc not yet expired")
	}

	amount, ok := sdkmath.NewIntFromString(htlc.Amount)
	if !ok {
		return fmt.Errorf("invalid HTLC amount")
	}

	senderAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddr, sdk.NewCoins(sdk.NewCoin(htlc.Denom, amount))); err != nil {
		return err
	}

	htlc.Status = htlcStatusRefunded
	k.setHTLC(ctx, htlcID, htlc)
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
	k.cdc.MustUnmarshal(bz, &data)
	return &data, true
}

func (k Keeper) setHTLC(ctx sdk.Context, htlcID string, htlc *types.HTLCData) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.HTLCKey(htlcID), k.cdc.MustMarshal(htlc))
}

// CleanupExpiredHTLCs refunds expired HTLCs during EndBlock.
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
		k.cdc.MustUnmarshal(iterator.Value(), &data)

		if data.Status != htlcStatusActive {
			continue
		}
		if data.Timelock == nil || now.Before(data.Timelock.AsTime()) {
			continue
		}

		amount, ok := sdkmath.NewIntFromString(data.Amount)
		if !ok || amount.IsZero() {
			ctx.Logger().Error("htlc cleanup skipped due to invalid amount", "amount", data.Amount)
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

		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(data.Denom, amount))); err != nil {
			ctx.Logger().Error("failed to refund expired htlc", "address", refundAddr, "error", err)
			continue
		}

		data.Status = htlcStatusRefunded
		htlcID := string(iterator.Key()[len(types.HTLCPrefix):])
		k.setHTLC(ctx, htlcID, &data)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"htlc_refunded",
				sdk.NewAttribute("htlc_id", htlcID),
				sdk.NewAttribute("recipient", refundAddr),
			),
		)
	}
}
