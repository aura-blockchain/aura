// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package wasm_test

import (
	"testing"

	"github.com/aequitas/aura/chain/x/wasm/types"
	"github.com/stretchr/testify/require"
)

// TestWASMModuleTypes verifies that WASM module types are correctly defined
func TestWASMModuleTypes(t *testing.T) {
	// Verify store key is defined
	require.NotEmpty(t, types.StoreKey, "StoreKey should be defined")
	require.Equal(t, "aura_wasm_security", types.StoreKey)

	// Verify module name
	require.NotEmpty(t, types.ModuleName, "ModuleName should be defined")
	require.Equal(t, "aura_wasm_security", types.ModuleName)
}

// TestWASMParamsDefaults verifies default parameters are valid
func TestWASMParamsDefaults(t *testing.T) {
	params := types.DefaultParams()

	// Verify code size limits are set
	require.Greater(t, params.MaxWasmCodeSize, uint64(0), "Max WASM code size should be positive")
	require.Equal(t, uint64(600*1024), params.MaxWasmCodeSize, "Default max code size should be 600KB")

	// Verify gas limits are set
	require.Greater(t, params.MaxGasWasmExecution, uint64(0), "Max gas execution should be positive")
	require.Equal(t, uint64(10_000_000), params.MaxGasWasmExecution, "Default max gas should be 10M")

	// Verify security settings
	require.True(t, params.SecurityAnalysisEnabled, "Security analysis should be enabled by default")
	require.True(t, params.RequireAdminForMigrate, "Admin should be required for migration by default")
}

// TestWASMParamsValidation verifies parameter validation
func TestWASMParamsValidation(t *testing.T) {
	tests := []struct {
		name        string
		params      *types.Params
		expectError bool
	}{
		{
			name:        "default params are valid",
			params:      types.DefaultParams(),
			expectError: false,
		},
		{
			name: "zero gas limit is invalid",
			params: &types.Params{
				CodeUploadAccess:             types.AccessConfig{Permission: types.AccessTypeEverybody},
				InstantiateDefaultPermission: types.AccessTypeEverybody,
				MaxWasmCodeSize:              600 * 1024,
				MaxGasWasmExecution:          0, // Invalid
				SecurityAnalysisEnabled:      true,
				RequireAdminForMigrate:       true,
			},
			expectError: true,
		},
		{
			name: "zero contract size is invalid",
			params: &types.Params{
				CodeUploadAccess:             types.AccessConfig{Permission: types.AccessTypeEverybody},
				InstantiateDefaultPermission: types.AccessTypeEverybody,
				MaxWasmCodeSize:              0, // Invalid
				MaxGasWasmExecution:          10_000_000,
				SecurityAnalysisEnabled:      true,
				RequireAdminForMigrate:       true,
			},
			expectError: true,
		},
		{
			name: "excessive contract size is invalid",
			params: &types.Params{
				CodeUploadAccess:             types.AccessConfig{Permission: types.AccessTypeEverybody},
				InstantiateDefaultPermission: types.AccessTypeEverybody,
				MaxWasmCodeSize:              20 * 1024 * 1024, // 20MB - exceeds 10MB limit
				MaxGasWasmExecution:          10_000_000,
				SecurityAnalysisEnabled:      true,
				RequireAdminForMigrate:       true,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateParams(tc.params)
			if tc.expectError {
				require.Error(t, err, "Expected validation error")
			} else {
				require.NoError(t, err, "Expected no validation error")
			}
		})
	}
}

// TestWASMSecuritySettings verifies security-related settings
func TestWASMSecuritySettings(t *testing.T) {
	params := types.DefaultParams()

	// Security: Security analysis enabled by default
	require.True(t, params.SecurityAnalysisEnabled,
		"SECURITY: Security analysis should be enabled by default")

	// Security: Admin required for migration by default
	require.True(t, params.RequireAdminForMigrate,
		"SECURITY: Admin should be required for contract migration")

	// Security: Reasonable gas limits
	require.LessOrEqual(t, params.MaxGasWasmExecution, uint64(100_000_000),
		"SECURITY: Gas limit should have reasonable upper bound")

	// Security: Contract size limits
	require.LessOrEqual(t, params.MaxWasmCodeSize, uint64(10*1024*1024),
		"SECURITY: Contract size should be limited (10MB max)")
}

// TestWASMAccessTypes verifies access type constants
func TestWASMAccessTypes(t *testing.T) {
	// Verify access types are defined
	require.NotEqual(t, types.AccessTypeUnspecified, types.AccessTypeNobody)
	require.NotEqual(t, types.AccessTypeNobody, types.AccessTypeOnlyAddress)
	require.NotEqual(t, types.AccessTypeOnlyAddress, types.AccessTypeEverybody)
}

// Note: Full integration tests are in chain/x/wasm/keeper/integration_test.go
// These include:
// - TestFullContractLifecycle: Complete contract instantiation, execution, pause/unpause flow
// - TestPolicyEnforcement_Blacklist: Sender blacklist enforcement
// - TestPolicyEnforcement_Whitelist: Sender whitelist enforcement
// - TestSecurityValidation: Security controls and limits
// - TestMetricsTracking: Gas and execution metrics
// - TestCircuitBreaker: Circuit breaker functionality
//
// Run with: go test -v ./chain/x/wasm/keeper/...
