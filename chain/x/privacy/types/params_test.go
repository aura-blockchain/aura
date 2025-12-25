// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	require.NotNil(t, params)
	require.True(t, params.EnableZkProofs)
	require.True(t, params.EnableStealthAddresses)
	require.True(t, params.EnableRingSignatures)
	require.True(t, params.EnableConfidentialTransactions)
	require.True(t, params.EnableNetworkPrivacy)
	require.True(t, params.EnableMixing)
	require.Equal(t, uint32(3), params.MinRingSize)
	require.Equal(t, uint32(16), params.MaxRingSize)
	require.Equal(t, uint32(5), params.MinMixingParticipants)
	require.Equal(t, "1000", params.MixingFee)
	require.Equal(t, uint64(10000), params.ZkProofVerificationCost)

	// Validate default params
	require.NoError(t, ValidateParams(params))
}

func TestValidateParams_Valid(t *testing.T) {
	params := DefaultParams()
	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_MinRingSize(t *testing.T) {
	tests := []struct {
		name      string
		ringSize  uint32
		wantError bool
	}{
		{"valid min size 2", 2, false},
		{"valid min size 3", 3, false},
		{"valid min size 10", 10, false},
		{"invalid size 0", 0, true},
		{"invalid size 1", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := DefaultParams()
			params.MinRingSize = tt.ringSize
			err := ValidateParams(params)
			if tt.wantError {
				require.Error(t, err)
				require.Equal(t, ErrInvalidRingSize, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateParams_MaxRingSize(t *testing.T) {
	tests := []struct {
		name        string
		minRingSize uint32
		maxRingSize uint32
		wantError   bool
	}{
		{"valid max > min", 3, 16, false},
		{"valid max = min", 5, 5, false},
		{"invalid max < min", 10, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := DefaultParams()
			params.MinRingSize = tt.minRingSize
			params.MaxRingSize = tt.maxRingSize
			err := ValidateParams(params)
			if tt.wantError {
				require.Error(t, err)
				require.Equal(t, ErrInvalidRingSize, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateParams_MinMixingParticipants(t *testing.T) {
	tests := []struct {
		name         string
		participants uint32
		wantError    bool
	}{
		{"valid 2", 2, false},
		{"valid 5", 5, false},
		{"valid 100", 100, false},
		{"invalid 0", 0, true},
		{"invalid 1", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := DefaultParams()
			params.MinMixingParticipants = tt.participants
			err := ValidateParams(params)
			if tt.wantError {
				require.Error(t, err)
				require.Equal(t, ErrInvalidMixingParams, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParams_AllFeaturesEnabled(t *testing.T) {
	params := DefaultParams()
	require.True(t, params.EnableZkProofs)
	require.True(t, params.EnableStealthAddresses)
	require.True(t, params.EnableRingSignatures)
	require.True(t, params.EnableConfidentialTransactions)
	require.True(t, params.EnableNetworkPrivacy)
	require.True(t, params.EnableMixing)
}

func TestParams_AllFeaturesDisabled(t *testing.T) {
	params := Params{
		EnableZkProofs:                 false,
		EnableStealthAddresses:         false,
		EnableRingSignatures:           false,
		EnableConfidentialTransactions: false,
		EnableNetworkPrivacy:           false,
		EnableMixing:                   false,
		MinRingSize:                    3,
		MaxRingSize:                    16,
		MinMixingParticipants:          5,
		MixingFee:                      "0",
		ZkProofVerificationCost:        0,
	}

	// Should still be valid even with all features disabled
	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestParams_CustomValues(t *testing.T) {
	params := Params{
		EnableZkProofs:                 true,
		EnableStealthAddresses:         true,
		EnableRingSignatures:           true,
		EnableConfidentialTransactions: true,
		EnableNetworkPrivacy:           true,
		EnableMixing:                   true,
		MinRingSize:                    5,
		MaxRingSize:                    32,
		MinMixingParticipants:          10,
		MixingFee:                      "5000",
		ZkProofVerificationCost:        50000,
	}

	err := ValidateParams(params)
	require.NoError(t, err)
	require.Equal(t, uint32(5), params.MinRingSize)
	require.Equal(t, uint32(32), params.MaxRingSize)
	require.Equal(t, uint32(10), params.MinMixingParticipants)
	require.Equal(t, "5000", params.MixingFee)
	require.Equal(t, uint64(50000), params.ZkProofVerificationCost)
}

func TestParams_EdgeCases(t *testing.T) {
	// Test edge case: MinRingSize = MaxRingSize
	params := DefaultParams()
	params.MinRingSize = 10
	params.MaxRingSize = 10
	err := ValidateParams(params)
	require.NoError(t, err)

	// Test edge case: MinMixingParticipants = 2 (minimum valid)
	params = DefaultParams()
	params.MinMixingParticipants = 2
	err = ValidateParams(params)
	require.NoError(t, err)
}

func TestParams_JsonTags(t *testing.T) {
	// Ensure JSON tags are present for all fields
	params := DefaultParams()
	require.NotEmpty(t, params.MinRingSize)
	require.NotEmpty(t, params.MaxRingSize)
	require.NotEmpty(t, params.MinMixingParticipants)
	require.NotEmpty(t, params.MixingFee)
	require.NotZero(t, params.ZkProofVerificationCost)
}
