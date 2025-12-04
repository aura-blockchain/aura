package keeper

import (
	"fmt"
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

func TestQueryUserScore(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	// Query non-existent user
	totalScore, isVerified, anchorInfo, arenaScores, irCount, _, status, verificationHeight, err := k.QueryUserScore(ctx, walletAddr)

	if err != nil {
		t.Fatalf("expected no error for non-existent user, got %v", err)
	}

	if totalScore != 0 {
		t.Errorf("expected score 0, got %d", totalScore)
	}

	if isVerified {
		t.Error("expected not verified")
	}

	if anchorInfo != nil {
		t.Error("expected nil anchor info")
	}

	if len(arenaScores) != 0 {
		t.Errorf("expected empty arena scores, got %d", len(arenaScores))
	}

	// Create user record
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
		Status:        types.VerificationStatusVerified,
		HasAnchor:     true,
		AnchorInfo: &types.AnchorInfo{
			Completed: true,
		},
		CompletedIrs: []*types.IRCompletion{
			{IrId: "IR-001", FinalScore: 5000},
			{IrId: "IR-002", FinalScore: 5000},
		},
		ArenaScores: map[string]*types.ArenaScore{
			"Biometric": {
				ArenaType:  "Biometric",
				TotalScore: 8000,
				IrCount:    2,
			},
		},
		VerificationAchievedHeight: 50,
	}
	k.SetUserRecord(ctx, record)

	// Query existing user
	totalScore, isVerified, anchorInfo, arenaScores, irCount, _, status, verificationHeight, err = k.QueryUserScore(ctx, walletAddr)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if totalScore != 10000 {
		t.Errorf("expected score 10000, got %d", totalScore)
	}

	if !isVerified {
		t.Error("expected verified")
	}

	if anchorInfo == nil {
		t.Error("expected anchor info")
	}

	if len(arenaScores) != 1 {
		t.Errorf("expected 1 arena score, got %d", len(arenaScores))
	}

	if irCount != 2 {
		t.Errorf("expected 2 IR completions, got %d", irCount)
	}

	if status != types.VerificationStatusVerified {
		t.Errorf("expected verified status, got %v", status)
	}

	if verificationHeight != 50 {
		t.Errorf("expected verification height 50, got %d", verificationHeight)
	}
}

func TestQueryUserCompletions(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	// Query non-existent user
	completions, total := k.QueryUserCompletions(ctx, walletAddr, "", 0, 10)

	if len(completions) != 0 {
		t.Errorf("expected 0 completions, got %d", len(completions))
	}

	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}

	// Create user with completions
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		CompletedIrs: []*types.IRCompletion{
			{IrId: "IR-001", Arena: "Biometric", FinalScore: 500},
			{IrId: "IR-002", Arena: "Biometric", FinalScore: 750},
			{IrId: "IR-003", Arena: "Social", FinalScore: 1000},
			{IrId: "IR-004", Arena: "GeoLocation", FinalScore: 600},
		},
	}
	k.SetUserRecord(ctx, record)

	// Query all completions
	completions, total = k.QueryUserCompletions(ctx, walletAddr, "", 0, 10)

	if len(completions) != 4 {
		t.Errorf("expected 4 completions, got %d", len(completions))
	}

	if total != 4 {
		t.Errorf("expected total 4, got %d", total)
	}

	// Query with arena filter
	completions, total = k.QueryUserCompletions(ctx, walletAddr, "Biometric", 0, 10)

	if len(completions) != 2 {
		t.Errorf("expected 2 Biometric completions, got %d", len(completions))
	}

	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}

	// Query with pagination
	completions, total = k.QueryUserCompletions(ctx, walletAddr, "", 1, 2)

	if len(completions) != 2 {
		t.Errorf("expected 2 completions (offset 1, limit 2), got %d", len(completions))
	}

	if total != 4 {
		t.Errorf("expected total 4, got %d", total)
	}

	// Query with offset beyond total
	completions, total = k.QueryUserCompletions(ctx, walletAddr, "", 10, 5)

	if len(completions) != 0 {
		t.Errorf("expected 0 completions (offset beyond total), got %d", len(completions))
	}

	if total != 4 {
		t.Errorf("expected total 4, got %d", total)
	}
}

func TestQueryThresholds(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	verifiedThreshold, vcThresholds, arenaThresholds := k.QueryThresholds()

	params := k.GetParams()

	if verifiedThreshold != params.VerificationThreshold {
		t.Errorf("expected verified threshold %d, got %d",
			params.VerificationThreshold, verifiedThreshold)
	}

	// Check VC thresholds
	if vcThresholds["VerifiedHuman"] != params.VerificationThreshold {
		t.Error("VerifiedHuman threshold mismatch")
	}

	if vcThresholds["HighAssurance"] != params.HighAssuranceThreshold {
		t.Error("HighAssurance threshold mismatch")
	}

	if vcThresholds["AgeOver18"] != params.VerificationThreshold {
		t.Error("AgeOver18 threshold mismatch")
	}

	// Check arena thresholds
	if arenaThresholds["Biometric"] != params.ArenaFocusThreshold {
		t.Error("Biometric threshold mismatch")
	}

	if arenaThresholds["Social"] != params.ArenaFocusThreshold {
		t.Error("Social threshold mismatch")
	}
}

func TestGetUserScore(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	// No record
	score, ok := k.GetUserScore(ctx, walletAddr)
	if ok {
		t.Error("expected false for non-existent user")
	}
	if score != 0 {
		t.Errorf("expected score 0, got %d", score)
	}

	// Create record
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    7500,
	}
	k.SetUserRecord(ctx, record)

	score, ok = k.GetUserScore(ctx, walletAddr)
	if !ok {
		t.Error("expected true for existing user")
	}
	if score != 7500 {
		t.Errorf("expected score 7500, got %d", score)
	}
}

func TestGetAnchorInfo(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	// No record
	anchorInfo, ok := k.GetAnchorInfo(ctx, walletAddr)
	if ok {
		t.Error("expected false for non-existent user")
	}
	if anchorInfo != nil {
		t.Error("expected nil anchor info")
	}

	// Record without anchor
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		HasAnchor:     false,
		AnchorInfo:    nil,
	}
	k.SetUserRecord(ctx, record)

	anchorInfo, ok = k.GetAnchorInfo(ctx, walletAddr)
	if ok {
		t.Error("expected false for user without anchor")
	}

	// Record with anchor
	record.HasAnchor = true
	record.AnchorInfo = &types.AnchorInfo{
		Completed:   true,
		BlockHeight: 100,
	}
	k.SetUserRecord(ctx, record)

	anchorInfo, ok = k.GetAnchorInfo(ctx, walletAddr)
	if !ok {
		t.Error("expected true for user with anchor")
	}
	if anchorInfo == nil {
		t.Fatal("expected anchor info")
	}

	// Type assertion to check fields
	if info, ok := anchorInfo.(*types.AnchorInfo); ok {
		if !info.Completed {
			t.Error("expected anchor to be completed")
		}
		if info.BlockHeight != 100 {
			t.Errorf("expected block height 100, got %d", info.BlockHeight)
		}
	} else {
		t.Error("anchor info type mismatch")
	}
}

func TestListVerifiedUsers(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	// No users
	wallets, scores := k.ListVerifiedUsers(ctx, 0, 10)
	if len(wallets) != 0 {
		t.Errorf("expected 0 users, got %d", len(wallets))
	}

	// Create multiple users
	users := []struct {
		wallet string
		score  uint64
		status types.VerificationStatus
	}{
		{"aura1user1", 15000, types.VerificationStatusVerified},
		{"aura1user2", 12000, types.VerificationStatusVerified},
		{"aura1user3", 9000, types.VerificationStatusUnverified}, // Below threshold
		{"aura1user4", 11000, types.VerificationStatusVerified},
		{"aura1user5", 10000, types.VerificationStatusSuspended}, // Suspended
	}

	for _, u := range users {
		record := types.UserConfidenceRecord{
			WalletAddress: u.wallet,
			TotalScore:    u.score,
			Status:        u.status,
		}
		k.SetUserRecord(ctx, record)
	}

	// Query all verified users
	wallets, scores = k.ListVerifiedUsers(ctx, 0, 10)

	// Should get 3 verified users (user1, user2, user4)
	expectedCount := 3
	if len(wallets) != expectedCount {
		t.Errorf("expected %d verified users, got %d", expectedCount, len(wallets))
	}

	if len(scores) != len(wallets) {
		t.Error("wallets and scores length mismatch")
	}

	// Query with min score
	wallets, scores = k.ListVerifiedUsers(ctx, 12000, 10)

	// Should get 2 users (user1: 15000, user2: 12000)
	expectedCount = 2
	if len(wallets) != expectedCount {
		t.Errorf("expected %d users with score >= 12000, got %d",
			expectedCount, len(wallets))
	}

	// Query with limit
	wallets, scores = k.ListVerifiedUsers(ctx, 0, 2)

	if len(wallets) > 2 {
		t.Errorf("expected max 2 users with limit, got %d", len(wallets))
	}
}

func TestGetScoreHistory(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr1 := "aura1user1"
	walletAddr2 := "aura1user2"

	// Test 1: No history for new wallet
	history := k.GetScoreHistory(ctx, walletAddr1, 0, 0, 10)
	if len(history) != 0 {
		t.Errorf("expected 0 history entries for new wallet, got %d", len(history))
	}

	// Test 2: Add history entries for wallet 1
	for i := uint64(1); i <= 5; i++ {
		change := types.ScoreChange{
			ScoreDelta:    int64(i * 100),
			NewTotal:      i * 500,
			Reason:        types.ChangeReasonIRCompletion,
			TxHash:        fmt.Sprintf("tx-user1-%d", i),
			PreviousScore: (i - 1) * 500,
		}
		if err := k.AddScoreChange(ctx, walletAddr1, change); err != nil {
			t.Fatalf("failed to add score change for wallet 1: %v", err)
		}
	}

	// Test 3: Add history entries for wallet 2
	for i := uint64(1); i <= 3; i++ {
		change := types.ScoreChange{
			ScoreDelta:    int64(i * 200),
			NewTotal:      i * 1000,
			Reason:        types.ChangeReasonIRCompletion,
			TxHash:        fmt.Sprintf("tx-user2-%d", i),
			PreviousScore: (i - 1) * 1000,
		}
		if err := k.AddScoreChange(ctx, walletAddr2, change); err != nil {
			t.Fatalf("failed to add score change for wallet 2: %v", err)
		}
	}

	// Test 4: Get all history for wallet 1 (should have 5 entries)
	history = k.GetScoreHistory(ctx, walletAddr1, 0, 0, 0)
	if len(history) != 5 {
		t.Errorf("expected 5 history entries for wallet 1, got %d", len(history))
	}
	// Verify wallet address is set correctly
	for _, h := range history {
		if h.WalletAddress != walletAddr1 {
			t.Errorf("expected wallet address %s, got %s", walletAddr1, h.WalletAddress)
		}
	}

	// Test 5: Get all history for wallet 2 (should have 3 entries)
	history = k.GetScoreHistory(ctx, walletAddr2, 0, 0, 0)
	if len(history) != 3 {
		t.Errorf("expected 3 history entries for wallet 2, got %d", len(history))
	}
	// Verify wallet address is set correctly
	for _, h := range history {
		if h.WalletAddress != walletAddr2 {
			t.Errorf("expected wallet address %s, got %s", walletAddr2, h.WalletAddress)
		}
	}

	// Test 6: Verify histories are isolated (wallet 1 shouldn't see wallet 2's history)
	history = k.GetScoreHistory(ctx, walletAddr1, 0, 0, 0)
	for _, h := range history {
		if h.WalletAddress == walletAddr2 {
			t.Errorf("wallet 1 history contains wallet 2 entries - histories are not isolated!")
		}
	}

	// Test 7: Test with limit
	history = k.GetScoreHistory(ctx, walletAddr1, 0, 0, 2)
	if len(history) != 2 {
		t.Errorf("expected 2 history entries with limit, got %d", len(history))
	}

	// Test 8: Verify AddScoreChange validation
	err := k.AddScoreChange(ctx, "", types.ScoreChange{
		ScoreDelta: 100,
		TxHash:     "test",
	})
	if err == nil {
		t.Error("expected error when adding score change with empty wallet address")
	}
}

func TestGetSlashRecords(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	// Test 1: No slash records for new wallet
	records := k.GetSlashRecords(ctx, walletAddr)
	if len(records) != 0 {
		t.Errorf("expected 0 slash records, got %d", len(records))
	}

	// Test 2: Create user and perform slashes
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
	}
	k.SetUserRecord(ctx, record)

	// First slash
	_, _, _, slashTxHash1, err := k.SlashScore(ctx, walletAddr, "IR-001", 1000, types.SlashReasonFraudDetected, "gov1", "ev1")
	if err != nil {
		t.Fatalf("failed to create first slash: %v", err)
	}

	// Second slash (move forward in time to ensure different tx hash)
	ctx = ctx.WithBlockHeight(101)
	_, _, _, slashTxHash2, err := k.SlashScore(ctx, walletAddr, "IR-002", 500, types.SlashReasonCollusion, "gov1", "ev2")
	if err != nil {
		t.Fatalf("failed to create second slash: %v", err)
	}

	// Test 3: Get all slash records
	records = k.GetSlashRecords(ctx, walletAddr)
	if len(records) != 2 {
		t.Errorf("expected 2 slash records, got %d", len(records))
	}

	// Test 4: Verify specific slash records exist and have correct data
	slash1, ok1 := k.GetSlashRecord(ctx, walletAddr, slashTxHash1)
	if !ok1 {
		t.Error("expected first slash record to exist")
	} else {
		if slash1.SlashAmount != 1000 {
			t.Errorf("expected slash amount 1000, got %d", slash1.SlashAmount)
		}
		if slash1.Reason != types.SlashReasonFraudDetected {
			t.Errorf("expected fraud reason, got %v", slash1.Reason)
		}
		if slash1.RelatedIrId != "IR-001" {
			t.Errorf("expected IR-001, got %s", slash1.RelatedIrId)
		}
		if slash1.WalletAddress != walletAddr {
			t.Errorf("expected wallet %s, got %s", walletAddr, slash1.WalletAddress)
		}
	}

	slash2, ok2 := k.GetSlashRecord(ctx, walletAddr, slashTxHash2)
	if !ok2 {
		t.Error("expected second slash record to exist")
	} else {
		if slash2.SlashAmount != 500 {
			t.Errorf("expected slash amount 500, got %d", slash2.SlashAmount)
		}
		if slash2.Reason != types.SlashReasonCollusion {
			t.Errorf("expected collusion reason, got %v", slash2.Reason)
		}
		if slash2.RelatedIrId != "IR-002" {
			t.Errorf("expected IR-002, got %s", slash2.RelatedIrId)
		}
	}

	// Test 5: Verify records match what GetSlashRecords returns
	foundSlash1 := false
	foundSlash2 := false
	for _, rec := range records {
		if rec.SlashTxHash == slashTxHash1 {
			foundSlash1 = true
			if rec.SlashAmount != 1000 {
				t.Error("slash 1 amount mismatch in GetSlashRecords")
			}
		}
		if rec.SlashTxHash == slashTxHash2 {
			foundSlash2 = true
			if rec.SlashAmount != 500 {
				t.Error("slash 2 amount mismatch in GetSlashRecords")
			}
		}
	}
	if !foundSlash1 {
		t.Error("slash 1 not found in GetSlashRecords")
	}
	if !foundSlash2 {
		t.Error("slash 2 not found in GetSlashRecords")
	}

	// Test 6: Test isolation - different wallet should not see these slashes
	otherWallet := "aura1other"
	otherRecords := k.GetSlashRecords(ctx, otherWallet)
	if len(otherRecords) != 0 {
		t.Errorf("expected 0 records for different wallet, got %d", len(otherRecords))
	}
}

func TestUpdateSlashRecord(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	// No record
	err := k.UpdateSlashRecord(ctx, walletAddr, types.SlashRecord{})
	if err != types.ErrSlashNotFound {
		t.Errorf("expected ErrSlashNotFound, got %v", err)
	}

	// Create slash
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
	}
	k.SetUserRecord(ctx, record)

	_, _, _, slashTxHash, _ := k.SlashScore(ctx,
		walletAddr,
		"IR-001",
		1000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	// Get and update
	slashRecord, ok := k.GetSlashRecord(ctx, walletAddr, slashTxHash)
	if !ok {
		t.Fatal("expected slash record to exist")
	}

	slashRecord.Appealed = true
	slashRecord.Evidence = "new_evidence"

	err = k.UpdateSlashRecord(ctx, walletAddr, slashRecord)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify update
	updated, ok := k.GetSlashRecord(ctx, walletAddr, slashTxHash)
	if !ok {
		t.Fatal("expected updated slash record to exist")
	}

	if !updated.Appealed {
		t.Error("expected slash to be appealed")
	}

	if updated.Evidence != "new_evidence" {
		t.Errorf("expected evidence to be updated, got %s", updated.Evidence)
	}

	// Try to update non-existent slash
	wrongSlash := slashRecord
	wrongSlash.SlashTxHash = "wrong"
	err = k.UpdateSlashRecord(ctx, walletAddr, wrongSlash)
	if err != types.ErrSlashNotFound {
		t.Errorf("expected ErrSlashNotFound for wrong hash, got %v", err)
	}
}
