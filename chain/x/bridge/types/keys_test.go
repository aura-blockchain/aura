package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

func TestTransferKey(t *testing.T) {
	transferID := "transfer-123"
	key := types.TransferKey(transferID)

	require.NotNil(t, key)
	require.Contains(t, string(key), transferID)
	require.Equal(t, append(types.TransferPrefix, []byte(transferID)...), key)
}

func TestWrappedTokenKey(t *testing.T) {
	wrappedDenom := "wrapped-eth"
	key := types.WrappedTokenKey(wrappedDenom)

	require.NotNil(t, key)
	require.Contains(t, string(key), wrappedDenom)
	require.Equal(t, append(types.WrappedTokenPrefix, []byte(wrappedDenom)...), key)
}

func TestSharedIdentityKey(t *testing.T) {
	auraAddress := "aura1abcdef"
	key := types.SharedIdentityKey(auraAddress)

	require.NotNil(t, key)
	require.Contains(t, string(key), auraAddress)
	require.Equal(t, append(types.SharedIdentityPrefix, []byte(auraAddress)...), key)
}

func TestRelayerKey(t *testing.T) {
	relayerAddress := "aura1relayer"
	key := types.RelayerKey(relayerAddress)

	require.NotNil(t, key)
	require.Contains(t, string(key), relayerAddress)
	require.Equal(t, append(types.RelayerPrefix, []byte(relayerAddress)...), key)
}

func TestTransferKey_Empty(t *testing.T) {
	key := types.TransferKey("")
	require.NotNil(t, key)
	require.Equal(t, types.TransferPrefix, key)
}

func TestWrappedTokenKey_Empty(t *testing.T) {
	key := types.WrappedTokenKey("")
	require.NotNil(t, key)
	require.Equal(t, types.WrappedTokenPrefix, key)
}

func TestSharedIdentityKey_Empty(t *testing.T) {
	key := types.SharedIdentityKey("")
	require.NotNil(t, key)
	require.Equal(t, types.SharedIdentityPrefix, key)
}

func TestRelayerKey_Empty(t *testing.T) {
	key := types.RelayerKey("")
	require.NotNil(t, key)
	require.Equal(t, types.RelayerPrefix, key)
}

func TestKeyUniqueness(t *testing.T) {
	// Ensure different key types don't collide
	transferKey := types.TransferKey("test")
	wrappedKey := types.WrappedTokenKey("test")
	identityKey := types.SharedIdentityKey("test")
	relayerKey := types.RelayerKey("test")

	require.NotEqual(t, transferKey, wrappedKey)
	require.NotEqual(t, transferKey, identityKey)
	require.NotEqual(t, transferKey, relayerKey)
	require.NotEqual(t, wrappedKey, identityKey)
	require.NotEqual(t, wrappedKey, relayerKey)
	require.NotEqual(t, identityKey, relayerKey)
}
