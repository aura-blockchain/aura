package identitychange

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/types"
)

type queryServer struct {
	keeper *keeper.Keeper
}

func NewQueryServer(k *keeper.Keeper) QueryServer {
	return &queryServer{keeper: k}
}

func (s *queryServer) IdentityRecord(ctx context.Context, req *QueryIdentityRecordRequest) (*QueryIdentityRecordResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}
	record, ok := s.keeper.GetIdentityRecord(req.DID)
	if !ok {
		return nil, fmt.Errorf("identity record %s not found", req.DID)
	}
	return &QueryIdentityRecordResponse{Record: &record}, nil
}

func (s *queryServer) IdentityChangeRequest(ctx context.Context, req *QueryIdentityChangeRequestRequest) (*QueryIdentityChangeRequestResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}
	request, ok := s.keeper.GetRequest(req.RequestID)
	if !ok {
		return nil, fmt.Errorf("identity change request %s not found", req.RequestID)
	}
	return &QueryIdentityChangeRequestResponse{Request: &request}, nil
}

func (s *queryServer) IdentityChangeHistory(ctx context.Context, req *QueryIdentityChangeHistoryRequest) (*QueryIdentityChangeHistoryResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}
	entries := s.keeper.ListHistory(req.DID)
	response := &QueryIdentityChangeHistoryResponse{Entries: make([]*types.IdentityChangeHistory, 0, len(entries))}
	for _, entry := range entries {
		entryCopy := entry
		response.Entries = append(response.Entries, &entryCopy)
	}
	return response, nil
}
