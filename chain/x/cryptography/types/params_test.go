package types

import (
	"testing"

	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
	"github.com/stretchr/testify/require"
)

func TestNewParams(t *testing.T) {
	params := NewParams()

	require.Equal(t, int32(90), params.DefaultRotationIntervalDays)
	require.True(t, params.EnableAutoRotation)
	require.Equal(t, int32(2), params.MinThresholdParticipants)
	require.Equal(t, int32(100), params.MaxThresholdParticipants)
	require.Equal(t, int32(256), params.MinEntropyBits)
	require.Equal(t, int32(100000), params.MinPbkdf2Iterations)
	require.Equal(t, int32(65536), params.MinArgon2MemoryKb)
	require.Equal(t, int32(3), params.MinArgon2Iterations)
	require.True(t, params.EnforceCertificatePinning)
	require.Equal(t, int32(365), params.CertificatePinValidityDays)
	require.Equal(t, int32(16), params.MinSaltLengthBytes)
	require.Equal(t, int32(256), params.MinKeyLengthBits)
}

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	require.NotNil(t, params)
	require.Equal(t, int32(90), params.DefaultRotationIntervalDays)
	require.True(t, params.EnableAutoRotation)
}

func TestValidateParams(t *testing.T) {
	defaultParams := DefaultParams()
	tests := []struct {
		name    string
		params  *cryptoproto.Params
		wantErr bool
		errType error
	}{
		{
			name:    "valid params",
			params:  &defaultParams,
			wantErr: false,
		},
		{
			name: "invalid rotation interval",
			params: &cryptoproto.Params{
				DefaultRotationIntervalDays: 0,
				MinThresholdParticipants:    2,
				MaxThresholdParticipants:    100,
				MinEntropyBits:              256,
				MinPbkdf2Iterations:         100000,
				MinArgon2MemoryKb:           65536,
				MinArgon2Iterations:         3,
				MinSaltLengthBytes:          16,
				MinKeyLengthBits:            256,
			},
			wantErr: true,
			errType: ErrInvalidRotationInterval,
		},
		{
			name: "invalid min threshold - too low",
			params: &cryptoproto.Params{
				DefaultRotationIntervalDays: 90,
				MinThresholdParticipants:    1,
				MaxThresholdParticipants:    100,
				MinEntropyBits:              256,
				MinPbkdf2Iterations:         100000,
				MinArgon2MemoryKb:           65536,
				MinArgon2Iterations:         3,
				MinSaltLengthBytes:          16,
				MinKeyLengthBits:            256,
			},
			wantErr: true,
			errType: ErrInvalidThreshold,
		},
		{
			name: "invalid participant count - max less than min",
			params: &cryptoproto.Params{
				DefaultRotationIntervalDays: 90,
				MinThresholdParticipants:    10,
				MaxThresholdParticipants:    5,
				MinEntropyBits:              256,
				MinPbkdf2Iterations:         100000,
				MinArgon2MemoryKb:           65536,
				MinArgon2Iterations:         3,
				MinSaltLengthBytes:          16,
				MinKeyLengthBits:            256,
			},
			wantErr: true,
			errType: ErrInvalidParticipantCount,
		},
		{
			name: "insufficient entropy",
			params: &cryptoproto.Params{
				DefaultRotationIntervalDays: 90,
				MinThresholdParticipants:    2,
				MaxThresholdParticipants:    100,
				MinEntropyBits:              64,
				MinPbkdf2Iterations:         100000,
				MinArgon2MemoryKb:           65536,
				MinArgon2Iterations:         3,
				MinSaltLengthBytes:          16,
				MinKeyLengthBits:            256,
			},
			wantErr: true,
			errType: ErrInsufficientEntropy,
		},
		{
			name: "invalid PBKDF2 iteration count",
			params: &cryptoproto.Params{
				DefaultRotationIntervalDays: 90,
				MinThresholdParticipants:    2,
				MaxThresholdParticipants:    100,
				MinEntropyBits:              256,
				MinPbkdf2Iterations:         5000,
				MinArgon2MemoryKb:           65536,
				MinArgon2Iterations:         3,
				MinSaltLengthBytes:          16,
				MinKeyLengthBits:            256,
			},
			wantErr: true,
			errType: ErrInvalidIterationCount,
		},
		{
			name: "invalid Argon2 memory",
			params: &cryptoproto.Params{
				DefaultRotationIntervalDays: 90,
				MinThresholdParticipants:    2,
				MaxThresholdParticipants:    100,
				MinEntropyBits:              256,
				MinPbkdf2Iterations:         100000,
				MinArgon2MemoryKb:           1024,
				MinArgon2Iterations:         3,
				MinSaltLengthBytes:          16,
				MinKeyLengthBits:            256,
			},
			wantErr: true,
			errType: ErrInvalidKeyStretchingConfig,
		},
		{
			name: "invalid Argon2 iterations",
			params: &cryptoproto.Params{
				DefaultRotationIntervalDays: 90,
				MinThresholdParticipants:    2,
				MaxThresholdParticipants:    100,
				MinEntropyBits:              256,
				MinPbkdf2Iterations:         100000,
				MinArgon2MemoryKb:           65536,
				MinArgon2Iterations:         0,
				MinSaltLengthBytes:          16,
				MinKeyLengthBits:            256,
			},
			wantErr: true,
			errType: ErrInvalidIterationCount,
		},
		{
			name: "invalid salt length",
			params: &cryptoproto.Params{
				DefaultRotationIntervalDays: 90,
				MinThresholdParticipants:    2,
				MaxThresholdParticipants:    100,
				MinEntropyBits:              256,
				MinPbkdf2Iterations:         100000,
				MinArgon2MemoryKb:           65536,
				MinArgon2Iterations:         3,
				MinSaltLengthBytes:          4,
				MinKeyLengthBits:            256,
			},
			wantErr: true,
			errType: ErrInvalidSaltLength,
		},
		{
			name: "invalid key length",
			params: &cryptoproto.Params{
				DefaultRotationIntervalDays: 90,
				MinThresholdParticipants:    2,
				MaxThresholdParticipants:    100,
				MinEntropyBits:              256,
				MinPbkdf2Iterations:         100000,
				MinArgon2MemoryKb:           65536,
				MinArgon2Iterations:         3,
				MinSaltLengthBytes:          16,
				MinKeyLengthBits:            64,
			},
			wantErr: true,
			errType: ErrInvalidKeyStretchingConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateParams(tt.params)
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
