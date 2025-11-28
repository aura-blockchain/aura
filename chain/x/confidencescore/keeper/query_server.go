package keeper

import (
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

// QueryServer implements the confidencescore Query service
type QueryServer struct {
	confidencescorepb.UnimplementedQueryServer
	keeper *Keeper
}

// NewQueryServer returns a new QueryServer
func NewQueryServer(k *Keeper) confidencescorepb.QueryServer {
	return &QueryServer{keeper: k}
}

var _ confidencescorepb.QueryServer = &QueryServer{}
