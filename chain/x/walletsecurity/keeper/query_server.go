// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	wspb "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

var _ wspb.QueryServer = (*queryServer)(nil)

type queryServer struct {
	wspb.UnimplementedQueryServer
	Keeper *Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper *Keeper) wspb.QueryServer {
	return &queryServer{Keeper: keeper}
}

// GetHardwareWallet retrieves hardware wallet configuration
func (qs queryServer) GetHardwareWallet(goCtx context.Context, req *wspb.QueryGetHardwareWalletRequest) (*wspb.QueryGetHardwareWalletResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.WalletId == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	configBytes, err := qs.Keeper.GetHardwareWallet(ctx, req.WalletId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "hardware wallet not found")
	}

	var config wspb.HardwareWalletConfig
	if err := qs.Keeper.cdc.Unmarshal(configBytes, &config); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.QueryGetHardwareWalletResponse{Config: &config}, nil
}

// GetMultiSigWallet retrieves multi-sig wallet configuration
func (qs queryServer) GetMultiSigWallet(goCtx context.Context, req *wspb.QueryGetMultiSigWalletRequest) (*wspb.QueryGetMultiSigWalletResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.WalletId == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	walletBytes, err := qs.Keeper.GetMultiSigWallet(ctx, req.WalletId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "multi-sig wallet not found")
	}

	var wallet wspb.MultiSigWallet
	if err := qs.Keeper.cdc.Unmarshal(walletBytes, &wallet); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.QueryGetMultiSigWalletResponse{Wallet: &wallet}, nil
}

// GetPendingMultiSigTx retrieves a pending multi-sig transaction
func (qs queryServer) GetPendingMultiSigTx(goCtx context.Context, req *wspb.QueryGetPendingMultiSigTxRequest) (*wspb.QueryGetPendingMultiSigTxResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.TxId == "" {
		return nil, status.Error(codes.InvalidArgument, "tx id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	txBytes, err := qs.Keeper.GetPendingMultiSigTx(ctx, req.TxId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "pending transaction not found")
	}

	var tx wspb.PendingMultiSigTransaction
	if err := qs.Keeper.cdc.Unmarshal(txBytes, &tx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.QueryGetPendingMultiSigTxResponse{Tx: &tx}, nil
}

// GetSocialRecoveryConfig retrieves social recovery configuration
func (qs queryServer) GetSocialRecoveryConfig(goCtx context.Context, req *wspb.QueryGetSocialRecoveryConfigRequest) (*wspb.QueryGetSocialRecoveryConfigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.WalletId == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	configBytes, err := qs.Keeper.GetSocialRecoveryConfig(ctx, req.WalletId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "social recovery not configured")
	}

	var config wspb.SocialRecoveryConfig
	if err := qs.Keeper.cdc.Unmarshal(configBytes, &config); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.QueryGetSocialRecoveryConfigResponse{Config: &config}, nil
}

// GetRecoveryRequest retrieves a recovery request
func (qs queryServer) GetRecoveryRequest(goCtx context.Context, req *wspb.QueryGetRecoveryRequestRequest) (*wspb.QueryGetRecoveryRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.RequestId == "" {
		return nil, status.Error(codes.InvalidArgument, "request id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requestBytes, err := qs.Keeper.GetRecoveryRequest(ctx, req.RequestId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "recovery request not found")
	}

	var request wspb.RecoveryRequest
	if err := qs.Keeper.cdc.Unmarshal(requestBytes, &request); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.QueryGetRecoveryRequestResponse{Request: &request}, nil
}

// GetSpendingLimit retrieves spending limit configuration
func (qs queryServer) GetSpendingLimit(goCtx context.Context, req *wspb.QueryGetSpendingLimitRequest) (*wspb.QueryGetSpendingLimitResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.WalletId == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	limitBytes, err := qs.Keeper.GetSpendingLimit(ctx, req.WalletId, req.Denom)
	if err != nil {
		return nil, status.Error(codes.NotFound, "spending limit not found")
	}

	var limit wspb.SpendingLimit
	if err := qs.Keeper.cdc.Unmarshal(limitBytes, &limit); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.QueryGetSpendingLimitResponse{Limit: &limit}, nil
}

// GetSessionConfig retrieves session configuration
func (qs queryServer) GetSessionConfig(goCtx context.Context, req *wspb.QueryGetSessionConfigRequest) (*wspb.QueryGetSessionConfigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	configBytes, err := qs.Keeper.GetSessionConfig(ctx, req.SessionId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "session config not found")
	}

	var config wspb.SessionConfig
	if err := qs.Keeper.cdc.Unmarshal(configBytes, &config); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.QueryGetSessionConfigResponse{Config: &config}, nil
}

// GetSecurityMetrics retrieves wallet security metrics
func (qs queryServer) GetSecurityMetrics(goCtx context.Context, req *wspb.QueryGetSecurityMetricsRequest) (*wspb.QueryGetSecurityMetricsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.WalletId == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	metricsBytes, err := qs.Keeper.GetSecurityMetrics(ctx, req.WalletId)
	if err != nil {
		// Return empty metrics if not found
		return &wspb.QueryGetSecurityMetricsResponse{
			Metrics: &wspb.WalletSecurityMetrics{
				WalletId:           req.WalletId,
				SecurityScore:      0,
			},
		}, nil
	}

	var metrics wspb.WalletSecurityMetrics
	if err := qs.Keeper.cdc.Unmarshal(metricsBytes, &metrics); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.QueryGetSecurityMetricsResponse{Metrics: &metrics}, nil
}

// GetDomainVerification retrieves domain verification status
func (qs queryServer) GetDomainVerification(goCtx context.Context, req *wspb.QueryGetDomainVerificationRequest) (*wspb.QueryGetDomainVerificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Domain == "" {
		return nil, status.Error(codes.InvalidArgument, "domain cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	verificationBytes, err := qs.Keeper.GetDomainVerification(ctx, req.Domain)
	if err != nil {
		return nil, status.Error(codes.NotFound, "domain verification not found")
	}

	var verification wspb.DomainVerification
	if err := qs.Keeper.cdc.Unmarshal(verificationBytes, &verification); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.QueryGetDomainVerificationResponse{Verification: &verification}, nil
}

// GetDustFilter retrieves dust filter configuration
func (qs queryServer) GetDustFilter(goCtx context.Context, req *wspb.QueryGetDustFilterRequest) (*wspb.QueryGetDustFilterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.WalletId == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	filterBytes, err := qs.Keeper.GetDustFilter(ctx, req.WalletId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "dust filter not configured")
	}

	var filter wspb.DustAttackFilter
	if err := qs.Keeper.cdc.Unmarshal(filterBytes, &filter); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.QueryGetDustFilterResponse{Filter: &filter}, nil
}

// Params queries the module parameters
func (qs queryServer) Params(goCtx context.Context, req *wspb.QueryParamsRequest) (*wspb.QueryParamsResponse, error) {
	if req == nil {
		req = &wspb.QueryParamsRequest{}
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := qs.Keeper.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.QueryParamsResponse{Params: &params}, nil
}
