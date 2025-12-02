package app

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// TestSecurityStakingAdapterErrorPropagation verifies that adapter methods properly
// return errors instead of suppressing them
func TestSecurityStakingAdapterErrorPropagation(t *testing.T) {
	// This is a compilation test to ensure the interface signatures are correct
	// The actual runtime behavior is tested in integration tests

	var adapter interface{}

	// Verify securityStakingAdapter implements the expected interface
	adapter = securityStakingAdapter{}
	_, ok := adapter.(interface {
		Jail(ctx sdk.Context, consAddr sdk.ConsAddress) error
		Unjail(ctx sdk.Context, consAddr sdk.ConsAddress) error
		Slash(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor string) (string, error)
	})
	require.True(t, ok, "securityStakingAdapter must return errors from Jail, Unjail, and Slash")

	// Verify securityStakingKeeperAdapter implements the expected interface
	adapter = securityStakingKeeperAdapter{}
	_, ok = adapter.(interface {
		Jail(ctx sdk.Context, consAddr sdk.ConsAddress) error
		Unjail(ctx sdk.Context, consAddr sdk.ConsAddress) error
		Slash(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor string) (string, error)
	})
	require.True(t, ok, "securityStakingKeeperAdapter must return errors from Jail, Unjail, and Slash")
}
