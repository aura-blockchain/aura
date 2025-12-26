// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = queryServer{}

// queryServer is the gRPC server implementation for wasm module queries
type queryServer struct {
	types.UnimplementedQueryServer
	Keeper
}

// NewQueryServerImpl returns an implementation of the wasm QueryServer interface
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

// Params returns the module parameters
func (qs queryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, types.ErrUnauthorized.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := qs.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// Code returns contract code by code ID
func (qs queryServer) Code(goCtx context.Context, req *types.QueryCodeRequest) (*types.QueryCodeResponse, error) {
	if req == nil {
		return nil, types.ErrUnauthorized.Wrap("empty request")
	}

	_ = sdk.UnwrapSDKContext(goCtx)

	// Note: In production, this would query wasmd keeper
	// For now, return stub response
	return &types.QueryCodeResponse{
		CodeId: req.CodeId,
		Data:   []byte{},
	}, nil
}

// Codes returns all stored contract codes with pagination
func (qs queryServer) Codes(goCtx context.Context, req *types.QueryCodesRequest) (*types.QueryCodesResponse, error) {
	if req == nil {
		return nil, types.ErrUnauthorized.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.Logger().Info("Codes query called")

	// If wasmd keeper is not available, return empty response
	if qs.Keeper.wasmKeeper == nil {
		ctx.Logger().Warn("wasmd keeper is nil")
		return &types.QueryCodesResponse{
			CodeInfos:  []types.CodeInfo{},
			Pagination: &query.PageResponse{},
		}, nil
	}

	ctx.Logger().Info("wasmd keeper is available")

	// Initialize with pre-allocated capacity for efficiency.
	// maxCodeID of 100 is a reasonable upper limit for initial implementation.
	codeInfos := make([]types.CodeInfo, 0, 100)

	// Iterate through all code IDs by trying sequential IDs
	// wasmd stores codes with sequential IDs starting from 1
	maxCodeID := uint64(100) // Reasonable upper limit for initial implementation

	for codeID := uint64(1); codeID <= maxCodeID; codeID++ {
		wasmCodeInfo := qs.Keeper.wasmKeeper.GetCodeInfo(ctx, codeID)
		ctx.Logger().Info("GetCodeInfo result", "code_id", codeID, "found", wasmCodeInfo != nil)
		if wasmCodeInfo == nil {
			// No more codes found, but continue searching for gaps
			continue
		}

		ctx.Logger().Info("Found code", "code_id", codeID, "creator", wasmCodeInfo.Creator)

		// Convert wasmd types to our proto types
		// Note: wasmd CodeInfo has fewer fields than our proto, so we set defaults
		codeInfo := types.CodeInfo{
			CodeHash: wasmCodeInfo.CodeHash,
			Creator:  wasmCodeInfo.Creator,
			InstantiateConfig: types.AccessConfig{
				Permission: types.AccessType(wasmCodeInfo.InstantiateConfig.Permission),
				Addresses:  wasmCodeInfo.InstantiateConfig.Addresses,
			},
			CreatedAt: 0,      // wasmd doesn't track this in CodeInfo
			Source:    "",     // wasmd doesn't track this in CodeInfo
			Builder:   "",     // wasmd doesn't track this in CodeInfo
		}
		codeInfos = append(codeInfos, codeInfo)
	}

	ctx.Logger().Info("query codes result", "count", len(codeInfos))

	// Apply pagination manually (simplified - in production use proper pagination)
	return &types.QueryCodesResponse{
		CodeInfos:  codeInfos,
		Pagination: &query.PageResponse{},
	}, nil
}

// ContractInfo returns contract info
func (qs queryServer) ContractInfo(goCtx context.Context, req *types.QueryContractInfoRequest) (*types.QueryContractInfoResponse, error) {
	if req == nil {
		return nil, types.ErrUnauthorized.Wrap("empty request")
	}

	if req.Address == "" {
		return nil, types.ErrInvalidContractAddress.Wrap("address cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid address: %s", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if contract is paused
	isPaused := qs.Keeper.IsContractPaused(ctx, req.Address)

	// Note: In production, this would query wasmd keeper
	// For now, return stub response with pause status
	return &types.QueryContractInfoResponse{
		Address:  req.Address,
		IsPaused: isPaused,
	}, nil
}

// ContractHistory returns contract history
func (qs queryServer) ContractHistory(goCtx context.Context, req *types.QueryContractHistoryRequest) (*types.QueryContractHistoryResponse, error) {
	if req == nil {
		return nil, types.ErrUnauthorized.Wrap("empty request")
	}

	if req.Address == "" {
		return nil, types.ErrInvalidContractAddress.Wrap("address cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid address: %s", err)
	}

	// Note: In production, this would query wasmd keeper
	// For now, return stub response
	return &types.QueryContractHistoryResponse{
		Entries:    []types.ContractHistoryEntry{},
		Pagination: &query.PageResponse{},
	}, nil
}

// AllContractState returns all contract state with pagination
func (qs queryServer) AllContractState(goCtx context.Context, req *types.QueryAllContractStateRequest) (*types.QueryAllContractStateResponse, error) {
	if req == nil {
		return nil, types.ErrUnauthorized.Wrap("empty request")
	}

	if req.Address == "" {
		return nil, types.ErrInvalidContractAddress.Wrap("address cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid address: %s", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if contract is paused
	if qs.Keeper.IsContractPaused(ctx, req.Address) {
		return nil, types.ErrContractPaused.Wrapf("contract %s is paused", req.Address)
	}

	// Note: In production, this would query wasmd keeper
	// For now, return stub response
	return &types.QueryAllContractStateResponse{
		Models:     []types.Model{},
		Pagination: &query.PageResponse{},
	}, nil
}

// RawContractState returns raw contract state by key
func (qs queryServer) RawContractState(goCtx context.Context, req *types.QueryRawContractStateRequest) (*types.QueryRawContractStateResponse, error) {
	if req == nil {
		return nil, types.ErrUnauthorized.Wrap("empty request")
	}

	if req.Address == "" {
		return nil, types.ErrInvalidContractAddress.Wrap("address cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid address: %s", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if contract is paused
	if qs.Keeper.IsContractPaused(ctx, req.Address) {
		return nil, types.ErrContractPaused.Wrapf("contract %s is paused", req.Address)
	}

	// Note: In production, this would query wasmd keeper
	// For now, return stub response
	return &types.QueryRawContractStateResponse{
		Data: []byte{},
	}, nil
}

// SmartContractState returns smart contract state via query
func (qs queryServer) SmartContractState(goCtx context.Context, req *types.QuerySmartContractStateRequest) (*types.QuerySmartContractStateResponse, error) {
	if req == nil {
		return nil, types.ErrUnauthorized.Wrap("empty request")
	}

	if req.Address == "" {
		return nil, types.ErrInvalidContractAddress.Wrap("address cannot be empty")
	}

	contractAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid address: %s", err)
	}

	if len(req.QueryData) == 0 {
		return nil, types.ErrUnauthorized.Wrap("query data cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Query smart contract
	data, err := qs.Keeper.QuerySmart(ctx, contractAddr, req.QueryData)
	if err != nil {
		return nil, err
	}

	return &types.QuerySmartContractStateResponse{
		Data: data,
	}, nil
}

// SecurityStats returns security statistics
func (qs queryServer) SecurityStats(goCtx context.Context, req *types.QuerySecurityStatsRequest) (*types.QuerySecurityStatsResponse, error) {
	if req == nil {
		return nil, types.ErrUnauthorized.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	stats := qs.Keeper.GetSecurityStats(ctx)

	return &types.QuerySecurityStatsResponse{
		Stats: stats,
	}, nil
}

// AuthorizedUploaders returns all authorized uploaders with pagination
func (qs queryServer) AuthorizedUploaders(goCtx context.Context, req *types.QueryAuthorizedUploadersRequest) (*types.QueryAuthorizedUploadersResponse, error) {
	if req == nil {
		return nil, types.ErrUnauthorized.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)

	uploaders := make([]string, 0, 64)
	iterator := storetypes.KVStorePrefixIterator(store, types.ContractAuthKey)
	defer iterator.Close()

	// Manually paginate since we can't use query.FilteredPaginate with iterator directly
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		address := string(key[len(types.ContractAuthKey):])
		uploaders = append(uploaders, address)
	}

	pageRes := &query.PageResponse{
		Total: uint64(len(uploaders)),
	}
	err := error(nil)
	if err != nil {
		return nil, err
	}

	return &types.QueryAuthorizedUploadersResponse{
		Uploaders:  uploaders,
		Pagination: pageRes,
	}, nil
}

// PausedContracts returns all paused contracts with pagination
func (qs queryServer) PausedContracts(goCtx context.Context, req *types.QueryPausedContractsRequest) (*types.QueryPausedContractsResponse, error) {
	if req == nil {
		return nil, types.ErrUnauthorized.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)

	contracts := make([]string, 0, 64)
	iterator := storetypes.KVStorePrefixIterator(store, types.ContractPauseKey)
	defer iterator.Close()

	// Manually paginate since we can't use query.FilteredPaginate with iterator directly
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		address := string(key[len(types.ContractPauseKey):])
		contracts = append(contracts, address)
	}

	pageRes := &query.PageResponse{
		Total: uint64(len(contracts)),
	}
	err := error(nil)
	if err != nil {
		return nil, err
	}

	return &types.QueryPausedContractsResponse{
		Contracts:  contracts,
		Pagination: pageRes,
	}, nil
}

// IsAuthorizedUploader checks if an address is authorized to upload
func (qs queryServer) IsAuthorizedUploader(goCtx context.Context, req *types.QueryIsAuthorizedUploaderRequest) (*types.QueryIsAuthorizedUploaderResponse, error) {
	if req == nil {
		return nil, types.ErrUnauthorized.Wrap("empty request")
	}

	if req.Address == "" {
		return nil, types.ErrUnauthorized.Wrap("address cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrapf("invalid address: %s", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	isAuthorized := qs.Keeper.IsAuthorizedUploader(ctx, req.Address)

	return &types.QueryIsAuthorizedUploaderResponse{
		IsAuthorized: isAuthorized,
	}, nil
}

// IsContractPaused checks if a contract is paused
func (qs queryServer) IsContractPaused(goCtx context.Context, req *types.QueryIsContractPausedRequest) (*types.QueryIsContractPausedResponse, error) {
	if req == nil {
		return nil, types.ErrUnauthorized.Wrap("empty request")
	}

	if req.Address == "" {
		return nil, types.ErrInvalidContractAddress.Wrap("address cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid address: %s", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	isPaused := qs.Keeper.IsContractPaused(ctx, req.Address)

	return &types.QueryIsContractPausedResponse{
		IsPaused: isPaused,
	}, nil
}

// ContractAdmin returns the admin of a contract
func (qs queryServer) ContractAdmin(goCtx context.Context, req *types.QueryContractAdminRequest) (*types.QueryContractAdminResponse, error) {
	if req == nil {
		return nil, types.ErrUnauthorized.Wrap("empty request")
	}

	if req.Address == "" {
		return nil, types.ErrInvalidContractAddress.Wrap("address cannot be empty")
	}

	contractAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid address: %s", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	admin, err := qs.Keeper.GetContractAdmin(ctx, contractAddr)
	if err != nil {
		return nil, err
	}

	// Return empty string if no admin is set
	adminStr := ""
	if !admin.Empty() {
		adminStr = admin.String()
	}

	return &types.QueryContractAdminResponse{
		Admin: adminStr,
	}, nil
}
