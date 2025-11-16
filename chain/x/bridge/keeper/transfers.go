package keeper

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/bridge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// CROSS-CHAIN TRANSFERS (Lock & Mint / Burn & Unlock)
// ============================================================================

// InitiateTransferToChain initiates a cross-chain transfer from AURA to PAW/XAI
// Uses lock & mint mechanism: lock AURA tokens, mint wrapped tokens on target chain
func (k Keeper) InitiateTransferToChain(
	ctx sdk.Context,
	sender string,
	recipient string,
	amount sdk.Coin,
	targetChain string, // "paw" or "xai"
) (*types.CrossChainTransfer, error) {
	params := k.GetParams(ctx)

	// Validate target chain
	if !k.IsSupportedChain(targetChain) {
		return nil, fmt.Errorf("unsupported target chain: %s", targetChain)
	}

	// Check minimum transfer amount
	if amount.Amount.LT(params.MinTransferAmount) {
		return nil, fmt.Errorf(
			"transfer amount %s is below minimum %s",
			amount.Amount.String(),
			params.MinTransferAmount.String(),
		)
	}

	// Lock sender's tokens
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
		return nil, fmt.Errorf("failed to lock tokens: %w", err)
	}

	// Generate transfer ID
	transferID := k.GenerateTransferID(ctx, sender, targetChain)

	// Create transfer record
	transfer := &types.CrossChainTransfer{
		TransferId:   transferID,
		SourceChain:  "aura",
		TargetChain:  targetChain,
		Sender:       sender,
		Recipient:    recipient,
		Amount:       amount.String(),
		Status:       types.TransferStatus_PENDING,
		InitiatedAt:  ctx.BlockTime(),
		SourceTxHash: ctx.TxBytes(),
	}

	// Store transfer
	k.SetTransfer(ctx, transfer)

	// Emit event for relayers to pick up
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"cross_chain_transfer_initiated",
			sdk.NewAttribute("transfer_id", transferID),
			sdk.NewAttribute("source_chain", "aura"),
			sdk.NewAttribute("target_chain", targetChain),
			sdk.NewAttribute("sender", sender),
			sdk.NewAttribute("recipient", recipient),
			sdk.NewAttribute("amount", amount.String()),
		),
	)

	return transfer, nil
}

// CompleteTransferFromChain completes an incoming transfer from PAW/XAI to AURA
// Uses burn & unlock mechanism: burn wrapped tokens on source chain, unlock AURA
func (k Keeper) CompleteTransferFromChain(
	ctx sdk.Context,
	transferID string,
	sourceChain string,
	sender string,
	recipient string,
	amount sdk.Coin,
	proof []byte, // Merkle proof of burn on source chain
) error {
	// Verify proof (in production, verify Merkle proof against source chain state root)
	// For now, trust the input

	// Check if transfer already completed
	existing := k.GetTransfer(ctx, transferID)
	if existing != nil {
		return fmt.Errorf("transfer already exists: %s", transferID)
	}

	// Unlock tokens to recipient
	recipientAddr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		recipientAddr,
		sdk.NewCoins(amount),
	); err != nil {
		return fmt.Errorf("failed to unlock tokens: %w", err)
	}

	// Create transfer record
	transfer := &types.CrossChainTransfer{
		TransferId:  transferID,
		SourceChain: sourceChain,
		TargetChain: "aura",
		Sender:      sender,
		Recipient:   recipient,
		Amount:      amount.String(),
		Status:      types.TransferStatus_COMPLETED,
		InitiatedAt: ctx.BlockTime(),
		CompletedAt: ctx.BlockTime(),
		Proof:       proof,
	}

	// Store transfer
	k.SetTransfer(ctx, transfer)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"cross_chain_transfer_completed",
			sdk.NewAttribute("transfer_id", transferID),
			sdk.NewAttribute("source_chain", sourceChain),
			sdk.NewAttribute("recipient", recipient),
			sdk.NewAttribute("amount", amount.String()),
		),
	)

	return nil
}

// ConfirmTransfer confirms a pending transfer (called by relayers after observing mint on target chain)
func (k Keeper) ConfirmTransfer(
	ctx sdk.Context,
	transferID string,
	targetTxHash []byte,
	relayer string,
) error {
	transfer := k.GetTransfer(ctx, transferID)
	if transfer == nil {
		return fmt.Errorf("transfer not found: %s", transferID)
	}

	if transfer.Status != types.TransferStatus_PENDING {
		return fmt.Errorf("transfer is not pending: %s", transfer.Status.String())
	}

	// Update transfer
	transfer.Status = types.TransferStatus_COMPLETED
	transfer.CompletedAt = ctx.BlockTime()
	transfer.TargetTxHash = targetTxHash
	transfer.RelayerAddress = relayer

	k.SetTransfer(ctx, transfer)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"transfer_confirmed",
			sdk.NewAttribute("transfer_id", transferID),
			sdk.NewAttribute("relayer", relayer),
		),
	)

	return nil
}

// ============================================================================
// WRAPPED TOKENS (PAW.token, XAI.coin)
// ============================================================================

// MintWrappedToken mints wrapped tokens on AURA for tokens locked on PAW/XAI
func (k Keeper) MintWrappedToken(
	ctx sdk.Context,
	recipient string,
	wrappedDenom string, // "paw.token" or "xai.coin"
	amount sdk.Int,
	proof []byte,
) error {
	// Verify wrapped token is supported
	wrappedToken := k.GetWrappedToken(ctx, wrappedDenom)
	if wrappedToken == nil {
		return fmt.Errorf("unsupported wrapped token: %s", wrappedDenom)
	}

	// Verify proof of lock on source chain
	// (in production, verify Merkle proof)

	// Mint wrapped tokens
	recipientAddr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}

	coins := sdk.NewCoins(sdk.NewCoin(wrappedDenom, amount))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return fmt.Errorf("failed to mint wrapped tokens: %w", err)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, coins); err != nil {
		return fmt.Errorf("failed to send wrapped tokens: %w", err)
	}

	// Update total supply
	wrappedToken.TotalSupply = wrappedToken.TotalSupply.Add(amount).String()
	k.SetWrappedToken(ctx, wrappedToken)

	return nil
}

// BurnWrappedToken burns wrapped tokens on AURA to unlock tokens on PAW/XAI
func (k Keeper) BurnWrappedToken(
	ctx sdk.Context,
	sender string,
	wrappedDenom string,
	amount sdk.Int,
	targetRecipient string, // Address on source chain to unlock to
) error {
	// Verify wrapped token exists
	wrappedToken := k.GetWrappedToken(ctx, wrappedDenom)
	if wrappedToken == nil {
		return fmt.Errorf("unsupported wrapped token: %s", wrappedDenom)
	}

	// Burn tokens
	senderAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return fmt.Errorf("invalid sender address: %w", err)
	}

	coins := sdk.NewCoins(sdk.NewCoin(wrappedDenom, amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, coins); err != nil {
		return fmt.Errorf("failed to receive tokens: %w", err)
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
		return fmt.Errorf("failed to burn tokens: %w", err)
	}

	// Update total supply
	currentSupply, _ := sdk.NewIntFromString(wrappedToken.TotalSupply)
	wrappedToken.TotalSupply = currentSupply.Sub(amount).String()
	k.SetWrappedToken(ctx, wrappedToken)

	// Emit event for relayers to unlock on source chain
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"wrapped_token_burned",
			sdk.NewAttribute("wrapped_denom", wrappedDenom),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("target_chain", wrappedToken.SourceChain),
			sdk.NewAttribute("target_recipient", targetRecipient),
		),
	)

	return nil
}

// ============================================================================
// STORAGE
// ============================================================================

// GetTransfer returns a transfer by ID
func (k Keeper) GetTransfer(ctx sdk.Context, transferID string) *types.CrossChainTransfer {
	store := ctx.KVStore(k.storeKey)
	key := types.TransferKey(transferID)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var transfer types.CrossChainTransfer
	k.cdc.MustUnmarshal(bz, &transfer)
	return &transfer
}

// SetTransfer stores a transfer
func (k Keeper) SetTransfer(ctx sdk.Context, transfer *types.CrossChainTransfer) {
	store := ctx.KVStore(k.storeKey)
	key := types.TransferKey(transfer.TransferId)

	bz := k.cdc.MustMarshal(transfer)
	store.Set(key, bz)
}

// GetWrappedToken returns a wrapped token by denom
func (k Keeper) GetWrappedToken(ctx sdk.Context, wrappedDenom string) *types.WrappedToken {
	store := ctx.KVStore(k.storeKey)
	key := types.WrappedTokenKey(wrappedDenom)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var token types.WrappedToken
	k.cdc.MustUnmarshal(bz, &token)
	return &token
}

// SetWrappedToken stores a wrapped token
func (k Keeper) SetWrappedToken(ctx sdk.Context, token *types.WrappedToken) {
	store := ctx.KVStore(k.storeKey)
	key := types.WrappedTokenKey(token.WrappedDenom)

	bz := k.cdc.MustMarshal(token)
	store.Set(key, bz)
}

// ============================================================================
// HELPERS
// ============================================================================

// IsSupportedChain checks if a chain is supported
func (k Keeper) IsSupportedChain(chain string) bool {
	supportedChains := []string{"paw", "xai", "osmosis"}
	for _, supported := range supportedChains {
		if chain == supported {
			return true
		}
	}
	return false
}

// GenerateTransferID generates a unique transfer ID
func (k Keeper) GenerateTransferID(ctx sdk.Context, sender, targetChain string) string {
	return fmt.Sprintf("transfer-%s-%s-%d", sender, targetChain, ctx.BlockHeight())
}
