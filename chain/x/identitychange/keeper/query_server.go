package keeper

import (
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
