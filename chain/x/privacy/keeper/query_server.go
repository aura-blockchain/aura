package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

	// Convert to proto params
	protoParams := &privacypb.Params{
		EnableZkProofs:                 params.EnableZkProofs,
		EnableStealthAddresses:         params.EnableStealthAddresses,
		EnableRingSignatures:           params.EnableRingSignatures,
		EnableConfidentialTransactions: params.EnableConfidentialTransactions,
		EnableNetworkPrivacy:           params.EnableNetworkPrivacy,
		EnableMixing:                   params.EnableMixing,
		MinRingSize:                    params.MinRingSize,
		MaxRingSize:                    params.MaxRingSize,
		MinMixingParticipants:          params.MinMixingParticipants,
		MixingFee:                      params.MixingFee,
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

// MixingPools queries all mixing pools
func (qs queryServer) MixingPools(goCtx context.Context, req *privacypb.QueryMixingPoolsRequest) (*privacypb.QueryMixingPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pools := qs.Keeper.GetAllMixingPools(ctx)

	// Filter by status if provided
	if req.Status != "" {
		var filtered []*privacypb.MixingPool
		for _, pool := range pools {
			if pool.Status == req.Status {
				filtered = append(filtered, pool)
			}
		}
		pools = filtered
	}

	return &privacypb.QueryMixingPoolsResponse{MixingPools: pools}, nil
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

// ViewKeys queries all view keys for an address
func (qs queryServer) ViewKeys(goCtx context.Context, req *privacypb.QueryViewKeysRequest) (*privacypb.QueryViewKeysResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	viewKeys := qs.Keeper.GetViewKeys(ctx, req.Address)

	return &privacypb.QueryViewKeysResponse{ViewKeys: viewKeys}, nil
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
