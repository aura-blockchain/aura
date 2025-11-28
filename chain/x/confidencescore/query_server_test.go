package confidencescore

import (
	"context"
	"testing"

	"github.com/aequitas/aura/chain/x/confidencescore/keeper"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

func TestQueryServer_UserScore(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	server := keeper.NewQueryServer(k)
	ctx := context.Background()

	walletAddr := "aura1test"

	// Query non-existent user
	req := &confidencescorepb.QueryUserScoreRequest{
		WalletAddress: walletAddr,
	}

	resp, err := server.UserScore(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.TotalScore != 0 {
		t.Errorf("expected score 0, got %d", resp.TotalScore)
	}

	// Create user
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    12000,
		Status:        types.VerificationStatusVerified,
		CompletedIrs: []*types.IRCompletion{
			{IrId: "IR-001", FinalScore: 12000},
		},
		ArenaScores: map[string]*types.ArenaScore{
			"Biometric": {
				ArenaType:  "Biometric",
				TotalScore: 12000,
				IrCount:    1,
			},
		},
		VerificationAchievedHeight: 50,
	}
	k.SetUserRecord(record)

	// Query existing user
	resp, err = server.UserScore(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.TotalScore != 12000 {
		t.Errorf("expected score 12000, got %d", resp.TotalScore)
	}

	if !resp.IsVerified {
		t.Error("expected user to be verified")
	}

	if resp.IrCount != 1 {
		t.Errorf("expected 1 IR, got %d", resp.IrCount)
	}

	if len(resp.ArenaScores) != 1 {
		t.Errorf("expected 1 arena score, got %d", len(resp.ArenaScores))
	}

	if resp.VerificationAchievedHeight != 50 {
		t.Errorf("expected verification height 50, got %d", resp.VerificationAchievedHeight)
	}
}

func TestQueryServer_UserScore_NilRequest(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	server := keeper.NewQueryServer(k)
	ctx := context.Background()

	_, err := server.UserScore(ctx, nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
}

func TestQueryServer_UserCompletions(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	server := keeper.NewQueryServer(k)
	ctx := context.Background()

	walletAddr := "aura1test"

	// Create user with completions
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		CompletedIrs: []*types.IRCompletion{
			{IrId: "IR-001", Arena: "Biometric", FinalScore: 500},
			{IrId: "IR-002", Arena: "Biometric", FinalScore: 750},
			{IrId: "IR-003", Arena: "Social", FinalScore: 1000},
		},
	}
	k.SetUserRecord(record)

	// Query all completions
	req := &confidencescorepb.QueryUserCompletionsRequest{
		WalletAddress: walletAddr,
	}

	resp, err := server.UserCompletions(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Completions) != 3 {
		t.Errorf("expected 3 completions, got %d", len(resp.Completions))
	}

	if resp.Pagination.Total != 3 {
		t.Errorf("expected total 3, got %d", resp.Pagination.Total)
	}

	// Query with arena filter
	req.ArenaFilter = "Biometric"
	resp, err = server.UserCompletions(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Completions) != 2 {
		t.Errorf("expected 2 Biometric completions, got %d", len(resp.Completions))
	}
}

func TestQueryServer_ScoreHistory(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)

	server := keeper.NewQueryServer(k)
	ctx := context.Background()

	walletAddr := "aura1test"

	// Add history
	for i := uint64(1); i <= 3; i++ {
		k.SetCurrentHeight(100 + i)
		change := types.ScoreChange{
			BlockHeight:   100 + i,
			ScoreDelta:    int64(i * 100),
			NewTotal:      i * 500,
			TxHash:        walletAddr,
			PreviousScore: (i - 1) * 500,
		}
		k.AddScoreChange(change)
	}

	req := &confidencescorepb.QueryScoreHistoryRequest{
		WalletAddress: walletAddr,
	}

	resp, err := server.ScoreHistory(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Changes) != 3 {
		t.Errorf("expected 3 changes, got %d", len(resp.Changes))
	}

	if resp.Pagination.Total != 3 {
		t.Errorf("expected total 3, got %d", resp.Pagination.Total)
	}
}

func TestQueryServer_Thresholds(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	server := keeper.NewQueryServer(k)
	ctx := context.Background()

	req := &confidencescorepb.QueryThresholdsRequest{}

	resp, err := server.Thresholds(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	params := k.GetParams()

	if resp.VerifiedHumanThreshold != params.VerificationThreshold {
		t.Errorf("expected threshold %d, got %d",
			params.VerificationThreshold, resp.VerifiedHumanThreshold)
	}

	if len(resp.VcThresholds) == 0 {
		t.Error("expected VC thresholds to be set")
	}

	if len(resp.ArenaFocusThresholds) == 0 {
		t.Error("expected arena thresholds to be set")
	}
}

func TestQueryServer_VerifiedUsers(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	server := keeper.NewQueryServer(k)
	ctx := context.Background()

	// Create verified users
	users := []struct {
		wallet string
		score  uint64
	}{
		{"aura1user1", 15000},
		{"aura1user2", 12000},
		{"aura1user3", 11000},
	}

	for _, u := range users {
		record := types.UserConfidenceRecord{
			WalletAddress: u.wallet,
			TotalScore:    u.score,
			Status:        types.VerificationStatusVerified,
		}
		k.SetUserRecord(record)
	}

	req := &confidencescorepb.QueryVerifiedUsersRequest{}

	resp, err := server.VerifiedUsers(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.WalletAddresses) != 3 {
		t.Errorf("expected 3 verified users, got %d", len(resp.WalletAddresses))
	}

	if len(resp.Scores) != 3 {
		t.Errorf("expected 3 scores, got %d", len(resp.Scores))
	}

	// Query with min score
	req.MinScore = 12000
	resp, err = server.VerifiedUsers(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.WalletAddresses) != 2 {
		t.Errorf("expected 2 users with score >= 12000, got %d", len(resp.WalletAddresses))
	}
}

func TestQueryServer_ArenaBreakdown(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	server := keeper.NewQueryServer(k)
	ctx := context.Background()

	walletAddr := "aura1test"

	// Create user with arena scores
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		ArenaScores: map[string]*types.ArenaScore{
			"Biometric": {
				ArenaType:        "Biometric",
				TotalScore:       5000,
				IrCount:          5,
				FocusBonusActive: true,
			},
			"Social": {
				ArenaType:  "Social",
				TotalScore: 3000,
				IrCount:    3,
			},
		},
	}
	k.SetUserRecord(record)

	req := &confidencescorepb.QueryArenaBreakdownRequest{
		WalletAddress: walletAddr,
	}

	resp, err := server.ArenaBreakdown(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.ArenaScores) != 2 {
		t.Errorf("expected 2 arena scores, got %d", len(resp.ArenaScores))
	}

	if len(resp.FocusArenas) != 1 {
		t.Errorf("expected 1 focus arena, got %d", len(resp.FocusArenas))
	}

	if resp.ArenaScores["Biometric"].TotalScore != 5000 {
		t.Errorf("expected Biometric score 5000, got %d",
			resp.ArenaScores["Biometric"].TotalScore)
	}
}

func TestQueryServer_SlashRecord(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)

	server := keeper.NewQueryServer(k)
	ctx := context.Background()

	walletAddr := "aura1test"

	// Create user and slash
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
	}
	k.SetUserRecord(record)

	k.SlashScore(walletAddr, "IR-001", 1000, types.SlashReasonFraudDetected, "gov1", "ev1")

	req := &confidencescorepb.QuerySlashRecordRequest{
		WalletAddress: walletAddr,
	}

	resp, err := server.SlashRecord(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.SlashRecords) != 1 {
		t.Errorf("expected 1 slash record, got %d", len(resp.SlashRecords))
	}

	if resp.Pagination.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Pagination.Total)
	}
}

func TestQueryServer_Params(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	server := keeper.NewQueryServer(k)
	ctx := context.Background()

	req := &confidencescorepb.QueryParamsRequest{}

	resp, err := server.Params(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Params == nil {
		t.Error("expected params to be set")
	}

	params := k.GetParams()

	if resp.Params.VerificationThreshold != params.VerificationThreshold {
		t.Error("params mismatch")
	}
}

func TestQueryServer_IRCompletion(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	server := keeper.NewQueryServer(k)
	ctx := context.Background()

	walletAddr := "aura1test"
	irID := "IR-001"

	// Query non-existent completion
	req := &confidencescorepb.QueryIRCompletionRequest{
		WalletAddress: walletAddr,
		IrId:          irID,
	}

	resp, err := server.IRCompletion(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Completed {
		t.Error("expected completion not found")
	}

	if resp.Completion != nil {
		t.Error("expected nil completion")
	}

	// Create completion
	completion := types.IRCompletion{
		IrId:       irID,
		FinalScore: 500,
		Arena:      "Biometric",
	}
	k.SetIRCompletion(walletAddr, completion)

	// Query existing completion
	resp, err = server.IRCompletion(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !resp.Completed {
		t.Error("expected completion to be found")
	}

	if resp.Completion == nil {
		t.Fatal("expected completion to be set")
	}

	if resp.Completion.IrId != irID {
		t.Errorf("expected IR ID %s, got %s", irID, resp.Completion.IrId)
	}

	if resp.Completion.FinalScore != 500 {
		t.Errorf("expected score 500, got %d", resp.Completion.FinalScore)
	}
}
