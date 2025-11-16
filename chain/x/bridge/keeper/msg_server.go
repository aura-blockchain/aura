package keeper

import (
	"context"
	"fmt"

	bridgetypes "github.com/aequitas/aura/chain/x/bridge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ bridgetypes.MsgServer = msgServer{}

// msgServer implements the MsgServer interface
type msgServer struct {
	types.UnimplementedMsgServer
	Keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper *Keeper) bridgetypes.MsgServer {
	return &msgServer{Keeper: keeper}
}

// LockTokens locks tokens on AURA for transfer to PAW or XAI
func (ms msgServer) LockTokens(goCtx context.Context, msg *bridgetypes.MsgLockTokens) (*bridgetypes.MsgLockTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.Sender == "" || msg.TargetChain == "" || msg.Recipient == "" {
		return nil, bridgetypes.ErrInvalidInput.Wrap("sender, target_chain, and recipient are required")
	}

	// Validate target chain
	if msg.TargetChain != "paw" && msg.TargetChain != "xai" {
		return nil, bridgetypes.ErrInvalidChain.Wrapf("unsupported target chain: %s", msg.TargetChain)
	}

	// Validate amount
	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return nil, bridgetypes.ErrInvalidAmount.Wrap("invalid or zero amount")
	}

	// Lock tokens
	transferID, estimatedCompletion, err := ms.Keeper.LockTokens(ctx, msg.Sender, msg.TargetChain, msg.Recipient, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &bridgetypes.MsgLockTokensResponse{
		TransferId:          transferID,
		EstimatedCompletion: estimatedCompletion,
	}, nil
}

// MintTokens mints wrapped tokens on AURA from PAW or XAI (validator-only)
func (ms msgServer) MintTokens(goCtx context.Context, msg *bridgetypes.MsgMintTokens) (*bridgetypes.MsgMintTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate validator
	if !ms.Keeper.IsAuthorizedValidator(ctx, msg.Validator) {
		return nil, bridgetypes.ErrUnauthorized.Wrap("only authorized validators can mint tokens")
	}

	// Validate inputs
	if msg.SourceChain == "" || msg.SourceTxHash == "" || msg.Recipient == "" || msg.Denom == "" {
		return nil, bridgetypes.ErrInvalidInput.Wrap("all fields are required")
	}

	// Validate source chain
	if msg.SourceChain != "paw" && msg.SourceChain != "xai" {
		return nil, bridgetypes.ErrInvalidChain.Wrapf("unsupported source chain: %s", msg.SourceChain)
	}

	// Verify validator signature
	if err := ms.Keeper.VerifyValidatorSignature(ctx, msg); err != nil {
		return nil, err
	}

	// Mint tokens
	wrappedDenom, err := ms.Keeper.MintWrappedTokens(ctx, msg.SourceChain, msg.SourceTxHash, msg.Recipient, msg.Amount, msg.Denom)
	if err != nil {
		return nil, err
	}

	return &bridgetypes.MsgMintTokensResponse{
		Success:      true,
		WrappedDenom: wrappedDenom,
	}, nil
}

// UnlockTokens unlocks tokens on AURA after burn proof from target chain
func (ms msgServer) UnlockTokens(goCtx context.Context, msg *bridgetypes.MsgUnlockTokens) (*bridgetypes.MsgUnlockTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.Sender == "" || msg.SourceChain == "" || msg.BurnTxHash == "" || msg.Denom == "" {
		return nil, bridgetypes.ErrInvalidInput.Wrap("all fields are required")
	}

	// Validate source chain
	if msg.SourceChain != "paw" && msg.SourceChain != "xai" {
		return nil, bridgetypes.ErrInvalidChain.Wrapf("unsupported source chain: %s", msg.SourceChain)
	}

	// Verify multisig validator signatures
	if err := ms.Keeper.VerifyValidatorMultisig(ctx, msg.ValidatorSignatures, msg.BurnTxHash); err != nil {
		return nil, err
	}

	// Unlock tokens
	if err := ms.Keeper.UnlockTokens(ctx, msg.Sender, msg.SourceChain, msg.BurnTxHash, msg.Amount, msg.Denom); err != nil {
		return nil, err
	}

	return &bridgetypes.MsgUnlockTokensResponse{Success: true}, nil
}

// BurnTokens burns wrapped tokens on AURA to unlock on source chain
func (ms msgServer) BurnTokens(goCtx context.Context, msg *bridgetypes.MsgBurnTokens) (*bridgetypes.MsgBurnTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.Sender == "" || msg.TargetChain == "" || msg.Recipient == "" {
		return nil, bridgetypes.ErrInvalidInput.Wrap("sender, target_chain, and recipient are required")
	}

	// Validate target chain
	if msg.TargetChain != "paw" && msg.TargetChain != "xai" {
		return nil, bridgetypes.ErrInvalidChain.Wrapf("unsupported target chain: %s", msg.TargetChain)
	}

	// Validate amount
	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return nil, bridgetypes.ErrInvalidAmount.Wrap("invalid or zero amount")
	}

	// Burn tokens
	transferID, estimatedCompletion, err := ms.Keeper.BurnWrappedTokens(ctx, msg.Sender, msg.TargetChain, msg.Recipient, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &bridgetypes.MsgBurnTokensResponse{
		TransferId:          transferID,
		EstimatedCompletion: estimatedCompletion,
	}, nil
}

// LinkAddress links addresses across AURA, PAW, and XAI for shared identity
func (ms msgServer) LinkAddress(goCtx context.Context, msg *bridgetypes.MsgLinkAddress) (*bridgetypes.MsgLinkAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.AuraAddress == "" || msg.Signer == "" {
		return nil, bridgetypes.ErrInvalidInput.Wrap("aura_address and signer are required")
	}

	// At least one external address must be provided
	if msg.PawAddress == "" && msg.XaiAddress == "" {
		return nil, bridgetypes.ErrInvalidInput.Wrap("at least one external address (PAW or XAI) must be provided")
	}

	// Verify signatures
	if msg.PawAddress != "" {
		if err := ms.Keeper.VerifyPawSignature(ctx, msg.PawAddress, msg.PawSignature, msg.AuraAddress); err != nil {
			return nil, fmt.Errorf("invalid PAW signature: %w", err)
		}
	}

	if msg.XaiAddress != "" {
		if err := ms.Keeper.VerifyXaiSignature(ctx, msg.XaiAddress, msg.XaiSignature, msg.AuraAddress); err != nil {
			return nil, fmt.Errorf("invalid XAI signature: %w", err)
		}
	}

	// Link addresses
	identityID, err := ms.Keeper.LinkAddresses(ctx, msg.AuraAddress, msg.PawAddress, msg.XaiAddress)
	if err != nil {
		return nil, err
	}

	return &bridgetypes.MsgLinkAddressResponse{
		Success:          true,
		LinkedIdentityId: identityID,
	}, nil
}

// CrossChainSwap initiates cross-chain swap
func (ms msgServer) CrossChainSwap(goCtx context.Context, msg *bridgetypes.MsgCrossChainSwap) (*bridgetypes.MsgCrossChainSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.Sender == "" || msg.SourceChain == "" || msg.TargetChain == "" || msg.TargetDenom == "" {
		return nil, bridgetypes.ErrInvalidInput.Wrap("sender, chains, and target_denom are required")
	}

	// Validate amount
	if !msg.InputCoin.IsValid() || msg.InputCoin.IsZero() {
		return nil, bridgetypes.ErrInvalidAmount.Wrap("invalid or zero input amount")
	}

	// Validate slippage
	if msg.MaxSlippageBps > 10000 {
		return nil, bridgetypes.ErrInvalidInput.Wrap("max_slippage_bps cannot exceed 10000 (100%)")
	}

	// Execute cross-chain swap
	swapID, route, estimatedCompletion, err := ms.Keeper.ExecuteCrossChainSwap(
		ctx,
		msg.Sender,
		msg.SourceChain,
		msg.InputCoin,
		msg.TargetChain,
		msg.TargetDenom,
		msg.MinTargetAmount,
		msg.Recipient,
		msg.MaxSlippageBps,
	)
	if err != nil {
		return nil, err
	}

	return &bridgetypes.MsgCrossChainSwapResponse{
		SwapId:              swapID,
		Route:               route,
		EstimatedCompletion: estimatedCompletion,
	}, nil
}

// RelayTransfer relays cross-chain transfer (relayer-only)
func (ms msgServer) RelayTransfer(goCtx context.Context, msg *bridgetypes.MsgRelayTransfer) (*bridgetypes.MsgRelayTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate relayer is authorized
	if !ms.Keeper.IsAuthorizedRelayer(ctx, msg.Relayer) {
		return nil, bridgetypes.ErrUnauthorized.Wrap("only authorized relayers can relay transfers")
	}

	// Validate inputs
	if msg.TransferId == "" || msg.TargetTxHash == "" || msg.Status == "" {
		return nil, bridgetypes.ErrInvalidInput.Wrap("transfer_id, target_tx_hash, and status are required")
	}

	// Relay transfer
	if err := ms.Keeper.RelayTransferUpdate(ctx, msg.TransferId, msg.TargetTxHash, msg.Status); err != nil {
		return nil, err
	}

	return &bridgetypes.MsgRelayTransferResponse{Success: true}, nil
}
