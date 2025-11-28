package keeper

import (
	economicsecuritypb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
)

// QueryServer implements the economicsecurity Query service
type QueryServer struct {
	economicsecuritypb.UnimplementedQueryServer
	keeper *Keeper
}

// NewQueryServer returns a new QueryServer
func NewQueryServer(k *Keeper) economicsecuritypb.QueryServer {
	return &QueryServer{keeper: k}
}

var _ economicsecuritypb.QueryServer = &QueryServer{}
