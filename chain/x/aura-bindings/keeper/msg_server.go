// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	"github.com/aequitas/aura/chain/x/aura-bindings/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type msgServer struct {
	Keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// Note: The aura-bindings module primarily serves as a bridge for CosmWasm
// custom bindings and doesn't define its own Msg types. The actual message
// handling is done through the message_plugin.go which processes custom
// CosmWasm messages. This file exists for module completeness and to satisfy
// the module interface requirements.

// If custom Msg types are added in the future, they should be handled here.
// For now, this serves as a placeholder for any future direct message handling
// that may be needed outside of the CosmWasm binding context.

// Example placeholder method (can be removed if not needed):
func (ms msgServer) EmptyMethod(goCtx context.Context, req *types.EmptyRequest) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ms.Keeper.Logger(ctx).Debug("empty method called")

	return &types.EmptyResponse{}, nil
}
