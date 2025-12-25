// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

// QueryServer implements the inclusionroutines Query service
type QueryServer struct {
	inclusionroutinespb.UnimplementedQueryServer
	keeper *Keeper
}

// NewQueryServer returns a new QueryServer
func NewQueryServer(k *Keeper) inclusionroutinespb.QueryServer {
	return &QueryServer{keeper: k}
}

var _ inclusionroutinespb.QueryServer = &QueryServer{}
