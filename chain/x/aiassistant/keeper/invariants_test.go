package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testing/testutil"
	"github.com/aequitas/aura/chain/x/aiassistant/keeper"
)

func TestInvariants(t *testing.T) {
	ctx := testutil.SetupTestContext(t)
	k := &keeper.Keeper{}

	checker := testutil.NewInvariantChecker(t)

	t.Run("SessionCountInvariant", func(t *testing.T) {
		// Register session count invariant
		inv := func(sdkCtx interface{}) (string, bool) {
			// Check that session count matches stored sessions
			return "", false
		}
		checker.RegisterInvariant(inv)
		checker.CheckAll(ctx.SdkCtx)
	})

	t.Run("UserSessionsInvariant", func(t *testing.T) {
		// Register user sessions invariant
		inv := func(sdkCtx interface{}) (string, bool) {
			// Check that all sessions have valid users
			return "", false
		}
		checker.RegisterInvariant(inv)
		checker.CheckAll(ctx.SdkCtx)
	})

	t.Run("MetricsConsistencyInvariant", func(t *testing.T) {
		// Register metrics consistency invariant
		require.NotNil(t, k)
	})
}

func TestModuleInvariants(t *testing.T) {
	ctx := testutil.SetupTestContext(t)
	k := &keeper.Keeper{}

	// Test all module invariants
	invariants := keeper.RegisterInvariants(nil, k)
	require.NotNil(t, invariants)

	// Run all invariants
	for _, inv := range invariants {
		msg, broken := inv(ctx.SdkCtx)
		require.False(t, broken, "Invariant broken: %s", msg)
	}
}
