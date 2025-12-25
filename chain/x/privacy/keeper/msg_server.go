// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/privacy/types"
	privacypb "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

var _ privacypb.MsgServer = (*msgServer)(nil)

type msgServer struct {
	privacypb.UnimplementedMsgServer
	Keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper *Keeper) privacypb.MsgServer {
	return &msgServer{Keeper: keeper}
}

// SubmitPrivateTransaction submits a private transaction
func (ms msgServer) SubmitPrivateTransaction(goCtx context.Context, msg *privacypb.MsgSubmitPrivateTransaction) (*privacypb.MsgSubmitPrivateTransactionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Sender == "" {
		return nil, status.Error(codes.InvalidArgument, "sender cannot be empty")
	}

	if msg.PrivateTransaction == nil {
		return nil, status.Error(codes.InvalidArgument, "private transaction cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate privacy features are enabled
	params := ms.Keeper.GetParams(ctx)
	if !params.EnableZkProofs && msg.PrivateTransaction.ZkProof != nil {
		return nil, status.Error(codes.FailedPrecondition, "zk proofs not enabled")
	}

	// Store private transaction (simplified)
	txHash := []byte(fmt.Sprintf("tx_%s_%d", msg.Sender, ctx.BlockHeight()))

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePrivateTransaction,
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyTxHash, string(txHash)),
		),
	)

	return &privacypb.MsgSubmitPrivateTransactionResponse{
		TxHash:  txHash,
		Success: true,
	}, nil
}

// CreateMixingPool creates a new coin mixing pool
func (ms msgServer) CreateMixingPool(goCtx context.Context, msg *privacypb.MsgCreateMixingPool) (*privacypb.MsgCreateMixingPoolResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Creator == "" {
		return nil, status.Error(codes.InvalidArgument, "creator cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate mixing is enabled
	params := ms.Keeper.GetParams(ctx)
	if !params.EnableMixing {
		return nil, status.Error(codes.FailedPrecondition, "mixing not enabled")
	}

	// Validate participant counts
	if msg.MinParticipants < 2 {
		return nil, status.Error(codes.InvalidArgument, "minimum participants must be at least 2")
	}

	if msg.MaxParticipants < msg.MinParticipants {
		return nil, status.Error(codes.InvalidArgument, "max participants must be >= min participants")
	}

	// Create mixing pool
	poolID := fmt.Sprintf("pool_%s_%d", msg.Creator, ctx.BlockHeight())

	pool := &privacypb.MixingPool{
		PoolId:          poolID,
		MinParticipants: msg.MinParticipants,
		MaxParticipants: msg.MaxParticipants,
		Denomination:    msg.Denomination,
		MixingRounds:    msg.MixingRounds,
		Status:          "pending",
		Participants:    [][]byte{},
	}

	// Store mixing pool
	if err := ms.Keeper.SetMixingPool(ctx, pool); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMixingPool,
			sdk.NewAttribute("pool_id", poolID),
			sdk.NewAttribute("creator", msg.Creator),
		),
	)

	return &privacypb.MsgCreateMixingPoolResponse{PoolId: poolID}, nil
}

// JoinMixingPool joins an existing mixing pool
func (ms msgServer) JoinMixingPool(goCtx context.Context, msg *privacypb.MsgJoinMixingPool) (*privacypb.MsgJoinMixingPoolResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Participant == "" {
		return nil, status.Error(codes.InvalidArgument, "participant cannot be empty")
	}

	if msg.PoolId == "" {
		return nil, status.Error(codes.InvalidArgument, "pool id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get mixing pool
	pool, err := ms.Keeper.GetMixingPool(ctx, msg.PoolId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "mixing pool not found")
	}

	// Check if pool is full
	if uint32(len(pool.Participants)) >= pool.MaxParticipants {
		return nil, status.Error(codes.FailedPrecondition, "mixing pool is full")
	}

	// Check if already participating
	participantBytes := []byte(msg.Participant)
	for _, p := range pool.Participants {
		if string(p) == msg.Participant {
			return nil, status.Error(codes.AlreadyExists, "already participating in pool")
		}
	}

	// Add participant
	pool.Participants = append(pool.Participants, participantBytes)
	participantIndex := uint32(len(pool.Participants) - 1)

	// Update pool status if minimum reached
	if uint32(len(pool.Participants)) >= pool.MinParticipants {
		pool.Status = "ready"
	}

	// Store updated pool
	if err := ms.Keeper.SetMixingPool(ctx, pool); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMixingPool,
			sdk.NewAttribute("pool_id", msg.PoolId),
			sdk.NewAttribute("participant", msg.Participant),
		),
	)

	return &privacypb.MsgJoinMixingPoolResponse{
		Success:          true,
		ParticipantIndex: participantIndex,
	}, nil
}

// RegisterViewKey registers a new view key
// SECURITY: Only public keys are stored. Private keys must never be transmitted or stored on-chain.
func (ms msgServer) RegisterViewKey(goCtx context.Context, msg *privacypb.MsgRegisterViewKey) (*privacypb.MsgRegisterViewKeyResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Owner == "" {
		return nil, status.Error(codes.InvalidArgument, "owner cannot be empty")
	}

	if msg.ViewKey == nil {
		return nil, status.Error(codes.InvalidArgument, "view key cannot be nil")
	}

	// Validate public view key is present and valid
	if len(msg.ViewKey.PublicViewKey) == 0 {
		return nil, status.Error(codes.InvalidArgument, "public view key cannot be empty")
	}

	// Validate key length (Ed25519/Curve25519 keys are 32 bytes, compressed secp256k1 is 33 bytes)
	keyLen := len(msg.ViewKey.PublicViewKey)
	if keyLen != 32 && keyLen != 33 && keyLen != 64 {
		return nil, status.Error(codes.InvalidArgument, "invalid public key length (must be 32, 33, or 64 bytes)")
	}

	// SECURITY CHECK: Ensure no private key material is being stored
	// The ViewKey proto no longer has private_view_key field, but we add this
	// defensive check in case someone tries to abuse the system
	if msg.ViewKey.KeyType == "PRIVATE" || msg.ViewKey.KeyType == "SECRET" {
		return nil, status.Error(codes.InvalidArgument, "private keys cannot be registered on-chain")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Store view key (only public component)
	if err := ms.Keeper.SetViewKey(ctx, msg.Owner, msg.ViewKey); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeViewKey,
			sdk.NewAttribute("owner", msg.Owner),
			sdk.NewAttribute("key_type", msg.ViewKey.KeyType),
		),
	)

	return &privacypb.MsgRegisterViewKeyResponse{Success: true}, nil
}

// RevokeViewKey revokes an existing view key
func (ms msgServer) RevokeViewKey(goCtx context.Context, msg *privacypb.MsgRevokeViewKey) (*privacypb.MsgRevokeViewKeyResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Owner == "" {
		return nil, status.Error(codes.InvalidArgument, "owner cannot be empty")
	}

	if len(msg.PublicViewKey) == 0 {
		return nil, status.Error(codes.InvalidArgument, "public view key cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Delete view key
	if err := ms.Keeper.DeleteViewKey(ctx, msg.Owner, msg.PublicViewKey); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeViewKey,
			sdk.NewAttribute("owner", msg.Owner),
		),
	)

	return &privacypb.MsgRevokeViewKeyResponse{Success: true}, nil
}

// UpdateNetworkPrivacy updates network privacy settings
func (ms msgServer) UpdateNetworkPrivacy(goCtx context.Context, msg *privacypb.MsgUpdateNetworkPrivacy) (*privacypb.MsgUpdateNetworkPrivacyResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Sender == "" {
		return nil, status.Error(codes.InvalidArgument, "sender cannot be empty")
	}

	if msg.NetworkPrivacy == nil {
		return nil, status.Error(codes.InvalidArgument, "network privacy cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate network privacy is enabled
	params := ms.Keeper.GetParams(ctx)
	if !params.EnableNetworkPrivacy {
		return nil, status.Error(codes.FailedPrecondition, "network privacy not enabled")
	}

	// Store network privacy settings
	if err := ms.Keeper.SetNetworkPrivacy(ctx, msg.NetworkPrivacy); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNetworkPrivacy,
			sdk.NewAttribute("sender", msg.Sender),
		),
	)

	return &privacypb.MsgUpdateNetworkPrivacyResponse{Success: true}, nil
}

// UpdateParams updates module parameters
func (ms msgServer) UpdateParams(goCtx context.Context, msg *privacypb.MsgUpdateParams) (*privacypb.MsgUpdateParamsResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Authority == "" {
		return nil, status.Error(codes.InvalidArgument, "authority cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority (simplified - in production, check against module authority)
	if msg.Authority != ms.Keeper.GetAuthority() && ms.Keeper.GetAuthority() != "" {
		return nil, status.Error(codes.PermissionDenied, "unauthorized")
	}

	// Convert proto params to types params
	params := types.Params{
		EnableZkProofs:                 msg.Params.EnableZkProofs,
		EnableStealthAddresses:         msg.Params.EnableStealthAddresses,
		EnableRingSignatures:           msg.Params.EnableRingSignatures,
		EnableConfidentialTransactions: msg.Params.EnableConfidentialTransactions,
		EnableNetworkPrivacy:           msg.Params.EnableNetworkPrivacy,
		EnableMixing:                   msg.Params.EnableMixing,
		MinRingSize:                    msg.Params.MinRingSize,
		MaxRingSize:                    msg.Params.MaxRingSize,
		MinMixingParticipants:          msg.Params.MinMixingParticipants,
		MixingFee:                      msg.Params.MixingFee.String(),
		ZkProofVerificationCost:        msg.Params.ZkProofVerificationCost,
	}

	// Update params
	if err := ms.Keeper.SetParams(ctx, params); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateParams,
			sdk.NewAttribute("authority", msg.Authority),
		),
	)

	return &privacypb.MsgUpdateParamsResponse{}, nil
}
