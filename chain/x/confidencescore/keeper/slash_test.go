package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

func TestSlashScore(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	wallet := "aura1test"
	record := types.UserConfidenceRecord{
		WalletAddress: wallet,
		TotalScore:    10000,
		Status:        types.VerificationStatusVerified,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	prev, newScore, revoked, txHash, err := k.SlashScore(ctx, wallet, "IR-001", 2000, types.SlashReasonFraudDetected, "gov1", "evidence")
	require.NoError(t, err)
	require.Equal(t, uint64(10000), prev)
	require.Equal(t, uint64(8000), newScore)
	_ = revoked
	require.NotEmpty(t, txHash)

	// Slash more than balance should floor at zero
	prev, newScore, revoked, _, err = k.SlashScore(ctx, wallet, "IR-002", 9000, types.SlashReasonCollusion, "gov1", "evidence2")
	require.NoError(t, err)
	require.Equal(t, uint64(8000), prev)
	require.Equal(t, uint64(0), newScore)
	_ = revoked
}

func TestAppealSlashFlow(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(50).WithBlockTime(time.Now())

	wallet := "aura1appeal"
	record := types.UserConfidenceRecord{
		WalletAddress: wallet,
		TotalScore:    5000,
		Status:        types.VerificationStatusUnverified,
	}
	require.NoError(t, k.SetUserRecord(ctx, record))

	_, _, _, txHash, err := k.SlashScore(ctx, wallet, "IR-010", 1000, types.SlashReasonFraudDetected, "gov1", "ev1")
	require.NoError(t, err)

	params := k.GetParams()
	accepted, reviewDeadline, err := k.AppealSlash(ctx, wallet, txHash, "counter", params.AppealDeposit)
	require.NoError(t, err)
	require.True(t, accepted)
	require.Greater(t, reviewDeadline, int64(0))

	restored, returned, err := k.ResolveAppeal(ctx, wallet, txHash, true, "gov1", "resolved")
	require.NoError(t, err)
	require.True(t, returned)
	require.Equal(t, uint64(1000), restored)
}
