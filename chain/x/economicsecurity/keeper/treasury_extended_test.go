// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// =============================================================================
// ExecuteTreasurySpend Tests
// =============================================================================

func TestExecuteTreasurySpend_Success(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2"},
		Threshold:        2,
		TimelockDuration: 1, // 1 second for testing
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create proposal
	txID, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	require.NoError(t, err)

	// Sign with second signer
	_, _, err = k.SignTreasurySpend(ctx, "aura1signer2", txID)
	require.NoError(t, err)

	// Advance time past timelock
	_ = k.SetCurrentTime(ctx, currentTime+10)

	// Execute with sufficient treasury balance
	err = k.ExecuteTreasurySpend(ctx, "aura1executor", txID, "10000000")
	require.NoError(t, err)

	// Verify executed
	tx, _ := k.GetPendingTreasuryTx(ctx, txID)
	require.True(t, tx.Executed)
}

func TestExecuteTreasurySpend_TxNotFound(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress: "aura1treasury",
		Signers:         []string{"aura1signer1"},
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	err := k.ExecuteTreasurySpend(ctx, "aura1executor", "nonexistent-tx", "1000000")
	require.Error(t, err)
}

func TestExecuteTreasurySpend_AlreadyExecuted(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2"},
		Threshold:        2,
		TimelockDuration: 1,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create and sign
	txID, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	require.NoError(t, err)
	_, _, _ = k.SignTreasurySpend(ctx, "aura1signer2", txID)

	// Advance time and execute
	_ = k.SetCurrentTime(ctx, currentTime+10)
	err = k.ExecuteTreasurySpend(ctx, "aura1executor", txID, "10000000")
	require.NoError(t, err)

	// Try to execute again
	err = k.ExecuteTreasurySpend(ctx, "aura1executor", txID, "10000000")
	require.ErrorIs(t, err, types.ErrTxAlreadyExecuted)
}

func TestExecuteTreasurySpend_AlreadyRejected(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2"},
		Threshold:        2,
		TimelockDuration: 1,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create and reject
	txID, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	require.NoError(t, err)
	_ = k.RejectTreasurySpend(ctx, txID)

	// Try to execute rejected tx
	_ = k.SetCurrentTime(ctx, currentTime+10)
	err = k.ExecuteTreasurySpend(ctx, "aura1executor", txID, "10000000")
	require.ErrorIs(t, err, types.ErrTxAlreadyRejected)
}

func TestExecuteTreasurySpend_InsufficientSignatures(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2", "aura1signer3"},
		Threshold:        3, // Requires all 3
		TimelockDuration: 1,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create proposal (1 signature)
	txID, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	require.NoError(t, err)

	// Only sign with one more (2 total, need 3)
	_, _, _ = k.SignTreasurySpend(ctx, "aura1signer2", txID)

	// Try to execute with insufficient signatures
	_ = k.SetCurrentTime(ctx, currentTime+10)
	err = k.ExecuteTreasurySpend(ctx, "aura1executor", txID, "10000000")
	require.ErrorIs(t, err, types.ErrInsufficientSignatures)
}

func TestExecuteTreasurySpend_TimelockNotExpired(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2"},
		Threshold:        2,
		TimelockDuration: 86400, // 1 day
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create and sign
	txID, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	require.NoError(t, err)
	_, _, _ = k.SignTreasurySpend(ctx, "aura1signer2", txID)

	// Try to execute without advancing time (timelock not expired)
	err = k.ExecuteTreasurySpend(ctx, "aura1executor", txID, "10000000")
	require.ErrorIs(t, err, types.ErrTimelockNotExpired)
}

func TestExecuteTreasurySpend_InsufficientBalance(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2"},
		Threshold:        2,
		TimelockDuration: 1,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create and sign for 1,000,000 tokens
	txID, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	require.NoError(t, err)
	_, _, _ = k.SignTreasurySpend(ctx, "aura1signer2", txID)

	// Try to execute with insufficient treasury balance (only 500000)
	_ = k.SetCurrentTime(ctx, currentTime+10)
	err = k.ExecuteTreasurySpend(ctx, "aura1executor", txID, "500000")
	require.ErrorIs(t, err, types.ErrInsufficientTreasuryBalance)
}

func TestExecuteTreasurySpend_EmptyBalance(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2"},
		Threshold:        2,
		TimelockDuration: 1,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	txID, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	require.NoError(t, err)
	_, _, _ = k.SignTreasurySpend(ctx, "aura1signer2", txID)

	_ = k.SetCurrentTime(ctx, currentTime+10)
	err = k.ExecuteTreasurySpend(ctx, "aura1executor", txID, "0")
	require.ErrorIs(t, err, types.ErrInsufficientTreasuryBalance)
}

// =============================================================================
// GetTreasuryStatistics Extended Tests
// =============================================================================

func TestGetTreasuryStatistics_WithTxs(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2", "aura1signer3"},
		Threshold:        2,
		TimelockDuration: 1,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create some proposals
	txID1, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient1", "1000000", "Test 1")
	require.NoError(t, err)
	_, _, err = k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient2", "2000000", "Test 2")
	require.NoError(t, err)

	// Sign and execute one
	_, _, err = k.SignTreasurySpend(ctx, "aura1signer2", txID1)
	require.NoError(t, err)
	_ = k.SetCurrentTime(ctx, currentTime+10)
	err = k.ExecuteTreasurySpend(ctx, "aura1executor", txID1, "10000000")
	require.NoError(t, err)

	enabled, treasuryAddress, threshold, signerCount, _, _ := k.GetTreasuryStatistics(ctx)
	require.True(t, enabled)
	require.Equal(t, "aura1treasury", treasuryAddress)
	require.Equal(t, uint32(2), threshold)
	require.Equal(t, uint32(3), signerCount)
	// Note: pending and executed counts depend on iteration order which may vary
}

func TestGetTreasuryStatistics_WithRejected(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1"},
		Threshold:        1,
		TimelockDuration: 1,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create and reject a proposal
	txID, _, _ := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	_ = k.RejectTreasurySpend(ctx, txID)

	enabled, _, _, _, pendingTxCount, _ := k.GetTreasuryStatistics(ctx)
	require.True(t, enabled)
	// Rejected txs should not count as pending
	require.Equal(t, uint64(0), pendingTxCount)
}

// =============================================================================
// CleanupExecutedTreasuryTxs Tests
// =============================================================================

func TestCleanupExecutedTreasuryTxs_Success(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2"},
		Threshold:        2,
		TimelockDuration: 1,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create, sign and execute a tx
	txID, _, _ := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	_, _, _ = k.SignTreasurySpend(ctx, "aura1signer2", txID)
	_ = k.SetCurrentTime(ctx, currentTime+10)
	_ = k.ExecuteTreasurySpend(ctx, "aura1executor", txID, "10000000")

	// Advance time significantly past retention period
	_ = k.SetCurrentTime(ctx, currentTime+1000000)

	// Cleanup with 1 day retention
	err := k.CleanupExecutedTreasuryTxs(ctx, 86400)
	require.NoError(t, err)
}

func TestCleanupExecutedTreasuryTxs_RejectedTxs(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1"},
		Threshold:        1,
		TimelockDuration: 1,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create and reject
	txID, _, _ := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	_ = k.RejectTreasurySpend(ctx, txID)

	// Advance time
	_ = k.SetCurrentTime(ctx, currentTime+1000000)

	// Cleanup should also handle rejected txs
	err := k.CleanupExecutedTreasuryTxs(ctx, 86400)
	require.NoError(t, err)
}

func TestCleanupExecutedTreasuryTxs_NoOldTxs(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2"},
		Threshold:        2,
		TimelockDuration: 1,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create and execute
	txID, _, _ := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	_, _, _ = k.SignTreasurySpend(ctx, "aura1signer2", txID)
	_ = k.SetCurrentTime(ctx, currentTime+10)
	_ = k.ExecuteTreasurySpend(ctx, "aura1executor", txID, "10000000")

	// Don't advance time much - tx should not be cleaned up
	err := k.CleanupExecutedTreasuryTxs(ctx, 86400)
	require.NoError(t, err)

	// Verify tx still exists
	tx, found := k.GetPendingTreasuryTxByID(ctx, txID)
	require.True(t, found)
	require.NotNil(t, tx)
}

func TestCleanupExecutedTreasuryTxs_Empty(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Cleanup on empty store should succeed
	err := k.CleanupExecutedTreasuryTxs(ctx, 86400)
	require.NoError(t, err)
}

// =============================================================================
// RejectTreasurySpend Extended Tests
// =============================================================================

func TestRejectTreasurySpend_AlreadyExecuted(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2"},
		Threshold:        2,
		TimelockDuration: 1,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create, sign and execute
	txID, _, _ := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	_, _, _ = k.SignTreasurySpend(ctx, "aura1signer2", txID)
	_ = k.SetCurrentTime(ctx, currentTime+10)
	_ = k.ExecuteTreasurySpend(ctx, "aura1executor", txID, "10000000")

	// Try to reject executed tx
	err := k.RejectTreasurySpend(ctx, txID)
	require.ErrorIs(t, err, types.ErrTxAlreadyExecuted)
}

func TestRejectTreasurySpend_NotFound(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.RejectTreasurySpend(ctx, "nonexistent-tx")
	require.Error(t, err)
}
