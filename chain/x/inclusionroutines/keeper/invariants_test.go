package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

type mockInvariantRegistry struct{}

func (mockInvariantRegistry) RegisterRoute(_ string, _ string, _ sdk.Invariant) {}

func TestAllInvariants(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)

	inv := AllInvariants(keeper)
	msg, broken := inv(ctx)
	require.False(t, broken, "all invariants should pass on empty store")
	require.Empty(t, msg)
}

func TestRegisterInvariants(t *testing.T) {
	_, keeper := setupInclusionKeeper(t)
	registry := mockInvariantRegistry{}

	require.NotPanics(t, func() {
		RegisterInvariants(registry, keeper)
	})
}
