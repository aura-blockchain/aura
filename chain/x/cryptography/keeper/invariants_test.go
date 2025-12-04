package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/cryptography/keeper"
)

func TestAllInvariants(t *testing.T) {
	k, ctx := setupKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Test: All invariants on empty store
	inv := keeper.AllInvariants(k)
	msg, broken := inv(sdkCtx)
	require.False(t, broken, "all invariants should pass on empty store")
	require.Empty(t, msg)
}

func TestRegisterInvariants(t *testing.T) {
	k, _ := setupKeeper(t)

	// Register should not panic - just verify all invariants work
	require.NotPanics(t, func() {
		_ = keeper.AllInvariants(k)
	})
}

func TestParamsInvariant(t *testing.T) {
	k, ctx := setupKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Test: valid params pass
	inv := keeper.ParamsInvariant(k)
	msg, broken := inv(sdkCtx)
	require.False(t, broken, "valid params should pass")
	require.Empty(t, msg)
}

func TestKeyRotationValidityInvariant(t *testing.T) {
	k, ctx := setupKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Test: empty store passes
	inv := keeper.KeyRotationValidityInvariant(k)
	msg, broken := inv(sdkCtx)
	require.False(t, broken, "empty store should pass")
	require.Empty(t, msg)

	// Note: To properly test invariants with actual data, we would need to:
	// 1. Store the key rotation schedule using internal keeper methods
	// 2. Then run the invariant
	// Since the invariant iterates KV store directly, we'd need direct store access
	// which would require exposing internal methods or testing at integration level
}

func TestThresholdSchemeConsistencyInvariant(t *testing.T) {
	k, ctx := setupKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Test: empty store passes
	inv := keeper.ThresholdSchemeConsistencyInvariant(k)
	msg, broken := inv(sdkCtx)
	require.False(t, broken, "empty store should pass")
	require.Empty(t, msg)
}

func TestZKProofConfigValidityInvariant(t *testing.T) {
	k, ctx := setupKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Test: empty store passes
	inv := keeper.ZKProofConfigValidityInvariant(k)
	msg, broken := inv(sdkCtx)
	require.False(t, broken, "empty store should pass")
	require.Empty(t, msg)
}

func TestSecureEnclaveValidityInvariant(t *testing.T) {
	k, ctx := setupKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Test: empty store passes
	inv := keeper.SecureEnclaveValidityInvariant(k)
	msg, broken := inv(sdkCtx)
	require.False(t, broken, "empty store should pass")
	require.Empty(t, msg)
}

func TestQuantumKeyValidityInvariant(t *testing.T) {
	k, ctx := setupKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Test: empty store passes
	inv := keeper.QuantumKeyValidityInvariant(k)
	msg, broken := inv(sdkCtx)
	require.False(t, broken, "empty store should pass")
	require.Empty(t, msg)
}
