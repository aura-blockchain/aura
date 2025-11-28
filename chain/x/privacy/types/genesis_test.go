package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultGenesis(t *testing.T) {
	gs := DefaultGenesis()
	require.NotNil(t, gs)
	require.NotNil(t, gs.Params)

	// Validate default genesis
	require.NoError(t, gs.Validate())
}

func TestGenesisState_Validate_Valid(t *testing.T) {
	gs := DefaultGenesis()
	err := gs.Validate()
	require.NoError(t, err)
}

func TestGenesisState_Validate_InvalidParams(t *testing.T) {
	gs := DefaultGenesis()
	gs.Params.MinRingSize = 1 // Invalid
	err := gs.Validate()
	require.Error(t, err)
	require.Equal(t, ErrInvalidRingSize, err)
}

func TestGenesisState_CustomParams(t *testing.T) {
	gs := &GenesisState{
		Params: Params{
			EnableZkProofs:                 true,
			EnableStealthAddresses:         false,
			EnableRingSignatures:           true,
			EnableConfidentialTransactions: true,
			EnableNetworkPrivacy:           false,
			EnableMixing:                   true,
			MinRingSize:                    4,
			MaxRingSize:                    20,
			MinMixingParticipants:          8,
			MixingFee:                      "2000",
			ZkProofVerificationCost:        20000,
		},
	}

	err := gs.Validate()
	require.NoError(t, err)
}

func TestGenesisState_MultipleValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*GenesisState)
		wantError error
	}{
		{
			name: "invalid min ring size",
			mutate: func(gs *GenesisState) {
				gs.Params.MinRingSize = 0
			},
			wantError: ErrInvalidRingSize,
		},
		{
			name: "max ring size less than min",
			mutate: func(gs *GenesisState) {
				gs.Params.MinRingSize = 10
				gs.Params.MaxRingSize = 5
			},
			wantError: ErrInvalidRingSize,
		},
		{
			name: "invalid mixing participants",
			mutate: func(gs *GenesisState) {
				gs.Params.MinMixingParticipants = 1
			},
			wantError: ErrInvalidMixingParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := DefaultGenesis()
			tt.mutate(gs)
			err := gs.Validate()
			require.Error(t, err)
			require.Equal(t, tt.wantError, err)
		})
	}
}

func TestGenesisState_DefaultValues(t *testing.T) {
	gs := DefaultGenesis()

	require.True(t, gs.Params.EnableZkProofs)
	require.True(t, gs.Params.EnableStealthAddresses)
	require.True(t, gs.Params.EnableRingSignatures)
	require.True(t, gs.Params.EnableConfidentialTransactions)
	require.True(t, gs.Params.EnableNetworkPrivacy)
	require.True(t, gs.Params.EnableMixing)
	require.Equal(t, uint32(3), gs.Params.MinRingSize)
	require.Equal(t, uint32(16), gs.Params.MaxRingSize)
	require.Equal(t, uint32(5), gs.Params.MinMixingParticipants)
	require.Equal(t, "1000", gs.Params.MixingFee)
	require.Equal(t, uint64(10000), gs.Params.ZkProofVerificationCost)
}
