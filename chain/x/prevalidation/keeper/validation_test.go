package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/prevalidation/keeper"
	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

func TestValidateTransactionNonceBounds(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	// Set existing nonce so lower values are rejected.
	k.SetNonce(input.Ctx, "aura1sender", 5)

	tx := types.Transaction{Sender: "aura1sender", Recipient: "aura1rcpt", Amount: "1", Nonce: 3}
	valid, err := k.ValidateTransaction(input.Ctx, tx)
	require.False(t, valid)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid nonce")

	// Nonce equal to current should pass.
	tx.Nonce = 5
	valid, err = k.ValidateTransaction(input.Ctx, tx)
	require.True(t, valid)
	require.NoError(t, err)
}

func TestCheckRateLimitAndRecordTransaction(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	addr := "aura1rate"

	require.True(t, k.CheckRateLimit(input.Ctx, addr))
	for i := 0; i < 49; i++ {
		k.RecordTransaction(input.Ctx, addr)
	}
	require.True(t, k.CheckRateLimit(input.Ctx, addr), "49 records still under limit")

	k.RecordTransaction(input.Ctx, addr) // 50th should hit limit
	require.False(t, k.CheckRateLimit(input.Ctx, addr), "50th record should exceed limit")
}

func TestValidateSignaturePaths(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	require.False(t, k.ValidateSignature(input.Ctx, "signer", []byte("short"), []byte("tiny")))
	require.False(t, k.ValidateSignature(input.Ctx, "signer", []byte("msg"), []byte("invalid_signature")))
	require.True(t, k.ValidateSignature(input.Ctx, "signer", []byte("msg"), []byte("valid_signature")))
}
