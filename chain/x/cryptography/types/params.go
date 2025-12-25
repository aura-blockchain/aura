// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// Type aliases for proto types
type (
	KeyRotationSchedule       = cryptoproto.KeyRotationSchedule
	ThresholdSignatureScheme  = cryptoproto.ThresholdSignatureScheme
	ZKProofConfig             = cryptoproto.ZKProofConfig
	SecureEnclaveConfig       = cryptoproto.SecureEnclaveConfig
	QuantumResistantKey       = cryptoproto.QuantumResistantKey
	CertificatePin            = cryptoproto.CertificatePin
	KeyStretchingConfig       = cryptoproto.KeyStretchingConfig
	CryptoRandomSource        = cryptoproto.CryptoRandomSource
)

// NewParams creates a new Params instance with default values
func NewParams() cryptoproto.Params {
	return cryptoproto.Params{
		DefaultRotationIntervalDays: 90,
		EnableAutoRotation:          true,
		MinThresholdParticipants:    2,
		MaxThresholdParticipants:    100,
		MinEntropyBits:              256,
		MinPbkdf2Iterations:         100000,
		MinArgon2MemoryKb:           65536, // 64 MB
		MinArgon2Iterations:         3,
		EnforceCertificatePinning:   true,
		CertificatePinValidityDays:  365,
		MinSaltLengthBytes:          16,
		MinKeyLengthBits:            256,
	}
}

// DefaultParams returns default module parameters
func DefaultParams() cryptoproto.Params {
	return NewParams()
}

// ValidateParams validates module parameters
func ValidateParams(params *cryptoproto.Params) error {
	if params.DefaultRotationIntervalDays < 1 {
		return ErrInvalidRotationInterval
	}
	if params.MinThresholdParticipants < 2 {
		return ErrInvalidThreshold
	}
	if params.MaxThresholdParticipants < params.MinThresholdParticipants {
		return ErrInvalidParticipantCount
	}
	if params.MinEntropyBits < 128 {
		return ErrInsufficientEntropy
	}
	if params.MinPbkdf2Iterations < 10000 {
		return ErrInvalidIterationCount
	}
	if params.MinArgon2MemoryKb < 8192 {
		return ErrInvalidKeyStretchingConfig
	}
	if params.MinArgon2Iterations < 1 {
		return ErrInvalidIterationCount
	}
	if params.MinSaltLengthBytes < 8 {
		return ErrInvalidSaltLength
	}
	if params.MinKeyLengthBits < 128 {
		return ErrInvalidKeyStretchingConfig
	}
	return nil
}
