// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

// MsgServer implements the inclusionroutines Msg service
type MsgServer struct {
	inclusionroutinespb.UnimplementedMsgServer
	keeper *Keeper
}

// NewMsgServer returns a new MsgServer
func NewMsgServer(k *Keeper) inclusionroutinespb.MsgServer {
	return &MsgServer{keeper: k}
}

var _ inclusionroutinespb.MsgServer = &MsgServer{}

func (m *MsgServer) CreateIR(goCtx context.Context, msg *inclusionroutinespb.MsgCreateIR) (*inclusionroutinespb.MsgCreateIRResponse, error) {
	if err := m.assertAuthority(msg.GetAuthority()); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	ir := types.IRDefinition{
		Id:               msg.GetId(),
		Name:             msg.GetName(),
		Arena:            msg.GetArena(),
		Description:      msg.GetDescription(),
		Score:            msg.GetScore(),
		PoiReward:        msg.GetPoiReward(),
		LocaleTags:       append([]string{}, msg.GetLocaleTags()...),
		PrivacyTier:      msg.GetPrivacyTier(),
		Version:          msg.GetVersion(),
		MetadataHash:     msg.GetMetadataHash(),
		ActivationHeight: msg.GetActivationHeight(),
		SunsetHeight:     msg.GetSunsetHeight(),
	}

	if err := m.keeper.CreateIR(ctx, ir); err != nil {
		return nil, mapIRAuthError(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeIRCreated,
			sdk.NewAttribute(types.AttributeKeyIRID, msg.GetId()),
			sdk.NewAttribute(types.AttributeKeyName, msg.GetName()),
			sdk.NewAttribute(types.AttributeKeyCreator, msg.GetAuthority()),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return &inclusionroutinespb.MsgCreateIRResponse{Id: msg.GetId()}, nil
}

func (m *MsgServer) UpdateIR(goCtx context.Context, msg *inclusionroutinespb.MsgUpdateIR) (*inclusionroutinespb.MsgUpdateIRResponse, error) {
	if err := m.assertAuthority(msg.GetAuthority()); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	existing, found := m.keeper.GetIR(ctx, msg.GetId())
	if !found {
		return nil, types.ErrIRNotFound
	}

	updated := existing

	if msg.GetName() != "" {
		updated.Name = msg.GetName()
	}
	if msg.GetDescription() != "" {
		updated.Description = msg.GetDescription()
	}
	if msg.GetScore() != 0 || existing.Score == 0 {
		updated.Score = msg.GetScore()
	}
	if msg.GetPoiReward() != 0 || existing.PoiReward == 0 {
		updated.PoiReward = msg.GetPoiReward()
	}
	if len(msg.GetLocaleTags()) > 0 {
		updated.LocaleTags = append([]string{}, msg.GetLocaleTags()...)
	}
	if msg.GetPrivacyTier() != inclusionroutinespb.PrivacyTier_PRIVACY_TIER_UNSPECIFIED {
		updated.PrivacyTier = msg.GetPrivacyTier()
	}
	if msg.GetVersion() != "" {
		updated.Version = msg.GetVersion()
	}
	if msg.GetMetadataHash() != "" {
		updated.MetadataHash = msg.GetMetadataHash()
	}
	if msg.GetSunsetHeight() != 0 || existing.SunsetHeight == 0 {
		updated.SunsetHeight = msg.GetSunsetHeight()
	}

	if err := m.keeper.UpdateIR(ctx, updated); err != nil {
		return nil, mapIRAuthError(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeIRUpdated,
			sdk.NewAttribute(types.AttributeKeyIRID, msg.GetId()),
			sdk.NewAttribute(types.AttributeKeyName, updated.Name),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return &inclusionroutinespb.MsgUpdateIRResponse{}, nil
}

func (m *MsgServer) DeleteIR(goCtx context.Context, msg *inclusionroutinespb.MsgDeleteIR) (*inclusionroutinespb.MsgDeleteIRResponse, error) {
	if err := m.assertAuthority(msg.GetAuthority()); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := m.keeper.DeleteIR(ctx, msg.GetId()); err != nil {
		return nil, mapIRAuthError(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeIRDeleted,
			sdk.NewAttribute(types.AttributeKeyIRID, msg.GetId()),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return &inclusionroutinespb.MsgDeleteIRResponse{}, nil
}

func (m *MsgServer) SetIRPrerequisites(goCtx context.Context, msg *inclusionroutinespb.MsgSetIRPrerequisites) (*inclusionroutinespb.MsgSetIRPrerequisitesResponse, error) {
	if err := m.assertAuthority(msg.GetAuthority()); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := m.keeper.SetPrerequisites(ctx, msg.GetIrId(), msg.GetRequiredIrIds()); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeIRPrerequisitesUpdated,
			sdk.NewAttribute(types.AttributeKeyIRID, msg.GetIrId()),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return &inclusionroutinespb.MsgSetIRPrerequisitesResponse{}, nil
}

func (m *MsgServer) SetIRRateLimit(goCtx context.Context, msg *inclusionroutinespb.MsgSetIRRateLimit) (*inclusionroutinespb.MsgSetIRRateLimitResponse, error) {
	if err := m.assertAuthority(msg.GetAuthority()); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	limit := types.IRRateLimit{
		IrId:             msg.GetIrId(),
		PerWalletPerHour: msg.GetPerWalletPerHour(),
		PerWalletPerDay:  msg.GetPerWalletPerDay(),
		PerBlockGlobal:   msg.GetPerBlockGlobal(),
	}
	if err := m.keeper.SetRateLimitConfig(ctx, limit); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeIRRateLimitUpdated,
			sdk.NewAttribute(types.AttributeKeyIRID, msg.GetIrId()),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return &inclusionroutinespb.MsgSetIRRateLimitResponse{}, nil
}

func (m *MsgServer) SuspendIR(goCtx context.Context, msg *inclusionroutinespb.MsgSuspendIR) (*inclusionroutinespb.MsgSuspendIRResponse, error) {
	if err := m.assertAuthority(msg.GetAuthority()); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := m.keeper.SuspendIR(ctx, msg.GetIrId()); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeIRDeactivated,
			sdk.NewAttribute(types.AttributeKeyIRID, msg.GetIrId()),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return &inclusionroutinespb.MsgSuspendIRResponse{}, nil
}

func (m *MsgServer) ActivateIR(goCtx context.Context, msg *inclusionroutinespb.MsgActivateIR) (*inclusionroutinespb.MsgActivateIRResponse, error) {
	if err := m.assertAuthority(msg.GetAuthority()); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := m.keeper.ActivateIR(ctx, msg.GetIrId()); err != nil {
		return nil, mapIRAuthError(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeIRActivated,
			sdk.NewAttribute(types.AttributeKeyIRID, msg.GetIrId()),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return &inclusionroutinespb.MsgActivateIRResponse{}, nil
}

func (m *MsgServer) assertAuthority(authority string) error {
	expected := m.keeper.GetAuthority()
	if expected == "" {
		return types.ErrInvalidAuthority
	}

	if authority == "" {
		return types.ErrInvalidAuthority
	}

	if authority != expected {
		return errorsmod.Wrapf(types.ErrUnauthorized, "expected %s, got %s", expected, authority)
	}

	return nil
}

func mapIRAuthError(err error) error {
	if err == nil {
		return nil
	}
	if errorsmod.IsOf(err, types.ErrUnauthorized) {
		return status.Error(codes.PermissionDenied, err.Error())
	}
	return fmt.Errorf("error in mapIRAuthError: %w", err)
}
