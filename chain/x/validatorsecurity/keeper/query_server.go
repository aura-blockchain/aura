package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

type queryServer struct {
	types.UnimplementedQueryServer
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// Params queries module parameters
func (qs queryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := qs.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// ValidatorSecurityInfo queries security info for a validator
func (qs queryServer) ValidatorSecurityInfo(goCtx context.Context, req *types.QueryValidatorSecurityInfoRequest) (*types.QueryValidatorSecurityInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	info, err := qs.Keeper.GetValidatorSecurityInfo(ctx, req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryValidatorSecurityInfoResponse{Info: info}, nil
}

// AllValidators queries all validator security info
func (qs queryServer) AllValidators(goCtx context.Context, req *types.QueryAllValidatorsRequest) (*types.QueryAllValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validators := qs.Keeper.GetAllValidators(ctx)

	return &types.QueryAllValidatorsResponse{Validators: validators}, nil
}

// JailedValidators queries all jailed validators
func (qs queryServer) JailedValidators(goCtx context.Context, req *types.QueryJailedValidatorsRequest) (*types.QueryJailedValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validators := qs.Keeper.GetJailedValidators(ctx)

	return &types.QueryJailedValidatorsResponse{Validators: validators}, nil
}

// TombstonedValidators queries all tombstoned validators
func (qs queryServer) TombstonedValidators(goCtx context.Context, req *types.QueryTombstonedValidatorsRequest) (*types.QueryTombstonedValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validators := qs.Keeper.GetTombstonedValidators(ctx)

	return &types.QueryTombstonedValidatorsResponse{Validators: validators}, nil
}

// DoubleSignEvidences queries all double sign evidences
func (qs queryServer) DoubleSignEvidences(goCtx context.Context, req *types.QueryDoubleSignEvidencesRequest) (*types.QueryDoubleSignEvidencesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	evidences := qs.Keeper.GetAllDoubleSignEvidences(ctx)

	return &types.QueryDoubleSignEvidencesResponse{Evidences: evidences}, nil
}

// ValidatorAlerts queries alerts for a validator
func (qs queryServer) ValidatorAlerts(goCtx context.Context, req *types.QueryValidatorAlertsRequest) (*types.QueryValidatorAlertsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	alerts := qs.Keeper.GetValidatorAlerts(ctx, req.ValidatorAddress)

	return &types.QueryValidatorAlertsResponse{Alerts: alerts}, nil
}

// SentryNodes queries sentry nodes for a validator
func (qs queryServer) SentryNodes(goCtx context.Context, req *types.QuerySentryNodesRequest) (*types.QuerySentryNodesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	nodes := qs.Keeper.GetValidatorSentryNodes(ctx, req.ValidatorAddress)

	return &types.QuerySentryNodesResponse{Nodes: nodes}, nil
}
