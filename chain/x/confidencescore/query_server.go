package confidencescore

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/confidencescore/keeper"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
	query "github.com/cosmos/cosmos-sdk/types/query"
)

type queryServer struct {
	confidencescorepb.UnimplementedQueryServer
	keeper *keeper.Keeper
}

// NewQueryServer creates a new query server
func NewQueryServer(k *keeper.Keeper) confidencescorepb.QueryServer {
	return &queryServer{keeper: k}
}

// UserScore queries a user's confidence score and status
func (s *queryServer) UserScore(ctx context.Context, req *confidencescorepb.QueryUserScoreRequest) (*confidencescorepb.QueryUserScoreResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	totalScore, isVerified, anchorInfo, arenaScores, irCount, _, status, verificationHeight, err := s.keeper.QueryUserScore(req.WalletAddress)
	if err != nil {
		return nil, err
	}

	// Convert arena scores to proto
	arenaScoresProto := make(map[string]*confidencescorepb.ArenaScore)
	for arena, score := range arenaScores {
		arenaScoresProto[arena] = types.ArenaScoreToProto(score)
	}

	var anchorInfoProto *confidencescorepb.AnchorInfo
	if anchorInfo != nil {
		anchorInfoProto = types.AnchorInfoToProto(*anchorInfo)
	}

	return &confidencescorepb.QueryUserScoreResponse{
		TotalScore:                 totalScore,
		IsVerified:                 isVerified,
		AnchorInfo:                 anchorInfoProto,
		ArenaScores:                arenaScoresProto,
		IrCount:                    irCount,
		LastUpdated:                nil, // TODO: Convert timestamp if needed
		Status:                     confidencescorepb.VerificationStatus(status),
		VerificationAchievedHeight: verificationHeight,
	}, nil
}

// UserCompletions queries a user's IR completions
func (s *queryServer) UserCompletions(ctx context.Context, req *confidencescorepb.QueryUserCompletionsRequest) (*confidencescorepb.QueryUserCompletionsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	offset := 0
	limit := 100
	if req.Pagination != nil {
		offset = int(req.Pagination.Offset)
		if req.Pagination.Limit > 0 {
			limit = int(req.Pagination.Limit)
		}
	}

	completions, total := s.keeper.QueryUserCompletions(req.WalletAddress, req.ArenaFilter, offset, limit)

	// Convert to proto
	completionsProto := make([]*confidencescorepb.IRCompletion, len(completions))
	for i, completion := range completions {
		completionsProto[i] = types.IRCompletionToProto(completion)
	}

	return &confidencescorepb.QueryUserCompletionsResponse{
		Completions: completionsProto,
		Pagination: &query.PageResponse{
			Total: uint64(total),
		},
	}, nil
}

// ScoreHistory queries a user's score change history
func (s *queryServer) ScoreHistory(ctx context.Context, req *confidencescorepb.QueryScoreHistoryRequest) (*confidencescorepb.QueryScoreHistoryResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	limit := 100
	if req.Pagination != nil && req.Pagination.Limit > 0 {
		limit = int(req.Pagination.Limit)
	}

	changes := s.keeper.GetScoreHistory(req.WalletAddress, req.FromHeight, req.ToHeight, limit)

	// Convert to proto
	changesProto := make([]*confidencescorepb.ScoreChange, len(changes))
	for i, change := range changes {
		changesProto[i] = types.ScoreChangeToProto(change)
	}

	return &confidencescorepb.QueryScoreHistoryResponse{
		Changes: changesProto,
		Pagination: &query.PageResponse{
			Total: uint64(len(changes)),
		},
	}, nil
}

// Thresholds queries verification thresholds
func (s *queryServer) Thresholds(ctx context.Context, req *confidencescorepb.QueryThresholdsRequest) (*confidencescorepb.QueryThresholdsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	verifiedThreshold, vcThresholds, arenaThresholds := s.keeper.QueryThresholds()

	return &confidencescorepb.QueryThresholdsResponse{
		VerifiedHumanThreshold: verifiedThreshold,
		VcThresholds:           vcThresholds,
		ArenaFocusThresholds:   arenaThresholds,
	}, nil
}

// VerifiedUsers queries verified users
func (s *queryServer) VerifiedUsers(ctx context.Context, req *confidencescorepb.QueryVerifiedUsersRequest) (*confidencescorepb.QueryVerifiedUsersResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	limit := 100
	if req.Pagination != nil && req.Pagination.Limit > 0 {
		limit = int(req.Pagination.Limit)
	}

	wallets, scores := s.keeper.ListVerifiedUsers(req.MinScore, limit)

	return &confidencescorepb.QueryVerifiedUsersResponse{
		WalletAddresses: wallets,
		Scores:          scores,
		Pagination: &query.PageResponse{
			Total: uint64(len(wallets)),
		},
	}, nil
}

// ArenaBreakdown queries a user's arena score breakdown
func (s *queryServer) ArenaBreakdown(ctx context.Context, req *confidencescorepb.QueryArenaBreakdownRequest) (*confidencescorepb.QueryArenaBreakdownResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	arenaScores, focusArenas, err := s.keeper.GetArenaBreakdown(req.WalletAddress)
	if err != nil {
		return nil, err
	}

	// Convert to proto
	arenaScoresProto := make(map[string]*confidencescorepb.ArenaScore)
	for arena, score := range arenaScores {
		arenaScoresProto[arena] = types.ArenaScoreToProto(score)
	}

	return &confidencescorepb.QueryArenaBreakdownResponse{
		ArenaScores: arenaScoresProto,
		FocusArenas: focusArenas,
	}, nil
}

// SlashRecord queries slash records for a user
func (s *queryServer) SlashRecord(ctx context.Context, req *confidencescorepb.QuerySlashRecordRequest) (*confidencescorepb.QuerySlashRecordResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	slashRecords := s.keeper.GetSlashRecords(req.WalletAddress)

	// Convert to proto
	slashRecordsProto := make([]*confidencescorepb.SlashRecord, len(slashRecords))
	for i, record := range slashRecords {
		slashRecordsProto[i] = types.SlashRecordToProto(record)
	}

	return &confidencescorepb.QuerySlashRecordResponse{
		SlashRecords: slashRecordsProto,
		Pagination: &query.PageResponse{
			Total: uint64(len(slashRecords)),
		},
	}, nil
}

// Params queries module parameters
func (s *queryServer) Params(ctx context.Context, req *confidencescorepb.QueryParamsRequest) (*confidencescorepb.QueryParamsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	params := s.keeper.GetParams()

	return &confidencescorepb.QueryParamsResponse{
		Params: types.ParamsToProto(params),
	}, nil
}

// IRCompletion queries a specific IR completion
func (s *queryServer) IRCompletion(ctx context.Context, req *confidencescorepb.QueryIRCompletionRequest) (*confidencescorepb.QueryIRCompletionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	completion, ok := s.keeper.GetIRCompletion(req.WalletAddress, req.IrId)
	if !ok {
		return &confidencescorepb.QueryIRCompletionResponse{
			Completion: nil,
			Completed:  false,
		}, nil
	}

	return &confidencescorepb.QueryIRCompletionResponse{
		Completion: types.IRCompletionToProto(completion),
		Completed:  true,
	}, nil
}
