// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1beta1 "github.com/aequitas/aura/proto/aura/validatorsecurity/v1beta1"
	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
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
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.storeKey)
	validatorStore := prefix.NewStore(store, types.ValidatorSecurityInfoKey)

	var validators []v1beta1.ValidatorSecurityInfo
	pageRes, err := query.Paginate(validatorStore, req.Pagination, func(key, value []byte) error {
		var info v1beta1.ValidatorSecurityInfo
		if err := qs.cdc.Unmarshal(value, &info); err != nil {
			return fmt.Errorf("failed to unmarshal validator security info: %w", err)
		}
		validators = append(validators, info)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1beta1.QueryAllValidatorsResponse{
		Validators: validators,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) JailedValidators(ctx context.Context, req *v1beta1.QueryJailedValidatorsRequest) (*v1beta1.QueryJailedValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Get all jailed validators and apply in-memory pagination
	allJailed := qs.GetJailedValidators(ctx)
	total := uint64(len(allJailed))

	// Parse pagination params
	var offset, limit uint64
	if req.Pagination != nil {
		offset = req.Pagination.Offset
		limit = req.Pagination.Limit
		if limit == 0 {
			limit = query.DefaultLimit
		}
	} else {
		limit = query.DefaultLimit
	}

	// Apply offset and limit
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var validators []v1beta1.ValidatorSecurityInfo
	for i := start; i < end; i++ {
		validators = append(validators, allJailed[i])
	}

	pageRes := &query.PageResponse{Total: total}

	return &v1beta1.QueryJailedValidatorsResponse{
		Validators: validators,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) TombstonedValidators(ctx context.Context, req *v1beta1.QueryTombstonedValidatorsRequest) (*v1beta1.QueryTombstonedValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Get all tombstoned validators and apply in-memory pagination
	allTombstoned := qs.GetTombstonedValidators(ctx)
	total := uint64(len(allTombstoned))

	// Parse pagination params
	var offset, limit uint64
	if req.Pagination != nil {
		offset = req.Pagination.Offset
		limit = req.Pagination.Limit
		if limit == 0 {
			limit = query.DefaultLimit
		}
	} else {
		limit = query.DefaultLimit
	}

	// Apply offset and limit
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var validators []v1beta1.ValidatorSecurityInfo
	for i := start; i < end; i++ {
		validators = append(validators, allTombstoned[i])
	}

	pageRes := &query.PageResponse{Total: total}

	return &v1beta1.QueryTombstonedValidatorsResponse{
		Validators: validators,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) DoubleSignEvidences(ctx context.Context, req *v1beta1.QueryDoubleSignEvidencesRequest) (*v1beta1.QueryDoubleSignEvidencesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.storeKey)
	evidenceStore := prefix.NewStore(store, types.DoubleSignEvidenceKey)

	var evidences []v1beta1.DoubleSignEvidence
	pageRes, err := query.Paginate(evidenceStore, req.Pagination, func(key, value []byte) error {
		var evidence v1beta1.DoubleSignEvidence
		if err := qs.cdc.Unmarshal(value, &evidence); err != nil {
			return fmt.Errorf("failed to unmarshal double sign evidence: %w", err)
		}
		evidences = append(evidences, evidence)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1beta1.QueryDoubleSignEvidencesResponse{
		Evidences:  evidences,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) ValidatorAlerts(ctx context.Context, req *v1beta1.QueryValidatorAlertsRequest) (*v1beta1.QueryValidatorAlertsResponse, error) {
	if req == nil || req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address required")
	}

	// Get all alerts for validator and apply in-memory pagination
	allAlerts := qs.GetValidatorAlerts(ctx, req.ValidatorAddress)
	total := uint64(len(allAlerts))

	// Parse pagination params
	var offset, limit uint64
	if req.Pagination != nil {
		offset = req.Pagination.Offset
		limit = req.Pagination.Limit
		if limit == 0 {
			limit = query.DefaultLimit
		}
	} else {
		limit = query.DefaultLimit
	}

	// Apply offset and limit
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var alerts []v1beta1.ValidatorAlert
	for i := start; i < end; i++ {
		alerts = append(alerts, allAlerts[i])
	}

	pageRes := &query.PageResponse{Total: total}

	return &v1beta1.QueryValidatorAlertsResponse{
		Alerts:     alerts,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) SentryNodes(ctx context.Context, req *v1beta1.QuerySentryNodesRequest) (*v1beta1.QuerySentryNodesResponse, error) {
	if req == nil || req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address required")
	}

	// Get all sentry nodes for validator and apply in-memory pagination
	allNodes := qs.GetValidatorSentryNodes(ctx, req.ValidatorAddress)
	total := uint64(len(allNodes))

	// Parse pagination params
	var offset, limit uint64
	if req.Pagination != nil {
		offset = req.Pagination.Offset
		limit = req.Pagination.Limit
		if limit == 0 {
			limit = query.DefaultLimit
		}
	} else {
		limit = query.DefaultLimit
	}

	// Apply offset and limit
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var nodes []v1beta1.SentryNodeInfo
	for i := start; i < end; i++ {
		nodes = append(nodes, allNodes[i])
	}

	pageRes := &query.PageResponse{Total: total}

	return &v1beta1.QuerySentryNodesResponse{
		Nodes:      nodes,
		Pagination: pageRes,
	}, nil
}
