// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
	v1beta1 "github.com/aequitas/aura/proto/aura/validatorsecurity/v1beta1"
)

type msgServer struct {
	Keeper
	v1beta1.UnimplementedMsgServer
}

// NewMsgServerImpl wires the keeper into the generated Msg service.
func NewMsgServerImpl(k Keeper) v1beta1.MsgServer {
	return &msgServer{Keeper: k}
}

var _ v1beta1.MsgServer = (*msgServer)(nil)

func (ms msgServer) RegisterValidator(goCtx context.Context, msg *v1beta1.MsgRegisterValidator) (*v1beta1.MsgRegisterValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := ms.Keeper.RegisterValidator(
		ctx,
		msg.ValidatorAddress,
		msg.HotKey,
		msg.ColdKey,
		msg.Region,
		msg.CountryCode,
		msg.Latitude,
		msg.Longitude,
		msg.BackupValidatorAddresses,
	); err != nil {
		return nil, err
	}

	return &v1beta1.MsgRegisterValidatorResponse{}, nil
}

func (ms msgServer) UpdateSecurityInfo(goCtx context.Context, msg *v1beta1.MsgUpdateSecurityInfo) (*v1beta1.MsgUpdateSecurityInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := ms.Keeper.UpdateValidatorSecurityInfo(
		ctx,
		msg.ValidatorAddress,
		msg.HotKey,
		msg.ColdKey,
		msg.Region,
		msg.CountryCode,
		msg.Latitude,
		msg.Longitude,
		msg.BackupValidatorAddresses,
	); err != nil {
		return nil, err
	}

	return &v1beta1.MsgUpdateSecurityInfoResponse{}, nil
}

func (ms msgServer) RegisterSentryNode(goCtx context.Context, msg *v1beta1.MsgRegisterSentryNode) (*v1beta1.MsgRegisterSentryNodeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := ms.Keeper.RegisterSentryNode(ctx, msg.ValidatorAddress, msg.SentryAddress, msg.IpAddress, msg.Port); err != nil {
		return nil, err
	}

	return &v1beta1.MsgRegisterSentryNodeResponse{}, nil
}

func (ms msgServer) ReportDoubleSign(goCtx context.Context, msg *v1beta1.MsgReportDoubleSign) (*v1beta1.MsgReportDoubleSignResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	slashAmount, err := ms.Keeper.HandleDoubleSign(ctx, msg.ValidatorAddress, msg.Height, msg.VoteA, msg.VoteB)
	if err != nil {
		return nil, err
	}

	return &v1beta1.MsgReportDoubleSignResponse{
		Slashed:     slashAmount.IsPositive(),
		SlashAmount: slashAmount.String(),
	}, nil
}

func (ms msgServer) Unjail(goCtx context.Context, msg *v1beta1.MsgUnjail) (*v1beta1.MsgUnjailResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := ms.Keeper.UnjailValidator(ctx, msg.ValidatorAddress); err != nil {
		return nil, err
	}

	return &v1beta1.MsgUnjailResponse{}, nil
}

func (ms msgServer) AcknowledgeAlert(goCtx context.Context, msg *v1beta1.MsgAcknowledgeAlert) (*v1beta1.MsgAcknowledgeAlertResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := ms.Keeper.AcknowledgeAlert(ctx, msg.AlertId, msg.AcknowledgerAddress); err != nil {
		return nil, err
	}

	return &v1beta1.MsgAcknowledgeAlertResponse{}, nil
}

func (ms msgServer) UpdateParams(goCtx context.Context, msg *v1beta1.MsgUpdateParams) (*v1beta1.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != ms.GetAuthority() {
		return nil, fmt.Errorf("unauthorized authority: %s", msg.Authority)
	}

	if err := types.ValidateParams(&msg.Params); err != nil {
		return nil, err
	}

	if err := ms.Keeper.SetParams(ctx, &msg.Params); err != nil {
		return nil, err
	}

	return &v1beta1.MsgUpdateParamsResponse{}, nil
}
