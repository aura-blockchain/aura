// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModuleConstants(t *testing.T) {
	require.Equal(t, "contractregistry", ModuleName)
	require.Equal(t, "contractregistry", StoreKey)
	require.Equal(t, "contractregistry", RouterKey)
	require.Equal(t, "contractregistry", QuerierRoute)
}

func TestContractStatusValues(t *testing.T) {
	require.Equal(t, ContractStatus(0), ContractStatus_CONTRACT_STATUS_UNSPECIFIED)
	require.Equal(t, ContractStatus(1), ContractStatus_CONTRACT_STATUS_ACTIVE)
	require.Equal(t, ContractStatus(2), ContractStatus_CONTRACT_STATUS_PAUSED)
	require.Equal(t, ContractStatus(3), ContractStatus_CONTRACT_STATUS_DEPRECATED)
	require.Equal(t, ContractStatus(4), ContractStatus_CONTRACT_STATUS_FROZEN)
}

func TestDefaultParamsValues(t *testing.T) {
	params := DefaultParams()

	require.True(t, params.OpenRegistration)
	require.Greater(t, params.MaxContractsPerCreator, uint64(0))
	require.False(t, params.RequireMetadata)
	require.False(t, params.RequireSecurityPolicy)
	require.Greater(t, params.DefaultRateLimit, uint64(0))
	require.Greater(t, params.DefaultMaxGas, uint64(0))
}

func TestContractInfoFields(t *testing.T) {
	info := &ContractInfo{
		ContractAddress: "cosmos1contract",
		CodeId:          1,
		Status:          ContractStatus_CONTRACT_STATUS_ACTIVE,
	}

	require.Equal(t, "cosmos1contract", info.ContractAddress)
	require.Equal(t, uint64(1), info.CodeId)
	require.Equal(t, ContractStatus_CONTRACT_STATUS_ACTIVE, info.Status)
}

func TestContractMetadata(t *testing.T) {
	metadata := &ContractMetadata{
		Name:        "Test Contract",
		Description: "A test smart contract",
		Version:     "1.0.0",
		Tags:        []string{"test", "demo"},
	}

	require.Equal(t, "Test Contract", metadata.Name)
	require.Equal(t, "A test smart contract", metadata.Description)
	require.Equal(t, "1.0.0", metadata.Version)
	require.Len(t, metadata.Tags, 2)
	require.Contains(t, metadata.Tags, "test")
	require.Contains(t, metadata.Tags, "demo")
}

func TestSecurityPolicy(t *testing.T) {
	policy := &SecurityPolicy{
		AllowPause:         true,
		MaxGasPerExecution: 1000000,
		RateLimitPerUser:   100,
	}

	require.True(t, policy.AllowPause)
	require.Equal(t, uint64(1000000), policy.MaxGasPerExecution)
	require.Equal(t, uint64(100), policy.RateLimitPerUser)
}

func TestComplianceRequirements(t *testing.T) {
	compliance := &ComplianceRequirements{
		MinKycLevel: 2,
	}

	require.Equal(t, uint32(2), compliance.MinKycLevel)
}

func TestContractMetrics(t *testing.T) {
	metrics := &ContractMetrics{
		TotalExecutions:      100,
		SuccessfulExecutions: 95,
		FailedExecutions:     5,
		TotalGasUsed:         5000000,
		RateLimitViolations:  2,
		ComplianceFailures:   1,
	}

	require.Equal(t, uint64(100), metrics.TotalExecutions)
	require.Equal(t, uint64(95), metrics.SuccessfulExecutions)
	require.Equal(t, uint64(5), metrics.FailedExecutions)
	require.Equal(t, uint64(5000000), metrics.TotalGasUsed)
	require.Equal(t, uint64(2), metrics.RateLimitViolations)
	require.Equal(t, uint64(1), metrics.ComplianceFailures)
}
