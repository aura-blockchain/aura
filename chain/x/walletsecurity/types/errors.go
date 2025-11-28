package types

import (
	"errors"
)

var (
	// Hardware Wallet Errors
	ErrInvalidHardwareWallet     = errors.New("invalid hardware wallet")
	ErrHardwareWalletNotFound    = errors.New("hardware wallet not found")
	ErrHardwareWalletExists      = errors.New("hardware wallet already exists")
	ErrInvalidDeviceSignature    = errors.New("invalid device signature")
	ErrUnsupportedHardwareWallet = errors.New("unsupported hardware wallet type")

	// Multi-Sig Wallet Errors
	ErrInvalidMultiSigConfig  = errors.New("invalid multi-sig configuration")
	ErrMultiSigWalletNotFound = errors.New("multi-sig wallet not found")
	ErrInsufficientSignatures = errors.New("insufficient signatures")
	ErrInvalidSigner          = errors.New("invalid signer")
	ErrSignatureExists        = errors.New("signature already exists")
	ErrInvalidThreshold       = errors.New("invalid threshold")
	ErrInvalidWeights         = errors.New("invalid signer weights")
	ErrMultiSigTxNotFound     = errors.New("multi-sig transaction not found")
	ErrMultiSigTxExpired      = errors.New("multi-sig transaction expired")

	// Social Recovery Errors
	ErrInvalidRecoveryConfig    = errors.New("invalid recovery configuration")
	ErrRecoveryNotEnabled       = errors.New("recovery not enabled")
	ErrInvalidGuardian          = errors.New("invalid guardian")
	ErrGuardianExists           = errors.New("guardian already exists")
	ErrGuardianNotFound         = errors.New("guardian not found")
	ErrRecoveryRequestNotFound  = errors.New("recovery request not found")
	ErrRecoveryNotReady         = errors.New("recovery not ready to execute")
	ErrRecoveryDelayNotElapsed  = errors.New("recovery delay not elapsed")
	ErrInsufficientApprovals    = errors.New("insufficient guardian approvals")
	ErrRecoveryAlreadyExecuted  = errors.New("recovery already executed")
	ErrInvalidRecoveryThreshold = errors.New("invalid recovery threshold")

	// Transaction Simulation Errors
	ErrSimulationFailed       = errors.New("transaction simulation failed")
	ErrInvalidTransactionData = errors.New("invalid transaction data")
	ErrSimulationRiskTooHigh  = errors.New("simulation detected high risk")

	// Phishing Protection Errors
	ErrDomainNotVerified     = errors.New("domain not verified")
	ErrDomainBlacklisted     = errors.New("domain blacklisted")
	ErrInvalidCertificate    = errors.New("invalid SSL certificate")
	ErrSuspiciousTransaction = errors.New("suspicious transaction detected")

	// Address Checksum Errors
	ErrInvalidChecksum         = errors.New("invalid address checksum")
	ErrChecksumMismatch        = errors.New("checksum mismatch")
	ErrUnsupportedChecksumAlgo = errors.New("unsupported checksum algorithm")

	// Spending Limit Errors
	ErrSpendingLimitExceeded = errors.New("spending limit exceeded")
	ErrInvalidSpendingLimit  = errors.New("invalid spending limit")
	ErrSpendingLimitNotFound = errors.New("spending limit not found")

	// Session Errors
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionExpired       = errors.New("session expired")
	ErrSessionLocked        = errors.New("session locked")
	ErrSessionInactive      = errors.New("session inactive due to timeout")
	ErrInvalidSessionConfig = errors.New("invalid session configuration")
	ErrSessionTimeout       = errors.New("session timeout")

	// Biometric Errors
	ErrBiometricNotEnrolled     = errors.New("biometric not enrolled")
	ErrBiometricAuthFailed      = errors.New("biometric authentication failed")
	ErrBiometricLockedOut       = errors.New("biometric locked out")
	ErrInvalidBiometricData     = errors.New("invalid biometric data")
	ErrBiometricAlreadyEnrolled = errors.New("biometric already enrolled")

	// Secure Enclave Errors
	ErrEnclaveNotAvailable    = errors.New("secure enclave not available")
	ErrEnclaveStorageFailed   = errors.New("secure enclave storage failed")
	ErrEnclaveRetrievalFailed = errors.New("secure enclave retrieval failed")
	ErrInvalidAttestation     = errors.New("invalid attestation certificate")

	// Backup Errors
	ErrBackupNotFound           = errors.New("backup not found")
	ErrBackupDecryptionFailed   = errors.New("backup decryption failed")
	ErrBackupVerificationFailed = errors.New("backup verification failed")
	ErrInvalidBackupData        = errors.New("invalid backup data")

	// Dust Attack Errors
	ErrDustTransactionBlocked = errors.New("dust transaction blocked")
	ErrSuspiciousPattern      = errors.New("suspicious pattern detected")
	ErrDustFilterNotEnabled   = errors.New("dust filter not enabled")

	// General Errors
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInvalidInput  = errors.New("invalid input")
	ErrInternalError = errors.New("internal error")
)
