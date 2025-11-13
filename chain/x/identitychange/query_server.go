package identitychange

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
)

type queryServer struct {
	identitychangepb.UnimplementedQueryServer
	keeper *keeper.Keeper
}

func NewQueryServer(k *keeper.Keeper) identitychangepb.QueryServer {
	return &queryServer{keeper: k}
}

func (s *queryServer) IdentityRecord(ctx context.Context, req *identitychangepb.QueryIdentityRecordRequest) (*identitychangepb.QueryIdentityRecordResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}
	record, ok := s.keeper.GetIdentityRecord(req.Did)
	if !ok {
		return nil, fmt.Errorf("identity record %s not found", req.Did)
	}
	return &identitychangepb.QueryIdentityRecordResponse{Record: recordToProto(record)}, nil
}

func (s *queryServer) IdentityChangeRequest(ctx context.Context, req *identitychangepb.QueryIdentityChangeRequestRequest) (*identitychangepb.QueryIdentityChangeRequestResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}
	request, ok := s.keeper.GetRequest(req.RequestId)
	if !ok {
		return nil, fmt.Errorf("identity change request %s not found", req.RequestId)
	}
	return &identitychangepb.QueryIdentityChangeRequestResponse{Request: requestToProto(request)}, nil
}

func (s *queryServer) IdentityChangeHistory(ctx context.Context, req *identitychangepb.QueryIdentityChangeHistoryRequest) (*identitychangepb.QueryIdentityChangeHistoryResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}
	entries := s.keeper.ListHistory(req.Did)
	pageEntries, page := paginateHistory(entries, req.Pagination)
	return &identitychangepb.QueryIdentityChangeHistoryResponse{
		Entries:    historySliceToProto(pageEntries),
		Pagination: page,
	}, nil
}
