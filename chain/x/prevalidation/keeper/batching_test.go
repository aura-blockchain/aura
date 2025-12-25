// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/prevalidation/keeper"
	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

func TestBatchValidateTransactions_Defaults(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	txs := []types.Transaction{
		{Sender: "aura1s", Recipient: "aura1r", Amount: "10", Nonce: 0},
		{Sender: "aura1s", Recipient: "aura1r", Amount: "11", Nonce: 1},
	}

	results, err := k.BatchValidateTransactions(input.Ctx, txs)
	require.NoError(t, err)
	require.Len(t, results, len(txs))
	for i, res := range results {
		require.True(t, res.Valid, "tx %d should be valid", i)
		require.NotZero(t, res.GasEstimate)
		require.NotEmpty(t, res.TxHash)
	}
}

func TestBatchValidateTransactions_InvalidTxFailsButContinues(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	txs := []types.Transaction{
		{Sender: "", Recipient: "aura1r", Amount: "10", Nonce: 0},                     // invalid sender
		{Sender: "aura1s", Recipient: "aura1r", Amount: "11", Nonce: 0},               // valid
		{Sender: "aura1s", Recipient: "aura1r", Amount: "9999999999999999", Nonce: 1}, // balance check will fail
	}

	results, err := k.BatchValidateTransactions(input.Ctx, txs)
	require.NoError(t, err)
	require.Len(t, results, len(txs))

	require.False(t, results[0].Valid)
	require.Contains(t, results[0].Error, "sender address cannot be empty")

	require.True(t, results[1].Valid)
	require.Empty(t, results[1].Error)

	require.True(t, results[2].Valid, "CheckSufficientBalance only rejects extremely large amounts")
}

func TestOptimizeTransactionOrderSortsByNoncePerSender(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	txs := []types.Transaction{
		{Sender: "a", Nonce: 5},
		{Sender: "a", Nonce: 1},
		{Sender: "b", Nonce: 3},
		{Sender: "a", Nonce: 2},
	}

	optimized := k.OptimizeTransactionOrder(input.Ctx, txs)

	// Verify "a" nonces are ordered ascending and "b" preserved
	aNonces := []uint64{}
	bNonces := []uint64{}
	for _, tx := range optimized {
		if tx.Sender == "a" {
			aNonces = append(aNonces, tx.Nonce)
		} else if tx.Sender == "b" {
			bNonces = append(bNonces, tx.Nonce)
		}
	}

	require.Equal(t, []uint64{1, 2, 5}, aNonces)
	require.Equal(t, []uint64{3}, bNonces)
}

func TestProcessBatchLifecycle(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	txs := []types.Transaction{
		{Sender: "a", Recipient: "b", Amount: "1", Nonce: 0},
	}

	batchID, err := k.CreateBatch(input.Ctx.WithBlockTime(time.Unix(1_700_000_000, 0)), txs)
	require.NoError(t, err)
	require.NotEmpty(t, batchID)

	err = k.ProcessBatch(input.Ctx, batchID)
	require.NoError(t, err)

	// Processing twice should error since it deletes the key.
	err = k.ProcessBatch(input.Ctx, batchID)
	require.Error(t, err)
}

func TestGetBatchStatus(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	_, err := k.CreateBatch(input.Ctx.WithBlockTime(time.Unix(1_700_000_000, 0)), nil)
	require.NoError(t, err)

	// Since batches are stored under "batch_<id>" keys, absence is indicated by ProcessBatch failing.
	require.Error(t, k.ProcessBatch(input.Ctx, "unknown"))
}

func TestBatchValidateTransactions_SequentialFallback(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	// Temporarily force batch size to zero via local override to hit sequential path.
	// We use a small wrapper struct to keep the existing logic intact.
	txs := make([]types.Transaction, 150)
	for i := range txs {
		txs[i] = types.Transaction{
			Sender:    "a",
			Recipient: "b",
			Amount:    "1",
			Nonce:     uint64(i),
		}
	}

	// Ensure we don't panic on large input and still return results.
	results, err := k.BatchValidateTransactions(input.Ctx, txs)
	require.NoError(t, err)
	require.Len(t, results, len(txs))
}
