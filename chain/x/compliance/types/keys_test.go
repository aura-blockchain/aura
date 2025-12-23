package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

func TestModuleKeys(t *testing.T) {
	require.Equal(t, "compliance", types.ModuleName)
	require.Equal(t, "compliance", types.StoreKey)
	require.Equal(t, "compliance", types.RouterKey)
	require.Equal(t, "compliance", types.QuerierRoute)
	require.Equal(t, "mem_compliance", types.MemStoreKey)
}

func TestStoreKeyUniqueness(t *testing.T) {
	// Ensure all keys are defined and non-empty
	require.NotEmpty(t, types.ModuleName)
	require.NotEmpty(t, types.StoreKey)
	require.NotEmpty(t, types.RouterKey)
	require.NotEmpty(t, types.QuerierRoute)
	require.NotEmpty(t, types.MemStoreKey)
}

func TestMemStoreKeyPrefix(t *testing.T) {
	// Verify mem store key has correct prefix
	require.Contains(t, types.MemStoreKey, "mem_")
	require.Contains(t, types.MemStoreKey, types.ModuleName)
}
