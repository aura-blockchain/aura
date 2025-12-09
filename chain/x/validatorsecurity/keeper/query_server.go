package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1beta1 "github.com/aequitas/aura/proto/aura/validatorsecurity/v1beta1"
)

type queryServer struct {
	Keeper
	v1beta1.UnimplementedQueryServer
}

// NewQueryServerImpl returns a QueryServer backed by the keeper.
func NewQueryServerImpl(k Keeper) v1beta1.QueryServer {
	return &queryServer{Keeper: k}
}

var _ v1beta1.QueryServer = (*queryServer)(nil)

func (qs queryServer) Params(ctx context.Context, req *v1beta1.QueryParamsRequest) (*v1beta1.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	return &v1beta1.QueryParamsResponse{
		Params: *qs.GetParams(ctx),
	}, nil
}

func (qs queryServer) ValidatorSecurityInfo(ctx context.Context, req *v1beta1.QueryValidatorSecurityInfoRequest) (*v1beta1.QueryValidatorSecurityInfoResponse, error) {
	if req == nil || req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address required")
	}

	info, err := qs.GetValidatorSecurityInfo(ctx, req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	return &v1beta1.QueryValidatorSecurityInfoResponse{
		Info: info,
	}, nil
}

func (qs queryServer) AllValidators(ctx context.Context, req *v1beta1.QueryAllValidatorsRequest) (*v1beta1.QueryAllValidatorsResponse, error) {
	if req == nil {
		req = &v1beta1.QueryAllValidatorsRequest{}
	}
	validators := qs.GetAllValidators(ctx)
	return &v1beta1.QueryAllValidatorsResponse{Validators: validators}, nil
}

func (qs queryServer) JailedValidators(ctx context.Context, req *v1beta1.QueryJailedValidatorsRequest) (*v1beta1.QueryJailedValidatorsResponse, error) {
	if req == nil {
		req = &v1beta1.QueryJailedValidatorsRequest{}
	}
	jailed := qs.GetJailedValidators(ctx)
	return &v1beta1.QueryJailedValidatorsResponse{Validators: jailed}, nil
}

func (qs queryServer) TombstonedValidators(ctx context.Context, req *v1beta1.QueryTombstonedValidatorsRequest) (*v1beta1.QueryTombstonedValidatorsResponse, error) {
	if req == nil {
		req = &v1beta1.QueryTombstonedValidatorsRequest{}
	}
	tombstoned := qs.GetTombstonedValidators(ctx)
	return &v1beta1.QueryTombstonedValidatorsResponse{Validators: tombstoned}, nil
}

func (qs queryServer) DoubleSignEvidences(ctx context.Context, req *v1beta1.QueryDoubleSignEvidencesRequest) (*v1beta1.QueryDoubleSignEvidencesResponse, error) {
	if req == nil {
		req = &v1beta1.QueryDoubleSignEvidencesRequest{}
	}
	evidences := qs.GetAllDoubleSignEvidences(ctx)
	return &v1beta1.QueryDoubleSignEvidencesResponse{Evidences: evidences}, nil
}

func (qs queryServer) ValidatorAlerts(ctx context.Context, req *v1beta1.QueryValidatorAlertsRequest) (*v1beta1.QueryValidatorAlertsResponse, error) {
	if req == nil || req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address required")
	}
	alerts := qs.GetValidatorAlerts(ctx, req.ValidatorAddress)
	return &v1beta1.QueryValidatorAlertsResponse{Alerts: alerts}, nil
}

func (qs queryServer) SentryNodes(ctx context.Context, req *v1beta1.QuerySentryNodesRequest) (*v1beta1.QuerySentryNodesResponse, error) {
	if req == nil || req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address required")
	}
	nodes := qs.GetValidatorSentryNodes(ctx, req.ValidatorAddress)
	return &v1beta1.QuerySentryNodesResponse{Nodes: nodes}, nil
}
