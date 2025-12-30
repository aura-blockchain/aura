// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	pb "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesisState(t *testing.T) {
	genesis := DefaultGenesisState()

	require.NotNil(t, genesis)
	require.NotNil(t, genesis.Params)

	// Check default params
	require.True(t, genesis.Params.HardwareWalletEnabled)
	require.Equal(t, int32(5), genesis.Params.MaxSigners)
	require.Equal(t, int32(2), genesis.Params.MinThreshold)
	require.Equal(t, int32(5), genesis.Params.MaxThreshold)
	require.True(t, genesis.Params.SocialRecoveryEnabled)
	require.Equal(t, int32(5), genesis.Params.MaxGuardians)
	require.Equal(t, int32(2), genesis.Params.MinRecoveryThreshold)
	require.Equal(t, uint64(86400), genesis.Params.DefaultRecoveryDelaySeconds)
	require.Equal(t, uint64(3600), genesis.Params.DefaultSessionTimeoutSeconds)
	require.Equal(t, uint64(86400), genesis.Params.MaxSessionDurationSeconds)
	require.True(t, genesis.Params.SpendingLimitsEnabled)
	require.Equal(t, "1000000000", genesis.Params.DefaultDailyLimit)
	require.True(t, genesis.Params.BiometricEnabled)
	require.Equal(t, int32(5), genesis.Params.MaxBiometricAttempts)
	require.Equal(t, uint64(300), genesis.Params.LockoutDurationSeconds)
	require.True(t, genesis.Params.DustFilterEnabled)
	require.Equal(t, "1000", genesis.Params.MinDustAmount)
	require.True(t, genesis.Params.PhishingProtectionEnabled)
	require.True(t, genesis.Params.RequireDomainVerification)

	// Check empty slices are initialized
	require.NotNil(t, genesis.HardwareWallets)
	require.NotNil(t, genesis.MultisigWallets)
	require.NotNil(t, genesis.PendingTransactions)
	require.NotNil(t, genesis.RecoveryConfigs)
	require.NotNil(t, genesis.RecoveryRequests)
	require.NotNil(t, genesis.DomainVerifications)
	require.NotNil(t, genesis.PhishingConfigs)
	require.NotNil(t, genesis.SpendingLimits)
	require.NotNil(t, genesis.SessionConfigs)
	require.NotNil(t, genesis.BiometricConfigs)
	require.NotNil(t, genesis.EnclaveConfigs)
	require.NotNil(t, genesis.EncryptedBackups)
	require.NotNil(t, genesis.DustFilters)
	require.NotNil(t, genesis.DustTransactions)
	require.NotNil(t, genesis.SecurityMetrics)
}

func TestDefaultGenesis(t *testing.T) {
	genesis := DefaultGenesis()
	require.NotNil(t, genesis)

	// Should be equivalent to DefaultGenesisState
	defaultState := DefaultGenesisState()
	require.Equal(t, defaultState.Params.MaxSigners, genesis.Params.MaxSigners)
	require.Equal(t, defaultState.Params.MinThreshold, genesis.Params.MinThreshold)
}

func TestValidateGenesis(t *testing.T) {
	tests := []struct {
		name    string
		genesis *GenesisState
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil genesis",
			genesis: nil,
			wantErr: true,
			errMsg:  "genesis state cannot be nil",
		},
		{
			name: "valid genesis",
			genesis: &GenesisState{
				Params: pb.WalletSecurityParams{
					MinThreshold: 2,
					MaxThreshold: 5,
					MaxSigners:   5,
				},
			},
			wantErr: false,
		},
		{
			name: "min threshold less than 1",
			genesis: &GenesisState{
				Params: pb.WalletSecurityParams{
					MinThreshold: 0,
					MaxThreshold: 5,
					MaxSigners:   5,
				},
			},
			wantErr: true,
			errMsg:  "invalid multisig thresholds",
		},
		{
			name: "max threshold less than min threshold",
			genesis: &GenesisState{
				Params: pb.WalletSecurityParams{
					MinThreshold: 5,
					MaxThreshold: 2,
					MaxSigners:   5,
				},
			},
			wantErr: true,
			errMsg:  "invalid multisig thresholds",
		},
		{
			name: "max signers less than min threshold",
			genesis: &GenesisState{
				Params: pb.WalletSecurityParams{
					MinThreshold: 5,
					MaxThreshold: 10,
					MaxSigners:   3, // Less than min_threshold
				},
			},
			wantErr: true,
			errMsg:  "max_signers must be >= min_threshold",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateGenesis(tc.genesis)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultGenesisValidates(t *testing.T) {
	// Default genesis should always be valid
	genesis := DefaultGenesisState()
	err := ValidateGenesis(genesis)
	require.NoError(t, err)
}
