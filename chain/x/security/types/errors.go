// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"cosmossdk.io/errors"
)

// Security module sentinel errors
// Organized by security domain for clarity

var (
	// Module-level errors (1-9)
	ErrInvalidRequest = errors.Register(ModuleName, 1, "invalid request")
	ErrUnauthorized   = errors.Register(ModuleName, 2, "unauthorized")
	ErrInvalidParams  = errors.Register(ModuleName, 3, "invalid parameters")
	ErrNotFound       = errors.Register(ModuleName, 4, "not found")
	ErrAlreadyExists  = errors.Register(ModuleName, 5, "already exists")
	ErrInvalidState   = errors.Register(ModuleName, 6, "invalid state")
	ErrCorruptedState = errors.Register(ModuleName, 7, "corrupted state data")

	// Network security errors (10-29)
	ErrRateLimitExceeded      = errors.Register(ModuleName, 10, "rate limit exceeded")
	ErrPeerBlacklisted        = errors.Register(ModuleName, 11, "peer is blacklisted")
	ErrInvalidPeerID          = errors.Register(ModuleName, 12, "invalid peer ID")
	ErrSybilDetected          = errors.Register(ModuleName, 13, "sybil attack detected")
	ErrEclipseDetected        = errors.Register(ModuleName, 14, "eclipse attack detected")
	ErrGossipViolation        = errors.Register(ModuleName, 15, "gossip protocol violation")
	ErrMempoolSpam            = errors.Register(ModuleName, 16, "mempool spam detected")
	ErrForkDetected           = errors.Register(ModuleName, 17, "fork detected")
	ErrPartitionDetected      = errors.Register(ModuleName, 18, "network partition detected")
	ErrMaxConnectionsExceeded = errors.Register(ModuleName, 19, "max connections exceeded")

	// Validator security errors (30-49)
	ErrValidatorNotFound = errors.Register(ModuleName, 30, "validator not found")
	ErrValidatorJailed   = errors.Register(ModuleName, 31, "validator is jailed")
	ErrDoubleSign        = errors.Register(ModuleName, 32, "double sign detected")
	ErrDowntime          = errors.Register(ModuleName, 33, "excessive downtime")
	ErrInvalidSentryNode = errors.Register(ModuleName, 34, "invalid sentry node configuration")
	ErrSlashingFailed    = errors.Register(ModuleName, 35, "slashing operation failed")
	ErrJailingFailed     = errors.Register(ModuleName, 36, "jailing operation failed")
	ErrInsufficientStake = errors.Register(ModuleName, 37, "insufficient stake")

	// Wallet security errors (50-69)
	ErrWalletNotFound         = errors.Register(ModuleName, 50, "wallet not found")
	ErrMultiSigFailed         = errors.Register(ModuleName, 51, "multisig validation failed")
	ErrInsufficientSigs       = errors.Register(ModuleName, 52, "insufficient signatures")
	ErrRecoveryFailed         = errors.Register(ModuleName, 53, "social recovery failed")
	ErrRecoveryNotReady       = errors.Register(ModuleName, 54, "recovery period not elapsed")
	ErrInvalidSession         = errors.Register(ModuleName, 55, "invalid or expired session")
	ErrDeviceNotTrusted       = errors.Register(ModuleName, 56, "device not trusted")
	ErrAnomalyDetected        = errors.Register(ModuleName, 57, "anomalous activity detected")
	ErrHardwareWalletRequired = errors.Register(ModuleName, 58, "hardware wallet required")
	ErrBiometricFailed        = errors.Register(ModuleName, 59, "biometric verification failed")

	// Incident response errors (70-89)
	ErrIncidentNotFound    = errors.Register(ModuleName, 70, "incident not found")
	ErrSystemPaused        = errors.Register(ModuleName, 71, "system is paused")
	ErrCircuitBreakerOpen  = errors.Register(ModuleName, 72, "circuit breaker is open")
	ErrInvalidPauseLevel   = errors.Register(ModuleName, 73, "invalid pause level")
	ErrWalletLimitExceeded = errors.Register(ModuleName, 74, "wallet limit exceeded")
	ErrEmergencyOnly       = errors.Register(ModuleName, 75, "only emergency actions allowed")
	ErrIncidentActive      = errors.Register(ModuleName, 76, "incident is still active")

	// Cryptography errors (90-109)
	ErrKeyNotFound             = errors.Register(ModuleName, 90, "key not found")
	ErrKeyRotationFailed       = errors.Register(ModuleName, 91, "key rotation failed")
	ErrEncryptionFailed        = errors.Register(ModuleName, 92, "encryption failed")
	ErrDecryptionFailed        = errors.Register(ModuleName, 93, "decryption failed")
	ErrInvalidSignature        = errors.Register(ModuleName, 94, "invalid signature")
	ErrThresholdNotMet         = errors.Register(ModuleName, 95, "threshold not met")
	ErrZKProofInvalid          = errors.Register(ModuleName, 96, "zero-knowledge proof invalid")
	ErrEnclaveError            = errors.Register(ModuleName, 97, "secure enclave error")
	ErrRandomGenFailed         = errors.Register(ModuleName, 98, "random generation failed")
	ErrCertificateInvalid      = errors.Register(ModuleName, 99, "certificate invalid or expired")
	ErrKeyRotationNotFound     = errors.Register(ModuleName, 100, "key rotation schedule not found")
	ErrKeyRotationDisabled     = errors.Register(ModuleName, 101, "key rotation is disabled")
	ErrQuantumKeyNotFound      = errors.Register(ModuleName, 102, "quantum-resistant key not found")
	ErrThresholdSchemeNotFound = errors.Register(ModuleName, 103, "threshold scheme not found")

	// Privacy errors (110-129)
	ErrMixingPoolNotFound   = errors.Register(ModuleName, 110, "mixing pool not found")
	ErrInsufficientMixers   = errors.Register(ModuleName, 111, "insufficient mixing participants")
	ErrRingTooSmall         = errors.Register(ModuleName, 112, "ring size too small")
	ErrRingTooLarge         = errors.Register(ModuleName, 113, "ring size too large")
	ErrStealthAddrFailed    = errors.Register(ModuleName, 114, "stealth address generation failed")
	ErrConfidentialTxFailed = errors.Register(ModuleName, 115, "confidential transaction failed")
	ErrViewKeyInvalid       = errors.Register(ModuleName, 116, "view key invalid")
	ErrMixingFeeRequired    = errors.Register(ModuleName, 117, "mixing fee required")

	// Security guard errors (130-149)
	ErrReentrancyDetected = errors.Register(ModuleName, 130, "reentrancy attack detected")
)
