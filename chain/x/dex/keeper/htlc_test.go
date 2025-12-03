package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
)

func TestCreateAndClaimHTLC(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	senderAddr := keepertest.GenTestAddr()
	sender := senderAddr.String()
	recipient := keepertest.GenTestAddr().String()
	secret := "super-secret"
	hash := k.GenerateSecureHash([]byte(secret))
	amount := sdk.NewCoin("uaura", sdkmath.NewInt(1_000000))

	// Fund sender account before creating HTLC
	mockBank.SetBalance(senderAddr, "uaura", sdkmath.NewInt(1_000000))

	htlcID, err := k.CreateHTLC(ctx, sender, recipient, amount, hash, 600)
	require.NoError(t, err)
	require.NotEmpty(t, htlcID)

	data, found := k.GetHTLC(ctx, htlcID)
	require.True(t, found)
	require.Equal(t, "active", data.Status)

	err = k.ClaimHTLC(ctx, recipient, htlcID, secret)
	require.NoError(t, err)

	claimed, found := k.GetHTLC(ctx, htlcID)
	require.True(t, found)
	require.Equal(t, "claimed", claimed.Status)
	require.Equal(t, secret, claimed.Secret)
}

func TestRefundHTLC(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	senderAddr := keepertest.GenTestAddr()
	sender := senderAddr.String()
	recipient := keepertest.GenTestAddr().String()
	hash := k.GenerateSecureHash([]byte("refund-secret"))
	amount := sdk.NewCoin("uaura", sdkmath.NewInt(500_000))

	// Fund sender account before creating HTLC
	mockBank.SetBalance(senderAddr, "uaura", sdkmath.NewInt(500_000))

	htlcID, err := k.CreateHTLC(ctx, sender, recipient, amount, hash, 1)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Second))

	err = k.RefundHTLC(ctx, sender, htlcID)
	require.NoError(t, err)

	data, found := k.GetHTLC(ctx, htlcID)
	require.True(t, found)
	require.Equal(t, "refunded", data.Status)

	err = k.RefundHTLC(ctx, sender, htlcID)
	require.Error(t, err)
}

func TestCleanupExpiredHTLCs(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	senderAddr := keepertest.GenTestAddr()
	sender := senderAddr.String()
	recipient := keepertest.GenTestAddr().String()
	hash := k.GenerateSecureHash([]byte("cleanup-secret"))
	amount := sdk.NewCoin("uaura", sdkmath.NewInt(250_000))

	// Fund sender account before creating HTLC
	mockBank.SetBalance(senderAddr, "uaura", sdkmath.NewInt(250_000))

	htlcID, err := k.CreateHTLC(ctx, sender, recipient, amount, hash, 1)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Second))

	k.CleanupExpiredHTLCs(ctx)

	data, found := k.GetHTLC(ctx, htlcID)
	require.True(t, found)
	require.Equal(t, "refunded", data.Status)
}
