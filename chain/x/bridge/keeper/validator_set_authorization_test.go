package keeper_test

import (
	"crypto/sha256"
	"testing"
	"time"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestValidatorSetAuthorization_Issue019 tests the security fixes for issue #019
// Bridge Validator Signature Verification Weakness
//
// This test suite verifies three critical security requirements:
//   1. Validator authorization check against governance-approved list
//   2. Active validator set verification at current block height
//   3. Replay protection with signature tracking
//
// CVSS Score: 9.1 (CRITICAL)
func TestValidatorSetAuthorization_Issue019(t *testing.T) {
	t.Run("01_ActiveValidatorSetRetrieved_Success", testActiveValidatorSetRetrieved)
	t.Run("02_InactiveValidatorExcluded_Success", testInactiveValidatorExcluded)
	t.Run("03_ValidatorDeactivated_SetUpdated", testValidatorDeactivatedSetUpdated)
	t.Run("04_NoActiveValidators_EmptySet", testNoActiveValidatorsEmptySet)
	t.Run("05_SignatureSetReplay_Prevented", testSignatureSetReplayPrevented)
	t.Run("06_SourceHashReplay_Prevented", testSourceHashReplayPrevented)
}

// Test that getActiveValidatorSet returns only active validators
func testActiveValidatorSetRetrieved(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Create three validators: 2 active, 1 inactive
	val1PrivKey, val1PubKey := generateTestKeyPair(t)
	val1PubKeyBytes := val1PubKey.SerializeCompressed()
	var val1CosmosAddr sdk.AccAddress = val1PubKeyBytes[:20]
	val1Addr := val1CosmosAddr.String()

	val2PrivKey, val2PubKey := generateTestKeyPair(t)
	val2PubKeyBytes := val2PubKey.SerializeCompressed()
	var val2CosmosAddr sdk.AccAddress = val2PubKeyBytes[:20]
	val2Addr := val2CosmosAddr.String()

	val3PrivKey, val3PubKey := generateTestKeyPair(t)
	val3PubKeyBytes := val3PubKey.SerializeCompressed()
	var val3CosmosAddr sdk.AccAddress = val3PubKeyBytes[:20]
	val3Addr := val3CosmosAddr.String()

	_ = val1PrivKey
	_ = val2PrivKey
	_ = val3PrivKey

	// Create marshaled pub keys
	val1MarshaledPubKey, err := input.Cdc.MarshalInterface(val1PubKey)
	require.NoError(t, err)
	val2MarshaledPubKey, err := input.Cdc.MarshalInterface(val2PubKey)
	require.NoError(t, err)
	val3MarshaledPubKey, err := input.Cdc.MarshalInterface(val3PubKey)
	require.NoError(t, err)

	k.SetValidator(ctx, &types.BridgeValidator{
		Address:   val1Addr,
		Active:    true, // ACTIVE
		PublicKey: val1MarshaledPubKey,
	})
	k.SetValidator(ctx, &types.BridgeValidator{
		Address:   val2Addr,
		Active:    true, // ACTIVE
		PublicKey: val2MarshaledPubKey,
	})
	k.SetValidator(ctx, &types.BridgeValidator{
		Address:   val3Addr,
		Active:    false, // INACTIVE
		PublicKey: val3MarshaledPubKey,
	})

	// Get active validator set
	activeVals := k.GetActiveValidatorSet(ctx, ctx.BlockHeight())

	// Should return only val1 and val2
	require.Len(t, activeVals, 2, "should return only active validators")
	require.Contains(t, activeVals, val1Addr, "should include val1")
	require.Contains(t, activeVals, val2Addr, "should include val2")
	require.NotContains(t, activeVals, val3Addr, "should NOT include val3 (inactive)")
}

// Test that inactive validators are excluded from active set
func testInactiveValidatorExcluded(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Create validator marked as inactive
	val1PrivKey, val1PubKey := generateTestKeyPair(t)
	val1PubKeyBytes := val1PubKey.SerializeCompressed()
	var val1CosmosAddr sdk.AccAddress = val1PubKeyBytes[:20]
	val1Addr := val1CosmosAddr.String()

	_ = val1PrivKey

	val1MarshaledPubKey, err := input.Cdc.MarshalInterface(val1PubKey)
	require.NoError(t, err)

	k.SetValidator(ctx, &types.BridgeValidator{
		Address:   val1Addr,
		Active:    false, // Inactive (slashed/jailed)
		PublicKey: val1MarshaledPubKey,
	})

	// Get active validator set
	activeVals := k.GetActiveValidatorSet(ctx, ctx.BlockHeight())

	// Should be empty
	require.Len(t, activeVals, 0, "inactive validator should not be in active set")
}

// Test that when a validator is deactivated, it's removed from active set
func testValidatorDeactivatedSetUpdated(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Create validator initially active
	val1PrivKey, val1PubKey := generateTestKeyPair(t)
	val1PubKeyBytes := val1PubKey.SerializeCompressed()
	var val1CosmosAddr sdk.AccAddress = val1PubKeyBytes[:20]
	val1Addr := val1CosmosAddr.String()

	_ = val1PrivKey

	val1MarshaledPubKey, err := input.Cdc.MarshalInterface(val1PubKey)
	require.NoError(t, err)

	k.SetValidator(ctx, &types.BridgeValidator{
		Address:   val1Addr,
		Active:    true, // Initially active
		PublicKey: val1MarshaledPubKey,
	})

	// Get active validator set
	activeVals := k.GetActiveValidatorSet(ctx, ctx.BlockHeight())
	require.Len(t, activeVals, 1, "should have one active validator")

	// SCENARIO: Validator is removed/slashed
	k.SetValidator(ctx, &types.BridgeValidator{
		Address:   val1Addr,
		Active:    false, // Now inactive
		PublicKey: val1MarshaledPubKey,
	})

	// Get active validator set again
	activeValsAfter := k.GetActiveValidatorSet(ctx, ctx.BlockHeight())
	require.Len(t, activeValsAfter, 0, "should have zero active validators after deactivation")
}

// Test that when no validators are active, empty set is returned
func testNoActiveValidatorsEmptySet(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Don't create any validators

	// Get active validator set
	activeVals := k.GetActiveValidatorSet(ctx, ctx.BlockHeight())

	// Should be empty
	require.Len(t, activeVals, 0, "should return empty set when no active validators")
}

// Test that signature sets cannot be replayed
func testSignatureSetReplayPrevented(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Create some dummy signatures
	sig1 := []byte("signature1_dummy_65_bytes_0000000000000000000000000000000000000000000")
	sig2 := []byte("signature2_dummy_65_bytes_0000000000000000000000000000000000000000000")

	transferID := "transfer-replay-001"

	// Compute signature set hash
	sigs := [][]byte{sig1, sig2}
	sigSetHash := k.ComputeSignatureSetHash(sigs)

	// Check if used (should be false initially)
	used := k.IsSignatureSetUsed(ctx, transferID, sigSetHash)
	require.False(t, used, "signature set should not be marked as used initially")

	// Mark as used
	k.MarkSignatureSetUsed(ctx, transferID, sigSetHash)

	// Check again (should be true now)
	usedAfter := k.IsSignatureSetUsed(ctx, transferID, sigSetHash)
	require.True(t, usedAfter, "signature set should be marked as used after marking")

	// Try to mark again (should be idempotent)
	k.MarkSignatureSetUsed(ctx, transferID, sigSetHash)
	stillUsed := k.IsSignatureSetUsed(ctx, transferID, sigSetHash)
	require.True(t, stillUsed, "signature set should remain marked as used")
}

// Test that source hashes cannot be replayed
func testSourceHashReplayPrevented(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	sourceChain := "ethereum"
	sourceHash := "0xabcd1234567890abcdef"

	// Check if processed (should be false initially)
	processed := k.IsSourceHashProcessed(ctx, sourceChain, sourceHash)
	require.False(t, processed, "source hash should not be processed initially")

	// Mark as processed
	k.MarkSourceHashProcessed(ctx, sourceChain, sourceHash)

	// Check again (should be true now)
	processedAfter := k.IsSourceHashProcessed(ctx, sourceChain, sourceHash)
	require.True(t, processedAfter, "source hash should be marked as processed")

	// Try to mark again (should be idempotent)
	k.MarkSourceHashProcessed(ctx, sourceChain, sourceHash)
	stillProcessed := k.IsSourceHashProcessed(ctx, sourceChain, sourceHash)
	require.True(t, stillProcessed, "source hash should remain marked as processed")
}

// Test integration scenario: validator set rotation during fraud proof window
func TestValidatorSetRotation_DuringFraudProofWindow(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Create initial validator set (2 validators)
	val1PrivKey, val1PubKey := generateTestKeyPair(t)
	val1PubKeyBytes := val1PubKey.SerializeCompressed()
	var val1CosmosAddr sdk.AccAddress = val1PubKeyBytes[:20]
	val1Addr := val1CosmosAddr.String()

	val2PrivKey, val2PubKey := generateTestKeyPair(t)
	val2PubKeyBytes := val2PubKey.SerializeCompressed()
	var val2CosmosAddr sdk.AccAddress = val2PubKeyBytes[:20]
	val2Addr := val2CosmosAddr.String()

	_ = val1PrivKey
	_ = val2PrivKey

	val1MarshaledPubKey, err := input.Cdc.MarshalInterface(val1PubKey)
	require.NoError(t, err)
	val2MarshaledPubKey, err := input.Cdc.MarshalInterface(val2PubKey)
	require.NoError(t, err)

	k.SetValidator(ctx, &types.BridgeValidator{
		Address:   val1Addr,
		Active:    true,
		PublicKey: val1MarshaledPubKey,
	})
	k.SetValidator(ctx, &types.BridgeValidator{
		Address:   val2Addr,
		Active:    true,
		PublicKey: val2MarshaledPubKey,
	})

	// Create transfer
	transfer := &types.CrossChainTransfer{
		TransferId:  "transfer-rotation-001",
		SourceChain: "ethereum",
		TargetChain: "aura",
		Sender:      "0x123",
		Recipient:   "aura1recipient",
		Amount:      "1000000",
		Denom:       "ueth",
		Status:      types.TransferStatus_PENDING,
		Timestamp:   timestamppb.New(ctx.BlockTime()),
	}
	k.SetTransfer(ctx, transfer)

	// Validators sign the transfer at time T1
	// (In reality, validators would sign off-chain)

	// SCENARIO: During fraud proof window, val1 is slashed
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(5 * time.Minute))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 50)

	k.SetValidator(ctx, &types.BridgeValidator{
		Address:   val1Addr,
		Active:    false, // Slashed/removed
		PublicKey: val1MarshaledPubKey,
	})

	// Get active validator set NOW (after rotation)
	activeVals := k.GetActiveValidatorSet(ctx, ctx.BlockHeight())

	// Should only have val2
	require.Len(t, activeVals, 1, "should have only one active validator after rotation")
	require.Contains(t, activeVals, val2Addr, "should include val2")
	require.NotContains(t, activeVals, val1Addr, "should NOT include val1 (deactivated)")
}

// Test that IsValidatorActive correctly checks validator status
func TestIsValidatorActive(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Create active validator
	val1PrivKey, val1PubKey := generateTestKeyPair(t)
	val1PubKeyBytes := val1PubKey.SerializeCompressed()
	var val1CosmosAddr sdk.AccAddress = val1PubKeyBytes[:20]
	val1Addr := val1CosmosAddr.String()

	// Create inactive validator
	val2PrivKey, val2PubKey := generateTestKeyPair(t)
	val2PubKeyBytes := val2PubKey.SerializeCompressed()
	var val2CosmosAddr sdk.AccAddress = val2PubKeyBytes[:20]
	val2Addr := val2CosmosAddr.String()

	_ = val1PrivKey
	_ = val2PrivKey

	val1MarshaledPubKey, err := input.Cdc.MarshalInterface(val1PubKey)
	require.NoError(t, err)
	val2MarshaledPubKey, err := input.Cdc.MarshalInterface(val2PubKey)
	require.NoError(t, err)

	k.SetValidator(ctx, &types.BridgeValidator{
		Address:   val1Addr,
		Active:    true,
		PublicKey: val1MarshaledPubKey,
	})
	k.SetValidator(ctx, &types.BridgeValidator{
		Address:   val2Addr,
		Active:    false,
		PublicKey: val2MarshaledPubKey,
	})

	// Test active validator
	isActive1 := k.IsValidatorActive(ctx, val1Addr)
	require.True(t, isActive1, "val1 should be active")

	// Test inactive validator
	isActive2 := k.IsValidatorActive(ctx, val2Addr)
	require.False(t, isActive2, "val2 should NOT be active")

	// Test non-existent validator
	isActive3 := k.IsValidatorActive(ctx, "nonexistent_address")
	require.False(t, isActive3, "non-existent validator should NOT be active")

	// Test empty address
	isActive4 := k.IsValidatorActive(ctx, "")
	require.False(t, isActive4, "empty address should NOT be active")
}

// Test that GetActiveValidatorSet emits proper audit events
func TestGetActiveValidatorSet_EmitsAuditEvents(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Create active validator
	val1PrivKey, val1PubKey := generateTestKeyPair(t)
	val1PubKeyBytes := val1PubKey.SerializeCompressed()
	var val1CosmosAddr sdk.AccAddress = val1PubKeyBytes[:20]
	val1Addr := val1CosmosAddr.String()

	_ = val1PrivKey

	val1MarshaledPubKey, err := input.Cdc.MarshalInterface(val1PubKey)
	require.NoError(t, err)

	k.SetValidator(ctx, &types.BridgeValidator{
		Address:   val1Addr,
		Active:    true,
		PublicKey: val1MarshaledPubKey,
	})

	// Get active validator set (should emit event)
	_ = k.GetActiveValidatorSet(ctx, ctx.BlockHeight())

	// Check that event was emitted
	events := ctx.EventManager().Events()
	foundEvent := false
	for _, event := range events {
		if event.Type == "active_validator_set_retrieved" {
			foundEvent = true
			// Verify attributes
			for _, attr := range event.Attributes {
				if attr.Key == "active_count" {
					require.Equal(t, "1", attr.Value, "active_count should be 1")
				}
			}
		}
	}
	require.True(t, foundEvent, "should emit active_validator_set_retrieved event")
}
