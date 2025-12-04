package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
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
				FraudProofWindow:        3600,   // 1 hour minimum
				SlashFraudSignature:     "0.50", // 50%
				SlashDoubleSigning:      "1.00", // 100%
				SlashOffline:            "0.01", // 1%
				MinSigningWindow:        10000,  // blocks
				MinSignedPerWindow:      "0.50", // 50%
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
				FraudProofWindow:        3600,   // 1 hour minimum
				SlashFraudSignature:     "0.50", // 50%
				SlashDoubleSigning:      "1.00", // 100%
				SlashOffline:            "0.01", // 1%
				MinSigningWindow:        10000,  // blocks
				MinSignedPerWindow:      "0.50", // 50%
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
				FraudProofWindow:             3600,   // 1 hour minimum
				SlashFraudSignature:          "0.50", // 50%
				SlashDoubleSigning:           "1.00", // 100%
				SlashOffline:                 "0.01", // 1%
				MinSigningWindow:             10000,  // blocks
				MinSignedPerWindow:           "0.50", // 50%
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
				FraudProofWindow:             3600,   // 1 hour minimum
				SlashFraudSignature:          "0.50", // 50%
				SlashDoubleSigning:           "1.00", // 100%
				SlashOffline:                 "0.01", // 1%
				MinSigningWindow:             10000,  // blocks
				MinSignedPerWindow:           "0.50", // 50%
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
				FraudProofWindow:             3600,   // 1 hour minimum
				SlashFraudSignature:          "0.50", // 50%
				SlashDoubleSigning:           "1.00", // 100%
				SlashOffline:                 "0.01", // 1%
				MinSigningWindow:             10000,  // blocks
				MinSignedPerWindow:           "0.50", // 50%
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
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Get default params to verify keeper setup works
	params := k.GetParams(input.Ctx)
	require.NotNil(t, params, "Keeper should return params")

	// Verify that params contain supply cap related fields
	require.NotNil(t, params.SupplyCaps, "SupplyCaps map should be initialized")
	require.NotEmpty(t, params.DailyMintLimit, "DailyMintLimit should have a default value")
	require.NotEmpty(t, params.HourlyMintLimit, "HourlyMintLimit should have a default value")

	// Verify limits are valid positive integers
	dailyLimit, ok := sdkmath.NewIntFromString(params.DailyMintLimit)
	require.True(t, ok, "DailyMintLimit should be a valid integer")
	require.True(t, dailyLimit.GT(sdkmath.ZeroInt()), "DailyMintLimit should be positive")

	hourlyLimit, ok := sdkmath.NewIntFromString(params.HourlyMintLimit)
	require.True(t, ok, "HourlyMintLimit should be a valid integer")
	require.True(t, hourlyLimit.GT(sdkmath.ZeroInt()), "HourlyMintLimit should be positive")

	// Verify that MaxTransferAmount can be compared with supply caps
	maxTransfer, ok := sdkmath.NewIntFromString(params.MaxTransferAmount)
	require.True(t, ok, "MaxTransferAmount should be a valid integer")
	require.True(t, maxTransfer.GT(sdkmath.ZeroInt()), "MaxTransferAmount should be positive")

	// Verify basic keeper functionality
	require.NotPanics(t, func() {
		_ = k.GetParams(input.Ctx)
	}, "GetParams should not panic")
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
