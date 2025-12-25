// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
)

// QueryServer implements the identitychange Query service
type QueryServer struct {
	identitychangepb.UnimplementedQueryServer
	keeper *Keeper
}

// NewQueryServer returns a new QueryServer
func NewQueryServer(k *Keeper) identitychangepb.QueryServer {
	return &QueryServer{keeper: k}
}

var _ identitychangepb.QueryServer = &QueryServer{}

// Params queries the module parameters
func (qs *QueryServer) Params(ctx context.Context, req *identitychangepb.QueryParamsRequest) (*identitychangepb.QueryParamsResponse, error) {
	if req == nil {
		req = &identitychangepb.QueryParamsRequest{}
	}

	params, _ := qs.keeper.GetParams(ctx)

	return &identitychangepb.QueryParamsResponse{Params: &params}, nil
}
