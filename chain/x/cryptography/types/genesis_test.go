// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesis(t *testing.T) {
	genesis := DefaultGenesis()

	require.NotNil(t, genesis)
	require.NotNil(t, genesis.Params)
	require.NotNil(t, genesis.KeyRotationSchedules)
	require.NotNil(t, genesis.ThresholdSchemes)
	require.NotNil(t, genesis.ZkProofConfigs)
	require.NotNil(t, genesis.SecureEnclaves)
	require.NotNil(t, genesis.QuantumResistantKeys)
	require.NotNil(t, genesis.RandomSources)
	require.NotNil(t, genesis.KeyStretchingConfigs)
	require.NotNil(t, genesis.CertificatePins)

	// Verify default params
	require.Equal(t, int32(90), genesis.Params.DefaultRotationIntervalDays)
	require.True(t, genesis.Params.EnableAutoRotation)
}

func TestValidateGenesis_Valid(t *testing.T) {
	genesis := DefaultGenesis()
	err := ValidateGenesis(genesis)
	require.NoError(t, err)
}

func TestValidateGenesis_InvalidParams(t *testing.T) {
	genesis := DefaultGenesis()
	genesis.Params.MinEntropyBits = 64 // Too low

	err := ValidateGenesis(genesis)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInsufficientEntropy)
}

func TestValidateGenesis_KeyRotationSchedule(t *testing.T) {
	tests := []struct {
		name     string
		schedule *cryptoproto.KeyRotationSchedule
		wantErr  bool
		errType  error
	}{
		{
			name: "valid schedule",
			schedule: &cryptoproto.KeyRotationSchedule{
				KeyId:                   "key-1",
				RotationIntervalSeconds: 7200,
			},
			wantErr: false,
		},
		{
			name: "empty key ID",
			schedule: &cryptoproto.KeyRotationSchedule{
				KeyId:                   "",
				RotationIntervalSeconds: 7200,
			},
			wantErr: true,
			errType: ErrInvalidKeyID,
		},
		{
			name: "invalid interval - too short",
			schedule: &cryptoproto.KeyRotationSchedule{
				KeyId:                   "key-1",
				RotationIntervalSeconds: 1800, // Less than 1 hour
			},
			wantErr: true,
			errType: ErrInvalidRotationInterval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := DefaultGenesis()
			genesis.KeyRotationSchedules = []*cryptoproto.KeyRotationSchedule{tt.schedule}

			err := ValidateGenesis(genesis)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					require.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateGenesis_ThresholdScheme(t *testing.T) {
	tests := []struct {
		name    string
		scheme  *cryptoproto.ThresholdSignatureScheme
		wantErr bool
		errType error
	}{
		{
			name: "valid scheme",
			scheme: &cryptoproto.ThresholdSignatureScheme{
				SchemeId:          "scheme-1",
				Threshold:         3,
				TotalParticipants: 5,
				ParticipantIds:    []string{"p1", "p2", "p3", "p4", "p5"},
			},
			wantErr: false,
		},
		{
			name: "threshold exceeds total",
			scheme: &cryptoproto.ThresholdSignatureScheme{
				SchemeId:          "scheme-1",
				Threshold:         6,
				TotalParticipants: 5,
				ParticipantIds:    []string{"p1", "p2", "p3", "p4", "p5"},
			},
			wantErr: true,
			errType: ErrInvalidThreshold,
		},
		{
			name: "participant count mismatch",
			scheme: &cryptoproto.ThresholdSignatureScheme{
				SchemeId:          "scheme-1",
				Threshold:         3,
				TotalParticipants: 5,
				ParticipantIds:    []string{"p1", "p2", "p3"}, // Only 3 instead of 5
			},
			wantErr: true,
			errType: ErrInvalidParticipantCount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := DefaultGenesis()
			genesis.ThresholdSchemes = []*cryptoproto.ThresholdSignatureScheme{tt.scheme}

			err := ValidateGenesis(genesis)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					require.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateGenesis_ZKProofConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *cryptoproto.ZKProofConfig
		wantErr bool
		errType error
	}{
		{
			name: "valid config",
			config: &cryptoproto.ZKProofConfig{
				CircuitId: "circuit-1",
			},
			wantErr: false,
		},
		{
			name: "empty circuit ID",
			config: &cryptoproto.ZKProofConfig{
				CircuitId: "",
			},
			wantErr: true,
			errType: ErrInvalidZKProof,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := DefaultGenesis()
			genesis.ZkProofConfigs = []*cryptoproto.ZKProofConfig{tt.config}

			err := ValidateGenesis(genesis)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					require.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateGenesis_QuantumResistantKey(t *testing.T) {
	tests := []struct {
		name    string
		key     *cryptoproto.QuantumResistantKey
		wantErr bool
		errType error
	}{
		{
			name: "valid key",
			key: &cryptoproto.QuantumResistantKey{
				KeyId:     "qr-key-1",
				Algorithm: cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM,
			},
			wantErr: false,
		},
		{
			name: "unspecified algorithm",
			key: &cryptoproto.QuantumResistantKey{
				KeyId:     "qr-key-1",
				Algorithm: cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_UNSPECIFIED,
			},
			wantErr: true,
			errType: ErrInvalidQuantumAlgorithm,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := DefaultGenesis()
			genesis.QuantumResistantKeys = []*cryptoproto.QuantumResistantKey{tt.key}

			err := ValidateGenesis(genesis)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					require.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateGenesis_RandomSource(t *testing.T) {
	genesis := DefaultGenesis()

	tests := []struct {
		name    string
		source  *cryptoproto.CryptoRandomSource
		wantErr bool
		errType error
	}{
		{
			name: "valid source",
			source: &cryptoproto.CryptoRandomSource{
				SourceId:    "source-1",
				EntropyBits: 256,
			},
			wantErr: false,
		},
		{
			name: "insufficient entropy",
			source: &cryptoproto.CryptoRandomSource{
				SourceId:    "source-1",
				EntropyBits: 128, // Less than min (256)
			},
			wantErr: true,
			errType: ErrInsufficientEntropy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis.RandomSources = []*cryptoproto.CryptoRandomSource{tt.source}

			err := ValidateGenesis(genesis)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					require.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateGenesis_CertificatePin(t *testing.T) {
	tests := []struct {
		name    string
		pin     *cryptoproto.CertificatePin
		wantErr bool
		errType error
	}{
		{
			name: "valid pin",
			pin: &cryptoproto.CertificatePin{
				Hostname:          "example.com",
				CertificateHashes: [][]byte{make([]byte, 32)},
			},
			wantErr: false,
		},
		{
			name: "empty hostname",
			pin: &cryptoproto.CertificatePin{
				Hostname:          "",
				CertificateHashes: [][]byte{make([]byte, 32)},
			},
			wantErr: true,
			errType: ErrCertificatePinNotFound,
		},
		{
			name: "invalid hash length",
			pin: &cryptoproto.CertificatePin{
				Hostname:          "example.com",
				CertificateHashes: [][]byte{make([]byte, 16)}, // Wrong length
			},
			wantErr: true,
			errType: ErrInvalidCertificateHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := DefaultGenesis()
			genesis.CertificatePins = []*cryptoproto.CertificatePin{tt.pin}

			err := ValidateGenesis(genesis)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					require.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
