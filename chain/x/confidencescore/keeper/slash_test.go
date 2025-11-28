package keeper

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

func TestSlashScore_Success(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	walletAddr := "aura1test"

	// Create user record
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
		Status:        types.VerificationStatusVerified,
	}
	k.SetUserRecord(record)

	previousScore, newScore, verificationRevoked, slashTxHash, err := k.SlashScore(
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence_hash_123",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if previousScore != 10000 {
		t.Errorf("expected previous score 10000, got %d", previousScore)
	}

	if newScore != 8000 {
		t.Errorf("expected new score 8000, got %d", newScore)
	}

	// 8000 < 10000 threshold, so verification should be revoked
	if !verificationRevoked {
		t.Error("expected verification to be revoked when score drops below threshold")
	}

	if slashTxHash == "" {
		t.Error("expected slash tx hash to be set")
	}

	// Verify slash record was created
	slashRecord, ok := k.GetSlashRecord(walletAddr, slashTxHash)
	if !ok {
		t.Error("expected slash record to exist")
	}

	if slashRecord.SlashAmount != 2000 {
		t.Errorf("expected slash amount 2000, got %d", slashRecord.SlashAmount)
	}

	if slashRecord.Reason != types.SlashReasonFraudDetected {
		t.Errorf("expected fraud reason, got %v", slashRecord.Reason)
	}
}

func TestSlashScore_InvalidInputs(t *testing.T) {
	k := NewKeeper(nil, "gov1")

	tests := []struct {
		name        string
		walletAddr  string
		slashAmount uint64
		expectError error
	}{
		{
			name:        "empty wallet address",
			walletAddr:  "",
			slashAmount: 1000,
			expectError: types.ErrInvalidWalletAddress,
		},
		{
			name:        "zero slash amount",
			walletAddr:  "aura1test",
			slashAmount: 0,
			expectError: types.ErrInvalidSlashAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, _, err := k.SlashScore(
				tt.walletAddr,
				"IR-001",
				tt.slashAmount,
				types.SlashReasonFraudDetected,
				"gov1",
				"evidence",
			)

			if err != tt.expectError {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestSlashScore_UserNotFound(t *testing.T) {
	k := NewKeeper(nil, "gov1")

	_, _, _, _, err := k.SlashScore(
		"nonexistent",
		"IR-001",
		1000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	if err != types.ErrUserRecordNotFound {
		t.Errorf("expected ErrUserRecordNotFound, got %v", err)
	}
}

func TestSlashScore_ExceedsTotalScore(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	walletAddr := "aura1test"

	// Create user with low score
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    1000,
		Status:        types.VerificationStatusUnverified,
	}
	k.SetUserRecord(record)

	// Slash more than total
	_, newScore, _, _, err := k.SlashScore(
		walletAddr,
		"IR-001",
		5000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Score should be 0, not negative
	if newScore != 0 {
		t.Errorf("expected score 0, got %d", newScore)
	}
}

func TestSlashScore_VerificationRevoked(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	walletAddr := "aura1test"

	// Create verified user
	record := types.UserConfidenceRecord{
		WalletAddress:              walletAddr,
		TotalScore:                 10000,
		Status:                     types.VerificationStatusVerified,
		VerificationAchievedHeight: 50,
	}
	k.SetUserRecord(record)

	// Slash enough to drop below threshold
	_, newScore, verificationRevoked, _, err := k.SlashScore(
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if newScore != 8000 {
		t.Errorf("expected new score 8000, got %d", newScore)
	}

	// Verification should NOT be revoked since 8000 >= 10000 threshold is false
	// Actually, let's check - the threshold is 10000, current score is 8000
	// So verification SHOULD be revoked
	if !verificationRevoked {
		t.Error("expected verification to be revoked when score drops below threshold")
	}

	// Verify record was updated
	updatedRecord, _ := k.GetUserRecord(walletAddr)
	if updatedRecord.Status != types.VerificationStatusUnverified {
		t.Error("expected status to be unverified after revocation")
	}
	if updatedRecord.VerificationAchievedHeight != 0 {
		t.Error("expected verification height to be reset")
	}
}

func TestAppealSlash_Success(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	walletAddr := "aura1test"

	// Create slash record first
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	k.SetUserRecord(record)

	_, _, _, slashTxHash, _ := k.SlashScore(
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	// Appeal the slash
	params := k.GetParams()
	appealAccepted, reviewDeadline, err := k.AppealSlash(
		walletAddr,
		slashTxHash,
		"counter_evidence",
		params.AppealDeposit,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !appealAccepted {
		t.Error("expected appeal to be accepted")
	}

	if reviewDeadline == 0 {
		t.Error("expected review deadline to be set")
	}

	// Verify slash record was updated
	slashRecord, _ := k.GetSlashRecord(walletAddr, slashTxHash)
	if !slashRecord.Appealed {
		t.Error("expected slash to be marked as appealed")
	}
}

func TestAppealSlash_NotFound(t *testing.T) {
	k := NewKeeper(nil, "gov1")

	_, _, err := k.AppealSlash(
		"aura1test",
		"nonexistent_slash",
		"evidence",
		"1000uaura",
	)

	if err != types.ErrSlashNotFound {
		t.Errorf("expected ErrSlashNotFound, got %v", err)
	}
}

func TestAppealSlash_AlreadyAppealed(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	walletAddr := "aura1test"

	// Create and appeal slash
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	k.SetUserRecord(record)

	_, _, _, slashTxHash, _ := k.SlashScore(
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	params := k.GetParams()
	k.AppealSlash(walletAddr, slashTxHash, "evidence", params.AppealDeposit)

	// Try to appeal again
	_, _, err := k.AppealSlash(walletAddr, slashTxHash, "more_evidence", params.AppealDeposit)

	if err != types.ErrSlashAlreadyAppealed {
		t.Errorf("expected ErrSlashAlreadyAppealed, got %v", err)
	}
}

func TestAppealSlash_Expired(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	currentTime := time.Now().Unix()
	k.SetCurrentHeight(100)
	k.SetCurrentTime(currentTime)

	walletAddr := "aura1test"

	// Create slash record
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	k.SetUserRecord(record)

	_, _, _, slashTxHash, _ := k.SlashScore(
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	// Move time forward past appeal deadline (14 days + 1)
	k.SetCurrentTime(currentTime + (15 * 24 * 3600))

	params := k.GetParams()
	_, _, err := k.AppealSlash(walletAddr, slashTxHash, "evidence", params.AppealDeposit)

	if err != types.ErrAppealExpired {
		t.Errorf("expected ErrAppealExpired, got %v", err)
	}
}

func TestAppealSlash_InsufficientDeposit(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	walletAddr := "aura1test"

	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	k.SetUserRecord(record)

	_, _, _, slashTxHash, _ := k.SlashScore(
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	// Wrong deposit amount
	_, _, err := k.AppealSlash(walletAddr, slashTxHash, "evidence", "500uaura")

	if err != types.ErrInsufficientDeposit {
		t.Errorf("expected ErrInsufficientDeposit, got %v", err)
	}
}

func TestResolveAppeal_RestoreScore(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	walletAddr := "aura1test"

	// Create user with initial score
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
	}
	k.SetUserRecord(record)

	_, _, _, slashTxHash, _ := k.SlashScore(
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	params := k.GetParams()
	k.AppealSlash(walletAddr, slashTxHash, "counter_evidence", params.AppealDeposit)

	// Resolve appeal - restore score
	restoredScore, depositReturned, err := k.ResolveAppeal(
		walletAddr,
		slashTxHash,
		true,
		"gov1",
		"appeal upheld",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if restoredScore != 2000 {
		t.Errorf("expected restored score 2000, got %d", restoredScore)
	}

	if !depositReturned {
		t.Error("expected deposit to be returned")
	}

	// Verify score was restored
	// Note: The score was 10000, slashed 2000 to become 8000
	// When we restore 2000, it should go back to 10000
	updatedRecord, _ := k.GetUserRecord(walletAddr)
	expectedScore := uint64(10000) // 8000 + 2000 restored
	if updatedRecord.TotalScore != expectedScore {
		t.Errorf("expected total score %d, got %d", expectedScore, updatedRecord.TotalScore)
	}

	// Verify slash is marked as resolved
	slashRecord, _ := k.GetSlashRecord(walletAddr, slashTxHash)
	if !slashRecord.Resolved {
		t.Error("expected slash to be marked as resolved")
	}
}

func TestResolveAppeal_DenyAppeal(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	walletAddr := "aura1test"

	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
	}
	k.SetUserRecord(record)

	_, _, _, slashTxHash, _ := k.SlashScore(
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	params := k.GetParams()
	k.AppealSlash(walletAddr, slashTxHash, "counter_evidence", params.AppealDeposit)

	// Resolve appeal - deny (don't restore score)
	restoredScore, depositReturned, err := k.ResolveAppeal(
		walletAddr,
		slashTxHash,
		false,
		"gov1",
		"appeal denied",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if restoredScore != 0 {
		t.Errorf("expected no score restored, got %d", restoredScore)
	}

	if depositReturned {
		t.Error("expected deposit not to be returned")
	}

	// Verify score was not restored
	updatedRecord, _ := k.GetUserRecord(walletAddr)
	if updatedRecord.TotalScore != 8000 {
		t.Errorf("expected total score to remain 8000, got %d", updatedRecord.TotalScore)
	}
}

func TestResolveAppeal_NotAppealed(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	walletAddr := "aura1test"

	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	k.SetUserRecord(record)

	_, _, _, slashTxHash, _ := k.SlashScore(
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	// Try to resolve without appeal
	_, _, err := k.ResolveAppeal(walletAddr, slashTxHash, true, "gov1", "notes")

	if err == nil {
		t.Error("expected error for non-appealed slash")
	}
}

func TestResolveAppeal_AlreadyResolved(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	walletAddr := "aura1test"

	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	k.SetUserRecord(record)

	_, _, _, slashTxHash, _ := k.SlashScore(
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	params := k.GetParams()
	k.AppealSlash(walletAddr, slashTxHash, "evidence", params.AppealDeposit)

	// Resolve first time
	k.ResolveAppeal(walletAddr, slashTxHash, true, "gov1", "notes")

	// Try to resolve again
	_, _, err := k.ResolveAppeal(walletAddr, slashTxHash, true, "gov1", "notes")

	if err != types.ErrAppealAlreadyResolved {
		t.Errorf("expected ErrAppealAlreadyResolved, got %v", err)
	}
}

func TestCalculateSlashAmount(t *testing.T) {
	k := NewKeeper(nil, "gov1")

	walletAddr := "aura1test"

	// No record
	_, err := k.CalculateSlashAmount(walletAddr, 50)
	if err != types.ErrUserRecordNotFound {
		t.Errorf("expected ErrUserRecordNotFound, got %v", err)
	}

	// Create record
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
	}
	k.SetUserRecord(record)

	// Calculate 50% slash
	amount, err := k.CalculateSlashAmount(walletAddr, 50)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := uint64(5000)
	if amount != expected {
		t.Errorf("expected slash amount %d, got %d", expected, amount)
	}

	// Test percentage > max (should cap at max)
	amount, err = k.CalculateSlashAmount(walletAddr, 75)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should be capped at 50% (params default)
	if amount != 5000 {
		t.Errorf("expected slash amount capped at 5000, got %d", amount)
	}
}

func TestGetPendingAppeals(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	currentTime := time.Now().Unix()
	k.SetCurrentHeight(100)
	k.SetCurrentTime(currentTime)

	// Create multiple slash records with different states
	walletAddr1 := "aura1test1"
	walletAddr2 := "aura1test2"

	record1 := types.UserConfidenceRecord{
		WalletAddress: walletAddr1,
		TotalScore:    8000,
	}
	k.SetUserRecord(record1)

	record2 := types.UserConfidenceRecord{
		WalletAddress: walletAddr2,
		TotalScore:    7000,
	}
	k.SetUserRecord(record2)

	// Slash 1: Appealed, not resolved
	_, _, _, slashTxHash1, _ := k.SlashScore(walletAddr1, "IR-001", 2000, types.SlashReasonFraudDetected, "gov1", "ev1")
	params := k.GetParams()
	k.AppealSlash(walletAddr1, slashTxHash1, "counter_ev1", params.AppealDeposit)

	// Slash 2: Appealed and resolved
	_, _, _, slashTxHash2, _ := k.SlashScore(walletAddr2, "IR-002", 3000, types.SlashReasonCollusion, "gov1", "ev2")
	k.AppealSlash(walletAddr2, slashTxHash2, "counter_ev2", params.AppealDeposit)
	k.ResolveAppeal(walletAddr2, slashTxHash2, false, "gov1", "denied")

	pending := k.GetPendingAppeals()

	// Should only have slash 1
	if len(pending) != 1 {
		t.Errorf("expected 1 pending appeal, got %d", len(pending))
	}

	if len(pending) > 0 && pending[0].WalletAddress != walletAddr1 {
		t.Errorf("expected pending appeal for %s, got %s", walletAddr1, pending[0].WalletAddress)
	}
}

func TestIsSlashAppealed(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	walletAddr := "aura1test"

	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	k.SetUserRecord(record)

	_, _, _, slashTxHash, _ := k.SlashScore(
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	// Not yet appealed
	if k.IsSlashAppealed(walletAddr, slashTxHash) {
		t.Error("expected slash not to be appealed")
	}

	// Appeal it
	params := k.GetParams()
	k.AppealSlash(walletAddr, slashTxHash, "evidence", params.AppealDeposit)

	// Now should be appealed
	if !k.IsSlashAppealed(walletAddr, slashTxHash) {
		t.Error("expected slash to be appealed")
	}

	// Nonexistent slash
	if k.IsSlashAppealed(walletAddr, "nonexistent") {
		t.Error("expected false for nonexistent slash")
	}
}

func TestIsSlashResolved(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	walletAddr := "aura1test"

	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	k.SetUserRecord(record)

	_, _, _, slashTxHash, _ := k.SlashScore(
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	// Not yet resolved
	if k.IsSlashResolved(walletAddr, slashTxHash) {
		t.Error("expected slash not to be resolved")
	}

	// Appeal and resolve
	params := k.GetParams()
	k.AppealSlash(walletAddr, slashTxHash, "evidence", params.AppealDeposit)
	k.ResolveAppeal(walletAddr, slashTxHash, true, "gov1", "notes")

	// Now should be resolved
	if !k.IsSlashResolved(walletAddr, slashTxHash) {
		t.Error("expected slash to be resolved")
	}
}
