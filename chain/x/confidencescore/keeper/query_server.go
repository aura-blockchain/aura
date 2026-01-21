// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
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
func (q *QueryServer) Params(goCtx context.Context, _ *confidencescorepb.QueryParamsRequest) (*confidencescorepb.QueryParamsResponse, error) {
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
		TotalScore:                 record.TotalScore,
		IsVerified:                 isVerified,
		AnchorInfo:                 record.AnchorInfo,
		ArenaScores:                record.ArenaScores,
		IrCount:                    uint32(len(record.CompletedIrs)),
		LastUpdated:                record.LastUpdated,
		Status:                     record.Status,
		VerificationAchievedHeight: record.VerificationAchievedHeight,
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
func (q *QueryServer) Thresholds(goCtx context.Context, _ *confidencescorepb.QueryThresholdsRequest) (*confidencescorepb.QueryThresholdsResponse, error) {
	params, _ := q.keeper.GetParams(goCtx)

	// Build VC thresholds map
	// Based on the params structure, we have:
	// - verification_threshold (10,000)
	// - high_assurance_threshold (15,000)
	vcThresholds := map[string]uint64{
		"VerifiedHuman": params.VerificationThreshold,
		"HighAssurance": params.HighAssuranceThreshold,
	}

	// Build arena focus thresholds map
	// All arenas use the same threshold from params
	arenaFocusThresholds := map[string]uint64{
		"Biometric":     params.ArenaFocusThreshold,
		"Possession":    params.ArenaFocusThreshold,
		"Knowledge":     params.ArenaFocusThreshold,
		"Social":        params.ArenaFocusThreshold,
		"GeoLocation":   params.ArenaFocusThreshold,
		"HighAssurance": params.ArenaFocusThreshold,
		"Persistence":   params.ArenaFocusThreshold,
		"Specialized":   params.ArenaFocusThreshold,
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

// UserCompletions returns a user's IR completion history
func (q *QueryServer) UserCompletions(goCtx context.Context, req *confidencescorepb.QueryUserCompletionsRequest) (*confidencescorepb.QueryUserCompletionsResponse, error) {
	if req == nil || req.WalletAddress == "" {
		return nil, types.ErrInvalidWalletAddress
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	record, ok := q.keeper.GetUserRecord(ctx, req.WalletAddress)
	if !ok {
		return &confidencescorepb.QueryUserCompletionsResponse{
			Completions: []*confidencescorepb.IRCompletion{},
		}, nil
	}

	// Filter by arena if specified
	completions := record.CompletedIrs
	if req.ArenaFilter != "" {
		filtered := make([]*confidencescorepb.IRCompletion, 0)
		for _, c := range completions {
			if c.Arena == req.ArenaFilter {
				filtered = append(filtered, c)
			}
		}
		completions = filtered
	}

	return &confidencescorepb.QueryUserCompletionsResponse{
		Completions: completions,
	}, nil
}

// ArenaBreakdown returns a user's score breakdown by arena
func (q *QueryServer) ArenaBreakdown(goCtx context.Context, req *confidencescorepb.QueryArenaBreakdownRequest) (*confidencescorepb.QueryArenaBreakdownResponse, error) {
	if req == nil || req.WalletAddress == "" {
		return nil, types.ErrInvalidWalletAddress
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	record, ok := q.keeper.GetUserRecord(ctx, req.WalletAddress)
	if !ok {
		return &confidencescorepb.QueryArenaBreakdownResponse{
			ArenaScores: make(map[string]*confidencescorepb.ArenaScore),
			FocusArenas: []string{},
		}, nil
	}

	params, _ := q.keeper.GetParams(goCtx)

	// Build list of arenas with focus bonus (>= arena focus threshold)
	focusArenas := make([]string, 0)
	for arenaType, arenaScore := range record.ArenaScores {
		if arenaScore.TotalScore >= params.ArenaFocusThreshold {
			focusArenas = append(focusArenas, arenaType)
		}
	}

	return &confidencescorepb.QueryArenaBreakdownResponse{
		ArenaScores: record.ArenaScores,
		FocusArenas: focusArenas,
	}, nil
}

// SlashRecord returns slash records for a user
func (q *QueryServer) SlashRecord(goCtx context.Context, req *confidencescorepb.QuerySlashRecordRequest) (*confidencescorepb.QuerySlashRecordResponse, error) {
	if req == nil || req.WalletAddress == "" {
		return nil, types.ErrInvalidWalletAddress
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	slashRecords := q.keeper.GetSlashRecords(ctx, req.WalletAddress)

	// Convert []SlashRecord to []*SlashRecord for proto response
	recordPtrs := make([]*confidencescorepb.SlashRecord, len(slashRecords))
	for i := range slashRecords {
		recordPtrs[i] = &slashRecords[i]
	}

	return &confidencescorepb.QuerySlashRecordResponse{
		SlashRecords: recordPtrs,
	}, nil
}

// IRCompletion returns a specific IR completion record
func (q *QueryServer) IRCompletion(goCtx context.Context, req *confidencescorepb.QueryIRCompletionRequest) (*confidencescorepb.QueryIRCompletionResponse, error) {
	if req == nil || req.WalletAddress == "" || req.IrId == "" {
		return nil, types.ErrInvalidWalletAddress
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	record, ok := q.keeper.GetUserRecord(ctx, req.WalletAddress)
	if !ok {
		return &confidencescorepb.QueryIRCompletionResponse{
			Completed: false,
		}, nil
	}

	// Search for the specific IR completion
	for _, completion := range record.CompletedIrs {
		if completion.IrId == req.IrId {
			return &confidencescorepb.QueryIRCompletionResponse{
				Completion: completion,
				Completed:  true,
			}, nil
		}
	}

	return &confidencescorepb.QueryIRCompletionResponse{
		Completed: false,
	}, nil
}
