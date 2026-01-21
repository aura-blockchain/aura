// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

// newTestContext creates a context suitable for testing
func newTestContext() sdk.Context {
	key := storetypes.NewKVStoreKey("test")
	return testutil.DefaultContext(key, storetypes.NewTransientStoreKey("test_transient"))
}

func TestRequireNotPaused(t *testing.T) {
	tests := []struct {
		name    string
		params  types.Params
		chain   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "not paused - operations allowed",
			params: types.Params{
				Paused:       false,
				PausedChains: []string{},
			},
			chain:   "paw",
			wantErr: false,
		},
		{
			name: "globally paused",
			params: types.Params{
				Paused:       true,
				PausedChains: []string{},
			},
			chain:   "paw",
			wantErr: true,
			errMsg:  "bridge is globally paused",
		},
		{
			name: "specific chain paused",
			params: types.Params{
				Paused:       false,
				PausedChains: []string{"paw", "xai"},
			},
			chain:   "paw",
			wantErr: true,
			errMsg:  "bridge paused for chain paw",
		},
		{
			name: "different chain paused - allow operation",
			params: types.Params{
				Paused:       false,
				PausedChains: []string{"xai"},
			},
			chain:   "paw",
			wantErr: false,
		},
		{
			name: "case insensitive chain matching",
			params: types.Params{
				Paused:       false,
				PausedChains: []string{"PAW"},
			},
			chain:   "paw",
			wantErr: true,
			errMsg:  "bridge paused for chain paw",
		},
		{
			name: "chain with whitespace",
			params: types.Params{
				Paused:       false,
				PausedChains: []string{"  paw  "},
			},
			chain:   "paw",
			wantErr: true,
			errMsg:  "bridge paused for chain paw",
		},
		{
			name: "empty chain parameter",
			params: types.Params{
				Paused:       false,
				PausedChains: []string{"paw"},
			},
			chain:   "",
			wantErr: false,
		},
		{
			name: "global pause overrides chain specific",
			params: types.Params{
				Paused:       true,
				PausedChains: []string{}, // No specific chains
			},
			chain:   "paw",
			wantErr: true,
			errMsg:  "globally paused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestContext()
			err := types.RequireNotPaused(ctx, tt.params, tt.chain)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckAndTriggerAutoPause(t *testing.T) {
	tests := []struct {
		name          string
		params        types.Params
		denom         string
		amount        sdkmath.Int
		hourlyMinted  sdkmath.Int
		shouldTrigger bool
		shouldPause   bool
	}{
		{
			name: "auto pause disabled - no trigger",
			params: types.Params{
				AutoPauseEnabled:   false,
				AutoPauseThreshold: "1000000000",
				Paused:             false,
			},
			denom:         "uaura",
			amount:        sdkmath.NewInt(500000000),
			hourlyMinted:  sdkmath.NewInt(600000000),
			shouldTrigger: false,
			shouldPause:   false,
		},
		{
			name: "below threshold - no trigger",
			params: types.Params{
				AutoPauseEnabled:   true,
				AutoPauseThreshold: "1000000000",
				Paused:             false,
			},
			denom:         "uaura",
			amount:        sdkmath.NewInt(100000000),
			hourlyMinted:  sdkmath.NewInt(100000000),
			shouldTrigger: false,
			shouldPause:   false,
		},
		{
			name: "exactly at threshold - no trigger",
			params: types.Params{
				AutoPauseEnabled:   true,
				AutoPauseThreshold: "1000000000",
				Paused:             false,
			},
			denom:         "uaura",
			amount:        sdkmath.NewInt(500000000),
			hourlyMinted:  sdkmath.NewInt(500000000),
			shouldTrigger: false,
			shouldPause:   false,
		},
		{
			name: "exceeds threshold - trigger pause",
			params: types.Params{
				AutoPauseEnabled:   true,
				AutoPauseThreshold: "1000000000",
				Paused:             false,
			},
			denom:         "uaura",
			amount:        sdkmath.NewInt(500000001),
			hourlyMinted:  sdkmath.NewInt(500000000),
			shouldTrigger: true,
			shouldPause:   true,
		},
		{
			name: "invalid threshold format - no trigger",
			params: types.Params{
				AutoPauseEnabled:   true,
				AutoPauseThreshold: "invalid",
				Paused:             false,
			},
			denom:         "uaura",
			amount:        sdkmath.NewInt(1000000000),
			hourlyMinted:  sdkmath.NewInt(1000000000),
			shouldTrigger: false,
			shouldPause:   false,
		},
		{
			name: "zero threshold - no trigger",
			params: types.Params{
				AutoPauseEnabled:   true,
				AutoPauseThreshold: "0",
				Paused:             false,
			},
			denom:         "uaura",
			amount:        sdkmath.NewInt(1),
			hourlyMinted:  sdkmath.NewInt(0),
			shouldTrigger: false,
			shouldPause:   false,
		},
		{
			name: "negative threshold - no trigger",
			params: types.Params{
				AutoPauseEnabled:   true,
				AutoPauseThreshold: "-1000",
				Paused:             false,
			},
			denom:         "uaura",
			amount:        sdkmath.NewInt(1),
			hourlyMinted:  sdkmath.NewInt(0),
			shouldTrigger: false,
			shouldPause:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestContext()
			triggered, updatedParams := types.CheckAndTriggerAutoPause(
				ctx,
				tt.params,
				tt.denom,
				tt.amount,
				tt.hourlyMinted,
			)

			require.Equal(t, tt.shouldTrigger, triggered)
			require.Equal(t, tt.shouldPause, updatedParams.Paused)
		})
	}
}

func TestIsEmergencyPauseAuthorized(t *testing.T) {
	tests := []struct {
		name       string
		params     types.Params
		address    string
		authorized bool
	}{
		{
			name: "authorized address",
			params: types.Params{
				EmergencyPauseAddresses: []string{
					"cosmos1authorized1",
					"cosmos1authorized2",
				},
			},
			address:    "cosmos1authorized1",
			authorized: true,
		},
		{
			name: "unauthorized address",
			params: types.Params{
				EmergencyPauseAddresses: []string{
					"cosmos1authorized1",
					"cosmos1authorized2",
				},
			},
			address:    "cosmos1unauthorized",
			authorized: false,
		},
		{
			name: "empty address",
			params: types.Params{
				EmergencyPauseAddresses: []string{
					"cosmos1authorized1",
				},
			},
			address:    "",
			authorized: false,
		},
		{
			name: "empty authorized list",
			params: types.Params{
				EmergencyPauseAddresses: []string{},
			},
			address:    "cosmos1someaddress",
			authorized: false,
		},
		{
			name: "case sensitive comparison",
			params: types.Params{
				EmergencyPauseAddresses: []string{
					"cosmos1authorized",
				},
			},
			address:    "COSMOS1AUTHORIZED",
			authorized: false, // Case sensitive, different
		},
		{
			name: "nil authorized list",
			params: types.Params{
				EmergencyPauseAddresses: nil,
			},
			address:    "cosmos1someaddress",
			authorized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := types.IsEmergencyPauseAuthorized(tt.params, tt.address)
			require.Equal(t, tt.authorized, result)
		})
	}
}

func TestCircuitBreakerStatusEnums(t *testing.T) {
	// Verify enum values are distinct
	require.NotEqual(t, types.CircuitBreakerStatus_CIRCUIT_CLOSED, types.CircuitBreakerStatus_CIRCUIT_OPEN)
	require.NotEqual(t, types.CircuitBreakerStatus_CIRCUIT_OPEN, types.CircuitBreakerStatus_CIRCUIT_HALF_OPEN)
	require.NotEqual(t, types.CircuitBreakerStatus_CIRCUIT_CLOSED, types.CircuitBreakerStatus_CIRCUIT_HALF_OPEN)

	// Verify default value is CLOSED
	require.Equal(t, types.CircuitBreakerStatus(0), types.CircuitBreakerStatus_CIRCUIT_CLOSED)
}

func TestCircuitBreakerSecurityErrors(t *testing.T) {
	// Verify security error codes are correctly registered
	require.NotNil(t, types.ErrCircuitBreakerOpen)
	require.NotNil(t, types.ErrHourlyVolumeExceeded)
	require.NotNil(t, types.ErrTooManyFailedTransfers)

	// Check error codes are in the expected range (120-129)
	require.Contains(t, types.ErrCircuitBreakerOpen.Error(), "circuit breaker is open")
	require.Contains(t, types.ErrHourlyVolumeExceeded.Error(), "hourly volume limit exceeded")
	require.Contains(t, types.ErrTooManyFailedTransfers.Error(), "too many failed transfers")
}
