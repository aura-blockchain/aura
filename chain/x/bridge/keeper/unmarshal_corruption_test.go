// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
	"github.com/stretchr/testify/require"
)

// ========================================================================
// UNMARSHAL ERROR HANDLING TESTS
// ========================================================================

// TestUnmarshalError_CorruptedPendingTransfer tests corrupted pending transfer data handling
func TestUnmarshalError_CorruptedPendingTransfer(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Create a valid pending transfer
	validPending := &types.PendingTransfer{
		TransferId:   "valid-123",
		Recipient:    "recipient-address",
		Amount:       sdkmath.NewInt(1000),
		Denom:        "uaura",
		SourceChain:  "ethereum",
		SourceTxHash: "0xabc123",
		CreatedAt:    input.Ctx.BlockTime(),
		UnlockTime:   input.Ctx.BlockTime().Add(types.DefaultFraudProofWindow),
		Challenged:   false,
	}
	k.SetPendingTransfer(input.Ctx, validPending)

	// Manually write corrupted data
	store := input.Ctx.KVStore(input.StoreKey)
	corruptedKey := types.PendingTransferKey("corrupted-456")
	corruptedData := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF} // Invalid protobuf
	store.Set(corruptedKey, corruptedData)

	// Should NOT panic, should return only valid transfer
	pending := k.GetAllPendingTransfers(input.Ctx)
	require.Len(t, pending, 1)
	require.Equal(t, "valid-123", pending[0].TransferId)
}

// TestUnmarshalError_TruncatedData tests handling of truncated protobuf data
func TestUnmarshalError_TruncatedData(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Create a chain config and marshal it
	validConfig := types.ChainConfig{
		ChainId:          "test-chain",
		ChainName:        "Test",
		Enabled:          true,
		MinConfirmations: 10,
	}
	validBytes := input.Cdc.MustMarshal(&validConfig)

	// Truncate the data
	truncatedBytes := validBytes[:len(validBytes)/2]

	// Write truncated data to storage
	store := input.Ctx.KVStore(input.StoreKey)
	store.Set(types.ChainConfigKey("truncated"), truncatedBytes)

	// Should NOT panic, should return not found
	_, found := k.GetSupportedChain(input.Ctx, "truncated")
	require.False(t, found)
}

// TestUnmarshalError_EmptyData tests handling of empty data
// Note: Empty protobuf data is valid and deserializes to default values
func TestUnmarshalError_EmptyData(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Write empty data
	store := input.Ctx.KVStore(input.StoreKey)
	store.Set(types.ChainConfigKey("empty"), []byte{})

	// Empty protobuf is valid and deserializes to default values
	cfg, found := k.GetSupportedChain(input.Ctx, "empty")
	require.True(t, found)
	require.Empty(t, cfg.ChainId) // ChainId will be empty string (default value)
}

// TestUnmarshalError_WrongTypeData tests data of wrong type in storage
// Note: Protobuf is flexible - compatible field types may partially deserialize
func TestUnmarshalError_WrongTypeData(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Create a chain config and marshal it
	chainConfig := types.ChainConfig{
		ChainId:          "test",
		ChainName:        "Test",
		Enabled:          true,
		MinConfirmations: 10,
	}
	configBytes := input.Cdc.MustMarshal(&chainConfig)

	// Write chain config bytes to a PENDING TRANSFER key (wrong type)
	store := input.Ctx.KVStore(input.StoreKey)
	store.Set(types.PendingTransferKey("wrong-type"), configBytes)

	// Protobuf may partially deserialize compatible types
	// What matters is that it doesn't PANIC
	_, found := k.GetPendingTransfer(input.Ctx, "wrong-type")
	// We don't assert found == false because protobuf might partially deserialize
	// The important thing is no panic occurred
	_ = found // Use the variable
}

// TestUnmarshalError_MixedValidAndCorrupted tests mixed valid/corrupted data
func TestUnmarshalError_MixedValidAndCorrupted(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Create multiple valid pending transfers
	for i := 0; i < 5; i++ {
		pending := &types.PendingTransfer{
			TransferId:   "transfer-" + string(rune('0'+i)),
			Recipient:    "recipient",
			Amount:       sdkmath.NewInt(1000),
			Denom:        "uaura",
			SourceChain:  "ethereum",
			SourceTxHash: "0xabc",
			CreatedAt:    input.Ctx.BlockTime(),
			UnlockTime:   input.Ctx.BlockTime().Add(types.DefaultFraudProofWindow),
			Challenged:   false,
		}
		k.SetPendingTransfer(input.Ctx, pending)
	}

	// Inject corrupted data between valid entries
	store := input.Ctx.KVStore(input.StoreKey)
	for i := 0; i < 3; i++ {
		corruptedKey := types.PendingTransferKey("corrupted-" + string(rune('0'+i)))
		corruptedData := make([]byte, 20+i*10)
		for j := range corruptedData {
			corruptedData[j] = byte((i * j) % 256)
		}
		store.Set(corruptedKey, corruptedData)
	}

	// Should return only the 5 valid transfers, skipping 3 corrupted ones
	pending := k.GetAllPendingTransfers(input.Ctx)
	require.Len(t, pending, 5)

	// Verify all returned transfers are valid
	for _, p := range pending {
		require.NotEmpty(t, p.TransferId)
		require.NotEmpty(t, p.SourceChain)
	}
}

// TestUnmarshalError_VeryLargeCorruptedData tests handling of very large corrupted data
func TestUnmarshalError_VeryLargeCorruptedData(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Write very large corrupted data (10KB)
	store := input.Ctx.KVStore(input.StoreKey)
	largeCorruptedData := make([]byte, 10*1024)
	for i := range largeCorruptedData {
		largeCorruptedData[i] = byte(i % 256)
	}
	store.Set(types.PendingTransferKey("large-corrupted"), largeCorruptedData)

	// Should NOT panic, should handle large corrupted data gracefully
	_, found := k.GetPendingTransfer(input.Ctx, "large-corrupted")
	require.False(t, found)
}

// TestUnmarshalError_GetSingleCorrupted tests getting a single corrupted item
func TestUnmarshalError_GetSingleCorrupted(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Write corrupted pending transfer
	store := input.Ctx.KVStore(input.StoreKey)
	corruptedKey := types.PendingTransferKey("corrupted")
	corruptedData := []byte{0xDE, 0xAD, 0xBE, 0xEF} // Invalid protobuf
	store.Set(corruptedKey, corruptedData)

	// Get should return false for corrupted data
	transfer, found := k.GetPendingTransfer(input.Ctx, "corrupted")
	require.False(t, found)
	require.Nil(t, transfer)
}

// TestUnmarshalError_AllZeros tests all-zero corrupted data
func TestUnmarshalError_AllZeros(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Write all-zero data
	store := input.Ctx.KVStore(input.StoreKey)
	allZeros := make([]byte, 100)
	store.Set(types.ChainConfigKey("all-zeros"), allZeros)

	// Should NOT panic
	_, found := k.GetSupportedChain(input.Ctx, "all-zeros")
	require.False(t, found)
}

// TestUnmarshalError_RandomBytes tests random corrupted bytes
func TestUnmarshalError_RandomBytes(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Write random bytes
	store := input.Ctx.KVStore(input.StoreKey)
	randomBytes := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}
	store.Set(types.ChainConfigKey("random"), randomBytes)

	// Should NOT panic
	_, found := k.GetSupportedChain(input.Ctx, "random")
	require.False(t, found)
}
