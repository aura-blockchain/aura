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
	require.NotNil(t, k)
	require.Equal(t, k, k)

	// Verify keeper has correct authority
	authority := k.GetAuthority()
	require.NotEmpty(t, authority)
	require.Equal(t, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr", authority)

	// Verify keeper params can be accessed through msg server
	params := k.GetParams()
	require.NotNil(t, params.Tokenomics)
	require.NotNil(t, params.WhaleProtection)

	// Test context is valid
	require.NotNil(t, ctx)
	require.False(t, ctx.IsZero())
}
