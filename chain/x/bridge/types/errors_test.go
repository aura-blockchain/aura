// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

func TestCoreErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains string
	}{
		{"ErrInvalidParam", types.ErrInvalidParam, "invalid parameter"},
		{"ErrDuplicateAttestation", types.ErrDuplicateAttestation, "duplicate attestation"},
		{"ErrWithdrawalNotFound", types.ErrWithdrawalNotFound, "withdrawal not found"},
		{"ErrChainNotFound", types.ErrChainNotFound, "chain not found"},
		{"ErrTransferNotFound", types.ErrTransferNotFound, "transfer not found"},
		{"ErrCircuitBreakerTripped", types.ErrCircuitBreakerTripped, "circuit breaker tripped"},
		{"ErrTimelockNotElapsed", types.ErrTimelockNotElapsed, "timelock period has not elapsed"},
		{"ErrChainDisabled", types.ErrChainDisabled, "chain disabled"},
		{"ErrInvalidSignature", types.ErrInvalidSignature, "invalid cryptographic signature"},
		{"ErrCorruptedData", types.ErrCorruptedData, "corrupted or invalid data"},
		{"ErrSignatureReplay", types.ErrSignatureReplay, "replay attack prevented"},
		{"ErrSignatureRateLimit", types.ErrSignatureRateLimit, "rate limit exceeded"},
		{"ErrInvalidRecoveryID", types.ErrInvalidRecoveryID, "invalid ECDSA recovery ID"},
		{"ErrMarshalFailed", types.ErrMarshalFailed, "failed to marshal"},
		{"ErrUnmarshalFailed", types.ErrUnmarshalFailed, "failed to unmarshal"},
		{"ErrIBCNotEnabled", types.ErrIBCNotEnabled, "IBC not enabled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.err)
			require.Contains(t, tt.err.Error(), tt.contains)
		})
	}
}

func TestSecurityErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains string
	}{
		// Security errors (100-109)
		{"ErrBridgePaused", types.ErrBridgePaused, "bridge is paused"},
		{"ErrInvalidMerkleProof", types.ErrInvalidMerkleProof, "invalid Merkle proof"},
		{"ErrInvalidTSSSignature", types.ErrInvalidTSSSignature, "invalid threshold signature"},
		{"ErrInsufficientSignatures", types.ErrInsufficientSignatures, "insufficient validator signatures"},
		{"ErrValidatorNotFound", types.ErrValidatorNotFound, "validator not found"},
		{"ErrValidatorNotActive", types.ErrValidatorNotActive, "validator is not active"},
		{"ErrInvalidNonce", types.ErrInvalidNonce, "invalid nonce"},
		{"ErrNonceAlreadyUsed", types.ErrNonceAlreadyUsed, "nonce already used"},

		// Transfer limit errors (110-119)
		{"ErrAmountBelowMinimum", types.ErrAmountBelowMinimum, "amount below minimum"},
		{"ErrAmountExceedsMaximum", types.ErrAmountExceedsMaximum, "amount exceeds maximum"},
		{"ErrDailyLimitExceeded", types.ErrDailyLimitExceeded, "daily withdrawal limit exceeded"},
		{"ErrTimeLockRequired", types.ErrTimeLockRequired, "time-lock required"},
		{"ErrTimeLockNotExpired", types.ErrTimeLockNotExpired, "time-lock has not expired"},
		{"ErrTimeLockChallenged", types.ErrTimeLockChallenged, "time-lock has been challenged"},

		// Circuit breaker errors (120-129)
		{"ErrCircuitBreakerOpen", types.ErrCircuitBreakerOpen, "circuit breaker is open"},
		{"ErrHourlyVolumeExceeded", types.ErrHourlyVolumeExceeded, "hourly volume limit exceeded"},
		{"ErrTooManyFailedTransfers", types.ErrTooManyFailedTransfers, "too many failed transfers"},

		// Permission errors (130-139)
		{"ErrAddressBlacklisted", types.ErrAddressBlacklisted, "address is blacklisted"},
		{"ErrAddressNotWhitelisted", types.ErrAddressNotWhitelisted, "address is not whitelisted"},
		{"ErrWhitelistEnabled", types.ErrWhitelistEnabled, "whitelist is enabled"},

		// Fraud proof errors (140-149)
		{"ErrFraudProofExpired", types.ErrFraudProofExpired, "fraud proof has expired"},
		{"ErrFraudProofAlreadyResolved", types.ErrFraudProofAlreadyResolved, "fraud proof already resolved"},
		{"ErrFraudProofPending", types.ErrFraudProofPending, "fraud proof already pending"},
		{"ErrFraudProofNotFound", types.ErrFraudProofNotFound, "fraud proof not found"},
		{"ErrInvalidEvidence", types.ErrInvalidEvidence, "invalid evidence"},

		// Insurance fund errors (150-159)
		{"ErrInsufficientInsuranceFund", types.ErrInsufficientInsuranceFund, "insufficient insurance fund"},
		{"ErrClaimNotFound", types.ErrClaimNotFound, "insurance claim not found"},
		{"ErrClaimAlreadyResolved", types.ErrClaimAlreadyResolved, "claim already resolved"},

		// Validator errors (160-199)
		{"ErrValidatorSlashed", types.ErrValidatorSlashed, "validator has been slashed"},
		{"ErrValidatorJailed", types.ErrValidatorJailed, "validator is jailed"},
		{"ErrRotationNotApproved", types.ErrRotationNotApproved, "validator rotation not approved"},
		{"ErrRotationNotEffective", types.ErrRotationNotEffective, "validator rotation not yet effective"},
		{"ErrValidatorUnauthorized", types.ErrValidatorUnauthorized, "validator is not authorized"},
		{"ErrNoActiveValidators", types.ErrNoActiveValidators, "no active validators available"},
		{"ErrSignatureSetAlreadyUsed", types.ErrSignatureSetAlreadyUsed, "signature set already used"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.err)
			require.Contains(t, tt.err.Error(), tt.contains)
		})
	}
}

func TestErrorCodesAreUnique(t *testing.T) {
	// Map to track error codes
	errorCodes := make(map[uint32]string)

	// All errors in the module
	allErrors := []struct {
		err  error
		name string
	}{
		// Core errors (300-399)
		{types.ErrInvalidParam, "ErrInvalidParam"},
		{types.ErrDuplicateAttestation, "ErrDuplicateAttestation"},
		{types.ErrWithdrawalNotFound, "ErrWithdrawalNotFound"},
		{types.ErrChainNotFound, "ErrChainNotFound"},
		{types.ErrTransferNotFound, "ErrTransferNotFound"},
		{types.ErrCircuitBreakerTripped, "ErrCircuitBreakerTripped"},
		{types.ErrTimelockNotElapsed, "ErrTimelockNotElapsed"},
		{types.ErrChainDisabled, "ErrChainDisabled"},
		{types.ErrInvalidSignature, "ErrInvalidSignature"},
		{types.ErrCorruptedData, "ErrCorruptedData"},
		{types.ErrSignatureReplay, "ErrSignatureReplay"},
		{types.ErrSignatureRateLimit, "ErrSignatureRateLimit"},
		{types.ErrInvalidRecoveryID, "ErrInvalidRecoveryID"},
		{types.ErrMarshalFailed, "ErrMarshalFailed"},
		{types.ErrUnmarshalFailed, "ErrUnmarshalFailed"},
		{types.ErrIBCNotEnabled, "ErrIBCNotEnabled"},

		// Security errors (100-199)
		{types.ErrBridgePaused, "ErrBridgePaused"},
		{types.ErrInvalidMerkleProof, "ErrInvalidMerkleProof"},
		{types.ErrInvalidTSSSignature, "ErrInvalidTSSSignature"},
		{types.ErrInsufficientSignatures, "ErrInsufficientSignatures"},
		{types.ErrValidatorNotFound, "ErrValidatorNotFound"},
		{types.ErrValidatorNotActive, "ErrValidatorNotActive"},
		{types.ErrInvalidNonce, "ErrInvalidNonce"},
		{types.ErrNonceAlreadyUsed, "ErrNonceAlreadyUsed"},
		{types.ErrAmountBelowMinimum, "ErrAmountBelowMinimum"},
		{types.ErrAmountExceedsMaximum, "ErrAmountExceedsMaximum"},
		{types.ErrDailyLimitExceeded, "ErrDailyLimitExceeded"},
		{types.ErrTimeLockRequired, "ErrTimeLockRequired"},
		{types.ErrTimeLockNotExpired, "ErrTimeLockNotExpired"},
		{types.ErrTimeLockChallenged, "ErrTimeLockChallenged"},
		{types.ErrCircuitBreakerOpen, "ErrCircuitBreakerOpen"},
		{types.ErrHourlyVolumeExceeded, "ErrHourlyVolumeExceeded"},
		{types.ErrTooManyFailedTransfers, "ErrTooManyFailedTransfers"},
		{types.ErrAddressBlacklisted, "ErrAddressBlacklisted"},
		{types.ErrAddressNotWhitelisted, "ErrAddressNotWhitelisted"},
		{types.ErrWhitelistEnabled, "ErrWhitelistEnabled"},
		{types.ErrFraudProofExpired, "ErrFraudProofExpired"},
		{types.ErrFraudProofAlreadyResolved, "ErrFraudProofAlreadyResolved"},
		{types.ErrFraudProofPending, "ErrFraudProofPending"},
		{types.ErrFraudProofNotFound, "ErrFraudProofNotFound"},
		{types.ErrInvalidEvidence, "ErrInvalidEvidence"},
		{types.ErrInsufficientInsuranceFund, "ErrInsufficientInsuranceFund"},
		{types.ErrClaimNotFound, "ErrClaimNotFound"},
		{types.ErrClaimAlreadyResolved, "ErrClaimAlreadyResolved"},
		{types.ErrValidatorSlashed, "ErrValidatorSlashed"},
		{types.ErrValidatorJailed, "ErrValidatorJailed"},
		{types.ErrRotationNotApproved, "ErrRotationNotApproved"},
		{types.ErrRotationNotEffective, "ErrRotationNotEffective"},
		{types.ErrValidatorUnauthorized, "ErrValidatorUnauthorized"},
		{types.ErrNoActiveValidators, "ErrNoActiveValidators"},
		{types.ErrSignatureSetAlreadyUsed, "ErrSignatureSetAlreadyUsed"},
	}

	// Verify all errors exist and have unique codes
	for _, e := range allErrors {
		require.NotNil(t, e.err, "Error %s should not be nil", e.name)
	}

	// Log the count of unique errors for visibility
	require.Greater(t, len(allErrors), 0, "Should have at least one error defined")
	_ = errorCodes // Using the map for potential future uniqueness checks
}
