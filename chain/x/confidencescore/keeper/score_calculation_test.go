// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

func TestCalculateScoreEarned(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100)

	registry := newMockIRRegistry()
	registry.irScores["IR-001"] = 500
	registry.irArenas["IR-001"] = "Biometric"
	k.SetIRRegistry(registry)

	walletAddr := "aura1test"

	score, velocity, arena, jackpot, err := k.CalculateScoreEarned(ctx, walletAddr, "IR-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if score == 0 {
		t.Error("expected non-zero score")
	}

	// All multipliers are in basis points (10000 = 1.0x)
	if velocity < BasisPointsBase {
		t.Errorf("expected velocity >= %d (1.0x), got %d", BasisPointsBase, velocity)
	}

	if arena < BasisPointsBase {
		t.Errorf("expected arena >= %d (1.0x), got %d", BasisPointsBase, arena)
	}

	if jackpot < BasisPointsBase {
		t.Errorf("expected jackpot >= %d (1.0x), got %d", BasisPointsBase, jackpot)
	}
}

func TestCalculateScoreEarned_NoRegistry(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100)

	// No registry set - should use default score
	walletAddr := "aura1test"

	score, _, _, _, err := k.CalculateScoreEarned(ctx, walletAddr, "IR-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if score == 0 {
		t.Error("expected non-zero default score")
	}
}

func TestCalculateArenaMultiplier(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100)

	walletAddr := "aura1test"
	arena := "Biometric"

	// No record - should return 10000 (1.0x in basis points)
	multiplier := k.CalculateArenaMultiplier(ctx, walletAddr, arena)
	if multiplier != BasisPointsBase {
		t.Errorf("expected %d (1.0x) for no record, got %d", BasisPointsBase, multiplier)
	}

	// Empty arena - should return 10000 (1.0x)
	multiplier = k.CalculateArenaMultiplier(ctx, walletAddr, "")
	if multiplier != BasisPointsBase {
		t.Errorf("expected %d (1.0x) for empty arena, got %d", BasisPointsBase, multiplier)
	}

	// Create record with arena score below threshold
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		ArenaScores: map[string]*types.ArenaScore{
			arena: {
				ArenaType:  arena,
				TotalScore: 2000, // Below 3000 threshold
				IrCount:    3,
			},
		},
	}
	k.SetUserRecord(ctx, record)

	multiplier = k.CalculateArenaMultiplier(ctx, walletAddr, arena)
	if multiplier != BasisPointsBase {
		t.Errorf("expected %d (1.0x) for score below threshold, got %d", BasisPointsBase, multiplier)
	}

	// Arena score at 3000 threshold
	record.ArenaScores[arena].TotalScore = 3000
	k.SetUserRecord(ctx, record)

	multiplier = k.CalculateArenaMultiplier(ctx, walletAddr, arena)
	if multiplier != 11000 { // 1.1x = 11000 basis points
		t.Errorf("expected 11000 (1.1x) for 3000 score, got %d", multiplier)
	}

	// Arena score at 4000 threshold
	record.ArenaScores[arena].TotalScore = 4000
	k.SetUserRecord(ctx, record)

	multiplier = k.CalculateArenaMultiplier(ctx, walletAddr, arena)
	if multiplier != 12000 { // 1.2x = 12000 basis points
		t.Errorf("expected 12000 (1.2x) for 4000 score, got %d", multiplier)
	}

	// Arena score at 5000 threshold (focus unlocked)
	record.ArenaScores[arena].TotalScore = 5000
	k.SetUserRecord(ctx, record)

	multiplier = k.CalculateArenaMultiplier(ctx, walletAddr, arena)
	if multiplier != 15000 { // 1.5x = 15000 basis points
		t.Errorf("expected 15000 (1.5x) for 5000 score, got %d", multiplier)
	}
}

func TestCalculateTotalScore(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100)

	walletAddr := "aura1test"

	// No record
	_, err := k.CalculateTotalScore(ctx, walletAddr)
	if err != types.ErrUserRecordNotFound {
		t.Errorf("expected ErrUserRecordNotFound, got %v", err)
	}

	// Create record with completions
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		CompletedIrs: []*types.IRCompletion{
			{IrId: "IR-001", FinalScore: 500},
			{IrId: "IR-002", FinalScore: 750},
			{IrId: "IR-003", FinalScore: 1000},
		},
	}
	k.SetUserRecord(ctx, record)

	total, err := k.CalculateTotalScore(ctx, walletAddr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := uint64(2250)
	if total != expected {
		t.Errorf("expected total %d, got %d", expected, total)
	}
}

func TestCalculateArenaScore(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100)

	walletAddr := "aura1test"
	arena := "Biometric"

	// No record
	_, err := k.CalculateArenaScore(ctx, walletAddr, arena)
	if err != types.ErrUserRecordNotFound {
		t.Errorf("expected ErrUserRecordNotFound, got %v", err)
	}

	// Create record with arena completions
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		CompletedIrs: []*types.IRCompletion{
			{IrId: "IR-001", Arena: "Biometric", FinalScore: 500},
			{IrId: "IR-002", Arena: "Biometric", FinalScore: 750},
			{IrId: "IR-003", Arena: "Social", FinalScore: 1000},
		},
	}
	k.SetUserRecord(ctx, record)

	arenaScore, err := k.CalculateArenaScore(ctx, walletAddr, arena)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := uint64(1250) // Only Biometric completions
	if arenaScore != expected {
		t.Errorf("expected arena score %d, got %d", expected, arenaScore)
	}

	// Test different arena
	socialScore, err := k.CalculateArenaScore(ctx, walletAddr, "Social")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if socialScore != 1000 {
		t.Errorf("expected social score 1000, got %d", socialScore)
	}
}

func TestRecalculateScore(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100)

	walletAddr := "aura1test"

	// No record
	_, _, _, err := k.RecalculateScore(ctx, walletAddr)
	if err != types.ErrUserRecordNotFound {
		t.Errorf("expected ErrUserRecordNotFound, got %v", err)
	}

	// Create record with score mismatch
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    5000, // Incorrect
		CompletedIrs: []*types.IRCompletion{
			{IrId: "IR-001", Arena: "Biometric", FinalScore: 500},
			{IrId: "IR-002", Arena: "Biometric", FinalScore: 750},
		},
		ArenaScores: map[string]*types.ArenaScore{
			"Biometric": {
				ArenaType:  "Biometric",
				TotalScore: 1250,
				IrCount:    2,
			},
		},
	}
	k.SetUserRecord(ctx, record)

	previous, recalculated, discrepancies, err := k.RecalculateScore(ctx, walletAddr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if previous != 5000 {
		t.Errorf("expected previous score 5000, got %d", previous)
	}

	if recalculated != 1250 {
		t.Errorf("expected recalculated score 1250, got %d", recalculated)
	}

	if len(discrepancies) == 0 {
		t.Error("expected discrepancies to be found")
	}

	// Verify record was updated
	updatedRecord, _ := k.GetUserRecord(ctx, walletAddr)
	if updatedRecord.TotalScore != 1250 {
		t.Errorf("expected updated score 1250, got %d", updatedRecord.TotalScore)
	}
}

func TestRecalculateScore_ArenaDiscrepancy(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100)

	walletAddr := "aura1test"

	// Create record with arena score mismatch
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    1250,
		CompletedIrs: []*types.IRCompletion{
			{IrId: "IR-001", Arena: "Biometric", FinalScore: 500},
			{IrId: "IR-002", Arena: "Biometric", FinalScore: 750},
		},
		ArenaScores: map[string]*types.ArenaScore{
			"Biometric": {
				ArenaType:  "Biometric",
				TotalScore: 1000, // Wrong
				IrCount:    1,    // Wrong
			},
		},
	}
	k.SetUserRecord(ctx, record)

	_, _, discrepancies, err := k.RecalculateScore(ctx, walletAddr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(discrepancies) < 2 {
		t.Errorf("expected at least 2 discrepancies, got %d", len(discrepancies))
	}

	// Verify arena scores were corrected
	updatedRecord, _ := k.GetUserRecord(ctx, walletAddr)
	if updatedRecord.ArenaScores["Biometric"].TotalScore != 1250 {
		t.Errorf("expected corrected arena score 1250, got %d",
			updatedRecord.ArenaScores["Biometric"].TotalScore)
	}
	if updatedRecord.ArenaScores["Biometric"].IrCount != 2 {
		t.Errorf("expected corrected IR count 2, got %d",
			updatedRecord.ArenaScores["Biometric"].IrCount)
	}
}

func TestCheckVerificationStatus(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100)

	walletAddr := "aura1test"

	// No record
	status, err := k.CheckVerificationStatus(ctx, walletAddr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != types.VerificationStatusUnverified {
		t.Errorf("expected unverified status, got %v", status)
	}

	// Below threshold
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    9000,
		Status:        types.VerificationStatusUnverified,
	}
	k.SetUserRecord(ctx, record)

	status, err = k.CheckVerificationStatus(ctx, walletAddr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != types.VerificationStatusUnverified {
		t.Errorf("expected unverified status, got %v", status)
	}

	// At threshold
	record.TotalScore = 10000
	k.SetUserRecord(ctx, record)

	status, err = k.CheckVerificationStatus(ctx, walletAddr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != types.VerificationStatusVerified {
		t.Errorf("expected verified status, got %v", status)
	}

	// Suspended (should remain suspended regardless of score)
	record.Status = types.VerificationStatusSuspended
	k.SetUserRecord(ctx, record)

	status, err = k.CheckVerificationStatus(ctx, walletAddr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != types.VerificationStatusSuspended {
		t.Errorf("expected suspended status, got %v", status)
	}

	// Revoked
	record.Status = types.VerificationStatusRevoked
	k.SetUserRecord(ctx, record)

	status, err = k.CheckVerificationStatus(ctx, walletAddr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != types.VerificationStatusRevoked {
		t.Errorf("expected revoked status, got %v", status)
	}
}

func TestApplyArenaFocusBonus(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100)

	walletAddr := "aura1test"

	// No record
	_, err := k.ApplyArenaFocusBonus(ctx, walletAddr)
	if err != types.ErrUserRecordNotFound {
		t.Errorf("expected ErrUserRecordNotFound, got %v", err)
	}

	// Create record with multiple arenas
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		ArenaScores: map[string]*types.ArenaScore{
			"Biometric": {
				ArenaType:        "Biometric",
				TotalScore:       5000, // At focus threshold
				IrCount:          10,
				FocusBonusActive: false,
			},
			"Social": {
				ArenaType:        "Social",
				TotalScore:       3000, // Below focus threshold
				IrCount:          6,
				FocusBonusActive: false,
			},
			"GeoLocation": {
				ArenaType:        "GeoLocation",
				TotalScore:       6000, // Above focus threshold
				IrCount:          12,
				FocusBonusActive: false,
			},
		},
	}
	k.SetUserRecord(ctx, record)

	focusArenas, err := k.ApplyArenaFocusBonus(ctx, walletAddr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should have 2 focus arenas (Biometric and GeoLocation)
	if len(focusArenas) != 2 {
		t.Errorf("expected 2 focus arenas, got %d", len(focusArenas))
	}

	// Verify focus bonus flags were set
	updatedRecord, _ := k.GetUserRecord(ctx, walletAddr)
	if !updatedRecord.ArenaScores["Biometric"].FocusBonusActive {
		t.Error("expected Biometric focus bonus to be active")
	}
	if !updatedRecord.ArenaScores["GeoLocation"].FocusBonusActive {
		t.Error("expected GeoLocation focus bonus to be active")
	}
	if updatedRecord.ArenaScores["Social"].FocusBonusActive {
		t.Error("expected Social focus bonus to be inactive")
	}
}

func TestGetArenaBreakdown(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100)

	walletAddr := "aura1test"

	// No record
	_, _, err := k.GetArenaBreakdown(ctx, walletAddr)
	if err != types.ErrUserRecordNotFound {
		t.Errorf("expected ErrUserRecordNotFound, got %v", err)
	}

	// Create record with multiple arenas
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		ArenaScores: map[string]*types.ArenaScore{
			"Biometric": {
				ArenaType:        "Biometric",
				TotalScore:       5000,
				IrCount:          10,
				FocusBonusActive: true,
			},
			"Social": {
				ArenaType:  "Social",
				TotalScore: 3000,
				IrCount:    6,
			},
		},
	}
	k.SetUserRecord(ctx, record)

	arenaScores, focusArenas, err := k.GetArenaBreakdown(ctx, walletAddr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(arenaScores) != 2 {
		t.Errorf("expected 2 arenas, got %d", len(arenaScores))
	}

	if len(focusArenas) != 1 {
		t.Errorf("expected 1 focus arena, got %d", len(focusArenas))
	}

	if arenaScores["Biometric"].TotalScore != 5000 {
		t.Errorf("expected Biometric score 5000, got %d",
			arenaScores["Biometric"].TotalScore)
	}
}

func TestRecalculateScore_VerificationStatusChange(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100)

	walletAddr := "aura1test"

	// Create record that should be verified but isn't
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    12000, // Stored score (wrong)
		Status:        types.VerificationStatusUnverified,
		CompletedIrs: []*types.IRCompletion{
			{IrId: "IR-001", FinalScore: 10000},
		},
	}
	k.SetUserRecord(ctx, record)

	_, recalculated, _, err := k.RecalculateScore(ctx, walletAddr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if recalculated != 10000 {
		t.Errorf("expected recalculated score 10000, got %d", recalculated)
	}

	// Should still be verified since recalculated >= threshold
	updatedRecord, _ := k.GetUserRecord(ctx, walletAddr)
	if updatedRecord.Status != types.VerificationStatusVerified {
		t.Errorf("expected verified status, got %v", updatedRecord.Status)
	}

	// Now test dropping below threshold
	record2 := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000, // Stored score
		Status:        types.VerificationStatusVerified,
		CompletedIrs: []*types.IRCompletion{
			{IrId: "IR-001", FinalScore: 5000}, // Actually only 5000
		},
		VerificationAchievedHeight: 50,
	}
	k.SetUserRecord(ctx, record2)

	_, recalculated, _, err = k.RecalculateScore(ctx, walletAddr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if recalculated != 5000 {
		t.Errorf("expected recalculated score 5000, got %d", recalculated)
	}

	// Should now be unverified
	updatedRecord2, _ := k.GetUserRecord(ctx, walletAddr)
	if updatedRecord2.Status != types.VerificationStatusUnverified {
		t.Errorf("expected unverified status, got %v", updatedRecord2.Status)
	}
	if updatedRecord2.VerificationAchievedHeight != 0 {
		t.Error("expected verification height to be reset")
	}
}
