package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
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

	params, _ := q.keeper.GetParams(goCtx)

	return &confidencescorepb.QueryParamsResponse{Params: &params}, nil
}

// UserScore returns a user's confidence score and verification status
func (q *QueryServer) UserScore(goCtx context.Context, req *confidencescorepb.QueryUserScoreRequest) (*confidencescorepb.QueryUserScoreResponse, error) {
	if req == nil || req.WalletAddress == "" {
		return nil, types.ErrInvalidWalletAddress
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	record, ok := q.keeper.GetUserRecord(ctx, req.WalletAddress)
	if !ok {
		// Return zero score for non-existent users instead of error
		return &confidencescorepb.QueryUserScoreResponse{
			TotalScore:  0,
			IsVerified:  false,
			Status:      confidencescorepb.VerificationStatus_VERIFICATION_STATUS_UNVERIFIED,
			ArenaScores: make(map[string]*confidencescorepb.ArenaScore),
			IrCount:     0,
		}, nil
	}

	params, _ := q.keeper.GetParams(goCtx)
	isVerified := record.TotalScore >= params.VerificationThreshold &&
		record.Status == confidencescorepb.VerificationStatus_VERIFICATION_STATUS_VERIFIED

	return &confidencescorepb.QueryUserScoreResponse{
		TotalScore:                  record.TotalScore,
		IsVerified:                  isVerified,
		AnchorInfo:                  record.AnchorInfo,
		ArenaScores:                 record.ArenaScores,
		IrCount:                     uint32(len(record.CompletedIrs)),
		LastUpdated:                 record.LastUpdated,
		Status:                      record.Status,
		VerificationAchievedHeight:  record.VerificationAchievedHeight,
	}, nil
}

// ScoreHistory returns the score change history for a user
func (q *QueryServer) ScoreHistory(goCtx context.Context, req *confidencescorepb.QueryScoreHistoryRequest) (*confidencescorepb.QueryScoreHistoryResponse, error) {
	if req == nil || req.WalletAddress == "" {
		return nil, types.ErrInvalidWalletAddress
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Determine limit from pagination
	limit := 100 // default
	if req.Pagination != nil && req.Pagination.Limit > 0 {
		limit = int(req.Pagination.Limit)
	}

	// Get score history with optional height filtering
	changes := q.keeper.GetScoreHistory(ctx, req.WalletAddress, req.FromHeight, req.ToHeight, limit)

	// Convert []ScoreChange to []*ScoreChange for proto response
	changePtrs := make([]*confidencescorepb.ScoreChange, len(changes))
	for i := range changes {
		changePtrs[i] = &changes[i]
	}

	return &confidencescorepb.QueryScoreHistoryResponse{
		Changes: changePtrs,
	}, nil
}

// Thresholds returns verification thresholds and arena focus thresholds
func (q *QueryServer) Thresholds(goCtx context.Context, req *confidencescorepb.QueryThresholdsRequest) (*confidencescorepb.QueryThresholdsResponse, error) {
	if req == nil {
		req = &confidencescorepb.QueryThresholdsRequest{}
	}

	params, _ := q.keeper.GetParams(goCtx)

	// Build VC thresholds map
	// Based on the params structure, we have:
	// - verification_threshold (10,000)
	// - high_assurance_threshold (15,000)
	vcThresholds := map[string]uint64{
		"VerifiedHuman":  params.VerificationThreshold,
		"HighAssurance":  params.HighAssuranceThreshold,
	}

	// Build arena focus thresholds map
	// All arenas use the same threshold from params
	arenaFocusThresholds := map[string]uint64{
		"Biometric":      params.ArenaFocusThreshold,
		"Possession":     params.ArenaFocusThreshold,
		"Knowledge":      params.ArenaFocusThreshold,
		"Social":         params.ArenaFocusThreshold,
		"GeoLocation":    params.ArenaFocusThreshold,
		"HighAssurance":  params.ArenaFocusThreshold,
		"Persistence":    params.ArenaFocusThreshold,
		"Specialized":    params.ArenaFocusThreshold,
	}

	return &confidencescorepb.QueryThresholdsResponse{
		VerifiedHumanThreshold: params.VerificationThreshold,
		VcThresholds:           vcThresholds,
		ArenaFocusThresholds:   arenaFocusThresholds,
	}, nil
}

// VerifiedUsers returns a list of verified users (CS >= 10,000)
func (q *QueryServer) VerifiedUsers(goCtx context.Context, req *confidencescorepb.QueryVerifiedUsersRequest) (*confidencescorepb.QueryVerifiedUsersResponse, error) {
	if req == nil {
		req = &confidencescorepb.QueryVerifiedUsersRequest{}
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params, _ := q.keeper.GetParams(goCtx)

	// Use minimum score from request or default to verification threshold
	minScore := req.MinScore
	if minScore == 0 {
		minScore = params.VerificationThreshold
	}

	// Determine limit from pagination
	limit := 100 // default
	if req.Pagination != nil && req.Pagination.Limit > 0 {
		limit = int(req.Pagination.Limit)
	}

	// Get verified users list
	wallets, scores := q.keeper.ListVerifiedUsers(ctx, minScore, limit)

	return &confidencescorepb.QueryVerifiedUsersResponse{
		WalletAddresses: wallets,
		Scores:          scores,
	}, nil
}
