package inclusionroutines

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/inclusionroutines/keeper"
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
	query "github.com/cosmos/cosmos-sdk/types/query"
)

type queryServer struct {
	inclusionroutinespb.UnimplementedQueryServer
	keeper *keeper.Keeper
}

// NewQueryServer creates a new query server
func NewQueryServer(k *keeper.Keeper) inclusionroutinespb.QueryServer {
	return &queryServer{keeper: k}
}

// IR queries a single IR definition by ID
func (s *queryServer) IR(ctx context.Context, req *inclusionroutinespb.QueryIRRequest) (*inclusionroutinespb.QueryIRResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	ir, ok := s.keeper.GetIR(req.Id)
	if !ok {
		return nil, fmt.Errorf("IR %s not found", req.Id)
	}

	return &inclusionroutinespb.QueryIRResponse{
		Ir: types.IRDefinitionToProto(ir),
	}, nil
}

// ListIRs queries a list of IR definitions with filters
func (s *queryServer) ListIRs(ctx context.Context, req *inclusionroutinespb.QueryListIRsRequest) (*inclusionroutinespb.QueryListIRsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	// Extract pagination parameters
	offset := 0
	limit := 100 // Default limit
	if req.Pagination != nil {
		offset = int(req.Pagination.Offset)
		if req.Pagination.Limit > 0 {
			limit = int(req.Pagination.Limit)
		}
	}

	// Query with filters
	irs, total := s.keeper.ListIRs(
		req.StatusFilter,
		req.ArenaFilter,
		req.LocaleFilter,
		offset,
		limit,
	)

	return &inclusionroutinespb.QueryListIRsResponse{
		Irs: types.IRDefinitionsSliceToProto(irs),
		Pagination: &query.PageResponse{
			Total: uint64(total),
		},
	}, nil
}

// IRGraph queries the prerequisite dependency graph
func (s *queryServer) IRGraph(ctx context.Context, req *inclusionroutinespb.QueryIRGraphRequest) (*inclusionroutinespb.QueryIRGraphResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	// Get the graph (full or for specific IR)
	nodes := s.keeper.GetIRGraph(req.IrId)

	return &inclusionroutinespb.QueryIRGraphResponse{
		Nodes: types.IRGraphNodesSliceToProto(nodes),
	}, nil
}

// RateLimit queries the rate limit configuration for an IR
func (s *queryServer) RateLimit(ctx context.Context, req *inclusionroutinespb.QueryRateLimitRequest) (*inclusionroutinespb.QueryRateLimitResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	limit, ok := s.keeper.GetRateLimit(req.IrId)
	if !ok {
		// Return default rate limit if not configured
		params := s.keeper.GetParams()
		limit = types.IRRateLimit{
			IRID:             req.IrId,
			PerWalletPerHour: params.DefaultRateLimitHour,
			PerWalletPerDay:  params.DefaultRateLimitHour * 24,
			PerBlockGlobal:   0,
		}
	}

	return &inclusionroutinespb.QueryRateLimitResponse{
		RateLimit: types.IRRateLimitToProto(limit),
	}, nil
}

// Params queries the module parameters
func (s *queryServer) Params(ctx context.Context, req *inclusionroutinespb.QueryParamsRequest) (*inclusionroutinespb.QueryParamsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	params := s.keeper.GetParams()

	return &inclusionroutinespb.QueryParamsResponse{
		Params: types.ParamsToProto(params),
	}, nil
}
