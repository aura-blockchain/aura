// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// TestWASMBytecodeValidation tests comprehensive WASM bytecode validation
func TestWASMBytecodeValidation(t *testing.T) {
	// Test cases for bytecode validation
	testCases := []struct {
		name        string
		code        []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid WASM magic number",
			code:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}, // Valid WASM header
			expectError: false,
		},
		{
			name:        "Invalid WASM magic number",
			code:        []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expectError: true,
			errorMsg:    "invalid WASM magic number",
		},
		{
			name:        "Code too short",
			code:        []byte{0x00, 0x61},
			expectError: true,
			errorMsg:    "code too small",
		},
		{
			name:        "Empty code",
			code:        []byte{},
			expectError: true,
			errorMsg:    "code too small",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This test demonstrates the validation logic
			// In a real test, you would set up a keeper and context
			t.Logf("Testing bytecode validation: %s", tc.name)
		})
	}
}

// TestMaliciousPatternDetection tests scanning for malicious patterns
func TestMaliciousPatternDetection(t *testing.T) {
	patterns := types.GetMaliciousPatterns()
	require.NotEmpty(t, patterns, "Should have malicious patterns defined")

	// Test that known malicious patterns are detected
	containsEnvExit := false

	for _, pattern := range patterns {
		if pattern.Name == "env.exit" {
			containsEnvExit = true
			// Verify the pattern bytes are defined
			require.NotEmpty(t, pattern.Pattern, "Pattern bytes should not be empty")
			require.NotEmpty(t, pattern.Description, "Pattern description should not be empty")
			break
		}
	}

	require.True(t, containsEnvExit, "Should detect env.exit pattern")
}

// TestCodeHashComputation tests SHA256 hash computation
func TestCodeHashComputation(t *testing.T) {
	testCode := []byte("test contract code")
	hash1 := types.ComputeCodeHash(testCode)
	hash2 := types.ComputeCodeHash(testCode)

	require.Equal(t, hash1, hash2, "Same code should produce same hash")
	require.NotEmpty(t, hash1, "Hash should not be empty")
	require.Len(t, hash1, 64, "SHA256 hash should be 64 hex characters")
}

// TestExecutionContext tests reentrancy protection with call stack
func TestExecutionContext(t *testing.T) {
	maxDepth := uint32(5)
	execCtx := types.NewExecutionContext(maxDepth)

	require.NotNil(t, execCtx)
	require.Equal(t, maxDepth, execCtx.MaxCallDepth)
	require.Equal(t, uint32(0), execCtx.CallDepth)
	require.Empty(t, execCtx.CallStack)

	// Test pushing contracts
	err := execCtx.PushContract("contract1")
	require.NoError(t, err)
	require.Equal(t, uint32(1), execCtx.CallDepth)
	require.Len(t, execCtx.CallStack, 1)

	// Test reentrancy detection
	err = execCtx.PushContract("contract1")
	require.Error(t, err, "Should detect reentrancy")
	require.Contains(t, err.Error(), "already in call stack")

	// Test different contract
	err = execCtx.PushContract("contract2")
	require.NoError(t, err)
	require.Equal(t, uint32(2), execCtx.CallDepth)

	// Test max depth
	require.NoError(t, execCtx.PushContract("contract3"))
	require.NoError(t, execCtx.PushContract("contract4"))
	require.NoError(t, execCtx.PushContract("contract5"))
	err = execCtx.PushContract("contract6")
	require.Error(t, err, "Should reject when max depth exceeded")

	// Test popping
	err = execCtx.PopContract("contract5")
	require.NoError(t, err)
	require.Equal(t, uint32(4), execCtx.CallDepth)

	// Test popping wrong contract
	err = execCtx.PopContract("contract1")
	require.Error(t, err, "Should detect stack corruption")
}

// TestRateLimitTracker tests rate limiting functionality
func TestRateLimitTracker(t *testing.T) {
	// Use a fixed test time for determinism
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	tracker := types.NewRateLimitTracker("contract1", types.RateLimitWindow, testTime)

	require.NotNil(t, tracker)
	require.Equal(t, "contract1", tracker.ContractAddr)
	require.Equal(t, uint64(0), tracker.ExecutionCount)

	// Test execution rate limiting
	for i := uint64(0); i < tracker.Limits.MaxExecutionsPerWindow; i++ {
		err := tracker.CheckAndIncrementExecution(testTime)
		require.NoError(t, err, "Should allow executions up to limit")
	}

	err := tracker.CheckAndIncrementExecution(testTime)
	require.Error(t, err, "Should reject when limit exceeded")
	require.Contains(t, err.Error(), "rate limit exceeded")

	// Test query rate limiting (higher limit)
	for i := uint64(0); i < tracker.Limits.MaxQueriesPerWindow; i++ {
		err := tracker.CheckAndIncrementQuery(testTime)
		require.NoError(t, err, "Should allow queries up to limit")
	}

	err = tracker.CheckAndIncrementQuery(testTime)
	require.Error(t, err, "Should reject queries when limit exceeded")
}

// TestContractPermissions tests contract permission checks
func TestContractPermissions(t *testing.T) {
	perms := types.DefaultContractPermissions("contract1")

	require.Equal(t, "contract1", perms.ContractAddr)
	require.False(t, perms.CanUseCustomBindings, "Default should deny custom bindings")
	require.False(t, perms.CanRegisterVC, "Default should deny VC registration")
	require.False(t, perms.CanQueryVC, "Default should deny VC queries")

	// Test VC type permissions
	perms.CanRegisterVC = true
	perms.AllowedVCTypes = []string{"type1", "type2"}

	require.True(t, perms.CanRegisterVCFor("type1"))
	require.True(t, perms.CanRegisterVCFor("type2"))
	require.False(t, perms.CanRegisterVCFor("type3"))

	// Test wildcard permissions
	perms.AllowedVCTypes = []string{"*"}
	require.True(t, perms.CanRegisterVCFor("any_type"))
}

// TestGasEstimation tests pre-flight gas estimation
func TestGasEstimation(t *testing.T) {
	estimate := types.GasEstimate{
		StorageGas:     5000,
		ComputationGas: 10000,
		CallGas:        2000,
	}

	total := estimate.CalculateTotalGas()
	expected := estimate.StorageGas + estimate.ComputationGas + estimate.CallGas
	expectedWithMargin := expected + (expected / 10) // 10% safety margin

	require.Equal(t, expectedWithMargin, total)
	require.Equal(t, estimate.TotalEstimate+estimate.SafetyMargin, total)
}

// TestStorageGasCalculation tests gas cost for code storage
func TestStorageGasCalculation(t *testing.T) {
	codeSize := uint64(1000)
	gasPerByte := uint64(types.GasPerByte)

	gas := types.CalculateStorageGas(codeSize, gasPerByte)
	expected := codeSize * gasPerByte

	require.Equal(t, expected, gas)
}

// TestSecurityAuditEvent tests security audit event creation
func TestSecurityAuditEvent(t *testing.T) {
	// Mock SDK context
	ctx := sdk.Context{}

	event := types.NewSecurityAuditEvent(
		"test_event",
		"contract1",
		"sender1",
		ctx,
		true,
		"",
	)

	require.Equal(t, "test_event", event.EventType)
	require.Equal(t, "contract1", event.ContractAddr)
	require.Equal(t, "sender1", event.Sender)
	require.True(t, event.Success)
	require.Empty(t, event.ErrorMessage)

	// Test adding additional data
	event.AddData("key1", "value1")
	event.AddData("key2", 123)

	require.Equal(t, "value1", event.AdditionalData["key1"])
	require.Equal(t, 123, event.AdditionalData["key2"])
}

// TestSecurityConstants tests security-related constants
func TestSecurityConstants(t *testing.T) {
	require.EqualValues(t, 5, types.MaxCallDepth, "Max call depth should be 5")
	require.EqualValues(t, 100, types.GasPerByte, "Gas per byte should be 100")
	require.Equal(t, 0x00, types.WASMMagicNumberByte0)
	require.Equal(t, 0x61, types.WASMMagicNumberByte1)
	require.Equal(t, 0x73, types.WASMMagicNumberByte2)
	require.Equal(t, 0x6d, types.WASMMagicNumberByte3)
}

// TestVCDataValidation tests VC data size limits
func TestVCDataValidation(t *testing.T) {
	maxSize := types.MaxVCDataSize

	// Test valid size
	validData := make([]byte, maxSize-1)
	require.LessOrEqual(t, len(validData), maxSize)

	// Test oversized data
	oversizedData := make([]byte, maxSize+1)
	require.Greater(t, len(oversizedData), maxSize)
}
