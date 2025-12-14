package keeper

import (
	"context"

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

// Params returns the module parameters
func (q *QueryServer) Params(goCtx context.Context, req *confidencescorepb.QueryParamsRequest) (*confidencescorepb.QueryParamsResponse, error) {
	if req == nil {
		req = &confidencescorepb.QueryParamsRequest{}
	}

	params := q.keeper.GetParams()

	return &confidencescorepb.QueryParamsResponse{Params: &params}, nil
}
