// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

// TestUnlockTokens_ValidatorAuthorization tests that only active validators can sign unlock operations
func TestUnlockTokens_ValidatorAuthorization(t *testing.T) {
	// This test verifies that:
	// 1. Active validators can sign unlock operations
	// 2. Inactive validators are rejected
	// 3. Non-existent validators are rejected
	// 4. Minimum threshold is enforced

	t.Run("reject_inactive_validator", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)
		msgServer := keeper.NewMsgServerImpl(k)
		ctx := sdk.WrapSDKContext(input.Ctx)

		// Create transfer
		transferID := "transfer-001"
		burnTxHash := "0xabcd1234"
		amount := "1000"
		amountInt, _ := sdkmath.NewIntFromString(amount)
		sender := keepertest.GenTestAddr().String()
		createTestTransferForUnlock(t, input, k, transferID, burnTxHash, amount, types.MinAllowedConfirmations)

		// Register 1 active validator and 1 inactive validator
		activeVal := createTestValidatorWithStatus(t, input, k, true)
		inactiveVal := createTestValidatorWithStatus(t, input, k, false)

		// Create unlock message with message hash - MUST match keeper format exactly:
		// transfer.SourceChain:msg.BurnTxHash:msg.Sender:msg.Amount:msg.Denom
		msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s", "paw", burnTxHash, sender, amount, "uaura")
		msgHash := sha256.Sum256([]byte(msgToSign))

		// Get signatures from both validators
		sig1 := signMessageWithValidatorKey(t, activeVal.PrivKey, msgHash[:])
		sig2 := signMessageWithValidatorKey(t, inactiveVal.PrivKey, msgHash[:])

		msg := &bridgepb.MsgUnlockTokens{
			Sender:              sender,
			BurnTxHash:          burnTxHash,
			Amount:              amountInt,
			Denom:               "uaura",
			SourceChain:         "paw",
			ValidatorSignatures: [][]byte{sig1, sig2},
		}

		// Attempt unlock - should fail because only 1 active signature (need at least 3)
		_, err := msgServer.UnlockTokens(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "insufficient validator signatures")
	})

	t.Run("reject_unregistered_validator", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)
		msgServer := keeper.NewMsgServerImpl(k)
		ctx := sdk.WrapSDKContext(input.Ctx)

		// Create transfer
		transferID := "transfer-002"
		burnTxHash := "0xdef56789"
		amount := "2000"
		amountInt, _ := sdkmath.NewIntFromString(amount)
		sender := keepertest.GenTestAddr().String()
		createTestTransferForUnlock(t, input, k, transferID, burnTxHash, amount, types.MinAllowedConfirmations)

		// Register 1 active validator
		activeVal := createTestValidatorWithStatus(t, input, k, true)

		// Create unregistered validator (not stored in keeper)
		unregisteredPrivKey := secp256k1.GenPrivKey()

		// Create unlock message
		msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s", "paw", burnTxHash, sender, amount, "uaura")
		msgHash := sha256.Sum256([]byte(msgToSign))

		// Get signatures from active and unregistered validators
		sig1 := signMessageWithValidatorKey(t, activeVal.PrivKey, msgHash[:])
		sig2 := signMessageWithValidatorKey(t, unregisteredPrivKey, msgHash[:])

		msg := &bridgepb.MsgUnlockTokens{
			Sender:              sender,
			BurnTxHash:          burnTxHash,
			Amount:              amountInt,
			Denom:               "uaura",
			SourceChain:         "paw",
			ValidatorSignatures: [][]byte{sig1, sig2},
		}

		// Attempt unlock - should fail because only 1 registered active validator
		_, err := msgServer.UnlockTokens(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "insufficient validator signatures")
	})

	t.Run("accept_active_validators_only", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)
		msgServer := keeper.NewMsgServerImpl(k)
		ctx := sdk.WrapSDKContext(input.Ctx)

		// Create transfer
		transferID := "transfer-003"
		burnTxHash := "0x11223344"
		amount := "3000"
		amountInt, _ := sdkmath.NewIntFromString(amount)
		sender := keepertest.GenTestAddr().String()
		createTestTransferForUnlock(t, input, k, transferID, burnTxHash, amount, types.MinAllowedConfirmations)

		// Register 3 active validators (need at least 3 for quorum)
		activeVal1 := createTestValidatorWithStatus(t, input, k, true)
		activeVal2 := createTestValidatorWithStatus(t, input, k, true)
		activeVal3 := createTestValidatorWithStatus(t, input, k, true)

		// Register 2 inactive validators
		inactiveVal1 := createTestValidatorWithStatus(t, input, k, false)
		inactiveVal2 := createTestValidatorWithStatus(t, input, k, false)

		// Create unlock message - MUST match keeper format
		msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s", "paw", burnTxHash, sender, amount, "uaura")
		msgHash := sha256.Sum256([]byte(msgToSign))

		// Get signatures from all validators (active and inactive)
		sig1 := signMessageWithValidatorKey(t, activeVal1.PrivKey, msgHash[:])
		sig2 := signMessageWithValidatorKey(t, activeVal2.PrivKey, msgHash[:])
		sig3 := signMessageWithValidatorKey(t, activeVal3.PrivKey, msgHash[:])
		sig4 := signMessageWithValidatorKey(t, inactiveVal1.PrivKey, msgHash[:])
		sig5 := signMessageWithValidatorKey(t, inactiveVal2.PrivKey, msgHash[:])

		msg := &bridgepb.MsgUnlockTokens{
			Sender:              sender,
			BurnTxHash:          burnTxHash,
			Amount:              amountInt,
			Denom:               "uaura",
			SourceChain:         "paw",
			ValidatorSignatures: [][]byte{sig1, sig2, sig3, sig4, sig5},
		}

		// Attempt unlock - should succeed with 3 active validators
		resp, err := msgServer.UnlockTokens(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.Success)
	})

	t.Run("enforce_minimum_threshold", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)
		msgServer := keeper.NewMsgServerImpl(k)
		ctx := sdk.WrapSDKContext(input.Ctx)

		// Create transfer
		transferID := "transfer-004"
		burnTxHash := "0x55667788"
		amount := "4000"
		amountInt, _ := sdkmath.NewIntFromString(amount)
		sender := keepertest.GenTestAddr().String()
		createTestTransferForUnlock(t, input, k, transferID, burnTxHash, amount, types.MinAllowedConfirmations)

		// Register 5 active validators
		activeVals := make([]TestValidatorUnlock, 5)
		for i := 0; i < 5; i++ {
			activeVals[i] = createTestValidatorWithStatus(t, input, k, true)
		}

		// Create unlock message
		msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s", "paw", burnTxHash, sender, amount, "uaura")
		msgHash := sha256.Sum256([]byte(msgToSign))

		// Provide only 1 signature (below MinAllowedConfirmations = 3)
		sig1 := signMessageWithValidatorKey(t, activeVals[0].PrivKey, msgHash[:])

		msg := &bridgepb.MsgUnlockTokens{
			Sender:              sender,
			BurnTxHash:          burnTxHash,
			Amount:              amountInt,
			Denom:               "uaura",
			SourceChain:         "paw",
			ValidatorSignatures: [][]byte{sig1},
		}

		// Attempt unlock - should fail with insufficient signatures
		_, err := msgServer.UnlockTokens(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "insufficient validator signatures")
	})
}

// TestUnlockTokens_SignatureSetReplay tests signature set replay attack prevention
func TestUnlockTokens_SignatureSetReplay(t *testing.T) {
	// This test verifies that:
	// 1. Same signature set cannot be used twice for same transfer
	// 2. Different signature sets can be used (if source hash allows)
	// 3. Signature set hash is deterministic

	t.Run("prevent_exact_signature_set_replay", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)
		msgServer := keeper.NewMsgServerImpl(k)
		ctx := sdk.WrapSDKContext(input.Ctx)

		// Create transfer
		transferID := "transfer-replay-001"
		burnTxHash := "0xreplay1234"
		amount := "5000"
		amountInt, _ := sdkmath.NewIntFromString(amount)
		sender := keepertest.GenTestAddr().String()
		createTestTransferForUnlock(t, input, k, transferID, burnTxHash, amount, types.MinAllowedConfirmations)

		// Register 3 active validators
		validators := make([]TestValidatorUnlock, 3)
		for i := 0; i < 3; i++ {
			validators[i] = createTestValidatorWithStatus(t, input, k, true)
		}

		// Create unlock message - MUST match keeper format
		msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s", "paw", burnTxHash, sender, amount, "uaura")
		msgHash := sha256.Sum256([]byte(msgToSign))

		// Get signatures from validators
		sigs := make([][]byte, 3)
		for i := 0; i < 3; i++ {
			sigs[i] = signMessageWithValidatorKey(t, validators[i].PrivKey, msgHash[:])
		}

		msg := &bridgepb.MsgUnlockTokens{
			Sender:              sender,
			BurnTxHash:          burnTxHash,
			Amount:              amountInt,
			Denom:               "uaura",
			SourceChain:         "paw",
			ValidatorSignatures: sigs,
		}

		// First unlock - should succeed
		resp, err := msgServer.UnlockTokens(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.Success)

		// Second unlock with SAME signatures - should fail (source hash replay)
		_, err = msgServer.UnlockTokens(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already processed")
	})

	t.Run("signature_set_hash_deterministic", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)

		// Create signature set
		sig1 := []byte("signature_aaa")
		sig2 := []byte("signature_bbb")
		sig3 := []byte("signature_ccc")

		// Compute hash with different orderings
		hash1 := k.ComputeSignatureSetHash([][]byte{sig1, sig2, sig3})
		hash2 := k.ComputeSignatureSetHash([][]byte{sig3, sig1, sig2})
		hash3 := k.ComputeSignatureSetHash([][]byte{sig2, sig3, sig1})

		// All hashes should be identical (order-independent)
		require.NotNil(t, hash1)
		require.Equal(t, hash1, hash2, "Hash should be order-independent")
		require.Equal(t, hash1, hash3, "Hash should be order-independent")
	})

	t.Run("different_signature_sets_allowed_if_source_hash_unique", func(t *testing.T) {
		// NOTE: This test documents defense-in-depth behavior
		// Source hash replay protection is the primary defense
		// Signature set tracking provides additional protection

		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)

		// Create two different transfers with different burn hashes
		transferID1 := "transfer-001"
		transferID2 := "transfer-002"

		// Even if we could somehow use the same signature set,
		// the source hash check would prevent replay
		// This test verifies signature set tracking is per-transfer

		signatureSetHash := []byte("samehash")

		// Mark signature set as used for transfer 1
		k.MarkSignatureSetUsed(input.Ctx, transferID1, signatureSetHash)

		// Verify transfer 1 marked
		require.True(t, k.IsSignatureSetUsed(input.Ctx, transferID1, signatureSetHash))

		// Verify transfer 2 NOT marked (independent tracking)
		require.False(t, k.IsSignatureSetUsed(input.Ctx, transferID2, signatureSetHash))
	})
}

// TestUnlockTokens_ValidatorRotation tests validator set changes during fraud proof window
func TestUnlockTokens_ValidatorRotation(t *testing.T) {
	// This test verifies that:
	// 1. Unlock uses validators active at unlock time, not lock time
	// 2. Rotated validators cannot sign after being deactivated
	// 3. New validators can sign immediately after activation

	t.Run("use_current_active_validators", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)
		msgServer := keeper.NewMsgServerImpl(k)
		ctx := sdk.WrapSDKContext(input.Ctx)

		// Create transfer
		transferID := "transfer-rotation-001"
		burnTxHash := "0xrot1234"
		amount := "6000"
		amountInt, _ := sdkmath.NewIntFromString(amount)
		sender := keepertest.GenTestAddr().String()
		createTestTransferForUnlock(t, input, k, transferID, burnTxHash, amount, types.MinAllowedConfirmations)

		// Register validator set A (active initially)
		validatorSetA := make([]TestValidatorUnlock, 2)
		for i := 0; i < 2; i++ {
			validatorSetA[i] = createTestValidatorWithStatus(t, input, k, true)
		}

		// Deactivate set A and activate new set B (need at least 3 active validators)
		for i := 0; i < 2; i++ {
			deactivateValidator(t, input, k, validatorSetA[i].Address)
		}

		validatorSetB := make([]TestValidatorUnlock, 3) // Need 3 for quorum
		for i := 0; i < 3; i++ {
			validatorSetB[i] = createTestValidatorWithStatus(t, input, k, true)
		}

		// Unlock with signatures from set B (current active set)
		msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s", "paw", burnTxHash, sender, amount, "uaura")
		msgHash := sha256.Sum256([]byte(msgToSign))

		sigs := make([][]byte, 3)
		for i := 0; i < 3; i++ {
			sigs[i] = signMessageWithValidatorKey(t, validatorSetB[i].PrivKey, msgHash[:])
		}

		msg := &bridgepb.MsgUnlockTokens{
			Sender:              sender,
			BurnTxHash:          burnTxHash,
			Amount:              amountInt,
			Denom:               "uaura",
			SourceChain:         "paw",
			ValidatorSignatures: sigs,
		}

		// Should succeed with current active set B
		resp, err := msgServer.UnlockTokens(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.Success)
	})

	t.Run("reject_rotated_out_validators", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)
		msgServer := keeper.NewMsgServerImpl(k)
		ctx := sdk.WrapSDKContext(input.Ctx)

		// Create transfer
		transferID := "transfer-rotation-002"
		burnTxHash := "0xrot5678"
		amount := "7000"
		amountInt, _ := sdkmath.NewIntFromString(amount)
		sender := keepertest.GenTestAddr().String()
		createTestTransferForUnlock(t, input, k, transferID, burnTxHash, amount, types.MinAllowedConfirmations)

		// Register validator set A (active initially)
		validatorSetA := make([]TestValidatorUnlock, 2)
		for i := 0; i < 2; i++ {
			validatorSetA[i] = createTestValidatorWithStatus(t, input, k, true)
		}

		// Deactivate set A (rotation)
		for i := 0; i < 2; i++ {
			deactivateValidator(t, input, k, validatorSetA[i].Address)
		}

		// Register new active validators (set B)
		_ = createTestValidatorWithStatus(t, input, k, true)
		_ = createTestValidatorWithStatus(t, input, k, true)

		// Try to unlock with signatures from deactivated set A
		msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s", "paw", burnTxHash, sender, amount, "uaura")
		msgHash := sha256.Sum256([]byte(msgToSign))

		sigs := make([][]byte, 2)
		for i := 0; i < 2; i++ {
			sigs[i] = signMessageWithValidatorKey(t, validatorSetA[i].PrivKey, msgHash[:])
		}

		msg := &bridgepb.MsgUnlockTokens{
			Sender:              sender,
			BurnTxHash:          burnTxHash,
			Amount:              amountInt,
			Denom:               "uaura",
			SourceChain:         "paw",
			ValidatorSignatures: sigs,
		}

		// Should fail - set A validators are inactive
		_, err := msgServer.UnlockTokens(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "insufficient validator signatures")
	})

	t.Run("accept_newly_rotated_in_validators", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)
		msgServer := keeper.NewMsgServerImpl(k)
		ctx := sdk.WrapSDKContext(input.Ctx)

		// Create transfer
		transferID := "transfer-rotation-003"
		burnTxHash := "0xrot9999"
		amount := "8000"
		amountInt, _ := sdkmath.NewIntFromString(amount)
		sender := keepertest.GenTestAddr().String()
		createTestTransferForUnlock(t, input, k, transferID, burnTxHash, amount, types.MinAllowedConfirmations)

		// Register initial validator set (1 validator)
		initialVal := createTestValidatorWithStatus(t, input, k, true)

		// Add new validators (rotation) - need total of 3 active
		newValidator1 := createTestValidatorWithStatus(t, input, k, true)
		newValidator2 := createTestValidatorWithStatus(t, input, k, true)

		// Unlock with signatures including new validators
		msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s", "paw", burnTxHash, sender, amount, "uaura")
		msgHash := sha256.Sum256([]byte(msgToSign))

		// All 3 active validators sign
		sigs := make([][]byte, 3)
		sigs[0] = signMessageWithValidatorKey(t, initialVal.PrivKey, msgHash[:])
		sigs[1] = signMessageWithValidatorKey(t, newValidator1.PrivKey, msgHash[:])
		sigs[2] = signMessageWithValidatorKey(t, newValidator2.PrivKey, msgHash[:])

		msg := &bridgepb.MsgUnlockTokens{
			Sender:              sender,
			BurnTxHash:          burnTxHash,
			Amount:              amountInt,
			Denom:               "uaura",
			SourceChain:         "paw",
			ValidatorSignatures: sigs,
		}

		// Should succeed - new validators can sign immediately
		resp, err := msgServer.UnlockTokens(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.Success)
	})
}

// TestUnlockTokens_CombinedSecurityChecks tests multiple security features together
func TestUnlockTokens_CombinedSecurityChecks(t *testing.T) {
	// This test verifies all security checks work together:
	// 1. Source hash replay protection
	// 2. Signature set replay protection
	// 3. Validator authorization
	// 4. Signature verification
	// 5. Minimum threshold enforcement

	t.Run("all_security_checks_enforced", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)
		msgServer := keeper.NewMsgServerImpl(k)
		ctx := sdk.WrapSDKContext(input.Ctx)

		// Create transfer
		transferID := "transfer-combined-001"
		burnTxHash := "0xcombined1234"
		amount := "9000"
		amountInt, _ := sdkmath.NewIntFromString(amount)
		sender := keepertest.GenTestAddr().String()
		createTestTransferForUnlock(t, input, k, transferID, burnTxHash, amount, types.MinAllowedConfirmations)

		// Register 3 active validators
		validators := make([]TestValidatorUnlock, 3)
		for i := 0; i < 3; i++ {
			validators[i] = createTestValidatorWithStatus(t, input, k, true)
		}

		// Create unlock message - MUST match keeper format
		msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s", "paw", burnTxHash, sender, amount, "uaura")
		msgHash := sha256.Sum256([]byte(msgToSign))

		// Get signatures from all 3 validators
		sigs := make([][]byte, 3)
		for i := 0; i < 3; i++ {
			sigs[i] = signMessageWithValidatorKey(t, validators[i].PrivKey, msgHash[:])
		}

		msg := &bridgepb.MsgUnlockTokens{
			Sender:              sender,
			BurnTxHash:          burnTxHash,
			Amount:              amountInt,
			Denom:               "uaura",
			SourceChain:         "paw",
			ValidatorSignatures: sigs,
		}

		// First unlock - should succeed (all checks pass)
		resp, err := msgServer.UnlockTokens(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.Success)

		// Verify source hash marked as processed
		require.True(t, k.IsSourceHashProcessed(input.Ctx, "paw", burnTxHash))

		// Second unlock - should fail (source hash replay)
		_, err = msgServer.UnlockTokens(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already processed")
	})

	t.Run("attack_scenario_invalid_validator_with_valid_signature", func(t *testing.T) {
		// ATTACK: Attacker has valid cryptographic signature but validator is inactive
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)
		msgServer := keeper.NewMsgServerImpl(k)
		ctx := sdk.WrapSDKContext(input.Ctx)

		// Create transfer
		transferID := "transfer-attack-001"
		burnTxHash := "0xattack1234"
		amount := "10000"
		amountInt, _ := sdkmath.NewIntFromString(amount)
		sender := keepertest.GenTestAddr().String()
		createTestTransferForUnlock(t, input, k, transferID, burnTxHash, amount, types.MinAllowedConfirmations)

		// Register 2 validators - 1 active, 1 inactive
		activeVal := createTestValidatorWithStatus(t, input, k, true)
		inactiveVal := createTestValidatorWithStatus(t, input, k, false)

		// Create unlock message - MUST match keeper format
		msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s", "paw", burnTxHash, sender, amount, "uaura")
		msgHash := sha256.Sum256([]byte(msgToSign))

		// ATTACK: Use valid signatures, but one is from inactive validator
		sig1 := signMessageWithValidatorKey(t, activeVal.PrivKey, msgHash[:])
		sig2 := signMessageWithValidatorKey(t, inactiveVal.PrivKey, msgHash[:])

		msg := &bridgepb.MsgUnlockTokens{
			Sender:              sender,
			BurnTxHash:          burnTxHash,
			Amount:              amountInt,
			Denom:               "uaura",
			SourceChain:         "paw",
			ValidatorSignatures: [][]byte{sig1, sig2},
		}

		// DEFENSE: Should fail - inactive validator not counted
		_, err := msgServer.UnlockTokens(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "insufficient validator signatures")
	})

	t.Run("attack_scenario_reuse_valid_signatures", func(t *testing.T) {
		// ATTACK: Attacker reuses previously valid signature set
		// NOTE: This is caught by source hash replay protection first
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)
		msgServer := keeper.NewMsgServerImpl(k)
		ctx := sdk.WrapSDKContext(input.Ctx)

		// Create transfer
		transferID := "transfer-attack-002"
		burnTxHash := "0xattack5678"
		amount := "11000"
		amountInt, _ := sdkmath.NewIntFromString(amount)
		sender := keepertest.GenTestAddr().String()
		createTestTransferForUnlock(t, input, k, transferID, burnTxHash, amount, types.MinAllowedConfirmations)

		// Register 3 active validators
		validators := make([]TestValidatorUnlock, 3)
		for i := 0; i < 3; i++ {
			validators[i] = createTestValidatorWithStatus(t, input, k, true)
		}

		// Create unlock message - MUST match keeper format
		msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s", "paw", burnTxHash, sender, amount, "uaura")
		msgHash := sha256.Sum256([]byte(msgToSign))

		// Get valid signatures
		sigs := make([][]byte, 3)
		for i := 0; i < 3; i++ {
			sigs[i] = signMessageWithValidatorKey(t, validators[i].PrivKey, msgHash[:])
		}

		msg := &bridgepb.MsgUnlockTokens{
			Sender:              sender,
			BurnTxHash:          burnTxHash,
			Amount:              amountInt,
			Denom:               "uaura",
			SourceChain:         "paw",
			ValidatorSignatures: sigs,
		}

		// First unlock - succeeds
		resp, err := msgServer.UnlockTokens(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// ATTACK: Reuse same signatures
		// DEFENSE: Should fail - source hash already processed
		_, err = msgServer.UnlockTokens(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already processed")
	})

	t.Run("attack_scenario_reuse_burn_transaction", func(t *testing.T) {
		// ATTACK: Attacker tries to unlock same burn transaction multiple times
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)
		msgServer := keeper.NewMsgServerImpl(k)
		ctx := sdk.WrapSDKContext(input.Ctx)

		// Create transfer
		transferID := "transfer-attack-003"
		burnTxHash := "0xattack9999"
		amount := "12000"
		amountInt, _ := sdkmath.NewIntFromString(amount)
		sender := keepertest.GenTestAddr().String()
		createTestTransferForUnlock(t, input, k, transferID, burnTxHash, amount, types.MinAllowedConfirmations)

		// Register 3 active validators
		validators := make([]TestValidatorUnlock, 3)
		for i := 0; i < 3; i++ {
			validators[i] = createTestValidatorWithStatus(t, input, k, true)
		}

		// Create first unlock message - MUST match keeper format
		msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s", "paw", burnTxHash, sender, amount, "uaura")
		msgHash := sha256.Sum256([]byte(msgToSign))

		sigs := make([][]byte, 3)
		for i := 0; i < 3; i++ {
			sigs[i] = signMessageWithValidatorKey(t, validators[i].PrivKey, msgHash[:])
		}

		msg := &bridgepb.MsgUnlockTokens{
			Sender:              sender,
			BurnTxHash:          burnTxHash,
			Amount:              amountInt,
			Denom:               "uaura",
			SourceChain:         "paw",
			ValidatorSignatures: sigs,
		}

		// First unlock - succeeds
		resp, err := msgServer.UnlockTokens(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// ATTACK: Try to unlock same burn transaction again
		// DEFENSE: Source hash replay protection blocks this
		_, err = msgServer.UnlockTokens(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already processed")
	})
}

// NOTE: TestComputeSignatureSetHash and TestIsValidatorActive are now implemented
// in keeper_coverage_test.go to provide full test coverage for these functions.

// TestGetActiveValidators tests the active validator list retrieval
func TestGetActiveValidators(t *testing.T) {
	t.Run("returns_only_active_validators", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)

		// Register 3 active and 2 inactive validators
		activeVals := make([]TestValidatorUnlock, 3)
		for i := 0; i < 3; i++ {
			activeVals[i] = createTestValidatorWithStatus(t, input, k, true)
		}

		inactiveVals := make([]TestValidatorUnlock, 2)
		for i := 0; i < 2; i++ {
			inactiveVals[i] = createTestValidatorWithStatus(t, input, k, false)
		}

		// Get active validator set
		activeSet := k.GetActiveValidatorSet(input.Ctx, input.Ctx.BlockHeight())

		// Verify only active validators returned
		require.Len(t, activeSet, 3)
		for i := 0; i < 3; i++ {
			require.Contains(t, activeSet, activeVals[i].Address)
		}
		for i := 0; i < 2; i++ {
			require.NotContains(t, activeSet, inactiveVals[i].Address)
		}
	})

	t.Run("empty_when_no_validators", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)

		// Get active validators without registering any
		activeSet := k.GetActiveValidatorSet(input.Ctx, input.Ctx.BlockHeight())

		// Expected: Returns empty list
		require.Empty(t, activeSet)
	})

	t.Run("empty_when_all_inactive", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)

		// Register only inactive validators
		for i := 0; i < 3; i++ {
			_ = createTestValidatorWithStatus(t, input, k, false)
		}

		// Get active validators
		activeSet := k.GetActiveValidatorSet(input.Ctx, input.Ctx.BlockHeight())

		// Expected: Returns empty list
		require.Empty(t, activeSet)
	})
}

// TestSignatureSetTracking tests the signature set usage tracking
func TestSignatureSetTracking(t *testing.T) {
	t.Run("new_signature_set_not_used", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)

		transferID := "transfer-tracking-001"
		signatureSetHash := []byte("newsighash")

		// Expected: isSignatureSetUsed returns false for new signature set
		isUsed := k.IsSignatureSetUsed(input.Ctx, transferID, signatureSetHash)
		require.False(t, isUsed, "New signature set should not be marked as used")
	})

	t.Run("marked_signature_set_is_used", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)

		transferID := "transfer-tracking-002"
		signatureSetHash := []byte("markedhash")

		// Mark signature set as used
		k.MarkSignatureSetUsed(input.Ctx, transferID, signatureSetHash)

		// Expected: isSignatureSetUsed returns true
		isUsed := k.IsSignatureSetUsed(input.Ctx, transferID, signatureSetHash)
		require.True(t, isUsed, "Marked signature set should be detected as used")
	})

	t.Run("different_transfers_independent", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)

		transferID1 := "transfer-tracking-003"
		transferID2 := "transfer-tracking-004"
		signatureSetHash := []byte("samehash")

		// Mark signature set for transfer1
		k.MarkSignatureSetUsed(input.Ctx, transferID1, signatureSetHash)

		// Verify transfer1 marked
		require.True(t, k.IsSignatureSetUsed(input.Ctx, transferID1, signatureSetHash))

		// Expected: Same signature set for transfer2 should NOT be marked
		require.False(t, k.IsSignatureSetUsed(input.Ctx, transferID2, signatureSetHash),
			"Signature set tracking should be independent per transfer")
	})

	t.Run("emit_event_on_mark", func(t *testing.T) {
		input := keepertest.CreateTestInput(t)
		k := createTestBridgeKeeper(t, &input)

		transferID := "transfer-tracking-005"
		signatureSetHash := []byte("eventhash")

		// Mark signature set as used
		k.MarkSignatureSetUsed(input.Ctx, transferID, signatureSetHash)

		// Expected: markSignatureSetUsed emits audit event
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
				}
			}
		}
		require.True(t, foundEvent, "Expected signature_set_marked_used event to be emitted")
	})
}

// Benchmark tests for performance
func BenchmarkComputeSignatureSetHash(b *testing.B) {
	input := keepertest.CreateTestInput(&testing.T{})
	k := createTestBridgeKeeper(&testing.T{}, &input)

	signatures := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		signatures[i] = make([]byte, 64) // Typical signature size
		for j := 0; j < 64; j++ {
			signatures[i][j] = byte(i + j)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = k.ComputeSignatureSetHash(signatures)
	}
}

func BenchmarkVerifyValidatorSignatures(b *testing.B) {
	input := keepertest.CreateTestInput(&testing.T{})
	k := createTestBridgeKeeper(&testing.T{}, &input)
	msgServer := keeper.NewMsgServerImpl(k)
	ctx := sdk.WrapSDKContext(input.Ctx)

	// Setup: Create validators
	validators := make([]TestValidatorUnlock, 5)
	for i := 0; i < 5; i++ {
		validators[i] = createTestValidatorWithStatus(&testing.T{}, input, k, true)
	}

	// Create message to sign
	msgToSign := "test:0xbench1234:sender:1000:uaura"
	msgHash := sha256.Sum256([]byte(msgToSign))

	// Get signatures
	sigs := make([][]byte, 5)
	for i := 0; i < 5; i++ {
		sigs[i] = signMessageWithValidatorKey(&testing.T{}, validators[i].PrivKey, msgHash[:])
	}

	// Create unlock message
	msg := &bridgepb.MsgUnlockTokens{
		Sender:              keepertest.GenTestAddr().String(),
		BurnTxHash:          fmt.Sprintf("0xbench%d", b.N),
		Amount:              sdkmath.NewInt(1000),
		Denom:               "uaura",
		SourceChain:         "paw",
		ValidatorSignatures: sigs,
	}

	// Create transfer for each benchmark iteration
	for i := 0; i < b.N; i++ {
		transferID := fmt.Sprintf("bench-transfer-%d", i)
		burnTxHash := fmt.Sprintf("0xbench%d", i)
		createTestTransferForUnlock(&testing.T{}, input, k, transferID, burnTxHash, "1000", types.MinAllowedConfirmations)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg.BurnTxHash = fmt.Sprintf("0xbench%d", i)
		_, _ = msgServer.UnlockTokens(ctx, msg)
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// TestValidatorUnlock represents a test validator with keys for unlock tests
type TestValidatorUnlock struct {
	Address string
	PrivKey cryptotypes.PrivKey
	PubKey  cryptotypes.PubKey
	Active  bool
}

// createTestBridgeKeeper creates a keeper for testing and returns updated input
func createTestBridgeKeeper(t *testing.T, input *keepertest.TestInput) *keeper.Keeper {
	t.Helper()

	// Register crypto interfaces so secp256k1 keys can be marshaled
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Update input codec to use the one with registered crypto interfaces
	input.Cdc = cdc

	// Create LegacyAmino codec for paramstore
	legacyAmino := codec.NewLegacyAmino()

	// Create a paramstore with KeyTable
	ps := paramtypes.NewSubspace(cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge").
		WithKeyTable(types.ParamKeyTable())

	// Set default params
	params := types.DefaultParams()
	ps.SetParamSet(input.Ctx, &params)

	return keeper.NewKeeper(
		cdc,
		input.StoreKey,
		&ps,
		nil, // bankKeeper
		nil, // accountKeeper
		nil, // vcKeeper
		nil, // stakingKeeper
	)
}

// createTestValidatorWithStatus creates a test validator with specified active status
func createTestValidatorWithStatus(t *testing.T, input keepertest.TestInput, k *keeper.Keeper, active bool) TestValidatorUnlock {
	t.Helper()

	// Generate Cosmos SDK secp256k1 key pair
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()

	// Derive address from public key
	addr := sdk.AccAddress(pubKey.Address()).String()

	// Marshal public key for storage
	pubKeyAny, err := input.Cdc.MarshalInterface(pubKey)
	require.NoError(t, err, "Failed to marshal public key")

	// Create bridge validator
	bridgeValidator := &types.BridgeValidator{
		Address:   addr,
		PublicKey: pubKeyAny,
		Power:     1000,
		Active:    active,
		Chains:    []string{"aura", "paw", "xai"},
	}

	// Register validator
	k.SetValidator(input.Ctx, bridgeValidator)

	return TestValidatorUnlock{
		Address: addr,
		PrivKey: privKey,
		PubKey:  pubKey,
		Active:  active,
	}
}

// deactivateValidator marks a validator as inactive
func deactivateValidator(t *testing.T, input keepertest.TestInput, k *keeper.Keeper, address string) {
	t.Helper()

	// Retrieve validator directly from store
	store := input.Ctx.KVStore(input.StoreKey)
	bz := store.Get(types.ValidatorKey(address))
	require.NotNil(t, bz, "Validator should exist")

	var validator types.BridgeValidator
	err := input.Cdc.Unmarshal(bz, &validator)
	require.NoError(t, err)

	validator.Active = false
	k.SetValidator(input.Ctx, &validator)
}

// createTestTransferForUnlock creates a transfer and indexes it for unlock testing
func createTestTransferForUnlock(t *testing.T, input keepertest.TestInput, k *keeper.Keeper, transferID string, burnTxHash string, amount string, requiredConfirmations uint64) {
	t.Helper()

	amountInt, _ := sdkmath.NewIntFromString(amount)
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:            transferID,
		SourceChain:           "paw",
		TargetChain:           "aura",
		Sender:                keepertest.GenTestAddr().String(),
		Recipient:             keepertest.GenTestAddr().String(),
		Denom:                 "uaura",
		Amount:                amountInt,
		Status:                bridgepb.TransferStatus_PENDING,
		Timestamp:             input.Ctx.BlockTime(),
		RequiredConfirmations: requiredConfirmations,
	}

	// Store the transfer
	k.SetTransfer(input.Ctx, transfer)

	// Index the transfer by hash so UnlockTokens can find it
	k.IndexTransferHash(input.Ctx, burnTxHash, transferID)
}

// signMessageWithValidatorKey signs a message hash with a validator's private key
func signMessageWithValidatorKey(t *testing.T, privKey cryptotypes.PrivKey, msgHash []byte) []byte {
	t.Helper()

	sig, err := privKey.Sign(msgHash)
	require.NoError(t, err, "Failed to sign message")

	return sig
}
