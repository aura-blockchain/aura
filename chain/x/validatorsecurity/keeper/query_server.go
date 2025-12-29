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

// Query limits to prevent DoS via unbounded iteration
const (
	// maxJailedValidatorQuery limits jailed validator queries
	maxJailedValidatorQuery = 500
	// maxTombstonedValidatorQuery limits tombstoned validator queries
	maxTombstonedValidatorQuery = 500
	// maxAlertQuery limits alert queries per validator
	maxAlertQuery = 100
	// maxSentryNodeQuery limits sentry node queries per validator
	maxSentryNodeQuery = 50
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

// JailedValidators queries jailed validators with store-based pagination.
// Uses direct iterator with early termination to avoid loading all validators into memory.
func (qs queryServer) JailedValidators(ctx context.Context, req *v1beta1.QueryJailedValidatorsRequest) (*v1beta1.QueryJailedValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.storeKey)
	jailedStore := prefix.NewStore(store, types.JailedValidatorsKey)

	var validators []v1beta1.ValidatorSecurityInfo
	count := 0

	pageRes, err := query.Paginate(jailedStore, req.Pagination, func(key, value []byte) error {
		// Hard cap to prevent DoS
		if count >= maxJailedValidatorQuery {
			return nil
		}
		count++

		// Key contains the validator address
		validatorAddr := string(key)
		info, err := qs.GetValidatorSecurityInfo(ctx, validatorAddr)
		if err != nil {
			return nil // Skip invalid entries
		}
		if info.IsJailed {
			validators = append(validators, info)
		}
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1beta1.QueryJailedValidatorsResponse{
		Validators: validators,
		Pagination: pageRes,
	}, nil
}

// TombstonedValidators queries tombstoned validators with store-based pagination.
// Uses direct iterator with early termination to avoid loading all validators into memory.
func (qs queryServer) TombstonedValidators(ctx context.Context, req *v1beta1.QueryTombstonedValidatorsRequest) (*v1beta1.QueryTombstonedValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.storeKey)
	tombstonedStore := prefix.NewStore(store, types.TombstonedValidatorsKey)

	var validators []v1beta1.ValidatorSecurityInfo
	count := 0

	pageRes, err := query.Paginate(tombstonedStore, req.Pagination, func(key, value []byte) error {
		// Hard cap to prevent DoS
		if count >= maxTombstonedValidatorQuery {
			return nil
		}
		count++

		// Key contains the validator address
		validatorAddr := string(key)
		info, err := qs.GetValidatorSecurityInfo(ctx, validatorAddr)
		if err != nil {
			return nil // Skip invalid entries
		}
		if info.IsTombstoned {
			validators = append(validators, info)
		}
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

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

// ValidatorAlerts queries alerts for a validator with store-based pagination.
// Uses the secondary index for efficient lookup by validator address.
func (qs queryServer) ValidatorAlerts(ctx context.Context, req *v1beta1.QueryValidatorAlertsRequest) (*v1beta1.QueryValidatorAlertsResponse, error) {
	if req == nil || req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.storeKey)

	// Use secondary index prefix for this validator's alerts
	alertPrefix := types.GetValidatorAlertByAddrPrefix(req.ValidatorAddress)
	alertStore := prefix.NewStore(store, alertPrefix)

	var alerts []v1beta1.ValidatorAlert
	count := 0

	pageRes, err := query.Paginate(alertStore, req.Pagination, func(key, value []byte) error {
		// Hard cap to prevent DoS
		if count >= maxAlertQuery {
			return nil
		}
		count++

		// Value contains the alert ID - fetch full alert
		alertID := string(value)
		alertKey := types.GetValidatorAlertKey(alertID)
		bz := store.Get(alertKey)
		if bz == nil {
			return nil // Stale index entry
		}

		var alert v1beta1.ValidatorAlert
		if err := qs.cdc.Unmarshal(bz, &alert); err != nil {
			return nil // Skip invalid entries
		}
		alerts = append(alerts, alert)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1beta1.QueryValidatorAlertsResponse{
		Alerts:     alerts,
		Pagination: pageRes,
	}, nil
}

// SentryNodes queries sentry nodes for a validator with store-based pagination.
// Uses bounded iteration with early termination.
func (qs queryServer) SentryNodes(ctx context.Context, req *v1beta1.QuerySentryNodesRequest) (*v1beta1.QuerySentryNodesResponse, error) {
	if req == nil || req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.storeKey)
	sentryStore := prefix.NewStore(store, types.SentryNodeInfoKey)

	var nodes []v1beta1.SentryNodeInfo
	count := 0
	matchCount := 0

	pageRes, err := query.Paginate(sentryStore, req.Pagination, func(key, value []byte) error {
		// Hard cap total iterations to prevent DoS
		count++
		if count > maxSentryNodeQuery*10 { // Allow some scan overhead for filtering
			return nil
		}

		var node v1beta1.SentryNodeInfo
		if err := qs.cdc.Unmarshal(value, &node); err != nil {
			return nil // Skip invalid entries
		}

		// Filter by validator address
		if node.ValidatorAddress != req.ValidatorAddress {
			return nil
		}

		// Cap matched results
		if matchCount >= maxSentryNodeQuery {
			return nil
		}
		matchCount++
		nodes = append(nodes, node)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1beta1.QuerySentryNodesResponse{
		Nodes:      nodes,
		Pagination: pageRes,
	}, nil
}
