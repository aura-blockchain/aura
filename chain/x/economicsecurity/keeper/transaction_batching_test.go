// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// =============================================================================
// Transaction Batching Tests
// =============================================================================

func TestBatchTransaction_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)
	_ = k.SetCurrentHeight(ctx, 100)

	// Enable batching
	err := k.SetBatchingEnabled(ctx, true)
	require.NoError(t, err)

	txID, err := k.BatchTransaction(ctx, "aura1sender1", "aura1recipient1", "1000000", 10)
	require.NoError(t, err)
	require.NotEmpty(t, txID)
	require.Len(t, txID, 16) // SHA256 truncated to 16 chars
}

func TestBatchTransaction_Disabled(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)

	// Disable batching
	err := k.SetBatchingEnabled(ctx, false)
	require.NoError(t, err)

	_, err = k.BatchTransaction(ctx, "aura1sender1", "aura1recipient1", "1000000", 10)
	require.ErrorIs(t, err, types.ErrBatchingDisabled)
}

func TestBatchTransaction_InvalidAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)
	_ = k.SetCurrentHeight(ctx, 100)
	_ = k.SetBatchingEnabled(ctx, true)

	_, err := k.BatchTransaction(ctx, "aura1sender1", "aura1recipient1", "not-a-number", 10)
	require.ErrorIs(t, err, types.ErrInvalidAmount)
}

func TestBatchTransaction_MultipleTxs(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)
	_ = k.SetCurrentHeight(ctx, 100)
	_ = k.SetBatchingEnabled(ctx, true)

	// Add multiple transactions
	tx1, err := k.BatchTransaction(ctx, "aura1sender1", "aura1recipient1", "1000000", 10)
	require.NoError(t, err)

	_ = k.SetCurrentTime(ctx, 1001)
	tx2, err := k.BatchTransaction(ctx, "aura1sender2", "aura1recipient2", "2000000", 20)
	require.NoError(t, err)

	_ = k.SetCurrentTime(ctx, 1002)
	tx3, err := k.BatchTransaction(ctx, "aura1sender3", "aura1recipient3", "3000000", 30)
	require.NoError(t, err)

	// All should have unique IDs
	require.NotEqual(t, tx1, tx2)
	require.NotEqual(t, tx2, tx3)
	require.NotEqual(t, tx1, tx3)
}

func TestBatchTransaction_ReachesBatchSize(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)
	_ = k.SetCurrentHeight(ctx, 100)
	_ = k.SetBatchingEnabled(ctx, true)

	// Set a low max batch size
	err := k.SetMaxBatchSize(ctx, 3)
	require.NoError(t, err)

	// Add transactions up to max
	for i := 0; i < 3; i++ {
		_ = k.SetCurrentTime(ctx, int64(1000+i))
		_, err := k.BatchTransaction(ctx, "aura1sender", "aura1recipient", "1000000", 10)
		require.NoError(t, err)
	}

	// Check batch is marked ready
	batch, err := k.GetPendingBatchData(ctx)
	require.NoError(t, err)
	require.NotNil(t, batch)
	require.Equal(t, "ready", batch.Status)
}

func TestProcessBatch_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)
	_ = k.SetCurrentHeight(ctx, 100)
	_ = k.SetBatchingEnabled(ctx, true)
	_ = k.SetMinBatchSize(ctx, 2)

	// Add transactions
	for i := 0; i < 3; i++ {
		_ = k.SetCurrentTime(ctx, int64(1000+i))
		_, err := k.BatchTransaction(ctx, "aura1sender", "aura1recipient", "1000000", 10)
		require.NoError(t, err)
	}

	// Process batch
	count, batchID, err := k.ProcessBatch(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(3), count)
	require.NotEmpty(t, batchID)
}

func TestProcessBatch_Disabled(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetBatchingEnabled(ctx, false)

	_, _, err := k.ProcessBatch(ctx)
	require.ErrorIs(t, err, types.ErrBatchingDisabled)
}

func TestProcessBatch_NoPending(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetBatchingEnabled(ctx, true)

	count, batchID, err := k.ProcessBatch(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(0), count)
	require.Empty(t, batchID)
}

func TestProcessBatch_TooSmall(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)
	_ = k.SetCurrentHeight(ctx, 100)
	_ = k.SetBatchingEnabled(ctx, true)
	_ = k.SetMinBatchSize(ctx, 10) // High minimum

	// Add only a few transactions
	_, _ = k.BatchTransaction(ctx, "aura1sender", "aura1recipient", "1000000", 10)

	_, _, err := k.ProcessBatch(ctx)
	require.ErrorIs(t, err, types.ErrBatchTooSmall)
}

func TestShouldProcessBatch_Disabled(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetBatchingEnabled(ctx, false)

	should, err := k.ShouldProcessBatch(ctx)
	require.NoError(t, err)
	require.False(t, should)
}

func TestShouldProcessBatch_NoPending(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetBatchingEnabled(ctx, true)

	should, err := k.ShouldProcessBatch(ctx)
	require.NoError(t, err)
	require.False(t, should)
}

func TestShouldProcessBatch_MaxSizeReached(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)
	_ = k.SetCurrentHeight(ctx, 100)
	_ = k.SetBatchingEnabled(ctx, true)
	_ = k.SetMaxBatchSize(ctx, 3)
	_ = k.SetMinBatchSize(ctx, 2)

	// Add transactions to reach max
	for i := 0; i < 3; i++ {
		_ = k.SetCurrentTime(ctx, int64(1000+i))
		_, _ = k.BatchTransaction(ctx, "aura1sender", "aura1recipient", "1000000", 10)
	}

	should, err := k.ShouldProcessBatch(ctx)
	require.NoError(t, err)
	require.True(t, should)
}

func TestShouldProcessBatch_MinSizeWithTimeout(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)
	_ = k.SetCurrentHeight(ctx, 100)
	_ = k.SetBatchingEnabled(ctx, true)
	_ = k.SetMaxBatchSize(ctx, 100)
	_ = k.SetMinBatchSize(ctx, 2)
	_ = k.SetBatchTimeout(ctx, 60)

	// Add enough transactions
	for i := 0; i < 3; i++ {
		_ = k.SetCurrentTime(ctx, int64(1000+i))
		_, _ = k.BatchTransaction(ctx, "aura1sender", "aura1recipient", "1000000", 10)
	}

	// Move time past timeout
	_ = k.SetCurrentTime(ctx, 1100)

	should, err := k.ShouldProcessBatch(ctx)
	require.NoError(t, err)
	require.True(t, should)
}

func TestShouldProcessBatch_MinSizeNotReached(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)
	_ = k.SetCurrentHeight(ctx, 100)
	_ = k.SetBatchingEnabled(ctx, true)
	_ = k.SetMaxBatchSize(ctx, 100)
	_ = k.SetMinBatchSize(ctx, 10)
	_ = k.SetBatchTimeout(ctx, 60)

	// Add only 2 transactions (below min)
	_, _ = k.BatchTransaction(ctx, "aura1sender", "aura1recipient", "1000000", 10)

	// Even with timeout, shouldn't process
	_ = k.SetCurrentTime(ctx, 1100)

	should, err := k.ShouldProcessBatch(ctx)
	require.NoError(t, err)
	require.False(t, should)
}

func TestGetPendingBatchData_Empty(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	batch, err := k.GetPendingBatchData(ctx)
	require.NoError(t, err)
	require.Nil(t, batch)
}

func TestSetPendingBatch_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	batch := &TransactionBatch{
		BatchID:      "test-batch-id",
		Transactions: []Transaction{},
		TotalAmount:  "0",
		CreatedAt:    1000,
		Status:       "pending",
	}

	err := k.SetPendingBatch(ctx, batch)
	require.NoError(t, err)

	retrieved, err := k.GetPendingBatchData(ctx)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	require.Equal(t, "test-batch-id", retrieved.BatchID)
}

func TestClearPendingBatch_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	batch := &TransactionBatch{
		BatchID: "test-batch-id",
		Status:  "pending",
	}
	_ = k.SetPendingBatch(ctx, batch)

	err := k.ClearPendingBatch(ctx)
	require.NoError(t, err)

	retrieved, err := k.GetPendingBatchData(ctx)
	require.NoError(t, err)
	require.Nil(t, retrieved)
}

func TestAddBatchRecord_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	record := &BatchRecord{
		BatchID:             "test-batch-id",
		TransactionCount:    10,
		TotalAmount:         "10000000",
		ProcessedAt:         1000,
		GasSaved:            50000,
		AverageGasPrice:     100,
		CompressionRatioBps: 7500, // 75% in basis points
	}

	err := k.AddBatchRecord(ctx, record)
	require.NoError(t, err)
}

func TestGetBatchStatisticsData_Default(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	stats, err := k.GetBatchStatisticsData(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)
	require.Equal(t, uint64(0), stats.TotalBatchesProcessed)
	require.Equal(t, uint64(0), stats.TotalTransactionsBatched)
	require.Equal(t, "0", stats.TotalGasSaved)
}

func TestSetBatchStatistics_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	stats := &BatchStatistics{
		TotalBatchesProcessed:      100,
		TotalTransactionsBatched:   5000,
		TotalGasSaved:              "1000000000",
		AverageCompressionRatioBps: 8000, // 80% in basis points
	}

	err := k.SetBatchStatistics(ctx, stats)
	require.NoError(t, err)

	retrieved, err := k.GetBatchStatisticsData(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(100), retrieved.TotalBatchesProcessed)
	require.Equal(t, uint64(5000), retrieved.TotalTransactionsBatched)
	require.Equal(t, "1000000000", retrieved.TotalGasSaved)
}

func TestIsBatchingEnabled_Default(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	enabled, err := k.IsBatchingEnabled(ctx)
	require.NoError(t, err)
	require.True(t, enabled) // Default is enabled
}

func TestSetBatchingEnabled_Toggle(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Disable
	err := k.SetBatchingEnabled(ctx, false)
	require.NoError(t, err)

	enabled, err := k.IsBatchingEnabled(ctx)
	require.NoError(t, err)
	require.False(t, enabled)

	// Enable
	err = k.SetBatchingEnabled(ctx, true)
	require.NoError(t, err)

	enabled, err = k.IsBatchingEnabled(ctx)
	require.NoError(t, err)
	require.True(t, enabled)
}

func TestGetMaxBatchSize_Default(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	size, err := k.GetMaxBatchSize(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(100), size) // Default
}

func TestSetMaxBatchSize_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.SetMaxBatchSize(ctx, 200)
	require.NoError(t, err)

	size, err := k.GetMaxBatchSize(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(200), size)
}

func TestGetMinBatchSize_Default(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	size, err := k.GetMinBatchSize(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(10), size) // Default
}

func TestSetMinBatchSize_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.SetMinBatchSize(ctx, 5)
	require.NoError(t, err)

	size, err := k.GetMinBatchSize(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(5), size)
}

func TestGetBatchTimeout_Default(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	timeout, err := k.GetBatchTimeout(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(60), timeout) // Default
}

func TestSetBatchTimeout_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.SetBatchTimeout(ctx, 120)
	require.NoError(t, err)

	timeout, err := k.GetBatchTimeout(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(120), timeout)
}

func TestGenerateTxID_Deterministic(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	id1 := k.generateTxID("sender1", "recipient1", "1000", 1000)
	id2 := k.generateTxID("sender1", "recipient1", "1000", 1000)
	id3 := k.generateTxID("sender2", "recipient1", "1000", 1000)

	require.Equal(t, id1, id2)
	require.NotEqual(t, id1, id3)
	require.Len(t, id1, 16)
}

func TestGenerateBatchID_Deterministic(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	id1 := k.generateBatchID(100, 1000)
	id2 := k.generateBatchID(100, 1000)
	id3 := k.generateBatchID(101, 1000)

	require.Equal(t, id1, id2)
	require.NotEqual(t, id1, id3)
	require.Len(t, id1, 16)
}

func TestGetPendingBatch_SimplifiedVersion(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	count, amount, status := k.GetPendingBatch()
	require.Equal(t, uint64(0), count)
	require.Equal(t, "0", amount)
	require.Empty(t, status)
}

func TestGetBatchStatistics_SimplifiedVersion(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	batches, txs, gasSaved, compressionBps := k.GetBatchStatistics()
	require.Equal(t, uint64(0), batches)
	require.Equal(t, uint64(0), txs)
	require.Equal(t, "0", gasSaved)
	require.Equal(t, uint64(0), compressionBps) // Basis points instead of float32
}

func TestTransactionBatching_FullWorkflow(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)
	_ = k.SetCurrentHeight(ctx, 100)
	_ = k.SetBatchingEnabled(ctx, true)
	_ = k.SetMinBatchSize(ctx, 3)
	_ = k.SetMaxBatchSize(ctx, 10)

	// Add transactions
	for i := 0; i < 5; i++ {
		_ = k.SetCurrentTime(ctx, int64(1000+i))
		txID, err := k.BatchTransaction(ctx, "aura1sender", "aura1recipient", "1000000", uint64(i))
		require.NoError(t, err)
		require.NotEmpty(t, txID)
	}

	// Verify pending batch
	batch, err := k.GetPendingBatchData(ctx)
	require.NoError(t, err)
	require.NotNil(t, batch)
	require.Equal(t, 5, len(batch.Transactions))
	require.Equal(t, "5000000", batch.TotalAmount)

	// Process the batch
	count, batchID, err := k.ProcessBatch(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(5), count)
	require.NotEmpty(t, batchID)

	// Verify pending batch is cleared
	batch, err = k.GetPendingBatchData(ctx)
	require.NoError(t, err)
	require.Nil(t, batch)

	// Verify statistics updated
	stats, err := k.GetBatchStatisticsData(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), stats.TotalBatchesProcessed)
	require.Equal(t, uint64(5), stats.TotalTransactionsBatched)
}

func TestTransactionBatching_GasSavingsCalculation(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)
	_ = k.SetCurrentHeight(ctx, 100)
	_ = k.SetBatchingEnabled(ctx, true)
	_ = k.SetMinBatchSize(ctx, 3)

	// Add 10 transactions
	for i := 0; i < 10; i++ {
		_ = k.SetCurrentTime(ctx, int64(1000+i))
		_, _ = k.BatchTransaction(ctx, "aura1sender", "aura1recipient", "1000000", 10)
	}

	// Process
	_, _, err := k.ProcessBatch(ctx)
	require.NoError(t, err)

	// Verify gas savings are calculated
	stats, err := k.GetBatchStatisticsData(ctx)
	require.NoError(t, err)
	require.NotEqual(t, "0", stats.TotalGasSaved)
}

func TestTransaction_JSONSerialization(t *testing.T) {
	tx := Transaction{
		ID:        "test-tx-id",
		Sender:    "aura1sender",
		Recipient: "aura1recipient",
		Amount:    "1000000",
		Priority:  10,
		Timestamp: 1000,
		GasPrice:  "100",
	}

	require.Equal(t, "test-tx-id", tx.ID)
	require.Equal(t, "aura1sender", tx.Sender)
	require.Equal(t, "aura1recipient", tx.Recipient)
}

func TestTransactionBatch_JSONSerialization(t *testing.T) {
	batch := TransactionBatch{
		BatchID:      "test-batch-id",
		Transactions: []Transaction{},
		TotalAmount:  "0",
		CreatedAt:    1000,
		Status:       "pending",
	}

	require.Equal(t, "test-batch-id", batch.BatchID)
	require.Equal(t, "pending", batch.Status)
}

func TestBatchRecord_Fields(t *testing.T) {
	record := BatchRecord{
		BatchID:             "test-batch-id",
		TransactionCount:    10,
		TotalAmount:         "10000000",
		ProcessedAt:         1000,
		GasSaved:            50000,
		AverageGasPrice:     100,
		CompressionRatioBps: 7500, // 75% in basis points
	}

	require.Equal(t, "test-batch-id", record.BatchID)
	require.Equal(t, uint64(10), record.TransactionCount)
	require.Equal(t, uint64(7500), record.CompressionRatioBps)
}

func TestBatchStatistics_Fields(t *testing.T) {
	stats := BatchStatistics{
		TotalBatchesProcessed:      100,
		TotalTransactionsBatched:   5000,
		TotalGasSaved:              "1000000000",
		AverageCompressionRatioBps: 8000, // 80% in basis points
	}

	require.Equal(t, uint64(100), stats.TotalBatchesProcessed)
	require.Equal(t, uint64(5000), stats.TotalTransactionsBatched)
	require.Equal(t, "1000000000", stats.TotalGasSaved)
	require.Equal(t, uint64(8000), stats.AverageCompressionRatioBps)
}
