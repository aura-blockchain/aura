package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

func TestSlashScore_Success(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	// Create user record
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
		Status:        types.VerificationStatusVerified,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	previousScore, newScore, verificationRevoked, slashTxHash, err := k.SlashScore(
		ctx,
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence_hash_123",
	)

	require.NoError(t, err)
	require.Equal(t, uint64(10000), previousScore)
	require.Equal(t, uint64(8000), newScore)

	// 8000 < 10000 threshold, so verification should be revoked
	require.True(t, verificationRevoked, "expected verification to be revoked when score drops below threshold")
	require.NotEmpty(t, slashTxHash)

	// Verify slash record was created
	slashRecord, ok := k.GetSlashRecord(ctx, walletAddr, slashTxHash)
	require.True(t, ok, "expected slash record to exist")
	require.Equal(t, uint64(2000), slashRecord.SlashAmount)
	require.Equal(t, types.SlashReasonFraudDetected, slashRecord.Reason)
}

func TestSlashScore_InvalidInputs(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)

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
				ctx,
				tt.walletAddr,
				"IR-001",
				tt.slashAmount,
				types.SlashReasonFraudDetected,
				"gov1",
				"evidence",
			)

			require.ErrorIs(t, err, tt.expectError)
		})
	}
}

func TestSlashScore_UserNotFound(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)

	_, _, _, _, err := k.SlashScore(
		ctx,
		"nonexistent",
		"IR-001",
		1000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	require.ErrorIs(t, err, types.ErrUserRecordNotFound)
}

func TestSlashScore_ExceedsTotalScore(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	// Create user with low score
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    1000,
		Status:        types.VerificationStatusUnverified,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	// Slash more than total
	_, newScore, _, _, err := k.SlashScore(
		ctx,
		walletAddr,
		"IR-001",
		5000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	require.NoError(t, err)

	// Score should be 0, not negative
	require.Equal(t, uint64(0), newScore)
}

func TestSlashScore_VerificationRevoked(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	// Create verified user
	record := types.UserConfidenceRecord{
		WalletAddress:              walletAddr,
		TotalScore:                 10000,
		Status:                     types.VerificationStatusVerified,
		VerificationAchievedHeight: 50,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	// Slash enough to drop below threshold
	_, newScore, verificationRevoked, _, err := k.SlashScore(
		ctx,
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	require.NoError(t, err)
	require.Equal(t, uint64(8000), newScore)

	// Verification should be revoked since 8000 < 10000 threshold
	require.True(t, verificationRevoked, "expected verification to be revoked when score drops below threshold")

	// Verify record was updated
	updatedRecord, ok := k.GetUserRecord(ctx, walletAddr)
	require.True(t, ok)
	require.Equal(t, types.VerificationStatusUnverified, updatedRecord.Status)
	require.Equal(t, uint64(0), updatedRecord.VerificationAchievedHeight)
}

func TestAppealSlash_Success(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	// Create slash record first
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	_, _, _, slashTxHash, _ := k.SlashScore(
		ctx,
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	// Appeal the slash
	params, _ := k.GetParams(ctx)
	appealAccepted, reviewDeadline, err := k.AppealSlash(
		ctx,
		walletAddr,
		slashTxHash,
		"counter_evidence",
		params.AppealDeposit,
	)

	require.NoError(t, err)
	require.True(t, appealAccepted, "expected appeal to be accepted")
	require.NotZero(t, reviewDeadline, "expected review deadline to be set")

	// Verify slash record was updated
	slashRecord, ok := k.GetSlashRecord(ctx, walletAddr, slashTxHash)
	require.True(t, ok)
	require.True(t, slashRecord.Appealed, "expected slash to be marked as appealed")
}

func TestAppealSlash_NotFound(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)

	_, _, err := k.AppealSlash(
		ctx,
		"aura1test",
		"nonexistent_slash",
		"evidence",
		"1000uaura",
	)

	require.ErrorIs(t, err, types.ErrSlashNotFound)
}

func TestAppealSlash_AlreadyAppealed(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	// Create and appeal slash
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	_, _, _, slashTxHash, _ := k.SlashScore(
		ctx,
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	params, _ := k.GetParams(ctx)
	_, _, err := k.AppealSlash(ctx, walletAddr, slashTxHash, "evidence", params.AppealDeposit)
	require.NoError(t, err)

	// Try to appeal again
	_, _, err = k.AppealSlash(ctx, walletAddr, slashTxHash, "more_evidence", params.AppealDeposit)

	require.ErrorIs(t, err, types.ErrSlashAlreadyAppealed)
}

func TestAppealSlash_Expired(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	currentTime := time.Now()
	ctx = ctx.WithBlockHeight(100).WithBlockTime(currentTime)

	walletAddr := "aura1test"

	// Create slash record
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	_, _, _, slashTxHash, _ := k.SlashScore(
		ctx,
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	// Move time forward past appeal deadline (14 days + 1)
	ctx = ctx.WithBlockTime(currentTime.Add(15 * 24 * time.Hour))

	params, _ := k.GetParams(ctx)
	_, _, err := k.AppealSlash(ctx, walletAddr, slashTxHash, "evidence", params.AppealDeposit)

	require.ErrorIs(t, err, types.ErrAppealExpired)
}

func TestAppealSlash_InsufficientDeposit(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	_, _, _, slashTxHash, _ := k.SlashScore(
		ctx,
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	// Wrong deposit amount
	_, _, err := k.AppealSlash(ctx, walletAddr, slashTxHash, "evidence", "500uaura")

	require.ErrorIs(t, err, types.ErrInsufficientDeposit)
}

func TestResolveAppeal_RestoreScore(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	// Create user with initial score
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	_, _, _, slashTxHash, _ := k.SlashScore(
		ctx,
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	params, _ := k.GetParams(ctx)
	_, _, err := k.AppealSlash(ctx, walletAddr, slashTxHash, "counter_evidence", params.AppealDeposit)
	require.NoError(t, err)

	// Resolve appeal - restore score
	restoredScore, depositReturned, err := k.ResolveAppeal(
		ctx,
		walletAddr,
		slashTxHash,
		true,
		"gov1",
		"appeal upheld",
	)

	require.NoError(t, err)
	require.Equal(t, uint64(2000), restoredScore)
	require.True(t, depositReturned, "expected deposit to be returned")

	// Verify score was restored
	// Note: The score was 10000, slashed 2000 to become 8000
	// When we restore 2000, it should go back to 10000
	updatedRecord, ok := k.GetUserRecord(ctx, walletAddr)
	require.True(t, ok)
	expectedScore := uint64(10000) // 8000 + 2000 restored
	require.Equal(t, expectedScore, updatedRecord.TotalScore)

	// Verify slash is marked as resolved
	slashRecord, ok := k.GetSlashRecord(ctx, walletAddr, slashTxHash)
	require.True(t, ok)
	require.True(t, slashRecord.Resolved, "expected slash to be marked as resolved")
}

func TestResolveAppeal_DenyAppeal(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	_, _, _, slashTxHash, _ := k.SlashScore(
		ctx,
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	params, _ := k.GetParams(ctx)
	_, _, err := k.AppealSlash(ctx, walletAddr, slashTxHash, "counter_evidence", params.AppealDeposit)
	require.NoError(t, err)

	// Resolve appeal - deny (don't restore score)
	restoredScore, depositReturned, err := k.ResolveAppeal(
		ctx,
		walletAddr,
		slashTxHash,
		false,
		"gov1",
		"appeal denied",
	)

	require.NoError(t, err)
	require.Equal(t, uint64(0), restoredScore)
	require.False(t, depositReturned, "expected deposit not to be returned")

	// Verify score was not restored
	updatedRecord, ok := k.GetUserRecord(ctx, walletAddr)
	require.True(t, ok)
	require.Equal(t, uint64(8000), updatedRecord.TotalScore)
}

func TestResolveAppeal_NotAppealed(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	_, _, _, slashTxHash, _ := k.SlashScore(
		ctx,
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	// Try to resolve without appeal
	_, _, err := k.ResolveAppeal(ctx, walletAddr, slashTxHash, true, "gov1", "notes")

	require.Error(t, err, "expected error for non-appealed slash")
}

func TestResolveAppeal_AlreadyResolved(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	_, _, _, slashTxHash, _ := k.SlashScore(
		ctx,
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	params, _ := k.GetParams(ctx)
	_, _, err := k.AppealSlash(ctx, walletAddr, slashTxHash, "evidence", params.AppealDeposit)
	require.NoError(t, err)

	// Resolve first time
	_, _, err = k.ResolveAppeal(ctx, walletAddr, slashTxHash, true, "gov1", "notes")
	require.NoError(t, err)

	// Try to resolve again
	_, _, err = k.ResolveAppeal(ctx, walletAddr, slashTxHash, true, "gov1", "notes")

	require.ErrorIs(t, err, types.ErrAppealAlreadyResolved)
}

func TestCalculateSlashAmount(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)

	walletAddr := "aura1test"

	// No record
	_, err := k.CalculateSlashAmount(ctx, walletAddr, 50)
	require.ErrorIs(t, err, types.ErrUserRecordNotFound)

	// Create record
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	// Calculate 50% slash
	amount, err := k.CalculateSlashAmount(ctx, walletAddr, 50)
	require.NoError(t, err)

	expected := uint64(5000)
	require.Equal(t, expected, amount)

	// Test percentage > max (should cap at max)
	amount, err = k.CalculateSlashAmount(ctx, walletAddr, 75)
	require.NoError(t, err)

	// Should be capped at 50% (params default)
	require.Equal(t, uint64(5000), amount)
}

func TestGetPendingAppeals(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	// Create multiple slash records with different states
	walletAddr1 := "aura1test1"
	walletAddr2 := "aura1test2"

	record1 := types.UserConfidenceRecord{
		WalletAddress: walletAddr1,
		TotalScore:    8000,
	}
	require.NoError(t, k.SetUserRecord(ctx, record1))

	record2 := types.UserConfidenceRecord{
		WalletAddress: walletAddr2,
		TotalScore:    7000,
	}
	require.NoError(t, k.SetUserRecord(ctx, record2))

	// Slash 1: Appealed, not resolved
	_, _, _, slashTxHash1, _ := k.SlashScore(ctx, walletAddr1, "IR-001", 2000, types.SlashReasonFraudDetected, "gov1", "ev1")
	params, _ := k.GetParams(ctx)
	_, _, err := k.AppealSlash(ctx, walletAddr1, slashTxHash1, "counter_ev1", params.AppealDeposit)
	require.NoError(t, err)

	// Slash 2: Appealed and resolved
	_, _, _, slashTxHash2, _ := k.SlashScore(ctx, walletAddr2, "IR-002", 3000, types.SlashReasonCollusion, "gov1", "ev2")
	_, _, err = k.AppealSlash(ctx, walletAddr2, slashTxHash2, "counter_ev2", params.AppealDeposit)
	require.NoError(t, err)
	_, _, err = k.ResolveAppeal(ctx, walletAddr2, slashTxHash2, false, "gov1", "denied")
	require.NoError(t, err)

	pending := k.GetPendingAppeals(ctx)

	// Should only have slash 1
	require.Len(t, pending, 1)
	require.Equal(t, walletAddr1, pending[0].WalletAddress)
}

func TestIsSlashAppealed(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	_, _, _, slashTxHash, _ := k.SlashScore(
		ctx,
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	// Not yet appealed
	require.False(t, k.IsSlashAppealed(ctx, walletAddr, slashTxHash), "expected slash not to be appealed")

	// Appeal it
	params, _ := k.GetParams(ctx)
	_, _, err := k.AppealSlash(ctx, walletAddr, slashTxHash, "evidence", params.AppealDeposit)
	require.NoError(t, err)

	// Now should be appealed
	require.True(t, k.IsSlashAppealed(ctx, walletAddr, slashTxHash), "expected slash to be appealed")

	// Nonexistent slash
	require.False(t, k.IsSlashAppealed(ctx, walletAddr, "nonexistent"), "expected false for nonexistent slash")
}

func TestIsSlashResolved(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1test"

	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    8000,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	_, _, _, slashTxHash, _ := k.SlashScore(
		ctx,
		walletAddr,
		"IR-001",
		2000,
		types.SlashReasonFraudDetected,
		"gov1",
		"evidence",
	)

	// Not yet resolved
	require.False(t, k.IsSlashResolved(ctx, walletAddr, slashTxHash), "expected slash not to be resolved")

	// Appeal and resolve
	params, _ := k.GetParams(ctx)
	_, _, err := k.AppealSlash(ctx, walletAddr, slashTxHash, "evidence", params.AppealDeposit)
	require.NoError(t, err)
	_, _, err = k.ResolveAppeal(ctx, walletAddr, slashTxHash, true, "gov1", "notes")
	require.NoError(t, err)

	// Now should be resolved
	require.True(t, k.IsSlashResolved(ctx, walletAddr, slashTxHash), "expected slash to be resolved")
}
