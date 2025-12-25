// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/aequitas/aura/chain/x/prevalidation/types"
	pb "github.com/aequitas/aura/proto/aura/prevalidation/v1beta1"
)

type queryServer struct {
	pb.UnimplementedQueryServer
	keeper *Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper *Keeper) pb.QueryServer {
	return &queryServer{keeper: keeper}
}

var _ pb.QueryServer = (*queryServer)(nil)

// Params queries module parameters
func (qs queryServer) Params(goCtx context.Context, req *pb.QueryParamsRequest) (*pb.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params, err := qs.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.QueryParamsResponse{
		Params: &params,
	}, nil
}

// PreValidatedTransaction queries a pre-validated transaction by ID
func (qs queryServer) PreValidatedTransaction(goCtx context.Context, req *pb.QueryPreValidatedTransactionRequest) (*pb.QueryPreValidatedTransactionResponse, error) {
	// In production, would retrieve from state
	// For now, return empty response
	return &pb.QueryPreValidatedTransactionResponse{
		Transaction: nil,
	}, nil
}

// PreValidatedTransactionsByStatus queries pre-validated transactions by status
func (qs queryServer) PreValidatedTransactionsByStatus(goCtx context.Context, req *pb.QueryPreValidatedTransactionsByStatusRequest) (*pb.QueryPreValidatedTransactionsByStatusResponse, error) {
	// In production, would filter by status from state
	return &pb.QueryPreValidatedTransactionsByStatusResponse{
		Transactions: []*pb.PreValidatedTransaction{},
		Pagination:   nil,
	}, nil
}

// PreValidatedTransactionsBySigner queries pre-validated transactions by signer
func (qs queryServer) PreValidatedTransactionsBySigner(goCtx context.Context, req *pb.QueryPreValidatedTransactionsBySignerRequest) (*pb.QueryPreValidatedTransactionsBySignerResponse, error) {
	// In production, would filter by signer from state
	return &pb.QueryPreValidatedTransactionsBySignerResponse{
		Transactions: []*pb.PreValidatedTransaction{},
		Pagination:   nil,
	}, nil
}

// Template queries a validation template by ID
func (qs queryServer) Template(goCtx context.Context, req *pb.QueryTemplateRequest) (*pb.QueryTemplateResponse, error) {
	// In production, would retrieve from state
	return &pb.QueryTemplateResponse{
		Template: nil,
	}, nil
}

// AllTemplates queries all validation templates
func (qs queryServer) AllTemplates(goCtx context.Context, req *pb.QueryAllTemplatesRequest) (*pb.QueryAllTemplatesResponse, error) {
	// In production, would retrieve all from state
	return &pb.QueryAllTemplatesResponse{
		Templates:  []*pb.ValidationTemplate{},
		Pagination: nil,
	}, nil
}

// TemplatesByType queries templates by transaction type
func (qs queryServer) TemplatesByType(goCtx context.Context, req *pb.QueryTemplatesByTypeRequest) (*pb.QueryTemplatesByTypeResponse, error) {
	// In production, would filter by type from state
	return &pb.QueryTemplatesByTypeResponse{
		Templates:  []*pb.ValidationTemplate{},
		Pagination: nil,
	}, nil
}

// Metrics queries pre-validation metrics
func (qs queryServer) Metrics(goCtx context.Context, req *pb.QueryMetricsRequest) (*pb.QueryMetricsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get metrics from keeper
	metrics := qs.keeper.GetMetrics(ctx)

	return &pb.QueryMetricsResponse{
		Metrics: metrics,
	}, nil
}

// MetricsByType queries metrics for a specific transaction type
func (qs queryServer) MetricsByType(goCtx context.Context, req *pb.QueryMetricsByTypeRequest) (*pb.QueryMetricsByTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get type-specific metrics from keeper
	metrics := qs.keeper.GetTypeMetrics(ctx, req.TxType)

	return &pb.QueryMetricsByTypeResponse{
		Metrics: metrics,
	}, nil
}

// EstimateGas estimates gas for a transaction
func (qs queryServer) EstimateGas(goCtx context.Context, req *pb.QueryEstimateGasRequest) (*pb.QueryEstimateGasResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tx := types.Transaction{
		Sender:    req.Sender,
		Recipient: req.Recipient,
		Amount:    req.Amount,
		Data:      req.Data,
	}

	gasEstimate := qs.keeper.EstimateGas(ctx, tx)

	return &pb.QueryEstimateGasResponse{
		GasEstimate: gasEstimate,
		GasLimit:    gasEstimate + (gasEstimate / 10), // Add 10% buffer
	}, nil
}

// ValidateTransaction validates a transaction without pre-validating it
func (qs queryServer) ValidateTransaction(goCtx context.Context, req *pb.QueryValidateTransactionRequest) (*pb.QueryValidateTransactionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tx := types.Transaction{
		Sender:    req.Sender,
		Recipient: req.Recipient,
		Amount:    req.Amount,
		Data:      req.Data,
		Nonce:     req.Nonce,
	}

	valid, err := qs.keeper.ValidateTransaction(ctx, tx)

	response := &pb.QueryValidateTransactionResponse{
		Valid:       valid,
		GasEstimate: qs.keeper.EstimateGas(ctx, tx),
	}

	if err != nil {
		response.Error = err.Error()
	}

	// Check balance
	hasBalance := qs.keeper.CheckSufficientBalance(ctx, tx.Sender, tx.Amount)
	response.SufficientBalance = hasBalance

	return response, nil
}

// GetNonce queries the current nonce for an address
func (qs queryServer) GetNonce(goCtx context.Context, req *pb.QueryGetNonceRequest) (*pb.QueryGetNonceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	nonce := qs.keeper.GetNonce(ctx, req.Address)

	return &pb.QueryGetNonceResponse{
		Nonce: nonce,
	}, nil
}
