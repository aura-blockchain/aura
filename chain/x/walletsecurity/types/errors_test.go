// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHardwareWalletErrors_Defined(t *testing.T) {
	errors := []error{
		ErrInvalidHardwareWallet,
		ErrHardwareWalletNotFound,
		ErrHardwareWalletExists,
		ErrInvalidDeviceSignature,
		ErrUnsupportedHardwareWallet,
	}

	for _, err := range errors {
		require.NotNil(t, err)
	}
}

func TestMultiSigWalletErrors_Defined(t *testing.T) {
	errors := []error{
		ErrInvalidMultiSigConfig,
		ErrMultiSigWalletNotFound,
		ErrInsufficientSignatures,
		ErrInvalidSigner,
		ErrSignatureExists,
		ErrInvalidThreshold,
		ErrInvalidWeights,
		ErrMultiSigTxNotFound,
		ErrMultiSigTxExpired,
	}

	for _, err := range errors {
		require.NotNil(t, err)
	}
}

func TestSocialRecoveryErrors_Defined(t *testing.T) {
	errors := []error{
		ErrInvalidRecoveryConfig,
		ErrRecoveryNotEnabled,
		ErrInvalidGuardian,
		ErrGuardianExists,
		ErrGuardianNotFound,
		ErrRecoveryRequestNotFound,
		ErrRecoveryNotReady,
		ErrRecoveryDelayNotElapsed,
		ErrInsufficientApprovals,
		ErrRecoveryAlreadyExecuted,
		ErrInvalidRecoveryThreshold,
	}

	for _, err := range errors {
		require.NotNil(t, err)
	}
}

func TestTransactionSimulationErrors_Defined(t *testing.T) {
	errors := []error{
		ErrSimulationFailed,
		ErrInvalidTransactionData,
		ErrSimulationRiskTooHigh,
	}

	for _, err := range errors {
		require.NotNil(t, err)
	}
}

func TestPhishingProtectionErrors_Defined(t *testing.T) {
	errors := []error{
		ErrDomainNotVerified,
		ErrDomainBlacklisted,
		ErrInvalidCertificate,
		ErrSuspiciousTransaction,
	}

	for _, err := range errors {
		require.NotNil(t, err)
	}
}

func TestAddressChecksumErrors_Defined(t *testing.T) {
	errors := []error{
		ErrInvalidChecksum,
		ErrChecksumMismatch,
		ErrUnsupportedChecksumAlgo,
	}

	for _, err := range errors {
		require.NotNil(t, err)
	}
}

func TestSpendingLimitErrors_Defined(t *testing.T) {
	errors := []error{
		ErrSpendingLimitExceeded,
		ErrInvalidSpendingLimit,
		ErrSpendingLimitNotFound,
	}

	for _, err := range errors {
		require.NotNil(t, err)
	}
}

func TestSessionErrors_Defined(t *testing.T) {
	errors := []error{
		ErrSessionNotFound,
		ErrSessionExpired,
		ErrSessionLocked,
		ErrInvalidSessionConfig,
		ErrSessionTimeout,
	}

	for _, err := range errors {
		require.NotNil(t, err)
	}
}

func TestBiometricErrors_Defined(t *testing.T) {
	errors := []error{
		ErrBiometricNotEnrolled,
		ErrBiometricAuthFailed,
		ErrBiometricLockedOut,
		ErrInvalidBiometricData,
		ErrBiometricAlreadyEnrolled,
	}

	for _, err := range errors {
		require.NotNil(t, err)
	}
}

func TestSecureEnclaveErrors_Defined(t *testing.T) {
	errors := []error{
		ErrEnclaveNotAvailable,
		ErrEnclaveStorageFailed,
		ErrEnclaveRetrievalFailed,
		ErrInvalidAttestation,
	}

	for _, err := range errors {
		require.NotNil(t, err)
	}
}

func TestBackupErrors_Defined(t *testing.T) {
	errors := []error{
		ErrBackupNotFound,
		ErrBackupDecryptionFailed,
		ErrBackupVerificationFailed,
		ErrInvalidBackupData,
	}

	for _, err := range errors {
		require.NotNil(t, err)
	}
}

func TestDustAttackErrors_Defined(t *testing.T) {
	errors := []error{
		ErrDustTransactionBlocked,
		ErrSuspiciousPattern,
		ErrDustFilterNotEnabled,
	}

	for _, err := range errors {
		require.NotNil(t, err)
	}
}

func TestGeneralErrors_Defined(t *testing.T) {
	errors := []error{
		ErrUnauthorized,
		ErrInvalidInput,
		ErrInternalError,
	}

	for _, err := range errors {
		require.NotNil(t, err)
	}
}

func TestErrors_Messages(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		message string
	}{
		{"invalid hardware wallet", ErrInvalidHardwareWallet, "invalid hardware wallet"},
		{"hardware wallet not found", ErrHardwareWalletNotFound, "hardware wallet not found"},
		{"invalid multi-sig config", ErrInvalidMultiSigConfig, "invalid multi-sig configuration"},
		{"insufficient signatures", ErrInsufficientSignatures, "insufficient signatures"},
		{"invalid guardian", ErrInvalidGuardian, "invalid guardian"},
		{"simulation failed", ErrSimulationFailed, "transaction simulation failed"},
		{"domain not verified", ErrDomainNotVerified, "domain not verified"},
		{"invalid checksum", ErrInvalidChecksum, "invalid address checksum"},
		{"spending limit exceeded", ErrSpendingLimitExceeded, "spending limit exceeded"},
		{"session not found", ErrSessionNotFound, "session not found"},
		{"biometric not enrolled", ErrBiometricNotEnrolled, "biometric not enrolled"},
		{"enclave not available", ErrEnclaveNotAvailable, "secure enclave not available"},
		{"backup not found", ErrBackupNotFound, "backup not found"},
		{"dust transaction blocked", ErrDustTransactionBlocked, "dust transaction blocked"},
		{"unauthorized", ErrUnauthorized, "unauthorized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.message, tt.err.Error())
		})
	}
}

func TestErrors_Unique(t *testing.T) {
	allErrors := []error{
		// Hardware Wallet
		ErrInvalidHardwareWallet,
		ErrHardwareWalletNotFound,
		ErrHardwareWalletExists,
		ErrInvalidDeviceSignature,
		ErrUnsupportedHardwareWallet,
		// Multi-Sig
		ErrInvalidMultiSigConfig,
		ErrMultiSigWalletNotFound,
		ErrInsufficientSignatures,
		ErrInvalidSigner,
		ErrSignatureExists,
		ErrInvalidThreshold,
		ErrInvalidWeights,
		ErrMultiSigTxNotFound,
		ErrMultiSigTxExpired,
		// Social Recovery
		ErrInvalidRecoveryConfig,
		ErrRecoveryNotEnabled,
		ErrInvalidGuardian,
		ErrGuardianExists,
		ErrGuardianNotFound,
		ErrRecoveryRequestNotFound,
		ErrRecoveryNotReady,
		ErrRecoveryDelayNotElapsed,
		ErrInsufficientApprovals,
		ErrRecoveryAlreadyExecuted,
		ErrInvalidRecoveryThreshold,
		// Transaction Simulation
		ErrSimulationFailed,
		ErrInvalidTransactionData,
		ErrSimulationRiskTooHigh,
		// Phishing Protection
		ErrDomainNotVerified,
		ErrDomainBlacklisted,
		ErrInvalidCertificate,
		ErrSuspiciousTransaction,
		// Address Checksum
		ErrInvalidChecksum,
		ErrChecksumMismatch,
		ErrUnsupportedChecksumAlgo,
		// Spending Limit
		ErrSpendingLimitExceeded,
		ErrInvalidSpendingLimit,
		ErrSpendingLimitNotFound,
		// Session
		ErrSessionNotFound,
		ErrSessionExpired,
		ErrSessionLocked,
		ErrInvalidSessionConfig,
		ErrSessionTimeout,
		// Biometric
		ErrBiometricNotEnrolled,
		ErrBiometricAuthFailed,
		ErrBiometricLockedOut,
		ErrInvalidBiometricData,
		ErrBiometricAlreadyEnrolled,
		// Secure Enclave
		ErrEnclaveNotAvailable,
		ErrEnclaveStorageFailed,
		ErrEnclaveRetrievalFailed,
		ErrInvalidAttestation,
		// Backup
		ErrBackupNotFound,
		ErrBackupDecryptionFailed,
		ErrBackupVerificationFailed,
		ErrInvalidBackupData,
		// Dust Attack
		ErrDustTransactionBlocked,
		ErrSuspiciousPattern,
		ErrDustFilterNotEnabled,
		// General
		ErrUnauthorized,
		ErrInvalidInput,
		ErrInternalError,
	}

	// Ensure all errors are unique instances
	for i, err1 := range allErrors {
		for j, err2 := range allErrors {
			if i != j {
				require.NotEqual(t, err1, err2, "errors at index %d and %d should not be equal", i, j)
			}
		}
	}
}

func TestErrors_CanBeCompared(t *testing.T) {
	// Test that errors can be compared using ==
	err := ErrInvalidHardwareWallet
	require.True(t, err == ErrInvalidHardwareWallet)
	require.False(t, err == ErrHardwareWalletNotFound)
}
