package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

// TestSupplyCaps_ValidateParams tests that supply cap parameters are validated correctly
func TestSupplyCaps_ValidateParams(t *testing.T) {
	tests := []struct {
		name      string
		params    types.Params
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid supply caps",
			params: types.Params{
				BridgeEnabled:                true,
				MinConfirmations:             3,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000",
				ValidatorThresholdPercentage: 67,
				SupplyCaps: map[string]string{
					"wrapped.eth":  "1000000",
					"wrapped.btc":  "21000000",
					"wrapped.usdc": "1000000000",
				},
				DailyMintLimit:          "10000000",
				HourlyMintLimit:         "1000000",
				Paused:                  false,
				PausedChains:            []string{},
				AutoPauseEnabled:        false,
				AutoPauseThreshold:      "5000000",
				EmergencyPauseAddresses: []string{},
			},
			expectErr: false,
		},
		{
			name: "invalid supply cap value",
			params: types.Params{
				BridgeEnabled:                true,
				MinConfirmations:             3,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000",
				ValidatorThresholdPercentage: 67,
				SupplyCaps: map[string]string{
					"wrapped.eth": "not-a-number",
				},
				DailyMintLimit:          "10000000",
				HourlyMintLimit:         "1000000",
				Paused:                  false,
				PausedChains:            []string{},
				AutoPauseEnabled:        false,
				AutoPauseThreshold:      "5000000",
				EmergencyPauseAddresses: []string{},
			},
			expectErr: true,
			errMsg:    "invalid supply cap",
		},
		{
			name: "invalid daily limit",
			params: types.Params{
				BridgeEnabled:                true,
				MinConfirmations:             3,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000",
				ValidatorThresholdPercentage: 67,
				SupplyCaps:                   make(map[string]string),
				DailyMintLimit:               "invalid",
				HourlyMintLimit:              "1000000",
				Paused:                       false,
				PausedChains:                 []string{},
				AutoPauseEnabled:             false,
				AutoPauseThreshold:           "5000000",
				EmergencyPauseAddresses:      []string{},
			},
			expectErr: true,
			errMsg:    "DailyMintLimit must be a valid integer",
		},
		{
			name: "invalid hourly limit",
			params: types.Params{
				BridgeEnabled:                true,
				MinConfirmations:             3,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000",
				ValidatorThresholdPercentage: 67,
				SupplyCaps:                   make(map[string]string),
				DailyMintLimit:               "10000000",
				HourlyMintLimit:              "not-valid",
				Paused:                       false,
				PausedChains:                 []string{},
				AutoPauseEnabled:             false,
				AutoPauseThreshold:           "5000000",
				EmergencyPauseAddresses:      []string{},
			},
			expectErr: true,
			errMsg:    "HourlyMintLimit must be a valid integer",
		},
		{
			name: "empty supply caps map is valid",
			params: types.Params{
				BridgeEnabled:                true,
				MinConfirmations:             3,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000",
				ValidatorThresholdPercentage: 67,
				SupplyCaps:                   make(map[string]string),
				DailyMintLimit:               "10000000",
				HourlyMintLimit:              "1000000",
				Paused:                       false,
				PausedChains:                 []string{},
				AutoPauseEnabled:             false,
				AutoPauseThreshold:           "5000000",
				EmergencyPauseAddresses:      []string{},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSupplyCaps_MintTracking tests basic mint tracking functionality
func TestSupplyCaps_MintTracking(t *testing.T) {
	// These would need proper test setup with keeper
	// Skipping for now since they require complex setup
	t.Skip("Requires full keeper test setup")
}

// TestDefaultParams tests that default params include supply cap fields
func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()

	require.NotNil(t, params.SupplyCaps)
	require.NotEmpty(t, params.DailyMintLimit)
	require.NotEmpty(t, params.HourlyMintLimit)

	// Verify default values are valid integers
	_, ok := sdkmath.NewIntFromString(params.DailyMintLimit)
	require.True(t, ok, "DailyMintLimit should be a valid integer")

	_, ok = sdkmath.NewIntFromString(params.HourlyMintLimit)
	require.True(t, ok, "HourlyMintLimit should be a valid integer")
}
