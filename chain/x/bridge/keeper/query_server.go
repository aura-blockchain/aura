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

// UserTransfers queries transfers for specific user with pagination
func (qs queryServer) UserTransfers(goCtx context.Context, req *bridgeproto.QueryUserTransfersRequest) (*bridgeproto.QueryUserTransfersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if strings.TrimSpace(req.Address) == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	chainFilter := normalizeChain(strings.TrimSpace(req.Chain))

	store := ctx.KVStore(qs.Keeper.storeKey)
	transferStore := prefix.NewStore(store, types.TransferPrefix)

	var transfers []bridgeproto.CrossChainTransfer
	pageRes, err := query.Paginate(transferStore, req.Pagination, func(key, value []byte) error {
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

// BridgeStats queries bridge statistics
func (qs queryServer) BridgeStats(goCtx context.Context, req *bridgeproto.QueryBridgeStatsRequest) (*bridgeproto.QueryBridgeStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	transfers := qs.Keeper.getAllTransfers(ctx)
	transfersByStatus := make(map[string]uint64)
	volumeByChain := make(map[string]string)

	for _, transfer := range transfers {
		if transfer == nil {
			continue
		}
		statusStr := transfer.Status.String()
		transfersByStatus[statusStr]++
		chainID := normalizeChain(transfer.SourceChain)
		if chainID == "" {
			chainID = "unknown"
		}
		amount := transfer.Amount
		if existing, ok := sdkmath.NewIntFromString(volumeByChain[chainID]); ok {
			volumeByChain[chainID] = existing.Add(amount).String()
		} else {
			volumeByChain[chainID] = amount.String()
		}
	}

	wrappedTokens := qs.Keeper.getAllWrappedTokens(ctx)
	chainConfigs := qs.Keeper.getAllChainConfigs(ctx)
	validatorSet := make(map[string]struct{})
	for _, cfg := range chainConfigs {
		for _, val := range cfg.Validators {
			addr := strings.ToLower(strings.TrimSpace(val))
			if addr == "" {
				continue
			}
			validatorSet[addr] = struct{}{}
		}
	}

	return &bridgeproto.QueryBridgeStatsResponse{
		TotalTransfers:     uint64(len(transfers)),
		TransfersByStatus:  transfersByStatus,
		VolumeByChain:      volumeByChain,
		TotalWrappedTokens: uint64(len(wrappedTokens)),
		ActiveValidators:   uint64(len(validatorSet)),
		ActiveRelayers:     qs.countRelayers(ctx),
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

func (qs queryServer) collectValidators(ctx sdk.Context) []*bridgeproto.BridgeValidator {
	store := ctx.KVStore(qs.Keeper.storeKey)
	iterator := store.Iterator(types.ValidatorPrefix, storetypes.PrefixEndBytes(types.ValidatorPrefix))
	defer iterator.Close()
	var validators []*bridgeproto.BridgeValidator
	for ; iterator.Valid(); iterator.Next() {
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

func (qs queryServer) validatorsFromChains(ctx sdk.Context) []*bridgeproto.BridgeValidator {
	cfgs := qs.Keeper.getAllChainConfigs(ctx)
	seen := make(map[string]struct{})
	var validators []*bridgeproto.BridgeValidator
	for _, cfg := range cfgs {
		for _, addr := range cfg.Validators {
			normalized := strings.ToLower(strings.TrimSpace(addr))
			if normalized == "" {
				continue
			}
			if _, exists := seen[normalized]; exists {
				continue
			}
			seen[normalized] = struct{}{}
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
	store := ctx.KVStore(qs.Keeper.storeKey)
	iterator := store.Iterator(types.RelayerPrefix, storetypes.PrefixEndBytes(types.RelayerPrefix))
	defer iterator.Close()
	var count uint64
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	return count
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
