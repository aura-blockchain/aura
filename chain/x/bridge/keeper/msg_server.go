package keeper

import (
	context "context"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

var _ bridgepb.MsgServer = (*msgServer)(nil)

type msgServer struct {
	bridgepb.UnimplementedMsgServer
	Keeper *Keeper
}

// NewMsgServerImpl wires the keeper into the protobuf msg server implementation.
func NewMsgServerImpl(k *Keeper) bridgepb.MsgServer {
	return &msgServer{Keeper: k}
}

func normalizeChain(chain string) string {
	return strings.ToLower(strings.TrimSpace(chain))
}

// LockTokens locks native tokens on Aura for cross-chain transfer.
func (ms msgServer) LockTokens(goCtx context.Context, msg *bridgepb.MsgLockTokens) (*bridgepb.MsgLockTokensResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.Amount == nil || !msg.Amount.Amount.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.Keeper.ensureBridgeEnabled(ctx); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	chainID := normalizeChain(msg.TargetChain)
	if chainID == "" {
		return nil, status.Error(codes.InvalidArgument, "target chain required")
	}
	chainCfg, found := ms.Keeper.getChainConfig(ctx, chainID)
	if !found {
		return nil, status.Error(codes.NotFound, types.ErrChainNotFound.Error())
	}
	if !chainCfg.Enabled {
		return nil, status.Error(codes.FailedPrecondition, types.ErrChainDisabled.Error())
	}
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	amnt := sdk.NewCoin(msg.Amount.Denom, msg.Amount.Amount)
	params := ms.Keeper.GetParams(ctx)
	maxAmt, ok := sdkmath.NewIntFromString(params.MaxTransferAmount)
	if ok && amnt.Amount.GT(maxAmt) {
		return nil, status.Error(codes.InvalidArgument, types.ErrCircuitBreakerTripped.Error())
	}
	if ms.Keeper.bankKeeper != nil {
		if err := ms.Keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(amnt)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	transferID := ms.Keeper.nextTransferID(ctx)
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:            transferID,
		SourceChain:           sourceChainAura,
		TargetChain:           chainID,
		Sender:                msg.Sender,
		Recipient:             msg.Recipient,
		Amount:                amnt.Amount.String(),
		Denom:                 amnt.Denom,
		Status:                bridgepb.TransferStatus_PENDING,
		Timestamp:             timestamppb.New(ctx.BlockTime()),
		RequiredConfirmations: params.MinConfirmations,
	}
	ms.Keeper.setTransfer(ctx, transfer)
	return &bridgepb.MsgLockTokensResponse{
		TransferId:          transferID,
		EstimatedCompletion: 600,
	}, nil
}

// MintTokens mints wrapped assets on Aura once sufficient attestations exist.
func (ms msgServer) MintTokens(goCtx context.Context, msg *bridgepb.MsgMintTokens) (*bridgepb.MsgMintTokensResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.Validator == "" {
		return nil, status.Error(codes.InvalidArgument, "validator required")
	}
	amount, ok := sdkmath.NewIntFromString(msg.Amount)
	if !ok || !amount.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	transferID, hasIndex := ms.Keeper.transferIDByHash(ctx, msg.SourceTxHash)
	if !hasIndex {
		transferID = ms.Keeper.nextTransferID(ctx)
	}
	transfer, found := ms.Keeper.getTransfer(ctx, transferID)
	if !found {
		transfer = &bridgepb.CrossChainTransfer{
			TransferId:  transferID,
			SourceChain: normalizeChain(msg.SourceChain),
			TargetChain: sourceChainAura,
			Sender:      msg.SourceChain,
			Recipient:   msg.Recipient,
			Amount:      amount.String(),
			Denom:       msg.Denom,
			Status:      bridgepb.TransferStatus_CONFIRMED,
			Timestamp:   timestamppb.New(ctx.BlockTime()),
		}
		ms.Keeper.setTransfer(ctx, transfer)
		ms.Keeper.indexTransferHash(ctx, msg.SourceTxHash, transferID)
	}
	if err := ms.Keeper.SubmitAttestation(ctx, transferID, msg.Validator, true); err != nil && err != types.ErrDuplicateAttestation {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if ms.Keeper.CheckAttestationThreshold(ctx, transferID) && transfer.Status != bridgepb.TransferStatus_COMPLETED {
		recipient, err := sdk.AccAddressFromBech32(transfer.Recipient)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		coin := sdk.NewCoin(msg.Denom, amount)
		if ms.Keeper.bankKeeper != nil {
			coins := sdk.NewCoins(coin)
			if err := ms.Keeper.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			if err := ms.Keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coins); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
		transfer.Status = bridgepb.TransferStatus_COMPLETED
		transfer.TargetTxHash = msg.SourceTxHash
		transfer.Timestamp = timestamppb.New(ctx.BlockTime())
		ms.Keeper.setTransfer(ctx, transfer)
	}
	wrappedDenom := fmt.Sprintf("%s.%s", normalizeChain(msg.SourceChain), msg.Denom)
	token, _ := ms.Keeper.getWrappedToken(ctx, wrappedDenom)
	if token == nil {
		token = &bridgepb.WrappedToken{
			WrappedDenom:  wrappedDenom,
			OriginalDenom: msg.Denom,
			SourceChain:   normalizeChain(msg.SourceChain),
			TotalSupply:   amount.String(),
		}
	} else {
		if total, ok := sdkmath.NewIntFromString(token.TotalSupply); ok {
			token.TotalSupply = total.Add(amount).String()
		}
	}
	ms.Keeper.setWrappedToken(ctx, token)
	return &bridgepb.MsgMintTokensResponse{Success: true, WrappedDenom: wrappedDenom}, nil
}

// UnlockTokens unlocks locked assets after a burn proof on the destination chain.
func (ms msgServer) UnlockTokens(goCtx context.Context, msg *bridgepb.MsgUnlockTokens) (*bridgepb.MsgUnlockTokensResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	amount, ok := sdkmath.NewIntFromString(msg.Amount)
	if !ok || !amount.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	transferID, hasIndex := ms.Keeper.transferIDByHash(ctx, msg.BurnTxHash)
	if !hasIndex {
		transferID = msg.BurnTxHash
	}
	transfer, found := ms.Keeper.getTransfer(ctx, transferID)
	if !found {
		return nil, status.Error(codes.NotFound, types.ErrTransferNotFound.Error())
	}
	required := ms.Keeper.GetParams(ctx).MinConfirmations
	if uint64(len(msg.ValidatorSignatures)) < required {
		return nil, status.Error(codes.FailedPrecondition, "insufficient signatures")
	}
	recipient, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	coin := sdk.NewCoin(msg.Denom, amount)
	if ms.Keeper.bankKeeper != nil {
		if err := ms.Keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(coin)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	transfer.Status = bridgepb.TransferStatus_COMPLETED
	transfer.TargetTxHash = msg.BurnTxHash
	transfer.Timestamp = timestamppb.New(ctx.BlockTime())
	ms.Keeper.setTransfer(ctx, transfer)
	return &bridgepb.MsgUnlockTokensResponse{Success: true}, nil
}

// BurnTokens burns wrapped tokens on Aura to unlock on the origin chain.
func (ms msgServer) BurnTokens(goCtx context.Context, msg *bridgepb.MsgBurnTokens) (*bridgepb.MsgBurnTokensResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	chainID := normalizeChain(msg.TargetChain)
	if chainID == "" {
		return nil, status.Error(codes.InvalidArgument, "target chain required")
	}
	if cfg, found := ms.Keeper.getChainConfig(ctx, chainID); !found {
		return nil, status.Error(codes.NotFound, types.ErrChainNotFound.Error())
	} else if !cfg.Enabled {
		return nil, status.Error(codes.FailedPrecondition, types.ErrChainDisabled.Error())
	}
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	amnt := msg.Amount
	if ms.Keeper.bankKeeper != nil {
		if err := ms.Keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(*amnt)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if err := ms.Keeper.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(*amnt)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	transferID := ms.Keeper.nextTransferID(ctx)
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  transferID,
		SourceChain: sourceChainAura,
		TargetChain: chainID,
		Sender:      msg.Sender,
		Recipient:   msg.Recipient,
		Amount:      amnt.Amount.String(),
		Denom:       amnt.Denom,
		Status:      bridgepb.TransferStatus_CONFIRMED,
		Timestamp:   timestamppb.New(ctx.BlockTime()),
	}
	ms.Keeper.setTransfer(ctx, transfer)
	ms.Keeper.indexTransferHash(ctx, transferID, transferID)
	return &bridgepb.MsgBurnTokensResponse{TransferId: transferID, EstimatedCompletion: 600}, nil
}

// LinkAddress links Aura/PAW/XAI addresses for shared identity.
func (ms msgServer) LinkAddress(goCtx context.Context, msg *bridgepb.MsgLinkAddress) (*bridgepb.MsgLinkAddressResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.AuraAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "aura address required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	linked := map[string]string{"aura": msg.AuraAddress}
	if msg.PawAddress != "" {
		linked["paw"] = msg.PawAddress
	}
	if msg.XaiAddress != "" {
		linked["xai"] = msg.XaiAddress
	}
	identity := &bridgepb.SharedIdentity{
		Address:         msg.AuraAddress,
		VerifiedAura:    msg.AuraAddress != "",
		VerifiedPaw:     msg.PawAddress != "",
		VerifiedXai:     msg.XaiAddress != "",
		LinkedAddresses: linked,
		VerifiedAt:      timestamppb.New(ctx.BlockTime()),
	}
	if ms.Keeper.vcKeeper != nil {
		identity.AuraIrScore = ms.Keeper.vcKeeper.GetIRScore(ctx, msg.AuraAddress)
		identity.ReputationScore = identity.AuraIrScore * 10
	}
	ms.Keeper.setSharedIdentity(ctx, identity)
	return &bridgepb.MsgLinkAddressResponse{Success: true, LinkedIdentityId: msg.AuraAddress}, nil
}

// CrossChainSwap stores metadata about a requested swap route.
func (ms msgServer) CrossChainSwap(goCtx context.Context, msg *bridgepb.MsgCrossChainSwap) (*bridgepb.MsgCrossChainSwapResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.InputCoin == nil || !msg.InputCoin.Amount.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "input amount required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	swapID := fmt.Sprintf("swap-%s", ms.Keeper.nextTransferID(ctx))
	swap := &bridgepb.CrossChainSwap{
		SwapId:      swapID,
		Sender:      msg.Sender,
		SourceChain: normalizeChain(msg.SourceChain),
		TargetChain: normalizeChain(msg.TargetChain),
		SourceCoin:  msg.InputCoin,
		TargetDenom: msg.TargetDenom,
		Status:      "pending",
		InitiatedAt: timestamppb.New(ctx.BlockTime()),
	}
	ms.Keeper.setSwap(ctx, swap)
	route := []string{normalizeChain(msg.SourceChain), normalizeChain(msg.TargetChain)}
	if msg.SourceChain == msg.TargetChain {
		route = []string{normalizeChain(msg.SourceChain)}
	}
	return &bridgepb.MsgCrossChainSwapResponse{SwapId: swapID, Route: route, EstimatedCompletion: 900}, nil
}

// RelayTransfer updates transfer state based on relayer reports.
func (ms msgServer) RelayTransfer(goCtx context.Context, msg *bridgepb.MsgRelayTransfer) (*bridgepb.MsgRelayTransferResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.TransferId == "" {
		return nil, status.Error(codes.InvalidArgument, "transfer id required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	transfer, found := ms.Keeper.getTransfer(ctx, msg.TransferId)
	if !found {
		return nil, status.Error(codes.NotFound, types.ErrTransferNotFound.Error())
	}
	switch strings.ToUpper(msg.Status) {
	case "PENDING":
		transfer.Status = bridgepb.TransferStatus_PENDING
	case "CONFIRMED":
		transfer.Status = bridgepb.TransferStatus_CONFIRMED
	case "RELAYED":
		transfer.Status = bridgepb.TransferStatus_RELAYED
	case "COMPLETED":
		transfer.Status = bridgepb.TransferStatus_COMPLETED
	case "FAILED":
		transfer.Status = bridgepb.TransferStatus_FAILED
	case "REFUNDED":
		transfer.Status = bridgepb.TransferStatus_REFUNDED
	}
	transfer.TargetTxHash = msg.TargetTxHash
	transfer.Timestamp = timestamppb.New(ctx.BlockTime())
	ms.Keeper.setTransfer(ctx, transfer)
	vol, _ := sdkmath.NewIntFromString(transfer.Amount)
	ms.Keeper.recordRelayerStats(ctx, msg.Relayer, true, vol)
	return &bridgepb.MsgRelayTransferResponse{Success: true}, nil
}
