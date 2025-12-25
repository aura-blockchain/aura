// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/privacy/types"
	privacypb "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

var _ privacypb.QueryServer = (*queryServer)(nil)

type queryServer struct {
	privacypb.UnimplementedQueryServer
	Keeper *Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper *Keeper) privacypb.QueryServer {
	return &queryServer{Keeper: keeper}
}

// Params queries the module parameters
func (qs queryServer) Params(goCtx context.Context, req *privacypb.QueryParamsRequest) (*privacypb.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := qs.Keeper.GetParams(ctx)

	// Convert MixingFee from string to math.Int
	mixingFee, ok := sdkmath.NewIntFromString(params.MixingFee)
	if !ok {
		mixingFee = sdkmath.ZeroInt()
	}

	// Convert to proto params
	protoParams := privacypb.Params{
		EnableZkProofs:                 params.EnableZkProofs,
		EnableStealthAddresses:         params.EnableStealthAddresses,
		EnableRingSignatures:           params.EnableRingSignatures,
		EnableConfidentialTransactions: params.EnableConfidentialTransactions,
		EnableNetworkPrivacy:           params.EnableNetworkPrivacy,
		EnableMixing:                   params.EnableMixing,
		MinRingSize:                    params.MinRingSize,
		MaxRingSize:                    params.MaxRingSize,
		MinMixingParticipants:          params.MinMixingParticipants,
		MixingFee:                      mixingFee,
		ZkProofVerificationCost:        params.ZkProofVerificationCost,
	}

	return &privacypb.QueryParamsResponse{Params: protoParams}, nil
}

// MixingPool queries a specific mixing pool
func (qs queryServer) MixingPool(goCtx context.Context, req *privacypb.QueryMixingPoolRequest) (*privacypb.QueryMixingPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.PoolId == "" {
		return nil, status.Error(codes.InvalidArgument, "pool id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pool, err := qs.Keeper.GetMixingPool(ctx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "mixing pool not found")
	}

	return &privacypb.QueryMixingPoolResponse{MixingPool: pool}, nil
}

// MixingPools queries all mixing pools with pagination
func (qs queryServer) MixingPools(goCtx context.Context, req *privacypb.QueryMixingPoolsRequest) (*privacypb.QueryMixingPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)
	poolStore := prefix.NewStore(store, types.MixingPoolPrefix)

	// Initialize to empty slice (not nil) so response always has a valid array
	pools := make([]*privacypb.MixingPool, 0)
	pageRes, err := query.Paginate(poolStore, req.Pagination, func(key, value []byte) error {
		var pool privacypb.MixingPool
		if err := qs.Keeper.cdc.Unmarshal(value, &pool); err != nil {
			return err
		}
		// Filter by status if provided
		if req.Status != "" && pool.Status != req.Status {
			return nil
		}
		pools = append(pools, &pool)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &privacypb.QueryMixingPoolsResponse{
		MixingPools: pools,
		Pagination:  pageRes,
	}, nil
}

// ViewKey queries a specific view key
func (qs queryServer) ViewKey(goCtx context.Context, req *privacypb.QueryViewKeyRequest) (*privacypb.QueryViewKeyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.PublicViewKey) == 0 {
		return nil, status.Error(codes.InvalidArgument, "public view key cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	viewKey, err := qs.Keeper.GetViewKeyByPublic(ctx, req.PublicViewKey)
	if err != nil {
		return nil, status.Error(codes.NotFound, "view key not found")
	}

	return &privacypb.QueryViewKeyResponse{ViewKey: viewKey}, nil
}

// ViewKeys queries all view keys for an address with pagination
func (qs queryServer) ViewKeys(goCtx context.Context, req *privacypb.QueryViewKeysRequest) (*privacypb.QueryViewKeysResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)
	// Create prefix for this address's view keys
	addressPrefix := append(types.ViewKeyPrefix, []byte(req.Address)...)
	viewKeyStore := prefix.NewStore(store, addressPrefix)

	// Initialize to empty slice (not nil) so response always has a valid array
	viewKeys := make([]*privacypb.ViewKey, 0)
	pageRes, err := query.Paginate(viewKeyStore, req.Pagination, func(key, value []byte) error {
		var viewKey privacypb.ViewKey
		if err := qs.Keeper.cdc.Unmarshal(value, &viewKey); err != nil {
			return err
		}
		viewKeys = append(viewKeys, &viewKey)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &privacypb.QueryViewKeysResponse{
		ViewKeys:   viewKeys,
		Pagination: pageRes,
	}, nil
}

// VerifyZKProof verifies a zero-knowledge proof
func (qs queryServer) VerifyZKProof(goCtx context.Context, req *privacypb.QueryVerifyZKProofRequest) (*privacypb.QueryVerifyZKProofResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ZkProof == nil {
		return nil, status.Error(codes.InvalidArgument, "zk proof cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate ZK proofs are enabled
	params := qs.Keeper.GetParams(ctx)
	if !params.EnableZkProofs {
		return &privacypb.QueryVerifyZKProofResponse{
			Valid:        false,
			ErrorMessage: "zk proofs not enabled",
		}, nil
	}

	// Verify proof (simplified)
	valid := qs.Keeper.VerifyZKProofSimple(ctx, req.ZkProof)
	errorMsg := ""
	if !valid {
		errorMsg = "proof verification failed"
	}

	return &privacypb.QueryVerifyZKProofResponse{
		Valid:        valid,
		ErrorMessage: errorMsg,
	}, nil
}

// SECURITY NOTE: DecryptWithViewKey has been removed.
// Decryption must be performed client-side using private keys that never leave the client.
// The blockchain should never receive or process private keys in any form.
