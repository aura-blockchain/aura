// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

func TestParams_Validate(t *testing.T) {
	tests := []struct {
		name        string
		modifyParam func(*types.Params)
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid default params",
			modifyParam: func(p *types.Params) {},
			wantErr:     false,
		},
		{
			name: "min confirmations below minimum allowed",
			modifyParam: func(p *types.Params) {
				p.MinConfirmations = 1 // Below MinAllowedConfirmations (2)
			},
			wantErr:     true,
			errContains: "MinConfirmations must be >= 2",
		},
		{
			name: "zero min confirmations",
			modifyParam: func(p *types.Params) {
				p.MinConfirmations = 0
			},
			wantErr:     true,
			errContains: "MinConfirmations must be >= 2",
		},
		{
			name: "excessive bridge fee",
			modifyParam: func(p *types.Params) {
				p.BridgeFeeBasisPoints = 1001 // > 1000 (10%)
			},
			wantErr:     true,
			errContains: "BridgeFeeBasisPoints cannot exceed 1000",
		},
		{
			name: "validator threshold below 51",
			modifyParam: func(p *types.Params) {
				p.ValidatorThresholdPercentage = 50
			},
			wantErr:     true,
			errContains: "ValidatorThresholdPercentage must be between 51-100",
		},
		{
			name: "validator threshold above 100",
			modifyParam: func(p *types.Params) {
				p.ValidatorThresholdPercentage = 101
			},
			wantErr:     true,
			errContains: "ValidatorThresholdPercentage must be between 51-100",
		},
		{
			name: "invalid max transfer amount",
			modifyParam: func(p *types.Params) {
				p.MaxTransferAmount = "not-a-number"
			},
			wantErr:     true,
			errContains: "MaxTransferAmount must be a valid integer",
		},
		{
			name: "invalid auto pause threshold",
			modifyParam: func(p *types.Params) {
				p.AutoPauseThreshold = "invalid"
			},
			wantErr:     true,
			errContains: "AutoPauseThreshold must be a valid integer",
		},
		{
			name: "invalid daily mint limit",
			modifyParam: func(p *types.Params) {
				p.DailyMintLimit = "invalid"
			},
			wantErr:     true,
			errContains: "DailyMintLimit must be a valid integer",
		},
		{
			name: "invalid hourly mint limit",
			modifyParam: func(p *types.Params) {
				p.HourlyMintLimit = "invalid"
			},
			wantErr:     true,
			errContains: "HourlyMintLimit must be a valid integer",
		},
		{
			name: "fraud proof window too short",
			modifyParam: func(p *types.Params) {
				p.FraudProofWindow = 3599 // < 3600 (1 hour)
			},
			wantErr:     true,
			errContains: "fraud proof window must be at least 1 hour",
		},
		{
			name: "fraud proof window too long",
			modifyParam: func(p *types.Params) {
				p.FraudProofWindow = 2592001 // > 2592000 (30 days)
			},
			wantErr:     true,
			errContains: "fraud proof window cannot exceed 30 days",
		},
		{
			name: "invalid slash fraud signature",
			modifyParam: func(p *types.Params) {
				p.SlashFraudSignature = "not-a-decimal"
			},
			wantErr:     true,
			errContains: "invalid SlashFraudSignature",
		},
		{
			name: "negative slash fraud signature",
			modifyParam: func(p *types.Params) {
				p.SlashFraudSignature = "-0.5"
			},
			wantErr:     true,
			errContains: "cannot be negative",
		},
		{
			name: "slash fraud signature above 1",
			modifyParam: func(p *types.Params) {
				p.SlashFraudSignature = "1.5"
			},
			wantErr:     true,
			errContains: "cannot exceed 1.0",
		},
		{
			name: "invalid slash double signing",
			modifyParam: func(p *types.Params) {
				p.SlashDoubleSigning = "invalid"
			},
			wantErr:     true,
			errContains: "invalid SlashDoubleSigning",
		},
		{
			name: "invalid slash offline",
			modifyParam: func(p *types.Params) {
				p.SlashOffline = ""
			},
			wantErr:     true,
			errContains: "invalid SlashOffline",
		},
		{
			name: "invalid min signed per window",
			modifyParam: func(p *types.Params) {
				p.MinSignedPerWindow = "2.0"
			},
			wantErr:     true,
			errContains: "invalid MinSignedPerWindow",
		},
		{
			name: "signing window too short",
			modifyParam: func(p *types.Params) {
				p.MinSigningWindow = 99 // < 100
			},
			wantErr:     true,
			errContains: "signing window must be at least 100 blocks",
		},
		{
			name: "signing window too long",
			modifyParam: func(p *types.Params) {
				p.MinSigningWindow = 100001 // > 100000
			},
			wantErr:     true,
			errContains: "signing window cannot exceed 100,000 blocks",
		},
		{
			name: "supply caps with empty denom",
			modifyParam: func(p *types.Params) {
				p.SupplyCaps = map[string]string{"": "1000"}
			},
			wantErr:     true,
			errContains: "supply cap denom cannot be empty",
		},
		{
			name: "supply caps with invalid cap",
			modifyParam: func(p *types.Params) {
				p.SupplyCaps = map[string]string{"uaura": "not-a-number"}
			},
			wantErr:     true,
			errContains: "invalid supply cap",
		},
		{
			name: "valid supply caps",
			modifyParam: func(p *types.Params) {
				p.SupplyCaps = map[string]string{
					"uaura": "1000000000",
					"wpaw":  "500000000",
				}
			},
			wantErr: false,
		},
		{
			name: "empty max transfer amount is valid",
			modifyParam: func(p *types.Params) {
				p.MaxTransferAmount = ""
			},
			wantErr: false,
		},
		{
			name: "empty auto pause threshold is valid",
			modifyParam: func(p *types.Params) {
				p.AutoPauseThreshold = ""
			},
			wantErr: false,
		},
		{
			name: "empty daily mint limit is valid",
			modifyParam: func(p *types.Params) {
				p.DailyMintLimit = ""
			},
			wantErr: false,
		},
		{
			name: "empty hourly mint limit is valid",
			modifyParam: func(p *types.Params) {
				p.HourlyMintLimit = ""
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.DefaultParams()
			tt.modifyParam(&params)

			err := params.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateMinConfirmations(t *testing.T) {
	params := types.DefaultParams()
	pairs := params.ParamSetPairs()

	var validator func(interface{}) error
	for _, pair := range pairs {
		if string(pair.Key) == "MinConfirmations" {
			validator = pair.ValidatorFn
			break
		}
	}
	require.NotNil(t, validator)

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
		errMsg  string
	}{
		{"valid 2", uint64(2), false, ""},
		{"valid 3", uint64(3), false, ""},
		{"valid 100", uint64(100), false, ""},
		{"invalid 0", uint64(0), true, "MinConfirmations must be >= 2"},
		{"invalid 1", uint64(1), true, "MinConfirmations must be >= 2"},
		{"wrong type string", "3", true, ""},
		{"wrong type int", 3, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateStringNotEmpty(t *testing.T) {
	params := types.DefaultParams()
	pairs := params.ParamSetPairs()

	var validator func(interface{}) error
	for _, pair := range pairs {
		if string(pair.Key) == "MaxTransferAmount" {
			validator = pair.ValidatorFn
			break
		}
	}
	require.NotNil(t, validator)

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"valid string", "1000000", false},
		{"empty string", "", true},
		{"wrong type int", 1000, true},
		{"wrong type nil", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateStringSlice(t *testing.T) {
	params := types.DefaultParams()
	pairs := params.ParamSetPairs()

	var validator func(interface{}) error
	for _, pair := range pairs {
		if string(pair.Key) == "PausedChains" {
			validator = pair.ValidatorFn
			break
		}
	}
	require.NotNil(t, validator)

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"valid empty slice", []string{}, false},
		{"valid slice with values", []string{"paw", "xai"}, false},
		{"wrong type string", "paw,xai", true},
		{"wrong type int slice", []int{1, 2}, true},
		{"wrong type nil", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateFraudProofWindow(t *testing.T) {
	params := types.DefaultParams()
	pairs := params.ParamSetPairs()

	var validator func(interface{}) error
	for _, pair := range pairs {
		if string(pair.Key) == "FraudProofWindow" {
			validator = pair.ValidatorFn
			break
		}
	}
	require.NotNil(t, validator)

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
		errMsg  string
	}{
		{"valid 1 hour", int64(3600), false, ""},
		{"valid 7 days", int64(604800), false, ""},
		{"valid 30 days", int64(2592000), false, ""},
		{"too short", int64(3599), true, "at least 1 hour"},
		{"too long", int64(2592001), true, "cannot exceed 30 days"},
		{"wrong type", "3600", true, ""},
		{"negative", int64(-1), true, "at least 1 hour"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateSlashFraction(t *testing.T) {
	params := types.DefaultParams()
	pairs := params.ParamSetPairs()

	var validator func(interface{}) error
	for _, pair := range pairs {
		if string(pair.Key) == "SlashFraudSignature" {
			validator = pair.ValidatorFn
			break
		}
	}
	require.NotNil(t, validator)

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
		errMsg  string
	}{
		{"valid 0", "0.00", false, ""},
		{"valid 0.5", "0.50", false, ""},
		{"valid 1.0", "1.00", false, ""},
		{"empty", "", true, "cannot be empty"},
		{"negative", "-0.5", true, "cannot be negative"},
		{"above 1", "1.5", true, "cannot exceed 1.0"},
		{"invalid format", "abc", true, "invalid decimal format"},
		{"wrong type int", 50, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateSigningWindow(t *testing.T) {
	params := types.DefaultParams()
	pairs := params.ParamSetPairs()

	var validator func(interface{}) error
	for _, pair := range pairs {
		if string(pair.Key) == "MinSigningWindow" {
			validator = pair.ValidatorFn
			break
		}
	}
	require.NotNil(t, validator)

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
		errMsg  string
	}{
		{"valid 100", int64(100), false, ""},
		{"valid 10000", int64(10000), false, ""},
		{"valid 100000", int64(100000), false, ""},
		{"too short", int64(99), true, "at least 100 blocks"},
		{"too long", int64(100001), true, "cannot exceed 100,000 blocks"},
		{"wrong type", "10000", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateSupplyCaps(t *testing.T) {
	params := types.DefaultParams()
	pairs := params.ParamSetPairs()

	var validator func(interface{}) error
	for _, pair := range pairs {
		if string(pair.Key) == "SupplyCaps" {
			validator = pair.ValidatorFn
			break
		}
	}
	require.NotNil(t, validator)

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
		errMsg  string
	}{
		{"valid empty map", map[string]string{}, false, ""},
		{"valid with entries", map[string]string{"uaura": "1000000"}, false, ""},
		{"empty denom", map[string]string{"": "1000"}, true, "denom cannot be empty"},
		{"invalid cap value", map[string]string{"uaura": "abc"}, true, "invalid supply cap"},
		{"wrong type", "not a map", true, ""},
		{"wrong map type", map[string]int{"uaura": 1000}, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultParams_AllFieldsSet(t *testing.T) {
	params := types.DefaultParams()

	// Verify all critical fields are set
	require.True(t, params.BridgeEnabled)
	require.Equal(t, types.DefaultMinConfirmations, params.MinConfirmations)
	require.Equal(t, uint64(30), params.BridgeFeeBasisPoints)
	require.NotEmpty(t, params.MaxTransferAmount)
	require.Equal(t, uint64(67), params.ValidatorThresholdPercentage)
	require.NotNil(t, params.SupplyCaps)
	require.NotEmpty(t, params.DailyMintLimit)
	require.NotEmpty(t, params.HourlyMintLimit)
	require.False(t, params.Paused)
	require.NotNil(t, params.PausedChains)
	require.False(t, params.AutoPauseEnabled)
	require.NotEmpty(t, params.AutoPauseThreshold)
	require.NotNil(t, params.EmergencyPauseAddresses)
	require.Greater(t, params.FraudProofWindow, int64(0))
	require.NotEmpty(t, params.SlashFraudSignature)
	require.NotEmpty(t, params.SlashDoubleSigning)
	require.NotEmpty(t, params.SlashOffline)
	require.Greater(t, params.MinSigningWindow, int64(0))
	require.NotEmpty(t, params.MinSignedPerWindow)

	// Verify the values make sense
	maxTransfer, ok := sdkmath.NewIntFromString(params.MaxTransferAmount)
	require.True(t, ok)
	require.True(t, maxTransfer.IsPositive())

	dailyLimit, ok := sdkmath.NewIntFromString(params.DailyMintLimit)
	require.True(t, ok)
	require.True(t, dailyLimit.IsPositive())

	hourlyLimit, ok := sdkmath.NewIntFromString(params.HourlyMintLimit)
	require.True(t, ok)
	require.True(t, hourlyLimit.IsPositive())
}

func TestDefaultGenesis_ValidParams(t *testing.T) {
	genesis := types.DefaultGenesis()

	require.NotNil(t, genesis)
	require.NotNil(t, genesis.Params)
	require.True(t, genesis.Params.Enabled)
	require.Equal(t, types.DefaultMinConfirmations, genesis.Params.MinConfirmations)
	require.Equal(t, uint64(30), genesis.Params.BridgeFeeBasisPoints)
	require.False(t, genesis.Params.MaxTransferAmount.IsNil())
	require.True(t, genesis.Params.MaxTransferAmount.IsPositive())
	require.Equal(t, uint64(67), genesis.Params.ValidatorThresholdPercentage)
}

func TestDefaultTimelockAndFraudWindow(t *testing.T) {
	// Verify default constants are sensible
	require.Equal(t, int64(86400), int64(types.DefaultTimelockDuration.Seconds()))  // 24 hours
	require.Equal(t, int64(604800), int64(types.DefaultFraudProofWindow.Seconds())) // 7 days
	require.Equal(t, uint64(3), types.DefaultMinConfirmations)                      // 3 confirmations
	require.Equal(t, uint64(2), types.MinAllowedConfirmations)                      // 2 minimum
}
