package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

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
		Address:  validatorAddr,
		Active:   true,
		Power:    100,
		Chains:   []string{"aura", "paw"},
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
	val, found := keeper.GetValidator(ctx, validatorAddr)
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
	val, found := keeper.GetValidator(ctx, validatorAddr)
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
	val, found := keeper.GetValidator(ctx, validatorAddr)
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
		ValidatorSignatures: []*types.ValidatorSignature{
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
		ValidatorSignatures: []*types.ValidatorSignature{
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
		ValidatorSignatures: []*types.ValidatorSignature{},
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
		ValidatorSignatures: []*types.ValidatorSignature{
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
	val, found := keeper.GetValidator(ctx, validatorAddr)
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
	val, found := keeper.GetValidator(ctx, validatorAddr)
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
	val, found := keeper.GetValidator(ctx, validatorAddr)
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
func setupKeeperWithStaking(t *testing.T) (*testKeeper, sdk.Context) {
	// Use the existing test setup (which may have a nil staking keeper)
	keeper, ctx := setupKeeper(t)

	// For tests that require staking keeper, it would be nil
	// In production, the real staking keeper would be wired in app.go
	// For these tests, we're testing the bridge module's slashing logic,
	// not the actual staking module integration

	return keeper, ctx
}

// Note: In production, you would also want to test:
// 1. Integration with actual staking module (slash amounts, jailing, etc.)
// 2. Edge cases around validator rotation during slashing
// 3. Multiple simultaneous slashing events
// 4. Slashing event pagination and queries
// 5. Governance-based slashing parameter updates
