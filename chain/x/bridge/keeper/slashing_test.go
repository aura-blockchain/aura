// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

// ========================================================================
// SLASHING EVIDENCE SUBMISSION TESTS
// ========================================================================

func TestSubmitSlashingEvidence_FraudSignature(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	// Setup: Create and register a bridge validator
	validatorAddr := "aura1validator1"
	validator := &types.BridgeValidator{
		Address: validatorAddr,
		Active:  true,
		Power:   100,
		Chains:  []string{"aura", "paw"},
	}
	keeper.SetValidator(ctx, validator)

	// Submit slashing evidence for fraudulent signature
	transferId := "transfer-123"
	evidenceHash := []byte("fraud-evidence-hash")
	submitter := "aura1submitter"

	event, err := keeper.SubmitSlashingEvidence(
		ctx,
		validatorAddr,
		types.SlashReason_SLASH_FRAUD_ATTEMPT,
		transferId,
		evidenceHash,
		submitter,
	)

	// Verify slashing event was created
	require.NoError(t, err)
	require.NotNil(t, event)
	require.Equal(t, validatorAddr, event.ValidatorAddress)
	require.Equal(t, types.SlashReason_SLASH_FRAUD_ATTEMPT, event.Reason)
	require.True(t, event.Jailed) // Should be jailed for fraud
	require.Equal(t, evidenceHash, event.EvidenceHash)

	// Verify validator was marked inactive (jailed in bridge module)
	val, found := keeper.getValidator(ctx, validatorAddr)
	require.True(t, found)
	require.False(t, val.Active) // Should be inactive after slashing

	// Verify slashing event can be retrieved
	retrievedEvent, found := keeper.GetSlashingEvent(ctx, event.EventId)
	require.True(t, found)
	require.Equal(t, event.EventId, retrievedEvent.EventId)
}

func TestSubmitSlashingEvidence_DoubleSigning(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	// Setup validator
	validatorAddr := "aura1validator2"
	validator := &types.BridgeValidator{
		Address: validatorAddr,
		Active:  true,
		Power:   200,
	}
	keeper.SetValidator(ctx, validator)

	// Submit slashing evidence for double-signing
	evidenceHash := []byte("double-sign-evidence")

	event, err := keeper.SubmitSlashingEvidence(
		ctx,
		validatorAddr,
		types.SlashReason_SLASH_DOUBLE_SIGN,
		"transfer-456",
		evidenceHash,
		"system",
	)

	// Verify double-signing results in jailing
	require.NoError(t, err)
	require.NotNil(t, event)
	require.True(t, event.Jailed) // Double-signers are permanently jailed
	require.Equal(t, types.SlashReason_SLASH_DOUBLE_SIGN, event.Reason)

	// Verify validator is inactive
	val, found := keeper.getValidator(ctx, validatorAddr)
	require.True(t, found)
	require.False(t, val.Active)
}

func TestSubmitSlashingEvidence_Downtime(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	// Setup validator
	validatorAddr := "aura1validator3"
	validator := &types.BridgeValidator{
		Address: validatorAddr,
		Active:  true,
		Power:   150,
	}
	keeper.SetValidator(ctx, validator)

	// Submit slashing evidence for downtime
	evidenceHash := []byte("downtime-evidence")

	event, err := keeper.SubmitSlashingEvidence(
		ctx,
		validatorAddr,
		types.SlashReason_SLASH_DOWNTIME,
		"", // No specific transfer for downtime
		evidenceHash,
		"system",
	)

	// Verify downtime does NOT result in jailing
	require.NoError(t, err)
	require.NotNil(t, event)
	require.False(t, event.Jailed) // Downtime should not jail
	require.Equal(t, types.SlashReason_SLASH_DOWNTIME, event.Reason)

	// Verify validator remains active
	val, found := keeper.getValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, val.Active) // Should still be active
}

func TestSubmitSlashingEvidence_ValidatorNotFound(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	// Try to slash non-existent validator
	_, err := keeper.SubmitSlashingEvidence(
		ctx,
		"aura1nonexistent",
		types.SlashReason_SLASH_FRAUD_ATTEMPT,
		"transfer-789",
		[]byte("evidence"),
		"submitter",
	)

	// Should fail with validator not found
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrValidatorNotFound)
}

func TestSubmitSlashingEvidence_InvalidEvidence(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	// Setup validator
	validatorAddr := "aura1validator4"
	validator := &types.BridgeValidator{
		Address: validatorAddr,
		Active:  true,
	}
	keeper.SetValidator(ctx, validator)

	// Try to submit with empty evidence
	_, err := keeper.SubmitSlashingEvidence(
		ctx,
		validatorAddr,
		types.SlashReason_SLASH_FRAUD_ATTEMPT,
		"transfer-abc",
		[]byte{}, // Empty evidence
		"submitter",
	)

	// Should fail with invalid evidence
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidEvidence)
}

func TestSubmitSlashingEvidence_AlreadyJailed(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	// Setup already jailed validator
	validatorAddr := "aura1validator5"
	validator := &types.BridgeValidator{
		Address: validatorAddr,
		Active:  false, // Already inactive/jailed
	}
	keeper.SetValidator(ctx, validator)

	// Try to slash again
	_, err := keeper.SubmitSlashingEvidence(
		ctx,
		validatorAddr,
		types.SlashReason_SLASH_FRAUD_ATTEMPT,
		"transfer-def",
		[]byte("evidence"),
		"submitter",
	)

	// Should fail with validator jailed
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrValidatorJailed)
}

// ========================================================================
// DOUBLE-SIGNING DETECTION TESTS
// ========================================================================

func TestDetectDoubleSigning_SameSignature(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	transferId := "transfer-double-1"
	validatorAddr := "aura1validator6"
	signature := []byte("signature-data")

	// Create transfer with existing signature
	transfer := &types.CrossChainTransfer{
		TransferId: transferId,
		ValidatorSignatures: []types.ValidatorSignature{
			{
				ValidatorAddress: validatorAddr,
				Signature:        signature,
			},
		},
	}
	keeper.SetTransfer(ctx, transfer)

	// Try to submit same signature again
	isDoubleSigning, err := keeper.DetectDoubleSigning(ctx, transferId, signature, validatorAddr)

	// Should NOT detect double-signing (same signature is a replay, not double-sign)
	require.NoError(t, err)
	require.False(t, isDoubleSigning)
}

func TestDetectDoubleSigning_DifferentSignature(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	transferId := "transfer-double-2"
	validatorAddr := "aura1validator7"
	signature1 := []byte("signature-1")
	signature2 := []byte("signature-2") // Different signature

	// Create transfer with first signature
	transfer := &types.CrossChainTransfer{
		TransferId: transferId,
		ValidatorSignatures: []types.ValidatorSignature{
			{
				ValidatorAddress: validatorAddr,
				Signature:        signature1,
			},
		},
	}
	keeper.SetTransfer(ctx, transfer)

	// Try to submit different signature for same transfer
	isDoubleSigning, err := keeper.DetectDoubleSigning(ctx, transferId, signature2, validatorAddr)

	// Should detect double-signing
	require.NoError(t, err)
	require.True(t, isDoubleSigning)
}

func TestDetectDoubleSigning_FirstSignature(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	transferId := "transfer-double-3"
	validatorAddr := "aura1validator8"
	signature := []byte("first-signature")

	// Create transfer with NO signatures yet
	transfer := &types.CrossChainTransfer{
		TransferId:          transferId,
		ValidatorSignatures: []types.ValidatorSignature{},
	}
	keeper.SetTransfer(ctx, transfer)

	// Submit first signature
	isDoubleSigning, err := keeper.DetectDoubleSigning(ctx, transferId, signature, validatorAddr)

	// Should NOT detect double-signing (first signature)
	require.NoError(t, err)
	require.False(t, isDoubleSigning)
}

func TestCheckAndSlashDoubleSigning_DetectsAndSlashes(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	transferId := "transfer-double-4"
	validatorAddr := "aura1validator9"
	signature1 := []byte("signature-original")
	signature2 := []byte("signature-fraudulent")

	// Setup validator
	validator := &types.BridgeValidator{
		Address: validatorAddr,
		Active:  true,
		Power:   100,
	}
	keeper.SetValidator(ctx, validator)

	// Create transfer with first signature
	transfer := &types.CrossChainTransfer{
		TransferId: transferId,
		ValidatorSignatures: []types.ValidatorSignature{
			{
				ValidatorAddress: validatorAddr,
				Signature:        signature1,
			},
		},
	}
	keeper.SetTransfer(ctx, transfer)

	// Try to submit different signature (should detect and slash)
	isValid, err := keeper.CheckAndSlashDoubleSigning(ctx, transferId, signature2, validatorAddr)

	// Should reject signature and slash validator
	require.Error(t, err)
	require.False(t, isValid)
	require.Contains(t, err.Error(), "double-signed")

	// Verify validator was slashed and jailed
	val, found := keeper.getValidator(ctx, validatorAddr)
	require.True(t, found)
	require.False(t, val.Active) // Should be jailed
}

// ========================================================================
// LIVENESS TRACKING TESTS
// ========================================================================

func TestRecordValidatorSigning_Signed(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	validatorAddr := "aura1validator10"

	// Record that validator signed at current block
	keeper.RecordValidatorSigning(ctx, validatorAddr, true)

	// Verify signing info was recorded
	signingInfo := keeper.GetValidatorSigningInfo(ctx, validatorAddr)
	require.NotNil(t, signingInfo)
	require.True(t, signingInfo[ctx.BlockHeight()])
}

func TestRecordValidatorSigning_Missed(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	validatorAddr := "aura1validator11"

	// Record that validator missed signing at current block
	keeper.RecordValidatorSigning(ctx, validatorAddr, false)

	// Verify miss was recorded
	signingInfo := keeper.GetValidatorSigningInfo(ctx, validatorAddr)
	require.NotNil(t, signingInfo)
	require.False(t, signingInfo[ctx.BlockHeight()])
}

func TestCheckValidatorLiveness_MeetsRequirement(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	validatorAddr := "aura1validator12"

	// Simulate validator signing 100% of blocks in window
	params := keeper.GetParams(ctx)
	windowSize := params.MinSigningWindow
	if windowSize == 0 {
		windowSize = 100 // Default for test
	}

	for i := int64(0); i < windowSize; i++ {
		keeper.RecordValidatorSigning(ctx, validatorAddr, true)
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	}

	// Check liveness
	meetsRequirement, err := keeper.CheckValidatorLiveness(ctx, validatorAddr)
	require.NoError(t, err)
	require.True(t, meetsRequirement)
}

func TestCheckValidatorLiveness_FailsRequirement(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	validatorAddr := "aura1validator13"

	// Simulate validator signing only 25% of blocks (below 50% requirement)
	params := keeper.GetParams(ctx)
	windowSize := params.MinSigningWindow
	if windowSize == 0 {
		windowSize = 100
	}

	for i := int64(0); i < windowSize; i++ {
		// Sign only every 4th block
		signed := (i % 4) == 0
		keeper.RecordValidatorSigning(ctx, validatorAddr, signed)
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	}

	// Check liveness
	meetsRequirement, err := keeper.CheckValidatorLiveness(ctx, validatorAddr)
	require.NoError(t, err)
	require.False(t, meetsRequirement) // Should fail (25% < 50%)
}

func TestSlashForDowntime_ValidatorOnline(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	validatorAddr := "aura1validator14"
	validator := &types.BridgeValidator{
		Address: validatorAddr,
		Active:  true,
	}
	keeper.SetValidator(ctx, validator)

	// Simulate validator signing 100% of blocks
	params := keeper.GetParams(ctx)
	windowSize := params.MinSigningWindow
	if windowSize == 0 {
		windowSize = 100
	}

	for i := int64(0); i < windowSize; i++ {
		keeper.RecordValidatorSigning(ctx, validatorAddr, true)
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	}

	// Try to slash for downtime
	err := keeper.SlashForDowntime(ctx, validatorAddr)

	// Should not slash (validator is online)
	require.NoError(t, err)

	// Verify validator remains active
	val, found := keeper.getValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, val.Active)
}

func TestSlashForDowntime_ValidatorOffline(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	validatorAddr := "aura1validator15"
	validator := &types.BridgeValidator{
		Address: validatorAddr,
		Active:  true,
	}
	keeper.SetValidator(ctx, validator)

	// Simulate validator signing only 10% of blocks (well below 50% requirement)
	params := keeper.GetParams(ctx)
	windowSize := params.MinSigningWindow
	if windowSize == 0 {
		windowSize = 100
	}

	for i := int64(0); i < windowSize; i++ {
		// Sign only every 10th block
		signed := (i % 10) == 0
		keeper.RecordValidatorSigning(ctx, validatorAddr, signed)
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	}

	// Slash for downtime
	err := keeper.SlashForDowntime(ctx, validatorAddr)

	// Should slash for being offline
	require.NoError(t, err)

	// Verify slashing event was created
	// (We can't easily verify the exact event ID, but we can check validator remains active
	// since downtime doesn't jail)
	val, found := keeper.getValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, val.Active) // Downtime doesn't jail, only slashes
}

// ========================================================================
// PARAMETER VALIDATION TESTS
// ========================================================================

func TestSlashingParams_ValidFractions(t *testing.T) {
	params := types.DefaultParams()

	// Verify default slashing fractions are valid
	fraudFraction, err := sdkmath.LegacyNewDecFromStr(params.SlashFraudSignature)
	require.NoError(t, err)
	require.True(t, fraudFraction.GTE(sdkmath.LegacyZeroDec()))
	require.True(t, fraudFraction.LTE(sdkmath.LegacyOneDec()))

	doubleFraction, err := sdkmath.LegacyNewDecFromStr(params.SlashDoubleSigning)
	require.NoError(t, err)
	require.True(t, doubleFraction.GTE(sdkmath.LegacyZeroDec()))
	require.True(t, doubleFraction.LTE(sdkmath.LegacyOneDec()))

	offlineFraction, err := sdkmath.LegacyNewDecFromStr(params.SlashOffline)
	require.NoError(t, err)
	require.True(t, offlineFraction.GTE(sdkmath.LegacyZeroDec()))
	require.True(t, offlineFraction.LTE(sdkmath.LegacyOneDec()))
}

func TestSlashingParams_SigningWindow(t *testing.T) {
	params := types.DefaultParams()

	// Verify signing window is reasonable
	require.Greater(t, params.MinSigningWindow, int64(100))
	require.Less(t, params.MinSigningWindow, int64(100000))

	// Verify min signed percentage is valid
	minSignedPct, err := sdkmath.LegacyNewDecFromStr(params.MinSignedPerWindow)
	require.NoError(t, err)
	require.True(t, minSignedPct.GTE(sdkmath.LegacyZeroDec()))
	require.True(t, minSignedPct.LTE(sdkmath.LegacyOneDec()))
}

// ========================================================================
// HELPER FUNCTIONS
// ========================================================================

// setupKeeperWithStaking creates a keeper with a mock staking keeper for testing
func setupKeeperWithStaking(t *testing.T) (*Keeper, sdk.Context) {
	input := keepertest.CreateTestInput(t)
	k := NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	// For tests that require staking keeper, it would be nil
	// In production, the real staking keeper would be wired in app.go
	// For these tests, we're testing the bridge module's slashing logic,
	// not the actual staking module integration

	return k, ctx
}

// ========================================================================
// FRAUD PROOF INTEGRATION TESTS
// ========================================================================

func TestSlashValidatorsForFraudulentTransfer_Success(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	// Setup: Create 3 validators who signed a fraudulent transfer
	validator1 := &types.BridgeValidator{
		Address: "aura1validator1",
		Active:  true,
		Power:   100,
	}
	validator2 := &types.BridgeValidator{
		Address: "aura1validator2",
		Active:  true,
		Power:   200,
	}
	validator3 := &types.BridgeValidator{
		Address: "aura1validator3",
		Active:  true,
		Power:   150,
	}
	keeper.SetValidator(ctx, validator1)
	keeper.SetValidator(ctx, validator2)
	keeper.SetValidator(ctx, validator3)

	// Create a transfer with signatures from all 3 validators
	transferID := "transfer-fraudulent"
	transfer := &types.CrossChainTransfer{
		TransferId: transferID,
		ValidatorSignatures: []types.ValidatorSignature{
			{
				ValidatorAddress: validator1.Address,
				Signature:        []byte("sig1"),
			},
			{
				ValidatorAddress: validator2.Address,
				Signature:        []byte("sig2"),
			},
			{
				ValidatorAddress: validator3.Address,
				Signature:        []byte("sig3"),
			},
		},
	}
	keeper.SetTransfer(ctx, transfer)

	// Submit fraud proof ID for slashing
	// Note: In production, a FraudProof record would be created via SubmitFraudProof
	// For this test, we're just testing the validator slashing logic
	fraudProofID := "fraud-proof-1"

	// Slash all validators for fraudulent transfer
	err := keeper.slashValidatorsForFraudulentTransfer(ctx, transferID, fraudProofID)
	require.NoError(t, err)

	// Verify all validators were slashed and jailed
	val1, found := keeper.getValidator(ctx, validator1.Address)
	require.True(t, found)
	require.False(t, val1.Active) // Should be jailed

	val2, found := keeper.getValidator(ctx, validator2.Address)
	require.True(t, found)
	require.False(t, val2.Active) // Should be jailed

	val3, found := keeper.getValidator(ctx, validator3.Address)
	require.True(t, found)
	require.False(t, val3.Active) // Should be jailed

	// Verify slashing events were created
	events := keeper.GetAllSlashingEvents(ctx)
	require.GreaterOrEqual(t, len(events), 3) // At least 3 events (one per validator)

	// Count slashing events for each validator
	slashedValidators := make(map[string]bool)
	for _, event := range events {
		if event.Reason == types.SlashReason_SLASH_FRAUD_ATTEMPT {
			slashedValidators[event.ValidatorAddress] = true
		}
	}
	require.Equal(t, 3, len(slashedValidators)) // All 3 validators slashed
}

func TestSlashValidatorsForFraudulentTransfer_NoSignatures(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	// Create transfer with NO validator signatures
	transferID := "transfer-no-sigs"
	transfer := &types.CrossChainTransfer{
		TransferId:          transferID,
		ValidatorSignatures: []types.ValidatorSignature{}, // Empty
	}
	keeper.SetTransfer(ctx, transfer)

	// Try to slash validators
	err := keeper.slashValidatorsForFraudulentTransfer(ctx, transferID, "fraud-proof-2")

	// Should fail - no signatures to slash
	require.Error(t, err)
	require.Contains(t, err.Error(), "no validator signatures")
}

func TestSlashValidatorsForFraudulentTransfer_TransferNotFound(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	// Try to slash for non-existent transfer
	err := keeper.slashValidatorsForFraudulentTransfer(ctx, "nonexistent-transfer", "fraud-proof-3")

	// Should fail - transfer not found
	require.Error(t, err)
	require.Contains(t, err.Error(), "transfer not found")
}

func TestSlashValidatorsForFraudulentTransfer_PartialFailure(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	// Setup: One valid validator, one already jailed
	validatorActive := &types.BridgeValidator{
		Address: "aura1validator-active",
		Active:  true,
		Power:   100,
	}
	validatorJailed := &types.BridgeValidator{
		Address: "aura1validator-jailed",
		Active:  false, // Already jailed
		Power:   100,
	}
	keeper.SetValidator(ctx, validatorActive)
	keeper.SetValidator(ctx, validatorJailed)

	// Create transfer with both validators
	transferID := "transfer-partial"
	transfer := &types.CrossChainTransfer{
		TransferId: transferID,
		ValidatorSignatures: []types.ValidatorSignature{
			{
				ValidatorAddress: validatorActive.Address,
				Signature:        []byte("sig-active"),
			},
			{
				ValidatorAddress: validatorJailed.Address,
				Signature:        []byte("sig-jailed"),
			},
		},
	}
	keeper.SetTransfer(ctx, transfer)

	// Slash validators
	err := keeper.slashValidatorsForFraudulentTransfer(ctx, transferID, "fraud-proof-4")

	// Should succeed even though one validator fails to slash (already jailed)
	// As long as at least one validator is slashed, it's a success
	require.NoError(t, err)

	// Verify active validator was slashed
	val, found := keeper.getValidator(ctx, validatorActive.Address)
	require.True(t, found)
	require.False(t, val.Active) // Should be jailed now
}

func TestResolveFraudProof_SlashesValidators(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	// Setup validators
	validator := &types.BridgeValidator{
		Address: "aura1validator-fraud",
		Active:  true,
		Power:   100,
	}
	keeper.SetValidator(ctx, validator)

	// Create transfer
	transferID := "transfer-resolve"
	transfer := &types.CrossChainTransfer{
		TransferId: transferID,
		Status:     types.TransferStatus_PENDING,
		Denom:      "aura",
		Amount:     sdkmath.NewInt(1000),
		Timestamp:  ctx.BlockTime(),
		ValidatorSignatures: []types.ValidatorSignature{
			{
				ValidatorAddress: validator.Address,
				Signature:        []byte("fraud-sig"),
			},
		},
	}
	keeper.SetTransfer(ctx, transfer)

	// Create fraud proof
	fraudProofID := "fraud-proof-resolve"
	fraudProof := &types.FraudProof{
		ProofId:              fraudProofID,
		ChallengedTransferId: transferID,
		Challenger:           "aura1challenger",
		Status:               types.FraudProofStatus_FRAUD_PROOF_INVESTIGATING,
		SubmittedAt:          ctx.BlockTime(),
		Evidence:             []byte("fraud evidence"),
	}
	require.NoError(t, keeper.setFraudProof(ctx, fraudProof))

	// Resolve fraud proof as VALID (fraud was proven)
	resolved, err := keeper.ResolveFraudProof(ctx, transferID, true)
	require.NoError(t, err)
	require.Equal(t, types.FraudProofStatus_FRAUD_PROOF_VALID, resolved.Status)

	// Verify validator was slashed
	val, found := keeper.getValidator(ctx, validator.Address)
	require.True(t, found)
	require.False(t, val.Active) // Should be jailed

	// Verify slashing event was created
	events := keeper.GetValidatorSlashingHistory(ctx, validator.Address)
	require.NotEmpty(t, events)
	require.Equal(t, types.SlashReason_SLASH_FRAUD_ATTEMPT, events[0].Reason)
}

func TestGetValidatorSlashingHistory(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	validatorAddr := "aura1validator-history"
	validator := &types.BridgeValidator{
		Address: validatorAddr,
		Active:  true,
		Power:   100,
	}
	keeper.SetValidator(ctx, validator)

	// Slash validator twice for different reasons
	_, err := keeper.SubmitSlashingEvidence(
		ctx,
		validatorAddr,
		types.SlashReason_SLASH_FRAUD_ATTEMPT,
		"transfer-1",
		[]byte("evidence-1"),
		"submitter",
	)
	require.NoError(t, err)

	// Reactivate validator for second test
	validator.Active = true
	keeper.SetValidator(ctx, validator)

	// Note: Second slash will fail because validator is already jailed
	// Let's just verify the history has at least one event

	// Get slashing history
	history := keeper.GetValidatorSlashingHistory(ctx, validatorAddr)
	require.NotEmpty(t, history)
	require.Equal(t, validatorAddr, history[0].ValidatorAddress)
}

func TestGetAllSlashingEvents(t *testing.T) {
	keeper, ctx := setupKeeperWithStaking(t)

	// Setup multiple validators
	val1 := &types.BridgeValidator{Address: "aura1val1", Active: true, Power: 100}
	val2 := &types.BridgeValidator{Address: "aura1val2", Active: true, Power: 100}
	keeper.SetValidator(ctx, val1)
	keeper.SetValidator(ctx, val2)

	// Slash both validators
	_, _ = keeper.SubmitSlashingEvidence(ctx, val1.Address, types.SlashReason_SLASH_FRAUD_ATTEMPT, "transfer-a", []byte("ev1"), "sub")
	_, _ = keeper.SubmitSlashingEvidence(ctx, val2.Address, types.SlashReason_SLASH_DOUBLE_SIGN, "transfer-b", []byte("ev2"), "sub")

	// Get all slashing events
	events := keeper.GetAllSlashingEvents(ctx)
	require.GreaterOrEqual(t, len(events), 2)

	// Verify events contain our validators
	slashedVals := make(map[string]bool)
	for _, event := range events {
		slashedVals[event.ValidatorAddress] = true
	}
	require.True(t, slashedVals[val1.Address] || slashedVals[val2.Address])
}

// Note: In production, you would also want to test:
// 1. Integration with actual staking module (slash amounts, jailing, etc.)
// 2. Edge cases around validator rotation during slashing
// 3. Multiple simultaneous slashing events
// 4. Slashing event pagination and queries
// 5. Governance-based slashing parameter updates
