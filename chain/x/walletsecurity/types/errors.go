package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Wallet security module error codes
var (
	// Hardware Wallet Errors (1-9)
	ErrInvalidHardwareWallet     = errorsmod.Register(ModuleName, 1, "invalid hardware wallet")
	ErrHardwareWalletNotFound    = errorsmod.Register(ModuleName, 2, "hardware wallet not found")
	ErrHardwareWalletExists      = errorsmod.Register(ModuleName, 3, "hardware wallet already exists")
	ErrInvalidDeviceSignature    = errorsmod.Register(ModuleName, 4, "invalid device signature")
	ErrUnsupportedHardwareWallet = errorsmod.Register(ModuleName, 5, "unsupported hardware wallet type")

	// Multi-Sig Wallet Errors (10-19)
	ErrInvalidMultiSigConfig  = errorsmod.Register(ModuleName, 10, "invalid multi-sig configuration")
	ErrMultiSigWalletNotFound = errorsmod.Register(ModuleName, 11, "multi-sig wallet not found")
	ErrInsufficientSignatures = errorsmod.Register(ModuleName, 12, "insufficient signatures")
	ErrInvalidSigner          = errorsmod.Register(ModuleName, 13, "invalid signer")
	ErrSignatureExists        = errorsmod.Register(ModuleName, 14, "signature already exists")
	ErrInvalidThreshold       = errorsmod.Register(ModuleName, 15, "invalid threshold")
	ErrInvalidWeights         = errorsmod.Register(ModuleName, 16, "invalid signer weights")
	ErrMultiSigTxNotFound     = errorsmod.Register(ModuleName, 17, "multi-sig transaction not found")
	ErrMultiSigTxExpired      = errorsmod.Register(ModuleName, 18, "multi-sig transaction expired")

	// Social Recovery Errors (20-29)
	ErrInvalidRecoveryConfig    = errorsmod.Register(ModuleName, 20, "invalid recovery configuration")
	ErrRecoveryNotEnabled       = errorsmod.Register(ModuleName, 21, "recovery not enabled")
	ErrInvalidGuardian          = errorsmod.Register(ModuleName, 22, "invalid guardian")
	ErrGuardianExists           = errorsmod.Register(ModuleName, 23, "guardian already exists")
	ErrGuardianNotFound         = errorsmod.Register(ModuleName, 24, "guardian not found")
	ErrRecoveryRequestNotFound  = errorsmod.Register(ModuleName, 25, "recovery request not found")
	ErrRecoveryNotReady         = errorsmod.Register(ModuleName, 26, "recovery not ready to execute")
	ErrRecoveryDelayNotElapsed  = errorsmod.Register(ModuleName, 27, "recovery delay not elapsed")
	ErrInsufficientApprovals    = errorsmod.Register(ModuleName, 28, "insufficient guardian approvals")
	ErrRecoveryAlreadyExecuted  = errorsmod.Register(ModuleName, 29, "recovery already executed")
	ErrInvalidRecoveryThreshold = errorsmod.Register(ModuleName, 30, "invalid recovery threshold")

	// Transaction Simulation Errors (40-49)
	ErrSimulationFailed       = errorsmod.Register(ModuleName, 40, "transaction simulation failed")
	ErrInvalidTransactionData = errorsmod.Register(ModuleName, 41, "invalid transaction data")
	ErrSimulationRiskTooHigh  = errorsmod.Register(ModuleName, 42, "simulation detected high risk")

	// Phishing Protection Errors (50-59)
	ErrDomainNotVerified     = errorsmod.Register(ModuleName, 50, "domain not verified")
	ErrDomainBlacklisted     = errorsmod.Register(ModuleName, 51, "domain blacklisted")
	ErrInvalidCertificate    = errorsmod.Register(ModuleName, 52, "invalid SSL certificate")
	ErrSuspiciousTransaction = errorsmod.Register(ModuleName, 53, "suspicious transaction detected")

	// Address Checksum Errors (60-69)
	ErrInvalidChecksum         = errorsmod.Register(ModuleName, 60, "invalid address checksum")
	ErrChecksumMismatch        = errorsmod.Register(ModuleName, 61, "checksum mismatch")
	ErrUnsupportedChecksumAlgo = errorsmod.Register(ModuleName, 62, "unsupported checksum algorithm")

	// Spending Limit Errors (70-79)
	ErrSpendingLimitExceeded = errorsmod.Register(ModuleName, 70, "spending limit exceeded")
	ErrInvalidSpendingLimit  = errorsmod.Register(ModuleName, 71, "invalid spending limit")
	ErrSpendingLimitNotFound = errorsmod.Register(ModuleName, 72, "spending limit not found")

	// Session Errors (80-89)
	ErrSessionNotFound      = errorsmod.Register(ModuleName, 80, "session not found")
	ErrSessionExpired       = errorsmod.Register(ModuleName, 81, "session expired")
	ErrSessionLocked        = errorsmod.Register(ModuleName, 82, "session locked")
	ErrSessionInactive      = errorsmod.Register(ModuleName, 83, "session inactive due to timeout")
	ErrInvalidSessionConfig = errorsmod.Register(ModuleName, 84, "invalid session configuration")
ErrSessionTimeout       = errorsmod.Register(ModuleName, 85, "session timeout")

// Biometric Errors (90-99)
ErrBiometricNotEnrolled     = errorsmod.Register(ModuleName, 90, "biometric not enrolled")
ErrBiometricAuthFailed      = errorsmod.Register(ModuleName, 91, "biometric authentication failed")
	ErrBiometricLockedOut       = errorsmod.Register(ModuleName, 92, "biometric locked out")
	ErrInvalidBiometricData     = errorsmod.Register(ModuleName, 93, "invalid biometric data")
	ErrBiometricAlreadyEnrolled = errorsmod.Register(ModuleName, 94, "biometric already enrolled")

	// Secure Enclave Errors (100-109)
	ErrEnclaveNotAvailable    = errorsmod.Register(ModuleName, 100, "secure enclave not available")
	ErrEnclaveStorageFailed   = errorsmod.Register(ModuleName, 101, "secure enclave storage failed")
	ErrEnclaveRetrievalFailed = errorsmod.Register(ModuleName, 102, "secure enclave retrieval failed")
	ErrInvalidAttestation     = errorsmod.Register(ModuleName, 103, "invalid attestation certificate")

	// Backup Errors (110-119)
	ErrBackupNotFound           = errorsmod.Register(ModuleName, 110, "backup not found")
	ErrBackupDecryptionFailed   = errorsmod.Register(ModuleName, 111, "backup decryption failed")
	ErrBackupVerificationFailed = errorsmod.Register(ModuleName, 112, "backup verification failed")
	ErrInvalidBackupData        = errorsmod.Register(ModuleName, 113, "invalid backup data")

	// Dust Attack Errors (120-129)
ErrDustTransactionBlocked = errorsmod.Register(ModuleName, 120, "dust transaction blocked")
ErrSuspiciousPattern      = errorsmod.Register(ModuleName, 121, "suspicious pattern detected")
ErrDustFilterNotEnabled   = errorsmod.Register(ModuleName, 122, "dust filter not enabled")

// General Errors (900-999)
ErrUnauthorized    = errorsmod.Register(ModuleName, 900, "unauthorized")
ErrInactiveSession = errorsmod.Register(ModuleName, 901, "inactive or missing auth session")
ErrAuthRateLimited = errorsmod.Register(ModuleName, 902, "too many authorization attempts")
ErrInvalidInput    = errorsmod.Register(ModuleName, 903, "invalid input")
ErrInternalError   = errorsmod.Register(ModuleName, 999, "internal error")
)
