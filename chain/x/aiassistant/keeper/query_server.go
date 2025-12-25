// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

type queryServer struct {
	keeper *Keeper
	types.UnimplementedQueryServer
}

func NewQueryServer(k *Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

func (q queryServer) Assistant(goCtx context.Context, req *types.QueryAssistantRequest) (*types.QueryAssistantResponse, error) {
	if req == nil || req.AssistantAddress == "" {
		return nil, types.ErrAssistantNotFound
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	assistant, err := q.keeper.mustGetAssistant(ctx, req.AssistantAddress)
	if err != nil {
		return nil, err
	}
	return &types.QueryAssistantResponse{Assistant: assistant}, nil
}

func (q queryServer) Assistants(goCtx context.Context, req *types.QueryAssistantsRequest) (*types.QueryAssistantsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	assistants, pageRes, err := q.keeper.ListAssistants(ctx, req.GetPagination())
	if err != nil {
		return nil, err
	}
	return &types.QueryAssistantsResponse{
		Assistants: assistants,
		Pagination: pageRes,
	}, nil
}

func (q queryServer) AssistantsByLocale(goCtx context.Context, req *types.QueryAssistantsByLocaleRequest) (*types.QueryAssistantsByLocaleResponse, error) {
	if req == nil || req.Locale == "" {
		return nil, types.ErrInvalidLocale
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	assistants, err := q.keeper.AssistantsByLocale(ctx, req.Locale)
	if err != nil {
		return nil, err
	}
	return &types.QueryAssistantsByLocaleResponse{
		Locale:     normalizeLocaleString(req.Locale),
		Assistants: assistants,
	}, nil
}

func (q queryServer) Params(goCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := q.keeper.GetParams(goCtx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: &params}, nil
}
