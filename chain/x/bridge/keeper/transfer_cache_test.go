// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

func TestTransferCache_BasicOperations(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	// Create a test transfer
	transfer := &types.CrossChainTransfer{
		TransferId:            "test-transfer-1",
		SourceChain:           "bitcoin",
		TargetChain:           "aura",
		Sender:                "sender123",
		Recipient:             "aura1recipient",
		Amount:                sdkmath.NewInt(1000),
		Denom:                 "uaura",
		SourceTxHash:          "hash123",
		Status:                types.TransferStatus_PENDING,
		Timestamp:             time.Now(),
		Confirmations:         0,
		RequiredConfirmations: 6,
	}

	// First call - should be a cache miss, fetches from store
	k.SetTransfer(ctx, transfer)

	// Second call - should be a cache hit
	retrieved, found := k.GetTransfer(ctx, "test-transfer-1")
	require.True(t, found)
	require.Equal(t, transfer.TransferId, retrieved.TransferId)
	require.Equal(t, transfer.Amount.String(), retrieved.Amount.String())
	require.Equal(t, transfer.SourceChain, retrieved.SourceChain)

	// Third call - should also be a cache hit (same ID)
	retrieved2, found := k.GetTransfer(ctx, "test-transfer-1")
	require.True(t, found)
	require.Equal(t, transfer.TransferId, retrieved2.TransferId)
}

func TestTransferCache_Invalidation(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	// Create and store a transfer
	transfer := &types.CrossChainTransfer{
		TransferId:            "test-transfer-2",
		SourceChain:           "bitcoin",
		TargetChain:           "aura",
		Sender:                "sender123",
		Recipient:             "aura1recipient",
		Amount:                sdkmath.NewInt(1000),
		Denom:                 "uaura",
		SourceTxHash:          "hash456",
		Status:                types.TransferStatus_PENDING,
		Timestamp:             time.Now(),
		Confirmations:         0,
		RequiredConfirmations: 6,
	}

	k.SetTransfer(ctx, transfer)

	// Get transfer (cache it)
	retrieved, found := k.GetTransfer(ctx, "test-transfer-2")
	require.True(t, found)
	require.Equal(t, types.TransferStatus_PENDING, retrieved.Status)

	// Update transfer (should invalidate cache)
	transfer.Status = types.TransferStatus_COMPLETED
	k.SetTransfer(ctx, transfer)

	// Get transfer again - should have updated status
	retrieved2, found := k.GetTransfer(ctx, "test-transfer-2")
	require.True(t, found)
	require.Equal(t, types.TransferStatus_COMPLETED, retrieved2.Status)
}

func TestTransferCache_MultipleTransfers(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	// Create multiple transfers
	transferIDs := make([]string, 10)
	for i := 1; i <= 10; i++ {
		transferID := fmt.Sprintf("test-transfer-%d", i)
		transferIDs[i-1] = transferID

		transfer := &types.CrossChainTransfer{
			TransferId:            transferID,
			SourceChain:           "bitcoin",
			TargetChain:           "aura",
			Sender:                "sender123",
			Recipient:             "aura1recipient",
			Amount:                sdkmath.NewInt(int64(i * 100)),
			Denom:                 "uaura",
			SourceTxHash:          fmt.Sprintf("hash-%d", i),
			Status:                types.TransferStatus_PENDING,
			Timestamp:             time.Now(),
			Confirmations:         0,
			RequiredConfirmations: 6,
		}

		k.SetTransfer(ctx, transfer)
	}

	// Retrieve all transfers - first call should cache them
	for i, transferID := range transferIDs {
		retrieved, found := k.GetTransfer(ctx, transferID)
		require.True(t, found, "Transfer %d not found", i)
		require.Equal(t, sdkmath.NewInt(int64((i+1)*100)).String(), retrieved.Amount.String())
	}

	// Second retrieval should hit cache
	for i, transferID := range transferIDs {
		retrieved, found := k.GetTransfer(ctx, transferID)
		require.True(t, found, "Transfer %d not found", i)
		require.Equal(t, sdkmath.NewInt(int64((i+1)*100)).String(), retrieved.Amount.String())
	}
}

func TestTransferCache_NotFound(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	// Try to get non-existent transfer
	retrieved, found := k.GetTransfer(ctx, "nonexistent-transfer")
	require.False(t, found)
	require.Nil(t, retrieved)

	// Try again - should still not be found
	retrieved2, found := k.GetTransfer(ctx, "nonexistent-transfer")
	require.False(t, found)
	require.Nil(t, retrieved2)
}

func TestTransferCache_ClearCache(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	// Create and store a transfer
	transfer := &types.CrossChainTransfer{
		TransferId:            "test-transfer-clear",
		SourceChain:           "bitcoin",
		TargetChain:           "aura",
		Sender:                "sender123",
		Recipient:             "aura1recipient",
		Amount:                sdkmath.NewInt(1000),
		Denom:                 "uaura",
		SourceTxHash:          "hash789",
		Status:                types.TransferStatus_PENDING,
		Timestamp:             time.Now(),
		Confirmations:         0,
		RequiredConfirmations: 6,
	}

	k.SetTransfer(ctx, transfer)

	// Get transfer (cache it)
	retrieved, found := k.GetTransfer(ctx, "test-transfer-clear")
	require.True(t, found)
	require.Equal(t, transfer.TransferId, retrieved.TransferId)

	// Clear cache
	k.ClearTransferCache()

	// Get transfer again - should still work (from store)
	retrieved2, found := k.GetTransfer(ctx, "test-transfer-clear")
	require.True(t, found)
	require.Equal(t, transfer.TransferId, retrieved2.TransferId)
}

func TestTransferCache_Stats(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	// Initial stats
	size, capacity := k.GetCacheStats()
	require.Equal(t, 0, size)
	require.Equal(t, 5000, capacity) // DefaultTransferCacheSize (increased from 1000 to 5000)

	// Add a transfer
	transfer := &types.CrossChainTransfer{
		TransferId:            "test-transfer-stats",
		SourceChain:           "bitcoin",
		TargetChain:           "aura",
		Sender:                "sender123",
		Recipient:             "aura1recipient",
		Amount:                sdkmath.NewInt(1000),
		Denom:                 "uaura",
		SourceTxHash:          "hashstats",
		Status:                types.TransferStatus_PENDING,
		Timestamp:             time.Now(),
		Confirmations:         0,
		RequiredConfirmations: 6,
	}

	k.SetTransfer(ctx, transfer)

	// Get transfer (cache it)
	_, found := k.GetTransfer(ctx, "test-transfer-stats")
	require.True(t, found)

	// Check stats - should have 1 entry now
	size, capacity = k.GetCacheStats()
	require.Equal(t, 1, size)
	require.Equal(t, 5000, capacity)
}
