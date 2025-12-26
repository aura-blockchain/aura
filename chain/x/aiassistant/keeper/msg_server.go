// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

type msgServer struct {
	keeper *Keeper
	types.UnimplementedMsgServer
}

func NewMsgServer(k *Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

func (m msgServer) RegisterAssistant(goCtx context.Context, msg *types.MsgRegisterAssistant) (*types.MsgRegisterAssistantResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	assistant, err := m.keeper.RegisterAssistant(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgRegisterAssistantResponse{Assistant: assistant}, nil
}

func (m msgServer) UpdateLocales(goCtx context.Context, msg *types.MsgUpdateLocales) (*types.MsgUpdateLocalesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	assistant, err := m.keeper.UpdateLocales(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgUpdateLocalesResponse{Assistant: assistant}, nil
}

func (m msgServer) Heartbeat(goCtx context.Context, msg *types.MsgHeartbeat) (*types.MsgHeartbeatResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	nextSlash, err := m.keeper.Heartbeat(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgHeartbeatResponse{NextSlashTime: nextSlash}, nil
}

func (m msgServer) ReportMisbehavior(goCtx context.Context, msg *types.MsgReportMisbehavior) (*types.MsgReportMisbehaviorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	assistant, slashed, err := m.keeper.ReportMisbehavior(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgReportMisbehaviorResponse{
		Assistant:   assistant,
		SlashAmount: coinToBalance(slashed),
	}, nil
}

func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if msg.Authority != m.keeper.authority {
		return nil, status.Error(codes.PermissionDenied, errorsmod.Wrapf(types.ErrUnauthorizedOperator, "expected %s", m.keeper.authority).Error())
	}
	if err := types.ValidateParams(msg.Params); err != nil {
		return nil, err
	}
	if err := m.keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}
	return &types.MsgUpdateParamsResponse{}, nil
}
