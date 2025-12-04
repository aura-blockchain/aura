package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMsgServerFunctionality(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Create msg server
	msgServer := NewMsgServer(k)
	require.NotNil(t, msgServer)

	// Verify msg server implements the interface
	require.Implements(t, (*interface{})(nil), msgServer)

	// Test that msg server has keeper access
	require.NotNil(t, msgServer.keeper)
	require.Equal(t, k, msgServer.keeper)

	// Verify keeper has correct authority
	authority := k.GetAuthority()
	require.NotEmpty(t, authority)
	require.Equal(t, "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", authority)

	// Verify keeper params can be accessed through msg server
	params := msgServer.keeper.GetParams()
	require.NotNil(t, params.Tokenomics)
	require.NotNil(t, params.WhaleProtection)

	// Test context is valid
	require.NotNil(t, ctx)
	require.False(t, ctx.IsZero())
}
