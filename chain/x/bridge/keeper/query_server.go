// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"encoding/hex"
	"strings"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgeproto "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Query limits to prevent DoS via unbounded iteration
const (
	// maxUserTransferFallbackScan limits the fallback O(n) scan for user transfers
	maxUserTransferFallbackScan = 1000
	// maxValidatorQueryLimit limits validator collection queries
	maxValidatorQueryLimit = 500
	// maxRelayerQueryLimit limits relayer counting queries
	maxRelayerQueryLimit = 500
)

var _ bridgeproto.QueryServer = queryServer{}

type queryServer struct {
	bridgeproto.UnimplementedQueryServer
	Keeper *Keeper
}

// NewQueryServerImpl creates a new query server implementation
func NewQueryServerImpl(keeper *Keeper) bridgeproto.QueryServer {
	return &queryServer{Keeper: keeper}
}

// Transfer queries a cross-chain transfer by ID
func (qs queryServer) Transfer(goCtx context.Context, req *bridgeproto.QueryTransferRequest) (*bridgeproto.QueryTransferResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.TransferId == "" {
		return nil, status.Error(codes.InvalidArgument, "transfer id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	transfer, found := qs.Keeper.getTransfer(ctx, req.TransferId)
	if !found {
		return nil, status.Error(codes.NotFound, "transfer not found")
	}

	return &bridgeproto.QueryTransferResponse{Transfer: *transfer}, nil
}

// AllTransfers queries all cross-chain transfers with pagination
func (qs queryServer) AllTransfers(goCtx context.Context, req *bridgeproto.QueryAllTransfersRequest) (*bridgeproto.QueryAllTransfersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	statusFilter := strings.TrimSpace(req.Status)
	var (
		statusValue bridgeproto.TransferStatus
		hasStatus   bool
	)
	if statusFilter != "" {
		canonical := strings.ToUpper(statusFilter)
		val, ok := bridgeproto.TransferStatus_value[canonical]
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "invalid status filter: %s", req.Status)
		}
		statusValue = bridgeproto.TransferStatus(val)
		hasStatus = true
	}

	store := ctx.KVStore(qs.Keeper.storeKey)
	transferStore := prefix.NewStore(store, types.TransferPrefix)

	var transfers []bridgeproto.CrossChainTransfer
	pageRes, err := query.Paginate(transferStore, req.Pagination, func(key, value []byte) error {
		var transfer bridgeproto.CrossChainTransfer
		if err := qs.Keeper.cdc.Unmarshal(value, &transfer); err != nil {
			return err
		}
		// Apply status filter if specified
		if hasStatus && transfer.Status != statusValue {
			return nil
		}
		transfers = append(transfers, transfer)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &bridgeproto.QueryAllTransfersResponse{
		Transfers:  transfers,
		Pagination: pageRes,
	}, nil
}

// UserTransfers queries transfers for specific user with pagination.
// Uses secondary index for O(m) where m is user's transfer count, not O(n) total.
func (qs queryServer) UserTransfers(goCtx context.Context, req *bridgeproto.QueryUserTransfersRequest) (*bridgeproto.QueryUserTransfersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if strings.TrimSpace(req.Address) == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	chainFilter := normalizeChain(strings.TrimSpace(req.Chain))

	// Use secondary index to get user's transfer IDs (O(m) instead of O(n))
	transferIDs := qs.Keeper.GetUserTransferIDs(ctx, req.Address)

	// If no indexed transfers, fall back to scanning (for backwards compatibility)
	if len(transferIDs) == 0 {
		return qs.userTransfersFallback(ctx, req, chainFilter)
	}

	// Apply pagination to the indexed results
	var transfers []bridgeproto.CrossChainTransfer
	offset := uint64(0)
	limit := uint64(100) // default limit
	if req.Pagination != nil {
		offset = req.Pagination.Offset
		if req.Pagination.Limit > 0 {
			limit = req.Pagination.Limit
		}
	}

	total := uint64(0)
	for i, transferID := range transferIDs {
		transfer, found := qs.Keeper.getTransfer(ctx, transferID)
		if !found {
			continue
		}

		// Apply chain filter if specified
		if chainFilter != "" {
			source := normalizeChain(transfer.SourceChain)
			target := normalizeChain(transfer.TargetChain)
			if source != chainFilter && target != chainFilter {
				continue
			}
		}

		total++
		// Apply pagination
		if uint64(i) < offset {
			continue
		}
		if uint64(len(transfers)) >= limit {
			continue
		}
		transfers = append(transfers, *transfer)
	}

	pageRes := &query.PageResponse{
		Total: total,
	}

	return &bridgeproto.QueryUserTransfersResponse{
		Transfers:  transfers,
		Pagination: pageRes,
	}, nil
}

// userTransfersFallback is the legacy O(n) implementation for backwards compatibility.
// Used when secondary index is empty (e.g., for transfers created before index existed).
// Limited to maxUserTransferFallbackScan iterations to prevent DoS.
func (qs queryServer) userTransfersFallback(ctx sdk.Context, req *bridgeproto.QueryUserTransfersRequest, chainFilter string) (*bridgeproto.QueryUserTransfersResponse, error) {
	store := ctx.KVStore(qs.Keeper.storeKey)
	transferStore := prefix.NewStore(store, types.TransferPrefix)

	var transfers []bridgeproto.CrossChainTransfer
	scannedCount := 0

	pageRes, err := query.Paginate(transferStore, req.Pagination, func(key, value []byte) error {
		// Hard cap to prevent DoS on fallback path
		scannedCount++
		if scannedCount > maxUserTransferFallbackScan {
			return nil // Stop scanning
		}

		var transfer bridgeproto.CrossChainTransfer
		if err := qs.Keeper.cdc.Unmarshal(value, &transfer); err != nil {
			return err
		}
		// Filter by address (sender or recipient)
		if transfer.Sender != req.Address && transfer.Recipient != req.Address {
			return nil
		}
		// Apply chain filter if specified
		if chainFilter != "" {
			source := normalizeChain(transfer.SourceChain)
			target := normalizeChain(transfer.TargetChain)
			if source != chainFilter && target != chainFilter {
				return nil
			}
		}
		transfers = append(transfers, transfer)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &bridgeproto.QueryUserTransfersResponse{
		Transfers:  transfers,
		Pagination: pageRes,
	}, nil
}

// ChainConfig queries configuration for a chain
func (qs queryServer) ChainConfig(goCtx context.Context, req *bridgeproto.QueryChainConfigRequest) (*bridgeproto.QueryChainConfigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "chain id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	config, found := qs.Keeper.getChainConfig(ctx, req.ChainId)
	if !found {
		return nil, status.Error(codes.NotFound, "chain config not found")
	}

	return &bridgeproto.QueryChainConfigResponse{Config: config}, nil
}

// AllChains queries all connected chains with pagination
func (qs queryServer) AllChains(goCtx context.Context, req *bridgeproto.QueryAllChainsRequest) (*bridgeproto.QueryAllChainsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)
	chainStore := prefix.NewStore(store, types.ChainConfigPrefix)

	var chains []bridgeproto.ChainConfig
	pageRes, err := query.Paginate(chainStore, req.Pagination, func(key, value []byte) error {
		var config bridgeproto.ChainConfig
		if err := qs.Keeper.cdc.Unmarshal(value, &config); err != nil {
			return err
		}
		chains = append(chains, config)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &bridgeproto.QueryAllChainsResponse{
		Chains:     chains,
		Pagination: pageRes,
	}, nil
}

// WrappedToken queries wrapped token info
func (qs queryServer) WrappedToken(goCtx context.Context, req *bridgeproto.QueryWrappedTokenRequest) (*bridgeproto.QueryWrappedTokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.WrappedDenom == "" {
		return nil, status.Error(codes.InvalidArgument, "wrapped denom cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	token, found := qs.Keeper.getWrappedToken(ctx, req.WrappedDenom)
	if !found {
		return nil, status.Error(codes.NotFound, "wrapped token not found")
	}

	return &bridgeproto.QueryWrappedTokenResponse{Token: *token}, nil
}

// AllWrappedTokens queries all wrapped tokens with pagination
func (qs queryServer) AllWrappedTokens(goCtx context.Context, req *bridgeproto.QueryAllWrappedTokensRequest) (*bridgeproto.QueryAllWrappedTokensResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)
	tokenStore := prefix.NewStore(store, types.WrappedTokenPrefix)

	var tokens []bridgeproto.WrappedToken
	pageRes, err := query.Paginate(tokenStore, req.Pagination, func(key, value []byte) error {
		var token bridgeproto.WrappedToken
		if err := qs.Keeper.cdc.Unmarshal(value, &token); err != nil {
			return err
		}
		tokens = append(tokens, token)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &bridgeproto.QueryAllWrappedTokensResponse{
		Tokens:     tokens,
		Pagination: pageRes,
	}, nil
}

// SharedIdentity queries shared identity across chains
func (qs queryServer) SharedIdentity(goCtx context.Context, req *bridgeproto.QuerySharedIdentityRequest) (*bridgeproto.QuerySharedIdentityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	identity, found := qs.Keeper.getSharedIdentity(ctx, req.Address)
	if !found {
		return nil, status.Error(codes.NotFound, "shared identity not found")
	}

	return &bridgeproto.QuerySharedIdentityResponse{Identity: *identity}, nil
}

// CrossChainSwap queries cross-chain swap status
func (qs queryServer) CrossChainSwap(goCtx context.Context, req *bridgeproto.QueryCrossChainSwapRequest) (*bridgeproto.QueryCrossChainSwapResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.SwapId == "" {
		return nil, status.Error(codes.InvalidArgument, "swap id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	swap, found := qs.Keeper.getSwap(ctx, req.SwapId)
	if !found {
		return nil, status.Error(codes.NotFound, "swap not found")
	}

	return &bridgeproto.QueryCrossChainSwapResponse{Swap: *swap}, nil
}

// BridgeStats queries bridge statistics using cached pre-computed values.
// Falls back to full recomputation if cache is empty or stale.
func (qs queryServer) BridgeStats(goCtx context.Context, req *bridgeproto.QueryBridgeStatsRequest) (*bridgeproto.QueryBridgeStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Try to use cached stats (O(1) instead of O(n) scan)
	cachedStats := qs.Keeper.GetCachedBridgeStats(ctx)

	// If no cache exists, recompute (happens on first query after genesis)
	if cachedStats == nil {
		cachedStats = qs.Keeper.RecomputeBridgeStats(ctx)
	}

	// Update validator and relayer counts (these can change without transfers)
	chainConfigs := qs.Keeper.getAllChainConfigs(ctx)
	validatorSet := make(map[string]struct{})
	for _, cfg := range chainConfigs {
		for _, val := range cfg.Validators {
			addr := strings.ToLower(strings.TrimSpace(val))
			if addr != "" {
				validatorSet[addr] = struct{}{}
			}
		}
	}
	activeValidators := uint64(len(validatorSet))
	activeRelayers := qs.countRelayers(ctx)

	return &bridgeproto.QueryBridgeStatsResponse{
		TotalTransfers:     cachedStats.TotalTransfers,
		TransfersByStatus:  cachedStats.TransfersByStatus,
		VolumeByChain:      cachedStats.VolumeByChain,
		TotalWrappedTokens: cachedStats.TotalWrappedTokens,
		ActiveValidators:   activeValidators,
		ActiveRelayers:     activeRelayers,
		AvgCompletionTime:  0,
	}, nil
}

// Validators queries active bridge validators
func (qs queryServer) Validators(goCtx context.Context, req *bridgeproto.QueryValidatorsRequest) (*bridgeproto.QueryValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)
	validatorStore := prefix.NewStore(store, types.ValidatorPrefix)

	var validators []bridgeproto.BridgeValidator
	pageRes, err := query.Paginate(validatorStore, req.Pagination, func(key, value []byte) error {
		var validator bridgeproto.BridgeValidator
		if err := qs.Keeper.cdc.Unmarshal(value, &validator); err != nil {
			// Log corrupted data but continue iteration
			qs.Keeper.Logger(ctx).Error("failed to unmarshal validator in paginated query",
				"key", hex.EncodeToString(key),
				"error", err.Error())
			return nil
		}
		validators = append(validators, validator)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// If no validators found in store, fall back to deriving from chain configs
	if len(validators) == 0 {
		validatorPtrs := qs.validatorsFromChains(ctx)
		for _, v := range validatorPtrs {
			if v != nil {
				validators = append(validators, *v)
			}
		}
	}

	return &bridgeproto.QueryValidatorsResponse{
		Validators: validators,
		Pagination: pageRes,
	}, nil
}

// RelayerStats queries relayer performance stats
func (qs queryServer) RelayerStats(goCtx context.Context, req *bridgeproto.QueryRelayerStatsRequest) (*bridgeproto.QueryRelayerStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.RelayerAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "relayer address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	stats, found := qs.Keeper.getRelayerStats(ctx, req.RelayerAddress)
	if !found {
		return nil, status.Error(codes.NotFound, "relayer stats not found")
	}

	return &bridgeproto.QueryRelayerStatsResponse{Stats: *stats}, nil
}

// collectValidators collects validators from store with bounded iteration.
// Limited to maxValidatorQueryLimit to prevent DoS.
func (qs queryServer) collectValidators(ctx sdk.Context) []*bridgeproto.BridgeValidator {
	store := ctx.KVStore(qs.Keeper.storeKey)
	iterator := store.Iterator(types.ValidatorPrefix, storetypes.PrefixEndBytes(types.ValidatorPrefix))
	defer iterator.Close()
	// Pre-allocate with reasonable capacity
	validators := make([]*bridgeproto.BridgeValidator, 0, 64)
	count := 0
	for ; iterator.Valid(); iterator.Next() {
		// Hard cap to prevent DoS
		if count >= maxValidatorQueryLimit {
			break
		}
		count++

		var validator bridgeproto.BridgeValidator
		if err := qs.Keeper.cdc.Unmarshal(iterator.Value(), &validator); err != nil {
			// Log corrupted data but continue iteration
			qs.Keeper.Logger(ctx).Error("failed to unmarshal validator in query",
				"key", hex.EncodeToString(iterator.Key()),
				"error", err.Error())
			continue
		}
		valCopy := validator
		validators = append(validators, &valCopy)
	}
	return validators
}

// validatorsFromChains derives validators from chain configs with bounded iteration.
// Limited to maxValidatorQueryLimit to prevent DoS.
func (qs queryServer) validatorsFromChains(ctx sdk.Context) []*bridgeproto.BridgeValidator {
	cfgs := qs.Keeper.getAllChainConfigs(ctx)
	// Pre-allocate maps and slices based on expected sizes
	seen := make(map[string]struct{}, 64)
	validators := make([]*bridgeproto.BridgeValidator, 0, 64)
	count := 0
	for _, cfg := range cfgs {
		for _, addr := range cfg.Validators {
			// Hard cap to prevent DoS
			if count >= maxValidatorQueryLimit {
				return validators
			}

			normalized := strings.ToLower(strings.TrimSpace(addr))
			if normalized == "" {
				continue
			}
			if _, exists := seen[normalized]; exists {
				continue
			}
			seen[normalized] = struct{}{}
			count++
			validators = append(validators, &bridgeproto.BridgeValidator{
				Address: addr,
				Active:  true,
				Chains:  []string{cfg.ChainId},
			})
		}
	}
	return validators
}

func (qs queryServer) countRelayers(ctx sdk.Context) uint64 {
	return qs.Keeper.countRelayers(ctx)
}

// Params queries the module parameters
func (qs queryServer) Params(goCtx context.Context, req *bridgeproto.QueryParamsRequest) (*bridgeproto.QueryParamsResponse, error) {
	if req == nil {
		req = &bridgeproto.QueryParamsRequest{}
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := qs.Keeper.GetParams(ctx)

	// Convert old Params struct to BridgeParams proto message
	maxTransferAmt, ok := sdkmath.NewIntFromString(params.MaxTransferAmount)
	if !ok {
		maxTransferAmt = sdkmath.NewInt(1000000000) // Default fallback
	}

	bridgeParams := bridgeproto.BridgeParams{
		MinConfirmations:             params.MinConfirmations,
		BridgeFeeBasisPoints:         params.BridgeFeeBasisPoints,
		MaxTransferAmount:            maxTransferAmt,
		Enabled:                      params.BridgeEnabled,
		ValidatorThresholdPercentage: params.ValidatorThresholdPercentage,
	}

	return &bridgeproto.QueryParamsResponse{Params: bridgeParams}, nil
}
