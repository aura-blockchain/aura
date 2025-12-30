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
// Treasury Tests
// =============================================================================

func TestProposeTreasurySpend_Success(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2", "aura1signer3"},
		Threshold:        2,
		TimelockDuration: 86400, // 1 day
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	txID, executableAt, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test proposal")
	require.NoError(t, err)
	require.NotEmpty(t, txID)
	require.False(t, executableAt.IsZero())
	require.True(t, executableAt.After(time.Unix(currentTime, 0)))
}

func TestProposeTreasurySpend_InvalidSigner(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress: "aura1treasury",
		Signers:         []string{"aura1signer1", "aura1signer2"},
		Threshold:       2,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	_, _, err := k.ProposeTreasurySpend(ctx, "aura1notasigner", "aura1recipient", "1000000", "Test")
	require.ErrorIs(t, err, types.ErrInvalidSigner)
}

func TestProposeTreasurySpend_NoTreasuryAddress(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress: "", // empty
		Signers:         []string{"aura1signer1"},
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	_, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	require.ErrorIs(t, err, types.ErrInvalidTreasuryAddress)
}

func TestProposeTreasurySpend_InvalidAmount(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress: "aura1treasury",
		Signers:         []string{"aura1signer1"},
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	_, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "not-a-number", "Test")
	require.ErrorIs(t, err, types.ErrInvalidAmount)
}

func TestSignTreasurySpend_Success(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2", "aura1signer3"},
		Threshold:        2,
		TimelockDuration: 86400,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create proposal
	txID, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	require.NoError(t, err)

	// Sign with second signer
	signatureCount, threshold, err := k.SignTreasurySpend(ctx, "aura1signer2", txID)
	require.NoError(t, err)
	require.Equal(t, uint32(2), signatureCount) // Proposer + signer2
	require.Equal(t, uint32(2), threshold)
}

func TestSignTreasurySpend_AlreadySigned(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2"},
		Threshold:        2,
		TimelockDuration: 86400,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create proposal
	txID, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	require.NoError(t, err)

	// Try to sign again with same signer (proposer)
	_, _, err = k.SignTreasurySpend(ctx, "aura1signer1", txID)
	require.ErrorIs(t, err, types.ErrAlreadySigned)
}

func TestSignTreasurySpend_TxNotFound(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress: "aura1treasury",
		Signers:         []string{"aura1signer1"},
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	_, _, err := k.SignTreasurySpend(ctx, "aura1signer1", "nonexistent-tx")
	require.Error(t, err)
}

func TestGetAllPendingTreasuryTxs_Empty(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	txs, err := k.GetAllPendingTreasuryTxs(ctx)
	require.NoError(t, err)
	require.Len(t, txs, 0)
}

func TestGetAllPendingTreasuryTxs_WithTxs(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2"},
		Threshold:        2,
		TimelockDuration: 86400,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create a proposal
	txID, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient1", "1000000", "Test 1")
	require.NoError(t, err)
	require.NotEmpty(t, txID)

	// Verify we can retrieve pending txs
	txs, err := k.GetAllPendingTreasuryTxs(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(txs), 1)
}

func TestRejectTreasurySpend_Success(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1", "aura1signer2"},
		Threshold:        2,
		TimelockDuration: 86400,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create proposal
	txID, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	require.NoError(t, err)

	// Reject
	err = k.RejectTreasurySpend(ctx, txID)
	require.NoError(t, err)

	// Verify rejected
	tx, _ := k.GetPendingTreasuryTx(ctx, txID)
	require.True(t, tx.Rejected)
}

func TestGetTreasuryStatistics(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	enabled, treasuryAddress, threshold, signerCount, pendingTxCount, executedTxCount := k.GetTreasuryStatistics(ctx)
	// Default params should have treasury disabled (empty address)
	require.False(t, enabled)
	require.Empty(t, treasuryAddress)
	require.Equal(t, uint32(0), threshold)
	require.Equal(t, uint32(0), signerCount)
	require.Equal(t, uint64(0), pendingTxCount)
	require.Equal(t, uint64(0), executedTxCount)
}

func TestGenerateTreasuryTxID(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	currentTime := time.Now().Unix()

	txID1 := k.generateTreasuryTxID("aura1proposer", "aura1recipient", "1000000", currentTime)
	require.NotEmpty(t, txID1)

	// Same inputs should produce same ID
	txID2 := k.generateTreasuryTxID("aura1proposer", "aura1recipient", "1000000", currentTime)
	require.Equal(t, txID1, txID2)

	// Different inputs should produce different ID
	txID3 := k.generateTreasuryTxID("aura1proposer", "aura1other", "1000000", currentTime)
	require.NotEqual(t, txID1, txID3)
}

func TestGetPendingTreasuryTxByID_NotFound(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	_, found := k.GetPendingTreasuryTxByID(ctx, "nonexistent")
	require.False(t, found)
}

func TestGetPendingTreasuryTxByID_Found(t *testing.T) {
	params := types.DefaultParams()
	params.TreasuryMultisig = &types.TreasuryMultisig{
		TreasuryAddress:  "aura1treasury",
		Signers:          []string{"aura1signer1"},
		Threshold:        1,
		TimelockDuration: 86400,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	txID, _, err := k.ProposeTreasurySpend(ctx, "aura1signer1", "aura1recipient", "1000000", "Test")
	require.NoError(t, err)

	tx, found := k.GetPendingTreasuryTxByID(ctx, txID)
	require.True(t, found)
	require.NotNil(t, tx)
	require.Equal(t, "1000000", tx.Amount)
}
