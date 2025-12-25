// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *cryptoproto.GenesisState {
	params := DefaultParams()
	return &cryptoproto.GenesisState{
		Params:               params,
		KeyRotationSchedules: []*cryptoproto.KeyRotationSchedule{},
		ThresholdSchemes:     []*cryptoproto.ThresholdSignatureScheme{},
		ZkProofConfigs:       []*cryptoproto.ZKProofConfig{},
		SecureEnclaves:       []*cryptoproto.SecureEnclaveConfig{},
		QuantumResistantKeys: []*cryptoproto.QuantumResistantKey{},
		RandomSources:        []*cryptoproto.CryptoRandomSource{},
		KeyStretchingConfigs: []*cryptoproto.KeyStretchingConfig{},
		CertificatePins:      []*cryptoproto.CertificatePin{},
	}
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(data *cryptoproto.GenesisState) error {
	if err := ValidateParams(&data.Params); err != nil {
		return err
	}

	// Validate key rotation schedules
	for _, schedule := range data.KeyRotationSchedules {
		if schedule.KeyId == "" {
			return ErrInvalidKeyID
		}
		if schedule.RotationIntervalSeconds < 3600 {
			return ErrInvalidRotationInterval
		}
	}

	// Validate threshold schemes
	for _, scheme := range data.ThresholdSchemes {
		if scheme.Threshold > scheme.TotalParticipants {
			return ErrInvalidThreshold
		}
		if len(scheme.ParticipantIds) != int(scheme.TotalParticipants) {
			return ErrInvalidParticipantCount
		}
	}

	// Validate ZK proof configs
	for _, config := range data.ZkProofConfigs {
		if config.CircuitId == "" {
			return ErrInvalidZKProof
		}
	}

	// Validate quantum keys
	for _, key := range data.QuantumResistantKeys {
		if key.Algorithm == cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_UNSPECIFIED {
			return ErrInvalidQuantumAlgorithm
		}
	}

	// Validate random sources
	for _, source := range data.RandomSources {
		if source.EntropyBits < int64(data.Params.MinEntropyBits) {
			return ErrInsufficientEntropy
		}
	}

	// Validate certificate pins
	for _, pin := range data.CertificatePins {
		if pin.Hostname == "" {
			return ErrCertificatePinNotFound
		}
		for _, hash := range pin.CertificateHashes {
			if len(hash) != 32 {
				return ErrInvalidCertificateHash
			}
		}
	}

	return nil
}
