// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

var _ cryptoproto.QueryServer = queryServer{}

type queryServer struct {
	cryptoproto.UnimplementedQueryServer
	Keeper *Keeper
}

func NewQueryServerImpl(keeper *Keeper) cryptoproto.QueryServer {
	return &queryServer{Keeper: keeper}
}

func (qs queryServer) Params(goCtx context.Context, req *cryptoproto.QueryParamsRequest) (*cryptoproto.QueryParamsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params, err := qs.Keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.QueryParamsResponse{Params: params}, nil
}

func (qs queryServer) KeyRotationSchedule(goCtx context.Context, req *cryptoproto.QueryKeyRotationScheduleRequest) (*cryptoproto.QueryKeyRotationScheduleResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	schedule, err := qs.Keeper.GetKeyRotationSchedule(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.QueryKeyRotationScheduleResponse{Schedule: schedule}, nil
}

func (qs queryServer) ThresholdScheme(goCtx context.Context, req *cryptoproto.QueryThresholdSchemeRequest) (*cryptoproto.QueryThresholdSchemeResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	scheme, err := qs.Keeper.GetThresholdScheme(ctx, req.SchemeId)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.QueryThresholdSchemeResponse{Scheme: scheme}, nil
}

func (qs queryServer) VerifyZKProof(goCtx context.Context, req *cryptoproto.QueryVerifyZKProofRequest) (*cryptoproto.QueryVerifyZKProofResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	verified, err := qs.Keeper.VerifyZKProof(ctx, req.ProofId, req.ProofData, req.PublicInputs)
	if err != nil {
		return &cryptoproto.QueryVerifyZKProofResponse{
			Valid:        false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &cryptoproto.QueryVerifyZKProofResponse{
		Valid:        verified,
		ErrorMessage: "",
	}, nil
}

func (qs queryServer) SecureEnclave(goCtx context.Context, req *cryptoproto.QuerySecureEnclaveRequest) (*cryptoproto.QuerySecureEnclaveResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	enclave, err := qs.Keeper.GetSecureEnclave(ctx, req.EnclaveId)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.QuerySecureEnclaveResponse{Enclave: enclave}, nil
}

func (qs queryServer) QuantumResistantKey(goCtx context.Context, req *cryptoproto.QueryQuantumResistantKeyRequest) (*cryptoproto.QueryQuantumResistantKeyResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	key, err := qs.Keeper.GetQuantumResistantKey(ctx, req.KeyId)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.QueryQuantumResistantKeyResponse{Key: key}, nil
}

func (qs queryServer) RandomSourceStatus(goCtx context.Context, req *cryptoproto.QueryRandomSourceStatusRequest) (*cryptoproto.QueryRandomSourceStatusResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	allSources := qs.Keeper.GetRandomSourceStatus(ctx)

	// Apply in-memory pagination
	total := uint64(len(allSources))
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

	// Calculate pagination bounds
	start := offset
	end := offset + limit
	if end > total {
		end = total
	}
	if start > total {
		start = total
	}

	// Apply pagination
	var paginatedSources []*cryptoproto.CryptoRandomSource
	for i := start; i < end; i++ {
		paginatedSources = append(paginatedSources, allSources[i])
	}

	// Ensure sources is never nil
	if paginatedSources == nil {
		paginatedSources = []*cryptoproto.CryptoRandomSource{}
	}

	return &cryptoproto.QueryRandomSourceStatusResponse{
		Sources:    paginatedSources,
		Pagination: &query.PageResponse{Total: total},
	}, nil
}

func (qs queryServer) CertificatePin(goCtx context.Context, req *cryptoproto.QueryCertificatePinRequest) (*cryptoproto.QueryCertificatePinResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pin, err := qs.Keeper.GetCertificatePin(ctx, req.Hostname)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.QueryCertificatePinResponse{Pin: pin}, nil
}
