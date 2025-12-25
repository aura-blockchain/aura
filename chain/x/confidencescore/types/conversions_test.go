// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
	"time"

	"github.com/cosmos/gogoproto/types"
)

func TestArenaScoreToProto(t *testing.T) {
	// Test nil
	if ArenaScoreToProto(nil) != nil {
		t.Error("expected nil for nil input")
	}

	// Test valid arena score
	score := &ArenaScore{
		ArenaType:        "Biometric",
		TotalScore:       5000,
		IrCount:          10,
		FocusBonusActive: true,
	}

	proto := ArenaScoreToProto(score)
	if proto == nil {
		t.Fatal("expected non-nil proto")
	}

	if proto.ArenaType != score.ArenaType {
		t.Errorf("expected arena type %s, got %s", score.ArenaType, proto.ArenaType)
	}

	if proto.TotalScore != score.TotalScore {
		t.Errorf("expected total score %d, got %d", score.TotalScore, proto.TotalScore)
	}

	if proto.IrCount != score.IrCount {
		t.Errorf("expected IR count %d, got %d", score.IrCount, proto.IrCount)
	}

	if proto.FocusBonusActive != score.FocusBonusActive {
		t.Errorf("expected focus bonus %v, got %v", score.FocusBonusActive, proto.FocusBonusActive)
	}
}

func TestAnchorInfoToProto(t *testing.T) {
	// Test nil
	if AnchorInfoToProto(nil) != nil {
		t.Error("expected nil for nil input")
	}

	// Test valid anchor info
	now, _ := types.TimestampProto(time.Now())
	proofHash := []byte("proof_hash")

	info := &AnchorInfo{
		Completed:          true,
		CompletedAt:        now,
		VerifierPluginHash: proofHash,
		BlockHeight:        100,
		ProofHash:          proofHash,
	}

	proto := AnchorInfoToProto(info)
	if proto == nil {
		t.Fatal("expected non-nil proto")
	}

	if proto.Completed != info.Completed {
		t.Error("completed mismatch")
	}

	if proto.BlockHeight != info.BlockHeight {
		t.Errorf("expected block height %d, got %d", info.BlockHeight, proto.BlockHeight)
	}
}

func TestIRCompletionToProto(t *testing.T) {
	// Test nil
	if IRCompletionToProto(nil) != nil {
		t.Error("expected nil for nil input")
	}

	// Test valid completion
	now, _ := types.TimestampProto(time.Now())
	completion := &IRCompletion{
		IrId:             "IR-001",
		BaseScore:        100,
		FinalScore:       150,
		CompletedAt:      now,
		CompletedHeight:  100,
		AssistantAddress: "assistant1",
		ProofHash:        []byte("proof"),
		VerifierHash:     []byte("verifier"),
		TxHash:           "tx123",
		VelocityBonusBps: 15000, // 1.5x in basis points
		ArenaBonusBps:    12000, // 1.2x in basis points
		JackpotBonusBps:  10000, // 1.0x in basis points
		Status:           IRCompletionStatusVerified,
		Arena:            "Biometric",
	}

	proto := IRCompletionToProto(completion)
	if proto == nil {
		t.Fatal("expected non-nil proto")
	}

	if proto.IrId != completion.IrId {
		t.Errorf("expected IR ID %s, got %s", completion.IrId, proto.IrId)
	}

	if proto.BaseScore != completion.BaseScore {
		t.Errorf("expected base score %d, got %d", completion.BaseScore, proto.BaseScore)
	}

	if proto.FinalScore != completion.FinalScore {
		t.Errorf("expected final score %d, got %d", completion.FinalScore, proto.FinalScore)
	}

	if proto.VelocityBonusBps != completion.VelocityBonusBps {
		t.Errorf("expected velocity bonus %d, got %d", completion.VelocityBonusBps, proto.VelocityBonusBps)
	}

	if proto.Arena != completion.Arena {
		t.Errorf("expected arena %s, got %s", completion.Arena, proto.Arena)
	}
}

func TestScoreChangeToProto(t *testing.T) {
	// Test nil
	if ScoreChangeToProto(nil) != nil {
		t.Error("expected nil for nil input")
	}

	// Test valid score change
	now, _ := types.TimestampProto(time.Now())
	change := &ScoreChange{
		BlockHeight:   100,
		ScoreDelta:    500,
		NewTotal:      5000,
		Reason:        ChangeReasonIRCompletion,
		RelatedIrId:   "IR-001",
		TxHash:        "tx123",
		Timestamp:     now,
		PreviousScore: 4500,
	}

	proto := ScoreChangeToProto(change)
	if proto == nil {
		t.Fatal("expected non-nil proto")
	}

	if proto.BlockHeight != change.BlockHeight {
		t.Errorf("expected block height %d, got %d", change.BlockHeight, proto.BlockHeight)
	}

	if proto.ScoreDelta != change.ScoreDelta {
		t.Errorf("expected score delta %d, got %d", change.ScoreDelta, proto.ScoreDelta)
	}

	if proto.NewTotal != change.NewTotal {
		t.Errorf("expected new total %d, got %d", change.NewTotal, proto.NewTotal)
	}

	if proto.PreviousScore != change.PreviousScore {
		t.Errorf("expected previous score %d, got %d", change.PreviousScore, proto.PreviousScore)
	}
}

func TestSlashRecordToProto(t *testing.T) {
	// Test nil
	if SlashRecordToProto(nil) != nil {
		t.Error("expected nil for nil input")
	}

	// Test valid slash record
	now, _ := types.TimestampProto(time.Now())
	record := &SlashRecord{
		WalletAddress:  "aura1test",
		SlashAmount:    1000,
		Reason:         SlashReasonFraudDetected,
		SlashHeight:    100,
		SlashTime:      now,
		RelatedIrId:    "IR-001",
		SlashTxHash:    "slash_tx_123",
		AppealDeadline: now,
		Appealed:       true,
		Resolved:       false,
		Authority:      "gov1",
		Evidence:       "evidence_hash",
	}

	proto := SlashRecordToProto(record)
	if proto == nil {
		t.Fatal("expected non-nil proto")
	}

	if proto.WalletAddress != record.WalletAddress {
		t.Errorf("expected wallet %s, got %s", record.WalletAddress, proto.WalletAddress)
	}

	if proto.SlashAmount != record.SlashAmount {
		t.Errorf("expected slash amount %d, got %d", record.SlashAmount, proto.SlashAmount)
	}

	if proto.Appealed != record.Appealed {
		t.Errorf("expected appealed %v, got %v", record.Appealed, proto.Appealed)
	}

	if proto.Resolved != record.Resolved {
		t.Errorf("expected resolved %v, got %v", record.Resolved, proto.Resolved)
	}
}

func TestParamsToProto(t *testing.T) {
	params := DefaultParams()

	proto := ParamsToProto(params)
	if proto == nil {
		t.Fatal("expected non-nil proto")
	}

	if proto.VerificationThreshold != params.VerificationThreshold {
		t.Error("verification threshold mismatch")
	}

	if proto.HighAssuranceThreshold != params.HighAssuranceThreshold {
		t.Error("high assurance threshold mismatch")
	}

	if proto.ArenaFocusThreshold != params.ArenaFocusThreshold {
		t.Error("arena focus threshold mismatch")
	}

	if len(proto.VelocityBonusDays) != len(params.VelocityBonusDays) {
		t.Error("velocity bonus days length mismatch")
	}

	if len(proto.ArenaMultipliersBps) != len(params.ArenaMultipliersBps) {
		t.Error("arena multipliers length mismatch")
	}

	if proto.SlashPercentage != params.SlashPercentage {
		t.Error("slash percentage mismatch")
	}

	if proto.MaxIrsPerDay != params.MaxIrsPerDay {
		t.Error("max IRs per day mismatch")
	}

	if proto.MaxIrsPerHour != params.MaxIrsPerHour {
		t.Error("max IRs per hour mismatch")
	}

	if proto.PoiRewardsEnabled != params.PoiRewardsEnabled {
		t.Error("PoI rewards enabled mismatch")
	}

	if proto.UserRewardSplitPercent != params.UserRewardSplitPercent {
		t.Error("user reward split mismatch")
	}

	if proto.VelocityBonusEnabled != params.VelocityBonusEnabled {
		t.Error("velocity bonus enabled mismatch")
	}
}

func TestEnumConstants(t *testing.T) {
	// Test that constants are properly defined
	if ChangeReasonIRCompletion != ChangeReason_CHANGE_REASON_IR_COMPLETION {
		t.Error("ChangeReasonIRCompletion constant mismatch")
	}

	if ChangeReasonFraudSlash != ChangeReason_CHANGE_REASON_FRAUD_SLASH {
		t.Error("ChangeReasonFraudSlash constant mismatch")
	}

	if ChangeReasonGovernanceAdjustment != ChangeReason_CHANGE_REASON_GOVERNANCE_ADJUSTMENT {
		t.Error("ChangeReasonGovernanceAdjustment constant mismatch")
	}

	if ChangeReasonAppealReversal != ChangeReason_CHANGE_REASON_APPEAL_REVERSAL {
		t.Error("ChangeReasonAppealReversal constant mismatch")
	}
}
