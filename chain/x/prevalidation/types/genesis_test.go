// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultGenesis(t *testing.T) {
	params := DefaultParams()
	gs := &GenesisState{
		Params: *params, // Dereference pointer
	}

	require.NotNil(t, gs)

	// Validate params - pass pointer to Params
	require.NoError(t, ValidateParams(&gs.Params))
}

func TestGenesisState_EmptyCollections(t *testing.T) {
	params := DefaultParams()
	gs := &GenesisState{
		Params: *params, // Dereference pointer
	}

	// Params is a value type, not a pointer, so we just check it's set correctly
	require.Equal(t, true, gs.Params.Enabled)
}

func TestTransactionType_Constants(t *testing.T) {
	// Test that transaction type constants are properly defined
	require.NotEqual(t, TransactionType_TX_TYPE_UNSPECIFIED, TransactionType_TX_TYPE_IR_COMPLETION)
	require.NotEqual(t, TransactionType_TX_TYPE_IR_COMPLETION, TransactionType_TX_TYPE_DEX_SWAP)
	require.NotEqual(t, TransactionType_TX_TYPE_DEX_SWAP, TransactionType_TX_TYPE_LP_DEPOSIT)
}

func TestValidationStatus_Constants(t *testing.T) {
	// Test that validation status constants are properly defined
	require.NotEqual(t, ValidationStatus_VALIDATION_STATUS_UNSPECIFIED, ValidationStatus_VALIDATION_STATUS_PENDING)
	require.NotEqual(t, ValidationStatus_VALIDATION_STATUS_PENDING, ValidationStatus_VALIDATION_STATUS_VALIDATED)
	require.NotEqual(t, ValidationStatus_VALIDATION_STATUS_VALIDATED, ValidationStatus_VALIDATION_STATUS_EXECUTED)
}

func TestCacheStrategy_Constants(t *testing.T) {
	// Test that cache strategy constants are properly defined
	require.NotEqual(t, CacheStrategy_CACHE_STRATEGY_UNSPECIFIED, CacheStrategy_CACHE_STRATEGY_LRU)
	require.NotEqual(t, CacheStrategy_CACHE_STRATEGY_LRU, CacheStrategy_CACHE_STRATEGY_LFU)
	require.NotEqual(t, CacheStrategy_CACHE_STRATEGY_LFU, CacheStrategy_CACHE_STRATEGY_FIFO)
	require.NotEqual(t, CacheStrategy_CACHE_STRATEGY_FIFO, CacheStrategy_CACHE_STRATEGY_ADAPTIVE)
}
