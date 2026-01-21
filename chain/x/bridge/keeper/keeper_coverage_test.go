// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

// TestGetActiveValidatorSet tests the public GetActiveValidatorSet method
func TestGetActiveValidatorSet(t *testing.T) {
	t.Run("returns_only_active_validators", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Register 3 active validators
		activeVal1 := &types.BridgeValidator{
			Address:   "cosmos1active1",
			PublicKey: []byte("pubkey1"),
			Active:    true,
		}
		activeVal2 := &types.BridgeValidator{
			Address:   "cosmos1active2",
			PublicKey: []byte("pubkey2"),
			Active:    true,
		}
		activeVal3 := &types.BridgeValidator{
			Address:   "cosmos1active3",
			PublicKey: []byte("pubkey3"),
			Active:    true,
		}

		// Register 2 inactive validators
		inactiveVal1 := &types.BridgeValidator{
			Address:   "cosmos1inactive1",
			PublicKey: []byte("pubkey4"),
			Active:    false,
		}
		inactiveVal2 := &types.BridgeValidator{
			Address:   "cosmos1inactive2",
			PublicKey: []byte("pubkey5"),
			Active:    false,
		}

		// Store validators
		store := input.Ctx.KVStore(input.StoreKey)
		store.Set(types.ValidatorKey(activeVal1.Address), input.Cdc.MustMarshal(activeVal1))
		store.Set(types.ValidatorKey(activeVal2.Address), input.Cdc.MustMarshal(activeVal2))
		store.Set(types.ValidatorKey(activeVal3.Address), input.Cdc.MustMarshal(activeVal3))
		store.Set(types.ValidatorKey(inactiveVal1.Address), input.Cdc.MustMarshal(inactiveVal1))
		store.Set(types.ValidatorKey(inactiveVal2.Address), input.Cdc.MustMarshal(inactiveVal2))

		// Get active validator set
		activeSet := k.GetActiveValidatorSet(input.Ctx, input.Ctx.BlockHeight())

		// Verify only active validators returned
		require.Len(t, activeSet, 3)
		require.Contains(t, activeSet, "cosmos1active1")
		require.Contains(t, activeSet, "cosmos1active2")
		require.Contains(t, activeSet, "cosmos1active3")
		require.NotContains(t, activeSet, "cosmos1inactive1")
		require.NotContains(t, activeSet, "cosmos1inactive2")
	})

	t.Run("empty_when_no_validators", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		activeSet := k.GetActiveValidatorSet(input.Ctx, input.Ctx.BlockHeight())
		require.Empty(t, activeSet)
	})

	t.Run("empty_when_all_inactive", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Register only inactive validators
		inactiveVal1 := &types.BridgeValidator{
			Address:   "cosmos1inactive1",
			PublicKey: []byte("pubkey1"),
			Active:    false,
		}
		inactiveVal2 := &types.BridgeValidator{
			Address:   "cosmos1inactive2",
			PublicKey: []byte("pubkey2"),
			Active:    false,
		}

		store := input.Ctx.KVStore(input.StoreKey)
		store.Set(types.ValidatorKey(inactiveVal1.Address), input.Cdc.MustMarshal(inactiveVal1))
		store.Set(types.ValidatorKey(inactiveVal2.Address), input.Cdc.MustMarshal(inactiveVal2))

		activeSet := k.GetActiveValidatorSet(input.Ctx, input.Ctx.BlockHeight())
		require.Empty(t, activeSet)
	})

	t.Run("emits_event_with_validator_count", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Register 2 active validators
		activeVal1 := &types.BridgeValidator{
			Address:   "cosmos1active1",
			PublicKey: []byte("pubkey1"),
			Active:    true,
		}
		activeVal2 := &types.BridgeValidator{
			Address:   "cosmos1active2",
			PublicKey: []byte("pubkey2"),
			Active:    true,
		}

		store := input.Ctx.KVStore(input.StoreKey)
		store.Set(types.ValidatorKey(activeVal1.Address), input.Cdc.MustMarshal(activeVal1))
		store.Set(types.ValidatorKey(activeVal2.Address), input.Cdc.MustMarshal(activeVal2))

		// Call function
		_ = k.GetActiveValidatorSet(input.Ctx, input.Ctx.BlockHeight())

		// Verify event was emitted
		events := input.Ctx.EventManager().Events()
		require.NotEmpty(t, events)

		// Find the active_validator_set_retrieved event
		foundEvent := false
		for _, event := range events {
			if event.Type == "active_validator_set_retrieved" {
				foundEvent = true
				// Verify event attributes
				for _, attr := range event.Attributes {
					if attr.Key == "active_count" {
						require.Equal(t, "2", attr.Value)
					}
				}
			}
		}
		require.True(t, foundEvent, "Expected active_validator_set_retrieved event to be emitted")
	})
}

// TestIsValidatorActive tests the public IsValidatorActive method
func TestIsValidatorActive(t *testing.T) {
	t.Run("active_validator_returns_true", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Register active validator
		validator := &types.BridgeValidator{
			Address:   "cosmos1active",
			PublicKey: []byte("pubkey1"),
			Active:    true,
		}

		store := input.Ctx.KVStore(input.StoreKey)
		store.Set(types.ValidatorKey(validator.Address), input.Cdc.MustMarshal(validator))

		// Check if validator is active
		isActive := k.IsValidatorActive(input.Ctx, "cosmos1active")
		require.True(t, isActive)
	})

	t.Run("inactive_validator_returns_false", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Register inactive validator
		validator := &types.BridgeValidator{
			Address:   "cosmos1inactive",
			PublicKey: []byte("pubkey1"),
			Active:    false,
		}

		store := input.Ctx.KVStore(input.StoreKey)
		store.Set(types.ValidatorKey(validator.Address), input.Cdc.MustMarshal(validator))

		// Check if validator is active
		isActive := k.IsValidatorActive(input.Ctx, "cosmos1inactive")
		require.False(t, isActive)
	})

	t.Run("nonexistent_validator_returns_false", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Check non-existent validator
		isActive := k.IsValidatorActive(input.Ctx, "cosmos1nonexistent")
		require.False(t, isActive)
	})

	t.Run("empty_address_returns_false", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Check empty address
		isActive := k.IsValidatorActive(input.Ctx, "")
		require.False(t, isActive)
	})
}

// TestComputeSignatureSetHash tests the public ComputeSignatureSetHash method
func TestComputeSignatureSetHash(t *testing.T) {
	t.Run("empty_signatures_returns_nil", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		hash := k.ComputeSignatureSetHash([][]byte{})
		require.Nil(t, hash)

		hash = k.ComputeSignatureSetHash(nil)
		require.Nil(t, hash)
	})

	t.Run("single_signature_hash", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sig1 := []byte("signature1")
		hash := k.ComputeSignatureSetHash([][]byte{sig1})

		// Compute expected hash manually
		expected := sha256.Sum256(sig1)

		require.NotNil(t, hash)
		require.Equal(t, expected[:], hash)
	})

	t.Run("multiple_signatures_order_independent", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sig1 := []byte("signature1")
		sig2 := []byte("signature2")
		sig3 := []byte("signature3")

		// Compute hash with different orderings
		hash1 := k.ComputeSignatureSetHash([][]byte{sig1, sig2, sig3})
		hash2 := k.ComputeSignatureSetHash([][]byte{sig3, sig1, sig2})
		hash3 := k.ComputeSignatureSetHash([][]byte{sig2, sig3, sig1})

		// All hashes should be identical
		require.NotNil(t, hash1)
		require.Equal(t, hash1, hash2)
		require.Equal(t, hash1, hash3)
	})

	t.Run("duplicate_signatures_handled_consistently", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sig1 := []byte("signature1")

		// Compute hash with duplicates
		hash1 := k.ComputeSignatureSetHash([][]byte{sig1, sig1, sig1})
		hash2 := k.ComputeSignatureSetHash([][]byte{sig1, sig1})

		// Both should produce a valid hash (consistency check)
		require.NotNil(t, hash1)
		require.NotNil(t, hash2)
	})

	t.Run("different_signatures_produce_different_hashes", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sig1 := []byte("signature1")
		sig2 := []byte("signature2")

		hash1 := k.ComputeSignatureSetHash([][]byte{sig1})
		hash2 := k.ComputeSignatureSetHash([][]byte{sig2})

		require.NotNil(t, hash1)
		require.NotNil(t, hash2)
		require.NotEqual(t, hash1, hash2)
	})
}

// TestIsSignatureSetUsed tests the public IsSignatureSetUsed method
func TestIsSignatureSetUsed(t *testing.T) {
	t.Run("new_signature_set_not_used", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		transferID := "transfer-001"
		signatureSetHash := []byte("somehash")

		isUsed := k.IsSignatureSetUsed(input.Ctx, transferID, signatureSetHash)
		require.False(t, isUsed)
	})

	t.Run("marked_signature_set_is_used", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		transferID := "transfer-001"
		signatureSetHash := []byte("somehash")

		// Mark as used
		k.MarkSignatureSetUsed(input.Ctx, transferID, signatureSetHash)

		// Check if used
		isUsed := k.IsSignatureSetUsed(input.Ctx, transferID, signatureSetHash)
		require.True(t, isUsed)
	})

	t.Run("empty_transfer_id_returns_false", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		signatureSetHash := []byte("somehash")

		isUsed := k.IsSignatureSetUsed(input.Ctx, "", signatureSetHash)
		require.False(t, isUsed)
	})

	t.Run("empty_hash_returns_false", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		transferID := "transfer-001"

		isUsed := k.IsSignatureSetUsed(input.Ctx, transferID, []byte{})
		require.False(t, isUsed)

		isUsed = k.IsSignatureSetUsed(input.Ctx, transferID, nil)
		require.False(t, isUsed)
	})
}

// TestMarkSignatureSetUsed tests the public MarkSignatureSetUsed method
func TestMarkSignatureSetUsed(t *testing.T) {
	t.Run("marks_signature_set_as_used", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		transferID := "transfer-001"
		signatureSetHash := []byte("somehash")

		// Verify not used initially
		require.False(t, k.IsSignatureSetUsed(input.Ctx, transferID, signatureSetHash))

		// Mark as used
		k.MarkSignatureSetUsed(input.Ctx, transferID, signatureSetHash)

		// Verify now marked as used
		require.True(t, k.IsSignatureSetUsed(input.Ctx, transferID, signatureSetHash))
	})

	t.Run("different_transfers_independent", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		transferID1 := "transfer-001"
		transferID2 := "transfer-002"
		signatureSetHash := []byte("samehash")

		// Mark for transfer1
		k.MarkSignatureSetUsed(input.Ctx, transferID1, signatureSetHash)

		// Verify transfer1 marked but transfer2 not marked
		require.True(t, k.IsSignatureSetUsed(input.Ctx, transferID1, signatureSetHash))
		require.False(t, k.IsSignatureSetUsed(input.Ctx, transferID2, signatureSetHash))
	})

	t.Run("emits_event_on_mark", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		transferID := "transfer-001"
		signatureSetHash := []byte("somehash")

		// Mark as used
		k.MarkSignatureSetUsed(input.Ctx, transferID, signatureSetHash)

		// Verify event was emitted
		events := input.Ctx.EventManager().Events()
		require.NotEmpty(t, events)

		// Find the signature_set_marked_used event
		foundEvent := false
		for _, event := range events {
			if event.Type == "signature_set_marked_used" {
				foundEvent = true
				// Verify event has transfer_id attribute
				for _, attr := range event.Attributes {
					if attr.Key == "transfer_id" {
						require.Equal(t, transferID, attr.Value)
					}
					if attr.Key == "signature_set_hash" {
						require.Equal(t, hex.EncodeToString(signatureSetHash), attr.Value)
					}
				}
			}
		}
		require.True(t, foundEvent, "Expected signature_set_marked_used event to be emitted")
	})

	t.Run("no_op_on_empty_transfer_id", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		signatureSetHash := []byte("somehash")

		// This should be a no-op
		k.MarkSignatureSetUsed(input.Ctx, "", signatureSetHash)

		// Verify not marked
		require.False(t, k.IsSignatureSetUsed(input.Ctx, "", signatureSetHash))
	})

	t.Run("no_op_on_empty_hash", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		transferID := "transfer-001"

		// This should be a no-op
		k.MarkSignatureSetUsed(input.Ctx, transferID, []byte{})

		// Verify not marked (can't check empty hash)
		require.False(t, k.IsSignatureSetUsed(input.Ctx, transferID, []byte{}))
	})
}

// TestVerifySourceBlock tests the public VerifySourceBlock method
func TestVerifySourceBlock(t *testing.T) {
	t.Run("verifies_valid_block_hash", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"
		blockHeight := uint64(100)
		blockHash := []byte("blockhash123")

		// Store verified block hash
		k.SetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight, blockHash)

		// Verify with correct hash
		isValid := k.VerifySourceBlock(input.Ctx, sourceChain, blockHeight, blockHash)
		require.True(t, isValid)
	})

	t.Run("rejects_wrong_block_hash", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"
		blockHeight := uint64(100)
		correctHash := []byte("correcthash")
		wrongHash := []byte("wronghash")

		// Store verified block hash
		k.SetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight, correctHash)

		// Verify with wrong hash
		isValid := k.VerifySourceBlock(input.Ctx, sourceChain, blockHeight, wrongHash)
		require.False(t, isValid)
	})

	t.Run("rejects_unverified_block", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"
		blockHeight := uint64(100)
		blockHash := []byte("blockhash")

		// Don't store any verified hash

		// Verify should fail
		isValid := k.VerifySourceBlock(input.Ctx, sourceChain, blockHeight, blockHash)
		require.False(t, isValid)
	})

	t.Run("rejects_empty_chain_name", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		blockHeight := uint64(100)
		blockHash := []byte("blockhash")

		isValid := k.VerifySourceBlock(input.Ctx, "", blockHeight, blockHash)
		require.False(t, isValid)
	})

	t.Run("rejects_zero_block_height", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"
		blockHash := []byte("blockhash")

		isValid := k.VerifySourceBlock(input.Ctx, sourceChain, 0, blockHash)
		require.False(t, isValid)
	})

	t.Run("rejects_empty_block_hash", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"
		blockHeight := uint64(100)

		isValid := k.VerifySourceBlock(input.Ctx, sourceChain, blockHeight, []byte{})
		require.False(t, isValid)

		isValid = k.VerifySourceBlock(input.Ctx, sourceChain, blockHeight, nil)
		require.False(t, isValid)
	})

	t.Run("normalizes_chain_name_case", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		blockHeight := uint64(100)
		blockHash := []byte("blockhash")

		// Store with lowercase
		k.SetVerifiedBlockHash(input.Ctx, "ethereum", blockHeight, blockHash)

		// Verify with uppercase
		isValid := k.VerifySourceBlock(input.Ctx, "ETHEREUM", blockHeight, blockHash)
		require.True(t, isValid)

		// Verify with mixed case
		isValid = k.VerifySourceBlock(input.Ctx, "EtHeReUm", blockHeight, blockHash)
		require.True(t, isValid)
	})

	t.Run("rejects_hash_with_different_length", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"
		blockHeight := uint64(100)
		correctHash := []byte("correcthash12345")
		wrongLengthHash := []byte("short")

		// Store verified block hash
		k.SetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight, correctHash)

		// Verify with different length hash
		isValid := k.VerifySourceBlock(input.Ctx, sourceChain, blockHeight, wrongLengthHash)
		require.False(t, isValid)
	})

	t.Run("rejects_hash_with_different_bytes", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"
		blockHeight := uint64(100)
		correctHash := []byte("correcthash12345")
		wrongHash := []byte("wronghash1234567") // Same length, different content

		// Store verified block hash
		k.SetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight, correctHash)

		// Verify with wrong hash (same length)
		isValid := k.VerifySourceBlock(input.Ctx, sourceChain, blockHeight, wrongHash)
		require.False(t, isValid)
	})
}

// TestGetVerifiedBlockHash tests the public GetVerifiedBlockHash method
func TestGetVerifiedBlockHash(t *testing.T) {
	t.Run("retrieves_stored_hash", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"
		blockHeight := uint64(100)
		blockHash := []byte("blockhash123")

		// Store hash
		k.SetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight, blockHash)

		// Retrieve hash
		retrieved := k.GetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight)
		require.NotNil(t, retrieved)
		require.Equal(t, blockHash, retrieved)
	})

	t.Run("returns_nil_for_nonexistent_hash", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"
		blockHeight := uint64(100)

		// Retrieve without storing
		retrieved := k.GetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight)
		require.Nil(t, retrieved)
	})

	t.Run("returns_nil_for_empty_chain", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		blockHeight := uint64(100)

		retrieved := k.GetVerifiedBlockHash(input.Ctx, "", blockHeight)
		require.Nil(t, retrieved)
	})

	t.Run("returns_nil_for_zero_height", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"

		retrieved := k.GetVerifiedBlockHash(input.Ctx, sourceChain, 0)
		require.Nil(t, retrieved)
	})

	t.Run("normalizes_chain_name", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		blockHeight := uint64(100)
		blockHash := []byte("blockhash")

		// Store with lowercase
		k.SetVerifiedBlockHash(input.Ctx, "ethereum", blockHeight, blockHash)

		// Retrieve with uppercase
		retrieved := k.GetVerifiedBlockHash(input.Ctx, "ETHEREUM", blockHeight)
		require.NotNil(t, retrieved)
		require.Equal(t, blockHash, retrieved)
	})
}

// TestSetVerifiedBlockHash tests the public SetVerifiedBlockHash method
func TestSetVerifiedBlockHash(t *testing.T) {
	t.Run("stores_block_hash", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"
		blockHeight := uint64(100)
		blockHash := []byte("blockhash123")

		// Store hash
		k.SetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight, blockHash)

		// Verify stored
		retrieved := k.GetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight)
		require.Equal(t, blockHash, retrieved)
	})

	t.Run("overwrites_existing_hash", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"
		blockHeight := uint64(100)
		oldHash := []byte("oldhash")
		newHash := []byte("newhash")

		// Store old hash
		k.SetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight, oldHash)

		// Overwrite with new hash
		k.SetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight, newHash)

		// Verify new hash stored
		retrieved := k.GetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight)
		require.Equal(t, newHash, retrieved)
	})

	t.Run("emits_event_on_store", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"
		blockHeight := uint64(100)
		blockHash := []byte("blockhash")

		// Store hash
		k.SetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight, blockHash)

		// Verify event was emitted
		events := input.Ctx.EventManager().Events()
		require.NotEmpty(t, events)

		// Find the verified_block_hash_stored event
		foundEvent := false
		for _, event := range events {
			if event.Type == "verified_block_hash_stored" {
				foundEvent = true
				// Verify event attributes
				for _, attr := range event.Attributes {
					if attr.Key == "source_chain" {
						require.Equal(t, sourceChain, attr.Value)
					}
				}
			}
		}
		require.True(t, foundEvent, "Expected verified_block_hash_stored event to be emitted")
	})

	t.Run("no_op_on_empty_chain", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		blockHeight := uint64(100)
		blockHash := []byte("blockhash")

		// Should be no-op
		k.SetVerifiedBlockHash(input.Ctx, "", blockHeight, blockHash)

		// Verify not stored
		retrieved := k.GetVerifiedBlockHash(input.Ctx, "", blockHeight)
		require.Nil(t, retrieved)
	})

	t.Run("no_op_on_zero_height", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"
		blockHash := []byte("blockhash")

		// Should be no-op
		k.SetVerifiedBlockHash(input.Ctx, sourceChain, 0, blockHash)

		// Verify not stored
		retrieved := k.GetVerifiedBlockHash(input.Ctx, sourceChain, 0)
		require.Nil(t, retrieved)
	})

	t.Run("no_op_on_empty_hash", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		sourceChain := "ethereum"
		blockHeight := uint64(100)

		// Should be no-op
		k.SetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight, []byte{})

		// Verify not stored
		retrieved := k.GetVerifiedBlockHash(input.Ctx, sourceChain, blockHeight)
		require.Nil(t, retrieved)
	})
}

// TestResetDailyMint tests the public ResetDailyMint method
func TestResetDailyMint(t *testing.T) {
	t.Run("deletes_old_daily_mint_records", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Current date
		currentDate := input.Ctx.BlockTime().UTC().Format("20060102")

		// Create old date (yesterday)
		yesterday := input.Ctx.BlockTime().AddDate(0, 0, -1).UTC().Format("20060102")

		// Store daily mint records for both dates
		store := input.Ctx.KVStore(input.StoreKey)
		store.Set(types.DailyMintKey(yesterday, "uaura"), []byte("1000"))
		store.Set(types.DailyMintKey(currentDate, "uaura"), []byte("2000"))

		// Reset daily mint
		k.ResetDailyMint(input.Ctx)

		// Verify old record deleted
		oldRecord := store.Get(types.DailyMintKey(yesterday, "uaura"))
		require.Nil(t, oldRecord)

		// Verify current record still exists
		currentRecord := store.Get(types.DailyMintKey(currentDate, "uaura"))
		require.NotNil(t, currentRecord)
	})

	t.Run("handles_empty_store", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Should not panic on empty store
		require.NotPanics(t, func() {
			k.ResetDailyMint(input.Ctx)
		})
	})

	t.Run("preserves_current_date_records", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		currentDate := input.Ctx.BlockTime().UTC().Format("20060102")

		// Store multiple records for current date
		store := input.Ctx.KVStore(input.StoreKey)
		store.Set(types.DailyMintKey(currentDate, "uaura"), []byte("1000"))
		store.Set(types.DailyMintKey(currentDate, "utoken"), []byte("2000"))

		// Reset daily mint
		k.ResetDailyMint(input.Ctx)

		// Verify all current records still exist
		record1 := store.Get(types.DailyMintKey(currentDate, "uaura"))
		require.NotNil(t, record1)

		record2 := store.Get(types.DailyMintKey(currentDate, "utoken"))
		require.NotNil(t, record2)
	})
}

// TestResetHourlyMint tests the public ResetHourlyMint method
func TestResetHourlyMint(t *testing.T) {
	t.Run("deletes_old_hourly_mint_records", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Current datetime (hour precision)
		currentDatetime := input.Ctx.BlockTime().UTC().Format("2006010215")

		// Create old datetime (1 hour ago)
		oldDatetime := input.Ctx.BlockTime().Add(-1 * time.Hour).UTC().Format("2006010215")

		// Store hourly mint records for both datetimes
		store := input.Ctx.KVStore(input.StoreKey)
		store.Set(types.HourlyMintKey(oldDatetime, "uaura"), []byte("500"))
		store.Set(types.HourlyMintKey(currentDatetime, "uaura"), []byte("1000"))

		// Reset hourly mint
		k.ResetHourlyMint(input.Ctx)

		// Verify old record deleted
		oldRecord := store.Get(types.HourlyMintKey(oldDatetime, "uaura"))
		require.Nil(t, oldRecord)

		// Verify current record still exists
		currentRecord := store.Get(types.HourlyMintKey(currentDatetime, "uaura"))
		require.NotNil(t, currentRecord)
	})

	t.Run("handles_empty_store", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Should not panic on empty store
		require.NotPanics(t, func() {
			k.ResetHourlyMint(input.Ctx)
		})
	})

	t.Run("preserves_current_hour_records", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		currentDatetime := input.Ctx.BlockTime().UTC().Format("2006010215")

		// Store multiple records for current hour
		store := input.Ctx.KVStore(input.StoreKey)
		store.Set(types.HourlyMintKey(currentDatetime, "uaura"), []byte("500"))
		store.Set(types.HourlyMintKey(currentDatetime, "utoken"), []byte("1000"))

		// Reset hourly mint
		k.ResetHourlyMint(input.Ctx)

		// Verify all current records still exist
		record1 := store.Get(types.HourlyMintKey(currentDatetime, "uaura"))
		require.NotNil(t, record1)

		record2 := store.Get(types.HourlyMintKey(currentDatetime, "utoken"))
		require.NotNil(t, record2)
	})
}

// TestProcessExpiredPendingTransfers tests the public ProcessExpiredPendingTransfers method
func TestProcessExpiredPendingTransfers(t *testing.T) {
	t.Run("handles_empty_pending_transfers", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Should not panic with no pending transfers
		require.NotPanics(t, func() {
			k.ProcessExpiredPendingTransfers(input.Ctx)
		})
	})

	t.Run("skips_challenged_transfers", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Create challenged pending transfer
		transferID := "transfer-challenged"
		pending := &types.PendingTransfer{
			TransferId:   transferID,
			Recipient:    "cosmos1recipient",
			Amount:       math.NewInt(1000),
			Denom:        "uaura",
			SourceChain:  "ethereum",
			SourceTxHash: "0xhash",
			CreatedAt:    input.Ctx.BlockTime(),
			UnlockTime:   input.Ctx.BlockTime().Add(-1 * time.Hour), // Expired
			Challenged:   true,                                      // Challenged
			FraudProofId: "fraud-001",
		}

		// Store pending transfer
		store := input.Ctx.KVStore(input.StoreKey)
		store.Set(types.PendingTransferKey(transferID), input.Cdc.MustMarshal(pending))

		// Process expired transfers
		k.ProcessExpiredPendingTransfers(input.Ctx)

		// Verify pending transfer still exists (not finalized because challenged)
		storedPending := store.Get(types.PendingTransferKey(transferID))
		require.NotNil(t, storedPending)
	})

	t.Run("processes_multiple_pending_transfers", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Create multiple pending transfers
		for i := 1; i <= 3; i++ {
			transferID := "transfer-" + string(rune('0'+i))
			pending := &types.PendingTransfer{
				TransferId:   transferID,
				Recipient:    "cosmos1recipient",
				Amount:       math.NewInt(1000),
				Denom:        "uaura",
				SourceChain:  "ethereum",
				SourceTxHash: "0xhash" + string(rune('0'+i)),
				CreatedAt:    input.Ctx.BlockTime(),
				UnlockTime:   input.Ctx.BlockTime().Add(1 * time.Hour), // Not expired
				Challenged:   false,
				FraudProofId: "",
			}

			// Store pending transfer
			store := input.Ctx.KVStore(input.StoreKey)
			store.Set(types.PendingTransferKey(transferID), input.Cdc.MustMarshal(pending))
		}

		// Should process without error
		require.NotPanics(t, func() {
			k.ProcessExpiredPendingTransfers(input.Ctx)
		})
	})

	t.Run("handles_corrupted_pending_transfer", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Store corrupted data
		store := input.Ctx.KVStore(input.StoreKey)
		store.Set(types.PendingTransferKey("corrupted"), []byte("invalid data"))

		// Should not panic on corrupted data
		require.NotPanics(t, func() {
			k.ProcessExpiredPendingTransfers(input.Ctx)
		})
	})

	t.Run("skips_unexpired_transfers", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Create unexpired pending transfer (unlock time in future)
		transferID := "transfer-unexpired"
		pending := &types.PendingTransfer{
			TransferId:   transferID,
			Recipient:    "cosmos1recipient",
			Amount:       math.NewInt(1000),
			Denom:        "uaura",
			SourceChain:  "ethereum",
			SourceTxHash: "0xhash",
			CreatedAt:    input.Ctx.BlockTime(),
			UnlockTime:   input.Ctx.BlockTime().Add(24 * time.Hour), // Future
			Challenged:   false,
			FraudProofId: "",
		}

		// Store pending transfer
		store := input.Ctx.KVStore(input.StoreKey)
		store.Set(types.PendingTransferKey(transferID), input.Cdc.MustMarshal(pending))

		// Process expired transfers
		k.ProcessExpiredPendingTransfers(input.Ctx)

		// Verify pending transfer still exists (not finalized)
		storedPending := store.Get(types.PendingTransferKey(transferID))
		require.NotNil(t, storedPending)
	})

	t.Run("emits_summary_event_when_processing", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestKeeper(t, input)

		// Create challenged pending transfer
		transferID := "transfer-challenged"
		pending := &types.PendingTransfer{
			TransferId:   transferID,
			Recipient:    "cosmos1recipient",
			Amount:       math.NewInt(1000),
			Denom:        "uaura",
			SourceChain:  "ethereum",
			SourceTxHash: "0xhash",
			CreatedAt:    input.Ctx.BlockTime(),
			UnlockTime:   input.Ctx.BlockTime().Add(-1 * time.Hour), // Expired
			Challenged:   true,                                      // Challenged
			FraudProofId: "fraud-001",
		}

		// Store pending transfer
		store := input.Ctx.KVStore(input.StoreKey)
		store.Set(types.PendingTransferKey(transferID), input.Cdc.MustMarshal(pending))

		// Process expired transfers
		k.ProcessExpiredPendingTransfers(input.Ctx)

		// Verify event was emitted
		events := input.Ctx.EventManager().Events()
		require.NotEmpty(t, events)

		// Find the pending_transfers_processed event
		foundEvent := false
		for _, event := range events {
			if event.Type == "pending_transfers_processed" {
				foundEvent = true
				// Verify event has challenged count
				for _, attr := range event.Attributes {
					if attr.Key == "challenged" {
						require.Equal(t, "1", attr.Value)
					}
				}
			}
		}
		require.True(t, foundEvent, "Expected pending_transfers_processed event to be emitted")
	})
}

// createTestKeeper is a helper function to create a keeper for testing
func createTestKeeper(t *testing.T, input keepertest.TestInput) *keeper.Keeper {
	t.Helper()

	// Create LegacyAmino codec for paramstore
	legacyAmino := codec.NewLegacyAmino()

	// Create a paramstore with KeyTable
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")

	return keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		&ps,
		nil, // bankKeeper
		nil, // accountKeeper
		nil, // vcKeeper
		nil, // stakingKeeper
	)
}
