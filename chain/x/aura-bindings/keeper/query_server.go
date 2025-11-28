package keeper

import (
	"context"

	"github.com/aequitas/aura/chain/x/aura-bindings/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// QueryStats returns query usage statistics
func (qs queryServer) QueryStats(goCtx context.Context, req *types.QueryStatsRequest) (*types.QueryStatsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidParam
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	qs.Logger(ctx).Debug("querying stats")

	stats := qs.GetQueryStats()

	return &types.QueryStatsResponse{
		QueryStats: stats,
	}, nil
}

// MessageStats returns message usage statistics
func (qs queryServer) MessageStats(goCtx context.Context, req *types.MessageStatsRequest) (*types.MessageStatsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidParam
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	qs.Logger(ctx).Debug("querying message stats")

	stats := qs.GetMessageStats()

	return &types.MessageStatsResponse{
		MessageStats: stats,
	}, nil
}

// AllStats returns both query and message statistics
func (qs queryServer) AllStats(goCtx context.Context, req *types.AllStatsRequest) (*types.AllStatsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidParam
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	qs.Logger(ctx).Debug("querying all stats")

	queryStats := qs.GetQueryStats()
	messageStats := qs.GetMessageStats()

	return &types.AllStatsResponse{
		QueryStats:   queryStats,
		MessageStats: messageStats,
	}, nil
}
