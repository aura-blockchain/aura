// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	bridgetypes "github.com/aequitas/aura/chain/x/bridge/types"
	dextypes "github.com/aequitas/aura/chain/x/dex/types"
)

func TestValidateStoreKeyNamesUnique(t *testing.T) {
	tests := []struct {
		name        string
		keys        map[string]*storetypes.KVStoreKey
		expectError bool
	}{
		{
			name: "unique keys",
			keys: map[string]*storetypes.KVStoreKey{
				"auth":    storetypes.NewKVStoreKey(authtypes.StoreKey),
				"bank":    storetypes.NewKVStoreKey(banktypes.StoreKey),
				"staking": storetypes.NewKVStoreKey(stakingtypes.StoreKey),
			},
			expectError: false,
		},
		{
			name: "duplicate key names",
			keys: map[string]*storetypes.KVStoreKey{
				"module1": storetypes.NewKVStoreKey("same_key"),
				"module2": storetypes.NewKVStoreKey("same_key"),
			},
			expectError: true,
		},
		{
			name: "nil key",
			keys: map[string]*storetypes.KVStoreKey{
				"auth": nil,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStoreKeyNamesUnique(tt.keys)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateModuleAccountPermissions(t *testing.T) {
	tests := []struct {
		name        string
		perms       map[string][]string
		expectError bool
	}{
		{
			name: "valid permissions",
			perms: map[string][]string{
				stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
				stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
			},
			expectError: false,
		},
		{
			name: "duplicate permissions in module",
			perms: map[string][]string{
				dextypes.ModuleName: {authtypes.Minter, authtypes.Minter},
			},
			expectError: true,
		},
		{
			name: "unknown permission type",
			perms: map[string][]string{
				bridgetypes.ModuleName: {"unknown_permission"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModuleAccountPermissions(tt.perms)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDetectCircularDependencies(t *testing.T) {
	tests := []struct {
		name         string
		dependencies map[string][]string
		hasCycle     bool
	}{
		{
			name: "no cycle",
			dependencies: map[string][]string{
				"bank":    {"auth"},
				"staking": {"auth", "bank"},
			},
			hasCycle: false,
		},
		{
			name: "simple cycle",
			dependencies: map[string][]string{
				"a": {"b"},
				"b": {"a"},
			},
			hasCycle: true,
		},
		{
			name: "complex cycle",
			dependencies: map[string][]string{
				"a": {"b"},
				"b": {"c"},
				"c": {"a"},
			},
			hasCycle: true,
		},
		{
			name: "no dependencies",
			dependencies: map[string][]string{
				"a": {},
				"b": {},
			},
			hasCycle: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasCycle := detectCircularDependencies(tt.dependencies)
			require.Equal(t, tt.hasCycle, hasCycle)
		})
	}
}

func TestValidateModuleInitializationOrder(t *testing.T) {
	dependencies := map[string][]string{
		"bank":    {"auth"},
		"staking": {"auth", "bank"},
		"bridge":  {"bank", "vcregistry"},
	}

	tests := []struct {
		name        string
		order       []string
		expectError bool
	}{
		{
			name:        "correct order",
			order:       []string{"auth", "bank", "vcregistry", "staking", "bridge"},
			expectError: false,
		},
		{
			name:        "incorrect order - bank before auth",
			order:       []string{"bank", "auth", "staking"},
			expectError: true,
		},
		{
			name:        "missing dependency",
			order:       []string{"bank", "staking"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModuleInitializationOrder(tt.order, dependencies)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetRecommendedModuleInitOrder(t *testing.T) {
	order := GetRecommendedModuleInitOrder()

	// Verify order is not empty
	require.NotEmpty(t, order)

	// Verify core modules come first
	authPos := findPosition(order, "auth")
	bankPos := findPosition(order, "bank")
	stakingPos := findPosition(order, "staking")

	require.NotEqual(t, -1, authPos, "auth should be in order")
	require.NotEqual(t, -1, bankPos, "bank should be in order")
	require.NotEqual(t, -1, stakingPos, "staking should be in order")

	// Bank depends on auth, so auth should come before bank
	require.Less(t, authPos, bankPos, "auth should come before bank")

	// Staking depends on bank, so bank should come before staking
	require.Less(t, bankPos, stakingPos, "bank should come before staking")

	// WASM should be near the end (depends on many modules)
	wasmPos := findPosition(order, "wasm")
	if wasmPos != -1 {
		require.Greater(t, wasmPos, stakingPos, "wasm should come after core modules")
	}
}

// Helper function to find position of a module in the initialization order
func findPosition(order []string, module string) int {
	for i, m := range order {
		if m == module {
			return i
		}
	}
	return -1
}
